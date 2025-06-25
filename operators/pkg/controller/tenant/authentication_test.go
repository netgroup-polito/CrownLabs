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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-logr/logr"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/mock"
	tntctrl "github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("Authenticator", func() {
	var (
		mockCtrl   *gomock.Controller
		ctx        context.Context
		reconciler *tntctrl.TenantReconciler
		tenant     *crownlabsv1alpha2.Tenant
		mKcAct     *mock.MockKeycloakActorIface
	)

	const (
		timeout    = time.Second * 10
		interval   = time.Millisecond * 250
		tenantName = "testuser"
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mKcAct = mock.NewMockKeycloakActorIface(mockCtrl)

		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())

		tenant = &crownlabsv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name: tenantName,
			},
			Spec: crownlabsv1alpha2.TenantSpec{
				FirstName: "Test",
				LastName:  "User",
				Email:     "test.user@example.com",
			},
		}

		cl := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
		reconciler = &tntctrl.TenantReconciler{
			Client:        cl,
			Scheme:        scheme.Scheme,
			KeycloakActor: mKcAct,
		}

	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("CheckKeycloakUserVerified", func() {
		Context("When Keycloak actor is not initialized", func() {
			It("should return true and no error", func() {
				mKcAct.EXPECT().IsInitialized().Return(false)
				verified, err := reconciler.CheckKeycloakUserVerified(ctx, tenant)
				Expect(err).NotTo(HaveOccurred())
				Expect(verified).To(BeTrue())
			})
		})

		Context("When Keycloak actor is initialized", func() {
			BeforeEach(func() {
				mKcAct.EXPECT().IsInitialized().Return(true).AnyTimes()
				mKcAct.EXPECT().GetAccessToken().Return("mock-access-token").AnyTimes()
			})

			It("should return true and no error if user is created and email is confirmed", func() {
				mKcAct.EXPECT().GetUser(gomock.Any(), tenantName).Return(&gocloak.User{
					ID:            gocloak.StringP("user-id"),
					EmailVerified: gocloak.BoolP(true),
				}, nil)

				verified, err := reconciler.CheckKeycloakUserVerified(ctx, tenant)
				Expect(err).NotTo(HaveOccurred())
				Expect(verified).To(BeTrue())
				Expect(tenant.Status.Keycloak.UserCreated).To(Equal(crownlabsv1alpha2.NameCreated{
					Name:    "user-id",
					Created: true,
				}))
				Expect(tenant.Status.Keycloak.UserConfirmed).To(BeTrue())
			})

			It("should return false and no error if user is created but email is not confirmed", func() {
				mKcAct.EXPECT().GetUser(gomock.Any(), tenantName).Return(&gocloak.User{
					ID:            gocloak.StringP("user-id"),
					EmailVerified: gocloak.BoolP(false),
				}, nil)

				verified, err := reconciler.CheckKeycloakUserVerified(ctx, tenant)
				Expect(err).NotTo(HaveOccurred())
				Expect(verified).To(BeFalse())
				Expect(tenant.Status.Keycloak.UserCreated).To(Equal(crownlabsv1alpha2.NameCreated{
					Name:    "user-id",
					Created: true,
				}))
				Expect(tenant.Status.Keycloak.UserConfirmed).To(BeFalse())
			})

			Context("When there is an error retrieving the user", func() {
				It("should return an error ", func() {
					mKcAct.EXPECT().GetUser(gomock.Any(), tenantName).Return(nil, gocloak.APIError{
						Message: "error retrieving user",
						Code:    500,
					})

					verified, err := reconciler.CheckKeycloakUserVerified(ctx, tenant)
					Expect(err).To(HaveOccurred())
					Expect(verified).To(BeFalse())
					Expect(err.Error()).To(ContainSubstring("error retrieving user"))
				})
			})

			Context("When the user does not exists in Keycloak", func() {
				It("should create the user", func() {
					gomock.InOrder(
						mKcAct.EXPECT().GetUser(gomock.Any(), tenantName).Return(nil, fmt.Errorf("404")),
						mKcAct.EXPECT().CreateUser(gomock.Any(), tenantName, tenant.Spec.Email, tenant.Spec.FirstName, tenant.Spec.LastName).Return("user-id", nil),
						mKcAct.EXPECT().GetUser(gomock.Any(), tenantName).Return(&gocloak.User{
							ID:            gocloak.StringP("user-id"),
							EmailVerified: gocloak.BoolP(false),
						}, nil),
					)

					verified, err := reconciler.CheckKeycloakUserVerified(ctx, tenant)
					Expect(err).NotTo(HaveOccurred())
					Expect(verified).To(BeFalse())
					Expect(tenant.Status.Keycloak.UserCreated).To(Equal(crownlabsv1alpha2.NameCreated{
						Name:    "user-id",
						Created: true,
					}))
					Expect(tenant.Status.Keycloak.UserConfirmed).To(BeFalse())
				})

				It("should return an error if user creation fails", func() {
					gomock.InOrder(
						mKcAct.EXPECT().GetUser(gomock.Any(), tenantName).Return(nil, fmt.Errorf("404")),
						mKcAct.EXPECT().CreateUser(gomock.Any(), tenantName, tenant.Spec.Email, tenant.Spec.FirstName, tenant.Spec.LastName).Return("", fmt.Errorf("error creating user")),
					)

					verified, err := reconciler.CheckKeycloakUserVerified(ctx, tenant)
					Expect(err).To(HaveOccurred())
					Expect(verified).To(BeFalse())
					Expect(err.Error()).To(ContainSubstring("error creating user"))
				})

				It("should return an error if it is unable to retrieve the newly created user", func() {
					gomock.InOrder(
						mKcAct.EXPECT().GetUser(gomock.Any(), tenantName).Return(nil, fmt.Errorf("404")),
						mKcAct.EXPECT().CreateUser(gomock.Any(), tenantName, tenant.Spec.Email, tenant.Spec.FirstName, tenant.Spec.LastName).Return("user-id", nil),
						mKcAct.EXPECT().GetUser(gomock.Any(), tenantName).Return(nil, fmt.Errorf("error retrieving user")),
					)

					verified, err := reconciler.CheckKeycloakUserVerified(ctx, tenant)
					Expect(err).To(HaveOccurred())
					Expect(verified).To(BeFalse())
					Expect(err.Error()).To(ContainSubstring("error retrieving user"))
				})
			})
		})
	})

	Describe("KeycloakEventHandler", func() {
		Context("When receiving a custom required action event", func() {
			It("should extract username and return 200", func() {
				ch := make(chan event.GenericEvent, 1)
				defer close(ch)

				reconciler.TriggerReconcileChannel = ch

				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "access.CUSTOM_REQUIRED_ACTION",
					"authDetails": {
						"username": "testuser"
					}
				}`)
				r := &http.Request{Body: io.NopCloser(body)}

				reconciler.KeycloakEventHandler(w, r)
				Expect(w.statusCode).To(Equal(http.StatusOK))
				Eventually(ch, timeout, interval).Should(Receive(WithTransform(func(e event.GenericEvent) string {
					return e.Object.(*crownlabsv1alpha2.Tenant).Name
				}, Equal(tenantName))))
			})
		})

		Context("When receiving a custom required action event with no username", func() {
			It("should return 400", func() {
				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "access.CUSTOM_REQUIRED_ACTION",
					"authDetails": {
						"username": ""
					}
				}`)
				r := &http.Request{Body: io.NopCloser(body)}
				reconciler.KeycloakEventHandler(w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("When receiving a malformed auth details on a custom required action event", func() {
			It("should return 400", func() {
				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "access.CUSTOM_REQUIRED_ACTION",
					"authDetails": "notajson"
				}`)
				r := &http.Request{Body: io.NopCloser(body)}
				reconciler.KeycloakEventHandler(w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("When receiving a user update event", func() {
			It("should extract username and return 200", func() {
				ch := make(chan event.GenericEvent, 1)
				defer close(ch)

				reconciler.TriggerReconcileChannel = ch

				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "admin.USER-UPDATE",
					"representation": "{\"username\":\"testuser\"}"
				}`)
				r := &http.Request{Body: io.NopCloser(body)}

				reconciler.KeycloakEventHandler(w, r)
				Expect(w.statusCode).To(Equal(http.StatusOK))
				Eventually(ch, timeout, interval).Should(Receive(WithTransform(func(e event.GenericEvent) string {
					return e.Object.(*crownlabsv1alpha2.Tenant).Name
				}, Equal(tenantName))))
			})
		})

		Context("When receiving a user update event with no username", func() {
			It("should return 400", func() {
				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "admin.USER-UPDATE",
					"representation": "{}"
				}`)
				r := &http.Request{Body: io.NopCloser(body)}
				reconciler.KeycloakEventHandler(w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("When receiving a user update event with malformed representation", func() {
			It("should return 400", func() {
				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "admin.USER-UPDATE",
					"representation": "notajson"
				}`)
				r := &http.Request{Body: io.NopCloser(body)}
				reconciler.KeycloakEventHandler(w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("When receiving a malformed user update event", func() {
			It("should return 400", func() {
				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "admin.USER-UPDATE",
					"representation": "{\"wrongField\":\"testuser\"}"
				}`)
				r := &http.Request{Body: io.NopCloser(body)}
				reconciler.KeycloakEventHandler(w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("When receiving an invalid event type", func() {
			It("should return 400", func() {
				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "invalid.EVENT",
					"representation": "{\"username\":\"testuser\"}"
				}`)
				r := &http.Request{Body: io.NopCloser(body)}

				reconciler.KeycloakEventHandler(w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("When receiving invalid JSON", func() {
			It("should return 400", func() {
				w := &mockResponseWriter{}
				body := strings.NewReader(`invalid json`)
				r := &http.Request{Body: io.NopCloser(body)}

				reconciler.KeycloakEventHandler(w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})
})

// mockResponseWriter implements http.ResponseWriter for testing
type mockResponseWriter struct {
	statusCode int
}

func (w *mockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
