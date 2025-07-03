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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var namespaceResource = &v1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "workspace-" + wsName,
		Labels: map[string]string{
			"crownlabs.polito.it/operator-selector": "test",
			"crownlabs.polito.it/type":              "workspace",
			"crownlabs.polito.it/managed-by":        "workspace",
		},
	},
}

var _ = Describe("Namespace", func() {
	Context("When a workspace is created", func() {
		It("Should create a namespace with the correct labels", func() {
			ns := &v1.Namespace{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "workspace-" + wsName}, ns, BeTrue(), timeout, interval)

			Expect(ns.Labels).To(HaveKeyWithValue("crownlabs.polito.it/type", "workspace"))
			Expect(ns.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(ns.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "workspace"))
		})

		It("Should update the workspace status with the namespace name", func() {
			ws := &v1alpha1.Workspace{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

			Expect(ws.Status.Namespace.Created).To(BeTrue())
			Expect(ws.Status.Namespace.Name).To(Equal("workspace-" + wsName))
		})

		Context("When there is a failure in creating the namespace", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption) error {
						if ns, ok := obj.(*v1.Namespace); ok && ns.Name == "workspace-"+wsName {
							return fmt.Errorf("failed to create namespace")
						}
						return nil
					},
				})

				wsReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("Should set the namespace status in the workspace to not created", func() {
				ws := &v1alpha1.Workspace{}

				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

				Expect(ws.Status.Namespace.Created).To(BeFalse())
			})

			It("Should set the workspace status to not ready", func() {
				ws := &v1alpha1.Workspace{}

				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

				Expect(ws.Status.Ready).To(BeFalse())
			})
		})
	})

	Context("When a workspace is deleted", func() {
		BeforeEach(func() {
			workspaceBeingDeleted()
		})

		Context("When the namespace exists", func() {
			BeforeEach(func() {
				wsResource.Status.Namespace = v1alpha2.NameCreated{
					Created: true,
					Name:    "workspace-" + wsName,
				}
				addObjToObjectsList(namespaceResource)
			})

			AfterEach(func() {
				removeObjFromObjectsList(namespaceResource)
			})

			It("Should delete the namespace", func() {
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "workspace-" + wsName}, &v1.Namespace{}, BeFalse(), timeout, interval)
			})

			Context("When there is an error deleting the namespace", func() {
				BeforeEach(func() {
					builder.WithInterceptorFuncs(interceptor.Funcs{
						Delete: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.DeleteOption) error {
							if ns, ok := obj.(*v1.Namespace); ok && ns.Name == "workspace-"+wsName {
								return fmt.Errorf("error deleting namespace")
							}
							return nil
						},
					})

					wsReconcileErrExpected = HaveOccurred()
				})

				It("Should return an error", func() {
					// checked in BeforeEach
				})

				It("Should prevent the workspace from being deleted", func() {
					ws := &v1alpha1.Workspace{}

					DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)
				})

				It("Should not remove the namespace from the workspace status", func() {
					ws := &v1alpha1.Workspace{}

					DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

					Expect(ws.Status.Namespace.Created).To(BeTrue())
					Expect(ws.Status.Namespace.Name).To(Equal("workspace-" + wsName))
				})

				It("Should set the workspace status to not ready", func() {
					ws := &v1alpha1.Workspace{}

					DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

					Expect(ws.Status.Ready).To(BeFalse())
				})
			})
		})

		Context("When the namespace does not exist", func() {
			It("Should not return an error", func() {
				// nothing to check here, the absence of the namespace is expected
			})
		})
	})
})
