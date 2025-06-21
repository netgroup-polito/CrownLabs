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
	ExpirationCheckInterval   time.Duration
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
				if !ok || template.Spec.DeleteAfter == "never" {
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

	isDeleted, err := r.HandleInstanceExpiration(ctx, &instance)
	if err != nil {
		return ctrl.Result{}, err
	}
	if isDeleted {
		tracer.Step("stale instance deleted and handled")
		return ctrl.Result{}, nil
	}

	tracer.Step("stale instance check complete")
	dbgLog.Info("requeueing instance")
	return ctrl.Result{RequeueAfter: r.ExpirationCheckInterval}, nil
}

// HandleInstanceExpiration checks if the instance is expired and handles its deletion and notification.
func (r *InstanceExpirationReconciler) HandleInstanceExpiration(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	log := ctrl.LoggerFrom(ctx)

	isDeleted, err := r.DeleteStaleInstance(ctx, instance)
	if err != nil {
		log.Error(err, "failed to delete stale instance")
		return false, err
	}
	if !isDeleted {
		return false, nil
	}

	tenant, err := GetTenantFromInstance(ctx, r.Client, instance)
	if err != nil {
		log.Error(err, "failed retrieving tenant from instance")
		return true, err // instance deleted, but email not sent
	}

	if err := NotifyInstanceExpiring(ctx, instance, tenant, r.MailClient); err != nil {
		log.Error(err, "failed sending notification email")
		return true, err // instance deleted, but email not sent
	}

	log.Info("Notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
	return true, nil
}

// DeleteStaleInstance checks if the instance is expired based on its creation timestamp and the deleteAfter field in the template.
func (r *InstanceExpirationReconciler) DeleteStaleInstance(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("delete-stale-instances")

	// Get the template associated with the instance.
	template, err := r.GetTemplateForInstance(ctx, instance)
	if err != nil {
		log.Error(err, "failed to get template for instance", "instance", instance.GetName(), "namespace", instance.GetNamespace())
		return false, fmt.Errorf("failed to get template for instance %s/%s: %w", instance.GetNamespace(), instance.GetName(), err)
	}

	// If the template's deleteAfter field is set to "never", skip deletion.
	if template.Spec.DeleteAfter == "never" {
		return false, nil
	}

	lifespan, err := GetLifespanFromTemplate(template)
	if err != nil {
		return false, err
	}

	expired, err := HasInstanceExpired(instance, lifespan)
	if err != nil {
		return false, fmt.Errorf("failed to compute expiration: %w", err)
	}

	if !expired {
		log.Info("Instance is not expired, skipping deletion", "name", instance.GetName(), "namespace", instance.GetNamespace())
		return false, nil
	}

	if err := r.DeleteInstance(ctx, instance); err != nil {
		return false, err
	}

	log.Info("Instance is expired and has been deleted", "name", instance.GetName(), "namespace", instance.GetNamespace())
	return true, nil
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
	if template.Spec.DeleteAfter == "never" {
		return math.Inf(1), nil
	}
	return ConvertToSeconds(template.Spec.DeleteAfter)
}

// ConvertToSeconds converts a deleteAfter string to seconds.
func ConvertToSeconds(deleteAfter string) (float64, error) {
	if deleteAfter == "never" {
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

// HasInstanceExpired determines if the instance creation time has exceeded the given lifespan (in seconds) relative to the current time.
func HasInstanceExpired(instance *clv1alpha2.Instance, lifespanSeconds float64) (bool, error) {
	created := instance.GetCreationTimestamp().Time
	duration := time.Since(created).Seconds()
	return duration > lifespanSeconds, nil
}

// DeleteInstance attempts to delete the instance and handles NotFound gracefully.
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
