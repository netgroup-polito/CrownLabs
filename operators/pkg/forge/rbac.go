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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// WorkspaceInstancesManagerRoleName -> the name of the ClusterRole for managing instances in workspaces.
	WorkspaceInstancesManagerRoleName = "crownlabs-manage-instances"

	// WorkspaceTenantsManagerRoleName -> the name of the ClusterRole for managing tenants in workspaces.
	WorkspaceTenantsManagerRoleName = "crownlabs-manage-tenants"

	// TenantInstancesManagerRoleName -> the name of the ClusterRole for managing instances in tenant namespaces.
	TenantInstancesManagerRoleName = "crownlabs-manage-instances"

	// TenantManageClusterRolePrefix -> the prefix for tenant management ClusterRoles resource name.
	TenantManageClusterRolePrefix = "crownlabs-manage-"
)

// GetWorkspaceInstancesManagerBindingName returns the name of the ClusterRoleBinding for managing instances in a workspace.
func GetWorkspaceInstancesManagerBindingName(ws *v1alpha1.Workspace) string {
	return fmt.Sprintf("%s-%s", WorkspaceInstancesManagerRoleName, ws.Name)
}

// GetWorkspaceTenantsManagerBindingName returns the name of the ClusterRoleBinding for managing tenants in a workspace.
func GetWorkspaceTenantsManagerBindingName(ws *v1alpha1.Workspace) string {
	return fmt.Sprintf("%s-%s", WorkspaceTenantsManagerRoleName, ws.Name)
}

// ConfigureWorkspaceInstancesManagerBinding configures the RoleRef and Subjects for a ClusterRoleBinding
// that grants permissions to manage instances in a workspace.
func ConfigureWorkspaceInstancesManagerBinding(ws *v1alpha1.Workspace, crb *rbacv1.ClusterRoleBinding) {
	// Configure the RoleRef for instances management
	crb.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     WorkspaceInstancesManagerRoleName,
		APIGroup: rbacv1.GroupName,
	}

	// Set the subjects (Workspace Managers)
	crb.Subjects = []rbacv1.Subject{
		{
			Kind:     rbacv1.GroupKind,
			Name:     fmt.Sprintf("kubernetes:%s", WorkspaceRoleName(ws.Name, clv1alpha2.Manager)),
			APIGroup: rbacv1.GroupName,
		},
	}
}

// ConfigureWorkspaceTenantsManagerBinding configures the RoleRef and Subjects for a ClusterRoleBinding
// that grants permissions to manage tenants in a workspace.
func ConfigureWorkspaceTenantsManagerBinding(ws *v1alpha1.Workspace, crb *rbacv1.ClusterRoleBinding) {
	// Configure the RoleRef for tenants management
	crb.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     WorkspaceTenantsManagerRoleName,
		APIGroup: rbacv1.GroupName,
	}

	// Set the subjects (Workspace Managers)
	crb.Subjects = []rbacv1.Subject{
		{
			Kind:     rbacv1.GroupKind,
			Name:     fmt.Sprintf("kubernetes:%s", WorkspaceRoleName(ws.Name, clv1alpha2.Manager)),
			APIGroup: rbacv1.GroupName,
		},
	}
}

// ResourceObjectMeta returns a generic ObjectMeta for a resource.
func ResourceObjectMeta(name string, labels map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   name,
		Labels: labels,
	}
}

// ConfigureTenantInstancesRoleBinding configures a RoleBinding for a tenant to manage instances.
func ConfigureTenantInstancesRoleBinding(rb *rbacv1.RoleBinding, tn *clv1alpha2.Tenant, labels map[string]string) {
	// Set the labels
	if rb.Labels == nil {
		rb.Labels = make(map[string]string)
	}

	// Copy the provided labels
	maps.Copy(rb.Labels, labels)

	// Configure the role binding spec
	rb.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     TenantInstancesManagerRoleName,
		APIGroup: rbacv1.GroupName,
	}
	rb.Subjects = []rbacv1.Subject{{
		Kind:     rbacv1.UserKind,
		Name:     tn.Name,
		APIGroup: rbacv1.GroupName,
	}}
}

// GetTenantClusterRoleResourceName returns the name for tenant cluster resources.
func GetTenantClusterRoleResourceName(tn *clv1alpha2.Tenant) string {
	namespace := GetTenantNamespaceName(tn)
	return fmt.Sprintf("%s%s", TenantManageClusterRolePrefix, namespace)
}

// ConfigureTenantClusterRole configures a ClusterRole for tenant access.
func ConfigureTenantClusterRole(cr *rbacv1.ClusterRole, tn *clv1alpha2.Tenant, labels map[string]string) {
	// Set the labels
	if cr.Labels == nil {
		cr.Labels = make(map[string]string)
	}

	// Copy the provided labels
	maps.Copy(cr.Labels, labels)

	// Configure the cluster role rules
	cr.Rules = []rbacv1.PolicyRule{{
		APIGroups:     []string{"crownlabs.polito.it"},
		Resources:     []string{"tenants"},
		ResourceNames: []string{tn.Name},
		Verbs:         []string{"get", "list", "watch", "patch", "update"},
	}}
}

// ConfigureTenantClusterRoleBinding configures a ClusterRoleBinding for tenant access.
func ConfigureTenantClusterRoleBinding(crb *rbacv1.ClusterRoleBinding, tn *clv1alpha2.Tenant, labels map[string]string) {
	// Set the labels
	if crb.Labels == nil {
		crb.Labels = make(map[string]string)
	}

	// Copy the provided labels
	for k, v := range labels {
		crb.Labels[k] = v
	}

	// Get the name for the resources
	resourceName := GetTenantClusterRoleResourceName(tn)

	// Configure the cluster role binding
	crb.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     resourceName,
		APIGroup: rbacv1.GroupName,
	}
	crb.Subjects = []rbacv1.Subject{{
		Kind:     rbacv1.UserKind,
		Name:     tn.Name,
		APIGroup: rbacv1.GroupName,
	}}
}
