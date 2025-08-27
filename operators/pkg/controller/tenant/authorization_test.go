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

	"github.com/Nerzal/gocloak/v13"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/mock"
)

var _ = Describe("Authorization", func() {
	BeforeEach(func() {
		keycloakActor = mock.NewMockKeycloakActorIface(mockCtrl)
	})

	Context("When Keycloak actor is initialized", func() {
		ws1 := v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ws1",
			},
		}
		ws2 := v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ws2",
			},
		}

		BeforeEach(func() {
			keycloakActor.EXPECT().IsInitialized().Return(true).AnyTimes()
			keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).Return(&gocloak.User{
				Username:      gocloak.StringP(tnName),
				ID:            gocloak.StringP("user-id"),
				EmailVerified: gocloak.BoolP(true),
			}, nil).AnyTimes()

			tnResource.Status = v1alpha2.TenantStatus{
				Subscriptions: map[string]v1alpha2.SubscriptionStatus{
					"keycloak": v1alpha2.SubscrOk,
				},
				Keycloak: v1alpha2.KeycloakStatus{
					UserCreated: v1alpha2.NameCreated{
						Name:    "user-id",
						Created: true,
					},
					UserConfirmed: true,
				},
			}

			addObjToObjectsList(&ws1)
			addObjToObjectsList(&ws2)
		})

		AfterEach(func() {
			removeObjFromObjectsList(&ws1)
			removeObjFromObjectsList(&ws2)
		})

		Context("When no workspaces are present in tenant and no roles in Keycloak", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{}, nil).AnyTimes()
			})

			It("Should not add/update/delete any roles to Keycloak", func() {
				// No expected call to add roles
			})
		})

		Context("When no workspaces are present in tenant but roles in Keycloak", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:manager"),
						Name: gocloak.StringP("workspace-ws1:manager"),
					},
					{
						ID:   gocloak.StringP("workspace-ws2:user"),
						Name: gocloak.StringP("workspace-ws2:user"),
					},
					{
						ID:   gocloak.StringP("no-ws-role"),
						Name: gocloak.StringP("no-ws-role"),
					},
				}, nil).AnyTimes()

				keycloakActor.EXPECT().RemoveUserFromRoles(gomock.Any(), "user-id", []*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:manager"),
						Name: gocloak.StringP("workspace-ws1:manager"),
					},
					{
						ID:   gocloak.StringP("workspace-ws2:user"),
						Name: gocloak.StringP("workspace-ws2:user"),
					},
				}).Return(nil).Times(1)
			})

			It("Should delete roles related to workspaces, and no other roles", func() {
				// Expected call to remove roles
			})
		})

		Context("When workspaces are present in tenant and roles in Keycloak", func() {
			BeforeEach(func() {
				tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{
					{
						Name: "ws1",
						Role: v1alpha2.Manager,
					},
					{
						Name: "ws2",
						Role: v1alpha2.User,
					},
				}

				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:manager"),
						Name: gocloak.StringP("workspace-ws1:manager"),
					},
					{
						ID:   gocloak.StringP("workspace-ws2:user"),
						Name: gocloak.StringP("workspace-ws2:user"),
					},
					{
						ID:   gocloak.StringP("no-ws-role"),
						Name: gocloak.StringP("no-ws-role"),
					},
				}, nil).AnyTimes()
			})

			It("Should not add nor delete roles", func() {
				// Expected no calls to add and remove roles
			})
		})

		Context("When workspaces are present in tenant and roles in Keycloak are missing", func() {
			BeforeEach(func() {
				tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{
					{
						Name: "ws1",
						Role: v1alpha2.Manager,
					},
					{
						Name: "ws2",
						Role: v1alpha2.User,
					},
				}

				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:manager"),
						Name: gocloak.StringP("workspace-ws1:manager"),
					},
				}, nil).AnyTimes()

				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-ws2:user").Return(&gocloak.Role{
					ID:   gocloak.StringP("workspace-ws2:user"),
					Name: gocloak.StringP("workspace-ws2:user"),
				}, nil).AnyTimes()

				keycloakActor.EXPECT().AddUserToRoles(gomock.Any(), "user-id", []*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws2:user"),
						Name: gocloak.StringP("workspace-ws2:user"),
					},
				}).Return(nil).Times(1)
			})

			It("Should add missing roles to Keycloak", func() {
				// Expected call to add roles
			})
		})

		Context("When workspaces are present in tenant and roles in Keycloak have different function (manager/user)", func() {
			BeforeEach(func() {
				tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{
					{
						Name: "ws1",
						Role: v1alpha2.User, // Changed from Manager to User
					},
					{
						Name: "ws2",
						Role: v1alpha2.Manager, // Changed from User to Manager
					},
				}

				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:manager"),
						Name: gocloak.StringP("workspace-ws1:manager"),
					},
					{
						ID:   gocloak.StringP("workspace-ws2:user"),
						Name: gocloak.StringP("workspace-ws2:user"),
					},
				}, nil).AnyTimes()

				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-ws1:user").Return(&gocloak.Role{
					ID:   gocloak.StringP("workspace-ws1:user"),
					Name: gocloak.StringP("workspace-ws1:user"),
				}, nil).AnyTimes()

				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-ws2:manager").Return(&gocloak.Role{
					ID:   gocloak.StringP("workspace-ws2:manager"),
					Name: gocloak.StringP("workspace-ws2:manager"),
				}, nil).AnyTimes()

				keycloakActor.EXPECT().AddUserToRoles(gomock.Any(), "user-id", []*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:user"),
						Name: gocloak.StringP("workspace-ws1:user"),
					},
					{
						ID:   gocloak.StringP("workspace-ws2:manager"),
						Name: gocloak.StringP("workspace-ws2:manager"),
					},
				}).Return(nil).Times(1)

				keycloakActor.EXPECT().RemoveUserFromRoles(gomock.Any(), "user-id", []*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:manager"),
						Name: gocloak.StringP("workspace-ws1:manager"),
					},
					{
						ID:   gocloak.StringP("workspace-ws2:user"),
						Name: gocloak.StringP("workspace-ws2:user"),
					},
				}).Return(nil).Times(1)
			})

			It("Should add roles with new functions and remove old ones", func() {
				// Expected calls to add and remove roles
			})
		})

		Context("When a workspace is not valid", func() {
			BeforeEach(func() {
				tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{
					{
						Name: "invalid-ws",
						Role: v1alpha2.User,
					},
				}

				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{}, nil).AnyTimes()
			})

			It("Should not add roles for invalid workspaces", func() {
				// Expected no call to add roles for invalid workspace
			})
		})

		Context("When tenant has candidate status in workspace", func() {
			BeforeEach(func() {
				tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{
					{
						Name: "ws1",
						Role: v1alpha2.Candidate,
					},
				}

				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{}, nil).AnyTimes()
			})

			It("Should not add roles for candidate workspaces", func() {
				// Expected no call to add roles for candidate workspace
			})
		})

		Context("When there is an error getting current user roles from Keycloak", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return(nil, fmt.Errorf("some error")).AnyTimes()
			})

			It("Should set Keycloak user sync to false", func() {
				var tn v1alpha2.Tenant
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, &tn, BeTrue(), timeout, interval)

				Expect(tn.Status.Keycloak.UserSynchronized).To(BeFalse())
			})
		})

		Context("When there is an error retrieving role information from Keycloak", func() {
			BeforeEach(func() {
				tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{
					{
						Name: "ws1",
						Role: v1alpha2.Manager,
					},
				}

				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{}, nil).AnyTimes()
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-ws1:manager").Return(nil, fmt.Errorf("some error")).AnyTimes()
			})

			It("Should set Keycloak user sync to false", func() {
				var tn v1alpha2.Tenant
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, &tn, BeTrue(), timeout, interval)

				Expect(tn.Status.Keycloak.UserSynchronized).To(BeFalse())
			})
		})

		Context("When there is an error adding roles to Keycloak", func() {
			BeforeEach(func() {
				tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{
					{
						Name: "ws1",
						Role: v1alpha2.Manager,
					},
				}

				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{}, nil).AnyTimes()
				keycloakActor.EXPECT().GetRole(gomock.Any(), "workspace-ws1:manager").Return(&gocloak.Role{
					ID:   gocloak.StringP("workspace-ws1:manager"),
					Name: gocloak.StringP("workspace-ws1:manager"),
				}, nil).AnyTimes()

				keycloakActor.EXPECT().AddUserToRoles(gomock.Any(), "user-id", []*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:manager"),
						Name: gocloak.StringP("workspace-ws1:manager"),
					},
				}).Return(fmt.Errorf("some error")).Times(1)
			})

			It("Should set Keycloak user sync to false", func() {
				var tn v1alpha2.Tenant
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, &tn, BeTrue(), timeout, interval)

				Expect(tn.Status.Keycloak.UserSynchronized).To(BeFalse())
			})
		})

		Context("When there is an error removing roles from Keycloak", func() {
			BeforeEach(func() {
				keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return([]*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:manager"),
						Name: gocloak.StringP("workspace-ws1:manager"),
					},
				}, nil).AnyTimes()

				keycloakActor.EXPECT().RemoveUserFromRoles(gomock.Any(), "user-id", []*gocloak.Role{
					{
						ID:   gocloak.StringP("workspace-ws1:manager"),
						Name: gocloak.StringP("workspace-ws1:manager"),
					},
				}).Return(fmt.Errorf("some error")).Times(1)
			})

			It("Should set Keycloak user sync to false", func() {
				var tn v1alpha2.Tenant
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, &tn, BeTrue(), timeout, interval)

				Expect(tn.Status.Keycloak.UserSynchronized).To(BeFalse())
			})
		})
	})

	Context("When Keycloak actor is not initialized", func() {
		BeforeEach(func() {
			keycloakActor.EXPECT().IsInitialized().Return(false).AnyTimes()
		})

		It("Should not get nor add any roles to Keycloak", func() {
			// No expectations on Keycloak actor methods
		})
	})
})
