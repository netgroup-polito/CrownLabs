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
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
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
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	pkgcontext "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/mail"
)

// InstanceInactiveTerminationReconciler watches for instances to be terminated.
type InstanceInactiveTerminationReconciler struct {
	client.Client
	EventsRecorder                   record.EventRecorder
	Scheme                           *runtime.Scheme
	NamespaceWhitelist               metav1.LabelSelector
	StatusCheckRequestTimeout        time.Duration
	InstanceMaxNumberOfAlerts        int
	EnableInactivityNotifications    bool
	NotificationInterval             time.Duration
	MailClient                       *mail.MailClient
	PrometheusURL                    string
	PrometheusNginxAvailability      string
	PrometheusBastionSSHAvailability string
	PrometheusNginxData              string
	PrometheusBastionSSHData         string
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
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				template, ok := obj.(*clv1alpha2.Template)
				if !ok || template.Spec.InactivityTimeout == NEVER_TIMEOUT_VALUE {
					return nil
				}
				return getTemplateInstanceRequests(ctx, r.Client, template)
			}),
			builder.WithPredicates(inactivityTimeoutChanged),
		).
		Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			var requests []reconcile.Request
			namespace, ok := obj.(*corev1.Namespace)
			if !ok || namespace.Labels[forge.InstanceInactivityIgnoreNamespace] == "true" {
				return requests
			}

			var instances clv1alpha2.InstanceList
			if err := r.List(ctx, &instances, client.InNamespace(namespace.Namespace)); err != nil {
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
			//rintln("Enqueued requests for namespace:", namespace.Name, "with", len(requests), "instances", requests)
			return requests
		}),
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

	instance, template, err := r.GetInstanceAndTemplate(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	tracer.Step("instance and template retrieved")

	// Add the instance and template to the context
	ctx, _ = pkgcontext.InstanceInto(ctx, instance)
	ctx, _ = pkgcontext.TemplateInto(ctx, template)

	// Get inactivityTimeout from the template
	inactivityTimeout := template.Spec.InactivityTimeout
	// If set to neverTimeoutValue, return without rescheduling
	if inactivityTimeout == NEVER_TIMEOUT_VALUE {
		log.Info("Instance marked as never stop", "name", instance.GetName(), "namespace", instance.GetNamespace())
		return ctrl.Result{}, nil
	}

	inactivityTimeoutDuration, err := ParseDurationWithDays(ctx, inactivityTimeout)
	if err != nil {
		log.Error(err, "failed to parse deleteAfter duration")
		return ctrl.Result{}, fmt.Errorf("failed to parse inactivityTimeout duration %s: %w", inactivityTimeout, err)
	}

	// Check if the reconciliation should be skipped based on the selector label and namespace labels.
	skip, err := r.CheckSkipReconciliation(ctx)
	if skip {
		return ctrl.Result{}, err
	}

	tracer.Step("labels checked")

	err = r.SetupInstanceAnnotations(ctx)
	if err != nil {
		log.Error(err, "failed setting up instance annotations")
		return ctrl.Result{}, err
	}

	tracer.Step("annotations setup done")

	// LOCAL: comment this check
	// update the last login time of the instance based on the Prometheus data
	if err := r.UpdateInstanceLastLogin(ctx, inactivityTimeoutDuration); err != nil {
		log.Error(err, "failed updating last login time of the instance")
		return ctrl.Result{}, err
	}

	tracer.Step("instance last login updated")

	remainingTime, err := r.GetRemainingInactivityTime(ctx, inactivityTimeoutDuration)
	if err != nil {
		log.Error(err, "failed checking instance termination")
		return ctrl.Result{}, err
	}

	log.Info("instance termination check", "remainingTime", remainingTime.String(), "instance", instance.Name)
	tracer.Step("Inactive termination check done")

	numberAlertSent, err := strconv.Atoi(instance.ObjectMeta.Annotations[forge.AlertAnnotationNum])
	if err != nil {
		log.Error(err, "failed converting string of alerts sent in int number")
		return ctrl.Result{}, err
	}
	remainingAlertToSend := r.InstanceMaxNumberOfAlerts - numberAlertSent
	notificationThreshold := inactivityTimeoutDuration - (r.NotificationInterval * time.Duration(remainingAlertToSend))

	// Check if the instance has expired
	if remainingTime <= 0 {
		// Check if all notifications have already been sent
		shouldSend, err := r.shouldSendNotification(ctx, instance)
		if err != nil {
			log.Error(err, "failed checking if should send notification")
			return ctrl.Result{}, err
		}

		if shouldSend {
			// If a notification should be sent, send the email and requeue after the notification interval
			if err := r.sendInactivityWarning(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: r.NotificationInterval}, nil
		}

		// If all notifications have been sent, terminate the instance
		if instance.Spec.Running {
			if err := r.TerminateInstance(ctx); err != nil {
				log.Error(err, "failed terminating instance", "instance", instance.Name)
				return ctrl.Result{}, err
			}
			log.Info("Instance has been paused/deleted due to inactivity", "instance", instance.Name)
			return ctrl.Result{}, nil
		}
	} else { // remainingTime > 0

		if remainingTime <= notificationThreshold {

			shouldSend, err := r.shouldSendNotification(ctx, instance)
			if err != nil {
				log.Error(err, "failed checking if should send notification")
				return ctrl.Result{}, err
			}

			if shouldSend {

				if err := r.sendInactivityWarning(ctx, instance); err != nil {
					return ctrl.Result{}, err
				}
				// Update the last notification time annotation
				patch := client.MergeFrom(instance.DeepCopy())
				instance.ObjectMeta.Annotations[forge.LastNotificationTimestampAnnotation] = time.Now().Format(time.RFC3339)
				if err := r.Patch(ctx, instance, patch); err != nil {
					log.Error(err, "failed updating instance annotations")
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{RequeueAfter: r.NotificationInterval}, nil
		}
	}

	tracer.Step("Inactive termination done")

	// Calculate requeue time at the instance inactive deadline time:
	// if the instance is not yet to be terminated, we requeue it after the remaining time
	requeueTime := notificationThreshold
	// add 1 minute to the remaining time to avoid requeueing just before the deadline
	// avoiding a double requeue
	requeueTime -= 1 * time.Minute

	dbgLog.Info("requeueing instance")
	return ctrl.Result{RequeueAfter: requeueTime}, nil
}

// GetInstanceAndTemplate retrieves the instance and associated template.
func (r *InstanceInactiveTerminationReconciler) GetInstanceAndTemplate(ctx context.Context, req ctrl.Request) (*clv1alpha2.Instance, *clv1alpha2.Template, error) {
	log := ctrl.LoggerFrom(ctx)

	var instance clv1alpha2.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving instance")
		}
		return nil, nil, err
	}

	var template clv1alpha2.Template
	if err := r.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.Template.Name,
		Namespace: instance.Spec.Template.Namespace,
	}, &template); err != nil {
		log.Error(err, "Unable to fetch the instance template.")
		return nil, nil, fmt.Errorf("failed to fetch instance template %s/%s: %w",
			instance.Spec.Template.Namespace, instance.Spec.Template.Name, err)
	}

	return &instance, &template, nil
}

