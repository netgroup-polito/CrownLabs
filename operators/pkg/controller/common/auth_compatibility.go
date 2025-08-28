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
	"net/http"
	"strings"
	"sync"
	"time"

	gk13 "github.com/Nerzal/gocloak/v13"
	"github.com/Nerzal/gocloak/v7"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
)

// KeycloakActorCompatibility contains the functionality to interact with Keycloak.
type KeycloakActorCompatibility struct {
	initialized    bool
	Client         gocloak.GoCloak
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

var actorCompatibility KeycloakActorCompatibility

// SetupKeycloakActorCompatibility creates and initializes a new KeycloakActor.
func SetupKeycloakActorCompatibility(
	ctx context.Context,
	url string,
	clientID string,
	clientSecret string,
	realm string,
	rolesClientID string,
	log logr.Logger,
) error {
	if actorCompatibility.IsInitialized() {
		return nil
	}

	actorIface = &actorCompatibility

	if actorCompatibility.Client == nil {
		actorCompatibility.Client = gocloak.NewClient(url)
	}

	// login to keycloak
	_, err := actorCompatibility.Client.LoginClient(ctx, clientID, clientSecret, realm)
	if err != nil {
		log.Error(err, "Unable to login on keycloak")
		return err
	}

	actorCompatibility.Realm = realm
	actorCompatibility.credentials.ClientID = clientID
	actorCompatibility.credentials.ClientSecret = clientSecret
	actorCompatibility.RolesClientID = rolesClientID
	actorCompatibility.clientIDCache = make(map[string]string)
	actorCompatibility.initialized = true
	return nil
}

// IsInitialized checks if the KeycloakActor has been initialized.
func (a *KeycloakActorCompatibility) IsInitialized() bool {
	return a.initialized
}

// Reset clears the KeycloakActor's token and cached data.
func (a *KeycloakActorCompatibility) Reset(log logr.Logger) {
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
func (a *KeycloakActorCompatibility) GetAccessToken(ctx context.Context) string {
	log := klog.FromContext(ctx)

	a.tokenMutex.RLock()
	defer a.tokenMutex.RUnlock()

	now := time.Now().Unix()
	// renew rhe token if it is nil or about to expire
	if a.token == nil || now >= (a.tokenExpiresAt-tokenRefreshBuffer) {
		log.Info("Keycloak token is not present or about to expire, refreshing it")
		token, err := a.Client.LoginClient(ctx, a.credentials.ClientID, a.credentials.ClientSecret, a.Realm)
		if err != nil {
			log.Error(err, "Unable to refresh keycloak token")
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
func (a *KeycloakActorCompatibility) GetUser(
	ctx context.Context,
	username string,
) (*gk13.User, error) {
	log := klog.FromContext(ctx)

	users, err := a.Client.GetUsers(ctx, a.GetAccessToken(ctx), a.Realm, gocloak.GetUsersParams{
		Username: &username,
	})
	if err != nil {
		log.Error(err, "Unable to get user from keycloak")
		return nil, err
	}

	// handle the case where it is a substring of another
	if len(users) == 0 {
		log.Info("User not found in Keycloak", "username", username)
		return nil, fmt.Errorf("%d", http.StatusNotFound)
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
			log.Info("No match for user in Keycloak", "username", username)
			return nil, fmt.Errorf("%d", http.StatusNotFound)
		}
	} else {
		user = users[0]
	}

	return a.convertUserV7to13(user), nil
}

// CreateUser creates a user in Keycloak.
func (a *KeycloakActorCompatibility) CreateUser(
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

	userID, err := a.Client.CreateUser(ctx, a.GetAccessToken(ctx), a.Realm, user)
	if err != nil {
		return "", err
	}

	// Set the required actions for the user
	requiredActions := []string{"UPDATE_PASSWORD", "VERIFY_EMAIL"}
	// user should do it in the next 30 days
	lifespan := 60 * 60 * 24 * 30
	err = a.Client.ExecuteActionsEmail(ctx, a.GetAccessToken(ctx), a.Realm, gocloak.ExecuteActionsEmail{
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
func (a *KeycloakActorCompatibility) DeleteUser(
	ctx context.Context,
	userID string,
) error {
	return a.Client.DeleteUser(ctx, a.GetAccessToken(ctx), a.Realm, userID)
}

//nolint:dupl // portability between kc versions
func (a *KeycloakActorCompatibility) getClientInternalIdentifierByClientID(
	ctx context.Context,
	clientID string,
) (string, error) {
	log := klog.FromContext(ctx)

	// Check if the client ID is already in cache
	a.cacheMutex.RLock()
	if internalID, exists := a.clientIDCache[clientID]; exists {
		a.cacheMutex.RUnlock()
		return internalID, nil
	}
	a.cacheMutex.RUnlock()

	// If not in cache, fetch the client from Keycloak
	clients, err := a.Client.GetClients(ctx, a.GetAccessToken(ctx), a.Realm, gocloak.GetClientsParams{
		ClientID: &clientID,
	})
	if err != nil {
		log.Error(err, "Unable to get client from keycloak")
		return "", err
	}
	if len(clients) != 1 {
		log.Info("Client not found in Keycloak", "clientID", clientID)
		return "", fmt.Errorf("%d", http.StatusNotFound)
	}

	// Store the client ID in the cache
	a.cacheMutex.Lock()
	a.clientIDCache[clientID] = *clients[0].ID
	a.cacheMutex.Unlock()

	return *clients[0].ID, nil
}

// GetRole gets a role from Keycloak.
func (a *KeycloakActorCompatibility) GetRole(
	ctx context.Context,
	roleName string,
) (*gk13.Role, error) {
	log := klog.FromContext(ctx)

	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		log.Error(err, "Unable to get client internal identifier from keycloak")
		return nil, err
	}

	role, err := a.Client.GetClientRole(
		ctx,
		a.GetAccessToken(ctx),
		a.Realm,
		clientID,
		roleName,
	)

	if err != nil && err.Error() == "404 Not Found: Could not find role" {
		log.Info("Role not found in Keycloak", "roleName", roleName)
		return nil, fmt.Errorf("%d", http.StatusNotFound)
	} else if err != nil {
		log.Error(err, "Unable to get roles from keycloak")
		return nil, err
	}

	return a.convertRoleV7to13(role), nil
}

// CreateRole creates a new role in Keycloak.
func (a *KeycloakActorCompatibility) CreateRole(
	ctx context.Context,
	roleName string,
	roleDescription string,
) (string, error) {
	log := klog.FromContext(ctx)

	role := gocloak.Role{
		Name:        &roleName,
		Description: &roleDescription,
	}

	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		log.Error(err, "Unable to get client internal identifier from keycloak")
		return "", err
	}

	createdRole, err := a.Client.CreateClientRole(
		ctx,
		a.GetAccessToken(ctx),
		a.Realm,
		clientID,
		role,
	)
	if err != nil {
		log.Error(err, "Unable to create role in keycloak")
		return "", err
	}

	return createdRole, nil
}

// DeleteRole removes a role from Keycloak.
func (a *KeycloakActorCompatibility) DeleteRole(
	ctx context.Context,
	roleName string,
) error {
	log := klog.FromContext(ctx)

	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		log.Error(err, "Unable to get client internal identifier from keycloak")
		return err
	}

	err = a.Client.DeleteClientRole(
		ctx,
		a.GetAccessToken(ctx),
		a.Realm,
		clientID,
		roleName,
	)

	if err != nil && err.Error() == "404 Not Found: Could not find role" {
		return nil
	} else if err != nil {
		log.Error(err, "Unable to delete role from keycloak")
		return err
	}

	return nil
}

// GetUserRoles gets the roles assigned to a user in Keycloak.
func (a *KeycloakActorCompatibility) GetUserRoles(
	ctx context.Context,
	userID string,
) ([]*gk13.Role, error) {
	log := klog.FromContext(ctx)

	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		log.Error(err, "Unable to get client internal identifier from keycloak")
		return nil, err
	}

	roles, err := a.Client.GetClientRolesByUserID(
		ctx,
		a.GetAccessToken(ctx),
		a.Realm,
		clientID,
		userID,
	)

	if err != nil && strings.Contains(err.Error(), "404 Not Found") {
		log.Info("User not found in Keycloak", "userID", userID)
		return nil, fmt.Errorf("%d", http.StatusNotFound)
	} else if err != nil {
		log.Error(err, "Unable to get user roles from keycloak")
		return nil, err
	}

	return a.convertRolesV7to13(roles), nil
}

// AddUserToRoles adds a user to the specified roles in Keycloak.
func (a *KeycloakActorCompatibility) AddUserToRoles(
	ctx context.Context,
	userID string,
	roles []*gk13.Role,
) error {
	log := klog.FromContext(ctx)

	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		log.Error(err, "Unable to get client internal identifier from keycloak")
		return err
	}

	// Convert []*gocloak.Role to []gocloak.Role
	rolesVal := make([]gocloak.Role, len(roles))
	for i, r := range roles {
		if r != nil {
			rolesVal[i] = *a.convertRoleV13to7(r)
		}
	}

	err = a.Client.AddClientRoleToUser(
		ctx,
		a.GetAccessToken(ctx),
		a.Realm,
		clientID,
		userID,
		rolesVal,
	)

	if err != nil {
		log.Error(err, "Unable to add user to role in keycloak")
		return err
	}

	return nil
}

// RemoveUserFromRoles removes a user from the specified roles in Keycloak.
func (a *KeycloakActorCompatibility) RemoveUserFromRoles(
	ctx context.Context,
	userID string,
	roles []*gk13.Role,
) error {
	log := klog.FromContext(ctx)

	clientID, err := a.getClientInternalIdentifierByClientID(ctx, a.RolesClientID)
	if err != nil {
		log.Error(err, "Unable to get client internal identifier from keycloak")
		return err
	}

	// Convert []*gocloak.Role to []gocloak.Role
	rolesVal := make([]gocloak.Role, len(roles))
	for i, r := range roles {
		if r != nil {
			rolesVal[i] = *a.convertRoleV13to7(r)
		}
	}

	err = a.Client.DeleteClientRoleFromUser(
		ctx,
		a.GetAccessToken(ctx),
		a.Realm,
		clientID,
		userID,
		rolesVal,
	)

	if err != nil {
		log.Error(err, "Unable to remove user from role in keycloak")
		return err
	}

	return nil
}

func (a *KeycloakActorCompatibility) convertUserV7to13(v7 *gocloak.User) *gk13.User {
	return &gk13.User{
		ID:            v7.ID,
		Username:      v7.Username,
		Email:         v7.Email,
		FirstName:     v7.FirstName,
		LastName:      v7.LastName,
		Enabled:       v7.Enabled,
		EmailVerified: v7.EmailVerified,
	}
}

func (a *KeycloakActorCompatibility) convertRoleV7to13(v7 *gocloak.Role) *gk13.Role {
	return &gk13.Role{
		ID:          v7.ID,
		Name:        v7.Name,
		Description: v7.Description,
	}
}

func (a *KeycloakActorCompatibility) convertRolesV7to13(v7 []*gocloak.Role) []*gk13.Role {
	roles := make([]*gk13.Role, len(v7))
	for i, role := range v7 {
		roles[i] = a.convertRoleV7to13(role)
	}
	return roles
}

func (a *KeycloakActorCompatibility) convertRoleV13to7(v13 *gk13.Role) *gocloak.Role {
	return &gocloak.Role{
		ID:          v13.ID,
		Name:        v13.Name,
		Description: v13.Description,
	}
}
