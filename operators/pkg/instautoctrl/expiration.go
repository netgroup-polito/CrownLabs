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

// SetupWithManager registers a new controller for InstanceTerminationReconciler resources.
func (r *InstanceExpirationReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Instance{}).
		Watches(
			&clv1alpha2.Template{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				var requests []reconcile.Request

				template, ok := obj.(*clv1alpha2.Template)
				if !ok || template.Spec.DeleteAfter == "never" {
					return requests
				}

				var instances clv1alpha2.InstanceList
				if err := r.Client.List(ctx, &instances,
					client.InNamespace(template.Namespace),
					client.MatchingLabels{"crownlabs.polito.it/template": template.Name},
				); err != nil {
					ctrl.LoggerFrom(ctx).Error(err, "failed listing instances for template", "template", template.Name)
					return requests
				}

				for i := range instances.Items {
					instance := &instances.Items[i]
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      instance.Name,
							Namespace: instance.Namespace,
						},
					})
				}

				return requests
			}),
			builder.WithPredicates(deleteAfterChanged), // opzionale ma consigliato
		).
		Named("instance-expiration-termination").
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "InstanceExpiration")).
		Complete(r)
}

// Reconcile reconciles the status of the InstanceSnapshot resource.
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

	// check if the instance reached the maximum time of lifetime and if so delete it
	isDeleted, err := r.deleteStaleInstance(ctx, &instance)
	if err != nil {
		log.Error(err, "failed delete-stale-instance")
	}
	if isDeleted {
		log.Info("Instance has been deleted", "instance", instance.Name)
		tenant, err := r.GetTenantFromInstance(ctx, &instance)
		if err != nil {
			log.Error(err, "failed retrieving tenant from instance")
			return ctrl.Result{}, err
		}
		err = r.SendNotification(ctx, &instance, tenant.Spec.Email)
		if err != nil {
			log.Error(err, "failed sending notification email")
			return ctrl.Result{}, err
		}
		log.Info("Notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
		tracer.Step("Stale instance deleted")
		return ctrl.Result{}, nil
	}

	tracer.Step("stale instance done")

	dbgLog.Info("requeueing instance")
	return ctrl.Result{RequeueAfter: 24 * time.Hour}, nil // TODO revisit this value, it is a placeholder.
}

// TerminateInstance terminates the Instance.
// TODO move somewhere else, as it is used in other controllers too.
func (r *InstanceExpirationReconciler) TerminateInstance(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx).WithName("termination")
	log.Info("Terminating instance", "instance", instance.Name, " in namespace", instance.Namespace)

	var template clv1alpha2.Template
	var err = r.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.Template.Name,
		Namespace: instance.Namespace,
	}, &template)
	if err != nil {
		log.Error(err, "Unable to fetch the instance template.")
		return err
	}

	var environment = template.Spec.EnvironmentList[0]
	if environment.Persistent {
		log.Info("Stopping persistent instance...")
		instance.Spec.Running = false
		return r.Update(ctx, instance)
	}
	log.Info("Deleting non-persistent instance...")
	return r.Delete(ctx, instance)
}

// SendNotification sends an email to the user to notify that the instance will be terminated/stopped if they do not use it anymore.
func (r *InstanceExpirationReconciler) SendNotification(ctx context.Context, instance *clv1alpha2.Instance, userEmail string) error {
	log := ctrl.LoggerFrom(ctx).WithName("notification-email-instance")
	log.Info("sending email notification to user", "instance", instance.Name, "email", userEmail)
	emailBody := fmt.Sprintf(
		"Dear user,\n\n"+
			"Your instance %s has reached the maximum lifetime and has now been terminated.\n\n"+
			"Best regards,\n"+
			"CrownLabs Team",
		instance.Name,
	)
	err := r.MailClient.SendMail([]string{userEmail}, "CrownLabs Instance Termination Alert", emailBody)
	if err != nil {
		log.Error(err, "failed sending email notification")
		return err
	}
	log.Info("The notification to the tenant has been sent", "instance", instance.Name)

	return nil
}

// GetTenantFromInstance retrieves the Tenant object associated with the Instance.
func (r *InstanceExpirationReconciler) GetTenantFromInstance(ctx context.Context, instance *clv1alpha2.Instance) (clv1alpha2.Tenant, error) {
	log := ctrl.LoggerFrom(ctx).WithName("get-user-from-instance")
	log.Info("getting user from instance", "instance", instance.Name)

	tenant := &clv1alpha2.Tenant{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      instance.Spec.Tenant.Name,
		Namespace: instance.Namespace,
	}, tenant); err != nil {
		if kerrors.IsNotFound(err) {
			log.Error(err, "user not found")
			return clv1alpha2.Tenant{}, fmt.Errorf("user %s not found", instance.Spec.Tenant.Name)
		}
		log.Error(err, "failed retrieving user")
		return clv1alpha2.Tenant{}, err
	}
	return *tenant, nil
}

func isInstanceExpired(creationTimestamp string, lifespan float64) (bool, error) {
	created, err := time.Parse(time.RFC3339, creationTimestamp)
	if err != nil {
		return false, err
	}
	duration := time.Since(created).Seconds()
	return duration > lifespan, nil
}

func convertToSeconds(deleteAfter string) (float64, error) {
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

func (r *InstanceExpirationReconciler) deleteStaleInstance(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("delete-stale-instances")

	// get the template from the instance
	template := &clv1alpha2.Template{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.Template.Name,
		Namespace: instance.Spec.Template.Namespace,
	}, template)

	if err != nil {
		if kerrors.IsNotFound(err) {
			return false, fmt.Errorf("template not found: name=%s, namespace=%s", instance.Spec.Template.Name, instance.Spec.Template.Namespace)
		}
		return false, fmt.Errorf("failed to retrieve template for instance %s: %w", instance.Name, err)
	}

	// get the deleteAfter field from the template
	deleteAfter := template.Spec.DeleteAfter
	if deleteAfter == "never" {
		return false, fmt.Errorf("template %s has deleteAfter set to 'never', skipping deletion", template.Name)
	}

	lifespan, err := convertToSeconds(deleteAfter)
	if err != nil {
		return false, err
	}

	creationTimestamp := instance.GetCreationTimestamp().Time.Format(time.RFC3339)
	expired, err := isInstanceExpired(creationTimestamp, lifespan)
	if err != nil {
		return false, fmt.Errorf("failed to compute expiration: %w", err)
	}

	if expired {
		err := r.Client.Delete(ctx, instance)
		if err != nil {
			if kerrors.IsNotFound(err) {
				log.Info("Instance already deleted", "instance", instance.GetName(), "namespace", instance.GetNamespace())
				return false, nil
			}
			return false, fmt.Errorf("failed to delete instance %s/%s: %w", instance.GetNamespace(), instance.GetName(), err)
		}
		log.Info("Instance is expired and has been deleted", instance.GetName(), instance.GetNamespace())
		return true, nil
	}
	log.Info("Instance is not expired, skipping deletion", instance.GetName(), instance.GetNamespace())
	return false, nil
}
