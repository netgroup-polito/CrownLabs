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

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// this function returns a map in which keys are kinds of ClusterRoleBindings
// and values are the names of the respective ClusterRoles.
var crbData = map[string]string{
	"instances": "crownlabs-manage-instances",
	"tenants":   "crownlabs-manage-tenants",
}

func (r *Reconciler) manageClusterRoleBindings(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	for kind, roleName := range crbData {
		if err := r.createOrUpdateSingleCrb(ctx, ws, kind, roleName); err != nil {
			return fmt.Errorf("error while creating/updating %s ClusterRoleBinding for workspace %s: %w",
				kind, ws.Name, err)
		}
	}

	return nil
}

func (r *Reconciler) createOrUpdateSingleCrb(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	kind string,
	roleName string,
) error {
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: getClusterRoleBindingName(ws, kind),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, crb, func() error {
		crb.Labels = r.updateWsResourceCommonLabels(crb.Labels)
		crb.RoleRef.Kind = "ClusterRole"
		crb.RoleRef.Name = roleName
		crb.RoleRef.APIGroup = "rbac.authorization.k8s.io"

		crb.Subjects = []rbacv1.Subject{
			{
				Kind:     "Group",
				Name:     fmt.Sprintf("kubernetes:%s", workspaceRoleName(ws, v1alpha2.Manager)),
				APIGroup: "rbac.authorization.k8s.io",
			},
		}

		return controllerutil.SetControllerReference(ws, crb, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error while creating/updating %s ClusterRoleBinding for workspace %s: %w",
			kind, ws.Name, err)
	}

	return nil
}

// deleteClusterRoleBinding deletes the ClusterRoleBinding associated with the Workspace.
func (r *Reconciler) deleteClusterRoleBindings(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	for kind := range crbData {
		if err := r.deleteSingleCrb(ctx, ws, kind); err != nil {
			// Check if the error is something other than "not found"
			if !errors.IsNotFound(err) {
				return fmt.Errorf("error while deleting %s ClusterRoleBinding for workspace %s: %w",
					kind, ws.Name, err)
			}
		}
	}

	return nil
}

func (r *Reconciler) deleteSingleCrb(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	kind string,
) error {
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: getClusterRoleBindingName(ws, kind),
		},
	}

	if err := r.Client.Delete(ctx, crb); err != nil {
		return fmt.Errorf("error while deleting %s ClusterRoleBinding for workspace %s: %w",
			kind, ws.Name, err)
	}

	return nil
}

func getClusterRoleBindingName(ws *v1alpha1.Workspace, kind string) string {
	return fmt.Sprintf("crownlabs-manage-%s-%s", kind, ws.Name)
}
