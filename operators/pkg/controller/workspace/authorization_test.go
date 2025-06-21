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
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-logr/logr"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/mock"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/workspace"
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
)

var _ = Describe("Authorization", func() {
	var (
		ctx                 context.Context
		builder             fake.ClientBuilder
		cl                  client.Client
		mockCtrl            *gomock.Controller
		keycloakActor       *mock.MockKeycloakActorIface
		workspaceReconciler workspace.WorkspaceReconciler

		wsResource             *v1alpha1.Workspace
		wsReconcileErrExpected gomegaTypes.GomegaMatcher
	)

	const (
		timeout      = time.Second * 10
		interval     = time.Millisecond * 250
		wsName       = "test-workspace"
		wsPrettyName = "Test Workspace"
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		builder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

		mockCtrl = gomock.NewController(GinkgoT())
		keycloakActor = mock.NewMockKeycloakActorIface(mockCtrl)

		wsResource = &v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name: wsName,
				Labels: map[string]string{
					"crownlabs.polito.it/target": "test",
				},
			},
			Spec: v1alpha1.WorkspaceSpec{
				PrettyName: wsPrettyName,
			},
		}
		wsReconcileErrExpected = Not(HaveOccurred())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	JustBeforeEach(func() {
		cl = builder.WithObjects(wsResource).WithStatusSubresource(wsResource).Build()

		workspaceReconciler = workspace.WorkspaceReconciler{
			Client:        cl,
			Scheme:        scheme.Scheme,
			KeycloakActor: keycloakActor,
			TargetLabel:   common.NewLabel("crownlabs.polito.it/target", "test"),
		}

		_, err := workspaceReconciler.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name: wsName,
			},
		})
		Expect(err).To(wsReconcileErrExpected)
	})

	Context("When Keycloak is initialized", func() {
		BeforeEach(func() {
			keycloakActor.EXPECT().IsInitialized().Return(true).AnyTimes()
		})

		Context("When Keycloak roles are not yet present", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-"+wsName+"-manager").Return(nil, nil).Times(1)
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-"+wsName+"-user").Return(nil, nil).Times(1)
				keycloakActor.EXPECT().CreateRole(gomock.Any(), "workspace-"+wsName+"-manager", wsPrettyName+" Manager Role").Times(1)
				keycloakActor.EXPECT().CreateRole(gomock.Any(), "workspace-"+wsName+"-user", wsPrettyName+" User Role").Times(1)
			})

			It("Should create Keycloak roles", func() {
				// checked in beforeEach
			})

			It("Should set the Keycloak subscription to ok", func() {
				ws := &v1alpha1.Workspace{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)
				Expect(ws.Status.Subscriptions["keycloak"]).To(Equal(v1alpha2.SubscrOk))
			})
		})

		Context("When Keycloak roles are already present", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-"+wsName+"-manager").Return(&gocloak.Role{Name: gocloak.StringP("workspace-" + wsName + "-manager")}, nil).Times(1)
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-"+wsName+"-user").Return(&gocloak.Role{Name: gocloak.StringP("workspace-" + wsName + "-user")}, nil).Times(1)
				keycloakActor.EXPECT().CreateRole(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			})

			It("Should not create Keycloak roles", func() {
				// checked in beforeEach
			})

			It("Should set the Keycloak subscription to ok", func() {
				ws := &v1alpha1.Workspace{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)
				Expect(ws.Status.Subscriptions["keycloak"]).To(Equal(v1alpha2.SubscrOk))
			})
		})

		Context("When the Workspace is being deleted", func() {
			BeforeEach(func() {
				wsResource.Finalizers = append(wsResource.Finalizers, v1alpha2.TnOperatorFinalizerName)
				wsResource.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				wsResource.Status.Namespace = v1alpha2.NameCreated{
					Created: true,
					Name:    "workspace-" + wsName,
				}
				wsResource.Status.Subscriptions = map[string]v1alpha2.SubscriptionStatus{
					"keycloak": v1alpha2.SubscrOk,
				}
				keycloakActor.EXPECT().DeleteRole(gomock.Any(), "workspace-"+wsName+"-manager").Return(nil).Times(1)
				keycloakActor.EXPECT().DeleteRole(gomock.Any(), "workspace-"+wsName+"-user").Return(nil).Times(1)
			})

			It("Should delete Keycloak roles", func() {
				// checked in beforeEach
			})
		})

		Context("When an error occurs in the getRole call", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-"+wsName+"-manager").Return(nil, fmt.Errorf("error getting role")).Times(1)
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-"+wsName+"-user").Return(nil, nil).AnyTimes()
				wsReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in beforeEach
			})
		})

		Context("When an error occurs in the createRole call", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-"+wsName+"-manager").Return(nil, nil).AnyTimes()
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-"+wsName+"-user").Return(nil, nil).AnyTimes()
				keycloakActor.EXPECT().CreateRole(gomock.Any(), "workspace-"+wsName+"-manager", wsPrettyName+" Manager Role").Return("", fmt.Errorf("error creating role")).Times(1)
				wsReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in beforeEach
			})
		})

		Context("When an error occurs in the deleteRole call", func() {
			BeforeEach(func() {
				wsResource.Finalizers = append(wsResource.Finalizers, v1alpha2.TnOperatorFinalizerName)
				wsResource.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				wsResource.Status.Namespace = v1alpha2.NameCreated{
					Created: true,
					Name:    "workspace-" + wsName,
				}
				wsResource.Status.Subscriptions = map[string]v1alpha2.SubscriptionStatus{
					"keycloak": v1alpha2.SubscrOk,
				}
				keycloakActor.EXPECT().DeleteRole(gomock.Any(), "workspace-"+wsName+"-user").Return(fmt.Errorf("error deleting role")).AnyTimes()
				keycloakActor.EXPECT().DeleteRole(gomock.Any(), "workspace-"+wsName+"-manager").Return(fmt.Errorf("error deleting role")).AnyTimes()
				wsReconcileErrExpected = HaveOccurred()
			})

			It("Should return an error", func() {
				// checked in beforeEach
			})
		})
	})

	Context("When Keycloak is not initialized", func() {
		BeforeEach(func() {
			keycloakActor.EXPECT().IsInitialized().Return(false).AnyTimes()
		})

		It("Should set the Keycloak subscription to failed", func() {
			ws := &v1alpha1.Workspace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: wsName}, ws, BeTrue(), timeout, interval)
			Expect(ws.Status.Subscriptions["keycloak"]).To(Equal(v1alpha2.SubscrFailed))
		})

		It("Should not create Keycloak roles", func() {
			keycloakActor.EXPECT().CreateRole(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			keycloakActor.EXPECT().GetRole(gomock.Any(), gomock.Any()).Times(0)
		})

		It("Should create the related resources", func() {
			namespaceName := "workspace-" + wsName
			ns := &corev1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: namespaceName}, ns, BeTrue(), timeout, interval)
			Expect(ns.Name).To(Equal(namespaceName))
		})

		Context("When the Workspace is being deleted", func() {
			BeforeEach(func() {
				wsResource.Finalizers = append(wsResource.Finalizers, v1alpha2.TnOperatorFinalizerName)
				wsResource.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				wsResource.Status.Namespace = v1alpha2.NameCreated{
					Created: true,
					Name:    "workspace-" + wsName,
				}
			})

			It("Should not delete Keycloak roles", func() {
				keycloakActor.EXPECT().DeleteRole(gomock.Any(), gomock.Any()).Times(0)
			})
		})
	})
})

func DoesEventuallyExists(ctx context.Context, cl client.Client, objLookupKey client.ObjectKey, targetObj client.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout, interval time.Duration) {
	Eventually(func() bool {
		err := cl.Get(ctx, objLookupKey, targetObj)
		return err == nil
	}, timeout, interval).Should(expectedStatus)
}
