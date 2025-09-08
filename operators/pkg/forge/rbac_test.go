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

package forge_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("RBAC forging", func() {
	var (
		workspace *v1alpha1.Workspace
	)

	BeforeEach(func() {
		workspace = &v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-workspace",
			},
			Spec: v1alpha1.WorkspaceSpec{
				PrettyName: "Test Workspace",
			},
		}
	})

	Describe("The forge.GetWorkspaceInstancesManagerBindingName function", func() {
		It("Should generate the correct binding name for instance managers", func() {
			expectedName := "crownlabs-manage-instances-test-workspace"
			Expect(forge.GetWorkspaceInstancesManagerBindingName(workspace)).To(Equal(expectedName))
		})
	})

	Describe("The forge.GetWorkspaceTenantsManagerBindingName function", func() {
		It("Should generate the correct binding name for tenant managers", func() {
			expectedName := "crownlabs-manage-tenants-test-workspace"
			Expect(forge.GetWorkspaceTenantsManagerBindingName(workspace)).To(Equal(expectedName))
		})
	})

	Describe("The forge.ConfigureWorkspaceInstancesManagerBinding function", func() {
		var (
			crb *rbacv1.ClusterRoleBinding
		)

		BeforeEach(func() {
			crb = &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: forge.GetWorkspaceInstancesManagerBindingName(workspace),
				},
			}
		})

		It("Should configure the correct RoleRef and Subject", func() {
			forge.ConfigureWorkspaceInstancesManagerBinding(workspace, crb)

			// Check RoleRef
			Expect(crb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(crb.RoleRef.Name).To(Equal(forge.WorkspaceInstancesManagerRoleName))
			Expect(crb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))

			// Check Subject
			Expect(crb.Subjects).To(HaveLen(1))
			Expect(crb.Subjects[0].Kind).To(Equal("Group"))
			Expect(crb.Subjects[0].Name).To(Equal("kubernetes:workspace-test-workspace:manager"))
			Expect(crb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
		})
	})

	Describe("The forge.ConfigureWorkspaceTenantsManagerBinding function", func() {
		var (
			crb *rbacv1.ClusterRoleBinding
		)

		BeforeEach(func() {
			crb = &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: forge.GetWorkspaceTenantsManagerBindingName(workspace),
				},
			}
		})

		It("Should configure the correct RoleRef and Subject", func() {
			forge.ConfigureWorkspaceTenantsManagerBinding(workspace, crb)

			// Check RoleRef
			Expect(crb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(crb.RoleRef.Name).To(Equal(forge.WorkspaceTenantsManagerRoleName))
			Expect(crb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))

			// Check Subject
			Expect(crb.Subjects).To(HaveLen(1))
			Expect(crb.Subjects[0].Kind).To(Equal("Group"))
			Expect(crb.Subjects[0].Name).To(Equal("kubernetes:workspace-test-workspace:manager"))
			Expect(crb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
		})
	})

	Describe("The forge.ResourceObjectMeta function", func() {
		It("Should return the correct ObjectMeta", func() {
			name := "test-resource"
			labels := map[string]string{
				"app":  "crownlabs",
				"tier": "backend",
			}

			meta := forge.ResourceObjectMeta(name, labels)

			Expect(meta.Name).To(Equal(name))
			Expect(meta.Labels).To(Equal(labels))
		})
	})
})

