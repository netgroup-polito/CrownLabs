// Copyright 2020-2026 Politecnico di Torino
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
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/clcontext"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/mail"
)

// InstanceInactiveTerminationReconciler watches for instances to be terminated.
type InstanceInactiveTerminationReconciler struct {
	client.Client
	EventsRecorder                  record.EventRecorder
	Scheme                          *runtime.Scheme
	NamespaceWhitelist              metav1.LabelSelector
	StatusCheckRequestTimeout       time.Duration
	InstanceMaxNumberOfAlerts       int
	EnableInactivityNotifications   bool
	NotificationInterval            time.Duration
	DestructionNotificationInterval time.Duration
	MailClient                      *mail.Client
	Prometheus                      PrometheusClientInterface
	MarginTime                      time.Duration
	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// SetupWithManager registers a new controller for InstanceTerminationReconciler resources.
// The controller is configured to watch for Instance resources and Template resources.
// For the instance resources, it is configured to only reconcile instances at the creation time (to calculate the expiration time) and at the deletion time. Updates on the instance resources are ignored by this reconciler.
// For the template resources, it is configured to reconcile instances when the template's inactivtyTimeout field is changed. In this case, it will enqueue all the instances that are associated with that template.
// To avoid unnecessary reconciliations, the controller avoid reconciling instances whose template's inactivtyTimeout field is set to neverTimeoutValue, which means that the instance will never be deleted.
func (r *InstanceInactiveTerminationReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Instance{},
			builder.WithPredicates(instanceTriggered)).
		Watches(
			&clv1alpha2.Template{},
			createTemplateWatchHandlerWithTimeout(r.Client, func(t *clv1alpha2.Template) string {
				if t.Spec.InactivityTimeout != NeverTimeoutValue {
					return t.Spec.InactivityTimeout
				}
				return t.Spec.DestroyAfterInactivity
			}),
			builder.WithPredicates(inactivityTimeoutChanged),
		).
		Watches(&corev1.Namespace{},
			createNamespaceWatchHandlerWithIgnore(r.Client, forge.InstanceInactivityIgnoreNamespace),
			builder.WithPredicates(inactivityIgnoreNamespace),
		).
		Named("instance-inactive-termination").
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "InstanceInactiveTermination")).
		Complete(r)
}

// Reconcile reconciles the status of the InstanceSnapshot resource.
func (r *InstanceInactiveTerminationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	// Add the instance, template and tenant to the context
	ctx, _ = clctx.InstanceInto(ctx, instance)
	ctx, _ = clctx.TemplateInto(ctx, template)
	ctx, _ = clctx.TenantInto(ctx, tenant)

	// Setup instance annotations
	if err := r.SetupInstanceAnnotations(ctx); err != nil {
		log.Error(err, "failed setting up instance annotations")
		return ctrl.Result{}, err
	}

	// Verify whether the instance annotations need to be reset, and reset them if necessary.
	if err := r.ResetAnnotations(ctx); err != nil {
		log.Error(err, "failed resetting instance annotations")
		return ctrl.Result{}, err
	}

	// Checks if the instance is running, if not, we start the countdown for destruction for persistent instances.
	if !instance.Spec.Running {
		return r.handlePoweredOffInstance(ctx, instance, tracer)
	}

	inactivityTimeout := template.Spec.InactivityTimeout
	// If set to neverTimeoutValue, return without rescheduling
	if inactivityTimeout == NeverTimeoutValue {
		dbgLog.Info("Instance marked as never stop", "name", instance.GetName(), "namespace", instance.GetNamespace())
		return ctrl.Result{}, nil
	}

	inactivityTimeoutDuration, err := ParseDurationWithDays(ctx, inactivityTimeout)
	if err != nil {
		log.Error(err, "failed to parse deleteAfter duration")
		return ctrl.Result{}, fmt.Errorf("failed to parse inactivityTimeout duration %s: %w", inactivityTimeout, err)
	}

	tracer.Step("labels checked")

	// Update the last login time of the instance based on the Prometheus data
	if err := r.UpdateInstanceLastLogin(ctx, inactivityTimeoutDuration); err != nil {
		log.Error(err, "failed updating last login time of the instance")
		return ctrl.Result{RequeueAfter: r.NotificationInterval}, err
	}

	tracer.Step("instance last login updated")

	remainingTime, err := r.GetRemainingInactivityTime(ctx, inactivityTimeoutDuration)
	if err != nil {
		log.Error(err, "failed checking instance termination")
		return ctrl.Result{}, err
	}

	dbgLog.Info("instance termination check", "remainingTime", remainingTime.String(), "instance", instance.Name)
	tracer.Step("Inactive termination check done")

	// Check if the instance has expired
	if remainingTime <= 0 {
		res, terminateEarly, err := r.handleInactivityInstance(ctx, instance)
		if terminateEarly || err != nil {
			return res, err
		}
	}

	tracer.Step("Inactive termination done")

	// Calculate requeue time at the instance inactive deadline time: if the instance is not yet to be terminated, we requeue it after the remaining time
	// Let's add margin time to the remaining time to avoid requeueing just before the deadline, avoiding a double requeue
	requeueTime := remainingTime + r.MarginTime
	dbgLog.Info("requeueing instance")
	return ctrl.Result{RequeueAfter: requeueTime}, nil
}

