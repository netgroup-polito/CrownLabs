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

package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/utils/mock_utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Auth", func() {
	var (
		mockCtrl  *gomock.Controller
		mKcClient *mock_utils.MockGoCloakIface
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mKcClient = mock_utils.NewMockGoCloakIface(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("SetupKeycloakActor", func() {
		Context("when the Keycloak actor is already initialized", func() {
			It("should return without error", func() {
				actor = KeycloakActor{
					Client:      mKcClient,
					initialized: true,
				}

				err := SetupKeycloakActor("url", "clientID", "clientSecret", "realm")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the Keycloak actor is not initialized", func() {
			It("should create a new client if it does not exist", func() {
				// ...and provide an error because the target URL is invalid
				actor = KeycloakActor{
					Client:      nil,
					initialized: false,
				}
				err := SetupKeycloakActor("url", "clientID", "clientSecret", "realm")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not get token"))
			})

			It("should initialize the actor and return without error", func() {
				actor = KeycloakActor{
					Client: mKcClient,
				}

				mKcClient.EXPECT().LoginClient(gomock.Any(), "clientID", "clientSecret", "realm").Return(&gocloak.JWT{}, nil)

				err := SetupKeycloakActor("url", "clientID", "clientSecret", "realm")
				Expect(err).NotTo(HaveOccurred())
				Expect(actor.initialized).To(BeTrue())
				Expect(actor.Client).To(Equal(mKcClient))
				Expect(actor.Realm).To(Equal("realm"))
				Expect(actor.credentials.ClientID).To(Equal("clientID"))
				Expect(actor.credentials.ClientSecret).To(Equal("clientSecret"))
			})

			It("should return an error if the login fails", func() {
				actor = KeycloakActor{
					Client:      mKcClient,
					initialized: false,
				}

				mKcClient.EXPECT().LoginClient(gomock.Any(), "clientID", "clientSecret", "realm").Return(nil, fmt.Errorf("login failed"))

				err := SetupKeycloakActor("url", "clientID", "clientSecret", "realm")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetKeycloakActor", func() {
		It("should return the current Keycloak actor", func() {
			actor = KeycloakActor{
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
			result := GetKeycloakActor()
			Expect(result).To(Equal(&actor))
			Expect(result.Client).To(Equal(mKcClient))
			Expect(result.Realm).To(Equal("test-realm"))
			Expect(result.credentials.ClientID).To(Equal("test-client"))
			Expect(result.credentials.ClientSecret).To(Equal("test-secret"))
			Expect(result.initialized).To(BeTrue())
		})
	})

	Describe("IsInitialized", func() {
		It("should return false if the actor is not initialized", func() {
			actor = KeycloakActor{initialized: false}
			Expect(actor.IsInitialized()).To(BeFalse())
		})

		It("should return true if the actor is initialized", func() {
			actor = KeycloakActor{initialized: true}
			Expect(actor.IsInitialized()).To(BeTrue())
		})
	})

	Describe("Reset", func() {
		It("should reset all fields", func() {
			actor = KeycloakActor{
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

			actor.Reset()

			Expect(actor.initialized).To(BeFalse())
			Expect(actor.Client).To(BeNil())
			Expect(actor.Realm).To(BeEmpty())
			Expect(actor.token).To(BeNil())
			Expect(actor.tokenExpiresAt).To(BeZero())
			Expect(actor.credentials.ClientID).To(BeEmpty())
			Expect(actor.credentials.ClientSecret).To(BeEmpty())
		})
	})

	Describe("GetAccessToken", func() {
		BeforeEach(func() {
			actor = KeycloakActor{
				Client:      mKcClient,
				Realm:       "test-realm",
				credentials: struct{ ClientID, ClientSecret string }{ClientID: "test-client", ClientSecret: "test-secret"},
				tokenMutex:  sync.RWMutex{},
			}
		})

		Context("when token is valid and not expired", func() {
			It("should return the existing token", func() {
				now := time.Now().Unix()
				actor.token = &gocloak.JWT{AccessToken: "test-token", ExpiresIn: 300}
				actor.tokenExpiresAt = now + 300 // valid for 5 minutes

				token := actor.GetAccessToken()
				Expect(token).To(Equal("test-token"))
			})
		})

		Context("when token is nil", func() {
			It("should initialize a new token", func() {
				actor.token = nil
				actor.tokenExpiresAt = 0

				mKcClient.EXPECT().LoginClient(gomock.Any(), "test-client", "test-secret", "test-realm").Return(&gocloak.JWT{AccessToken: "new-token", ExpiresIn: 300}, nil)

				token := actor.GetAccessToken()
				Expect(token).To(Equal("new-token"))
			})
		})

		Context("when token is expired", func() {
			It("should refresh the token", func() {
				now := time.Now().Unix()
				actor.token = &gocloak.JWT{AccessToken: "old-token", ExpiresIn: 300}
				actor.tokenExpiresAt = now - 100 // expired

				mKcClient.EXPECT().LoginClient(gomock.Any(), "test-client", "test-secret", "test-realm").Return(&gocloak.JWT{AccessToken: "new-token", ExpiresIn: 300}, nil)

				token := actor.GetAccessToken()
				Expect(token).To(Equal("new-token"))
			})
		})

		Context("when token is about to expire", func() {
			It("should refresh the token", func() {
				now := time.Now().Unix()
				actor.token = &gocloak.JWT{AccessToken: "old-token", ExpiresIn: 300}
				actor.tokenExpiresAt = now + 10 // expires in 10 seconds

				mKcClient.EXPECT().LoginClient(gomock.Any(), "test-client", "test-secret", "test-realm").Return(&gocloak.JWT{AccessToken: "new-token", ExpiresIn: 300}, nil)

				token := actor.GetAccessToken()
				Expect(token).To(Equal("new-token"))
			})
		})
	})

	Describe("GetUser", func() {
		BeforeEach(func() {
			actor = KeycloakActor{
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

				mKcClient.EXPECT().GetUsers(
					gomock.Any(),
					"test-token",
					"test-realm",
					gocloak.GetUsersParams{Username: &username},
				).Return([]*gocloak.User{expectedUser}, nil)

				user, err := actor.GetUser(context.Background(), username)
				Expect(err).NotTo(HaveOccurred())
				Expect(user).To(Equal(expectedUser))
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

				user, err := actor.GetUser(context.Background(), username)
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

				user, err := actor.GetUser(context.Background(), username)
				Expect(err).To(HaveOccurred())
				Expect(user).To(BeNil())
			})
		})
	})

	Describe("CreateUser", func() {
		BeforeEach(func() {
			actor = KeycloakActor{
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

			resultID, err := actor.CreateUser(context.Background(), username, email, firstName, lastName)
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
				resultID, err := actor.CreateUser(context.Background(), username, email, firstName, lastName)
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
				resultID, err := actor.CreateUser(context.Background(), username, email, firstName, lastName)
				Expect(err).To(HaveOccurred())
				Expect(resultID).To(BeEmpty())
			})
		})
	})

	Describe("DeleteUser", func() {
		BeforeEach(func() {
			actor = KeycloakActor{
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

			err := actor.DeleteUser(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
