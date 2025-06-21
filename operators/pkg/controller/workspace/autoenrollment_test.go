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

package workspace_test

import (
	"fmt"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var tenantResource = &v1alpha2.Tenant{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-tenant",
		Labels: map[string]string{
			"crownlabs.polito.it/operator-selector":                 "test",
			fmt.Sprintf("crownlabs.polito.it/workspace-%s", wsName): "candidate",
		},
	},
	Spec: v1alpha2.TenantSpec{
		FirstName: "Test",
		LastName:  "Tenant",
		Email:     "tenant@test.email",
		Workspaces: []v1alpha2.TenantWorkspaceEntry{{
			Name: wsName,
			Role: v1alpha2.Candidate,
		}},
	},
}

var _ = Describe("AutoEnrollment", func() {
	BeforeEach(func() {
		addObjToObjectsList(tenantResource)
	})

	AfterEach(func() {
		removeObjFromObjectsList(tenantResource)
	})

	Context("When no autoenrollment is configured", func() {
		It("Should set the autoenroll label to 'disabled'", func() {
			ws := &v1alpha1.Workspace{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

			Expect(ws.Labels).To(HaveKeyWithValue("crownlabs.polito.it/autoenroll", "disabled"))
		})

		It("Should update the Tenant spec to remove the workspace", func() {
			tenant := &v1alpha2.Tenant{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tenantResource.Name}, tenant, BeTrue(), timeout, interval)

			Expect(tenant.Spec.Workspaces).To(BeEmpty())
		})
	})

	Context("When autoenrollment is Immediate", func() {
		BeforeEach(func() {
			wsResource.Spec.AutoEnroll = v1alpha1.AutoenrollImmediate
		})

		It("Should set the autoenroll label to 'immediate'", func() {
			ws := &v1alpha1.Workspace{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

			Expect(ws.Labels).To(HaveKeyWithValue("crownlabs.polito.it/autoenroll", "immediate"))
		})

		It("Should update the Tenant spec to set the workspace with User role", func() {
			tenant := &v1alpha2.Tenant{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tenantResource.Name}, tenant, BeTrue(), timeout, interval)

			Expect(tenant.Spec.Workspaces).To(ContainElement(v1alpha2.TenantWorkspaceEntry{
				Name: wsName,
				Role: v1alpha2.User,
			}))
		})
	})

	Context("When autoenrollment is WithApproval", func() {
		BeforeEach(func() {
			wsResource.Spec.AutoEnroll = v1alpha1.AutoenrollWithApproval
		})

		It("Should set the autoenroll label to 'with-approval'", func() {
			ws := &v1alpha1.Workspace{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

			Expect(ws.Labels).To(HaveKeyWithValue("crownlabs.polito.it/autoenroll", "withApproval"))
		})
	})
})
