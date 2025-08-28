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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var roleBindingResources = []*rbacv1.RoleBinding{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-view-templates",
			Namespace: "workspace-" + wsName,
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector": "test",
				"crownlabs.polito.it/managed-by":        "workspace",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "Group",
				Name:     "kubernetes:workspace-" + wsName + ":user",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "crownlabs-view-templates",
			APIGroup: "rbac.authorization.k8s.io",
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-manage-templates",
			Namespace: "workspace-" + wsName,
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector": "test",
				"crownlabs.polito.it/managed-by":        "workspace",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "Group",
				Name:     "kubernetes:workspace-" + wsName + ":manager",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "crownlabs-manage-templates",
			APIGroup: "rbac.authorization.k8s.io",
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-manage-sharedvolumes",
			Namespace: "workspace-" + wsName,
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector": "test",
				"crownlabs.polito.it/managed-by":        "workspace",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "Group",
				Name:     "kubernetes:workspace-" + wsName + ":manager",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "crownlabs-manage-sharedvolumes",
			APIGroup: "rbac.authorization.k8s.io",
		},
	},
}

var _ = Describe("RoleBinding", func() {
	Context("When a workspace is created", func() {
		It("Should create a rolebinding for the workspace users to view templates", func() {
			rb := &rbacv1.RoleBinding{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-view-templates",
				Namespace: "workspace-" + wsName,
			}, rb, BeTrue(), timeout, interval)
			Expect(rb.RoleRef.Name).To(Equal("crownlabs-view-templates"))
			Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(rb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(rb.Subjects).To(HaveLen(1))
			Expect(rb.Subjects).To(ContainElement(rbacv1.Subject{
				Kind:     "Group",
				Name:     "kubernetes:workspace-" + wsName + ":user",
				APIGroup: "rbac.authorization.k8s.io",
			}))
		})

		It("Should create a rolebinding for the workspace managers to manage templates", func() {
			rb := &rbacv1.RoleBinding{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-manage-templates",
				Namespace: "workspace-" + wsName,
			}, rb, BeTrue(), timeout, interval)
			Expect(rb.RoleRef.Name).To(Equal("crownlabs-manage-templates"))
			Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(rb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(rb.Subjects).To(HaveLen(1))
			Expect(rb.Subjects).To(ContainElement(rbacv1.Subject{
				Kind:     "Group",
				Name:     "kubernetes:workspace-" + wsName + ":manager",
				APIGroup: "rbac.authorization.k8s.io",
			}))
		})

		It("Should create a rolebinding for the workspace managers to manage sharedvolumes", func() {
			rb := &rbacv1.RoleBinding{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-manage-sharedvolumes",
				Namespace: "workspace-" + wsName,
			}, rb, BeTrue(), timeout, interval)
			Expect(rb.RoleRef.Name).To(Equal("crownlabs-manage-sharedvolumes"))
			Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(rb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(rb.Subjects).To(HaveLen(1))
			Expect(rb.Subjects).To(ContainElement(rbacv1.Subject{
				Kind:     "Group",
				Name:     "kubernetes:workspace-" + wsName + ":manager",
				APIGroup: "rbac.authorization.k8s.io",
			}))
		})

		Context("When there is a failure in creating the rolebinding", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption) error {
						if rb, ok := obj.(*rbacv1.RoleBinding); ok && rb.Name == "crownlabs-view-templates" && rb.Namespace == "workspace-"+wsName {
							return fmt.Errorf("failed to create rolebinding")
						}
						return nil
					},
				})

				wsReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("Should set the workspace status as not ready", func() {
				ws := &v1alpha1.Workspace{}

				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)

				Expect(ws.Status.Ready).To(BeFalse())
			})
		})
	})

	Context("When the workspace is being deleted", func() {
		BeforeEach(func() {
			workspaceBeingDeleted()
		})

		Context("When the rolebindings exist", func() {
			BeforeEach(func() {
				wsResource.Status.Namespace = v1alpha2.NameCreated{
					Created: true,
					Name:    "workspace-" + wsName,
				}
				addObjToObjectsList(namespaceResource)
				for _, rb := range roleBindingResources {
					addObjToObjectsList(rb)
				}
			})

			AfterEach(func() {
				for _, rb := range roleBindingResources {
					removeObjFromObjectsList(rb)
				}
				removeObjFromObjectsList(namespaceResource)
			})

			It("Should delete the rolebindings", func() {
				for _, rb := range roleBindingResources {
					rbResource := &rbacv1.RoleBinding{}
					DoesEventuallyExists(ctx, cl, client.ObjectKey{
						Name:      rb.Name,
						Namespace: rb.Namespace,
					}, rbResource, BeFalse(), timeout, interval)
				}
			})

			Context("When there is an error deleting a rolebinding", func() {
				BeforeEach(func() {
					builder.WithInterceptorFuncs(interceptor.Funcs{
						Delete: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.DeleteOption) error {
							if rb, ok := obj.(*rbacv1.RoleBinding); ok && rb.Name == "crownlabs-view-templates" && rb.Namespace == "workspace-"+wsName {
								return fmt.Errorf("error deleting rolebinding")
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

					Expect(ws.Status.Ready).To(BeFalse())
				})
			})
		})

		Context("When the namespace does not exist", func() {
			It("Should not attempt to delete the rolebindings", func() {
				// nothing to check here, the absence of the namespace is expected
			})
		})

		Context("When the rolebindings do not exist", func() {
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

			It("Should not return an error", func() {
				// nothing to check here, the absence of the rolebindings is expected
			})
		})
	})
})
