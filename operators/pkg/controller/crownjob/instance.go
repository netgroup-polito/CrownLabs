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

package crownjob

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	ctrlcommon "github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// CrownJobInstanceIndex -> the field index of the CrownJobs on the name of their Instance.
const CrownJobInstanceIndex = "spec.instanceRef.name"

// InstanceReconciler keeps an Instance while it has CrownJobs that are not terminal, by removing the finalizer that the
// CrownJob reconciler adds right before creating a Job, once every CrownJob of the Instance is terminal. When the
// Instance is being deleted, its CrownJobs not admitted yet are canceled, while those admitted run until they end.
// It never adds the finalizer, so it cannot race with the CrownJob reconciler, which adds it while its CrownJob is not
// terminal; both write with optimistic locking.
type InstanceReconciler struct {
	client.Client
	// APIReader reads, without cache, the Jobs of the CrownJobs to cancel, to check that they are not admitted.
	APIReader      client.Reader
	TargetLabel    ctrlcommon.KVLabel
	EventsRecorder record.EventRecorder

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// IndexCrownJobsByInstance registers the field index of the CrownJobs on the name of their Instance.
func IndexCrownJobsByInstance(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &clv1alpha2.CrownJob{}, CrownJobInstanceIndex, func(obj client.Object) []string {
		crownJob, ok := obj.(*clv1alpha2.CrownJob)
		if !ok || crownJob.Spec.Instance.Name == "" {
			return nil
		}
		return []string{crownJob.Spec.Instance.Name}
	})
}

// SetupWithManager registers a new controller for the Instances that have CrownJobs. It watches the Instances
// metadata only, and the CrownJobs, each mapped to its Instance (also when deleted).
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	if err := IndexCrownJobsByInstance(context.Background(), mgr.GetFieldIndexer()); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("crownjob-instance").
		For(&clv1alpha2.Instance{}, builder.OnlyMetadata).
		Watches(&clv1alpha2.CrownJob{}, handler.EnqueueRequestsFromMapFunc(crownJobToInstance)).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "CrownJobInstance")).
		Complete(r)
}

// Reconcile cancels the CrownJobs not admitted yet of an Instance being deleted, and removes the finalizer of the
// Instance once all its CrownJobs are terminal.
func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	log := ctrl.LoggerFrom(ctx, "instance", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	instance := &metav1.PartialObjectMetadata{}
	instance.SetGroupVersionKind(clv1alpha2.GroupVersion.WithKind("Instance"))
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving instance")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if proceed, err := utils.CheckNamespaceTargetLabel(ctx, r.Client, instance.GetNamespace(), r.TargetLabel); !proceed {
		if err != nil {
			log.Error(err, "failed checking target label")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	var crownJobs clv1alpha2.CrownJobList
	if err := r.List(ctx, &crownJobs, client.InNamespace(instance.GetNamespace()), client.MatchingFields{CrownJobInstanceIndex: instance.GetName()}); err != nil {
		log.Error(err, "failed listing the crownjobs of the instance")
		return ctrl.Result{}, err
	}

	allTerminal := true
	for i := range crownJobs.Items {
		crownJob := &crownJobs.Items[i]
		if ns := crownJob.Spec.Instance.Namespace; ns != "" && ns != crownJob.Namespace {
			// Refused by the CrownJob reconciler (rule 1): it does not use this Instance.
			continue
		}
		if !instance.GetDeletionTimestamp().IsZero() && crownJob.Status.Phase == clv1alpha2.CrownJobPhaseWaiting {
			canceled, err := r.cancel(ctx, crownJob, instance.GetName())
			if kerrors.IsConflict(err) {
				log.Info("the crownjob changed while it was canceled, retrying", "crownjob", klog.KObj(crownJob))
				return ctrl.Result{RequeueAfter: requeueConflict}, nil
			} else if err != nil {
				return ctrl.Result{}, err
			}
			if canceled {
				continue
			}
		}
		if !isTerminal(crownJob.Status.Phase) {
			allTerminal = false
		}
	}

	if !allTerminal || !ctrlutil.ContainsFinalizer(instance, forge.CrownJobInstanceFinalizer) {
		return ctrl.Result{}, nil
	}

	original := instance.DeepCopy()
	ctrlutil.RemoveFinalizer(instance, forge.CrownJobInstanceFinalizer)
	if err := r.Patch(ctx, instance, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{})); err != nil {
		if kerrors.IsConflict(err) {
			// E.g., the CrownJob reconciler has just added the finalizer for a new job: check again.
			log.Info("the instance changed while its finalizer was removed, retrying")
			return ctrl.Result{RequeueAfter: requeueConflict}, nil
		}
		log.Error(err, "failed to remove the finalizer of the instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("finalizer removed: every crownjob of the instance is terminal", "crownjobs", len(crownJobs.Items))
	return ctrl.Result{}, nil
}

// cancel sets a Waiting CrownJob to Failed/Canceled and deletes its Job, unless Kueue has already admitted the Job
// (un-suspended it): then the job runs, and the CrownJob reconciler will report it as Starting. It returns whether
// the CrownJob was canceled.
func (r *InstanceReconciler) cancel(ctx context.Context, crownJob *clv1alpha2.CrownJob, instanceName string) (bool, error) {
	var job batchv1.Job
	jobFound := true
	if err := r.APIReader.Get(ctx, client.ObjectKeyFromObject(crownJob), &job); kerrors.IsNotFound(err) {
		jobFound = false
	} else if err != nil {
		return false, err
	}
	if jobFound && metav1.IsControlledBy(&job, crownJob) && !ptr.Deref(job.Spec.Suspend, false) {
		return false, nil
	}

	original := crownJob.DeepCopy()
	message := fmt.Sprintf(msgCanceled, instanceName)
	setStatus(crownJob, clv1alpha2.CrownJobPhaseFailed, ReasonCanceled, message)
	if err := r.Status().Patch(ctx, crownJob, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{})); err != nil {
		return false, err
	}
	r.EventsRecorder.Event(crownJob, corev1.EventTypeWarning, ReasonCanceled, message)
	ctrl.LoggerFrom(ctx).Info("crownjob canceled: its instance is being deleted", "crownjob", klog.KObj(crownJob))

	if jobFound && metav1.IsControlledBy(&job, crownJob) {
		if err := r.Delete(ctx, &job, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil && !kerrors.IsNotFound(err) {
			return true, err
		}
	}
	return true, nil
}

// crownJobToInstance maps a CrownJob to its Instance, in the namespace of the CrownJob.
func crownJobToInstance(_ context.Context, obj client.Object) []reconcile.Request {
	crownJob, ok := obj.(*clv1alpha2.CrownJob)
	if !ok || crownJob.Spec.Instance.Name == "" {
		return nil
	}
	return []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: crownJob.Namespace, Name: crownJob.Spec.Instance.Name}}}
}
