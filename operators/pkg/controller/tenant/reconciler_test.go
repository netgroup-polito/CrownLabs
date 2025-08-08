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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("TenantReconciler", func() {
	Context("When there is an error getting the resource from the cluster", func() {
		BeforeEach(func() {
			builder.WithInterceptorFuncs(interceptor.Funcs{
				Get: func(_ context.Context, _ client.WithWatch, key client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
					if key.Name == tnName {
						return fmt.Errorf("error getting resource %s", key.Name)
					}
					return nil
				},
			})

			tnReconcileErrExpected = HaveOccurred()
		})

		It("Should return an error", func() {
			// already set in BeforeEach
		})
	})

	Context("When the resource is not found in the cluster", func() {
		BeforeEach(func() {
			tnResource = &v1alpha2.Tenant{}
		})

		It("Should not return an error", func() {
			// because the resource is just been deleted

			// already set in BeforeEach
		})
	})

	Context("When the resource is not responsibility of the current reconciler", func() {
		BeforeEach(func() {
			tnResource.Labels["crownlabs.polito.it/operator-selector"] = "other"
		})

		It("Should not return an error", func() {
			// nothing to do, just skip the resource
		})

		It("Should not create any resources", func() {
			ns := &v1.Namespace{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName}, ns, BeFalse(), timeout, interval)
		})

		Context("When there are resources associated with the tenant", func() {
			namespaceResource := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tenant-" + tnName,
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "other",
					},
				},
			}

			BeforeEach(func() {
				addObjToObjectsList(namespaceResource)
			})

			AfterEach(func() {
				removeObjFromObjectsList(namespaceResource)
			})

			It("Should not delete the resources", func() {
				ns := &v1.Namespace{}

				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName}, ns, BeTrue(), timeout, interval)

				Expect(ns.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "other"))
			})
		})
	})

	It("Should add the finalizer to the tenant resource", func() {
		tn := &v1alpha2.Tenant{}

		DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

		Expect(tn.Finalizers).To(ContainElement("crownlabs.polito.it/tenant-operator"))
	})

	Context("When there is an error adding the finalizer", func() {
		BeforeEach(func() {
			builder.WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(_ context.Context, _ client.WithWatch, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
					if tn, ok := obj.(*v1alpha2.Tenant); ok && tn.Name == tnName {
						return fmt.Errorf("failed to add finalizer")
					}
					return nil
				},
			})

			tnReconcileErrExpected = HaveOccurred()
		})

		It("Should return an error", func() {
			// checked in BeforeEach
		})
	})

	Context("When the workspace is being deleted", func() {
		BeforeEach(func() {
			tenantBeingDeleted()
		})

		Context("When there is an error removing the finalizer", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					Patch: func(_ context.Context, _ client.WithWatch, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
						if tn, ok := obj.(*v1alpha2.Tenant); ok && tn.Name == tnName {
							return fmt.Errorf("failed to remove finalizer")
						}
						return nil
					},
				})

				tnReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in BeforeEach
			})

			It("Should prevent the deletion of the tenant resource", func() {
				tn := &v1alpha2.Tenant{}

				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)
			})
		})
	})

	Context("Testing WorkspaceToEnrolledTenants functionality", func() {
		var (
			workspaceName      string
			workspace          *v1alpha1.Workspace
			enrolledTenants    []*v1alpha2.Tenant
			notEnrolledTenants []*v1alpha2.Tenant
		)

		workspaceName = "test-workspace"

		workspace = &v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name: workspaceName,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test",
				},
			},
		}

		// Create enrolled tenants
		enrolledTenants = []*v1alpha2.Tenant{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "enrolled-tenant-1",
					Labels: map[string]string{
						fmt.Sprintf("%s%s", v1alpha2.WorkspaceLabelPrefix, workspaceName): "true",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "enrolled-tenant-2",
					Labels: map[string]string{
						fmt.Sprintf("%s%s", v1alpha2.WorkspaceLabelPrefix, workspaceName): "true",
					},
				},
			},
		}

		// Create not enrolled tenants
		notEnrolledTenants = []*v1alpha2.Tenant{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "not-enrolled-tenant-1",
					Labels: map[string]string{},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-enrolled-tenant-2",
					Labels: map[string]string{
						fmt.Sprintf("%s%s", v1alpha2.WorkspaceLabelPrefix, "other-workspace"): "true",
					},
				},
			},
		}

		BeforeEach(func() {
			// Add all resources to the fake client
			addObjToObjectsList(workspace)
			for _, t := range enrolledTenants {
				addObjToObjectsList(t)
			}
			for _, t := range notEnrolledTenants {
				addObjToObjectsList(t)
			}
		})

		AfterEach(func() {
			// Remove all resources from the fake client
			removeObjFromObjectsList(workspace)
			for _, t := range enrolledTenants {
				removeObjFromObjectsList(t)
			}
			for _, t := range notEnrolledTenants {
				removeObjFromObjectsList(t)
			}
		})

		It("Should return only tenants enrolled in the workspace", func() {
			// Execute the function being tested
			requests := tenantReconciler.WorkspaceNameToEnrolledTenants(ctx, workspaceName)

			// Verify the results
			Expect(requests).To(HaveLen(len(enrolledTenants)))

			// Create a map of expected tenant names
			expectedTenants := make(map[string]bool)
			for _, t := range enrolledTenants {
				expectedTenants[t.Name] = true
			}

			// Verify that all returned requests are for enrolled tenants
			for _, req := range requests {
				Expect(expectedTenants).To(HaveKey(req.Name))
			}
		})

		Context("When there is an error listing tenants", func() {
			BeforeEach(func() {
				builder.WithInterceptorFuncs(interceptor.Funcs{
					List: func(_ context.Context, _ client.WithWatch, list client.ObjectList, _ ...client.ListOption) error {
						if _, ok := list.(*v1alpha2.TenantList); ok {
							return fmt.Errorf("error listing tenants")
						}
						return nil
					},
				})
			})

			It("Should return nil", func() {
				requests := tenantReconciler.WorkspaceNameToEnrolledTenants(ctx, workspaceName)
				Expect(requests).To(BeNil())
			})
		})
	})
})
