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
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/trace"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// InstanceInactiveTerminationReconciler watches for instances to be terminated.
type InstanceInactiveTerminationReconciler struct {
	client.Client
	EventsRecorder            record.EventRecorder
	Scheme                    *runtime.Scheme
	NamespaceWhitelist        metav1.LabelSelector
	StatusCheckRequestTimeout time.Duration
	InstanceMaxNumberOfAlerts int
	MailClient                *MailClient
	PrometheusURL             string
	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

var alertAnnotation = "crownlabs.polito.it/number-alerts-sent"
var lastActivityAnnotation = "crownlabs.polito.it/last-activity"
var deleteAfterRegex = regexp.MustCompile(`^(\d+)([mhd])$`)

// SetupWithManager registers a new controller for InstanceTerminationReconciler resources.
func (r *InstanceInactiveTerminationReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Instance{}).
		Named("instance-inactive-termination").
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		// Do not requeue on update events
		// Inactive Instance Controller is triggered only by requeue events
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldInstance, oldOk := e.ObjectOld.(*clv1alpha2.Instance)
				newInstance, newOk := e.ObjectNew.(*clv1alpha2.Instance)
				if !oldOk || !newOk {
					return false
				}

				oldValue := oldInstance.Labels[forge.InstanceInactivityIgnoreNamespace]
				newValue := newInstance.Labels[forge.InstanceInactivityIgnoreNamespace]

				// Requeue only if the IstanceInactivityIgnoreNamespace label has changed
				return oldValue != newValue
			},
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

	// check the namespace labels, in order to know whether to perform or not reconciliation on a specific namespace.
	if stop := utils.CheckSingleLabel(&instance, forge.InstanceInactivityIgnoreNamespace, strconv.FormatBool(true)); stop {
		log.Info("label present, skipping inactivity reconciliation for namespace", "namespace", instance.Namespace, "label", forge.InstanceInactivityIgnoreNamespace)
		return ctrl.Result{}, nil
	}

	tracer.Step("labels checked")

	// add the annotation if not present to check the number of termination alerts
	if instance.ObjectMeta.Annotations == nil {
		instance.ObjectMeta.Annotations = make(map[string]string)
	}
	patch := client.MergeFrom(instance.DeepCopy())
	if _, ok := instance.ObjectMeta.Annotations[alertAnnotation]; !ok {
		log.Info("adding annotation to instance for the first time", "annotation", alertAnnotation)
		instance.ObjectMeta.Annotations[alertAnnotation] = "0"

		// update the instance with the new annotation
		_ = r.Patch(ctx, &instance, patch)
	}

	// add the annotation if not present to check the last activity time
	if instance.ObjectMeta.Annotations == nil {
		instance.ObjectMeta.Annotations = make(map[string]string)
	}
	patch = client.MergeFrom(instance.DeepCopy())
	if _, ok := instance.Annotations[lastActivityAnnotation]; !ok {
		log.Info("adding annotation to instance for the first time", "annotation", lastActivityAnnotation)
		instance.ObjectMeta.Annotations[lastActivityAnnotation] = time.Now().Format(time.RFC3339)
		// update the instance with the new annotation
		_ = r.Patch(ctx, &instance, patch)
	}

	tracer.Step("annotations checked")

	// update the last login time of the instance
	if err := r.UpdateInstanceLastLogin(ctx, &instance); err != nil {
		log.Error(err, "failed updating last login time of the instance")
		return ctrl.Result{}, err
	}

	tracer.Step("instance last login updated")

	// check if the instance reached the maximum time of lifetime and delete it if so
	isDeleted, err := r.deleteStaleInstances(ctx, &instance)
	if err != nil {
		log.Error(err, "failed delete-stale-instances")
	}
	if isDeleted {
		log.Info("instance is deleted, skipping inactivity check", "instance", instance.Name)
		return ctrl.Result{}, nil
	}

	tracer.Step("stale instances done")

	// check for inactivity and decide whether to terminate the instance or not
	terminate, remainingTime, err := r.CheckInstanceTermination(ctx, &instance)
	if err != nil {
		log.Error(err, "failed checking instance termination")
		return ctrl.Result{}, err
	}

	tracer.Step("Inactive termination check done")

	log.Info("instance termination check", "terminate", terminate)
	if terminate {
		// retrieve the user owner of the instance
		user, err := r.GetTenantFromInstance(ctx, &instance)
		if err != nil {
			log.Error(err, "failed retrieving user from instance")
			return ctrl.Result{}, err
		}

		// send notification to the user
		numberAlertSent, err := strconv.Atoi(instance.ObjectMeta.Annotations[alertAnnotation])
		if err != nil {
			log.Error(err, "failed converting string of alerts sent in int number")
			return ctrl.Result{}, err
		}

		if numberAlertSent < r.InstanceMaxNumberOfAlerts {
			err := r.SendNotification(ctx, &instance, user.Spec.Email)
			if err != nil {
				log.Error(err, "failed sending notification email to user", "email", user.Spec.Email)
				return ctrl.Result{}, err
			}
		} else if numberAlertSent >= r.InstanceMaxNumberOfAlerts && instance.Spec.Running {
			err := r.TerminateInstance(ctx, &instance)
			if err != nil {
				log.Error(err, "failed terminating instance", "instance", instance.Name)
				return ctrl.Result{}, err
			}
		}
	} else {
		log.Info("instance is not yet to be terminated", "instance", instance.Name)
	}

	tracer.Step("Inactive termination done")

	// Calculate requeue time at the instance inactive deadline time:
	// if the instance is not yet to be terminated, we requeue it after the remaining time
	requeueTime := remainingTime
	// add 1 minute to the remaining time to avoid requeueing just before the deadline
	// avoiding a double requeue
	requeueTime += 1 * time.Minute

	dbgLog.Info("requeueing instance")
	return ctrl.Result{RequeueAfter: requeueTime}, nil
}

