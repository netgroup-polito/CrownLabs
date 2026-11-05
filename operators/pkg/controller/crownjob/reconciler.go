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

// Package crownjob groups the functionalities related to the CrownJob controller, which runs each CrownJob as a
// batch Job queued by Kueue, and to the controller of the job queues (the Kueue objects) of the Workspaces.
package crownjob

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/trace"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	ctrlcommon "github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// Reconciler reconciles a CrownJob object.
type Reconciler struct {
	client.Client
	// APIReader reads, without cache, the objects that the controller does not watch: the Instance and the objects
	// it refers to (once, before the Job is created), the pod of the Job and its Kueue Workload.
	APIReader      client.Reader
	TargetLabel    ctrlcommon.KVLabel
	EventsRecorder record.EventRecorder
	// RuntimeClasses maps a resource name to the RuntimeClass that a job requesting it must use.
	RuntimeClasses map[string]string
	// DefaultMaxRunTime is the maximum run time of a job when neither its Template nor its Workspace set one.
	DefaultMaxRunTime time.Duration
	// LogHelperImage is the image of the Jobs that save the log of each job pod into its run folder.
	LogHelperImage string
	// FinishedJobTTL is how long a Job (the main Job, or a log helper) is kept after it ends, with its pods: then they
	// no longer keep the disk of the instance in use. The CrownJob is kept, with its terminal status.
	FinishedJobTTL time.Duration

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// SetupWithManager registers a new controller for CrownJob resources. It watches the CrownJobs and their Jobs (the main
// Job, with the name of the CrownJob, and the log helpers), and the metadata of the Deployments of the instances, to
// wake up the jobs that wait for them: pods and Workloads are read when needed, to keep the cache of the operator small.
// The field index of the CrownJobs on their Instance is registered by InstanceReconciler.SetupWithManager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.CrownJob{}).
		Owns(&batchv1.Job{}).
		Watches(&appsv1.Deployment{}, handler.EnqueueRequestsFromMapFunc(r.instanceDeploymentToCrownJobs), builder.OnlyMetadata,
			builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
				_, found := obj.GetLabels()[forge.LabelInstanceKey]
				return found
			}))).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "CrownJob")).
		Complete(r)
}

// Reconcile reconciles the state of a CrownJob resource.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	log := ctrl.LoggerFrom(ctx, "crownjob", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	tracer := trace.New("reconcile", trace.Field{Key: "crownjob", Value: req.NamespacedName})
	ctx = trace.ContextWithTrace(ctx, tracer)
	defer tracer.LogIfLong(utils.LongThreshold())

	// Get the CrownJob object.
	var crownJob clv1alpha2.CrownJob
	if err = r.Get(ctx, req.NamespacedName, &crownJob); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving crownjob")
		}
		// Reconcile was triggered by a delete request and object is already deleted.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check the target label, in order to know whether to perform or not reconciliation.
	if proceed, err := utils.CheckNamespaceTargetLabel(ctx, r.Client, crownJob.GetNamespace(), r.TargetLabel); !proceed {
		// If there was an error while checking, show the error and try again.
		if err != nil {
			log.Error(err, "failed checking target label")
			return ctrl.Result{}, err
		}
		log.V(utils.LogDebugLevel).Info("crownjob is not responsibility of this controller, skipping reconcile")
		return ctrl.Result{}, nil
	}

	// A CrownJob being deleted is left to the garbage collector.
	if !crownJob.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}
	// A terminal phase never changes again: only its log helpers are still looked at.
	if isTerminal(crownJob.Status.Phase) {
		return r.reconcileTerminal(ctx, &crownJob)
	}

	// Defer the function to update the CrownJob status depending on what is observed below,
	// and to record an event on every phase change. The patch is refused if the CrownJob changed meanwhile
	// (e.g., canceled by the instance controller), so that a terminal phase is never overwritten.
	defer func(original, updated *clv1alpha2.CrownJob) {
		if reflect.DeepEqual(original.Status, updated.Status) {
			return
		}
		if err2 := r.Status().Patch(ctx, updated, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{})); err2 != nil {
			if kerrors.IsConflict(err2) {
				log.Info("the crownjob changed while it was reconciled, retrying")
				result, err = ctrl.Result{RequeueAfter: requeueConflict}, nil
				return
			}
			log.Error(err2, "failed to update the crownjob status")
			err = err2
			return
		}
		tracer.Step("crownjob status updated")
		log.Info("crownjob status updated", "phase", updated.Status.Phase, "reason", updated.Status.Reason)
		if original.Status.Phase != updated.Status.Phase {
			eventType := corev1.EventTypeNormal
			if updated.Status.Phase == clv1alpha2.CrownJobPhaseFailed {
				eventType = corev1.EventTypeWarning
			}
			r.EventsRecorder.Event(updated, eventType, updated.Status.Reason, updated.Status.Message)
		}
	}(crownJob.DeepCopy(), &crownJob)

	// Once the Job exists, it is never rebuilt: its state is read.
	var job batchv1.Job
	if err = r.Get(ctx, req.NamespacedName, &job); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving the job")
			return ctrl.Result{}, err
		}
		if crownJob.Status.RunFolder == "" {
			return r.createJob(ctx, &crownJob)
		}
		// The Job was created before: the cache may not have seen it yet, so check with a live read.
		if err = r.APIReader.Get(ctx, req.NamespacedName, &job); kerrors.IsNotFound(err) {
			helpers, err := r.logHelpers(ctx, &crownJob)
			if err != nil {
				return ctrl.Result{}, err
			}
			setStatus(&crownJob, clv1alpha2.CrownJobPhaseFailed, ReasonError, msgJobDeleted+logSentence(&crownJob, nil, helpers))
			return ctrl.Result{}, nil
		} else if err != nil {
			log.Error(err, "failed retrieving the job")
			return ctrl.Result{}, err
		}
	}

	if !metav1.IsControlledBy(&job, &crownJob) {
		setStatus(&crownJob, clv1alpha2.CrownJobPhaseFailed, ReasonCannotStart, fmt.Sprintf(msgJobNameTaken, job.Name))
		return ctrl.Result{}, nil
	}

	return r.observeJob(ctx, &crownJob, &job)
}