// handlePoweredOffInstance manages the inactivity lifecycle for instances that are already powered off.
func (r *InstanceInactiveTerminationReconciler) handlePoweredOffInstance(ctx context.Context, instance *clv1alpha2.Instance, tracer *trace.Trace) (res ctrl.Result, err error) {
	log := ctrl.LoggerFrom(ctx)
	dbgLog := log.V(utils.LogDebugLevel)

	remainingPauseTime, isActive, err := r.GetRemainingInactivityDestructionTime(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	if isActive {
		if remainingPauseTime <= 0 {
			if r.EnableInactivityNotifications {
				// Email logic to inform the user that the instance will be destroyed
				shouldSendWarning, err := r.ShouldSendDestructionWarningNotification(ctx, instance)
				if err != nil {
					log.Error(err, "failed checking if should send destruction warning notification")
					return ctrl.Result{}, err
				}
				if shouldSendWarning {
					window, err := r.GetDestructionNotificationWindow(ctx, instance)
					if err != nil {
						log.Error(err, "failed getting destruction notification window")
						return ctrl.Result{}, err
					}
					// Send the email notification and requeue after the interval
					if err := r.SendDestructionWarning(ctx, instance, window); err != nil {
						log.Error(err, "failed sending destruction warning email")
						return ctrl.Result{}, err
					}
					return ctrl.Result{RequeueAfter: r.DestructionNotificationInterval}, nil
				}
				shouldDelete, err := r.ShouldDeleteInstance(ctx, instance)
				if err != nil {
					log.Error(err, "failed checking if should delete instance")
					return ctrl.Result{}, err
				}
				if !shouldDelete {
					// We have not sent all the emails, we are just waiting for the next interval.
					// Calculate how much time is left and requeue.
					lastNotificationTimeStr := instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation]
					lastNotificationTime, _ := time.Parse(time.RFC3339, lastNotificationTimeStr)
					requeueTime := r.DestructionNotificationInterval - time.Since(lastNotificationTime) + r.MarginTime
					if requeueTime < 0 {
						requeueTime = r.MarginTime
					}

					dbgLog.Info("requeueing paused instance to wait for next destruction notification interval")
					return ctrl.Result{RequeueAfter: requeueTime}, nil
				}

				// If all the emails have been sent, we delete the instance
				log.Info("Deleting paused persistent instance due to prolonged inactivity...")
				if err := r.DeleteInstance(ctx); err != nil {
					log.Error(err, "failed to delete inactive instance")
					return ctrl.Result{}, err
				}
				// Send notification for instance deletion
				if err := r.NotifyInstanceDeletion(ctx); err != nil {
					log.Error(err, "failed to send deletion notification")
					return ctrl.Result{}, err
				}
				if tracer != nil {
					tracer.Step("instance deleted")
				}
				return ctrl.Result{}, nil
			}

			// If notifications are disabled, we delete the instance immediately
			log.Info("Deleting paused persistent instance due to prolonged inactivity...")
			if err := r.DeleteInstance(ctx); err != nil {
				log.Error(err, "failed to delete inactive instance")
				return ctrl.Result{}, err
			}
			if tracer != nil {
				tracer.Step("instance deleted")
			}
			return ctrl.Result{}, nil
		}

		// Requeue based on the remaining time for the destruction
		dbgLog.Info("requeueing paused instance for destruction check")
		return ctrl.Result{RequeueAfter: remainingPauseTime + r.MarginTime}, nil
	}

	// Early return to avoid executing the normal inactivity logic for a machine that is already powered off
	return ctrl.Result{}, nil
}

