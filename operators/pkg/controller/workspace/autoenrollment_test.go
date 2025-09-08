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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("AutoEnrollment", func() {
	BeforeEach(func() {
		addObjToObjectsList(tenantResources[0])
	})

	AfterEach(func() {
		removeObjFromObjectsList(tenantResources[0])
	})

	Context("When no autoenrollment is configured", func() {
		It("Should set the autoenroll label to 'disabled'", func() {
			ws := &v1alpha1.Workspace{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

			Expect(ws.Labels).To(HaveKeyWithValue("crownlabs.polito.it/autoenroll", "disabled"))
		})

		It("Should update the Tenant spec to remove the workspace", func() {
			tenant := &v1alpha2.Tenant{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tenantResources[0].Name}, tenant, BeTrue(), timeout, interval)

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

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tenantResources[0].Name}, tenant, BeTrue(), timeout, interval)

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

	Context("When there is an error updating the label", func() {
		BeforeEach(func() {
			builder.WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(_ context.Context, _ client.WithWatch, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
					if ws, ok := obj.(*v1alpha1.Workspace); ok && ws.Name == wsName && ws.Labels["crownlabs.polito.it/autoenroll"] == "disabled" {
						return fmt.Errorf("error updating workspace label")
					}
					return nil
				},
			})

			wsReconcileErrExpected = HaveOccurred()
		})

		It("Should return an error", func() {
			// checked in BeforeEach
		})
	})

	Context("When there is an error listing the Tenants subscribed to the Workspace", func() {
		BeforeEach(func() {
			builder.WithInterceptorFuncs(interceptor.Funcs{
				List: func(_ context.Context, _ client.WithWatch, list client.ObjectList, _ ...client.ListOption) error {
					if _, ok := list.(*v1alpha2.TenantList); ok {
						return fmt.Errorf("error listing tenants")
					}
					return nil
				},
			})

			wsResource.Spec.AutoEnroll = v1alpha1.AutoenrollImmediate // Ensure autoenrollment is set to Immediate for the test

			wsReconcileErrExpected = HaveOccurred()
		})

		It("Should return an error", func() {
			// checked in BeforeEach
		})
	})

	Context("When there is an error updating the Tenant spec", func() {
		BeforeEach(func() {
			builder.WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(_ context.Context, _ client.WithWatch, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
					if tenant, ok := obj.(*v1alpha2.Tenant); ok && tenant.Name == tenantResources[0].Name {
						return fmt.Errorf("error updating tenant spec")
					}
					return nil
				},
			})

			wsResource.Spec.AutoEnroll = v1alpha1.AutoenrollImmediate // Ensure autoenrollment is set to Immediate for the test

			wsReconcileErrExpected = HaveOccurred()
		})

		It("Should return an error", func() {
			// checked in BeforeEach
		})

		It("Should set the workspace status to not ready", func() {
			ws := &v1alpha1.Workspace{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

			Expect(ws.Status.Ready).To(BeFalse())
		})
	})
})
