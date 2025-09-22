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

package tenant_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Sandbox", func() {
	Context("When the tenant sandbox is required", func() {
		BeforeEach(func() {
			tnResource.Spec.CreateSandbox = true
		})

		It("Should create a sandbox namespace for the tenant", func() {
			sbNs := v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "sandbox-" + tnName}, &sbNs, BeTrue(), timeout, interval)
		})

		It("Should create a role binding to edit the sandbox", func() {
			rb := &rbacv1.RoleBinding{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "sandbox-editor", Namespace: "sandbox-" + tnName}, rb, BeTrue(), timeout, interval)
			Expect(rb.RoleRef.Name).To(Equal("test-sandbox-editor"))
			Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(rb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(rb.Subjects).To(HaveLen(1))
			Expect(rb.Subjects[0].Kind).To(Equal("User"))
			Expect(rb.Subjects[0].Name).To(Equal(tnName))
			Expect(rb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
		})

		It("Should enforce resource quotas", func() {
			quota := &v1.ResourceQuota{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "sandbox-resource-quota", Namespace: "sandbox-" + tnName}, quota, BeTrue(), timeout, interval)
		})

		It("Should enforce limit ranges", func() {
			limitRange := &v1.LimitRange{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "sandbox-limit-range", Namespace: "sandbox-" + tnName}, limitRange, BeTrue(), timeout, interval)
		})

		Context("When there is an error creating the sandbox namespace", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption) error {
						if ns, ok := obj.(*v1.Namespace); ok && ns.Name == "sandbox-"+tnName {
							return fmt.Errorf("error creating sandbox namespace")
						}
						return nil
					},
				})

				tnReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("Should set the tenant status to not ready", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

				Expect(tn.Status.Ready).To(BeFalse())
			})
		})

		Context("When there is an error creating the role binding", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption) error {
						if rb, ok := obj.(*rbacv1.RoleBinding); ok && rb.Name == "sandbox-editor" && rb.Namespace == "sandbox-"+tnName {
							return fmt.Errorf("error creating role binding")
						}
						return nil
					},
				})

				tnReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("Should set the tenant status to not ready", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

				Expect(tn.Status.Ready).To(BeFalse())
			})
		})

		Context("When there is an error creating the resource quota", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption) error {
						if quota, ok := obj.(*v1.ResourceQuota); ok && quota.Name == "sandbox-resource-quota" && quota.Namespace == "sandbox-"+tnName {
							return fmt.Errorf("error creating resource quota")
						}
						return nil
					},
				})

				tnReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("Should set the tenant status to not ready", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

				Expect(tn.Status.Ready).To(BeFalse())
			})
		})

		Context("When there is an error creating the limit range", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption) error {
						if lr, ok := obj.(*v1.LimitRange); ok && lr.Name == "sandbox-limit-range" && lr.Namespace == "sandbox-"+tnName {
							return fmt.Errorf("error creating limit range")
						}
						return nil
					},
				})

				tnReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("Should set the tenant status to not ready", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

				Expect(tn.Status.Ready).To(BeFalse())
			})
		})

		Context("When the tenant is being deleted", func() {
			sbNs := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sandbox-" + tnName,
				},
			}

			BeforeEach(func() {
				addObjToObjectsList(sbNs)
				tenantBeingDeleted()
				tnResource.Status.SandboxNamespace.Created = true
				tnResource.Status.SandboxNamespace.Name = sbNs.Name
			})

			AfterEach(func() {
				removeObjFromObjectsList(sbNs)
			})

			It("Should delete the sandbox namespace", func() {
				ns := &v1.Namespace{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "sandbox-" + tnName}, ns, BeFalse(), timeout, interval)
			})
		})
	})

	Context("When the tenant sandbox is not required", func() {
		It("Should not create a sandbox for the tenant", func() {
			sbNs := v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "sandbox-" + tnName}, &sbNs, BeFalse(), timeout, interval)
		})

		Context("When the sandbox namespace exists", func() {
			sbNs := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sandbox-" + tnName,
				},
			}

			BeforeEach(func() {
				addObjToObjectsList(sbNs)
				tnResource.Status.SandboxNamespace.Created = true
				tnResource.Status.SandboxNamespace.Name = sbNs.Name
			})

			AfterEach(func() {
				removeObjFromObjectsList(sbNs)
			})

			It("Should delete the sandbox namespace", func() {
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "sandbox-" + tnName}, sbNs, BeFalse(), timeout, interval)
			})
		})

		Context("When there is an error deleting the sandbox namespace", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.DeleteOption) error {
						if ns, ok := obj.(*v1.Namespace); ok && ns.Name == "sandbox-"+tnName {
							return fmt.Errorf("error deleting sandbox namespace")
						}
						return nil
					},
				})

				tnReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("Should set the tenant status to not ready", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

				Expect(tn.Status.Ready).To(BeFalse())
			})
		})
	})
})
