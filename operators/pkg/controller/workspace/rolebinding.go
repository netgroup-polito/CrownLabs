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

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *WorkspaceReconciler) manageRoleBindings(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	if !ws.Status.Namespace.Created {
		return fmt.Errorf("cannot manage RoleBindings for Workspace %s: namespace not created", ws.Name)
	}
	namespace := ws.Status.Namespace.Name

	// all Users can view templates
	if err := r.createOrUpdateSingleRb(
		ctx,
		ws,
		namespace,
		"view-templates",
		"crownlabs-view-templates",
		v1alpha2.User,
	); err != nil {
		return err
	}

	// Managers can manage templates
	if err := r.createOrUpdateSingleRb(
		ctx,
		ws,
		namespace,
		"manage-templates",
		"crownlabs-manage-templates",
		v1alpha2.Manager,
	); err != nil {
		return err
	}

	// Managers can manage SharedVolumes
	if err := r.createOrUpdateSingleRb(
		ctx,
		ws,
		namespace,
		"manage-sharedvolumes",
		"crownlabs-manage-sharedvolumes",
		v1alpha2.Manager,
	); err != nil {
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) createOrUpdateSingleRb(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	namespace string,
	kind string,
	roleName string,
	authorized v1alpha2.WorkspaceUserRole,
) error {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("crownlabs-%s", kind),
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, rb, func() error {
		rb.Labels = r.updateWsResourceCommonLabels(rb.Labels)
		rb.RoleRef.Kind = "ClusterRole"
		rb.RoleRef.Name = roleName
		rb.RoleRef.APIGroup = "rbac.authorization.k8s.io"

		rb.Subjects = []rbacv1.Subject{
			{
				Kind:     "Group",
				Name:     fmt.Sprintf("kubernetes:%s", workspaceRoleName(ws, authorized)),
				APIGroup: "rbac.authorization.k8s.io",
			},
		}

		return controllerutil.SetControllerReference(ws, rb, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error while creating/updating RoleBinding %s: %w", rb.Name, err)
	}

	return nil
}

func (r *WorkspaceReconciler) deleteRoleBindings(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	if !ws.Status.Namespace.Created {
		return fmt.Errorf("cannot delete RoleBindings for Workspace %s: namespace not created", ws.Name)
	}
	namespace := ws.Status.Namespace.Name

	// Delete all RoleBindings related to the Workspace
	for _, kind := range []string{"view-templates", "manage-templates", "manage-sharedvolumes"} {
		if err := r.deleteSingleRb(ctx, ws, namespace, kind); err != nil {
			return fmt.Errorf("error while deleting RoleBinding %s: %w", kind, err)
		}
	}

	return nil
}

func (r *WorkspaceReconciler) deleteSingleRb(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	namespace string,
	kind string,
) error {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("crownlabs-%s", kind),
			Namespace: namespace,
		},
	}

	if err := r.Client.Delete(ctx, rb); err != nil {
		return fmt.Errorf("error while deleting RoleBinding %s: %w", rb.Name, err)
	}

	return nil
}