// getLastActivityTime retrieves the last time an instance was accessed.
func getLastActivityTime(query string, promClient v1.API, interval time.Duration) (time.Time, error) {
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
func (r *InstanceInactiveTerminationReconciler) UpdateInstanceLastLogin(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx).WithName("update-instance-last-login")

	promURL := r.PrometheusURL

	config := api.Config{
		Address: promURL,
	}

	promClient, err := api.NewClient(config)
	if err != nil {
		return fmt.Errorf("error creating prometheus client: %w", err)
	}

	v1api := v1.NewAPI(promClient)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Check Prometheus health first
	healthy, err := r.isPrometheusHealthy(ctx, v1api)
	if err != nil || !healthy {
		log.Info("Prometheus is not healthy", "error", err)
		return err
	}
	// Get instance activity data
	interval, err := r.getInactivityTimeout(ctx, instance)
	if err != nil {
		log.Error(err, "failed retrieving inactivity timeout from instance template")
		return err
	}
	intervalDuration, err := time.ParseDuration(interval)
	if err != nil {
		log.Error(err, "failed parsing inactivity timeout duration")
		return err
	}

	// Get instance activity data
	queryNginx := fmt.Sprintf(`nginx_ingress_controller_requests{exported_namespace=%q, exported_service=%q}`, instance.Namespace, instance.Name)
	lastActivityTimeNginx, errNginx := getLastActivityTime(queryNginx, v1api, intervalDuration)

	querySSH := fmt.Sprintf(`bation_ssh_conntections{namespace=%q, destination_Ip=%q}`, instance.Namespace, instance.Status.IP)
	lastActivityTimeSSH, errSSH := getLastActivityTime(querySSH, v1api, intervalDuration)

	if errNginx != nil && errSSH != nil {
		return fmt.Errorf("failed retrieving last activity time from both Nginx and SSH queries: %w", errNginx)
	}

	var maxLastActivityTime time.Time
	if lastActivityTimeNginx.After(lastActivityTimeSSH) {
		maxLastActivityTime = lastActivityTimeNginx
	} else {
		maxLastActivityTime = lastActivityTimeSSH
	}

	instance.ObjectMeta.Annotations[lastActivityAnnotation] = maxLastActivityTime.Format(time.RFC3339)
	return nil
}

