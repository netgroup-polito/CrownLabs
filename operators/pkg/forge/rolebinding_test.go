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
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("RoleBinding forging", func() {
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
			"app":  "crownlabs",
			"tier": "backend",
		}
	})

	Describe("The forge.ConfigureWorkspaceUserViewTemplatesBinding function", func() {
		var (
			rb *rbacv1.RoleBinding
		)

		BeforeEach(func() {
			rb = &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-view-templates",
					Namespace: "default",
				},
			}
		})

		Context("When the RoleBinding has no labels", func() {
			It("Should set the correct labels, RoleRef and Subject", func() {
				forge.ConfigureWorkspaceUserViewTemplatesBinding(workspace, rb, labels)

				// Check labels
				for k, v := range labels {
					Expect(rb.Labels).To(HaveKeyWithValue(k, v))
				}

				// Check RoleRef
				Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
				Expect(rb.RoleRef.Name).To(Equal(forge.ViewTemplatesRoleName))
				Expect(rb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))

				// Check Subject
				Expect(rb.Subjects).To(HaveLen(1))
				Expect(rb.Subjects[0].Kind).To(Equal("Group"))
				Expect(rb.Subjects[0].Name).To(Equal("kubernetes:workspace-test-workspace:user"))
				Expect(rb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
			})
		})

		Context("When the RoleBinding already has labels", func() {
			BeforeEach(func() {
				rb.Labels = map[string]string{
					"existing": "label",
				}
			})

			It("Should keep existing labels and add new ones", func() {
				forge.ConfigureWorkspaceUserViewTemplatesBinding(workspace, rb, labels)

				Expect(rb.Labels).To(HaveKeyWithValue("existing", "label"))
				for k, v := range labels {
					Expect(rb.Labels).To(HaveKeyWithValue(k, v))
				}
			})
		})
	})

	Describe("The forge.ConfigureWorkspaceManagerManageTemplatesBinding function", func() {
		var (
			rb *rbacv1.RoleBinding
		)

		BeforeEach(func() {
			rb = &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-manage-templates",
					Namespace: "default",
				},
			}
		})

		Context("When the RoleBinding has no labels", func() {
			It("Should set the correct labels, RoleRef and Subject", func() {
				forge.ConfigureWorkspaceManagerManageTemplatesBinding(workspace, rb, labels)

				// Check labels
				for k, v := range labels {
					Expect(rb.Labels).To(HaveKeyWithValue(k, v))
				}

				// Check RoleRef
				Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
				Expect(rb.RoleRef.Name).To(Equal(forge.ManageTemplatesRoleName))
				Expect(rb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))

				// Check Subject
				Expect(rb.Subjects).To(HaveLen(1))
				Expect(rb.Subjects[0].Kind).To(Equal("Group"))
				Expect(rb.Subjects[0].Name).To(Equal("kubernetes:workspace-test-workspace:manager"))
				Expect(rb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
			})
		})

		Context("When the RoleBinding already has labels", func() {
			BeforeEach(func() {
				rb.Labels = map[string]string{
					"existing": "label",
				}
			})

			It("Should keep existing labels and add new ones", func() {
				forge.ConfigureWorkspaceManagerManageTemplatesBinding(workspace, rb, labels)

				Expect(rb.Labels).To(HaveKeyWithValue("existing", "label"))
				for k, v := range labels {
					Expect(rb.Labels).To(HaveKeyWithValue(k, v))
				}
			})
		})
	})

	Describe("The forge.ConfigureWorkspaceManagerManageSharedVolumesBinding function", func() {
		var (
			rb *rbacv1.RoleBinding
		)

		BeforeEach(func() {
			rb = &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-manage-sharedvolumes",
					Namespace: "default",
				},
			}
		})

		Context("When the RoleBinding has no labels", func() {
			It("Should set the correct labels, RoleRef and Subject", func() {
				forge.ConfigureWorkspaceManagerManageSharedVolumesBinding(workspace, rb, labels)

				// Check labels
				for k, v := range labels {
					Expect(rb.Labels).To(HaveKeyWithValue(k, v))
				}

				// Check RoleRef
				Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
				Expect(rb.RoleRef.Name).To(Equal(forge.ManageSharedVolumesRoleName))
				Expect(rb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))

				// Check Subject
				Expect(rb.Subjects).To(HaveLen(1))
				Expect(rb.Subjects[0].Kind).To(Equal("Group"))
				Expect(rb.Subjects[0].Name).To(Equal("kubernetes:workspace-test-workspace:manager"))
				Expect(rb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
			})
		})

		Context("When the RoleBinding already has labels", func() {
			BeforeEach(func() {
				rb.Labels = map[string]string{
					"existing": "label",
				}
			})

			It("Should keep existing labels and add new ones", func() {
				forge.ConfigureWorkspaceManagerManageSharedVolumesBinding(workspace, rb, labels)

				Expect(rb.Labels).To(HaveKeyWithValue("existing", "label"))
				for k, v := range labels {
					Expect(rb.Labels).To(HaveKeyWithValue(k, v))
				}
			})
		})
	})
})
