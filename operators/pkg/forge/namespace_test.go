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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Namespace forging", func() {
	var (
		workspace *v1alpha1.Workspace
		labels    map[string]string
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

		labels = map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
	})

	Describe("The forge.GetWorkspaceNamespaceName function", func() {
		It("Should return the correct namespace name", func() {
			expectedNamespaceName := "workspace-test-workspace"
			Expect(forge.GetWorkspaceNamespaceName(workspace)).To(Equal(expectedNamespaceName))
		})
	})

	Describe("The forge.ConfigureWorkspaceNamespace function", func() {
		var namespace *corev1.Namespace

		BeforeEach(func() {
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: forge.GetWorkspaceNamespaceName(workspace),
				},
			}
		})

		Context("When the namespace has no labels", func() {
			It("Should initialize the labels map and set all labels", func() {
				forge.ConfigureWorkspaceNamespace(namespace, labels)

				// Check that all provided labels are set
				for k, v := range labels {
					Expect(namespace.Labels).To(HaveKeyWithValue(k, v))
				}

				// Check that the workspace type label is set
				Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/type", "workspace"))
			})
		})

		Context("When the namespace already has labels", func() {
			BeforeEach(func() {
				namespace.Labels = map[string]string{
					"existing": "label",
				}
			})

			It("Should keep existing labels and add new ones", func() {
				forge.ConfigureWorkspaceNamespace(namespace, labels)

				// Check that existing label is preserved
				Expect(namespace.Labels).To(HaveKeyWithValue("existing", "label"))

				// Check that all provided labels are set
				for k, v := range labels {
					Expect(namespace.Labels).To(HaveKeyWithValue(k, v))
				}

				// Check that the workspace type label is set
				Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/type", "workspace"))
			})
		})
	})

	var _ = Describe("GetTenantNamespaceName", func() {
		It("Should format namespace name correctly for simple tenant name", func() {
			tenant := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "student",
				},
			}

			namespaceName := forge.GetTenantNamespaceName(tenant)
			Expect(namespaceName).To(Equal("tenant-student"))
		})

		It("Should replace dots with dashes in tenant name", func() {
			tenant := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "s123456.student",
				},
			}

			namespaceName := forge.GetTenantNamespaceName(tenant)
			Expect(namespaceName).To(Equal("tenant-s123456-student"))
		})
	})

	var _ = Describe("ConfigureTenantNamespace", func() {
		It("Should initialize labels if nil and set required labels", func() {
			tenant := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "student",
				},
			}

			namespace := &corev1.Namespace{}
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantNamespace(namespace, tenant, labels)

			Expect(namespace.Labels).ToNot(BeNil())
			Expect(namespace.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/type", "tenant"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/name", "student"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/instance-resources-replication", "true"))
		})

		It("Should preserve existing labels and add new ones", func() {
			tenant := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "student",
				},
			}

			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
			}

			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantNamespace(namespace, tenant, labels)

			Expect(namespace.Labels).ToNot(BeNil())
			Expect(namespace.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(namespace.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/type", "tenant"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/name", "student"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/instance-resources-replication", "true"))
		})
	})

	var _ = Describe("GetTenantNamespaceName", func() {
		It("Should format namespace name correctly for simple tenant name", func() {
			tenant := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "student",
				},
			}

			namespaceName := forge.GetTenantNamespaceName(tenant)
			Expect(namespaceName).To(Equal("tenant-student"))
		})

		It("Should replace dots with dashes in tenant name", func() {
			tenant := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "s123456.student",
				},
			}

			namespaceName := forge.GetTenantNamespaceName(tenant)
			Expect(namespaceName).To(Equal("tenant-s123456-student"))
		})
	})

	var _ = Describe("ConfigureTenantNamespace", func() {
		It("Should initialize labels if nil and set required labels", func() {
			tenant := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "student",
				},
			}

			namespace := &corev1.Namespace{}
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantNamespace(namespace, tenant, labels)

			Expect(namespace.Labels).ToNot(BeNil())
			Expect(namespace.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/type", "tenant"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/name", "student"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/instance-resources-replication", "true"))
		})

		It("Should preserve existing labels and add new ones", func() {
			tenant := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "student",
				},
			}

			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
			}

			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantNamespace(namespace, tenant, labels)

			Expect(namespace.Labels).ToNot(BeNil())
			Expect(namespace.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(namespace.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/type", "tenant"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/name", "student"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/instance-resources-replication", "true"))
		})
	})
})
