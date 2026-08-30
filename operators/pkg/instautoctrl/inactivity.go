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
	"errors"
	"fmt"
	"math/rand"
	"reflect"
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
	MinLastActivityRequeueTime      time.Duration
	MaxLastActivityRequeueTime      time.Duration
	LastActivityCheckThreshold      time.Duration
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
				if t.Spec.Cleanup.StopAfterInactivity != NeverTimeoutValue {
					return t.Spec.Cleanup.StopAfterInactivity
				}
				return t.Spec.Cleanup.DeleteAfterInactivity
			}),
			builder.WithPredicates(stopAfterInactivityChanged),
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

// Reconcile reconciles the status of the Instance resource.
func (r *InstanceInactiveTerminationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}
	log := ctrl.LoggerFrom(ctx, "instance", req.NamespacedName)
	dbgLog := log.V(utils.LogDebugLevel)
	tracer := trace.New("reconcile", trace.Field{Key: "instance", Value: req.NamespacedName})
	ctx = ctrl.LoggerInto(trace.ContextWithTrace(ctx, tracer), log)

	// ── 1. Check selector label to early abort reconciliation ──
	if proceed, selectorErr := utils.CheckSelectorLabel(ctx, r.Client, req.Namespace, r.NamespaceWhitelist.MatchLabels); !proceed {
		if selectorErr != nil {
			log.Error(selectorErr, "failed checking selector label")
			return ctrl.Result{}, selectorErr
		}
		return ctrl.Result{}, nil
	}

	// ── 2. Fetch Instance & Template ──
	var instance clv1alpha2.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving instance")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var template clv1alpha2.Template
	if err := r.Get(ctx, forge.NamespacedNameFromGenericRef(instance.Spec.Template), &template); err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to retrieve instance template")
		return ctrl.Result{}, err
	}
	tracer.Step("instance and template retrieved")

	ctx, _ = clctx.InstanceInto(ctx, &instance)
	ctx, _ = clctx.TemplateInto(ctx, &template)

	// ── 3. Defer: patch instance object + delete if flagged ──
	var deleteInstance bool
	defer func(original *clv1alpha2.Instance) {
		// Delete instance if flagged by a handler.
		if deleteInstance {
			if deleteErr := r.Delete(ctx, &instance); deleteErr != nil && !kerrors.IsNotFound(deleteErr) {
				log.Error(deleteErr, "failed to delete instance")
				err = deleteErr
			} else if deleteErr == nil {
				tracer.Step("instance deleted")
				log.Info("Instance deleted", "instance", instance.Name)
			}
			return
		}

		// Patch annotations and spec if changed.
		annotationsChanged := !reflect.DeepEqual(original.Annotations, instance.Annotations)
		specChanged := !reflect.DeepEqual(original.Spec, instance.Spec)
		if annotationsChanged || specChanged {
			if patchErr := r.Patch(ctx, &instance, client.MergeFrom(original)); patchErr != nil {
				log.Error(patchErr, "failed to patch instance")
				err = patchErr
			} else {
				tracer.Step("instance patched")
			}
		}
	}(instance.DeepCopy())

	// ── 4. Ensure annotations (setup defaults + reset on state transition) ──
	r.EnsureAnnotations(ctx)

	// ── 5. Instance NOT running → check powered-off destruction ──
	if !instance.Spec.Running {
		return r.HandlePoweredOffInstance(ctx, &instance, &deleteInstance)
	}

	// ── 6. Instance running, Prometheus not available → exit ──
	if r.Prometheus == nil {
		dbgLog.Info("Prometheus not configured, skipping activity tracking")
		return ctrl.Result{}, nil
	}

	// ── 7. Update last activity from Prometheus (runs for ALL running instances) ──
	if r.ShouldSkipDueToThreshold(ctx, &instance) {
		return ctrl.Result{RequeueAfter: r.LastActivityCheckThreshold}, nil
	}

	if updateErr := r.UpdateLastActivity(ctx); updateErr != nil {
		log.Error(updateErr, "failed to update last activity annotation")
		return ctrl.Result{}, updateErr
	}
	tracer.Step("last activity checked")

	// ── 8. Check namespace label to skip inactivity termination ──
	skip, skipErr := r.CheckSkipInactivityByNSLabel(ctx, req.Namespace)
	if skip {
		return r.RequeueAfterRandom(), skipErr
	}

	stopAfterInactivity := template.Spec.Cleanup.StopAfterInactivity
	// If set to NeverTimeoutValue, return but schedule a requeue to keep refreshing activity
	if stopAfterInactivity == NeverTimeoutValue {
		dbgLog.Info("Instance marked as never stop", "name", instance.GetName(), "namespace", instance.GetNamespace())
		return r.RequeueAfterRandom(), nil
	}

	stopAfterInactivityDuration, parseErr := ParseDurationWithDays(ctx, stopAfterInactivity)
	if parseErr != nil {
		log.Error(parseErr, "failed to parse stopAfterInactivity duration")
		return ctrl.Result{}, fmt.Errorf("failed to parse stopAfterInactivity duration %s: %w", stopAfterInactivity, parseErr)
	}

	remainingTime, remainErr := r.GetRemainingInactivityTime(ctx, stopAfterInactivityDuration)
	if remainErr != nil {
		log.Error(remainErr, "failed checking instance termination")
		return ctrl.Result{}, remainErr
	}

	dbgLog.Info("instance termination check", "remainingTime", remainingTime.String(), "instance", instance.Name)
	tracer.Step("inactive termination check done")

	// Check if the instance has expired
	if remainingTime <= 0 {
		res, terminateEarly, handleErr := r.HandleInactivityInstance(ctx, &instance, &deleteInstance)
		if terminateEarly || handleErr != nil {
			return res, handleErr
		}
	}

	tracer.Step("inactive termination done")

	// Calculate requeue time, ensuring periodic activity refresh
	requeueTime := remainingTime + r.MarginTime
	activityResult := r.RequeueAfterRandom()
	if activityResult.RequeueAfter < requeueTime {
		requeueTime = activityResult.RequeueAfter
	}

	dbgLog.Info("requeueing instance")
	return ctrl.Result{RequeueAfter: requeueTime}, nil
}

