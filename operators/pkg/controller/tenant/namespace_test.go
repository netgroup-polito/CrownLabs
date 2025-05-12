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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var (
	testScheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(testScheme))
	utilruntime.Must(v1alpha2.AddToScheme(testScheme))
}

var _ = Describe("Namespace management", func() {
	Context("When tenant needs namespace resources", func() {
		It("Should create the personal namespace", func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)

			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/type", "tenant"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/name", tnName))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/instance-resources-replication", "true"))
		})

		It("Should create the resource quota", func() {
			rq := &v1.ResourceQuota{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-resource-quota",
				Namespace: "tenant-" + tnName,
			}, rq, BeTrue(), 10*time.Second, 250*time.Millisecond)

			Expect(rq.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(rq.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(rq.Spec.Hard).ToNot(BeEmpty())
		})

		It("Should create the role binding", func() {
			rb := &rbacv1.RoleBinding{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-manage-instances",
				Namespace: "tenant-" + tnName,
			}, rb, BeTrue(), 10*time.Second, 250*time.Millisecond)

			Expect(rb.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(rb.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(rb.RoleRef.Name).To(Equal("crownlabs-manage-instances"))
			Expect(rb.Subjects).To(HaveLen(1))
			Expect(rb.Subjects[0].Kind).To(Equal("User"))
			Expect(rb.Subjects[0].Name).To(Equal(tnName))
			Expect(rb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
		})

		It("Should create the deny network policy", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-deny-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeTrue(), 10*time.Second, 250*time.Millisecond)

			Expect(netPol.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(netPol.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(netPol.Spec.PodSelector.MatchLabels).To(HaveLen(0))
			Expect(netPol.Spec.Ingress).To(HaveLen(1))
			Expect(netPol.Spec.Ingress[0].From).To(HaveLen(1))
			Expect(netPol.Spec.Ingress[0].From[0].PodSelector.MatchLabels).To(HaveLen(0))
		})

		It("Should create the allow network policy", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-allow-trusted-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeTrue(), 10*time.Second, 250*time.Millisecond)

			Expect(netPol.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(netPol.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(netPol.Spec.PodSelector.MatchLabels).To(HaveLen(0))
			Expect(netPol.Spec.Ingress).To(HaveLen(1))
			Expect(netPol.Spec.Ingress[0].From).To(HaveLen(1))
			Expect(netPol.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels).To(HaveKeyWithValue("crownlabs.polito.it/allow-instance-access", "true"))
		})
	})

	Context("When tenant namespace should be deleted", func() {
		JustBeforeEach(func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)

			updatedTenant := &v1alpha2.Tenant{}
			err := cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = cl.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			tnResource = updatedTenant

			_, err = tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tnResource.Name},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should delete the namespace resources", func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should delete the resource quota", func() {
			rq := &v1.ResourceQuota{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-resource-quota",
				Namespace: "tenant-" + tnName,
			}, rq, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should delete the role binding", func() {
			rb := &rbacv1.RoleBinding{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-manage-instances",
				Namespace: "tenant-" + tnName,
			}, rb, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should delete the deny network policy", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-deny-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should delete the allow network policy", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-allow-trusted-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})
	})

	Context("When tenant has instances running", func() {
		JustBeforeEach(func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)

			instance := &v1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-instance",
					Namespace: "tenant-" + tnName,
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Spec: v1alpha2.InstanceSpec{
					Running: true,
				},
			}

			err := cl.Create(ctx, instance)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant := &v1alpha2.Tenant{}
			err = cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = cl.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			tnResource = updatedTenant

			_, err = tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tnResource.Name},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should keep the namespace when instances are running", func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should keep the resource quota when instances are running", func() {
			rq := &v1.ResourceQuota{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-resource-quota",
				Namespace: "tenant-" + tnName,
			}, rq, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should keep the role binding when instances are running", func() {
			rb := &rbacv1.RoleBinding{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-manage-instances",
				Namespace: "tenant-" + tnName,
			}, rb, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should keep the deny network policy when instances are running", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-deny-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should keep the allow network policy when instances are running", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-allow-trusted-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})
	})

	Context("When testing edge cases and error scenarios", func() {
		It("Should handle tenant with zero LastLogin time correctly", func() {

			updatedTenant := &v1alpha2.Tenant{}
			err := cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.Time{}
			err = cl.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			_, err = tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: updatedTenant.Name},
			})
			Expect(err).ToNot(HaveOccurred())

			namespace := &v1.Namespace{}
			Consistently(func() error {
				return cl.Get(ctx, client.ObjectKey{Name: "tenant-" + tnName}, namespace)
			}, 2*time.Second, 250*time.Millisecond).Should(HaveOccurred())
		})

		It("Should handle namespace with dots in tenant name", func() {
			tenantWithDots := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test.tenant.with.dots",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Spec: v1alpha2.TenantSpec{
					FirstName: "Test",
					LastName:  "TenantDots",
					Email:     "test.dots@tenant.example",
					LastLogin: metav1.Now(),
				},
			}

			err := cl.Create(ctx, tenantWithDots)
			Expect(err).ToNot(HaveOccurred())

			_, err = tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tenantWithDots.Name},
			})
			Expect(err).ToNot(HaveOccurred())

			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-test-tenant-with-dots"},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)

			err = cl.Delete(ctx, tenantWithDots)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should handle error when listing instances fails", func() {
			interceptorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(tnResource).
				WithInterceptorFuncs(interceptor.Funcs{
					List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
						if _, ok := list.(*v1alpha2.InstanceList); ok {
							return fmt.Errorf("simulated list error")
						}
						return client.List(ctx, list, opts...)
					},
				}).
				Build()

			updatedTenant := &v1alpha2.Tenant{}
			err := interceptorClient.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = interceptorClient.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = interceptorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: updatedTenant.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle error in deletePersonalNamespace", func() {
			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(tnResource).
				WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
						if _, ok := obj.(*v1.Namespace); ok {
							return fmt.Errorf("simulated delete error for Namespace")
						}
						return client.Delete(ctx, obj, opts...)
					},
				}).
				Build()

			updatedTenant := &v1alpha2.Tenant{}
			err := errorClient.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = errorClient.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: updatedTenant.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle multiple instances in namespace", func() {
			updatedTenant := &v1alpha2.Tenant{}
			err := cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i < 3; i++ {
				instance := &v1alpha2.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("test-instance-%d", i),
						Namespace: "tenant-" + tnName,
						Labels: map[string]string{
							"crownlabs.polito.it/operator-selector": "test",
						},
					},
					Spec: v1alpha2.InstanceSpec{Running: true},
				}
				err := cl.Create(ctx, instance)
				Expect(err).ToNot(HaveOccurred())
			}

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = cl.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			_, err = tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: updatedTenant.Name},
			})
			Expect(err).ToNot(HaveOccurred())

			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should test recent login time", func() {
			updatedTenant := &v1alpha2.Tenant{}
			err := cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-1 * time.Hour))
			err = cl.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			_, err = tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: updatedTenant.Name},
			})
			Expect(err).ToNot(HaveOccurred())

			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should handle error scenarios in create functions", func() {
			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects().
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						if _, ok := obj.(*v1.ResourceQuota); ok {
							return fmt.Errorf("simulated create error for ResourceQuota")
						}
						return client.Create(ctx, obj, opts...)
					},
					Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						if _, ok := obj.(*v1.ResourceQuota); ok {
							return fmt.Errorf("simulated update error for ResourceQuota")
						}
						return client.Update(ctx, obj, opts...)
					},
				}).
				Build()

			tenantForError := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "error-tenant",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Spec: v1alpha2.TenantSpec{
					FirstName: "Error",
					LastName:  "Tenant",
					Email:     "error@tenant.example",
					LastLogin: metav1.Now(),
				},
			}

			err := errorClient.Create(ctx, tenantForError)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tenantForError.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle error scenarios in delete functions", func() {
			_, err := tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tnResource.Name},
			})
			Expect(err).ToNot(HaveOccurred())

			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(tnResource).
				WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
						if _, ok := obj.(*rbacv1.RoleBinding); ok {
							return fmt.Errorf("simulated delete error for RoleBinding")
						}
						return client.Delete(ctx, obj, opts...)
					},
				}).
				Build()

			updatedTenant := &v1alpha2.Tenant{}
			err = errorClient.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = errorClient.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: updatedTenant.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle network policy creation errors", func() {
			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects().
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						if _, ok := obj.(*netv1.NetworkPolicy); ok {
							return fmt.Errorf("simulated create error for NetworkPolicy")
						}
						return client.Create(ctx, obj, opts...)
					},
					Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						if _, ok := obj.(*netv1.NetworkPolicy); ok {
							return fmt.Errorf("simulated update error for NetworkPolicy")
						}
						return client.Update(ctx, obj, opts...)
					},
				}).
				Build()

			tenantForNetpolError := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "netpol-error-tenant",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Spec: v1alpha2.TenantSpec{
					FirstName: "NetPol",
					LastName:  "Error",
					Email:     "netpol@error.example",
					LastLogin: metav1.Now(),
				},
			}

			err := errorClient.Create(ctx, tenantForNetpolError)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tenantForNetpolError.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle error in createPersonalNamespace", func() {
			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects().
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						if _, ok := obj.(*v1.Namespace); ok {
							return fmt.Errorf("simulated create error for Namespace")
						}
						return client.Create(ctx, obj, opts...)
					},
					Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						if _, ok := obj.(*v1.Namespace); ok {
							return fmt.Errorf("simulated update error for Namespace")
						}
						return client.Update(ctx, obj, opts...)
					},
				}).
				Build()

			tenantForNsError := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-error-tenant",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Spec: v1alpha2.TenantSpec{
					FirstName: "Namespace",
					LastName:  "Error",
					Email:     "ns@error.example",
					LastLogin: metav1.Now(),
				},
			}

			err := errorClient.Create(ctx, tenantForNsError)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tenantForNsError.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle error in createInstanceRoleBinding", func() {
			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects().
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						if _, ok := obj.(*rbacv1.RoleBinding); ok {
							return fmt.Errorf("simulated create error for RoleBinding")
						}
						return client.Create(ctx, obj, opts...)
					},
					Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						if _, ok := obj.(*rbacv1.RoleBinding); ok {
							return fmt.Errorf("simulated update error for RoleBinding")
						}
						return client.Update(ctx, obj, opts...)
					},
				}).
				Build()

			tenantForRbError := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rb-error-tenant",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Spec: v1alpha2.TenantSpec{
					FirstName: "RoleBinding",
					LastName:  "Error",
					Email:     "rb@error.example",
					LastLogin: metav1.Now(),
				},
			}

			err := errorClient.Create(ctx, tenantForRbError)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tenantForRbError.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle error in createDenyNetworkPolicy", func() {
			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects().
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						if netPol, ok := obj.(*netv1.NetworkPolicy); ok && netPol.Name == "crownlabs-deny-ingress-traffic" {
							return fmt.Errorf("simulated create error for Deny NetworkPolicy")
						}
						return client.Create(ctx, obj, opts...)
					},
					Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						if netPol, ok := obj.(*netv1.NetworkPolicy); ok && netPol.Name == "crownlabs-deny-ingress-traffic" {
							return fmt.Errorf("simulated update error for Deny NetworkPolicy")
						}
						return client.Update(ctx, obj, opts...)
					},
				}).
				Build()

			tenantForDenyError := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deny-error-tenant",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Spec: v1alpha2.TenantSpec{
					FirstName: "Deny",
					LastName:  "Error",
					Email:     "deny@error.example",
					LastLogin: metav1.Now(),
				},
			}

			err := errorClient.Create(ctx, tenantForDenyError)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tenantForDenyError.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle error in createAllowNetworkPolicy", func() {
			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects().
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						if netPol, ok := obj.(*netv1.NetworkPolicy); ok && netPol.Name == "crownlabs-allow-trusted-ingress-traffic" {
							return fmt.Errorf("simulated create error for Allow NetworkPolicy")
						}
						return client.Create(ctx, obj, opts...)
					},
					Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						if netPol, ok := obj.(*netv1.NetworkPolicy); ok && netPol.Name == "crownlabs-allow-trusted-ingress-traffic" {
							return fmt.Errorf("simulated update error for Allow NetworkPolicy")
						}
						return client.Update(ctx, obj, opts...)
					},
				}).
				Build()

			tenantForAllowError := &v1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: "allow-error-tenant",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Spec: v1alpha2.TenantSpec{
					FirstName: "Allow",
					LastName:  "Error",
					Email:     "allow@error.example",
					LastLogin: metav1.Now(),
				},
			}

			err := errorClient.Create(ctx, tenantForAllowError)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tenantForAllowError.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle error in deleteDenyNetworkPolicy", func() {

			_, err := tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tnResource.Name},
			})
			Expect(err).ToNot(HaveOccurred())

			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(tnResource).
				WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
						if netPol, ok := obj.(*netv1.NetworkPolicy); ok && netPol.Name == "crownlabs-deny-ingress-traffic" {
							return fmt.Errorf("simulated delete error for Deny NetworkPolicy")
						}
						return client.Delete(ctx, obj, opts...)
					},
				}).
				Build()

			updatedTenant := &v1alpha2.Tenant{}
			err = errorClient.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = errorClient.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: updatedTenant.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle error in deleteAllowNetworkPolicy", func() {

			_, err := tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tnResource.Name},
			})
			Expect(err).ToNot(HaveOccurred())

			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(tnResource).
				WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
						if netPol, ok := obj.(*netv1.NetworkPolicy); ok && netPol.Name == "crownlabs-allow-trusted-ingress-traffic" {
							return fmt.Errorf("simulated delete error for Allow NetworkPolicy")
						}
						return client.Delete(ctx, obj, opts...)
					},
				}).
				Build()

			updatedTenant := &v1alpha2.Tenant{}
			err = errorClient.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = errorClient.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: updatedTenant.Name},
			})
			Expect(err).To(HaveOccurred())
		})

		It("Should handle error in deleteResourceQuota", func() {

			_, err := tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tnResource.Name},
			})
			Expect(err).ToNot(HaveOccurred())

			errorClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(tnResource).
				WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
						if _, ok := obj.(*v1.ResourceQuota); ok {
							return fmt.Errorf("simulated delete error for ResourceQuota")
						}
						return client.Delete(ctx, obj, opts...)
					},
				}).
				Build()

			updatedTenant := &v1alpha2.Tenant{}
			err = errorClient.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = errorClient.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			tempReconciler := tenantReconciler
			tempReconciler.Client = errorClient

			_, err = tempReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: updatedTenant.Name},
			})
			Expect(err).To(HaveOccurred())
		})
	})
})