// createJob validates the CrownJob and, if it can start, applies the LocalQueue of its Workspace and creates its Job.
func (r *Reconciler) createJob(ctx context.Context, crownJob *clv1alpha2.CrownJob) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	inputs, err := r.gatherJobInputs(ctx, crownJob)
	var cannotStart *forge.CannotStartError
	switch {
	case errors.As(err, &cannotStart):
		setStatus(crownJob, clv1alpha2.CrownJobPhaseFailed, ReasonCannotStart, cannotStart.Error())
		return ctrl.Result{}, nil
	case errors.Is(err, errInstanceNotReady):
		setStatus(crownJob, clv1alpha2.CrownJobPhaseWaiting, ReasonInstanceNotReady, fmt.Sprintf(msgInstanceNotReady, inputs.instance.Name))
		return ctrl.Result{RequeueAfter: requeueInstanceNotReady}, nil
	case errors.Is(err, errDiskInUse):
		setStatus(crownJob, clv1alpha2.CrownJobPhaseWaiting, ReasonDiskInUse, fmt.Sprintf(msgDiskInUse, inputs.instance.Name))
		if errors.Is(err, errInstanceStopping) {
			return ctrl.Result{RequeueAfter: requeueDiskReleasing}, nil
		}
		return ctrl.Result{RequeueAfter: requeueDiskInUse}, nil
	case errors.Is(err, errQueueNotReady):
		setStatus(crownJob, clv1alpha2.CrownJobPhaseWaiting, ReasonQueueNotReady, fmt.Sprintf(msgQueueNotReady, inputs.workspace.Name))
		return ctrl.Result{RequeueAfter: requeueQueueNotReady}, nil
	case err != nil:
		log.Error(err, "failed gathering the job inputs")
		return ctrl.Result{}, err
	}

	job, err := forge.CrownJobBatchJob(crownJob, &inputs.opts)
	if err == nil {
		// 8. Every resource the pod requests must be covered by the queue, within maxAtOnce.
		err = forge.CrownJobCheckQueueLimits(inputs.workspace.Name, inputs.workspace.Spec.QueueLimits,
			forge.PodEffectiveRequests(&job.Spec.Template.Spec))
	}
	if errors.As(err, &cannotStart) {
		setStatus(crownJob, clv1alpha2.CrownJobPhaseFailed, ReasonCannotStart, cannotStart.Error())
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "failed forging the job")
		return ctrl.Result{}, err
	}

	// The Instance is kept while the job is not terminal (see InstanceReconciler), and read again after the finalizer
	// is added: an Instance deleted meanwhile no longer gets a job, which would keep it for the whole run.
	if err := r.addInstanceFinalizer(ctx, inputs.instance); err != nil {
		if kerrors.IsConflict(err) {
			return ctrl.Result{RequeueAfter: requeueConflict}, nil
		}
		log.Error(err, "failed to add the finalizer to the instance", "instance", klog.KObj(inputs.instance))
		return ctrl.Result{}, err
	}
	var instance clv1alpha2.Instance
	if found, err := r.read(ctx, client.ObjectKeyFromObject(inputs.instance), &instance); err != nil {
		log.Error(err, "failed retrieving the instance")
		return ctrl.Result{}, err
	} else if !found {
		setStatus(crownJob, clv1alpha2.CrownJobPhaseFailed, ReasonCannotStart, fmt.Sprintf(msgInstanceGone, inputs.instance.Name))
		return ctrl.Result{}, nil
	} else if !instance.DeletionTimestamp.IsZero() {
		setStatus(crownJob, clv1alpha2.CrownJobPhaseFailed, ReasonCannotStart, fmt.Sprintf(msgInstanceDeleting, instance.Name))
		return ctrl.Result{}, nil
	}

	localQueue := forge.KueueLocalQueue(inputs.workspace, crownJob.Namespace)
	if err := r.Patch(ctx, localQueue, client.Apply, client.FieldOwner(forge.CrownJobFieldManager), client.ForceOwnership); err != nil {
		log.Error(err, "failed to apply the local queue", "localqueue", klog.KObj(localQueue))
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, job); err != nil {
		switch {
		case kerrors.IsAlreadyExists(err):
			// Created by a previous reconcile whose status update failed: observe it at the next one.
			return ctrl.Result{RequeueAfter: time.Second}, nil
		case kerrors.IsInvalid(err):
			setStatus(crownJob, clv1alpha2.CrownJobPhaseFailed, ReasonCannotStart, fmt.Sprintf(msgJobInvalid, shorten(err.Error())))
			return ctrl.Result{}, nil
		default:
			log.Error(err, "failed to create the job", "job", klog.KObj(job))
			return ctrl.Result{}, err
		}
	}
	log.Info("job created", "job", klog.KObj(job), "queue", job.Labels[forge.KueueQueueNameLabel])

	crownJob.Status.RunFolder = forge.CrownJobRunFolder(crownJob.Name)
	setStatus(crownJob, clv1alpha2.CrownJobPhaseWaiting, ReasonInQueue, fmt.Sprintf(msgInQueue, inputs.workspace.Name))
	return ctrl.Result{RequeueAfter: requeueWaiting}, nil
}

