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

// Package instautoctrl contains the controller for Instance Termination and Submission automations.
package instautoctrl

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/trace"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// InstanceTerminationReconciler watches for instances to be terminated.
type InstanceTerminationReconciler struct {
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

// SetupWithManager registers a new controller for InstanceTerminationReconciler resources.
func (r *InstanceTerminationReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Instance{}).
		Named("instance-termination").
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "InstanceTermination")).
		Complete(r)
}

// Reconcile reconciles the status of the InstanceSnapshot resource.
func (r *InstanceTerminationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	log := ctrl.LoggerFrom(ctx, "instance", req.NamespacedName)
	dbgLog := log.V(utils.LogDebugLevel)
	tracer := trace.New("reconcile", trace.Field{Key: "instance", Value: req.NamespacedName})
	ctx = ctrl.LoggerInto(trace.ContextWithTrace(ctx, tracer), log)

	if dbgLog.Enabled() {
		defer tracer.Log()
	} else {
		defer tracer.LogIfLong(r.StatusCheckRequestTimeout / 2)
	}

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

	// Check if the instance has to be terminated.
	terminate, err := r.CheckInstanceTermination(ctx, &instance)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed checking instance termination: %w", err)
	}

	tracer.Step("status checked checked")

	if terminate {
		err := r.TerminateInstance(ctrl.LoggerInto(ctx, dbgLog), &instance)
		if err != nil {
			err = fmt.Errorf("failed terminating instance: %w", err)
		} else {
			tracer.Step("instance terminated")
			log.Info("instance terminated")
		}
		return ctrl.Result{}, err
	}

	tracer.Step("instance requeued")

	dbgLog.Info("requeueing instance")
	return ctrl.Result{RequeueAfter: r.InstanceStatusCheckInterval}, nil
}

// CheckInstanceTermination checks if the Instance has to be terminated.
func (r *InstanceTerminationReconciler) CheckInstanceTermination(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	if instance.Spec.CustomizationUrls == nil {
		return false, errors.New("customization urls field is not set for Instance")
	}

	statusCheckURL := instance.Spec.CustomizationUrls.StatusCheck
	if statusCheckURL == "" {
		return false, errors.New("status check url field is not set for Instance")
	}

	log := ctrl.LoggerFrom(ctx).WithName("status-check")
	log.Info("performing instance status check")

	statusCheckReponse := StatusCheckResponse{}
	statusCode, err := utils.HTTPGetJSONIntoStruct(ctx, statusCheckURL, &statusCheckReponse, r.StatusCheckRequestTimeout)
	if statusCode != http.StatusNotFound && err != nil {
		return false, err
	}

	instance.Status.Automation.LastCheckTime = metav1.Now()
	if statusCode == http.StatusOK {
		instance.Status.Automation.TerminationTime = metav1.Time{Time: statusCheckReponse.Deadline}
	} else if statusCode == http.StatusNotFound {
		instance.Status.Automation.TerminationTime = metav1.Now()
	}

	if err := r.Status().Update(ctx, instance); err != nil {
		log.Error(err, "failed updating instance status")
		return false, err
	}

	switch statusCode {
	case http.StatusOK:
		log.Info("termination not required")
		return false, nil
	case http.StatusNotFound:
		log.Info("termination required")
		return true, nil
	default:
		return false, fmt.Errorf("failed: unexpected status code %d, retrieved error='%s'", statusCode, statusCheckReponse.Error)
	}
}

// TerminateInstance terminates the Instance.
func (r *InstanceTerminationReconciler) TerminateInstance(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx).WithName("termination")
	log.Info("terminating instance")

	submissionRequired := false

	environment, err := RetrieveEnvironment(ctx, r.Client, instance)
	if err != nil {
		log.Info("failed retrieving environment", "error", err)
		return err
	}

	if err := CheckEnvironmentValidity(instance, environment); err != nil {
		log.Info("instance not eligible for submission", "error", err)
	} else {
		submissionRequired = true
		log.Info("submission required")
	}

	instance.SetLabels(forge.InstanceAutomationLabelsOnTermination(instance.GetLabels(), submissionRequired))

	instance.Spec.Running = false

	return r.Update(ctx, instance)
}