// handleInactivityInstance processes the instance when its inactivity timeout has been reached.
func (r *InstanceInactiveTerminationReconciler) handleInactivityInstance(ctx context.Context, instance *clv1alpha2.Instance) (res ctrl.Result, terminateEarly bool, err error) {
	log := ctrl.LoggerFrom(ctx)
	if r.EnableInactivityNotifications {
		// Check if all notifications have already been sent
		shouldSendWarning, err := r.ShouldSendWarningNotification(ctx, instance)
		if err != nil {
			log.Error(err, "failed checking if should send notification")
			return ctrl.Result{}, true, err
		}

		if shouldSendWarning {
			if err := r.SendInactivityWarning(ctx, instance); err != nil {
				log.Error(err, "failed sending inactivity warning email", "instance", instance.Name, "namespace", instance.Namespace)
				return ctrl.Result{}, true, err
			}
			return ctrl.Result{RequeueAfter: r.NotificationInterval}, true, nil
		}
		// If all notifications have been sent (or simply disabled), terminate the instance
		shouldTerminate, err := r.ShouldTerminateInstance(ctx, instance)
		if err != nil {
			log.Error(err, "failed checking if should terminate instance", "instance", instance.Name, "namespace", instance.Namespace)
			return ctrl.Result{}, true, err
		}
		if shouldTerminate {
			if err := r.TerminateInstance(ctx); err != nil {
				log.Error(err, "failed terminating inactive instance", "instance", instance.Name, "namespace", instance.Namespace)
				return ctrl.Result{}, true, err
			}
			log.Info("Inactive instance has been paused/deleted", "instance", instance.Name, "namespace", instance.Namespace)
			if err := r.SendTerminationNotification(ctx); err != nil {
				log.Error(err, "failed sending termination notification email", "instance", instance.Name, "namespace", instance.Namespace)
				return ctrl.Result{}, true, err
			}
			return ctrl.Result{}, true, nil
		}
	} else {
		// If notifications are disabled, terminate the instance immediately
		if err := r.TerminateInstance(ctx); err != nil {
			log.Error(err, "failed terminating inactive instance", "instance", instance.Name, "namespace", instance.Namespace)
			return ctrl.Result{}, true, err
		}
	}
	return ctrl.Result{}, false, nil
}

// UpdateInstanceLastLogin updates the last login time of the instance in the annotations.
func (r *InstanceInactiveTerminationReconciler) UpdateInstanceLastLogin(ctx context.Context, inactivityTimeoutDuration time.Duration) error {
	log := ctrl.LoggerFrom(ctx).WithName("update-instance-last-login")
	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	// Check Prometheus health first
	healthy, err := r.Prometheus.IsPrometheusHealthy(ctx, r.StatusCheckRequestTimeout)
	if err != nil || !healthy {
		log.Error(err, "Prometheus is not healthy")
		return err
	}

	// Get instance activity data
	queryNginx := fmt.Sprintf(r.Prometheus.GetQueryNginxData(), instance.Namespace, instance.Name)
	lastActivityTimeNginx, errNginx := r.Prometheus.GetLastActivityTime(queryNginx, inactivityTimeoutDuration)

	queryWebSSH := fmt.Sprintf(r.Prometheus.GetQueryWebSSHData(), instance.Namespace, instance.Name)
	lastActivityTimeWebSSH, errWebSSH := r.Prometheus.GetLastActivityTime(queryWebSSH, inactivityTimeoutDuration)

	// Aggregate SSH activity times across all environments (find minimum non-zero)
	var lastActivityTimeSSH time.Time
	lastActivityTimeSSHFound := false
	for envIdx := range instance.Status.Environments {
		env := &instance.Status.Environments[envIdx]
		querySSH := fmt.Sprintf(r.Prometheus.GetQuerySSHData(), env.IP)
		envActivityTime, errSSH := r.Prometheus.GetLastActivityTime(querySSH, inactivityTimeoutDuration)
		if errSSH == nil && !envActivityTime.IsZero() {
			if !lastActivityTimeSSHFound || envActivityTime.Before(lastActivityTimeSSH) {
				lastActivityTimeSSH = envActivityTime
				lastActivityTimeSSHFound = true
			}
		}
	}

	// If all queries failed, return error
	if errNginx != nil && !lastActivityTimeSSHFound && errWebSSH != nil {
		return fmt.Errorf("failed retrieving last activity time from all queries: %w", errNginx)
	}
	if lastActivityTimeNginx.IsZero() && !lastActivityTimeSSHFound && lastActivityTimeWebSSH.IsZero() {
		log.Info("No activity detected for the instance", "instance", instance.Name, "namespace", instance.Namespace)
		return nil // No activity detected, do not update the last activity time
	}

	var maxLastActivityTime time.Time
	maxLastActivityTime = lastActivityTimeNginx
	if lastActivityTimeSSH.After(maxLastActivityTime) {
		maxLastActivityTime = lastActivityTimeSSH
	}
	if lastActivityTimeWebSSH.After(maxLastActivityTime) {
		maxLastActivityTime = lastActivityTimeWebSSH
	}

	// patch the instance with the new last activity time
	patch := client.MergeFrom(instance.DeepCopy())
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}
	instance.Annotations[forge.LastActivityAnnotation] = maxLastActivityTime.Format(time.RFC3339)
	instance.Annotations[forge.AlertAnnotationNum] = "0"
	if err := r.Patch(ctx, instance, patch); err != nil {
		return err
	}

	return nil
}

