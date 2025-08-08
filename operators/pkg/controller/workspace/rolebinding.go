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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func (r *Reconciler) enforceRoleBindings(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	if !ws.Status.Namespace.Created {
		return fmt.Errorf("cannot manage RoleBindings for Workspace %s: namespace not created", ws.Name)
	}
	namespace := ws.Status.Namespace.Name

	// Enforce User View Templates RoleBinding
	if err := r.enforceUserViewTemplatesRoleBinding(ctx, ws, namespace); err != nil {
		return fmt.Errorf("error while managing User View Templates RoleBinding for workspace %s: %w", ws.Name, err)
	}

	// Enforce Manager Manage Templates RoleBinding
	if err := r.enforceManagerManageTemplatesRoleBinding(ctx, ws, namespace); err != nil {
		return fmt.Errorf("error while managing Manager Manage Templates RoleBinding for workspace %s: %w", ws.Name, err)
	}

	// Enforce Manager Manage SharedVolumes RoleBinding
	if err := r.enforceManagerManageSharedVolumesRoleBinding(ctx, ws, namespace); err != nil {
		return fmt.Errorf("error while managing Manager Manage SharedVolumes RoleBinding for workspace %s: %w", ws.Name, err)
	}

	return nil
}

func (r *Reconciler) enforceRoleBindingsAbsence(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	if !ws.Status.Namespace.Created {
		return nil // No RoleBindings to delete if the namespace is not created
	}
	namespace := ws.Status.Namespace.Name

	// Delete User View Templates RoleBinding
	if err := r.deleteSingleRb(ctx, namespace, forge.ViewTemplatesRoleName); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting User View Templates RoleBinding: %w", err)
		}
	}

	// Delete Manager Manage Templates RoleBinding
	if err := r.deleteSingleRb(ctx, namespace, forge.ManageTemplatesRoleName); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting Manager Manage Templates RoleBinding: %w", err)
		}
	}

	// Delete Manager Manage SharedVolumes RoleBinding
	if err := r.deleteSingleRb(ctx, namespace, forge.ManageSharedVolumesRoleName); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting Manager Manage SharedVolumes RoleBinding: %w", err)
		}
	}

	return nil
}

func (r *Reconciler) deleteSingleRb(
	ctx context.Context,
	namespace string,
	name string,
) error {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := utils.EnforceObjectAbsence(ctx, r.Client, rb, "RoleBinding"); err != nil {
		return fmt.Errorf("error while deleting RoleBinding %s: %w", rb.Name, err)
	}

	return nil
}

// enforceUserViewTemplatesRoleBinding creates or updates the RoleBinding for User View Templates.
func (r *Reconciler) enforceUserViewTemplatesRoleBinding(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	namespace string,
) error {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.ViewTemplatesRoleName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, rb, func() error {
		// Update labels
		rb.Labels = forge.UpdateWorkspaceResourceCommonLabels(rb.Labels, r.TargetLabel)

		// Configure the RoleBinding
		forge.ConfigureWorkspaceUserViewTemplatesBinding(ws, rb, rb.Labels)

		return controllerutil.SetControllerReference(ws, rb, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error while creating/updating User View Templates RoleBinding: %w", err)
	}

	return nil
}

// enforceManagerManageTemplatesRoleBinding creates or updates the RoleBinding for Manager Manage Templates.
func (r *Reconciler) enforceManagerManageTemplatesRoleBinding(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	namespace string,
) error {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.ManageTemplatesRoleName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, rb, func() error {
		// Update labels
		rb.Labels = forge.UpdateWorkspaceResourceCommonLabels(rb.Labels, r.TargetLabel)

		// Configure the RoleBinding
		forge.ConfigureWorkspaceManagerManageTemplatesBinding(ws, rb, rb.Labels)

		return controllerutil.SetControllerReference(ws, rb, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error while creating/updating Manager Manage Templates RoleBinding: %w", err)
	}

	return nil
}

// enforceManagerManageSharedVolumesRoleBinding creates or updates the RoleBinding for Manager Manage SharedVolumes.
func (r *Reconciler) enforceManagerManageSharedVolumesRoleBinding(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	namespace string,
) error {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.ManageSharedVolumesRoleName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, rb, func() error {
		// Update labels
		rb.Labels = forge.UpdateWorkspaceResourceCommonLabels(rb.Labels, r.TargetLabel)

		// Configure the RoleBinding
		forge.ConfigureWorkspaceManagerManageSharedVolumesBinding(ws, rb, rb.Labels)

		return controllerutil.SetControllerReference(ws, rb, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error while creating/updating Manager Manage SharedVolumes RoleBinding: %w", err)
	}

	return nil
}
