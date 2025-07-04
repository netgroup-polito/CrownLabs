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

// Package workspace implements the workspace controller functionality.
package workspace

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// Reconciler reconciles Workspace objects.
type Reconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	TargetLabel   common.KVLabel
	KeycloakActor common.KeycloakActorIface
}

// Reconcile reconciles the state of a Workspace resource.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "workspace", req.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	log.Info("Reconciling workspace", "name", req.Name)

	var ws v1alpha1.Workspace
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

	avoidStatusUpdate := false
	defer func() {
		if avoidStatusUpdate {
			return
		}
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
		avoidStatusUpdate = true
		return ctrl.Result{}, nil
	}

	// add the finalizer if not already present
	if !controllerutil.ContainsFinalizer(&ws, v1alpha2.TnOperatorFinalizerName) {
		controllerutil.AddFinalizer(&ws, v1alpha2.TnOperatorFinalizerName)
		if err := r.Update(ctx, &ws); err != nil {
			klog.Errorf("Error adding finalizer to workspace %s: %v", ws.Name, err)
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer to workspace", "name", ws.Name)
	}

	// manage subresources for the Workspace
	err := r.manageSubresources(ctx, log, &ws)
	if err != nil {
		log.Error(err, "Error managing subresources for workspace", "name", ws.Name)
		return ctrl.Result{}, fmt.Errorf("error managing subresources for workspace %s: %w", ws.Name, err)
	}

	// manage AutoEnrollment for the Workspace
	err = r.manageAutoEnrollment(ctx, &ws)
	if err != nil {
		klog.Errorf("Error managing AutoEnrollment for workspace %s: %v", ws.Name, err.Error())
		return ctrl.Result{}, fmt.Errorf("error managing AutoEnrollment for workspace %s: %w", ws.Name, err)
	}
	log.Info("AutoEnrollment managed for workspace", "name", ws.Name)

	if ws.Status.Subscriptions == nil {
		ws.Status.Subscriptions = make(map[string]v1alpha2.SubscriptionStatus)
	}

	// setup roles in Keycloak
	if err := r.createKeycloakRoles(ctx, &ws); err != nil {
		log.Error(err, "Error managing Keycloak roles for workspace", "name", ws.Name)
		return ctrl.Result{}, fmt.Errorf("error managing Keycloak roles for workspace %s: %w", ws.Name, err)
	}
	log.Info("Keycloak roles updated/created for workspace", "name", ws.Name)

	ws.Status.Ready = true

	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Workspace resources.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred, err := r.TargetLabel.GetPredicate()
	if err != nil {
		klog.Errorf("Error creating predicate for tenant controller: %v", err)
		return fmt.Errorf("error creating predicate for tenant controller: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Workspace{}, builder.WithPredicates(pred)).
		Owns(&v1.Namespace{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.RoleBinding{}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "Workspace")).
		Complete(r)
}

func (r *Reconciler) deleteWorkspace(
	ctx context.Context,
	log logr.Logger,
	ws *v1alpha1.Workspace,
) error {
	// remove the Workspace from Tenants
	if err := r.handleTenantWorkspaceDeletion(ctx, ws); err != nil {
		klog.Errorf("Error handling tenant workspace deletion for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error handling tenant workspace deletion for workspace %s: %w", ws.Name, err)
	}
	log.Info("Deleted workspace from all subscribed tenants", "name", ws.Name)

	// delete roles in Keycloak
	if err := r.deleteKeycloakRoles(ctx, ws); err != nil {
		klog.Errorf("Error deleting Keycloak roles for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error deleting Keycloak roles for workspace %s: %w", ws.Name, err)
	}
	log.Info("Keycloak roles deleted for workspace", "name", ws.Name)

	// delete subresources
	if err := r.deleteSubresources(ctx, log, ws); err != nil {
		klog.Errorf("Error deleting subresources for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error deleting subresources for workspace %s: %w", ws.Name, err)
	}
	log.Info("Subresources deleted for workspace", "name", ws.Name)

	// delete finalizer
	if controllerutil.ContainsFinalizer(ws, v1alpha2.TnOperatorFinalizerName) {
		controllerutil.RemoveFinalizer(ws, v1alpha2.TnOperatorFinalizerName)
		if err := r.Update(ctx, ws); err != nil {
			klog.Errorf("Error removing finalizer from workspace %s: %v", ws.Name, err)
			return fmt.Errorf("error removing finalizer from workspace %s: %w", ws.Name, err)
		}
		log.Info("Removed finalizer from workspace", "name", ws.Name)
	}

	return nil
}

func (r *Reconciler) manageSubresources(
	ctx context.Context,
	log logr.Logger,
	ws *v1alpha1.Workspace,
) error {
	// Manage the Namespace for the Workspace
	if err := r.manageNamespace(ctx, ws); err != nil {
		klog.Errorf("Error managing namespace for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error managing namespace for workspace %s: %w", ws.Name, err)
	}
	log.Info("Namespace created/updated for workspace", "name", ws.Name)

	// Manage the ClusterRoleBinding for the Workspace
	if err := r.manageClusterRoleBindings(ctx, ws); err != nil {
		klog.Errorf("Error managing ClusterRoleBinding for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error managing ClusterRoleBinding for workspace %s: %w", ws.Name, err)
	}
	log.Info("ClusterRoleBindings created/updated for workspace", "name", ws.Name)

	// Manage the RoleBindings for the Workspace
	if err := r.manageRoleBindings(ctx, ws); err != nil {
		klog.Errorf("Error managing RoleBindings for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error managing RoleBindings for workspace %s: %w", ws.Name, err)
	}
	log.Info("RoleBindings created/updated for workspace", "name", ws.Name)

	return nil
}

func (r *Reconciler) deleteSubresources(
	ctx context.Context,
	log logr.Logger,
	ws *v1alpha1.Workspace,
) error {
	// Delete the RoleBindings for the Workspace
	if err := r.deleteRoleBindings(ctx, ws); err != nil {
		klog.Errorf("Error deleting RoleBindings for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error deleting RoleBindings for workspace %s: %w", ws.Name, err)
	}
	log.Info("RoleBindings deleted for workspace", "name", ws.Name)

	// Delete the ClusterRoleBinding for the Workspace
	if err := r.deleteClusterRoleBindings(ctx, ws); err != nil {
		klog.Errorf("Error deleting ClusterRoleBinding for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error deleting ClusterRoleBinding for workspace %s: %w", ws.Name, err)
	}
	log.Info("ClusterRoleBindings deleted for workspace", "name", ws.Name)

	// Delete the Namespace for the Workspace
	if err := r.deleteNamespace(ctx, ws); err != nil {
		klog.Errorf("Error deleting namespace for workspace %s: %v", ws.Name, err)
		return fmt.Errorf("error deleting namespace for workspace %s: %w", ws.Name, err)
	}
	log.Info("Namespace deleted for workspace", "name", ws.Name)

	return nil
}
