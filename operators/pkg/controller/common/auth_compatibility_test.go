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

package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	gk13 "github.com/Nerzal/gocloak/v13"
	"github.com/Nerzal/gocloak/v7"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/mock"
)

var _ = Describe("Auth Compatibility", func() {
	var (
		mockCtrl   *gomock.Controller
		mKcClient  *mock.MockGoCloak
		testLogger logr.Logger
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mKcClient = mock.NewMockGoCloak(mockCtrl)
		testLogger = logr.Discard() // Use a discard logger for tests
	})

	AfterEach(func() {
		mockCtrl.Finish()
		actorIface = &actor
	})

	Describe("SetupKeycloakActorCompatibility", func() {
		Context("when the Keycloak actor is already initialized", func() {
			It("should return without error", func() {
				actorCompatibility = KeycloakActorCompatibility{
					Client:      mKcClient,
					initialized: true,
				}

				err := SetupKeycloakActorCompatibility(context.Background(), "url", "clientID", "clientSecret", "realm", "rolesClientID", testLogger)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the Keycloak actor is not initialized", func() {
			It("should create a new client if it does not exist", func() {
				// ...and provide an error because the target URL is invalid
				actorCompatibility = KeycloakActorCompatibility{
					Client:      nil,
					initialized: false,
				}
				err := SetupKeycloakActorCompatibility(context.Background(), "url", "clientID", "clientSecret", "realm", "rolesClientID", testLogger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not get token"))
			})

			It("should initialize the actor and return without error", func() {
				actorCompatibility = KeycloakActorCompatibility{
					Client: mKcClient,
				}

				mKcClient.EXPECT().LoginClient(gomock.Any(), "clientID", "clientSecret", "realm").Return(&gocloak.JWT{}, nil)

				err := SetupKeycloakActorCompatibility(context.Background(), "url", "clientID", "clientSecret", "realm", "rolesClientID", testLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(actorCompatibility.initialized).To(BeTrue())
				Expect(actorCompatibility.Client).To(Equal(mKcClient))
				Expect(actorCompatibility.Realm).To(Equal("realm"))
				Expect(actorCompatibility.credentials.ClientID).To(Equal("clientID"))
				Expect(actorCompatibility.credentials.ClientSecret).To(Equal("clientSecret"))
			})

			It("should return an error if the login fails", func() {
				actorCompatibility = KeycloakActorCompatibility{
					Client:      mKcClient,
					initialized: false,
				}

				mKcClient.EXPECT().LoginClient(gomock.Any(), "clientID", "clientSecret", "realm").Return(nil, fmt.Errorf("login failed"))

				err := SetupKeycloakActorCompatibility(context.Background(), "url", "clientID", "clientSecret", "realm", "rolesClientID", testLogger)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetKeycloakActor", func() {
		It("should return the current Keycloak actor", func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:      mKcClient,
				initialized: true,
				Realm:       "test-realm",
				credentials: struct {
					ClientID     string
					ClientSecret string
				}{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
				},
			}
			actorIface = &actorCompatibility
			r := GetKeycloakActor()
			result, ok := r.(*KeycloakActorCompatibility)
			Expect(ok).To(BeTrue())
			Expect(result).To(Equal(&actorCompatibility))
			Expect(result.Client).To(Equal(mKcClient))
			Expect(result.Realm).To(Equal("test-realm"))
			Expect(result.credentials.ClientID).To(Equal("test-client"))
			Expect(result.credentials.ClientSecret).To(Equal("test-secret"))
			Expect(result.initialized).To(BeTrue())
		})
	})

	Describe("IsInitialized", func() {
		It("should return false if the actor is not initialized", func() {
			actorCompatibility = KeycloakActorCompatibility{initialized: false}
			Expect(actorCompatibility.IsInitialized()).To(BeFalse())
		})

		It("should return true if the actor is initialized", func() {
			actorCompatibility = KeycloakActorCompatibility{initialized: true}
			Expect(actorCompatibility.IsInitialized()).To(BeTrue())
		})
	})

	Describe("Reset", func() {
		It("should reset all fields", func() {
			actorCompatibility = KeycloakActorCompatibility{
				initialized:    true,
				Client:         mKcClient,
				Realm:          "test-realm",
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: 12345,
				credentials: struct {
					ClientID     string
					ClientSecret string
				}{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
				},
			}

			actorCompatibility.Reset(testLogger)

			Expect(actorCompatibility.initialized).To(BeFalse())
			Expect(actorCompatibility.Client).To(BeNil())
			Expect(actorCompatibility.Realm).To(BeEmpty())
			Expect(actorCompatibility.token).To(BeNil())
			Expect(actorCompatibility.tokenExpiresAt).To(BeZero())
			Expect(actorCompatibility.credentials.ClientID).To(BeEmpty())
			Expect(actorCompatibility.credentials.ClientSecret).To(BeEmpty())
		})
	})

	Describe("GetAccessToken", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:      mKcClient,
				Realm:       "test-realm",
				credentials: struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:  sync.RWMutex{},
			}
		})

		Context("when token is valid and not expired", func() {
			It("should return the existing token", func() {
				now := time.Now().Unix()
				actorCompatibility.token = &gocloak.JWT{AccessToken: "test-token", ExpiresIn: 300}
				actorCompatibility.tokenExpiresAt = now + 300 // valid for 5 minutes

				token := actorCompatibility.GetAccessToken(context.Background())
				Expect(token).To(Equal("test-token"))
			})
		})

		Context("when token is nil", func() {
			It("should initialize a new token", func() {
				actorCompatibility.token = nil
				actorCompatibility.tokenExpiresAt = 0

				mKcClient.EXPECT().LoginClient(gomock.Any(), "test-client", "test-secret", "test-realm").Return(&gocloak.JWT{AccessToken: "new-token", ExpiresIn: 300}, nil)

				token := actorCompatibility.GetAccessToken(context.Background())
				Expect(token).To(Equal("new-token"))
			})
		})

		Context("when token is expired", func() {
			It("should refresh the token", func() {
				now := time.Now().Unix()
				actorCompatibility.token = &gocloak.JWT{AccessToken: "old-token", ExpiresIn: 300}
				actorCompatibility.tokenExpiresAt = now - 100 // expired

				mKcClient.EXPECT().LoginClient(gomock.Any(), "test-client", "test-secret", "test-realm").Return(&gocloak.JWT{AccessToken: "new-token", ExpiresIn: 300}, nil)

				token := actorCompatibility.GetAccessToken(context.Background())
				Expect(token).To(Equal("new-token"))
			})
		})

		Context("when token is about to expire", func() {
			It("should refresh the token", func() {
				now := time.Now().Unix()
				actorCompatibility.token = &gocloak.JWT{AccessToken: "old-token", ExpiresIn: 300}
				actorCompatibility.tokenExpiresAt = now + 10 // expires in 10 seconds

				mKcClient.EXPECT().LoginClient(gomock.Any(), "test-client", "test-secret", "test-realm").Return(&gocloak.JWT{AccessToken: "new-token", ExpiresIn: 300}, nil)

				token := actorCompatibility.GetAccessToken(context.Background())
				Expect(token).To(Equal("new-token"))
			})
		})
	})

	Describe("GetUser", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
			}
		})

		Context("when user exists", func() {
			It("should return the user", func() {
				username := "test-user"
				expectedUser := &gocloak.User{Username: &username}
				expectedUserV13 := &gk13.User{Username: gocloak.StringP(username)}

				mKcClient.EXPECT().GetUsers(
					gomock.Any(),
					"test-token",
					"test-realm",
					gocloak.GetUsersParams{Username: &username},
				).Return([]*gocloak.User{expectedUser}, nil)

				user, err := actorCompatibility.GetUser(context.Background(), username)
				Expect(err).NotTo(HaveOccurred())
				Expect(user).To(Equal(expectedUserV13))
			})
		})

		Context("when user does not exist", func() {
			It("should return an error", func() {
				username := "test-user"

				mKcClient.EXPECT().GetUsers(
					gomock.Any(),
					"test-token",
					"test-realm",
					gocloak.GetUsersParams{Username: &username},
				).Return([]*gocloak.User{}, nil)

				user, err := actorCompatibility.GetUser(context.Background(), username)
				Expect(err).To(MatchError("404"))
				Expect(user).To(BeNil())
			})
		})

		Context("when there is an error fetching the user", func() {
			It("should return an error", func() {
				username := "test-user"

				mKcClient.EXPECT().GetUsers(
					gomock.Any(),
					"test-token",
					"test-realm",
					gocloak.GetUsersParams{Username: &username},
				).Return(nil, fmt.Errorf("error fetching user"))

				user, err := actorCompatibility.GetUser(context.Background(), username)
				Expect(err).To(HaveOccurred())
				Expect(user).To(BeNil())
			})
		})

		Context("when the username is a substring of another username", func() {
			It("should return the right user", func() {
				username := "test-user"
				expectedUser := &gocloak.User{Username: &username}
				expectedUserV13 := &gk13.User{Username: gocloak.StringP(username)}
				otherUser := &gocloak.User{Username: gocloak.StringP("test-user-other")}

				mKcClient.EXPECT().GetUsers(
					gomock.Any(),
					"test-token",
					"test-realm",
					gocloak.GetUsersParams{Username: &username},
				).Return([]*gocloak.User{expectedUser, otherUser}, nil)

				user, err := actorCompatibility.GetUser(context.Background(), username)
				Expect(err).NotTo(HaveOccurred())
				Expect(user).To(Equal(expectedUserV13))
			})
		})
	})

	Describe("CreateUser", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
			}
		})

		It("should create a user and set required actions", func() {
			username := "test-user"
			email := "test@example.com"
			firstName := "Test"
			lastName := "User"
			userID := "test-user-id"

			expectedUser := gocloak.User{
				Username:      &username,
				Email:         &email,
				FirstName:     &firstName,
				LastName:      &lastName,
				Enabled:       gocloak.BoolP(true),
				EmailVerified: gocloak.BoolP(false),
			}

			mKcClient.EXPECT().CreateUser(
				gomock.Any(),
				"test-token",
				"test-realm",
				expectedUser,
			).Return(userID, nil)

			requiredActions := []string{"UPDATE_PASSWORD", "VERIFY_EMAIL"}
			lifespan := 60 * 60 * 24 * 30
			mKcClient.EXPECT().ExecuteActionsEmail(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.ExecuteActionsEmail{
					UserID:   &userID,
					Actions:  &requiredActions,
					Lifespan: &lifespan,
				},
			).Return(nil)

			resultID, err := actorCompatibility.CreateUser(context.Background(), username, email, firstName, lastName)
			Expect(err).NotTo(HaveOccurred())
			Expect(resultID).To(Equal(userID))
		})

		Context("when there is an error creating the user", func() {
			It("should return an error", func() {
				username := "test-user"
				email := "test@example.com"
				firstName := "Test"
				lastName := "User"
				expectedUser := gocloak.User{
					Username:      &username,
					Email:         &email,
					FirstName:     &firstName,
					LastName:      &lastName,
					Enabled:       gocloak.BoolP(true),
					EmailVerified: gocloak.BoolP(false),
				}
				mKcClient.EXPECT().CreateUser(
					gomock.Any(),
					"test-token",
					"test-realm",
					expectedUser,
				).Return("", fmt.Errorf("error creating user"))
				resultID, err := actorCompatibility.CreateUser(context.Background(), username, email, firstName, lastName)
				Expect(err).To(HaveOccurred())
				Expect(resultID).To(BeEmpty())
			})
		})

		Context("when there is an error setting required actions", func() {
			It("should return an error", func() {
				username := "test-user"
				email := "test@example.com"
				firstName := "Test"
				lastName := "User"
				expectedUser := gocloak.User{
					Username:      &username,
					Email:         &email,
					FirstName:     &firstName,
					LastName:      &lastName,
					Enabled:       gocloak.BoolP(true),
					EmailVerified: gocloak.BoolP(false),
				}
				userID := "test-user-id"
				mKcClient.EXPECT().CreateUser(
					gomock.Any(),
					"test-token",
					"test-realm",
					expectedUser,
				).Return(userID, nil)
				requiredActions := []string{"UPDATE_PASSWORD", "VERIFY_EMAIL"}
				lifespan := 60 * 60 * 24 * 30
				mKcClient.EXPECT().ExecuteActionsEmail(
					gomock.Any(),
					"test-token",
					"test-realm",
					gocloak.ExecuteActionsEmail{
						UserID:   &userID,
						Actions:  &requiredActions,
						Lifespan: &lifespan,
					},
				).Return(fmt.Errorf("error setting actions"))
				resultID, err := actorCompatibility.CreateUser(context.Background(), username, email, firstName, lastName)
				Expect(err).To(HaveOccurred())
				Expect(resultID).To(BeEmpty())
			})
		})
	})

	Describe("DeleteUser", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
			}
		})

		It("should delete the user", func() {
			userID := "test-user-id"

			mKcClient.EXPECT().DeleteUser(
				gomock.Any(),
				"test-token",
				"test-realm",
				userID,
			).Return(nil)

			err := actorCompatibility.DeleteUser(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("getClientInternalIdentifierByClientID", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
				clientIDCache:  make(map[string]string),
				cacheMutex:     sync.RWMutex{},
			}
		})

		It("should return the client internal identifier for a given client ID", func() {
			clientID := "test-client-id"
			expectedIdentifier := "internal-identifier"

			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &clientID},
			).Return([]*gocloak.Client{{ClientID: &clientID, ID: &expectedIdentifier}}, nil)

			result, err := actorCompatibility.getClientInternalIdentifierByClientID(context.Background(), clientID)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expectedIdentifier))
		})

		It("should return an error if the client is not found", func() {
			clientID := "non-existent-client"

			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &clientID},
			).Return([]*gocloak.Client{}, nil)

			result, err := actorCompatibility.getClientInternalIdentifierByClientID(context.Background(), clientID)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should return the cached identifier if it exists", func() {
			clientID := "cached-client-id"
			expectedIdentifier := "cached-identifier"

			actorCompatibility.cacheMutex.Lock()
			actorCompatibility.clientIDCache[clientID] = expectedIdentifier
			actorCompatibility.cacheMutex.Unlock()

			result, err := actorCompatibility.getClientInternalIdentifierByClientID(context.Background(), clientID)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expectedIdentifier))
		})

		It("should not make requests to Keycloak if the identifier is cached", func() {
			clientID := "cached-client-id"
			expectedIdentifier := "cached-identifier"

			actorCompatibility.cacheMutex.Lock()
			actorCompatibility.clientIDCache[clientID] = expectedIdentifier
			actorCompatibility.cacheMutex.Unlock()

			mKcClient.EXPECT().GetClients(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			result, err := actorCompatibility.getClientInternalIdentifierByClientID(context.Background(), clientID)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expectedIdentifier))
		})

		It("should store the client identifier in the cache after fetching it", func() {
			clientID := "test-client-id"
			expectedIdentifier := "internal-identifier"

			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &clientID},
			).Return([]*gocloak.Client{{ClientID: &clientID, ID: &expectedIdentifier}}, nil)

			result, err := actorCompatibility.getClientInternalIdentifierByClientID(context.Background(), clientID)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expectedIdentifier))

			actorCompatibility.cacheMutex.RLock()
			cachedIdentifier, exists := actorCompatibility.clientIDCache[clientID]
			actorCompatibility.cacheMutex.RUnlock()
			Expect(exists).To(BeTrue())
			Expect(cachedIdentifier).To(Equal(expectedIdentifier))
		})

		It("should propagate errors from Keycloak", func() {
			clientID := "test-client-id"

			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &clientID},
			).Return(nil, fmt.Errorf("keycloak error"))

			result, err := actorCompatibility.getClientInternalIdentifierByClientID(context.Background(), clientID)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})

	Describe("GetRole", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
				RolesClientID:  "roles-client-id",
				clientIDCache: map[string]string{
					"roles-client-id": "internal-roles-client-id",
				},
			}
		})

		It("should return the role for a given role name", func() {
			roleName := "test-role"
			expectedRole := &gocloak.Role{Name: &roleName}

			mKcClient.EXPECT().GetClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				roleName,
			).Return(expectedRole, nil)
			role, err := actorCompatibility.GetRole(context.Background(), roleName)
			Expect(err).NotTo(HaveOccurred())
			Expect(role).To(Equal(actorCompatibility.convertRoleV7to13(expectedRole)))
		})

		It("should return `404` error if the role is not found", func() {
			roleName := "non-existent-role"

			mKcClient.EXPECT().GetClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				roleName,
			).Return(nil, fmt.Errorf("404 Not Found: Could not find role"))
			role, err := actorCompatibility.GetRole(context.Background(), roleName)
			Expect(err).To(MatchError("404"))
			Expect(role).To(BeNil())
		})

		It("should propagate errors from Keycloak", func() {
			roleName := "test-role"

			mKcClient.EXPECT().GetClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				roleName,
			).Return(nil, fmt.Errorf("keycloak error"))
			role, err := actorCompatibility.GetRole(context.Background(), roleName)
			Expect(err).To(HaveOccurred())
			Expect(role).To(BeNil())
		})

		It("should return an error if the client ID is not found", func() {
			roleName := "test-role"
			actorCompatibility.clientIDCache = make(map[string]string) // Clear cache to simulate missing client ID

			mKcClient.EXPECT().GetClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				roleName,
			).Times(0) // No call to Keycloak since client ID is missing

			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &actorCompatibility.RolesClientID},
			).Return(nil, fmt.Errorf("error"))

			role, err := actorCompatibility.GetRole(context.Background(), roleName)
			Expect(err).To(HaveOccurred())
			Expect(role).To(BeNil())
		})
	})

	Describe("CreateRole", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
				RolesClientID:  "roles-client-id",
				clientIDCache: map[string]string{
					"roles-client-id": "internal-roles-client-id",
				},
			}
		})

		It("should create a role and return its name", func() {
			roleName := "test-role"
			roleDescription := "Test Role Description"

			mKcClient.EXPECT().CreateClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				gocloak.Role{
					Name:        &roleName,
					Description: &roleDescription,
				},
			).Return(roleName, nil)

			roleID, err := actorCompatibility.CreateRole(context.Background(), roleName, roleDescription)
			Expect(err).NotTo(HaveOccurred())
			Expect(roleID).To(Equal(roleName))
		})

		It("should return an error if the role creation fails", func() {
			roleName := "test-role"
			roleDescription := "Test Role Description"

			mKcClient.EXPECT().CreateClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				gocloak.Role{
					Name:        &roleName,
					Description: &roleDescription,
				},
			).Return("", fmt.Errorf("error creating role"))

			roleID, err := actorCompatibility.CreateRole(context.Background(), roleName, roleDescription)
			Expect(err).To(HaveOccurred())
			Expect(roleID).To(BeEmpty())
		})

		It("should return an error if the client ID is not found", func() {
			roleName := "test-role"
			roleDescription := "Test Role Description"
			actorCompatibility.clientIDCache = make(map[string]string) // Clear cache to simulate missing client ID

			mKcClient.EXPECT().CreateClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				gocloak.Role{
					Name:        &roleName,
					Description: &roleDescription,
				},
			).Times(0) // No call to Keycloak since client ID is missing

			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &actorCompatibility.RolesClientID},
			).Return(nil, fmt.Errorf("error"))

			roleID, err := actorCompatibility.CreateRole(context.Background(), roleName, roleDescription)
			Expect(err).To(HaveOccurred())
			Expect(roleID).To(BeEmpty())
		})
	})

	Describe("DeleteRole", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
				RolesClientID:  "roles-client-id",
				clientIDCache: map[string]string{
					"roles-client-id": "internal-roles-client-id",
				},
			}
		})

		It("should delete the role by name", func() {
			roleName := "test-role"

			mKcClient.EXPECT().DeleteClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				roleName,
			).Return(nil)

			err := actorCompatibility.DeleteRole(context.Background(), roleName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error if the role deletion fails", func() {
			roleName := "test-role"

			mKcClient.EXPECT().DeleteClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				roleName,
			).Return(fmt.Errorf("error deleting role"))

			err := actorCompatibility.DeleteRole(context.Background(), roleName)
			Expect(err).To(HaveOccurred())
		})

		It("should return an error if the client ID is not found", func() {
			roleName := "test-role"
			actorCompatibility.clientIDCache = make(map[string]string) // Clear cache to simulate missing client ID

			mKcClient.EXPECT().DeleteClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				roleName,
			).Times(0) // No call to Keycloak since client ID is missing

			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &actorCompatibility.RolesClientID},
			).Return(nil, fmt.Errorf("error"))

			err := actorCompatibility.DeleteRole(context.Background(), roleName)
			Expect(err).To(HaveOccurred())
		})

		It("should return successfully if the role does not exist", func() {
			roleName := "non-existent-role"

			mKcClient.EXPECT().DeleteClientRole(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				roleName,
			).Return(fmt.Errorf("404 Not Found: Could not find role"))

			err := actorCompatibility.DeleteRole(context.Background(), roleName)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetUserRoles", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
				RolesClientID:  "roles-client-id",
				clientIDCache: map[string]string{
					"roles-client-id": "internal-roles-client-id",
				},
			}
		})

		It("should return the roles for a given user ID", func() {
			userID := "test-user-id"
			expectedRoles := []*gocloak.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}}

			mKcClient.EXPECT().GetClientRolesByUserID(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
			).Return(expectedRoles, nil)

			roles, err := actorCompatibility.GetUserRoles(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(roles).To(Equal(actorCompatibility.convertRolesV7to13(expectedRoles)))
		})

		It("should return an error if the client ID is not found", func() {
			userID := "test-user-id"
			actorCompatibility.clientIDCache = make(map[string]string) // Clear cache to simulate missing client ID

			mKcClient.EXPECT().GetClientRolesByUserID(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
			).Times(0) // No call to Keycloak since client ID is missing

			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &actorCompatibility.RolesClientID},
			).Return(nil, fmt.Errorf("error"))

			roles, err := actorCompatibility.GetUserRoles(context.Background(), userID)
			Expect(err).To(HaveOccurred())
			Expect(roles).To(BeNil())
		})

		It("should return an error if the user ID is not found", func() {
			userID := "non-existent-user"

			mKcClient.EXPECT().GetClientRolesByUserID(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
			).Return(nil, fmt.Errorf("404 Not Found: Could not find user roles"))

			roles, err := actorCompatibility.GetUserRoles(context.Background(), userID)
			Expect(err).To(MatchError("404"))
			Expect(roles).To(BeNil())
		})

		It("should propagate errors from Keycloak", func() {
			userID := "test-user-id"

			mKcClient.EXPECT().GetClientRolesByUserID(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
			).Return(nil, fmt.Errorf("keycloak error"))

			roles, err := actorCompatibility.GetUserRoles(context.Background(), userID)
			Expect(err).To(HaveOccurred())
			Expect(roles).To(BeNil())
		})
	})

	Describe("AddUserToRoles", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
				RolesClientID:  "roles-client-id",
				clientIDCache: map[string]string{
					"roles-client-id": "internal-roles-client-id",
				},
			}
		})

		It("should add a user to multiple roles", func() {
			userID := "test-user-id"
			roles := []*gk13.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}}

			mKcClient.EXPECT().AddClientRoleToUser(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
				[]gocloak.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}},
			).Return(nil)

			err := actorCompatibility.AddUserToRoles(context.Background(), userID, roles)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error if the client ID is not found", func() {
			userID := "test-user-id"
			roles := []*gk13.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}}
			actorCompatibility.clientIDCache = make(map[string]string) // Clear cache to simulate missing client ID
			mKcClient.EXPECT().AddClientRoleToUser(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
				[]gocloak.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}},
			).Times(0) // No call to Keycloak since client ID is missing
			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &actorCompatibility.RolesClientID},
			).Return(nil, fmt.Errorf("error"))

			err := actorCompatibility.AddUserToRoles(context.Background(), userID, roles)
			Expect(err).To(HaveOccurred())
		})

		It("should propagate errors from Keycloak", func() {
			userID := "test-user-id"
			roles := []*gk13.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}}

			mKcClient.EXPECT().AddClientRoleToUser(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
				[]gocloak.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}},
			).Return(fmt.Errorf("keycloak error"))

			err := actorCompatibility.AddUserToRoles(context.Background(), userID, roles)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("RemoveUserFromRoles", func() {
		BeforeEach(func() {
			actorCompatibility = KeycloakActorCompatibility{
				Client:         mKcClient,
				Realm:          "test-realm",
				credentials:    struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:     sync.RWMutex{},
				token:          &gocloak.JWT{AccessToken: "test-token"},
				tokenExpiresAt: time.Now().Unix() + 3600, // valid for 1 hour
				RolesClientID:  "roles-client-id",
				clientIDCache: map[string]string{
					"roles-client-id": "internal-roles-client-id",
				},
			}
		})

		It("should remove a user from multiple roles", func() {
			userID := "test-user-id"
			roles := []*gk13.Role{{Name: gk13.StringP("role1")}, {Name: gk13.StringP("role2")}}

			mKcClient.EXPECT().DeleteClientRoleFromUser(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
				[]gocloak.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}},
			).Return(nil)

			err := actorCompatibility.RemoveUserFromRoles(context.Background(), userID, roles)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error if the client ID is not found", func() {
			userID := "test-user-id"
			roles := []*gk13.Role{{Name: gk13.StringP("role1")}, {Name: gk13.StringP("role2")}}
			actorCompatibility.clientIDCache = make(map[string]string) // Clear cache to simulate missing client ID
			mKcClient.EXPECT().DeleteClientRoleFromUser(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
				[]gocloak.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}},
			).Times(0) // No call to Keycloak since client ID is missing
			mKcClient.EXPECT().GetClients(
				gomock.Any(),
				"test-token",
				"test-realm",
				gocloak.GetClientsParams{ClientID: &actorCompatibility.RolesClientID},
			).Return(nil, fmt.Errorf("error"))

			err := actorCompatibility.RemoveUserFromRoles(context.Background(), userID, roles)
			Expect(err).To(HaveOccurred())
		})

		It("should propagate errors from Keycloak", func() {
			userID := "test-user-id"
			roles := []*gk13.Role{{Name: gk13.StringP("role1")}, {Name: gk13.StringP("role2")}}

			mKcClient.EXPECT().DeleteClientRoleFromUser(
				gomock.Any(),
				"test-token",
				"test-realm",
				"internal-roles-client-id",
				userID,
				[]gocloak.Role{{Name: gocloak.StringP("role1")}, {Name: gocloak.StringP("role2")}},
			).Return(fmt.Errorf("keycloak error"))

			err := actorCompatibility.RemoveUserFromRoles(context.Background(), userID, roles)
			Expect(err).To(HaveOccurred())
		})
	})
})
