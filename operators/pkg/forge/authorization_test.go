// Copyright 2020-2026 Politecnico di Torino
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

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Authorization forging", func() {
	const publicNamespace = "crownlabs-public"

	// The tenant name contains a dot on purpose: GetTenantNamespaceName rewrites it into a dash,
	// which is what makes the namespace-to-tenant mapping lossy and forward derivation mandatory.
	tenant := func() *clv1alpha2.Tenant {
		return &clv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{Name: "john.doe"},
			Spec: clv1alpha2.TenantSpec{
				Workspaces: []clv1alpha2.TenantWorkspaceEntry{
					{Name: "user-ws", Role: clv1alpha2.User},
					{Name: "manager-ws", Role: clv1alpha2.Manager},
					{Name: "candidate-ws", Role: clv1alpha2.Candidate},
					{Name: "failing-ws", Role: clv1alpha2.User},
				},
			},
			Status: clv1alpha2.TenantStatus{
				FailingWorkspaces: []string{"failing-ws"},
			},
		}
	}

	Describe("The forge.EnrolledWorkspaces function", func() {
		It("Should leave out candidates and failing workspaces", func() {
			names := []string{}
			for _, ws := range forge.EnrolledWorkspaces(tenant()) {
				names = append(names, ws.Name)
			}

			Expect(names).To(ConsistOf("user-ws", "manager-ws"))
		})
	})

	Describe("The forge.TenantCanReadNamespace function", func() {
		type ReadNamespaceCase struct {
			Namespace       string
			PublicNamespace string
			Expected        bool
		}

		DescribeTable("Correctly tells whether the tenant can read from the namespace",
			func(c ReadNamespaceCase) {
				Expect(forge.TenantCanReadNamespace(tenant(), c.Namespace, c.PublicNamespace)).To(Equal(c.Expected))
			},
			Entry("When the namespace is the tenant one", ReadNamespaceCase{
				Namespace: "tenant-john-doe", PublicNamespace: publicNamespace, Expected: true,
			}),
			Entry("When the namespace belongs to another tenant", ReadNamespaceCase{
				Namespace: "tenant-jane-doe", PublicNamespace: publicNamespace, Expected: false,
			}),
			Entry("When the namespace spells the tenant name verbatim instead of deriving it", ReadNamespaceCase{
				Namespace: "tenant-john.doe", PublicNamespace: publicNamespace, Expected: false,
			}),
			Entry("When the namespace is a workspace the tenant is a user of", ReadNamespaceCase{
				Namespace: "workspace-user-ws", PublicNamespace: publicNamespace, Expected: true,
			}),
			Entry("When the namespace is a workspace the tenant is a manager of", ReadNamespaceCase{
				Namespace: "workspace-manager-ws", PublicNamespace: publicNamespace, Expected: true,
			}),
			Entry("When the namespace is a workspace the tenant is only a candidate of", ReadNamespaceCase{
				Namespace: "workspace-candidate-ws", PublicNamespace: publicNamespace, Expected: false,
			}),
			Entry("When the namespace is a workspace that failed to be processed", ReadNamespaceCase{
				Namespace: "workspace-failing-ws", PublicNamespace: publicNamespace, Expected: false,
			}),
			Entry("When the namespace is a workspace the tenant is not subscribed to", ReadNamespaceCase{
				Namespace: "workspace-unrelated-ws", PublicNamespace: publicNamespace, Expected: false,
			}),
			Entry("When the namespace is the public catalog", ReadNamespaceCase{
				Namespace: publicNamespace, PublicNamespace: publicNamespace, Expected: true,
			}),
			Entry("When the public catalog is not configured", ReadNamespaceCase{
				Namespace: publicNamespace, PublicNamespace: "", Expected: false,
			}),
			Entry("When the namespace is empty and the public catalog is not configured", ReadNamespaceCase{
				Namespace: "", PublicNamespace: "", Expected: false,
			}),
		)
	})
})