// HandlePoweredOffInstance manages the inactivity lifecycle for instances that are already powered off.
// If the instance needs to be deleted, it sets *deleteInstance = true so the Reconcile's defer func handles the actual deletion.
func (r *InstanceInactiveTerminationReconciler) HandlePoweredOffInstance(ctx context.Context, instance *clv1alpha2.Instance, deleteInstance *bool) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	dbgLog := log.V(utils.LogDebugLevel)

	remainingPauseTime, isActive, err := r.GetRemainingInactivityDestructionTime(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !isActive {
		// No delete-after-inactivity configured, nothing to do for a powered-off instance.
		return ctrl.Result{}, nil
	}

	// Destruction timer still has time left — requeue.
	if remainingPauseTime > 0 {
		dbgLog.Info("requeueing paused instance for destruction check")
		return ctrl.Result{RequeueAfter: remainingPauseTime + r.MarginTime}, nil
	}

	// Destruction timer expired — handle notification and deletion.
	if r.EnableInactivityNotifications {
		// Check if a warning notification should be sent.
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
			if err := r.SendDestructionWarning(ctx, instance, window); err != nil {
				log.Error(err, "failed sending destruction warning email")
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: r.DestructionNotificationInterval}, nil
		}

		// Check if all notifications have been sent and instance should be deleted.
		shouldDelete, err := r.ShouldDeleteInstance(ctx, instance)
		if err != nil {
			log.Error(err, "failed checking if should delete instance")
			return ctrl.Result{}, err
		}
		if !shouldDelete {
			// Still waiting for the next notification interval.
			lastNotificationTimeStr := instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation]
			lastNotificationTime, _ := time.Parse(time.RFC3339, lastNotificationTimeStr)
			requeueTime := r.DestructionNotificationInterval - time.Since(lastNotificationTime) + r.MarginTime
			if requeueTime < 0 {
				requeueTime = r.MarginTime
			}
			dbgLog.Info("requeueing paused instance to wait for next destruction notification interval")
			return ctrl.Result{RequeueAfter: requeueTime}, nil
		}

		// All notifications sent — notify deletion then flag for deletion.
		log.Info("Deleting paused persistent instance due to prolonged inactivity...")
		if err := r.NotifyInstanceDeletion(ctx); err != nil {
			log.Error(err, "failed to send deletion notification")
			return ctrl.Result{}, err
		}
		*deleteInstance = true
		return ctrl.Result{}, nil
	}

	// Notifications disabled — flag for immediate deletion.
	log.Info("Deleting paused persistent instance due to prolonged inactivity...")
	*deleteInstance = true
	return ctrl.Result{}, nil
}