// GetLastActivityTime retrieves the last time an instance was accessed.
func GetLastActivityTime(query string, promClient v1.API, interval time.Duration) (time.Time, error) {
	end := time.Now()
	start := end.Add(-interval)

	r := v1.Range{
		Start: start,
		End:   end,
		Step:  time.Minute,
	}

	result, warnings, err := promClient.QueryRange(context.Background(), query, r)
	if err != nil {
		return time.Time{}, fmt.Errorf("query failed: %w", err)
	}
	if len(warnings) > 0 {
		fmt.Println("Warnings:", warnings)
	}

	matrix, ok := result.(model.Matrix)
	if !ok {
		return time.Time{}, fmt.Errorf("unexpected result format")
	}

	var lastChange time.Time

	for _, stream := range matrix {
		var prevValue model.SampleValue
		first := true
		for _, sample := range stream.Values {
			if first {
				prevValue = sample.Value
				first = false
				continue
			}
			if sample.Value != prevValue {
				lastChange = sample.Timestamp.Time()
				prevValue = sample.Value
			}
		}
	}

	if lastChange.IsZero() {
		return time.Time{}, fmt.Errorf("no changes detected")
	}
	return lastChange, nil
}

// UpdateInstanceLastLogin updates the last login time of the instance in the annotations.
func (r *InstanceInactiveTerminationReconciler) UpdateInstanceLastLogin(ctx context.Context, inactivityTimeoutDuration time.Duration) error {
	log := ctrl.LoggerFrom(ctx).WithName("update-instance-last-login")
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	// Create the Prometheus client
	promClient, err := api.NewClient(
		api.Config{
			Address: r.PrometheusURL,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	v1api := v1.NewAPI(promClient)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Check Prometheus health first
	healthy, err := r.IsPrometheusHealthy(ctx, v1api) // LOCAL: true, nil
	if err != nil || !healthy {
		log.Info("Prometheus is not healthy", "error", err)
		return err
	}

	// Get instance activity data
	queryNginx := fmt.Sprintf(r.PrometheusNginxData, instance.Namespace, instance.Name)
	lastActivityTimeNginx, errNginx := GetLastActivityTime(queryNginx, v1api, inactivityTimeoutDuration)

	querySSH := fmt.Sprintf(r.PrometheusBastionSSHData, instance.Status.IP)
	lastActivityTimeSSH, errSSH := GetLastActivityTime(querySSH, v1api, inactivityTimeoutDuration)

	if errNginx != nil && errSSH != nil {
		return fmt.Errorf("failed retrieving last activity time from both Nginx and SSH queries: %w", errNginx)
	}

	var maxLastActivityTime time.Time
	if lastActivityTimeNginx.After(lastActivityTimeSSH) {
		maxLastActivityTime = lastActivityTimeNginx
	} else {
		maxLastActivityTime = lastActivityTimeSSH
	}

	instance.ObjectMeta.Annotations[forge.LastActivityAnnotation] = maxLastActivityTime.Format(time.RFC3339)
	return nil
}

// GetRemainingInactivityTime checks if the Instance has to be terminated.
func (r *InstanceInactiveTerminationReconciler) GetRemainingInactivityTime(ctx context.Context, inactivityTimeoutDuration time.Duration) (time.Duration, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-instance-termination")
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return 0, fmt.Errorf("instance not found in context")
	}
	var remainingTime time.Duration

	lastLogin, err := time.Parse(time.RFC3339, instance.ObjectMeta.Annotations[forge.LastActivityAnnotation])
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

// IsPrometheusHealthy checks if Prometheus and required metrics are available.
func (r *InstanceInactiveTerminationReconciler) IsPrometheusHealthy(ctx context.Context, v1api v1.API) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("prometheus-health")

	// Verify connection to Prometheus health endpoint
	promURL := r.PrometheusURL
	healthEndpoint := fmt.Sprintf("%s/-/healthy", promURL)

	statusCode, _, err := utils.HTTPGet(ctx, healthEndpoint, 5*time.Second)
	if err != nil {
		log.Error(err, "Failed to connect to Prometheus health endpoint")
		return false, fmt.Errorf("prometheus health check failed: %w", err)
	}

	if statusCode != http.StatusOK {
		log.Info("Prometheus health check returned non-OK status", "statusCode", statusCode)
		return false, nil
	}

	// Check if ingress metrics and bastion metrics are available on worker nodes
	query1 := r.PrometheusNginxAvailability
	query2 := r.PrometheusBastionSSHAvailability

	result1, _, err1 := v1api.Query(ctx, query1, time.Now())
	result2, _, err2 := v1api.Query(ctx, query2, time.Now())

	if err1 != nil && err2 != nil {
		log.Error(err1, "Failed to query Prometheus for ingress metrics")
		log.Error(err2, "Failed to query Prometheus for bastion SSH metrics")
		return false, fmt.Errorf("both Prometheus queries failed: %v, %v", err1, err2)
	}

	active1 := false
	active2 := false

	if err1 == nil {
		vec1, ok1 := result1.(model.Vector)
		if ok1 && len(vec1) > 0 && int(vec1[0].Value) > 0 {
			active1 = true
		}
	}
	if err2 == nil {
		vec2, ok2 := result2.(model.Vector)
		if ok2 && len(vec2) > 0 && int(vec2[0].Value) > 0 {
			active2 = true
		}
	}

	if !active1 && !active2 {
		log.Info("Neither ingress metrics nor bastion SSH metrics are available on worker nodes")
		return false, nil
	}

	// At least one node has ingress metrics available
	return true, nil
}

// TerminateInstance terminates the Instance.
func (r *InstanceInactiveTerminationReconciler) TerminateInstance(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("termination")

	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}
	template := pkgcontext.TemplateFrom(ctx)
	if template == nil {
		return fmt.Errorf("template not found in context")
	}

	log.Info("Terminating instance", "instance", instance.Name, " in namespace", instance.Namespace)

	var environment = template.Spec.EnvironmentList[0]
	if environment.Persistent {
		log.Info("Stopping persistent instance...")
		instance.Spec.Running = false
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

	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}

	original := instance.DeepCopy()
	// add annotations to the instance if not present
	if instance.ObjectMeta.Annotations == nil {
		instance.ObjectMeta.Annotations = make(map[string]string)
	}

	updated := false

	// Check and set the alert annotation if not present
	if _, ok := instance.ObjectMeta.Annotations[forge.AlertAnnotationNum]; !ok {
		log.Info("adding annotation to instance for the first time", "annotation", forge.AlertAnnotationNum)
		instance.ObjectMeta.Annotations[forge.AlertAnnotationNum] = "0"
		updated = true
	}

	// Check and set the last activity annotation if not present
	if _, ok := instance.ObjectMeta.Annotations[forge.LastActivityAnnotation]; !ok {
		log.Info("adding annotation to instance for the first time", "annotation", forge.LastActivityAnnotation)
		instance.ObjectMeta.Annotations[forge.LastActivityAnnotation] = time.Now().Format(time.RFC3339)
		updated = true
	}

	// Check and set the last notification time annotation if not present
	if _, ok := instance.ObjectMeta.Annotations[forge.LastNotificationTimestampAnnotation]; !ok {
		log.Info("adding annotation to instance for the first time", "annotation", forge.LastNotificationTimestampAnnotation)
		updated = true
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
func (r *InstanceInactiveTerminationReconciler) CheckSkipReconciliation(ctx context.Context) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-skip-reconciliation")

	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return true, fmt.Errorf("instance not found in context")
	}

	// Check the selector label, in order to know whether to perform or not reconciliation.
	if proceed, err := utils.CheckSelectorLabel(ctx, r.Client, instance.GetNamespace(), r.NamespaceWhitelist.MatchLabels); !proceed {
		if err != nil {
			err = fmt.Errorf("failed checking selector label: %w", err)
		}
		return true, err
	}

	var namespace corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Namespace}, &namespace); err != nil {
		log.Error(err, "failed retrieving instance namespace", "instance", instance.Name)
		return true, err
	}

	// check the namespace labels, in order to know whether to perform or not reconciliation on a specific namespace.
	if stop := utils.CheckSingleLabel(&namespace, forge.InstanceInactivityIgnoreNamespace, strconv.FormatBool(true)); stop {
		log.Info("label present, skipping inactivity reconciliation for namespace", "namespace", instance.Namespace, "label", forge.InstanceInactivityIgnoreNamespace)
		return true, nil
	}

	log.Info("proceeding with inactivity reconciliation for instance", "instance", instance.Name, "namespace", instance.Namespace)
	return false, nil
}

