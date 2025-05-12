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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/mock"
)

var _ = Describe("Authentication Unit", func() {
	BeforeEach(func() {
		keycloakActor = mock.NewMockKeycloakActorIface(mockCtrl)
		runReconcile = false
	})

	AfterEach(func() {
		runReconcile = true
	})

	Describe("CheckKeycloakUserVerified", func() {
		Context("When Keycloak actor is not initialized", func() {
			It("should return true and no error", func() {
				keycloakActor.EXPECT().IsInitialized().Return(false)
				verified, err := tenantReconciler.CheckKeycloakUserVerified(ctx, log, tnResource)
				Expect(err).NotTo(HaveOccurred())
				Expect(verified).To(BeTrue())
			})
		})

		Context("When Keycloak actor is initialized", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().IsInitialized().Return(true).AnyTimes()
				keycloakActor.EXPECT().GetAccessToken(gomock.Any()).Return("mock-access-token").AnyTimes()
			})

			It("should return true and no error if user is created and email is confirmed", func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(&gocloak.User{
					ID:            gocloak.StringP("user-id"),
					EmailVerified: gocloak.BoolP(true),
				}, nil)

				verified, err := tenantReconciler.CheckKeycloakUserVerified(ctx, log, tnResource)
				Expect(err).NotTo(HaveOccurred())
				Expect(verified).To(BeTrue())
				Expect(tnResource.Status.Keycloak.UserCreated).To(Equal(v1alpha2.NameCreated{
					Name:    "user-id",
					Created: true,
				}))
				Expect(tnResource.Status.Keycloak.UserConfirmed).To(BeTrue())
			})

			It("should return false and no error if user is created but email is not confirmed", func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(&gocloak.User{
					ID:            gocloak.StringP("user-id"),
					EmailVerified: gocloak.BoolP(false),
				}, nil)

				verified, err := tenantReconciler.CheckKeycloakUserVerified(ctx, log, tnResource)
				Expect(err).NotTo(HaveOccurred())
				Expect(verified).To(BeFalse())
				Expect(tnResource.Status.Keycloak.UserCreated).To(Equal(v1alpha2.NameCreated{
					Name:    "user-id",
					Created: true,
				}))
				Expect(tnResource.Status.Keycloak.UserConfirmed).To(BeFalse())
			})

			Context("When there is an error retrieving the user", func() {
				It("should return an error ", func() {
					keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(nil, gocloak.APIError{
						Message: "error retrieving user",
						Code:    500,
					})

					verified, err := tenantReconciler.CheckKeycloakUserVerified(ctx, log, tnResource)
					Expect(err).To(HaveOccurred())
					Expect(verified).To(BeFalse())
					Expect(err.Error()).To(ContainSubstring("error retrieving user"))
				})
			})

			Context("When the user does not exists in Keycloak", func() {
				It("should create the user", func() {
					gomock.InOrder(
						keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(nil, fmt.Errorf("404")),
						keycloakActor.EXPECT().CreateUser(gomock.Any(), tnName, tnResource.Spec.Email, tnResource.Spec.FirstName, tnResource.Spec.LastName).Return("user-id", nil),
						keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(&gocloak.User{
							ID:            gocloak.StringP("user-id"),
							EmailVerified: gocloak.BoolP(false),
						}, nil),
					)

					verified, err := tenantReconciler.CheckKeycloakUserVerified(ctx, log, tnResource)
					Expect(err).NotTo(HaveOccurred())
					Expect(verified).To(BeFalse())
					Expect(tnResource.Status.Keycloak.UserCreated).To(Equal(v1alpha2.NameCreated{
						Name:    "user-id",
						Created: true,
					}))
					Expect(tnResource.Status.Keycloak.UserConfirmed).To(BeFalse())
				})

				It("should return an error if user creation fails", func() {
					gomock.InOrder(
						keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(nil, fmt.Errorf("404")),
						keycloakActor.EXPECT().CreateUser(gomock.Any(), tnName, tnResource.Spec.Email, tnResource.Spec.FirstName, tnResource.Spec.LastName).Return("", fmt.Errorf("error creating user")),
					)

					verified, err := tenantReconciler.CheckKeycloakUserVerified(ctx, log, tnResource)
					Expect(err).To(HaveOccurred())
					Expect(verified).To(BeFalse())
					Expect(err.Error()).To(ContainSubstring("error creating user"))
				})

				It("should return an error if it is unable to retrieve the newly created user", func() {
					gomock.InOrder(
						keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(nil, fmt.Errorf("404")),
						keycloakActor.EXPECT().CreateUser(gomock.Any(), tnName, tnResource.Spec.Email, tnResource.Spec.FirstName, tnResource.Spec.LastName).Return("user-id", nil),
						keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(nil, fmt.Errorf("error retrieving user")),
					)

					verified, err := tenantReconciler.CheckKeycloakUserVerified(ctx, log, tnResource)
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

				tenantReconciler.TriggerReconcileChannel = ch

				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "access.CUSTOM_REQUIRED_ACTION",
					"authDetails": {
						"username": "testuser"
					}
				}`)
				r := &http.Request{Body: io.NopCloser(body)}

				tenantReconciler.KeycloakEventHandler(log, w, r)
				Expect(w.statusCode).To(Equal(http.StatusOK))
				Eventually(ch, timeout, interval).Should(Receive(WithTransform(func(e event.GenericEvent) string {
					return e.Object.(*v1alpha2.Tenant).Name
				}, Equal(tnName))))
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
				tenantReconciler.KeycloakEventHandler(log, w, r)
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
				tenantReconciler.KeycloakEventHandler(log, w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("When receiving a user update event", func() {
			It("should extract username and return 200", func() {
				ch := make(chan event.GenericEvent, 1)
				defer close(ch)

				tenantReconciler.TriggerReconcileChannel = ch

				w := &mockResponseWriter{}
				body := strings.NewReader(`{
					"type": "admin.USER-UPDATE",
					"representation": "{\"username\":\"testuser\"}"
				}`)
				r := &http.Request{Body: io.NopCloser(body)}

				tenantReconciler.KeycloakEventHandler(log, w, r)
				Expect(w.statusCode).To(Equal(http.StatusOK))
				Eventually(ch, timeout, interval).Should(Receive(WithTransform(func(e event.GenericEvent) string {
					return e.Object.(*v1alpha2.Tenant).Name
				}, Equal(tnName))))
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
				tenantReconciler.KeycloakEventHandler(log, w, r)
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
				tenantReconciler.KeycloakEventHandler(log, w, r)
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
				tenantReconciler.KeycloakEventHandler(log, w, r)
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

				tenantReconciler.KeycloakEventHandler(log, w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("When receiving invalid JSON", func() {
			It("should return 400", func() {
				w := &mockResponseWriter{}
				body := strings.NewReader(`invalid json`)
				r := &http.Request{Body: io.NopCloser(body)}

				tenantReconciler.KeycloakEventHandler(log, w, r)
				Expect(w.statusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})
})

var _ = Describe("Authentication Reconciler", func() {
	BeforeEach(func() {
		keycloakActor = mock.NewMockKeycloakActorIface(mockCtrl)
		keycloakActor.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	})

	Context("When Keycloak is initialized", func() {
		BeforeEach(func() {
			keycloakActor.EXPECT().IsInitialized().Return(true).AnyTimes()
		})

		Context("When the Tenant is not created in Keycloak", func() {
			BeforeEach(func() {
				gomock.InOrder(
					keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(nil, fmt.Errorf("404")),
					keycloakActor.EXPECT().CreateUser(
						gomock.Any(),
						tnResource.Name,
						tnResource.Spec.Email,
						tnResource.Spec.FirstName,
						tnResource.Spec.LastName,
					).Return("test-user-id", nil),
					keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(&gocloak.User{
						ID:            gocloak.StringP("test-user-id"),
						EmailVerified: gocloak.BoolP(false),
					}, nil),
				)
			})

			It("Should set Tenant keycloak status as created but not confirmed", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Keycloak.UserCreated).To(Equal(v1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}))
				Expect(tn.Status.Keycloak.UserConfirmed).To(BeFalse())
			})

			It("Should not create the related resources", func() {
				namespace := &corev1.Namespace{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName}, namespace, BeFalse(), 10*time.Second, 250*time.Millisecond)
			})
		})

		Context("When the Tenant is created but not confirmed in Keycloak", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(&gocloak.User{
					ID:            gocloak.StringP("test-user-id"),
					EmailVerified: gocloak.BoolP(false),
				}, nil)
			})

			It("Should set Tenant keycloak status as created but not confirmed", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Keycloak.UserCreated).To(Equal(v1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}))
				Expect(tn.Status.Keycloak.UserConfirmed).To(BeFalse())
			})

			It("Should not create the related resources", func() {
				namespace := &corev1.Namespace{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName}, namespace, BeFalse(), 10*time.Second, 250*time.Millisecond)
			})
		})

		Context("When the Tenant is created and confirmed in Keycloak", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(&gocloak.User{
					ID:            gocloak.StringP("test-user-id"),
					EmailVerified: gocloak.BoolP(true),
				}, nil)
			})

			It("Should set Tenant keycloak status as confirmed", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Keycloak.UserCreated).To(Equal(v1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}))
				Expect(tn.Status.Keycloak.UserConfirmed).To(BeTrue())
			})

			It("Should create the related resources", func() {
				namespace := &corev1.Namespace{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName}, namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)
			})
		})

		Context("When the Tenant is deleted", func() {
			BeforeEach(func() {
				tenantBeingDeleted()
				tnResource.Status.Keycloak.UserCreated = v1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}
				tnResource.Status.Keycloak.UserConfirmed = true

				keycloakActor.EXPECT().DeleteUser(gomock.Any(), "test-user-id").Return(nil)
			})

			It("Should delete the Tenant keycloak user", func() {
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, &v1alpha2.Tenant{}, BeFalse(), 10*time.Second, 250*time.Millisecond)
			})
		})

		Context("When there is an error on Keycloak operations", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(nil, fmt.Errorf("some error"))
			})

			It("Should make the Tenant keycloak user sync to false", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Ready).To(BeFalse())
				Expect(tn.Status.Keycloak.UserSynchronized).To(BeFalse())
			})
		})

		Context("When the user-id in Keycloak does not match the one in Tenant status", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(&gocloak.User{
					ID:            gocloak.StringP("different-user-id"),
					EmailVerified: gocloak.BoolP(true),
				}, nil)
				tnResource.Status.Keycloak.UserCreated = v1alpha2.NameCreated{
					Name:    "test-user-id",
					Created: true,
				}
			})

			It("Should update the Tenant keycloak status with the new user-id", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
				Expect(tn.Status.Keycloak.UserCreated).To(Equal(v1alpha2.NameCreated{
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
			tn := &v1alpha2.Tenant{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), 10*time.Second, 250*time.Millisecond)
			Expect(tn.Status.Keycloak.UserCreated).To(Equal(v1alpha2.NameCreated{
				Name:    "",
				Created: false,
			}))
			Expect(tn.Status.Keycloak.UserConfirmed).To(BeFalse())
		})

		It("Should create the related resources", func() {
			namespace := &corev1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName}, namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})
	})
})

// mockResponseWriter implements http.ResponseWriter for testing.
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
