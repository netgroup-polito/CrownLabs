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
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

var _ = Describe("WorkspaceReconciler", func() {
	Context("When there is an error getting the resource from the cluster", func() {
		BeforeEach(func() {
			builder.WithInterceptorFuncs(interceptor.Funcs{
				Get: func(_ context.Context, _ client.WithWatch, key client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
					if key.Name == wsName {
						return fmt.Errorf("error getting resource %s", key.Name)
					}
					return nil
				},
			})

			wsReconcileErrExpected = HaveOccurred()
		})

		It("Should return an error", func() {
			// already set in BeforeEach
		})
	})

	Context("When the resource is not found in the cluster", func() {
		BeforeEach(func() {
			wsResource = &v1alpha1.Workspace{}
		})

		It("Should not return an error", func() {
			// because the resource is just been deleted

			// already set in BeforeEach
		})
	})

	Context("When the resource is not responsibility of the current reconciler", func() {
		BeforeEach(func() {
			wsResource.Labels["crownlabs.polito.it/operator-selector"] = "other"
		})

		It("Should not return an error", func() {
			// nothing to do, just skip the resource
		})

		It("Should not create any resources", func() {
			ns := &v1.Namespace{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "workspace-" + wsName}, ns, BeFalse(), timeout, interval)
		})

		Context("When there are resources associated with the workspace", func() {
			BeforeEach(func() {
				namespaceResource.Labels["crownlabs.polito.it/operator-selector"] = "other"
				addObjToObjectsList(namespaceResource)
			})

			AfterEach(func() {
				removeObjFromObjectsList(namespaceResource)
				namespaceResource.Labels["crownlabs.polito.it/operator-selector"] = "test"
			})

			It("Should not delete the resources", func() {
				ns := &v1.Namespace{}

				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "workspace-" + wsName}, ns, BeTrue(), timeout, interval)

				Expect(ns.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "other"))
			})
		})
	})

	It("Should add the finalizer to the workspace resource", func() {
		ws := &v1alpha1.Workspace{}

		DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

		Expect(ws.Finalizers).To(ContainElement("crownlabs.polito.it/tenant-operator"))
	})

	Context("When there is an error adding the finalizer", func() {
		BeforeEach(func() {
			builder.WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(_ context.Context, _ client.WithWatch, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
					if ws, ok := obj.(*v1alpha1.Workspace); ok && ws.Name == wsName {
						return fmt.Errorf("failed to add finalizer")
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

	Context("When the workspace is being deleted", func() {
		BeforeEach(func() {
			workspaceBeingDeleted()
		})

		Context("When there is an error removing the finalizer", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Patch: func(_ context.Context, _ client.WithWatch, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
						if ws, ok := obj.(*v1alpha1.Workspace); ok && ws.Name == wsName {
							return fmt.Errorf("failed to remove finalizer")
						}
						return nil
					},
				})

				wsReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("Should prevent the deletion of the workspace resource", func() {
				ws := &v1alpha1.Workspace{}

				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)
			})
		})
	})
})
