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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Keycloak forging", func() {
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

	Context("WorkspaceRoleName", func() {
		It("should return the correct role name for a workspace if the role is Manager", func() {
			ws := "test-workspace"
			role := v1alpha2.Manager
			expectedRoleName := "workspace-test-workspace:manager"

			roleName := forge.WorkspaceRoleName(ws, role)
			Expect(roleName).To(Equal(expectedRoleName))
		})

		It("should return the correct role name for a workspace if the role is User", func() {
			ws := "test-workspace"
			role := v1alpha2.User
			expectedRoleName := "workspace-test-workspace:user"

			roleName := forge.WorkspaceRoleName(ws, role)
			Expect(roleName).To(Equal(expectedRoleName))
		})
	})

	Describe("The forge.GetWorkspaceManagerRoleName function", func() {
		It("Should return the correct manager role name", func() {
			expectedRoleName := "workspace-test-workspace:manager"
			Expect(forge.GetWorkspaceManagerRoleName(workspace)).To(Equal(expectedRoleName))
		})
	})

	Describe("The forge.GetWorkspaceManagerRoleDescription function", func() {
		It("Should return the correct manager role description", func() {
			expectedDescription := "Test Workspace Manager Role"
			Expect(forge.GetWorkspaceManagerRoleDescription(workspace)).To(Equal(expectedDescription))
		})

		Context("When the workspace has a different pretty name", func() {
			BeforeEach(func() {
				workspace.Spec.PrettyName = "Another Name"
			})

			It("Should use the workspace pretty name in the description", func() {
				expectedDescription := "Another Name Manager Role"
				Expect(forge.GetWorkspaceManagerRoleDescription(workspace)).To(Equal(expectedDescription))
			})
		})
	})

	Describe("The forge.GetWorkspaceUserRoleName function", func() {
		It("Should return the correct user role name", func() {
			expectedRoleName := "workspace-test-workspace:user"
			Expect(forge.GetWorkspaceUserRoleName(workspace)).To(Equal(expectedRoleName))
		})
	})

	Describe("The forge.GetWorkspaceUserRoleDescription function", func() {
		It("Should return the correct user role description", func() {
			expectedDescription := "Test Workspace User Role"
			Expect(forge.GetWorkspaceUserRoleDescription(workspace)).To(Equal(expectedDescription))
		})

		Context("When the workspace has a different pretty name", func() {
			BeforeEach(func() {
				workspace.Spec.PrettyName = "Another Name"
			})

			It("Should use the workspace pretty name in the description", func() {
				expectedDescription := "Another Name User Role"
				Expect(forge.GetWorkspaceUserRoleDescription(workspace)).To(Equal(expectedDescription))
			})
		})
	})

	Describe("Integration with common.WorkspaceRoleName", func() {
		It("Should use the correct format for manager role", func() {
			managerRole := forge.GetWorkspaceManagerRoleName(workspace)
			Expect(managerRole).To(ContainSubstring(workspace.Name))
			Expect(managerRole).To(ContainSubstring(string(v1alpha2.Manager)))
		})

		It("Should use the correct format for user role", func() {
			userRole := forge.GetWorkspaceUserRoleName(workspace)
			Expect(userRole).To(ContainSubstring(workspace.Name))
			Expect(userRole).To(ContainSubstring(string(v1alpha2.User)))
		})
	})
})
