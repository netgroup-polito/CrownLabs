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
	"reflect"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	Reschedule    common.Rescheduler
}

// Reconcile reconciles the state of a Workspace resource.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("workspace", req.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	log.Info("Reconciling workspace")

	var ws v1alpha1.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); client.IgnoreNotFound(err) != nil {
		log.Error(err, "Error when getting workspace before starting reconcile", "workspace", req.Name)
		return ctrl.Result{}, err
	} else if err != nil {
		log.Info("Workspace deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !r.TargetLabel.IsIncluded(ws.Labels) {
		// the actual Workspace is not responsibility of this controller
		log.Info("Workspace is not responsibility of this controller, skipping reconcile")
		return ctrl.Result{}, nil
	}

	avoidStatusUpdate := false
	defer func(original, updated *v1alpha1.Workspace) {
		if avoidStatusUpdate {
			return
		}

		// Avoid status update if not necessary.
		if !reflect.DeepEqual(original.Status, updated.Status) {
			if err := r.Status().Update(ctx, updated); err != nil {
				log.Error(err, "Error updating workspace status")
			} else {
				log.Info("Workspace status updated")
			}
		}
	}(ws.DeepCopy(), &ws)

	reschedule := r.Reschedule.GetReconcileResult()
	hasErrors := false

	// check if the Workspace is being deleted
	if !ws.DeletionTimestamp.IsZero() {
		err := r.enforceWorkspaceAbsence(ctx, log, &ws)
		if err != nil {
			log.Error(err, "Error deleting workspace", "workspace", ws.Name)
			return reschedule, err
		}
		log.Info("Workspace deleted")
		avoidStatusUpdate = true
		return ctrl.Result{}, nil
	}

	// add the finalizer if not already present
	if !controllerutil.ContainsFinalizer(&ws, v1alpha2.TnOperatorFinalizerName) {
		if err := r.enforcePreservingStatus(ctx, log, &ws, func(workspace *v1alpha1.Workspace) *v1alpha1.Workspace {
			controllerutil.AddFinalizer(workspace, v1alpha2.TnOperatorFinalizerName)
			return workspace
		}); err != nil {
			log.Error(err, "Error adding finalizer to workspace")
			return reschedule, err
		}
		log.Info("Added finalizer to workspace")
	}

	// enforce subresources for the Workspace
	err := r.enforceSubresources(ctx, log, &ws)
	if err != nil {
		log.Error(err, "Error enforcing subresources for workspace")
		return reschedule, fmt.Errorf("error enforcing subresources for workspace %s: %w", ws.Name, err)
	}

	// enforce AutoEnrollment for the Workspace
	err = r.enforceAutoEnrollment(ctx, &ws, log)
	if err != nil {
		log.Error(err, "Error enforcing AutoEnrollment for workspace")
		return reschedule, fmt.Errorf("error enforcing AutoEnrollment for workspace %s: %w", ws.Name, err)
	}
	log.Info("AutoEnrollment enforced for workspace")

	if ws.Status.Subscriptions == nil {
		ws.Status.Subscriptions = make(map[string]v1alpha2.SubscriptionStatus)
	}

	// setup roles in Keycloak
	if err := r.createKeycloakRoles(ctx, &ws, log); err != nil {
		log.Error(err, "Error managing Keycloak roles for workspace")
		hasErrors = true
	} else {
		log.Info("Keycloak roles updated/created for workspace")
	}

	ws.Status.Ready = !hasErrors

	return reschedule, nil
}

// SetupWithManager registers a new controller for Workspace resources.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, log logr.Logger) error {
	pred, err := r.TargetLabel.GetPredicate()
	if err != nil {
		log.Error(err, "Error creating predicate for tenant controller")
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

func (r *Reconciler) enforceWorkspaceAbsence(
	ctx context.Context,
	log logr.Logger,
	ws *v1alpha1.Workspace,
) error {
	// remove the Workspace from Tenants
	if err := r.handleTenantWorkspaceDeletion(ctx, ws, log); err != nil {
		return fmt.Errorf("error handling tenant workspace deletion for workspace %s: %w", ws.Name, err)
	}
	log.Info("Deleted workspace from all subscribed tenants")

	// delete roles in Keycloak
	if err := r.deleteKeycloakRoles(ctx, ws, log); err != nil {
		return fmt.Errorf("error deleting Keycloak roles for workspace %s: %w", ws.Name, err)
	}
	log.Info("Keycloak roles deleted for workspace")

	// delete subresources
	if err := r.enforceSubresourcesAbsence(ctx, log, ws); err != nil {
		return fmt.Errorf("error deleting subresources for workspace %s: %w", ws.Name, err)
	}
	log.Info("Subresources deleted for workspace")

	// delete finalizer
	if controllerutil.ContainsFinalizer(ws, v1alpha2.TnOperatorFinalizerName) {
		if err := r.enforcePreservingStatus(ctx, log, ws, func(workspace *v1alpha1.Workspace) *v1alpha1.Workspace {
			controllerutil.RemoveFinalizer(workspace, v1alpha2.TnOperatorFinalizerName)
			return workspace
		}); err != nil {
			return fmt.Errorf("error removing finalizer from workspace %s: %w", ws.Name, err)
		}
		log.Info("Removed finalizer from workspace")
	}

	return nil
}

func (r *Reconciler) enforceSubresources(
	ctx context.Context,
	log logr.Logger,
	ws *v1alpha1.Workspace,
) error {
	// Enforce the Namespace for the Workspace
	if err := r.enforceNamespace(ctx, ws); err != nil {
		return fmt.Errorf("error enforcing namespace for workspace %s: %w", ws.Name, err)
	}
	log.Info("Namespace created/updated for workspace")

	// Enforce the ClusterRoleBinding for the Workspace
	if err := r.enforceClusterRoleBindings(ctx, ws); err != nil {
		return fmt.Errorf("error enforcing ClusterRoleBinding for workspace %s: %w", ws.Name, err)
	}
	log.Info("ClusterRoleBindings created/updated for workspace")

	// Enforce the RoleBindings for the Workspace
	if err := r.enforceRoleBindings(ctx, ws); err != nil {
		return fmt.Errorf("error enforcing RoleBindings for workspace %s: %w", ws.Name, err)
	}
	log.Info("RoleBindings created/updated for workspace")

	return nil
}

func (r *Reconciler) enforceSubresourcesAbsence(
	ctx context.Context,
	log logr.Logger,
	ws *v1alpha1.Workspace,
) error {
	// Delete the RoleBindings for the Workspace
	if err := r.enforceRoleBindingsAbsence(ctx, ws); err != nil {
		log.Error(err, "Error deleting RoleBindings for workspace")
		return fmt.Errorf("error deleting RoleBindings for workspace %s: %w", ws.Name, err)
	}
	log.Info("RoleBindings deleted for workspace")

	// Delete the ClusterRoleBinding for the Workspace
	if err := r.enforceClusterRoleBindingsAbsence(ctx, ws); err != nil {
		log.Error(err, "Error deleting ClusterRoleBinding for workspace")
		return fmt.Errorf("error deleting ClusterRoleBinding for workspace %s: %w", ws.Name, err)
	}
	log.Info("ClusterRoleBindings deleted for workspace")

	// Delete the Namespace for the Workspace
	if err := r.enforceNamespaceAbsence(ctx, ws); err != nil {
		log.Error(err, "Error deleting namespace for workspace")
		return fmt.Errorf("error deleting namespace for workspace %s: %w", ws.Name, err)
	}
	log.Info("Namespace deleted for workspace")

	return nil
}
