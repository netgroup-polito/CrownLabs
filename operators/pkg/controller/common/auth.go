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

// Package common provides shared functionality for the CrownLabs operators.
package common

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
)

// KeycloakActor contains the functionality to interact with Keycloak.
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
	RolesClientID string            // The client ID of the client in which the roles are defined.
	clientIDCache map[string]string // Cache for client IDs to avoid multiple requests to Keycloak
	cacheMutex    sync.RWMutex
}

const tokenRefreshBuffer = 30 // the token is considered about to expire if it has less than this many seconds left

var actor KeycloakActor
var actorIface KeycloakActorIface = &actor

// SetupKeycloakActor creates and initializes a new KeycloakActor.
func SetupKeycloakActor(
	url string,
	clientID string,
	clientSecret string,
	realm string,
	rolesClientID string,
	log logr.Logger,
) error {
	if actor.IsInitialized() {
		return nil
	}

	if actor.Client == nil {
		actor.Client = gocloak.NewClient(url)
	}

	// login to keycloak
	_, err := actor.Client.LoginClient(context.Background(), clientID, clientSecret, realm)
	if err != nil {
		log.Error(err, "Unable to login on keycloak: ")
		return err
	}

	actor.Realm = realm
	actor.credentials.ClientID = clientID
	actor.credentials.ClientSecret = clientSecret
	actor.RolesClientID = rolesClientID
	actor.clientIDCache = make(map[string]string)
	actor.initialized = true
	return nil
}

// GetKeycloakActor returns the KcActor currently used.
func GetKeycloakActor() KeycloakActorIface {
	return actorIface
}

// IsInitialized checks if the KeycloakActor has been initialized.
func (a *KeycloakActor) IsInitialized() bool {
	return a.initialized
}

// Reset clears the KeycloakActor's token and cached data.
func (a *KeycloakActor) Reset(log logr.Logger) {
	a.tokenMutex.Lock()
	a.cacheMutex.Lock()
	defer a.tokenMutex.Unlock()
	defer a.cacheMutex.Unlock()

	a.initialized = false
	a.Client = nil
	a.Realm = ""
	a.token = nil
	a.tokenExpiresAt = 0
	a.credentials.ClientID = ""
	a.credentials.ClientSecret = ""
	a.clientIDCache = nil
	log.Info("Keycloak actor has been reset")
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
			a.tokenExpiresAt = 0
			return ""
		}

		a.token = token
		// set the token expiration time
		a.tokenExpiresAt = now + int64(token.ExpiresIn)
		return token.AccessToken
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

	// handle the case where it is a substring of another
	if len(users) == 0 || users[0] == nil {
		klog.Warningf("User %s not found in Keycloak", username)
		return nil, fmt.Errorf("404")
	}

	// If there are multiple users, look for the one that exactly matches the searched username
	var user *gocloak.User
	if *users[0].Username != username {
		for _, u := range users {
			if u.Username != nil && *u.Username == username {
				user = u
				break
			}
		}
		if user == nil {
			klog.Warningf("No match for user %s in Keycloak", username)
			return nil, fmt.Errorf("404")
		}
	} else {
		user = users[0]
	}

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

// DeleteUser removes a user from Keycloak.
func (a *KeycloakActor) DeleteUser(
	ctx context.Context,
	userID string,
) error {
	return a.Client.DeleteUser(ctx, a.GetAccessToken(), a.Realm, userID)
}

//nolint:dupl // portability between kc versions: this directive should be removed when the compatibility layer is no longer needed
func (a *KeycloakActor) getClientInternalIdentifierByClientID(
	ctx context.Context,
	clientID string,
) (string, error) {
	// Check if the client ID is already in cache
	a.cacheMutex.RLock()
	if internalID, exists := a.clientIDCache[clientID]; exists {
		a.cacheMutex.RUnlock()
		return internalID, nil
	}
	a.cacheMutex.RUnlock()

	// If not in cache, fetch the client from Keycloak
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

	// Store the client ID in the cache
	a.cacheMutex.Lock()
	a.clientIDCache[clientID] = *clients[0].ID
	a.cacheMutex.Unlock()

	return *clients[0].ID, nil
}

// GetRole gets a role from Keycloak.
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

// CreateRole creates a new role in Keycloak.
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

// DeleteRole removes a role from Keycloak.
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

// GetUserRoles gets the roles assigned to a user in Keycloak.
func (a *KeycloakActor) GetUserRoles(
	ctx context.Context,
	userID string,
) ([]*gocloak.Role, error) {
	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		klog.Error("Unable to get client internal identifier from keycloak", err)
		return nil, err
	}

	roles, err := a.Client.GetClientRolesByUserID(
		ctx,
		a.GetAccessToken(),
		a.Realm,
		clientID,
		userID,
	)

	if err != nil && strings.Contains(err.Error(), "404 Not Found") {
		klog.Warningf("User %s not found in Keycloak", userID)
		return nil, fmt.Errorf("404")
	} else if err != nil {
		klog.Error("Unable to get user roles from keycloak", err)
		return nil, err
	}

	return roles, nil
}

// AddUserToRoles adds a user to the specified roles in Keycloak.
func (a *KeycloakActor) AddUserToRoles(
	ctx context.Context,
	userID string,
	roles []*gocloak.Role,
) error {
	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		klog.Error("Unable to get client internal identifier from keycloak", err)
		return err
	}

	// Convert []*gocloak.Role to []gocloak.Role
	rolesVal := make([]gocloak.Role, len(roles))
	for i, r := range roles {
		if r != nil {
			rolesVal[i] = *r
		}
	}

	err = a.Client.AddClientRolesToUser(
		ctx,
		a.GetAccessToken(),
		a.Realm,
		clientID,
		userID,
		rolesVal,
	)

	if err != nil {
		klog.Error("Unable to add user to role in keycloak", err)
		return err
	}

	return nil
}

// RemoveUserFromRoles removes a user from the specified roles in Keycloak.
func (a *KeycloakActor) RemoveUserFromRoles(
	ctx context.Context,
	userID string,
	roles []*gocloak.Role,
) error {
	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		klog.Error("Unable to get client internal identifier from keycloak", err)
		return err
	}

	// Convert []*gocloak.Role to []gocloak.Role
	rolesVal := make([]gocloak.Role, len(roles))
	for i, r := range roles {
		if r != nil {
			rolesVal[i] = *r
		}
	}

	err = a.Client.DeleteClientRolesFromUser(
		ctx,
		a.GetAccessToken(),
		a.Realm,
		clientID,
		userID,
		rolesVal,
	)

	if err != nil {
		klog.Error("Unable to remove user from role in keycloak", err)
		return err
	}

	return nil
}
