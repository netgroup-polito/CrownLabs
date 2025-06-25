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

package common

import (
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Context("WorkspaceRoleName", func() {
		It("should return the correct role name for a workspace if the role is Manager", func() {
			ws := "test-workspace"
			role := v1alpha2.Manager
			expectedRoleName := "workspace-test-workspace:manager"

			roleName := WorkspaceRoleName(ws, role)
			Expect(roleName).To(Equal(expectedRoleName))
		})

		It("should return the correct role name for a workspace if the role is User", func() {
			ws := "test-workspace"
			role := v1alpha2.User
			expectedRoleName := "workspace-test-workspace:user"

			roleName := WorkspaceRoleName(ws, role)
			Expect(roleName).To(Equal(expectedRoleName))
		})
	})
})
