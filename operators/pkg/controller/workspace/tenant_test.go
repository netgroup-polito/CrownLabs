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
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var tenantResources = []*v1alpha2.Tenant{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-candidate",
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector":                 "test",
				fmt.Sprintf("crownlabs.polito.it/workspace-%s", wsName): "candidate",
			},
		},
		Spec: v1alpha2.TenantSpec{
			FirstName: "Candidate",
			LastName:  "Tenant",
			Email:     "candidate-tenant@test.email",
			Workspaces: []v1alpha2.TenantWorkspaceEntry{{
				Name: wsName,
				Role: v1alpha2.Candidate,
			}},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector":                 "test",
				fmt.Sprintf("crownlabs.polito.it/workspace-%s", wsName): "user",
			},
		},
		Spec: v1alpha2.TenantSpec{
			FirstName: "User",
			LastName:  "Tenant",
			Email:     "user-tenant@test.email",
			Workspaces: []v1alpha2.TenantWorkspaceEntry{{
				Name: wsName,
				Role: v1alpha2.User,
			}},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-manager",
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector":                 "test",
				fmt.Sprintf("crownlabs.polito.it/workspace-%s", wsName): "manager",
			},
		},
		Spec: v1alpha2.TenantSpec{
			FirstName: "Manager",
			LastName:  "Tenant",
			Email:     "manager-tenant@test.email",
			Workspaces: []v1alpha2.TenantWorkspaceEntry{{
				Name: wsName,
				Role: v1alpha2.Manager,
			}},
		},
	},
}

var _ = Describe("Tenant", func() {
	Context("When the workspace is being deleted", func() {
		BeforeEach(func() {
			workspaceBeingDeleted()
			for _, tenant := range tenantResources {
				addObjToObjectsList(tenant)
			}
		})

		AfterEach(func() {
			for _, tenant := range tenantResources {
				removeObjFromObjectsList(tenant)
			}
		})

		It("Should remove the workspace from the Tenant spec", func() {
			tenant := &v1alpha2.Tenant{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tenantResources[0].Name}, tenant, BeTrue(), timeout, interval)

			Expect(tenant.Spec.Workspaces).To(BeEmpty())
		})

		Context("When there is a failure retrieving the Tenants subscribed to the Workspace", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					List: func(_ context.Context, _ client.WithWatch, list client.ObjectList, _ ...client.ListOption) error {
						if _, ok := list.(*v1alpha2.TenantList); ok {
							return fmt.Errorf("failed to list tenants")
						}
						return nil
					},
				})

				wsReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("should prevent the workspace from being deleted", func() {
				ws := &v1alpha1.Workspace{}

				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

				Expect(ws.Status.Ready).To(BeFalse())
			})
		})

		Context("When there is a failure in removing the workspace from the Tenant spec", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Patch: func(_ context.Context, _ client.WithWatch, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
						if tenant, ok := obj.(*v1alpha2.Tenant); ok && tenant.Name == tenantResources[0].Name {
							return fmt.Errorf("failed to update tenant")
						}
						return nil
					},
				})

				wsReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("should prevent the workspace from being deleted", func() {
				ws := &v1alpha1.Workspace{}

				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

				Expect(ws.Status.Ready).To(BeFalse())
			})
		})
	})
})