func (r *InstanceInactiveTerminationReconciler) shouldSendNotification(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("shouldSendNotification")

	if !r.EnableInactivityNotifications {
		log.Info("Inactivity notifications are disabled, skipping email notification", "instance", instance.Name)
		return false, nil
	}

	//TODO check last email sent time

	numAlerts, err := strconv.Atoi(instance.ObjectMeta.Annotations[forge.AlertAnnotationNum])
	if err != nil {
		log.Error(err, "failed converting string of alerts sent in int number", "annotation", instance.ObjectMeta.Annotations[forge.AlertAnnotationNum])
		return false, err
	}

	lastNotificationTimeStr, ok := instance.ObjectMeta.Annotations[forge.LastNotificationTimestampAnnotation]
	if !ok {
		log.Info("Last notification time annotation not found, sending notification", "instance", instance.Name)
		return true, nil
	}
	lastNotificationTime, err := time.Parse(time.RFC3339, lastNotificationTimeStr)
	if err != nil {
		log.Error(err, "failed parsing last notification time", "lastNotificationTime", lastNotificationTimeStr)
		return false, err
	}
	if numAlerts > 0 {
		if time.Since(lastNotificationTime) < r.NotificationInterval-1*time.Minute {
			log.Info("Last notification sent within the notification interval, skipping email notification", "instance", instance.Name)
			return false, nil
		}

	}

	return numAlerts <= r.InstanceMaxNumberOfAlerts-1, nil
}

func (r *InstanceInactiveTerminationReconciler) sendInactivityWarning(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx)
	tenant, err := GetTenantFromInstance(ctx, r.Client)
	if err != nil {
		log.Error(err, "failed retrieving tenant from instance")
	}
	ctx, _ = pkgcontext.TenantInto(ctx, tenant)

	err = SendInactivityNotification(ctx, r.MailClient)
	if err != nil {
		log.Error(err, "failed sending notification email to user", "email", tenant.Spec.Email)
		return err //LOCAL: return nil
	}

	newNumberOfAlerts, err := r.IncrementAnnotation(ctx, instance.ObjectMeta.Annotations[forge.AlertAnnotationNum])
	if err != nil {
		log.Error(err, "failed incrementing annotation")
		return err
	}

	// if err := r.Update(ctx, instance); err != nil {
	// 	log.Error(err, "failed updating instance annotations")
	// 	return err
	// }
	patch := client.MergeFrom(instance.DeepCopy())
	instance.ObjectMeta.Annotations[forge.AlertAnnotationNum] = newNumberOfAlerts
	if err := r.Patch(ctx, instance, patch); err != nil {
		log.Error(err, "failed updating instance annotations")
	}
	return nil
}
