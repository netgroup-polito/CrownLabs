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

// Package tenant_controller groups the functionalities related to the Tenant controller.
package workspace

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

type WorkspaceReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	TargetLabel common.KVLabel
}

// Reconcile reconciles the state of a Workspace resource.
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "workspace", req.NamespacedName.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	log.Info("Reconciling workspace", "name", req.NamespacedName.Name)

	var ws crownlabsv1alpha1.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting workspace %s before starting reconcile -> %s", req.Name, err)
		return ctrl.Result{}, err
	} else if err != nil {
		log.Info("Workspace deleted", "name", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !r.TargetLabel.IsIncluded(ws.Labels) {
		// the actual Workspace is not responsibility of this controller
		log.Info("Workspace is not responsibility of this controller, skipping reconcile")
		return ctrl.Result{}, nil
	}

	defer func() {
		// update the Tenant status
		if err := r.Status().Update(ctx, &ws); err != nil {
			klog.Errorf("Error updating status for workspace %s: %v", ws.Name, err)
		}
	}()

	// check if the Workspace is being deleted
	if !ws.DeletionTimestamp.IsZero() {
		err := r.deleteWorkspace(ctx, log, &ws)
		if err != nil {
			klog.Errorf("Error deleting workspace %s: %v", ws.Name, err)
			return ctrl.Result{}, err
		}
		log.Info("Workspace deleted", "name", ws.Name)
		return ctrl.Result{}, nil
	}

	// add the finalizer if not already present
	if !controllerutil.ContainsFinalizer(&ws, crownlabsv1alpha2.TnOperatorFinalizerName) {
		controllerutil.AddFinalizer(&ws, crownlabsv1alpha2.TnOperatorFinalizerName)
		if err := r.Update(ctx, &ws); err != nil {
			klog.Errorf("Error adding finalizer to workspace %s: %v", ws.Name, err)
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer to workspace", "name", ws.Name)
	}

	err := r.manageSubresources(ctx, log, &ws)
	if err != nil {
		// TODO: handle error properly
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Workspace resources.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred, err := r.TargetLabel.GetPredicate()
	if err != nil {
		klog.Errorf("Error creating predicate for tenant controller: %v", err)
		return fmt.Errorf("error creating predicate for tenant controller: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha1.Workspace{}, builder.WithPredicates(pred)).
		Owns(&v1.Namespace{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.RoleBinding{}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "Workspace")).
		Complete(r)
}

func (r *WorkspaceReconciler) deleteWorkspace(
	ctx context.Context,
	log logr.Logger,
	ws *crownlabsv1alpha1.Workspace,
) error {
	// delete subresources
	if err := r.deleteSubresources(ctx, log, ws); err != nil {
		klog.Errorf("Error deleting subresources for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error deleting subresources for workspace %s: %w", ws.Name, err)
	}
	log.Info("Subresources deleted for workspace", "name", ws.Name)

	// delete finalizer
	if controllerutil.ContainsFinalizer(ws, crownlabsv1alpha2.TnOperatorFinalizerName) {
		controllerutil.RemoveFinalizer(ws, crownlabsv1alpha2.TnOperatorFinalizerName)
		if err := r.Update(ctx, ws); err != nil {
			klog.Errorf("Error removing finalizer from workspace %s: %v", ws.Name, err)
			return fmt.Errorf("error removing finalizer from workspace %s: %w", ws.Name, err)
		}
		log.Info("Removed finalizer from workspace", "name", ws.Name)
	}

	return nil
}

func (r *WorkspaceReconciler) manageSubresources(
	ctx context.Context,
	log logr.Logger,
	ws *crownlabsv1alpha1.Workspace,
) error {
	// Manage the Namespace for the Workspace
	if err := r.manageNamespace(ctx, ws); err != nil {
		klog.Errorf("Error managing namespace for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error managing namespace for workspace %s: %w", ws.Name, err)
	}
	log.Info("Namespace created/updated for workspace", "name", ws.Name)

	// Manage the ClusterRoleBinding for the Workspace
	if err := r.manageClusterRoleBinding(ctx, ws); err != nil {
		klog.Errorf("Error managing ClusterRoleBinding for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error managing ClusterRoleBinding for workspace %s: %w", ws.Name, err)
	}
	log.Info("ClusterRoleBinding created/updated for workspace", "name", ws.Name)

	return nil
}

func (r *WorkspaceReconciler) deleteSubresources(
	ctx context.Context,
	log logr.Logger,
	ws *crownlabsv1alpha1.Workspace,
) error {
	// Delete the ClusterRoleBinding for the Workspace
	if err := r.deleteClusterRoleBinding(ctx, ws); err != nil {
		klog.Errorf("Error deleting ClusterRoleBinding for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error deleting ClusterRoleBinding for workspace %s: %w", ws.Name, err)
	}
	log.Info("ClusterRoleBinding deleted for workspace", "name", ws.Name)

	// Delete the Namespace for the Workspace
	if err := r.deleteNamespace(ctx, ws); err != nil {
		klog.Errorf("Error deleting namespace for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error deleting namespace for workspace %s: %w", ws.Name, err)
	}
	log.Info("Namespace deleted for workspace", "name", ws.Name)

	return nil
}
