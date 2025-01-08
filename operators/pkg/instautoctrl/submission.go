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
	"fmt"
	"strconv"

	batch "k8s.io/api/batch/v1"
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

// InstanceSubmissionReconciler watches for instances to be terminated.
type InstanceSubmissionReconciler struct {
	client.Client
	EventsRecorder     record.EventRecorder
	Scheme             *runtime.Scheme
	NamespaceWhitelist metav1.LabelSelector
	ContainerEnvOpts   forge.ContainerEnvOpts
	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// SetupWithManager registers a new controller for InstanceSubmissionReconciler resources.
func (r *InstanceSubmissionReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Instance{}).
		Owns(&batch.Job{}).
		Named("instance-submission").
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "InstanceSubmission")).
		Complete(r)
}

// Reconcile reconciles the status of the InstanceSnapshot resource.
func (r *InstanceSubmissionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	log := ctrl.LoggerFrom(ctx, "instance", req.NamespacedName)
	dbgLog := log.V(utils.LogDebugLevel)
	tracer := trace.New("reconcile", trace.Field{Key: "instance", Value: req.NamespacedName})
	ctx = ctrl.LoggerInto(trace.ContextWithTrace(ctx, tracer), log)

	defer tracer.LogIfLong(utils.LongThreshold())

	// Get the instance object.
	var instance clv1alpha2.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving instance")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	tracer.Step("instance retrieved")

	// Skip if the instance has not to be submitted.
	if proceed, err := r.CheckLabelSelectors(ctx, &instance); !proceed {
		if err == nil {
			dbgLog.Info("reconciliation skipped", "reason", "label selector not matching")
		}
		return ctrl.Result{}, err
	}
	tracer.Step("labels checked")

	environment, err := RetrieveEnvironment(ctx, r.Client, &instance)
	if err != nil {
		log.Error(err, "failed retrieving environment")
		return ctrl.Result{}, err
	}
	tracer.Step("retrieved the instance environment")

	if err := CheckEnvironmentValidity(&instance, environment); err != nil {
		instance.SetLabels(forge.InstanceAutomationLabelsOnSubmission(instance.GetLabels(), false))
		dbgLog.Info("instance submission aborted")
	} else {
		jobStatus, err := r.EnforceInstanceSubmissionJob(ctx, &instance, environment)
		switch {
		case err == nil && jobStatus.Succeeded == 0: // the job hasn't been completed yet
			tracer.Step("job enforced")
			dbgLog.Info("waiting for job completion")
			return ctrl.Result{}, nil
		case err == nil && jobStatus.Succeeded > 0: // the job has been completed successfully
			if jobStatus.CompletionTime != nil {
				instance.Status.Automation.SubmissionTime = *jobStatus.CompletionTime
			} else {
				instance.Status.Automation.SubmissionTime = metav1.Now()
			}
			if err := r.Status().Update(ctx, &instance); err != nil {
				log.Error(err, "failed updating instance status")
				return ctrl.Result{}, err
			}
			tracer.Step("instance status updated")
			log.Info("instance submission completed")
			instance.SetLabels(forge.InstanceAutomationLabelsOnSubmission(instance.GetLabels(), true))
		default: // any other error occurred
			return ctrl.Result{}, err
		}
	}

	if err := r.Update(ctx, &instance); err != nil {
		log.Error(err, "failed updating instance labels")
		return ctrl.Result{}, err
	}
	tracer.Step("instance labels updated")

	dbgLog.Info("instance submission enforced")
	return ctrl.Result{}, nil
}

// EnforceInstanceSubmissionJob ensures that the submission job for the given instance is present.
func (r *InstanceSubmissionReconciler) EnforceInstanceSubmissionJob(ctx context.Context, instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) (jobStatus *batch.JobStatus, err error) {
	// Get the submission job.
	submitterName := "submitter"
	job := batch.Job{ObjectMeta: forge.ObjectMetaWithSuffix(instance, submitterName)}

	jobSpec := forge.SubmissionJobSpec(instance, environment, &r.ContainerEnvOpts)

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, &job, func() error {
		if job.CreationTimestamp.IsZero() {
			job.Spec = jobSpec
		}
		job.SetLabels(forge.InstanceComponentLabels(instance, submitterName))
		return ctrl.SetControllerReference(instance, &job, r.Scheme)
	})

	if err != nil {
		return nil, fmt.Errorf("failed ensuring submission job (operation=%s): %w", op, err)
	}

	return &job.Status, nil
}

// CheckLabelSelectors checks whether the given instance is eligible for reconciliation.
func (r *InstanceSubmissionReconciler) CheckLabelSelectors(ctx context.Context, instance *clv1alpha2.Instance) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("selectors-check")

	if !utils.CheckSingleLabel(instance, forge.InstanceSubmissionSelectorLabel, strconv.FormatBool(true)) {
		log.V(utils.LogDebugLevel).Info("skipping instance", "reason", "label selector not matching", "label", forge.InstanceSubmissionSelectorLabel, "expected-value", strconv.FormatBool(true), "current-labels", instance.GetLabels())
		return false, nil
	} else if utils.CheckSingleLabel(instance, forge.InstanceSubmissionCompletedLabel, strconv.FormatBool(true)) {
		log.V(utils.LogDebugLevel).Info("skipping instance", "reason", "label selector not matching", "label", forge.InstanceSubmissionCompletedLabel, "expected-value", "!"+strconv.FormatBool(true), "current-labels", instance.GetLabels())
		return false, nil
	}

	// Check the selector label over namespace, in order to know whether to perform or not reconciliation.
	return utils.CheckSelectorLabel(ctx, r.Client, instance.GetNamespace(), r.NamespaceWhitelist.MatchLabels)
}