// GetRemainingInactivityTime checks if the Instance has to be terminated.
func (r *InstanceInactiveTerminationReconciler) GetRemainingInactivityTime(ctx context.Context, inactivityTimeoutDuration time.Duration) (time.Duration, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-instance-termination")
	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return 0, fmt.Errorf("instance not found in context")
	}
	var remainingTime time.Duration

	lastLogin, err := time.Parse(time.RFC3339, instance.Annotations[forge.LastActivityAnnotation])
	if err != nil {
		log.Error(err, "failed parsing LastLogin time")
		return 0, err
	}

	// Check if the instance has been inactive for longer than the timeout duration
	remainingTime = inactivityTimeoutDuration - time.Since(lastLogin)
	if remainingTime <= 0 {
		log.Info("Instance inactivity detected", "instance", instance.Name)
		return 0, nil
	}

	return remainingTime, nil
}

// GetInactivityNotificationWindow calculates the remaining time available for sending inactivity notifications to the given instance, based on the maximum allowed number of notifications and those already sent.
func (r *InstanceInactiveTerminationReconciler) GetInactivityNotificationWindow(ctx context.Context, instance *clv1alpha2.Instance) (time.Duration, error) {
	log := ctrl.LoggerFrom(ctx).WithName("GetInactivityNotificationWindow")

	template := clctx.TemplateFrom(ctx)

	// Calculate the remaining number of alerts that should be sent
	NumAlerts := r.InstanceMaxNumberOfAlerts

	if template != nil {
		if customMaxAlertsStr, ok := template.Annotations[forge.CustomNumberOfAlertsAnnotation]; ok {
			customMaxAlerts, err := strconv.Atoi(customMaxAlertsStr)
			if err == nil {
				NumAlerts = customMaxAlerts
			}
		}
	}

	numAlertsSent, err := strconv.Atoi(instance.Annotations[forge.AlertAnnotationNum])
	if err != nil {
		log.Error(err, "failed converting string of alerts sent in int number", "annotation", instance.Annotations[forge.AlertAnnotationNum])
		return 0, err
	}

	remainingAlerts := NumAlerts - numAlertsSent
	if remainingAlerts <= 0 {
		return 0, nil
	}

	// Calculate the remaining time before reaching the maximum number of alerts
	return time.Duration(remainingAlerts) * r.NotificationInterval, nil
}

// IsTemplatePersistent checks if the instance template has at least one persistent environment.
func IsTemplatePersistent(template *clv1alpha2.Template) bool {
	if template == nil || template.Spec.EnvironmentList == nil {
		return false
	}

	// Check if any environment in the template is persistent
	for i := range template.Spec.EnvironmentList {
		env := &template.Spec.EnvironmentList[i]
		if env.Persistent {
			return true
		}
	}
	return false
}

// TerminateInstance terminates the Instance.
func (r *InstanceInactiveTerminationReconciler) TerminateInstance(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("termination")

	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}
	template := clctx.TemplateFrom(ctx)
	if template == nil {
		return fmt.Errorf("template not found in context")
	}

	log.Info("Terminating instance", "instance", instance.Name, " in namespace", instance.Namespace)

	if IsTemplatePersistent(template) {
		log.Info("Stopping persistent instance...")
		instance.Spec.Running = false
		// Update the last running annotation
		currentRunningStr := strconv.FormatBool(instance.Spec.Running)
		lastRunningStr, ok := instance.Annotations[forge.LastRunningAnnotation]
		if !ok || lastRunningStr != currentRunningStr {
			instance.Annotations[forge.LastRunningAnnotation] = currentRunningStr
		}

		return r.Update(ctx, instance)
	}
	log.Info("Deleting non-persistent instance...")
	return r.Delete(ctx, instance)
}

// IncrementAnnotation increments the value of the annotation string by 1.
func (r *InstanceInactiveTerminationReconciler) IncrementAnnotation(ctx context.Context, annotationString string) (string, error) {
	log := ctrl.LoggerFrom(ctx).WithName("string-to-int-annotation")
	log.Info("converting string to int annotation", "annotation", annotationString)

	annotationInt, err := strconv.Atoi(annotationString)
	if err != nil {
		log.Error(err, "failed converting string to int")
		return "0", fmt.Errorf("failed converting string to int: %w", err)
	}
	annotationInt++
	log.Info("incrementing annotation", "annotation", annotationInt)

	annotationString = strconv.Itoa(annotationInt)
	log.Info("converting int to string updated annotation", "annotation", annotationString)
	return annotationString, nil
}

