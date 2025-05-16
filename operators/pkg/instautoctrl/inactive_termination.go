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
	"net/smtp"
	"strconv"
	"time"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/trace"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// InstanceInactiveTerminationReconciler watches for instances to be terminated.
type InstanceInactiveTerminationReconciler struct {
	client.Client
	EventsRecorder              record.EventRecorder
	Scheme                      *runtime.Scheme
	NamespaceWhitelist          metav1.LabelSelector
	StatusCheckRequestTimeout   time.Duration
	InstanceStatusCheckInterval time.Duration
	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

var alertAnnotation = "crownlabs.polito.it/number-alerts-sent"

const maxNumberOfAlerts = 4
const terminationInterval = 2 * time.Minute

var mailClient = &MailClient{
	SMTPServer: "smtp.polito.it",
	SMTPPort:   587,
	auth:       smtp.PlainAuth("identity", "crownlabs", "password", "smtp.polito.it"),
	From:       "crownlabs@polito.it",
}

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
				return false
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

	// if dbgLog.Enabled() {
	// 	defer tracer.Log()
	// } else {
	// 	defer tracer.LogIfLong(r.StatusCheckRequestTimeout / 2)
	// }

	// Get the instance object.
	var instance clv1alpha2.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving instance")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	tracer.Step("instance retrieved")

	// Skip if the instance has not to be terminated.
	if !utils.CheckSingleLabel(&instance, forge.InstanceTerminationSelectorLabel, strconv.FormatBool(true)) {
		dbgLog.Info("skipping instance", "reason", "label selector not matching", "label", forge.InstanceTerminationSelectorLabel)
		return ctrl.Result{}, nil
	}

	// Check the selector label, in order to know whether to perform or not reconciliation.
	if proceed, err := utils.CheckSelectorLabel(ctx, r.Client, instance.GetNamespace(), r.NamespaceWhitelist.MatchLabels); !proceed {
		if err != nil {
			err = fmt.Errorf("failed checking selector label: %w", err)
		}
		return ctrl.Result{}, err
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

	terminate, err := r.CheckInstanceTermination(ctx, &instance)
	if err != nil {
		log.Error(err, "failed checking instance termination")
		return ctrl.Result{}, err
	}
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

		if numberAlertSent < maxNumberOfAlerts {
			r.SendNotification(ctx, &instance, user.Spec.Email)
		} else if numberAlertSent >= maxNumberOfAlerts && instance.Spec.Running == true {
			r.TerminateInstance(ctx, &instance)
		}
	} else {
		log.Info("instance is not yet to be terminated", "instance", instance.Name)
	}

	dbgLog.Info("requeueing instance")
	// TODO: where to put delay time? general for all crownlab machines?
	return ctrl.Result{RequeueAfter: terminationInterval}, nil
}

// CheckInstanceTermination checks if the Instance has to be terminated.
func (r *InstanceInactiveTerminationReconciler) CheckInstanceTermination(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("check-instance-termination")

	if time.Since(instance.CreationTimestamp.Time) > 2*time.Minute {
		return true, nil
	}

	promURL := r.getPrometheusURL()

	config := api.Config{
		Address: promURL,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return false, fmt.Errorf("error creating prometheus client: %w", err)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Check Prometheus health first
	healthy, err := r.isPrometheusHealthy(ctx, v1api)
	if err != nil || !healthy {
		log.Info("Prometheus is not healthy", "error", err)
		return false, err
	}

	// Get instance activity data
	tenantNS := instance.Namespace
	instanceName := instance.Name

	// Proper PromQL query to check for activity in the last 2 weeks
	query := fmt.Sprintf(`sum(changes(nginx_ingress_controller_requests{exported_namespace="%s", exported_service="%s"}[2w])) > 0`,
		tenantNS, instanceName)

	result, warnings, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		return false, fmt.Errorf("error querying prometheus: %w", err)
	}

	if len(warnings) > 0 {
		log.Info("Prometheus query warnings", "warnings", warnings)
	}

	vec, ok := result.(model.Vector)
	if !ok {
		return false, fmt.Errorf("unexpected result format: %T", result)
	}

	// If we got any results with value > 0, there's been activity
	for _, sample := range vec {
		if sample.Value > 0 {
			return false, nil
		}
	}

	// No activity detected
	log.Info("No activity detected", "instance", instanceName)
	return true, nil
}

// isPrometheusHealthy checks if Prometheus and required metrics are available
func (r *InstanceInactiveTerminationReconciler) isPrometheusHealthy(ctx context.Context, v1api v1.API) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("prometheus-health")

	// Verify connection to Prometheus health endpoint
	promURL := r.getPrometheusURL()
	healthEndpoint := fmt.Sprintf("%s/-/healthy", promURL)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthEndpoint, nil)
	if err != nil {
		log.Error(err, "Failed to create HTTP request for Prometheus health check")
		return false, fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err, "Failed to connect to Prometheus health endpoint")
		return false, fmt.Errorf("prometheus health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Info("Prometheus health check returned non-OK status", "statusCode", resp.StatusCode)
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

func (r *InstanceInactiveTerminationReconciler) getPrometheusURL() string {
	// TODO: from inside the cluster use "http://prometheus-kube-prometheus-prometheus.monitoring.svc.cluster.local"
	// TODO: why port 80 from inside? Are we reaching this from inside or outside? We need internal unauthenticated access
	return "http://172.19.0.120:80"
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

	//TODO check all the environments?
	var environment = template.Spec.EnvironmentList[0]
	if environment.Persistent {
		log.Info("Stopping persistent instance...")
		instance.Spec.Running = false
		return r.Update(ctx, instance)
	} else {
		log.Info("Deleting non-persistent instance...")
		return r.Delete(ctx, instance)
	}

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
	mailClient.SendMail([]string{userEmail}, "CrownLabs Instance Termination Alert", emailBody)

	// increment the number of termination alerts
	newNumberOfAlerts, err := r.IncrementAnnotation(ctx, instance.ObjectMeta.Annotations[alertAnnotation])
	if err != nil {
		log.Error(err, "failed incrementing annotation")
	}
	instance.ObjectMeta.Annotations[alertAnnotation] = newNumberOfAlerts
	// update the status of the instance
	if err := r.Update(ctx, instance); err != nil {
		log.Error(err, "failed updating instance annotations")
	}

	return nil
}

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
