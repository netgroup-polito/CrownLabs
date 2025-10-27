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
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Personal workspace handling", func() {
	When("Tenant namespace exists", func() {
		JustBeforeEach(func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnPersonalNamespace}, namespace, BeTrue(), timeout, interval)
		})
		When("Personal workspace is enabled", func() {
			BeforeEach(func() {
				tnResource.Spec.CreatePersonalWorkspace = true
			})
			It("Should create the manage templates role binding for the tenant", func() {
				rb := &rbacv1.RoleBinding{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-templates", Namespace: tnPersonalNamespace}, rb, BeTrue(), timeout, interval)

				Expect(rb.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
				Expect(rb.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", tenantReconciler.TargetLabel.GetValue()))
				Expect(rb.RoleRef.Name).To(Equal("crownlabs-manage-templates"))
				Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
				Expect(len(rb.Subjects)).To(Equal(1))
				Expect(rb.Subjects[0].Kind).To(Equal("User"))
				Expect(rb.Subjects[0].Name).To(Equal(tnName))
				updatedTenant := &v1alpha2.Tenant{}
				err := cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
				Expect(err).ToNot(HaveOccurred())
				Expect(updatedTenant.Status.PersonalWorkspaceCreated).To(BeTrue())
				Expect(updatedTenant.Status.FailingWorkspaces).To(BeEmpty())
			})
			When("RoleBinding creation or update errors", func() {
				BeforeEach(func() {
					builder = *builder.WithInterceptorFuncs(interceptor.Funcs{
						Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
							if rb, ok := obj.(*rbacv1.RoleBinding); ok && rb.Name == "crownlabs-manage-templates" && rb.Namespace == tnPersonalNamespace {
								return errors.New("error creating role binding")
							}
							return client.Create(ctx, obj, opts...)
						},
					})
					tnReconcileErrExpected = HaveOccurred()
				})
				It("Should mark the personal workspace as failing", func() {
					updatedTenant := &v1alpha2.Tenant{}
					Expect(cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)).To(Succeed())
					Expect(updatedTenant.Status.PersonalWorkspaceCreated).To(BeFalse())
					Expect(updatedTenant.Status.FailingWorkspaces).To(HaveLen(1))
					Expect(updatedTenant.Status.FailingWorkspaces[0]).To(Equal("personal-workspace"))
				})
			})
		})
		When("Personal workspace is disabled", func() {
			It("Should not have the manage templates role binding", func() {
				rb := &rbacv1.RoleBinding{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-templates", Namespace: tnPersonalNamespace}, rb, BeFalse(), timeout, interval)
				updatedTenant := &v1alpha2.Tenant{}
				err := cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
				Expect(err).ToNot(HaveOccurred())
				Expect(updatedTenant.Status.PersonalWorkspaceCreated).To(BeFalse())
				Expect(updatedTenant.Status.FailingWorkspaces).To(BeEmpty())
			})
			When("RoleBinding deletion errors", func() {
				BeforeEach(func() {
					builder = *builder.WithInterceptorFuncs(interceptor.Funcs{
						Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
							if rb, ok := obj.(*rbacv1.RoleBinding); ok && rb.Name == "crownlabs-manage-templates" && rb.Namespace == tnPersonalNamespace {
								return errors.New("error deleting role binding")
							}
							return client.Delete(ctx, obj, opts...)
						},
					})
					tnReconcileErrExpected = HaveOccurred()
				})
				It("Should mark the personal workspace as not created", func() {
					updatedTenant := &v1alpha2.Tenant{}
					Expect(cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)).To(Succeed())
					Expect(updatedTenant.Status.PersonalWorkspaceCreated).To(BeFalse())
					Expect(updatedTenant.Status.FailingWorkspaces).To(BeEmpty())
				})
			})
		})
	})
	When("Tenant namespace does not exist", func() {
		BeforeEach(func() {
			tnResource.Spec.LastLogin = metav1.NewTime(time.Now().Add(-(tenantReconciler.TenantNSKeepAlive + time.Second)))
		})
		JustBeforeEach(func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnPersonalNamespace}, namespace, BeFalse(), timeout, interval)
		})
		When("Personal workspace is enabled", func() {
			It("Should not have the manage templates role binding", func() {
				rb := &rbacv1.RoleBinding{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-templates", Namespace: tnPersonalNamespace}, rb, BeFalse(), timeout, interval)
				updatedTenant := &v1alpha2.Tenant{}
				err := cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
				Expect(err).ToNot(HaveOccurred())
				Expect(updatedTenant.Status.PersonalWorkspaceCreated).To(BeFalse())
				Expect(updatedTenant.Status.FailingWorkspaces).To(BeEmpty())
			})
		})
		When("Personal workspace is disabled", func() {
			It("Should not have the manage templates role binding", func() {
				rb := &rbacv1.RoleBinding{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-templates", Namespace: tnPersonalNamespace}, rb, BeFalse(), timeout, interval)
				updatedTenant := &v1alpha2.Tenant{}
				err := cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
				Expect(err).ToNot(HaveOccurred())
				Expect(updatedTenant.Status.PersonalWorkspaceCreated).To(BeFalse())
				Expect(updatedTenant.Status.FailingWorkspaces).To(BeEmpty())
			})
		})
	})
})
