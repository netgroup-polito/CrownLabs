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

package forge

import (
	"fmt"
	"maps"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// ViewTemplatesRoleName -> the name of the ClusterRole for viewing templates in workspaces.
	ViewTemplatesRoleName = "crownlabs-view-templates"

	// ManageTemplatesRoleName -> the name of the ClusterRole for managing templates in workspaces.
	ManageTemplatesRoleName = "crownlabs-manage-templates"

	// ManageSharedVolumesRoleName -> the name of the ClusterRole for managing shared volumes in workspaces.
	ManageSharedVolumesRoleName = "crownlabs-manage-sharedvolumes"
)

// ConfigureWorkspaceUserViewTemplatesBinding configures a RoleBinding for a workspace user to view templates.
func ConfigureWorkspaceUserViewTemplatesBinding(ws *v1alpha1.Workspace, rb *rbacv1.RoleBinding, labels map[string]string) {
	// Set labels
	if rb.Labels == nil {
		rb.Labels = make(map[string]string)
	}
	maps.Copy(rb.Labels, labels)

	// Configure RoleRef
	rb.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     ViewTemplatesRoleName,
		APIGroup: rbacv1.GroupName,
	}

	// Configure Subjects
	rb.Subjects = []rbacv1.Subject{
		{
			Kind:     rbacv1.GroupKind,
			Name:     fmt.Sprintf("kubernetes:%s", WorkspaceRoleName(ws.Name, v1alpha2.User)),
			APIGroup: rbacv1.GroupName,
		},
	}
}

// ConfigureWorkspaceManagerManageTemplatesBinding configures a RoleBinding for a workspace manager to manage templates.
func ConfigureWorkspaceManagerManageTemplatesBinding(ws *v1alpha1.Workspace, rb *rbacv1.RoleBinding, labels map[string]string) {
	// Set labels
	if rb.Labels == nil {
		rb.Labels = make(map[string]string)
	}
	for k, v := range labels {
		rb.Labels[k] = v
	}

	// Configure RoleRef
	rb.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     ManageTemplatesRoleName,
		APIGroup: rbacv1.GroupName,
	}

	// Configure Subjects
	rb.Subjects = []rbacv1.Subject{
		{
			Kind:     rbacv1.GroupKind,
			Name:     fmt.Sprintf("kubernetes:%s", WorkspaceRoleName(ws.Name, v1alpha2.Manager)),
			APIGroup: rbacv1.GroupName,
		},
	}
}

// ConfigureWorkspaceManagerManageSharedVolumesBinding configures a RoleBinding for a workspace manager to manage shared volumes.
func ConfigureWorkspaceManagerManageSharedVolumesBinding(ws *v1alpha1.Workspace, rb *rbacv1.RoleBinding, labels map[string]string) {
	// Set labels
	if rb.Labels == nil {
		rb.Labels = make(map[string]string)
	}
	maps.Copy(rb.Labels, labels)

	// Configure RoleRef
	rb.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     ManageSharedVolumesRoleName,
		APIGroup: rbacv1.GroupName,
	}

	// Configure Subjects
	rb.Subjects = []rbacv1.Subject{
		{
			Kind:     rbacv1.GroupKind,
			Name:     fmt.Sprintf("kubernetes:%s", WorkspaceRoleName(ws.Name, v1alpha2.Manager)),
			APIGroup: rbacv1.GroupName,
		},
	}
}