// SetupInstanceAnnotations sets up the annotations for the instance.
func (r *InstanceInactiveTerminationReconciler) SetupInstanceAnnotations(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("setup-instance-annotations")

	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	original := instance.DeepCopy()
	// add annotations to the instance if not present
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}

	updated := false

	// Check and set the alert annotation if not present
	if _, ok := instance.Annotations[forge.AlertAnnotationNum]; !ok {
		log.Info("adding alert number annotation to instance for the first time", "annotation", forge.AlertAnnotationNum)
		instance.Annotations[forge.AlertAnnotationNum] = "0"
		updated = true
	}

	// Check and set the last activity annotation if not present
	if _, ok := instance.Annotations[forge.LastActivityAnnotation]; !ok {
		log.Info("adding last activity annotation to instance for the first time", "annotation", forge.LastActivityAnnotation)
		instance.Annotations[forge.LastActivityAnnotation] = time.Now().Format(time.RFC3339)
		updated = true
	}

	// Check and set the last notification time annotation if not present
	if _, ok := instance.Annotations[forge.LastNotificationTimestampAnnotation]; !ok {
		log.Info("adding last notification time annotation to instance for the first time", "annotation", forge.LastNotificationTimestampAnnotation)
		instance.Annotations[forge.LastNotificationTimestampAnnotation] = ""
		updated = true
	}

	// Check and set the destruction alert annotation if not present
	if _, ok := instance.Annotations[forge.DestructionAlertsSentAnnotation]; !ok {
		log.Info("adding destruction alert number annotation to instance for the first time", "annotation", forge.DestructionAlertsSentAnnotation)
		instance.Annotations[forge.DestructionAlertsSentAnnotation] = "0"
		updated = true
	}

	// Check and set the last destruction notification time annotation if not present
	if _, ok := instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation]; !ok {
		log.Info("adding last destruction notification time annotation to instance for the first time", "annotation", forge.LastDestructionNotificationTimestampAnnotation)
		instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation] = ""
		updated = true
	}

	// Check and set the last powered off timestamp annotation if not present
	if _, ok := instance.Annotations[forge.LastPoweredOffTimestampAnnotation]; !ok {
		log.Info("adding last powered off timestamp annotation to instance for the first time", "annotation", forge.LastPoweredOffTimestampAnnotation)
		instance.Annotations[forge.LastPoweredOffTimestampAnnotation] = ""
		updated = true
	}

	// Check and set the destroy-after-inactivity annotation from the template
	template := clctx.TemplateFrom(ctx)
	if template != nil {
		if val, ok := instance.Annotations["crownlabs.polito.it/destroy-after-inactivity"]; !ok || val != template.Spec.DestroyAfterInactivity {
			log.Info("updating destroy-after-inactivity annotation", "annotation", "crownlabs.polito.it/destroy-after-inactivity", "value", template.Spec.DestroyAfterInactivity)
			instance.Annotations["crownlabs.polito.it/destroy-after-inactivity"] = template.Spec.DestroyAfterInactivity
			updated = true
		}
	}

	// Apply the patch only if something changed
	if updated {
		patch := client.MergeFrom(original)
		if err := r.Patch(ctx, instance, patch); err != nil {
			log.Error(err, "failed updating instance annotations")
			return err
		}
	}

	log.Info("instance annotations setup completed", "instance", instance.Name)
	return nil
}

// CheckSkipReconciliation checks if the reconciliation should be skipped based on the selector label and namespace labels.
func (r *InstanceInactiveTerminationReconciler) CheckSkipReconciliation(ctx context.Context, namespace string) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-skip-reconciliation-inactivity")

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
	if stop := utils.CheckSingleLabel(&namespaceObj, forge.InstanceInactivityIgnoreNamespace, strconv.FormatBool(true)); stop {
		log.Info("label present, skipping inactivity reconciliation for namespace", "namespace", namespace, "label", forge.InstanceInactivityIgnoreNamespace)
		return true, nil
	}

	return false, nil
}

