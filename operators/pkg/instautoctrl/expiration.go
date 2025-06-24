// Copyright 2020-2025 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package instautoctrl contains the controller for Instance Inactive Termination
package instautoctrl

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/trace"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// InstanceExpirationReconciler watches for instances to be terminated.
type InstanceExpirationReconciler struct {
	client.Client
	EventsRecorder            record.EventRecorder
	Scheme                    *runtime.Scheme
	NamespaceWhitelist        metav1.LabelSelector
	StatusCheckRequestTimeout time.Duration
	MailClient                *utils.MailClient
	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

var deleteAfterRegex = regexp.MustCompile(`^(\d+)([mhd])$`)

// SetupWithManager registers a new controller for InstanceExpirationReconciler resources.
func (r *InstanceExpirationReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Instance{}).
		Watches(
			&clv1alpha2.Template{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				template, ok := obj.(*clv1alpha2.Template)
				if !ok || template.Spec.DeleteAfter == neverTimeoutValue {
					return nil
				}
				return getTemplateInstanceRequests(ctx, r.Client, template)
			}),
			builder.WithPredicates(deleteAfterChanged),
		).
		Named("instance-expiration-termination").
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "InstanceExpiration")).
		Complete(r)
}

// Reconcile reconciles the status of the InstanceExpirationReconciler resource.
func (r *InstanceExpirationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}
	log := ctrl.LoggerFrom(ctx, "instance", req.NamespacedName)
	dbgLog := log.V(utils.LogDebugLevel)
	tracer := trace.New("reconcile", trace.Field{Key: "instance", Value: req.NamespacedName})
	ctx = ctrl.LoggerInto(trace.ContextWithTrace(ctx, tracer), log)

	// Get the instance object.
	var instance clv1alpha2.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving instance")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	tracer.Step("instance retrieved")

	// Check the selector label, in order to know whether to perform or not reconciliation.
	if proceed, err := utils.CheckSelectorLabel(ctx, r.Client, instance.GetNamespace(), r.NamespaceWhitelist.MatchLabels); !proceed {
		if err != nil {
			err = fmt.Errorf("failed checking selector label: %w", err)
		}
		return ctrl.Result{}, err
	}

	// Get the template associated with the instance
	template, err := r.GetTemplateForInstance(ctx, &instance)
	if err != nil {
		log.Error(err, "failed to get template for instance", "instance", instance.GetName(), "namespace", instance.GetNamespace())
		return ctrl.Result{}, fmt.Errorf("failed to get template for instance %s/%s: %w", instance.GetNamespace(), instance.GetName(), err)
	}

	// If the template's deleteAfter field is set to neverTimeoutValue , never delete
	if template.Spec.DeleteAfter == neverTimeoutValue {
		log.Info("Instance marked as never delete", "name", instance.GetName(), "namespace", instance.GetNamespace())
		dbgLog.Info("Instance marked as never delete", "instance", instance.GetName(), "namespace", instance.GetNamespace())
		return ctrl.Result{}, nil
	}

	if remaining, err := GetRemainingTime(&instance, template); err != nil {
		log.Error(err, "failed to calculate remaining time for instance", "instance", instance.GetName(), "namespace", instance.GetNamespace())
		return ctrl.Result{}, fmt.Errorf("failed to calculate remaining time for instance %s/%s: %w", instance.GetNamespace(), instance.GetName(), err)
	} else if remaining > 0 {
		log.Info("Instance still active, requeuing", "remaining", remaining.String(), "instance", instance.GetName(), "namespace", instance.GetNamespace())
		dbgLog.Info("Instance still active, requeuing", "remaining", remaining.String(), "instance", instance.GetName(), "namespace", instance.GetNamespace())
		tracer.Step("instance still active, requeuing")
		return ctrl.Result{RequeueAfter: remaining}, nil
	}

	// If we reached this point, instance is expired and must be deleted
	if err := r.DeleteInstance(ctx, &instance); err != nil {
		log.Error(err, "failed to delete instance")
		return ctrl.Result{}, err
	}

	// Send notification
	if err := r.NotifyInstanceDeletion(ctx, &instance); err != nil {
		log.Error(err, "failed to send deletion notification")
		return ctrl.Result{}, fmt.Errorf("failed to send deletion notification: %w", err)
	}

	tracer.Step("instance deleted and notification sent")
	dbgLog.Info("Instance deletion and notification completed", "instance", instance.GetName(), "namespace", instance.GetNamespace())
	return ctrl.Result{}, nil
}

