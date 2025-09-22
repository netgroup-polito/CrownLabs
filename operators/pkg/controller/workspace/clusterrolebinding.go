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

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// enforceClusterRoleBindings ensures that all necessary ClusterRoleBindings exist for the workspace.
func (r *Reconciler) enforceClusterRoleBindings(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	// Create or update the ClusterRoleBinding for managing instances
	if err := r.enforceInstancesManagerBinding(ctx, ws); err != nil {
		return err
	}

	// Create or update the ClusterRoleBinding for managing tenants
	if err := r.enforceTenantsManagerBinding(ctx, ws); err != nil {
		return err
	}

	return nil
}

// enforceInstancesManagerBinding ensures that the ClusterRoleBinding for managing instances exists.
func (r *Reconciler) enforceInstancesManagerBinding(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	// Create only the skeleton of the ClusterRoleBinding with immutable information
	name := forge.GetWorkspaceInstancesManagerBindingName(ws)
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	// Update or create the resource, setting all mutable values in the callback
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, crb, func() error {
		// Update labels
		crb.Labels = forge.UpdateWorkspaceResourceCommonLabels(crb.Labels, r.TargetLabel)

		// Configure subjects and roleRef
		forge.ConfigureWorkspaceInstancesManagerBinding(ws, crb)

		return controllerutil.SetControllerReference(ws, crb, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error while creating/updating instances ClusterRoleBinding for workspace %s: %w",
			ws.Name, err)
	}

	return nil
}

// enforceTenantsManagerBinding ensures that the ClusterRoleBinding for managing tenants exists.
func (r *Reconciler) enforceTenantsManagerBinding(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	// Create only the skeleton of the ClusterRoleBinding with immutable information
	name := forge.GetWorkspaceTenantsManagerBindingName(ws)
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	// Update or create the resource, setting all mutable values in the callback
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, crb, func() error {
		// Update labels
		crb.Labels = forge.UpdateWorkspaceResourceCommonLabels(crb.Labels, r.TargetLabel)

		// Configure subjects and roleRef
		forge.ConfigureWorkspaceTenantsManagerBinding(ws, crb)

		return controllerutil.SetControllerReference(ws, crb, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error while creating/updating tenants ClusterRoleBinding for workspace %s: %w",
			ws.Name, err)
	}

	return nil
}

// deleteClusterRoleBindings deletes all ClusterRoleBindings associated with the Workspace.
func (r *Reconciler) enforceClusterRoleBindingsAbsence(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	// Delete the ClusterRoleBinding for managing instances
	if err := r.enforceInstancesManagerBindingAbsence(ctx, ws); err != nil {
		return err
	}

	// Delete the ClusterRoleBinding for managing tenants
	if err := r.enforceTenantsManagerBindingAbsence(ctx, ws); err != nil {
		return err
	}

	return nil
}

// deleteInstancesManagerBinding deletes the ClusterRoleBinding for managing instances.
func (r *Reconciler) enforceInstancesManagerBindingAbsence(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	// Create only the skeleton of the ClusterRoleBinding needed for deletion
	name := forge.GetWorkspaceInstancesManagerBindingName(ws)
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if err := utils.EnforceObjectAbsence(ctx, r.Client, crb, "ClusterRoleBinding"); err != nil {
		return fmt.Errorf("error while deleting instances ClusterRoleBinding for workspace %s: %w",
			ws.Name, err)
	}
	return nil
}

// deleteTenantsManagerBinding deletes the ClusterRoleBinding for managing tenants.
func (r *Reconciler) enforceTenantsManagerBindingAbsence(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	// Create only the skeleton of the ClusterRoleBinding needed for deletion
	name := forge.GetWorkspaceTenantsManagerBindingName(ws)
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if err := utils.EnforceObjectAbsence(ctx, r.Client, crb, "ClusterRoleBinding"); err != nil {
		return fmt.Errorf("error while deleting tenants ClusterRoleBinding for workspace %s: %w",
			ws.Name, err)
	}
	return nil
}