// CheckInstanceTermination checks if the Instance has to be terminated.
func (r *InstanceInactiveTerminationReconciler) CheckInstanceTermination(ctx context.Context, instance *clv1alpha2.Instance) (bool, time.Duration, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-instance-termination")
	var remainingTime time.Duration

	// get the inactivity timeout from the instance template
	inactivityTimeout, err := r.getInactivityTimeout(ctx, instance)
	if err != nil {
		log.Error(err, "failed retrieving inactivity timeout from instance template")
		return false, 0, err
	}

	lastLogin, err := time.Parse(time.RFC3339, instance.ObjectMeta.Annotations[lastActivityAnnotation])
	if err != nil {
		log.Error(err, "failed parsing LastLogin time")
		return false, 0, err
	}
	timeoutDuration, err := time.ParseDuration(inactivityTimeout)
	if err != nil {
		log.Error(err, "failed parsing inactivity timeout duration")
		return false, 0, err
	}

	// Check if the instance has been inactive for longer than the timeout duration
	remainingTime = timeoutDuration - time.Since(lastLogin)
	if remainingTime <= 0 {
		log.Info("Instance inactivity detected", "instance", instance.Name)
		return true, 0, nil
	}

	return false, remainingTime, nil
}

// isPrometheusHealthy checks if Prometheus and required metrics are available.
func (r *InstanceInactiveTerminationReconciler) isPrometheusHealthy(ctx context.Context, v1api v1.API) (bool, error) {
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

	// Check if ingress metrics are available on worker nodes
	query := `count(up{service="ingress-nginx-external-controller-metrics", node=~"worker-.*"} == 1)`
	result, _, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		log.Error(err, "Failed to query Prometheus for ingress metrics")
		return false, err
	}

	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		log.Info("No ingress metrics available on worker nodes")
		return false, nil
	}

	nodeCount := int(vec[0].Value)
	if nodeCount == 0 {
		// No nodes have ingress metrics available
		log.Info("No nodes have ingress metrics available")
		return false, nil
	}

	// At least one node has ingress metrics available
	return true, nil
}

// TerminateInstance terminates the Instance.
func (r *InstanceInactiveTerminationReconciler) TerminateInstance(ctx context.Context, instance *clv1alpha2.Instance) error {
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
func (r *InstanceInactiveTerminationReconciler) SendNotification(ctx context.Context, instance *clv1alpha2.Instance, userEmail string) error {
	log := ctrl.LoggerFrom(ctx).WithName("notification-email-instance")
	log.Info("sending email notification to user", "instance", instance.Name, "email", userEmail)
	emailBody := fmt.Sprintf("Dear user,\n\n"+
		"Your instance %s has been inactive for a while.\n"+
		"We will terminate it if you do not use it anymore.\n\n"+
		"Please log in to your instance if you wish to keep it running.\n\n"+
		"Best regards,\n"+
		"CrownLabs Team", instance.Name)
	err := r.MailClient.SendMail([]string{userEmail}, "CrownLabs Instance Termination Alert", emailBody)
	if err != nil {
		log.Error(err, "failed sending email notification")
		return err
	}

	// increment the number of termination alerts
	newNumberOfAlerts, err := r.IncrementAnnotation(ctx, instance.ObjectMeta.Annotations[alertAnnotation])
	if err != nil {
		log.Error(err, "failed incrementing annotation")
		return err
	}
	instance.ObjectMeta.Annotations[alertAnnotation] = newNumberOfAlerts
	// update the status of the instance
	if err := r.Update(ctx, instance); err != nil {
		log.Error(err, "failed updating instance annotations")
		return err
	}

	return nil
}

// GetTenantFromInstance retrieves the Tenant object associated with the Instance.
func (r *InstanceInactiveTerminationReconciler) GetTenantFromInstance(ctx context.Context, instance *clv1alpha2.Instance) (clv1alpha2.Tenant, error) {
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

// retrieve the inactivity timeout from the instance template
// This function should return the inactivity timeout for the instance based on its template.
func (r *InstanceInactiveTerminationReconciler) getInactivityTimeout(ctx context.Context, instance *clv1alpha2.Instance) (string, error) {
	log := ctrl.LoggerFrom(ctx).WithName("get-inactivity-timeout")
	// retrieve the template from the instance
	var template clv1alpha2.Template
	if err := r.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.Template.Name,
		Namespace: instance.Spec.Template.Namespace,
	}, &template); err != nil {
		log.Error(err, "Unable to fetch the instance template.")
		return "", err
	}

	templateInactivityTimeout := template.Spec.InactivityTimeout
	return templateInactivityTimeout, nil
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

func isInstanceExpired(creationTimestamp string, lifespan float64) (bool, error) {
	created, err := time.Parse(time.RFC3339, creationTimestamp)
	if err != nil {
		return false, err
	}
	duration := time.Since(created).Seconds()
	return duration > lifespan, nil
}

func (r *InstanceInactiveTerminationReconciler) deleteStaleInstances(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
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