// getAlertCounts returns the current number of alerts sent and the maximum allowed alerts for the instance.
func (r *InstanceInactiveTerminationReconciler) getAlertCounts(ctx context.Context, instance *clv1alpha2.Instance) (numAlerts, maxAlerts int, err error) {
	log := ctrl.LoggerFrom(ctx).WithName("getAlertCounts")

	numAlerts, err = strconv.Atoi(instance.Annotations[forge.AlertAnnotationNum])
	if err != nil {
		log.Error(err, "failed converting string of alerts sent in int number", "annotation", instance.Annotations[forge.AlertAnnotationNum])
		return 0, 0, err
	}

	maxAlerts = r.InstanceMaxNumberOfAlerts
	template := clctx.TemplateFrom(ctx)
	if template != nil {
		// if the CustomNumberOfAlertsAnnotation is set, override the default max alerts
		if customMaxAlertsStr, ok := template.Annotations[forge.CustomNumberOfAlertsAnnotation]; ok {
			customMaxAlerts, err := strconv.Atoi(customMaxAlertsStr)
			if err != nil {
				log.Error(err, "failed converting custom max alerts annotation to int, using default value", "annotation", customMaxAlertsStr)
			} else {
				maxAlerts = customMaxAlerts
			}
		}
	}
	return numAlerts, maxAlerts, nil
}

// ShouldTerminateInstance checks if the instance should be terminated based on its running state and the number of alerts sent.
func (r *InstanceInactiveTerminationReconciler) ShouldTerminateInstance(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	if !instance.Spec.Running {
		return false, nil
	}

	// If notifications are enabled, terminate the instance only if the maximum number of alerts has been sent
	if r.EnableInactivityNotifications {
		numAlerts, maxAlerts, err := r.getAlertCounts(ctx, instance)
		if err != nil {
			return false, err
		}
		return numAlerts >= maxAlerts, nil
	}

	// If notifications are disabled, terminate the instance immediately
	return true, nil
}

// ShouldSendWarningNotification checks if the notification should be sent based on the number of alerts sent and the last notification time.
func (r *InstanceInactiveTerminationReconciler) ShouldSendWarningNotification(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("ShouldSendWarningNotification")

	if !instance.Spec.Running {
		return false, nil // If the instance is not running, do not send a notification
	}

	if !r.EnableInactivityNotifications {
		log.Info("Inactivity notifications are disabled, skipping email notification", "instance", instance.Name)
		return false, nil
	}

	numAlerts, maxAlerts, err := r.getAlertCounts(ctx, instance)
	if err != nil {
		return false, err
	}

	lastNotificationTimeStr, ok := instance.Annotations[forge.LastNotificationTimestampAnnotation]
	if !ok {
		log.Info("Last notification time annotation not found, sending notification", "instance", instance.Name)
		return true, nil
	}

	// if this is the first notification, the annotation is still empty, therefore we can send a notification
	if lastNotificationTimeStr == "" {
		return true, nil
	}
	lastNotificationTime, err := time.Parse(time.RFC3339, lastNotificationTimeStr)
	if err != nil {
		log.Error(err, "failed parsing last notification time", "lastNotificationTime", lastNotificationTimeStr)
		return false, err
	}
	if numAlerts > 0 {
		if time.Since(lastNotificationTime) < r.NotificationInterval-r.MarginTime {
			log.Info("Last notification sent within the notification interval, skipping email notification", "instance", instance.Name)
			return false, nil
		}
	}
	return numAlerts < maxAlerts, nil
}

// SendInactivityWarning sends an inactivity warning email to the user and updates the instance annotations.
func (r *InstanceInactiveTerminationReconciler) SendInactivityWarning(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx)
	tenant := clctx.TenantFrom(ctx)
	if tenant == nil {
		return fmt.Errorf("tenant not found in context")
	}

	// Calculate the remaining time available for sending inactivity notifications
	remainingTime, err := r.GetInactivityNotificationWindow(ctx, instance)
	if err != nil {
		log.Error(err, "failed calculating remaining time for inactivity notifications")
		return err
	}

	if r.EnableInactivityNotifications {
		if err := SendInactivityDetectionNotification(ctx, r.MailClient, remainingTime); err != nil {
			log.Error(err, "failed sending notification email to user", "email", tenant.Spec.Email)
			return err
		}
		log.Info("Inactivity notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
	} else {
		log.Info("Inactivity notifications are disabled, skipping email notification", "instance", instance.Name, "email", tenant.Spec.Email)
	}

	newNumberOfAlerts, err := r.IncrementAnnotation(ctx, instance.Annotations[forge.AlertAnnotationNum])
	if err != nil {
		log.Error(err, "failed incrementing annotation")
		return err
	}

	patch := client.MergeFrom(instance.DeepCopy())
	instance.Annotations[forge.AlertAnnotationNum] = newNumberOfAlerts
	instance.Annotations[forge.LastNotificationTimestampAnnotation] = time.Now().Format(time.RFC3339)
	if err := r.Patch(ctx, instance, patch); err != nil {
		log.Error(err, "failed updating instance annotations")
		return err
	}

	return nil
}