// HandleInactivityInstance processes the instance when its inactivity timeout has been reached.
// If the instance needs to be deleted, it sets *deleteInstance = true so the Reconcile's defer func handles the actual deletion.
func (r *InstanceInactiveTerminationReconciler) HandleInactivityInstance(ctx context.Context, instance *clv1alpha2.Instance, deleteInstance *bool) (res ctrl.Result, terminateEarly bool, err error) {
	log := ctrl.LoggerFrom(ctx)
	if r.EnableInactivityNotifications {
		// Check if a warning notification should be sent.
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

		// Check if all notifications have been sent and instance should be terminated.
		shouldTerminate, err := r.ShouldTerminateInstance(ctx, instance)
		if err != nil {
			log.Error(err, "failed checking if should terminate instance", "instance", instance.Name, "namespace", instance.Namespace)
			return ctrl.Result{}, true, err
		}
		if shouldTerminate {
			if err := r.TerminateInstance(ctx, deleteInstance); err != nil {
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
		// Notifications disabled — terminate immediately.
		if err := r.TerminateInstance(ctx, deleteInstance); err != nil {
			log.Error(err, "failed terminating inactive instance", "instance", instance.Name, "namespace", instance.Namespace)
			return ctrl.Result{}, true, err
		}
	}
	return ctrl.Result{}, false, nil
}

// ShouldSkipDueToThreshold checks if the activity update should be skipped because
// the threshold since the last Prometheus check has not been reached yet.
func (r *InstanceInactiveTerminationReconciler) ShouldSkipDueToThreshold(ctx context.Context, instance *clv1alpha2.Instance) bool {
	// Check if the threshold has passed since the last Prometheus check
	if lastCheckStr, ok := instance.Annotations[forge.LastActivityCheckTimestampAnnotation]; ok {
		if lastCheckTime, err := time.Parse(time.RFC3339, lastCheckStr); err == nil {
			if time.Since(lastCheckTime) < r.LastActivityCheckThreshold {
				log := ctrl.LoggerFrom(ctx).WithName("update-instance-last-activity")
				log.Info("Skipping activity update, threshold not reached", "threshold", r.LastActivityCheckThreshold)
				return true
			}
		}
	}

	return false
}

// UpdateLastActivity updates the last activity time of the instance in the annotations.
func (r *InstanceInactiveTerminationReconciler) UpdateLastActivity(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	log := ctrl.LoggerFrom(ctx).WithName("update-instance-last-activity")

	log.Info("Checking updates for lastActivity")

	// Check Prometheus health
	healthy, err := r.Prometheus.IsPrometheusHealthy(ctx, r.StatusCheckRequestTimeout)
	if err != nil || !healthy {
		log.Error(err, "Prometheus not healthy, failing activity update")
		if err == nil {
			err = fmt.Errorf("prometheus is not healthy")
		}
		return err
	}

	// Use the max requeue time as the Prometheus lookback window
	lookback := r.MaxLastActivityRequeueTime
	var queryErrors []error

	// Query Nginx activity
	queryNginx := fmt.Sprintf(r.Prometheus.GetQueryNginxData(), instance.Namespace, instance.Name)
	lastActivityNginx, errNginx := r.Prometheus.GetLastActivityTime(queryNginx, lookback)
	if errNginx != nil {
		log.Error(errNginx, "failed querying Nginx activity")
		queryErrors = append(queryErrors, fmt.Errorf("failed querying Nginx activity: %w", errNginx))
	}

	// Query WebSSH activity across all environments (find maximum)
	var lastActivityWebSSH time.Time
	webSSHFound := false
	for envIdx := range instance.Status.Environments {
		env := &instance.Status.Environments[envIdx]
		q := fmt.Sprintf(r.Prometheus.GetQueryWebSSHData(), env.IP)
		t, errWebSSH := r.Prometheus.GetLastActivityTime(q, lookback)
		if errWebSSH != nil {
			log.Error(errWebSSH, "failed querying WebSSH activity", "environmentIP", env.IP)
			queryErrors = append(queryErrors, fmt.Errorf("failed querying WebSSH activity for environment %q: %w", env.IP, errWebSSH))
			continue
		}
		if !t.IsZero() && (!webSSHFound || t.After(lastActivityWebSSH)) {
			lastActivityWebSSH = t
			webSSHFound = true
		}
	}

	// Query SSH activity across all environments (find maximum)
	var lastActivitySSH time.Time
	sshFound := false
	for envIdx := range instance.Status.Environments {
		env := &instance.Status.Environments[envIdx]
		q := fmt.Sprintf(r.Prometheus.GetQuerySSHData(), env.IP)
		t, errSSH := r.Prometheus.GetLastActivityTime(q, lookback)
		if errSSH != nil {
			log.Error(errSSH, "failed querying SSH activity", "environmentIP", env.IP)
			queryErrors = append(queryErrors, fmt.Errorf("failed querying SSH activity for environment %q: %w", env.IP, errSSH))
			continue
		}
		if !t.IsZero() && (!sshFound || t.After(lastActivitySSH)) {
			lastActivitySSH = t
			sshFound = true
		}
	}

	// Compute the most recent activity timestamp
	var maxActivity time.Time
	if errNginx == nil && !lastActivityNginx.IsZero() {
		maxActivity = lastActivityNginx
	}
	if lastActivitySSH.After(maxActivity) {
		maxActivity = lastActivitySSH
	}
	if lastActivityWebSSH.After(maxActivity) {
		maxActivity = lastActivityWebSSH
	}

	activityFound := !maxActivity.IsZero()
	newStr := maxActivity.Format(time.RFC3339)

	// Preserve activity found by successful queries even if another source failed.
	// The Reconcile's defer func handles the actual patch.
	if activityFound {
		if instance.Annotations[forge.LastActivityAnnotation] != newStr {
			instance.Annotations[forge.LastActivityAnnotation] = newStr
			instance.Annotations[forge.AlertAnnotationNum] = "0"
			log.Info("Updated lastActivity annotation", "lastActivity", newStr)
		}
	}

	hasQueryErrors := len(queryErrors) > 0
	if activityFound || !hasQueryErrors {
		instance.Annotations[forge.LastActivityCheckTimestampAnnotation] = time.Now().Format(time.RFC3339)
	}

	// If there were any query errors, return an aggregated error to indicate that the activity update was not fully successful.
	if hasQueryErrors {
		return fmt.Errorf("one or more activity queries failed: %w", errors.Join(queryErrors...))
	}

	return nil
}

// RequeueAfterRandom returns a ctrl.Result with a randomized requeue time
// between MinLastActivityRequeueTime and MaxLastActivityRequeueTime.
func (r *InstanceInactiveTerminationReconciler) RequeueAfterRandom() ctrl.Result {
	requeue := r.MinLastActivityRequeueTime
	if r.MaxLastActivityRequeueTime > r.MinLastActivityRequeueTime {
		delta := r.MaxLastActivityRequeueTime - r.MinLastActivityRequeueTime
		//nolint:gosec // Cryptographic randomness is not needed for scheduling jitter
		requeue += time.Duration(rand.Int63n(int64(delta)))
	}
	return ctrl.Result{RequeueAfter: requeue}
}

// GetRemainingInactivityTime checks if the Instance has to be terminated.
func (r *InstanceInactiveTerminationReconciler) GetRemainingInactivityTime(ctx context.Context, stopAfterInactivityDuration time.Duration) (time.Duration, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-instance-termination")
	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return 0, fmt.Errorf("instance not found in context")
	}
	var remainingTime time.Duration

	lastActivity, err := time.Parse(time.RFC3339, instance.Annotations[forge.LastActivityAnnotation])
	if err != nil {
		log.Error(err, "failed parsing LastActivity time")
		return 0, err
	}

	// Check if the instance has been inactive for longer than the timeout duration
	remainingTime = stopAfterInactivityDuration - time.Since(lastActivity)
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
// For persistent instances, it sets Spec.Running = false (in-memory, patched by the Reconcile's defer).
// For non-persistent instances, it sets *deleteInstance = true so the Reconcile's defer handles deletion.
func (r *InstanceInactiveTerminationReconciler) TerminateInstance(ctx context.Context, deleteInstance *bool) error {
	log := ctrl.LoggerFrom(ctx).WithName("termination")

	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}
	template := clctx.TemplateFrom(ctx)
	if template == nil {
		return fmt.Errorf("template not found in context")
	}

	log.Info("Terminating instance", "instance", instance.Name, "namespace", instance.Namespace)

	if IsTemplatePersistent(template) {
		log.Info("Stopping persistent instance...")
		instance.Spec.Running = false
		instance.Annotations[forge.LastRunningAnnotation] = strconv.FormatBool(false)
		return nil
	}

	log.Info("Deleting non-persistent instance...")
	*deleteInstance = true
	return nil
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

// EnsureAnnotations initializes default annotations and resets counters on running state transitions.
// All mutations are in-memory only; the actual patch is deferred to the Reconcile's defer func.
func (r *InstanceInactiveTerminationReconciler) EnsureAnnotations(ctx context.Context) {
	log := ctrl.LoggerFrom(ctx).WithName("ensure-annotations")

	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return
	}

	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}

	// ── Setup defaults ──
	if _, ok := instance.Annotations[forge.AlertAnnotationNum]; !ok {
		log.Info("initializing alert number annotation", "annotation", forge.AlertAnnotationNum)
		instance.Annotations[forge.AlertAnnotationNum] = "0"
	}
	if _, ok := instance.Annotations[forge.LastActivityAnnotation]; !ok {
		log.Info("initializing last activity annotation", "annotation", forge.LastActivityAnnotation)
		instance.Annotations[forge.LastActivityAnnotation] = time.Now().Format(time.RFC3339)
	}
	if _, ok := instance.Annotations[forge.LastNotificationTimestampAnnotation]; !ok {
		log.Info("initializing last notification time annotation", "annotation", forge.LastNotificationTimestampAnnotation)
		instance.Annotations[forge.LastNotificationTimestampAnnotation] = ""
	}
	if _, ok := instance.Annotations[forge.DestructionAlertsSentAnnotation]; !ok {
		log.Info("initializing destruction alert number annotation", "annotation", forge.DestructionAlertsSentAnnotation)
		instance.Annotations[forge.DestructionAlertsSentAnnotation] = "0"
	}
	if _, ok := instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation]; !ok {
		log.Info("initializing last destruction notification time annotation", "annotation", forge.LastDestructionNotificationTimestampAnnotation)
		instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation] = ""
	}
	if _, ok := instance.Annotations[forge.LastPoweredOffTimestampAnnotation]; !ok {
		log.Info("initializing last powered off timestamp annotation", "annotation", forge.LastPoweredOffTimestampAnnotation)
		instance.Annotations[forge.LastPoweredOffTimestampAnnotation] = ""
	}

	// ── Reset on running state transition (false → true) ──
	lastRunningStr := instance.Annotations[forge.LastRunningAnnotation]
	lastRunning := false
	if lastRunningStr != "" {
		if val, err := strconv.ParseBool(lastRunningStr); err == nil {
			lastRunning = val
		}
	}

	if instance.Spec.Running && !lastRunning {
		log.Info("Detected transition from false to true: resetting alert counter and last activity field")
		instance.Annotations[forge.AlertAnnotationNum] = "0"
		instance.Annotations[forge.LastActivityAnnotation] = time.Now().Format(time.RFC3339)
		instance.Annotations[forge.DestructionAlertsSentAnnotation] = "0"
		instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation] = ""
	}

	// Update the LastRunningAnnotation
	currentRunningStr := strconv.FormatBool(instance.Spec.Running)
	if lastRunningStr != currentRunningStr {
		instance.Annotations[forge.LastRunningAnnotation] = currentRunningStr
	}

	log.Info("annotations ensured", "instance", instance.Name)
}