// instanceDeploymentToCrownJobs maps (an event of) the Deployment of an instance to the CrownJobs of that instance that
// have not started yet (no phase, or Waiting): among them, those waiting for its disk to be released (DiskInUse) or for
// the Deployment to exist (InstanceNotReady). The CrownJobs are listed from the cache, through their field index.
func (r *Reconciler) instanceDeploymentToCrownJobs(ctx context.Context, obj client.Object) []reconcile.Request {
	instance := obj.GetLabels()[forge.LabelInstanceKey]
	if instance == "" {
		return nil
	}
	var crownJobs clv1alpha2.CrownJobList
	if err := r.List(ctx, &crownJobs, client.InNamespace(obj.GetNamespace()), client.MatchingFields{CrownJobInstanceIndex: instance}); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed listing the crownjobs of the instance", "deployment", klog.KObj(obj))
		return nil
	}

	var requests []reconcile.Request
	for i := range crownJobs.Items {
		crownJob := &crownJobs.Items[i]
		if ns := crownJob.Spec.Instance.Namespace; ns != "" && ns != crownJob.Namespace {
			continue
		}
		if crownJob.Status.Phase == "" || crownJob.Status.Phase == clv1alpha2.CrownJobPhaseWaiting {
			requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(crownJob)})
		}
	}
	return requests
}

// addInstanceFinalizer adds the CrownJob finalizer to the given Instance, if missing, with a metadata patch that is
// refused if the Instance changed since it was read.
func (r *Reconciler) addInstanceFinalizer(ctx context.Context, instance *clv1alpha2.Instance) error {
	original := instance.DeepCopy()
	if !ctrlutil.AddFinalizer(instance, forge.CrownJobInstanceFinalizer) {
		return nil
	}
	return r.Patch(ctx, instance, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{}))
}

// reconcileTerminal handles a CrownJob in a terminal phase, whose phase and reason never change again: the Job of a
// CrownJob canceled before its admission is deleted (it may have been created just before the cancellation), the
// failures of its log helpers are reported, and the sentence about the log of its first pod is appended to its
// message once that log helper has ended.
func (r *Reconciler) reconcileTerminal(ctx context.Context, crownJob *clv1alpha2.CrownJob) (ctrl.Result, error) {
	if crownJob.Status.Reason == ReasonCanceled {
		var job batchv1.Job
		if err := r.Get(ctx, client.ObjectKeyFromObject(crownJob), &job); err == nil && metav1.IsControlledBy(&job, crownJob) {
			if err := r.Delete(ctx, &job, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil && !kerrors.IsNotFound(err) {
				ctrl.LoggerFrom(ctx).Error(err, "failed to delete the job of a canceled crownjob", "job", klog.KObj(&job))
				return ctrl.Result{}, err
			}
		} else if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
	}

	helpers, err := r.logHelpers(ctx, crownJob)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reportFailedLogHelpers(ctx, crownJob, helpers); err != nil {
		return ctrl.Result{}, err
	}
	return r.appendLogSentence(ctx, crownJob, helpers)
}

// isTerminal returns whether the given phase is terminal.
func isTerminal(phase clv1alpha2.CrownJobPhase) bool {
	return phase == clv1alpha2.CrownJobPhaseSucceeded || phase == clv1alpha2.CrownJobPhaseFailed
}

// setStatus sets the phase, reason and message of the status of the given CrownJob.
func setStatus(crownJob *clv1alpha2.CrownJob, phase clv1alpha2.CrownJobPhase, reason, message string) {
	crownJob.Status.Phase = phase
	crownJob.Status.Reason = reason
	crownJob.Status.Message = message
}

// shorten returns the given message, cut to maxMessageLength characters.
func shorten(message string) string {
	runes := []rune(message)
	if len(runes) <= maxMessageLength {
		return message
	}
	return string(runes[:maxMessageLength-1]) + "…"
}