var _ = Describe("Tenant RBAC forging", func() {
	var (
		tenant *clv1alpha2.Tenant
		labels map[string]string
	)

	BeforeEach(func() {
		tenant = &clv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-tenant",
			},
			Spec: clv1alpha2.TenantSpec{
				FirstName: "Test",
				LastName:  "Tenant",
			},
		}

		labels = map[string]string{
			"app": "test-app",
			"env": "testing",
		}
	})

	Describe("The forge.GetTenantClusterRoleResourceName function", func() {
		It("Should generate the correct resource name for tenant cluster role", func() {
			expectedName := "crownlabs-manage-tenant-test-tenant"
			Expect(forge.GetTenantClusterRoleResourceName(tenant)).To(Equal(expectedName))
		})
	})

	Describe("The forge.ConfigureTenantClusterRole function", func() {
		It("Should correctly configure a ClusterRole for tenant access", func() {
			cr := &rbacv1.ClusterRole{}

			forge.ConfigureTenantClusterRole(cr, tenant, labels)

			// Verify that the labels are properly set
			Expect(cr.Labels).ToNot(BeNil())
			Expect(cr.Labels).To(HaveKeyWithValue("app", "test-app"))
			Expect(cr.Labels).To(HaveKeyWithValue("env", "testing"))

			// Verify rules are correctly set
			Expect(cr.Rules).To(HaveLen(1))
			Expect(cr.Rules[0].APIGroups).To(ConsistOf("crownlabs.polito.it"))
			Expect(cr.Rules[0].Resources).To(ConsistOf("tenants"))
			Expect(cr.Rules[0].ResourceNames).To(ConsistOf("test-tenant"))
			Expect(cr.Rules[0].Verbs).To(ConsistOf("get", "list", "watch", "patch", "update"))
		})

		It("Should initialize labels if nil", func() {
			cr := &rbacv1.ClusterRole{}

			forge.ConfigureTenantClusterRole(cr, tenant, labels)

			Expect(cr.Labels).ToNot(BeNil())
			Expect(cr.Labels).To(HaveKeyWithValue("app", "test-app"))
			Expect(cr.Labels).To(HaveKeyWithValue("env", "testing"))
		})

		It("Should preserve existing labels", func() {
			cr := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
			}

			forge.ConfigureTenantClusterRole(cr, tenant, labels)

			Expect(cr.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(cr.Labels).To(HaveKeyWithValue("app", "test-app"))
			Expect(cr.Labels).To(HaveKeyWithValue("env", "testing"))
		})
	})

	Describe("The forge.ConfigureTenantClusterRoleBinding function", func() {
		It("Should correctly configure a ClusterRoleBinding for tenant access", func() {
			crb := &rbacv1.ClusterRoleBinding{}

			forge.ConfigureTenantClusterRoleBinding(crb, tenant, labels)

			// Verify labels
			Expect(crb.Labels).ToNot(BeNil())
			Expect(crb.Labels).To(HaveKeyWithValue("app", "test-app"))
			Expect(crb.Labels).To(HaveKeyWithValue("env", "testing"))

			// Verify RoleRef
			Expect(crb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(crb.RoleRef.Name).To(Equal("crownlabs-manage-tenant-test-tenant"))
			Expect(crb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))

			// Verify Subjects
			Expect(crb.Subjects).To(HaveLen(1))
			Expect(crb.Subjects[0].Kind).To(Equal("User"))
			Expect(crb.Subjects[0].Name).To(Equal("test-tenant"))
			Expect(crb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
		})

		It("Should initialize labels if nil", func() {
			crb := &rbacv1.ClusterRoleBinding{}

			forge.ConfigureTenantClusterRoleBinding(crb, tenant, labels)

			Expect(crb.Labels).ToNot(BeNil())
			Expect(crb.Labels).To(HaveKeyWithValue("app", "test-app"))
			Expect(crb.Labels).To(HaveKeyWithValue("env", "testing"))
		})

		It("Should preserve existing labels", func() {
			crb := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
			}

			forge.ConfigureTenantClusterRoleBinding(crb, tenant, labels)

			Expect(crb.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(crb.Labels).To(HaveKeyWithValue("app", "test-app"))
			Expect(crb.Labels).To(HaveKeyWithValue("env", "testing"))
		})
	})

	Describe("The forge.ConfigureTenantInstancesRoleBinding function", func() {
		It("Should correctly configure a RoleBinding for tenant instances", func() {
			rb := &rbacv1.RoleBinding{}

			forge.ConfigureTenantInstancesRoleBinding(rb, tenant, labels)

			// Verify labels
			Expect(rb.Labels).ToNot(BeNil())
			Expect(rb.Labels).To(HaveKeyWithValue("app", "test-app"))
			Expect(rb.Labels).To(HaveKeyWithValue("env", "testing"))

			// Verify RoleRef
			Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(rb.RoleRef.Name).To(Equal("crownlabs-manage-instances"))
			Expect(rb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))

			// Verify Subjects
			Expect(rb.Subjects).To(HaveLen(1))
			Expect(rb.Subjects[0].Kind).To(Equal("User"))
			Expect(rb.Subjects[0].Name).To(Equal("test-tenant"))
			Expect(rb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
		})

		It("Should initialize labels if nil", func() {
			rb := &rbacv1.RoleBinding{}

			forge.ConfigureTenantInstancesRoleBinding(rb, tenant, labels)

			Expect(rb.Labels).ToNot(BeNil())
			Expect(rb.Labels).To(HaveKeyWithValue("app", "test-app"))
			Expect(rb.Labels).To(HaveKeyWithValue("env", "testing"))
		})

		It("Should preserve existing labels", func() {
			rb := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
			}

			forge.ConfigureTenantInstancesRoleBinding(rb, tenant, labels)

			Expect(rb.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(rb.Labels).To(HaveKeyWithValue("app", "test-app"))
			Expect(rb.Labels).To(HaveKeyWithValue("env", "testing"))
		})
	})
})
