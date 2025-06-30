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

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-logr/logr"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/mock"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TODO --> questo file di test va riorganizzato con gli altri test del tenant, con un'organizzazione simile a quella adottata per il workspace

var _ = BeforeSuite(func() {
	Expect(crownlabsv1alpha2.AddToScheme(scheme.Scheme)).To(Succeed())
})

var _ = Describe("Authentication", func() {
	var (
		ctx              context.Context
		builder          fake.ClientBuilder
		cl               client.Client
		mockCtrl         *gomock.Controller
		keycloakActor    *mock.MockKeycloakActorIface
		tenantReconciler tenant.Reconciler

		tnResource             *crownlabsv1alpha2.Tenant
		tnReconcileErrExpected gomegaTypes.GomegaMatcher
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		builder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

		mockCtrl = gomock.NewController(GinkgoT())
		keycloakActor = mock.NewMockKeycloakActorIface(mockCtrl)

		tnResource = &crownlabsv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-tenant",
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "keycloak-test",
				},
			},
			Spec: crownlabsv1alpha2.TenantSpec{
				FirstName: "Test",
				LastName:  "Tenant",
				Email:     "test@tenant.example",
				LastLogin: metav1.Now(),
			},
		}
		tnReconcileErrExpected = Not(HaveOccurred())

		keycloakActor.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	JustBeforeEach(func() {
		cl = builder.WithObjects(tnResource).WithStatusSubresource(tnResource).Build()

		tenantReconciler = tenant.Reconciler{
			Client:            cl,
			Scheme:            scheme.Scheme,
			KeycloakActor:     keycloakActor,
			TargetLabel:       common.NewLabel("crownlabs.polito.it/operator-selector", "keycloak-test"),
			TenantNSKeepAlive: 24 * time.Hour,
		}

		_, err := tenantReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: tnResource.Name,
			},
		})
		Expect(err).To(tnReconcileErrExpected)
	})

	Context("When Keycloak is initialized", func() {
		BeforeEach(func() {
			keycloakActor.EXPECT().IsInitialized().Return(true).AnyTimes()
		})

		Context("When the Tenant is not created in Keycloak", func() {
			BeforeEach(func() {
				gomock.InOrder(
					keycloakActor.EXPECT().GetUser(gomock.Any(), "test-tenant").Return(nil, fmt.Errorf("404")),
					keycloakActor.EXPECT().CreateUser(
						gomock.Any(),
						tnResource.Name,
						tnResource.Spec.Email,
						tnResource.Spec.FirstName,
						tnResource.Spec.LastName,
					).Return("test-user-id", nil),
					keycloakActor.EXPECT().GetUser(gomock.Any(), "test-tenant").Return(&gocloak.User{
						ID:            gocloak.StringP("test-user-id"),
						EmailVerified: gocloak.BoolP(false),
					}, nil),
				)
			})

			It("Should set Tenant keycloak status as created but not confirmed", func() {
				tn := &crownlabsv1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "test-tenant"}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Keycloak.UserCreated).To(Equal(crownlabsv1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}))
				Expect(tn.Status.Keycloak.UserConfirmed).To(BeFalse())
			})

			It("Should not create the related resources", func() {
				namespace := &corev1.Namespace{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-test-tenant"}, namespace, BeFalse(), 10*time.Second, 250*time.Millisecond)
			})
		})

		Context("When the Tenant is created but not confirmed in Keycloak", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), "test-tenant").Return(&gocloak.User{
					ID:            gocloak.StringP("test-user-id"),
					EmailVerified: gocloak.BoolP(false),
				}, nil)
			})

			It("Should set Tenant keycloak status as created but not confirmed", func() {
				tn := &crownlabsv1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "test-tenant"}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Keycloak.UserCreated).To(Equal(crownlabsv1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}))
				Expect(tn.Status.Keycloak.UserConfirmed).To(BeFalse())
			})

			It("Should not create the related resources", func() {
				namespace := &corev1.Namespace{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-test-tenant"}, namespace, BeFalse(), 10*time.Second, 250*time.Millisecond)
			})
		})

		Context("When the Tenant is created and confirmed in Keycloak", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), "test-tenant").Return(&gocloak.User{
					ID:            gocloak.StringP("test-user-id"),
					EmailVerified: gocloak.BoolP(true),
				}, nil)
			})

			It("Should set Tenant keycloak status as confirmed", func() {
				tn := &crownlabsv1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "test-tenant"}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Keycloak.UserCreated).To(Equal(crownlabsv1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}))
				Expect(tn.Status.Keycloak.UserConfirmed).To(BeTrue())
			})

			It("Should create the related resources", func() {
				namespace := &corev1.Namespace{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-test-tenant"}, namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)
			})
		})

		Context("When the Tenant is deleted", func() {
			BeforeEach(func() {
				tnResource.Finalizers = append(tnResource.Finalizers, crownlabsv1alpha2.TnOperatorFinalizerName)
				tnResource.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: time.Now().Add(10 * time.Second)}
				tnResource.Status.Keycloak.UserCreated = crownlabsv1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}
				tnResource.Status.Keycloak.UserConfirmed = true

				keycloakActor.EXPECT().DeleteUser(gomock.Any(), "test-user-id").Return(nil)
			})

			It("Should delete the Tenant keycloak user", func() {
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "test-tenant"}, &crownlabsv1alpha2.Tenant{}, BeFalse(), 10*time.Second, 250*time.Millisecond)
			})
		})

		Context("When there is an error on Keycloak operations", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), "test-tenant").Return(nil, fmt.Errorf("some error"))
				tnReconcileErrExpected = HaveOccurred()
			})

			It("Should make the Tenant keycloak subscription failed", func() {
				tn := &crownlabsv1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "test-tenant"}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Ready).To(BeFalse())
				Expect(tn.Status.Subscriptions).To(HaveKeyWithValue("keycloak", crownlabsv1alpha2.SubscrFailed))
			})
		})

		Context("When the user-id in Keycloak does not match the one in Tenant status", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), "test-tenant").Return(&gocloak.User{
					ID:            gocloak.StringP("different-user-id"),
					EmailVerified: gocloak.BoolP(true),
				}, nil)
				tnResource.Status.Keycloak.UserCreated = crownlabsv1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}
			})

			It("Should update the Tenant keycloak status with the new user-id", func() {
				tn := &crownlabsv1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "test-tenant"}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Keycloak.UserCreated).To(Equal(crownlabsv1alpha2.NameCreated{
					Name:    "different-user-id",
					Created: true,
				}))
			})
		})
	})

	Context("When Keycloak is not initialized", func() {
		BeforeEach(func() {
			keycloakActor.EXPECT().IsInitialized().Return(false).AnyTimes()
		})

		It("Should set Tenant keycloak status as not confirmed", func() {
			tn := &crownlabsv1alpha2.Tenant{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "test-tenant"}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
			Expect(tn.Status.Keycloak.UserCreated).To(Equal(crownlabsv1alpha2.NameCreated{
				Name:    "",
				Created: false,
			}))
			Expect(tn.Status.Keycloak.UserConfirmed).To(BeFalse())
		})

		It("Should create the related resources", func() {
			namespace := &corev1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-test-tenant"}, namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})
	})
})

func DoesEventuallyExists(ctx context.Context, cl client.Client, objLookupKey client.ObjectKey, targetObj client.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout, interval time.Duration) {
	Eventually(func() bool {
		err := cl.Get(ctx, objLookupKey, targetObj)
		return err == nil
	}, timeout, interval).Should(expectedStatus)
}