// SendTerminationNotification handles sending notification emails when an instance is deleted.
func (r *InstanceInactiveTerminationReconciler) SendTerminationNotification(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("send-termination-notification")
	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	tenant := clctx.TenantFrom(ctx)
	if tenant == nil {
		return fmt.Errorf("tenant not found in context")
	}

	if r.EnableInactivityNotifications {
		if err := SendInactivityTerminationNotification(ctx, r.MailClient, 0); err != nil {
			return fmt.Errorf("failed sending termination notification email: %w", err)
		}
		log.Info("Termination notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
	} else {
		log.Info("Inactivity notifications are disabled, skipping email notification", "instance", instance.Name, "email", tenant.Spec.Email)
	}

	return nil
}

// ResetAnnotations resets some instance annotations (such as the number of alerts sent or the last activity field) when the instance Running state changes from false to true.
func (r *InstanceInactiveTerminationReconciler) ResetAnnotations(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("reset-annotation")

	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}
	original := instance.DeepCopy()
	updated := false

	lastRunningStr := instance.Annotations[forge.LastRunningAnnotation]
	lastRunning := false
	if lastRunningStr != "" {
		if val, err := strconv.ParseBool(lastRunningStr); err == nil {
			lastRunning = val
		}
	}

	// Reset if "Running" changed from false to true
	if instance.Spec.Running && !lastRunning {
		log.Info("Detected transition from false to true: resetting alert counter and last activity field")
		instance.Annotations[forge.AlertAnnotationNum] = "0"
		instance.Annotations[forge.LastActivityAnnotation] = time.Now().Format(time.RFC3339)

		// Reset the destruction mail counter
		instance.Annotations[forge.DestructionAlertsSentAnnotation] = "0"
		instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation] = ""

		updated = true
	}
	// update the LastRunningAnnotation
	currentRunningStr := strconv.FormatBool(instance.Spec.Running)
	if lastRunningStr != currentRunningStr {
		instance.Annotations[forge.LastRunningAnnotation] = currentRunningStr
		updated = true
	}

	if updated {
		patch := client.MergeFrom(original)
		if err := r.Patch(ctx, instance, patch); err != nil {
			log.Error(err, "failed updating instance annotations")
			return err
		}
	}

	return nil
}

// GetDestructionNotificationWindow the remaining time available for sending inactivity destruction notifications to the given instance, based on the maximum allowed number of notifications and those already sent.
func (r *InstanceInactiveTerminationReconciler) GetDestructionNotificationWindow(ctx context.Context, instance *clv1alpha2.Instance) (time.Duration, error) {
	log := ctrl.LoggerFrom(ctx).WithName("GetDestructionNotificationWindow")

	numAlertsStr := instance.Annotations[forge.DestructionAlertsSentAnnotation]
	numAlerts := 0
	if numAlertsStr != "" {
		var err error
		numAlerts, err = strconv.Atoi(numAlertsStr)
		if err != nil {
			log.Error(err, "failed converting string of destruction alerts sent in int number", "annotation", numAlertsStr)
			return 0, err
		}
	}

	maxAlerts := r.InstanceMaxNumberOfAlerts
	remainingAlerts := maxAlerts - numAlerts
	if remainingAlerts <= 0 {
		return 0, nil
	}
	return time.Duration(remainingAlerts) * r.DestructionNotificationInterval, nil
}

// ShouldSendDestructionWarningNotification checks if the notification should be sent based on the number of alerts sent and the last notification time.
func (r *InstanceInactiveTerminationReconciler) ShouldSendDestructionWarningNotification(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("ShouldSendDestructionWarningNotification")

	numAlertsStr := instance.Annotations[forge.DestructionAlertsSentAnnotation]
	numAlerts := 0
	if numAlertsStr != "" {
		var err error
		numAlerts, err = strconv.Atoi(numAlertsStr)
		if err != nil {
			log.Error(err, "failed converting string of destruction alerts sent in int number", "annotation", numAlertsStr)
			return false, err
		}
	}

	maxAlerts := r.InstanceMaxNumberOfAlerts

	lastNotificationTimeStr := instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation]
	if lastNotificationTimeStr == "" {
		log.Info("Last destruction notification time annotation not found or empty, sending notification", "instance", instance.Name)
		return true, nil // First email
	}

	lastNotificationTime, err := time.Parse(time.RFC3339, lastNotificationTimeStr)
	if err != nil {
		log.Error(err, "failed parsing last destruction notification time", "lastNotificationTime", lastNotificationTimeStr)
		return false, err
	}

	if numAlerts > 0 && time.Since(lastNotificationTime) < r.DestructionNotificationInterval-r.MarginTime {
		log.Info("Last destruction notification sent within the notification interval, skipping email notification", "instance", instance.Name)
		return false, nil // The interval has not yet passed
	}
	return numAlerts < maxAlerts, nil
}