// GetTemplateForInstance fetches the template associated with an instance.
func (r *InstanceExpirationReconciler) GetTemplateForInstance(ctx context.Context, instance *clv1alpha2.Instance) (*clv1alpha2.Template, error) {
	template := &clv1alpha2.Template{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.Template.Name,
		Namespace: instance.Spec.Template.Namespace,
	}, template)

	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, fmt.Errorf("template not found: name=%s, namespace=%s", instance.Spec.Template.Name, instance.Spec.Template.Namespace)
		}
		return nil, fmt.Errorf("failed to retrieve template: %w", err)
	}
	return template, nil
}

// GetLifespanFromTemplate converts the deleteAfter field into a duration in seconds.
func GetLifespanFromTemplate(template *clv1alpha2.Template) (float64, error) {
	if template.Spec.DeleteAfter == neverTimeoutValue {
		return math.Inf(1), nil
	}
	return ConvertToSeconds(template.Spec.DeleteAfter)
}

// ConvertToSeconds converts a deleteAfter string to seconds.
func ConvertToSeconds(deleteAfter string) (float64, error) {
	if deleteAfter == neverTimeoutValue {
		return math.Inf(1), nil
	}

	matches := deleteAfterRegex.FindStringSubmatch(deleteAfter)
	if matches == nil {
		return 0, fmt.Errorf("invalid deleteAfter format: %s", deleteAfter)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	unit := matches[2]
	switch unit {
	case "m":
		return float64(value * 60), nil
	case "h":
		return float64(value * 3600), nil
	case "d":
		return float64(value * 86400), nil
	default:
		return 0, fmt.Errorf("unsupported time unit: %s", unit)
	}
}

// GetRemainingTime returns the remaining time before expiration as a time.Duration.
// If the instance has already expired, the returned duration will be ≤ 0.
func GetRemainingTime(instance *clv1alpha2.Instance, template *clv1alpha2.Template) (time.Duration, error) {
	// Get lifespan from template's deleteAfter field
	if template.Spec.DeleteAfter == neverTimeoutValue {
		// Return maximum duration for neverTimeoutValue
		return time.Duration(math.MaxInt64), nil
	}

	lifespanSeconds, err := ConvertToSeconds(template.Spec.DeleteAfter)
	if err != nil {
		return 0, fmt.Errorf("failed to convert deleteAfter to seconds: %w", err)
	}

	// Calculate remaining time
	created := instance.GetCreationTimestamp().Time
	elapsed := time.Since(created)
	remaining := time.Duration(lifespanSeconds)*time.Second - elapsed
	requeueTime := remaining
	requeueTime += 1 * time.Minute
	return requeueTime, nil
}

// DeleteInstance attempts to delete the instance.
func (r *InstanceExpirationReconciler) DeleteInstance(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx)

	if err := r.Client.Delete(ctx, instance); err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("Instance already deleted", "name", instance.GetName(), "namespace", instance.GetNamespace())
			return nil
		}
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	log.Info("Instance has been deleted", "name", instance.GetName(), "namespace", instance.GetNamespace())
	return nil
}

// NotifyInstanceDeletion handles sending notification emails when an instance is deleted.
func (r *InstanceExpirationReconciler) NotifyInstanceDeletion(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx).WithName("notify-instance-deletion")

	// Get tenant information for notification
	tenant, err := GetTenantFromInstance(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("failed retrieving tenant from instance: %w", err)
	}

	// Send the notification email
	if err := SendExpiringNotification(ctx, instance, tenant, r.MailClient); err != nil {
		return fmt.Errorf("failed sending notification email: %w", err)
	}

	log.Info("Notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
	return nil
}
