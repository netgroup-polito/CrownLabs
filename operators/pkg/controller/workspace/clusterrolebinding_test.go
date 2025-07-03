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
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

var crbInstances = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "crownlabs-manage-instances-" + wsName,
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
		Name:     "crownlabs-manage-instances",
		APIGroup: "rbac.authorization.k8s.io",
	},
}

var crbTenants = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "crownlabs-manage-tenants-" + wsName,
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
		Name:     "crownlabs-manage-tenants",
		APIGroup: "rbac.authorization.k8s.io",
	},
}

var _ = Describe("ClusterRoleBinding", func() {
	Context("When a workspace is created", func() {
		It("Should create a ClusterRoleBinding to manage instances", func() {
			crb := &rbacv1.ClusterRoleBinding{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-instances-" + wsName}, crb, BeTrue(), timeout, interval)
			Expect(crb.Subjects).To(HaveLen(1))
			Expect(crb.Subjects[0].Kind).To(Equal("Group"))
			Expect(crb.Subjects[0].Name).To(Equal("kubernetes:workspace-" + wsName + ":manager"))
			Expect(crb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(crb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(crb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(crb.RoleRef.Name).To(Equal("crownlabs-manage-instances"))
		})

		It("Should create a ClusterRoleBinding to manage tenants", func() {
			crb := &rbacv1.ClusterRoleBinding{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "crownlabs-manage-tenants-" + wsName}, crb, BeTrue(), timeout, interval)
			Expect(crb.Subjects).To(HaveLen(1))
			Expect(crb.Subjects[0].Kind).To(Equal("Group"))
			Expect(crb.Subjects[0].Name).To(Equal("kubernetes:workspace-" + wsName + ":manager"))
			Expect(crb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(crb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(crb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(crb.RoleRef.Name).To(Equal("crownlabs-manage-tenants"))
		})

		Context("When there is an error creating the ClusterRoleBinding", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption) error {
						if crb, ok := obj.(*rbacv1.ClusterRoleBinding); ok && strings.Contains(crb.Name, "manage-instances") {
							return fmt.Errorf("error creating ClusterRoleBinding")
						}
						return nil
					},
				})

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

	Context("When the workspace is being deleted", func() {
		BeforeEach(func() {
			workspaceBeingDeleted()
		})

		Context("When the ClusterRoleBindings are present on the cluster", func() {
			BeforeEach(func() {
				addObjToObjectsList(crbInstances)
				addObjToObjectsList(crbTenants)
			})

			AfterEach(func() {
				removeObjFromObjectsList(crbInstances)
				removeObjFromObjectsList(crbTenants)
			})

			It("Should delete the ClusterRoleBinding to manage instances", func() {
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: crbInstances.Name}, &rbacv1.ClusterRoleBinding{}, BeFalse(), timeout, interval)
			})

			It("Should delete the ClusterRoleBinding to manage tenants", func() {
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: crbTenants.Name}, &rbacv1.ClusterRoleBinding{}, BeFalse(), timeout, interval)
			})

			Context("When there is an error deleting the ClusterRoleBinding", func() {
				BeforeEach(func() {
					builder.WithInterceptorFuncs(interceptor.Funcs{
						Delete: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.DeleteOption) error {
							if crb, ok := obj.(*rbacv1.ClusterRoleBinding); ok && strings.Contains(crb.Name, "manage-instances") {
								return fmt.Errorf("error deleting ClusterRoleBinding")
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
					Expect(ws.DeletionTimestamp).To(Not(BeNil()))
				})
			})
		})

		Context("When the ClusterRoleBindings are not present on the cluster", func() {
			It("Should not return an error", func() {
				wsReconcileErrExpected = Not(HaveOccurred())
			})

			It("Should not create the ClusterRoleBinding to manage instances", func() {
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: crbInstances.Name}, &rbacv1.ClusterRoleBinding{}, BeFalse(), timeout, interval)
			})

			It("Should not create the ClusterRoleBinding to manage tenants", func() {
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: crbTenants.Name}, &rbacv1.ClusterRoleBinding{}, BeFalse(), timeout, interval)
			})
		})
	})
})
