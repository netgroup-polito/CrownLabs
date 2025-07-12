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
	"time"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/trace"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	pkgcontext "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/mail"
)

// InstanceExpirationReconciler watches for instances to be terminated.
type InstanceExpirationReconciler struct {
	client.Client
	EventsRecorder                record.EventRecorder
	Scheme                        *runtime.Scheme
	NamespaceWhitelist            metav1.LabelSelector
	StatusCheckRequestTimeout     time.Duration
	EnableExpirationNotifications bool
	MailClient                    *mail.Client
	NotificationInterval          time.Duration
	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// SetupWithManager registers a new controller for InstanceExpirationReconciler resources.
// The controller is configured to watch for Instance resources and Template resources.
// For the instance resources, it is configured to only reconcile instances at the creation time (to calculate the expiration time) and at the deletion time. Updates on the instance resources are ignored by this reconciler.
// For the template resources, it is configured to reconcile instances when the template's deleteAfter field is changed. In this case, it will enqueue all the instances that are associated with that template.
// To avoid unnecessary reconciliations, the controller avoid reconciling instances whose template's deleteAfter field is set to neverTimeoutValue, which means that the instance will never be deleted.
func (r *InstanceExpirationReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Instance{}, builder.WithPredicates(instanceTriggered)).
		Watches(
			&clv1alpha2.Template{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				template, ok := obj.(*clv1alpha2.Template)
				if !ok || template.Spec.DeleteAfter == NeverTimeoutValue {
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

	instance, template, tenant, err := GetInstanceTemplateTenant(ctx, req, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	tracer.Step("instance, template and tenant retrieved")

	// Get lifespan from template's deleteAfter field
	deleteAfter := template.Spec.DeleteAfter

	ctx, _ = pkgcontext.TemplateInto(ctx, template)
	ctx, _ = pkgcontext.InstanceInto(ctx, instance)
	ctx, _ = pkgcontext.TenantInto(ctx, tenant)

	// If the template's deleteAfter field is set to neverTimeoutValue , never delete
	if deleteAfter == NeverTimeoutValue {
		log.Info("Instance marked as never delete", "name", instance.GetName(), "namespace", instance.GetNamespace())
		dbgLog.Info("Instance marked as never delete", "instance", instance.GetName(), "namespace", instance.GetNamespace())
		return ctrl.Result{}, nil
	}

	remainingTime, err := r.CheckInstanceExpiration(ctx, deleteAfter)
	if err != nil {
		log.Error(err, "failed to check instance expiration")
		return ctrl.Result{}, err
	}

	tracer.Step("expiration checked")

	if remainingTime <= 0 {
		tenant := pkgcontext.TenantFrom(ctx)
		if tenant == nil {
			return ctrl.Result{}, fmt.Errorf("tenant not found in context")
		}

		if r.EnableExpirationNotifications {
			shouldSendWarning, err := r.ShouldSendWarningNotification(ctx)
			if err != nil {
				log.Error(err, "failed to check if warning notification should be sent")
				return ctrl.Result{RequeueAfter: r.NotificationInterval}, err
			}
			if shouldSendWarning {
				if err := SendExpiringWarningNotification(ctx, r.MailClient); err != nil {
					return ctrl.Result{RequeueAfter: r.NotificationInterval}, err
				}
				return ctrl.Result{RequeueAfter: r.NotificationInterval}, nil
			}
		}

		// If we reached this point, instance is expired and must be deleted
		if r.EnableExpirationNotifications {
			if instance.Annotations[forge.ExpiringWarningNotificationAnnotation] == "true" {
				if err := r.DeleteInstance(ctx); err != nil {
					log.Error(err, "failed to delete instance")
					return ctrl.Result{}, err
				}
			} else {
				return ctrl.Result{RequeueAfter: r.NotificationInterval}, nil
			}
		} else {
			if err := r.DeleteInstance(ctx); err != nil {
				log.Error(err, "failed to delete instance")
				return ctrl.Result{}, err
			}
		}

		tracer.Step("instance deleted")

		// Send notification
		if err := r.NotifyInstanceDeletion(ctx); err != nil {
			log.Error(err, "failed to send deletion notification")
			return ctrl.Result{}, err
		}
		tracer.Step("deletion notification sent")
		dbgLog.Info("Instance deletion and notification completed", "instance", instance.GetName(), "namespace", instance.GetNamespace())
	}

	// Calculate requeue time at the instance inactive deadline time:
	// if the instance is not yet to be terminated, we requeue it after the remaining time
	requeueTime := remainingTime
	// add 1 minute to the remaining time to avoid requeueing just before the deadline
	// avoiding a double requeue
	requeueTime += 1 * time.Minute
	log.Info("Remaining time before expiration", "remainingTime", remainingTime.String(), "instance", instance.GetName(), "namespace", instance.GetNamespace())

	dbgLog.Info("requeueing instance")
	return ctrl.Result{RequeueAfter: requeueTime}, nil
}

// CheckInstanceExpiration returns the remaining time before expiration as a time.Duration.
func (r *InstanceExpirationReconciler) CheckInstanceExpiration(ctx context.Context, deleteAfter string) (time.Duration, error) {
	log := ctrl.LoggerFrom(ctx).WithName("get-remaining-time")
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return 0, fmt.Errorf("instance not found in context")
	}

	expirationDuration, err := ParseDurationWithDays(ctx, deleteAfter)
	if err != nil {
		log.Error(err, "failed to parse deleteAfter duration")
		return 0, err
	}

	// Check if the instance is expired
	remainingTime := expirationDuration - time.Since(instance.CreationTimestamp.Time)
	if remainingTime <= 0 {
		log.Info("Instance expiraton detected", "instance", instance.Name)
		return 0, nil
	}

	return remainingTime, nil
}

// DeleteInstance attempts to delete the instance.
func (r *InstanceExpirationReconciler) DeleteInstance(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

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
func (r *InstanceExpirationReconciler) NotifyInstanceDeletion(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("notify-instance-deletion")
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	tenant := pkgcontext.TenantFrom(ctx)
	if tenant == nil {
		return fmt.Errorf("tenant not found in context")
	}

	// Send the notification email
	if r.EnableExpirationNotifications {
		if err := SendExpiringNotification(ctx, r.MailClient); err != nil {
			return fmt.Errorf("failed sending notification email: %w", err)
		}
	} else {
		log.Info("Expiration notifications are disabled, skipping email notification", "instance", instance.Name, "email", tenant.Spec.Email)
	}

	log.Info("Notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
	return nil
}

// ShouldSendWarningNotification checks if a warning notification should be sent for an expiring instance.
func (r *InstanceExpirationReconciler) ShouldSendWarningNotification(ctx context.Context) (bool, error) {
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return false, fmt.Errorf("instance not found in context")
	}

	// Se l'annotation è già presente, ritorna false
	if _, ok := instance.Annotations[forge.ExpiringWarningNotificationAnnotation]; ok {
		return false, nil
	}

	// Se non è presente, aggiungila e ritorna true
	patch := client.MergeFrom(instance.DeepCopy())
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}
	instance.Annotations[forge.ExpiringWarningNotificationAnnotation] = "true"
	if err := r.Client.Patch(ctx, instance, patch); err != nil {
		return false, fmt.Errorf("failed to patch instance with expiring warning notification annotation: %w", err)
	}
	return true, nil
}
