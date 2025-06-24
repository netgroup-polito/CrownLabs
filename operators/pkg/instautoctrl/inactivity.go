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
)

// InstanceInactiveTerminationReconciler watches for instances to be terminated.
type InstanceInactiveTerminationReconciler struct {
	client.Client
	EventsRecorder            record.EventRecorder
	Scheme                    *runtime.Scheme
	NamespaceWhitelist        metav1.LabelSelector
	StatusCheckRequestTimeout time.Duration
	InstanceMaxNumberOfAlerts int
	MailClient                *utils.MailClient
	PrometheusURL             string
	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// SetupWithManager registers a new controller for InstanceTerminationReconciler resources.
func (r *InstanceInactiveTerminationReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Instance{}).
		Watches(
			&clv1alpha2.Template{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				template, ok := obj.(*clv1alpha2.Template)
				if !ok || template.Spec.InactivityTimeout == neverTimeoutValue {
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

			fmt.Println("Namespace changed:", namespace.Name)
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
			fmt.Println("Enqueued requests for namespace:", namespace.Name, "with", len(requests), "instances", requests)
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

	// Get the instance object.
	var instance clv1alpha2.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving instance")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	tracer.Step("instance retrieved")

	// Get the template associated with the instance
	var template clv1alpha2.Template
	if err := r.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.Template.Name,
		Namespace: instance.Namespace,
	}, &template); err != nil {
		log.Error(err, "Unable to fetch the instance template.")
		return ctrl.Result{}, fmt.Errorf("failed to fetch instance template %s/%s: %w", instance.Namespace, instance.Spec.Template.Name, err)
	}
	tracer.Step("template retrieved")

	// Add the instance and template to the context
	ctx, _ = pkgcontext.InstanceInto(ctx, &instance)
	ctx, _ = pkgcontext.TemplateInto(ctx, &template)

	// Get inactivityTimeout from the template
	inactivityTimeout := template.Spec.InactivityTimeout
	// If set to neverTimeoutValue , return without rescheduling
	if inactivityTimeout == neverTimeoutValue {
		log.Info("Instance marked as never delete", "name", instance.GetName(), "namespace", instance.GetNamespace())
		return ctrl.Result{}, nil
	}

	// Check the selector label, in order to know whether to perform or not reconciliation.
	if proceed, err := utils.CheckSelectorLabel(ctx, r.Client, instance.GetNamespace(), r.NamespaceWhitelist.MatchLabels); !proceed {
		if err != nil {
			err = fmt.Errorf("failed checking selector label: %w", err)
		}
		return ctrl.Result{}, err
	}

	var namespace corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Namespace}, &namespace); err != nil {
		log.Error(err, "failed retrieving instance namespace", "instance", instance.Name)
		return ctrl.Result{}, err
	}

	// check the namespace labels, in order to know whether to perform or not reconciliation on a specific namespace.
	if stop := utils.CheckSingleLabel(&namespace, forge.InstanceInactivityIgnoreNamespace, strconv.FormatBool(true)); stop {
		log.Info("label present, skipping inactivity reconciliation for namespace", "namespace", instance.Namespace, "label", forge.InstanceInactivityIgnoreNamespace)
		return ctrl.Result{}, nil
	}

	tracer.Step("labels checked")

	// add annotations to the instance if not present
	if instance.ObjectMeta.Annotations == nil {
		instance.ObjectMeta.Annotations = make(map[string]string)
	}
	patch := client.MergeFrom(instance.DeepCopy())
	// Check and set the alert annotation if not present
	if _, ok := instance.ObjectMeta.Annotations[forge.AlertAnnotation]; !ok {
		log.Info("adding annotation to instance for the first time", "annotation", forge.AlertAnnotation)
		instance.ObjectMeta.Annotations[forge.AlertAnnotation] = "0"
	}
	// Check and set the last activity annotation if not present
	if _, ok := instance.ObjectMeta.Annotations[forge.LastActivityAnnotation]; !ok {
		log.Info("adding annotation to instance for the first time", "annotation", forge.LastActivityAnnotation)
		instance.ObjectMeta.Annotations[forge.LastActivityAnnotation] = time.Now().Format(time.RFC3339)
	}
	// Apply the patch
	_, ok1 := instance.ObjectMeta.Annotations[forge.AlertAnnotation]
	_, ok2 := instance.ObjectMeta.Annotations[forge.LastActivityAnnotation]
	if !ok1 || !ok2 {
		if err := r.Patch(ctx, &instance, patch); err != nil {
			log.Error(err, "failed updating instance annotations")
			return ctrl.Result{}, err
		}
	}

	tracer.Step("annotations checked")

	// Create the Prometheus client
	promClient, err := api.NewClient(
		api.Config{
			Address: r.PrometheusURL,
		},
	)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed creating Prometheus client: %w", err)
	}
	// update the last login time of the instance based on the Prometheus data
	if err := r.UpdateInstanceLastLogin(ctx, inactivityTimeout, &promClient); err != nil {
		log.Error(err, "failed updating last login time of the instance")
		return ctrl.Result{}, err
	}

	tracer.Step("instance last login updated")

	// check for inactivity and decide whether to terminate the instance or not
	remainingTime, err := r.CheckInstanceTermination(ctx, inactivityTimeout)
	if err != nil {
		log.Error(err, "failed checking instance termination")
		return ctrl.Result{}, err
	}

	log.Info("instance termination check", "remainingTime", remainingTime.String(), "instance", instance.Name)
	tracer.Step("Inactive termination check done")

	// If the remaining time is less than or equal to 0, the instance is considered inactive
	if remainingTime <= 0 {
		// retrieve the user owner of the instance
		user, err := GetTenantFromInstance(ctx, r.Client)
		if err != nil {
			log.Error(err, "failed retrieving user from instance")
			return ctrl.Result{}, err
		}

		ctx, _ = pkgcontext.TenantInto(ctx, user)

		// send notification to the user
		numberAlertSent, err := strconv.Atoi(instance.ObjectMeta.Annotations[forge.AlertAnnotation])
		if err != nil {
			log.Error(err, "failed converting string of alerts sent in int number")
			return ctrl.Result{}, err
		}

		if numberAlertSent < r.InstanceMaxNumberOfAlerts {
			err := SendInactivityNotification(ctx, r.MailClient)
			if err != nil {
				log.Error(err, "failed sending notification email to user", "email", user.Spec.Email)
				return ctrl.Result{}, err
			}
			// increment the number of termination alerts
			newNumberOfAlerts, err := r.IncrementAnnotation(ctx, instance.ObjectMeta.Annotations[forge.AlertAnnotation])
			if err != nil {
				log.Error(err, "failed incrementing annotation")
				return ctrl.Result{}, err
			}
			instance.ObjectMeta.Annotations[forge.AlertAnnotation] = newNumberOfAlerts
			// update the status of the instance
			if err := r.Update(ctx, &instance); err != nil {
				log.Error(err, "failed updating instance annotations")
				return ctrl.Result{}, err
			}
		} else if numberAlertSent >= r.InstanceMaxNumberOfAlerts && instance.Spec.Running {
			err := r.TerminateInstance(ctx)
			if err != nil {
				log.Error(err, "failed terminating instance", "instance", instance.Name)
				return ctrl.Result{}, err
			}
		}
	} else {
		// TODO: remove
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
func (r *InstanceInactiveTerminationReconciler) UpdateInstanceLastLogin(ctx context.Context, inactivityTimeout string, promClient *api.Client) error {
	log := ctrl.LoggerFrom(ctx).WithName("update-instance-last-login")
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}
	v1api := v1.NewAPI(*promClient)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Check Prometheus health first
	healthy, err := r.IsPrometheusHealthy(ctx, v1api)
	if err != nil || !healthy {
		log.Info("Prometheus is not healthy", "error", err)
		return err
	}

	intervalDuration, err := time.ParseDuration(inactivityTimeout)
	if err != nil {
		log.Error(err, "failed parsing inactivity timeout duration")
		return err
	}

	// Get instance activity data
	queryNginx := fmt.Sprintf(`nginx_ingress_controller_requests{exported_namespace=%q, exported_service=%q}`, instance.Namespace, instance.Name)
	lastActivityTimeNginx, errNginx := GetLastActivityTime(queryNginx, v1api, intervalDuration)

	querySSH := fmt.Sprintf(`bation_ssh_conntections{namespace=%q, destination_Ip=%q}`, instance.Namespace, instance.Status.IP)
	lastActivityTimeSSH, errSSH := GetLastActivityTime(querySSH, v1api, intervalDuration)

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

// CheckInstanceTermination checks if the Instance has to be terminated.
func (r *InstanceInactiveTerminationReconciler) CheckInstanceTermination(ctx context.Context, inactivityTimeout string) (time.Duration, error) {
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
	timeoutDuration, err := time.ParseDuration(inactivityTimeout)
	if err != nil {
		log.Error(err, "failed parsing inactivity timeout duration")
		return 0, err
	}

	// Check if the instance has been inactive for longer than the timeout duration
	remainingTime = timeoutDuration - time.Since(lastLogin)
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
