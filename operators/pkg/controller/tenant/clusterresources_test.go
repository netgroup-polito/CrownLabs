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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Cluster Resources", func() {
	Context("When tenant is created", func() {
		It("Should create ClusterRole for tenant access", func() {
			cr := &rbacv1.ClusterRole{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-tenant-" + tnName}, cr, BeTrue(), timeout, interval)

			Expect(cr.Rules).To(HaveLen(1))
			Expect(cr.Rules[0].APIGroups).To(ContainElement("crownlabs.polito.it"))
			Expect(cr.Rules[0].Resources).To(ContainElement("tenants"))
			Expect(cr.Rules[0].Verbs).To(ContainElement("get"))
			Expect(cr.Rules[0].Verbs).To(ContainElement("list"))
			Expect(cr.Rules[0].Verbs).To(ContainElement("watch"))
			Expect(cr.Rules[0].Verbs).To(ContainElement("patch"))
			Expect(cr.Rules[0].Verbs).To(ContainElement("update"))
			Expect(cr.Rules[0].ResourceNames).To(HaveLen(1))
			Expect(cr.Rules[0].ResourceNames).To(ContainElement(tnName))
		})

		It("Should create ClusterRoleBinding for tenant access", func() {
			crb := &rbacv1.ClusterRoleBinding{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-tenant-" + tnName}, crb, BeTrue(), timeout, interval)

			Expect(crb.RoleRef.Name).To(Equal("crownlabs-manage-tenant-" + tnName))
			Expect(crb.Subjects).To(HaveLen(1))
			Expect(crb.Subjects[0].Kind).To(Equal("User"))
			Expect(crb.Subjects[0].Name).To(Equal(tnName))
			Expect(crb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(crb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(crb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(crb.RoleRef.Name).To(Equal("crownlabs-manage-tenant-" + tnName))
		})

		Context("When there is an error creating ClusterRole", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption) error {
						if cr, ok := obj.(*rbacv1.ClusterRole); ok && cr.Name == "crownlabs-manage-tenant-"+tnName {
							return fmt.Errorf("error creating cluster role")
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

		Context("When there is an error creating ClusterRoleBinding", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption) error {
						if crb, ok := obj.(*rbacv1.ClusterRoleBinding); ok && crb.Name == "crownlabs-manage-tenant-"+tnName {
							return fmt.Errorf("error creating cluster role binding")
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

	Context("When tenant is being deleted", func() {
		cr := &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "crownlabs-manage-tenant-" + tnName,
			},
		}

		crb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "crownlabs-manage-tenant-" + tnName,
			},
		}

		BeforeEach(func() {
			addObjToObjectsList(cr)
			addObjToObjectsList(crb)
			tenantBeingDeleted()
		})

		AfterEach(func() {
			removeObjFromObjectsList(cr)
			removeObjFromObjectsList(crb)
		})

		It("Should delete ClusterRole for tenant access", func() {
			cr := &rbacv1.ClusterRole{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-tenant-" + tnName}, cr, BeFalse(), timeout, interval)
		})

		It("Should delete ClusterRoleBinding for tenant access", func() {
			crb := &rbacv1.ClusterRoleBinding{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-tenant-" + tnName}, crb, BeFalse(), timeout, interval)
		})

		Context("When there is an error deleting ClusterRole", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.DeleteOption) error {
						if cr, ok := obj.(*rbacv1.ClusterRole); ok && cr.Name == "crownlabs-manage-tenant-"+tnName {
							return fmt.Errorf("error deleting cluster role")
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

			It("Should prevent the deletion of the tenant", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

				Expect(tn.Finalizers).To(ContainElement(v1alpha2.TnOperatorFinalizerName))
			})
		})

		Context("When there is an error deleting ClusterRoleBinding", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.DeleteOption) error {
						if crb, ok := obj.(*rbacv1.ClusterRoleBinding); ok && crb.Name == "crownlabs-manage-tenant-"+tnName {
							return fmt.Errorf("error deleting cluster role binding")
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

			It("Should prevent the deletion of the tenant", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

				Expect(tn.Finalizers).To(ContainElement(v1alpha2.TnOperatorFinalizerName))
			})
		})
	})
})