// SendDestructionWarning sends the destruction warning email to the user and updates the instance annotations.
func (r *InstanceInactiveTerminationReconciler) SendDestructionWarning(ctx context.Context, instance *clv1alpha2.Instance, remainingTime time.Duration) error {
	log := ctrl.LoggerFrom(ctx).WithName("SendDestructionWarning")
	tenant := clctx.TenantFrom(ctx)
	if tenant == nil {
		return fmt.Errorf("tenant not found in context")
	}

	// 1. Call the function to send the email that is in common.go.
	if err := SendDestructionWarningNotification(ctx, r.MailClient, remainingTime); err != nil {
		log.Error(err, "failed sending destruction notification email to user", "email", tenant.Spec.Email)
		return fmt.Errorf("failed to send destruction warning email: %w", err)
	}
	log.Info("Destruction notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)

	// 2. Update the annotations to count how many emails we have sent.
	numAlertsStr := instance.Annotations[forge.DestructionAlertsSentAnnotation]
	numAlerts := 0
	if numAlertsStr != "" {
		var err error
		numAlerts, err = strconv.Atoi(numAlertsStr)
		if err != nil {
			log.Error(err, "failed converting string to int")
			return err
		}
	}

	patch := client.MergeFrom(instance.DeepCopy())
	instance.Annotations[forge.DestructionAlertsSentAnnotation] = strconv.Itoa(numAlerts + 1)
	instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation] = time.Now().Format(time.RFC3339)
	if err := r.Patch(ctx, instance, patch); err != nil {
		log.Error(err, "failed updating instance annotations")
		return err
	}

	return nil
}

// GetRemainingInactivityDestructionTime checks the remaining time before the instance is destroyed due to prolonged inactivity while powered off.
func (r *InstanceInactiveTerminationReconciler) GetRemainingInactivityDestructionTime(ctx context.Context, instance *clv1alpha2.Instance) (time.Duration, bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-instance-destruction")
	template := clctx.TemplateFrom(ctx)
	if template == nil {
		return 0, false, fmt.Errorf("template not found in context")
	}

	destroyAfterInactivity := template.Spec.DestroyAfterInactivity
	if destroyAfterInactivity == NeverTimeoutValue || destroyAfterInactivity == "" {
		return 0, false, nil
	}

	destroyAfterInactivityDuration, err := ParseDurationWithDays(ctx, destroyAfterInactivity)
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse destroyAfterInactivity duration %s: %w", destroyAfterInactivity, err)
	}
	poweredOffTimeStr := instance.Annotations[forge.LastPoweredOffTimestampAnnotation]
	if poweredOffTimeStr == "" {
		return 0, false, nil // No powered-off timestamp, nothing to calculate
	}

	// Store the destroyAfterInactivity value as an annotation for validation/visibility
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}
	instance.Annotations["crownlabs.polito.it/destroy-after-inactivity"] = destroyAfterInactivity

	poweredOffTime, err := time.Parse(time.RFC3339, poweredOffTimeStr)
	if err != nil {
		log.Error(err, "failed to parse last powered off time", "timestamp", poweredOffTimeStr)
		return 0, false, err
	}

	remainingTime := destroyAfterInactivityDuration - time.Since(poweredOffTime)
	return remainingTime, true, nil
}

// ShouldDeleteInstance checks if the instance should be deleted based on the number of destruction alerts sent.
func (r *InstanceInactiveTerminationReconciler) ShouldDeleteInstance(_ context.Context, instance *clv1alpha2.Instance) (bool, error) {
	if r.EnableInactivityNotifications {
		numAlertsStr := instance.Annotations[forge.DestructionAlertsSentAnnotation]
		numAlerts := 0
		if numAlertsStr != "" {
			var err error
			numAlerts, err = strconv.Atoi(numAlertsStr)
			if err != nil {
				return false, err
			}
		}
		return numAlerts >= r.InstanceMaxNumberOfAlerts, nil
	}
	return true, nil
}

// DeleteInstance attempts to delete the instance.
func (r *InstanceInactiveTerminationReconciler) DeleteInstance(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
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
func (r *InstanceInactiveTerminationReconciler) NotifyInstanceDeletion(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("notify-instance-deletion")
	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	tenant := clctx.TenantFrom(ctx)
	if tenant == nil {
		return fmt.Errorf("tenant not found in context")
	}

	// Send the notification email
	if r.EnableInactivityNotifications {
		if err := SendDestructionNotification(ctx, r.MailClient); err != nil {
			return fmt.Errorf("failed sending notification email: %w", err)
		}
		log.Info("Notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
	} else {
		log.Info("Destruction notifications are disabled, skipping email notification", "instance", instance.Name, "email", tenant.Spec.Email)
	}

	return nil
}
