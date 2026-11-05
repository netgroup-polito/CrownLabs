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

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	ctrlcommon "github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// QueueReconciler reconciles the job queue of a Workspace: the ResourceFlavor shared by every job queue and the
// ClusterQueue of the Workspace. The LocalQueues are applied by the CrownJob controller, right before each Job.
type QueueReconciler struct {
	client.Client
	TargetLabel ctrlcommon.KVLabel

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// SetupWithManager registers a new controller for the job queues of the Workspaces. It is named explicitly,
// since the Workspace controller already takes the default name derived from the Workspace kind. The ClusterQueues
// are watched to restore them if changed or deleted, but not on the status updates written by Kueue.
func (r *QueueReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	pred, err := r.TargetLabel.GetPredicate()
	if err != nil {
		return fmt.Errorf("error creating predicate for the crownjob queue controller: %w", err)
	}

	clusterQueue := &unstructured.Unstructured{}
	clusterQueue.SetGroupVersionKind(forge.KueueClusterQueueGVK)

	return ctrl.NewControllerManagedBy(mgr).
		Named("crownjob-queue").
		For(&clv1alpha1.Workspace{}, builder.WithPredicates(pred)).
		Owns(clusterQueue, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "CrownJobQueue")).
		Complete(r)
}

// Reconcile reconciles the job queue of a Workspace.
func (r *QueueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	log := ctrl.LoggerFrom(ctx, "workspace", req.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	var workspace clv1alpha1.Workspace
	if err := r.Get(ctx, req.NamespacedName, &workspace); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving workspace")
		}
		// The ClusterQueue of a deleted Workspace is removed by the garbage collector.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !r.TargetLabel.IsIncluded(workspace.Labels) {
		log.V(utils.LogDebugLevel).Info("workspace is not responsibility of this controller, skipping reconcile")
		return ctrl.Result{}, nil
	}
	if !workspace.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}

	if workspace.Spec.QueueLimits == nil {
		return ctrl.Result{}, r.enforceClusterQueueAbsence(ctx, &workspace)
	}

	if err := r.enforceResourceFlavor(ctx); err != nil {
		return ctrl.Result{}, err
	}

	clusterQueue := forge.KueueClusterQueue(&workspace)
	if err := r.Patch(ctx, clusterQueue, client.Apply, client.FieldOwner(forge.CrownJobFieldManager), client.ForceOwnership); err != nil {
		log.Error(err, "failed to apply the cluster queue", "clusterqueue", klog.KObj(clusterQueue))
		return ctrl.Result{}, err
	}
	log.Info("cluster queue enforced", "clusterqueue", klog.KObj(clusterQueue))
	return ctrl.Result{}, nil
}

// enforceResourceFlavor creates the ResourceFlavor used by every job queue, if it is missing. It is never deleted.
func (r *QueueReconciler) enforceResourceFlavor(ctx context.Context) error {
	flavor := forge.KueueResourceFlavor()
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(forge.KueueResourceFlavorGVK)
	err := r.Get(ctx, types.NamespacedName{Name: flavor.GetName()}, existing)
	if err == nil || !kerrors.IsNotFound(err) {
		return err
	}
	if err := r.Create(ctx, flavor, client.FieldOwner(forge.CrownJobFieldManager)); err != nil && !kerrors.IsAlreadyExists(err) {
		ctrl.LoggerFrom(ctx).Error(err, "failed to create the resource flavor", "resourceflavor", flavor.GetName())
		return err
	}
	ctrl.LoggerFrom(ctx).Info("resource flavor created", "resourceflavor", flavor.GetName())
	return nil
}

// enforceClusterQueueAbsence deletes the ClusterQueue of a Workspace without queueLimits, if it is controlled by it.
func (r *QueueReconciler) enforceClusterQueueAbsence(ctx context.Context, workspace *clv1alpha1.Workspace) error {
	log := ctrl.LoggerFrom(ctx)

	clusterQueue := &unstructured.Unstructured{}
	clusterQueue.SetGroupVersionKind(forge.KueueClusterQueueGVK)
	if err := r.Get(ctx, types.NamespacedName{Name: forge.CrownJobQueueName(workspace.Name)}, clusterQueue); err != nil {
		return client.IgnoreNotFound(err)
	}
	if !metav1.IsControlledBy(clusterQueue, workspace) {
		log.Info("cluster queue not controlled by the workspace, not deleting it", "clusterqueue", clusterQueue.GetName())
		return nil
	}
	if !clusterQueue.GetDeletionTimestamp().IsZero() {
		return nil
	}

	if err := r.Delete(ctx, clusterQueue); err != nil && !kerrors.IsNotFound(err) {
		log.Error(err, "failed to delete the cluster queue", "clusterqueue", clusterQueue.GetName())
		return err
	}
	log.Info("cluster queue deleted", "clusterqueue", clusterQueue.GetName())
	return nil
}
