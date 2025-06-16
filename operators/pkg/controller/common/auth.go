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

// Package main contains the entrypoint for the Crownlabs unified operator.
package common

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"k8s.io/klog/v2"
)

// KcActor contains the needed objects and infos to use keycloak functionalities.
type KeycloakActor struct {
	initialized    bool
	Client         GoCloakIface
	Realm          string
	token          *gocloak.JWT
	tokenMutex     sync.RWMutex
	tokenExpiresAt int64
	credentials    struct {
		ClientID     string
		ClientSecret string
	}
	RolesClientID string // The client ID of the client in which the roles are defined.
}

const tokenRefreshBuffer = 30 // the token is considered about to expire if it has less than this many seconds left

var actor KeycloakActor

func SetupKeycloakActor(
	url string,
	clientID string,
	clientSecret string,
	realm string,
	rolesClientID string,
) error {
	if actor.initialized {
		return nil
	}

	if actor.Client == nil {
		actor.Client = gocloak.NewClient(url)
	}

	// login to keycloak
	_, err := actor.Client.LoginClient(context.Background(), clientID, clientSecret, realm)
	if err != nil {
		klog.Error("Unable to login as admin on keycloak: ", err)
		return err
	}

	actor.Realm = realm
	actor.credentials.ClientID = clientID
	actor.credentials.ClientSecret = clientSecret
	actor.RolesClientID = rolesClientID
	actor.initialized = true
	return nil
}

// GetKeycloakActor returns the KcActor currently used.
func GetKeycloakActor() *KeycloakActor {
	return &actor
}

func (a *KeycloakActor) IsInitialized() bool {
	return a.initialized
}

func (a *KeycloakActor) Reset() {
	a.tokenMutex.Lock()
	defer a.tokenMutex.Unlock()
	a.initialized = false
	a.Client = nil
	a.Realm = ""
	a.token = nil
	a.tokenExpiresAt = 0
	a.credentials.ClientID = ""
	a.credentials.ClientSecret = ""
	klog.Info("Keycloak actor has been reset")
}

// GetAccessToken returns the access token of the actor.
// It tries to refresh the token if it is nil or expired.
func (a *KeycloakActor) GetAccessToken() string {
	a.tokenMutex.RLock()
	defer a.tokenMutex.RUnlock()

	now := time.Now().Unix()
	// renew rhe token if it is nil or about to expire
	if a.token == nil || now >= (a.tokenExpiresAt-tokenRefreshBuffer) {
		klog.Info("Keycloak token is not present or about to expire, refreshing it")
		ctx := context.Background()
		token, err := a.Client.LoginClient(ctx, a.credentials.ClientID, a.credentials.ClientSecret, a.Realm)
		if err != nil {
			klog.Error("Unable to refresh keycloak token", err)
			os.Exit(1)
		}

		a.token = token
		// set the token expiration time
		a.tokenExpiresAt = now + int64(token.ExpiresIn)
	}

	return a.token.AccessToken
}

// GetUser returns the user associated with the given username.
func (a *KeycloakActor) GetUser(
	ctx context.Context,
	username string,
) (*gocloak.User, error) {

	users, err := a.Client.GetUsers(ctx, a.GetAccessToken(), a.Realm, gocloak.GetUsersParams{
		Username: &username,
	})
	if err != nil {
		klog.Error("Unable to get user from keycloak", err)
		return nil, err
	}

	if len(users) != 1 {
		klog.Warningf("User %s not found in Keycloak", username)
		return nil, fmt.Errorf("404")
	}

	user := users[0]

	return user, nil
}

// CreateUser creates a user in Keycloak.
func (a *KeycloakActor) CreateUser(
	ctx context.Context,
	username string,
	email string,
	firstName string,
	lastName string,
) (string, error) {
	user := gocloak.User{
		Username:      &username,
		Email:         &email,
		FirstName:     &firstName,
		LastName:      &lastName,
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(false),
	}

	userID, err := a.Client.CreateUser(ctx, a.GetAccessToken(), a.Realm, user)
	if err != nil {
		return "", err
	}

	// Set the required actions for the user
	requiredActions := []string{"UPDATE_PASSWORD", "VERIFY_EMAIL"}
	// user should do it in the next 30 days
	lifespan := 60 * 60 * 24 * 30
	err = a.Client.ExecuteActionsEmail(ctx, a.GetAccessToken(), a.Realm, gocloak.ExecuteActionsEmail{
		UserID:   &userID,
		Actions:  &requiredActions,
		Lifespan: &lifespan,
	})
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (a *KeycloakActor) DeleteUser(
	ctx context.Context,
	userID string,
) error {
	return a.Client.DeleteUser(ctx, a.GetAccessToken(), a.Realm, userID)
}

func (a *KeycloakActor) getClientInternalIdentifierByClientID(
	ctx context.Context,
	clientID string,
) (string, error) {
	clients, err := a.Client.GetClients(ctx, a.GetAccessToken(), a.Realm, gocloak.GetClientsParams{
		ClientID: &clientID,
	})
	if err != nil {
		klog.Error("Unable to get client from keycloak", err)
		return "", err
	}
	if len(clients) != 1 {
		klog.Warningf("Client %s not found in Keycloak", clientID)
		return "", fmt.Errorf("404")
	}
	return *clients[0].ID, nil
}

func (a *KeycloakActor) GetRole(
	ctx context.Context,
	roleName string,
) (*gocloak.Role, error) {
	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		klog.Error("Unable to get client internal identifier from keycloak", err)
		return nil, err
	}

	role, err := a.Client.GetClientRole(
		ctx,
		a.GetAccessToken(),
		a.Realm,
		clientID,
		roleName,
	)

	if err != nil && err.Error() == "404 Not Found: Could not find role" {
		klog.Warningf("Role %s not found in Keycloak", roleName)
		return nil, fmt.Errorf("404")
	} else if err != nil {
		klog.Error("Unable to get roles from keycloak", err)
		return nil, err
	}

	return role, nil
}

func (a *KeycloakActor) CreateRole(
	ctx context.Context,
	roleName string,
	roleDescription string,
) (string, error) {
	role := gocloak.Role{
		Name:        &roleName,
		Description: &roleDescription,
	}

	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		klog.Error("Unable to get client internal identifier from keycloak", err)
		return "", err
	}

	createdRole, err := a.Client.CreateClientRole(
		ctx,
		a.GetAccessToken(),
		a.Realm,
		clientID,
		role,
	)
	if err != nil {
		klog.Error("Unable to create role in keycloak", err)
		return "", err
	}

	return createdRole, nil
}

func (a *KeycloakActor) DeleteRole(
	ctx context.Context,
	roleName string,
) error {
	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		klog.Error("Unable to get client internal identifier from keycloak", err)
		return err
	}

	err = a.Client.DeleteClientRole(
		ctx,
		a.GetAccessToken(),
		a.Realm,
		clientID,
		roleName,
	)

	if err != nil && err.Error() == "404 Not Found: Could not find role" {
		return nil
	} else if err != nil {
		klog.Error("Unable to delete role from keycloak", err)
		return err
	}

	return nil
}