// CheckSkipInactivityByNSLabel checks if the inactivity reconciliation should be skipped
// based on the namespace labels.
func (r *InstanceInactiveTerminationReconciler) CheckSkipInactivityByNSLabel(ctx context.Context, namespace string) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-skip-inactivity-by-ns-label")

	var namespaceObj corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: namespace}, &namespaceObj); err != nil {
		log.Error(err, "failed retrieving namespace", "namespace", namespace)
		return true, err
	}

	if stop := utils.CheckSingleLabel(&namespaceObj, forge.InstanceInactivityIgnoreNamespace, strconv.FormatBool(true)); stop {
		log.Info("label present, skipping inactivity reconciliation for namespace", "namespace", namespace, "label", forge.InstanceInactivityIgnoreNamespace)
		return true, nil
	}

	return false, nil
}

// GetAlertCounts returns the current number of alerts sent and the maximum allowed alerts for the instance.
func (r *InstanceInactiveTerminationReconciler) GetAlertCounts(ctx context.Context, instance *clv1alpha2.Instance) (numAlerts, maxAlerts int, err error) {
	log := ctrl.LoggerFrom(ctx).WithName("GetAlertCounts")

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
		numAlerts, maxAlerts, err := r.GetAlertCounts(ctx, instance)
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

	numAlerts, maxAlerts, err := r.GetAlertCounts(ctx, instance)
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
	tenant, err := GetTenantFromInstance(ctx, r.Client)
	if err != nil {
		log.Error(err, "failed getting tenant from instance")
		return err
	}

	// Calculate the remaining time available for sending inactivity notifications
	remainingTime, err := r.GetInactivityNotificationWindow(ctx, instance)
	if err != nil {
		log.Error(err, "failed calculating remaining time for inactivity notifications")
		return err
	}

	if r.EnableInactivityNotifications {
		ctx, _ = clctx.TenantInto(ctx, tenant)
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

	instance.Annotations[forge.AlertAnnotationNum] = newNumberOfAlerts
	instance.Annotations[forge.LastNotificationTimestampAnnotation] = time.Now().Format(time.RFC3339)

	return nil
}

// SendTerminationNotification handles sending notification emails when an instance is deleted.
func (r *InstanceInactiveTerminationReconciler) SendTerminationNotification(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("send-termination-notification")
	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	tenant, err := GetTenantFromInstance(ctx, r.Client)
	if err != nil {
		log.Error(err, "failed getting tenant from instance")
		return err
	}

	if r.EnableInactivityNotifications {
		ctx, _ = clctx.TenantInto(ctx, tenant)
		if err := SendInactivityTerminationNotification(ctx, r.MailClient, 0); err != nil {
			return fmt.Errorf("failed sending termination notification email: %w", err)
		}
		log.Info("Termination notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
	} else {
		log.Info("Inactivity notifications are disabled, skipping email notification", "instance", instance.Name, "email", tenant.Spec.Email)
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
	tenant, err := GetTenantFromInstance(ctx, r.Client)
	if err != nil {
		log.Error(err, "failed getting tenant from instance")
		return err
	}

	// 1. Call the function to send the email that is in common.go.
	ctx, _ = clctx.TenantInto(ctx, tenant)
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

	instance.Annotations[forge.DestructionAlertsSentAnnotation] = strconv.Itoa(numAlerts + 1)
	instance.Annotations[forge.LastDestructionNotificationTimestampAnnotation] = time.Now().Format(time.RFC3339)

	return nil
}

// GetRemainingInactivityDestructionTime checks the remaining time before the instance is destroyed due to prolonged inactivity while powered off.
func (r *InstanceInactiveTerminationReconciler) GetRemainingInactivityDestructionTime(ctx context.Context, instance *clv1alpha2.Instance) (time.Duration, bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-instance-destruction")
	template := clctx.TemplateFrom(ctx)
	if template == nil {
		return 0, false, fmt.Errorf("template not found in context")
	}

	deleteAfterInactivity := template.Spec.Cleanup.DeleteAfterInactivity
	if deleteAfterInactivity == NeverTimeoutValue || deleteAfterInactivity == "" {
		return 0, false, nil
	}

	deleteAfterInactivityDuration, err := ParseDurationWithDays(ctx, deleteAfterInactivity)
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse deleteAfterInactivity duration %s: %w", deleteAfterInactivity, err)
	}
	poweredOffTimeStr := instance.Annotations[forge.LastPoweredOffTimestampAnnotation]
	if poweredOffTimeStr == "" {
		return 0, false, nil // No powered-off timestamp, nothing to calculate
	}

	// Store the deleteAfterInactivity value as an annotation for validation/visibility
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}
	instance.Annotations["crownlabs.polito.it/delete-after-inactivity"] = deleteAfterInactivity

	poweredOffTime, err := time.Parse(time.RFC3339, poweredOffTimeStr)
	if err != nil {
		log.Error(err, "failed to parse last powered off time", "timestamp", poweredOffTimeStr)
		return 0, false, err
	}

	remainingTime := deleteAfterInactivityDuration - time.Since(poweredOffTime)
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

// NotifyInstanceDeletion handles sending notification emails when an instance is deleted.
func (r *InstanceInactiveTerminationReconciler) NotifyInstanceDeletion(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("notify-instance-deletion")
	instance := clctx.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	tenant, err := GetTenantFromInstance(ctx, r.Client)
	if err != nil {
		log.Error(err, "failed getting tenant from instance")
		return err
	}

	// Send the notification email
	if r.EnableInactivityNotifications {
		ctx, _ = clctx.TenantInto(ctx, tenant)
		if err := SendDestructionNotification(ctx, r.MailClient); err != nil {
			return fmt.Errorf("failed sending notification email: %w", err)
		}
		log.Info("Notification email sent to user", "instance", instance.Name, "email", tenant.Spec.Email)
	} else {
		log.Info("Destruction notifications are disabled, skipping email notification", "instance", instance.Name, "email", tenant.Spec.Email)
	}

	return nil
}
