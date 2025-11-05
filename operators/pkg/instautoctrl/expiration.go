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
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
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
	EnableExpirationNotifications bool
	MailClient                    *mail.Client
	NotificationInterval          time.Duration
	MarginTime                    time.Duration
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
			createTemplateWatchHandlerWithTimeout(r.Client, func(t *clv1alpha2.Template) string { return t.Spec.DeleteAfter }),
			builder.WithPredicates(deleteAfterChanged),
		).
		Watches(&corev1.Namespace{},
			createNamespaceWatchHandlerWithIgnore(r.Client, forge.ExpirationIgnoreNamespace),
			builder.WithPredicates(expirationIgnoreNamespace),
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

	// Check if the reconciliation should be skipped based on the selector label and namespace labels.
	skip, err := r.CheckSkipReconciliation(ctx, req.Namespace)
	if skip {
		return ctrl.Result{}, err
	}

	instance, template, tenant, err := GetInstanceTemplateTenant(ctx, req, r.Client)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to retrieve instance/template/tenant")
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
		dbgLog.Info("Instance marked as never expiring", "instance", instance.GetName(), "namespace", instance.GetNamespace())
		return ctrl.Result{}, nil
	}

	remainingTime, err := r.CheckInstanceExpiration(ctx, deleteAfter)
	if err != nil {
		log.Error(err, "failed to check instance expiration")
		return ctrl.Result{}, err
	}

	tracer.Step("expiration checked")
	log.Info("Evaluate remaining time until expiration", remainingTime, instance.Name)

	if remainingTime <= 0 {
		tenant := pkgcontext.TenantFrom(ctx)
		if tenant == nil {
			return ctrl.Result{}, fmt.Errorf("tenant not found in context")
		}

		if r.EnableExpirationNotifications {
			shouldSendWarning, err := r.ShouldSendWarningNotification(ctx)
			if err != nil {
				log.Error(err, "failed to check if warning notification should be sent")
				return ctrl.Result{}, err
			}
			if shouldSendWarning {
				if err := SendExpiringWarningNotification(ctx, r.MailClient, r.NotificationInterval); err != nil {
					log.Error(err, "failed sending expiring warning notification email")
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: r.NotificationInterval}, nil
			}
			// If all notifications have been sent (or simply disabled), terminate the instance
			shouldTerminate, err := r.ShouldTerminateInstance(ctx)
			if err != nil {
				log.Error(err, "failed to check if instance should be terminated")
				return ctrl.Result{}, err
			}
			if shouldTerminate {
				if err := r.DeleteInstance(ctx); err != nil {
					log.Error(err, "failed to delete expired instance")
					return ctrl.Result{}, err
				}
				// Send notification for instance deletion
				if err := r.NotifyInstanceDeletion(ctx); err != nil {
					log.Error(err, "failed to send deletion notification")
					return ctrl.Result{}, err
				}
				tracer.Step("deletion notification sent")
			}
		} else {
			if err := r.DeleteInstance(ctx); err != nil {
				log.Error(err, "failed to delete expired instance")
				return ctrl.Result{}, err
			}
		}
	}
	// Calculate requeue time at the instance inactive deadline time:
	// if the instance is not yet to be terminated, we requeue it after the remaining time
	requeueTime := remainingTime
	// add margin time to the remaining time to avoid requeueing just before the deadline
	// avoiding a double requeue
	requeueTime += r.MarginTime
	dbgLog.Info("Remaining time before expiration", "remainingTime", remainingTime.String(), "instance", instance.Name, "namespace", instance.Namespace)

	dbgLog.Info("requeueing instance")
	return ctrl.Result{RequeueAfter: requeueTime}, nil
}

// ShouldTerminateInstance checks if the instance should be terminated.
func (r *InstanceExpirationReconciler) ShouldTerminateInstance(ctx context.Context) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("should-terminate-expiration")
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return false, fmt.Errorf("instance not found in context")
	}

	if r.EnableExpirationNotifications {
		if _, ok := instance.Annotations[forge.ExpiringWarningNotificationAnnotation]; !ok {
			return false, nil
		}
	}

	// Check if enough time is passed since the warning notification
	if timestampStr, ok := instance.Annotations[forge.ExpiringWarningNotificationTimestampAnnotation]; ok {
		timestampWarning, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return false, fmt.Errorf("failed to parse expiring warning notification timestamp: %w", err)
		}

		if time.Since(timestampWarning) < r.NotificationInterval {
			log.Info("Not enough time passed since warning notification, skipping termination for expiration", "instance", instance.Name)
			return false, nil
		}
	}

	return true, nil
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
		log.Info("Instance expiration detected", "instance", instance.Name)
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

	if err := r.Delete(ctx, instance); err != nil {
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
		log.Info("Notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
	} else {
		log.Info("Expiration notifications are disabled, skipping email notification", "instance", instance.Name, "email", tenant.Spec.Email)
	}

	return nil
}

// ShouldSendWarningNotification checks if a warning notification should be sent for an expiring instance.
func (r *InstanceExpirationReconciler) ShouldSendWarningNotification(ctx context.Context) (bool, error) {
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return false, fmt.Errorf("instance not found in context")
	}

	// If annotation already existing, return false
	if _, ok := instance.Annotations[forge.ExpiringWarningNotificationAnnotation]; ok {
		return false, nil
	}

	// If not present, add it and return true
	patch := client.MergeFrom(instance.DeepCopy())
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}
	instance.Annotations[forge.ExpiringWarningNotificationAnnotation] = "true"
	instance.Annotations[forge.ExpiringWarningNotificationTimestampAnnotation] = time.Now().Format(time.RFC3339)

	if err := r.Patch(ctx, instance, patch); err != nil {
		return false, fmt.Errorf("failed to patch instance with expiring warning notification annotation: %w", err)
	}
	return true, nil
}

// CheckSkipReconciliation checks if the reconciliation should be skipped based on the selector label and namespace labels.
func (r *InstanceExpirationReconciler) CheckSkipReconciliation(ctx context.Context, namespace string) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-skip-reconciliation-expiration")
	// Check the selector label, in order to know whether to perform or not reconciliation.
	if proceed, err := utils.CheckSelectorLabel(ctx, r.Client, namespace, r.NamespaceWhitelist.MatchLabels); !proceed {
		if err != nil {
			err = fmt.Errorf("failed checking selector label: %w", err)
		}
		return true, err
	}

	var namespaceObj corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: namespace}, &namespaceObj); err != nil {
		log.Error(err, "failed retrieving namespace", "namespace", namespace)
		return true, err
	}

	// check the namespace labels, in order to know whether to perform or not reconciliation on a specific namespace.
	if stop := utils.CheckSingleLabel(&namespaceObj, forge.ExpirationIgnoreNamespace, strconv.FormatBool(true)); stop {
		log.Info("label present, skipping expiration reconciliation for namespace", "namespace", namespace, "label", forge.ExpirationIgnoreNamespace)
		return true, nil
	}

	return false, nil
}
