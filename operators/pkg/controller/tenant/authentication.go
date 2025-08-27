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

// Package tenant contains functionality related to CrownLabs tenant management.
package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// CheckKeycloakUserVerified checks if the Tenant has already been created in Keycloak
// and its email address has already been verified.
// If it has not been created, it creates it.
// It returns true if the Tenant has confrmed his/her email, false otherwise.
func (r *Reconciler) CheckKeycloakUserVerified(
	ctx context.Context,
	log logr.Logger,
	tenant *v1alpha2.Tenant,
) (bool, error) {
	if !r.KeycloakActor.IsInitialized() {
		log.Info("Keycloak actor not initialized, skipping Keycloak status check")
		return true, nil
	}

	// Check if the tenant exists in Keycloak
	user, err := r.KeycloakActor.GetUser(ctx, tenant.Name)
	if err != nil {
		if err.Error() == fmt.Sprintf("%d", http.StatusNotFound) {
			log.Info("Tenant not found in Keycloak, creating it")

			// Create the tenant in Keycloak
			err = r.createTenantInKeycloak(ctx, log, tenant)
			if err != nil {
				log.Error(err, "Error creating tenant in Keycloak")
				return false, err
			}

			log.Info("Tenant created in Keycloak")

			// retrieve newly created user
			user, err = r.KeycloakActor.GetUser(ctx, tenant.Name)
			if err != nil {
				log.Error(err, "Error retrieving newly created tenant in Keycloak")
				return false, err
			}
		} else {
			log.Error(err, "Error checking Keycloak status")
			return false, err
		}
	} else if tenant.Status.Keycloak.UserCreated.Name != *user.ID {
		log.Info("Tenant exists in Keycloak but with a different ID, updating status", "id", *user.ID)
		// Update the tenant status in the cluster
		tenant.Status.Keycloak.UserCreated = v1alpha2.NameCreated{
			Name:    *user.ID,
			Created: true,
		}
	}

	if *user.EmailVerified != tenant.Status.Keycloak.UserConfirmed {
		log.Info("Email verification status updated in Keycloak", "verified", *user.EmailVerified)

		// Update the tenant status in the cluster
		tenant.Status.Keycloak.UserConfirmed = *user.EmailVerified
	}

	return *user.EmailVerified, nil
}

func (r *Reconciler) createTenantInKeycloak(
	ctx context.Context,
	log logr.Logger,
	tenant *v1alpha2.Tenant,
) error {
	if !r.KeycloakActor.IsInitialized() {
		log.Info("Keycloak actor not initialized, skipping Keycloak creation")
		return nil
	}

	// Create the tenant in Keycloak
	userID, err := r.KeycloakActor.CreateUser(
		ctx,
		tenant.Name,
		tenant.Spec.Email,
		tenant.Spec.FirstName,
		tenant.Spec.LastName,
	)
	if err != nil {
		log.Error(err, "Error creating tenant in Keycloak")
		return err
	}

	tenant.Status.Keycloak = v1alpha2.KeycloakStatus{
		UserCreated: v1alpha2.NameCreated{
			Name:    userID,
			Created: true,
		},
		UserConfirmed: false,
	}

	return nil
}

func (r *Reconciler) deleteTenantInKeycloak(
	ctx context.Context,
	log logr.Logger,
	tenant *v1alpha2.Tenant,
) error {
	if !r.KeycloakActor.IsInitialized() {
		log.Info("Keycloak actor not initialized, skipping Keycloak deletion")
		return nil
	}

	if !tenant.Status.Keycloak.UserCreated.Created {
		log.Info("Tenant not created in Keycloak, skipping deletion")
		return nil
	}

	// Delete the tenant in Keycloak
	if err := r.KeycloakActor.DeleteUser(ctx, tenant.Status.Keycloak.UserCreated.Name); err != nil {
		log.Error(err, "Error deleting tenant in Keycloak")
		return fmt.Errorf("error deleting tenant %s in Keycloak: %w", tenant.Name, err)
	}
	log.Info("Tenant deleted in Keycloak")

	// Reset the Keycloak status in the tenant
	tenant.Status.Keycloak = v1alpha2.KeycloakStatus{
		UserCreated: v1alpha2.NameCreated{
			Name:    "",
			Created: false,
		},
		UserConfirmed: false,
	}
	return nil
}

// KeycloakEventHandler handles Keycloak webhook events for tenant resources.
func (r *Reconciler) KeycloakEventHandler(
	log logr.Logger,
	hw http.ResponseWriter,
	hr *http.Request,
) {
	body, err := io.ReadAll(hr.Body)
	if err != nil {
		hw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer hr.Body.Close()

	username, err := extractUsernameFromKeycloakEvent(body)
	if err != nil {
		hw.WriteHeader(http.StatusBadRequest)
		return
	}

	if username == "" {
		log.Info("Received Keycloak event with empty username")
		hw.WriteHeader(http.StatusBadRequest)
		return
	}

	r.TriggerReconcileChannel <- event.GenericEvent{
		Object: &v1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name: username,
			},
		},
	}

	hw.WriteHeader(http.StatusOK)
}

func extractUsernameFromKeycloakEvent(
	body []byte,
) (string, error) {
	var baseEvent struct {
		Type string `json:"type"`
	}

	// Initial parse to determine the type of event
	if err := json.Unmarshal(body, &baseEvent); err != nil {
		return "", err
	}

	switch baseEvent.Type {
	case "access.CUSTOM_REQUIRED_ACTION":
		username, err := extractUsernameFromCustomRequiredActionEvent(body)
		if err != nil {
			return "", fmt.Errorf("error extracting username from custom required action event: %w", err)
		}
		return username, nil
	case "admin.USER-UPDATE":
		username, err := extractUsernameFromUserUpdateEvent(body)
		if err != nil {
			return "", fmt.Errorf("error extracting username from user update event: %w", err)
		}
		return username, nil
	default:
		return "", fmt.Errorf("unrecognized event type: %s", baseEvent.Type)
	}
}

func extractUsernameFromCustomRequiredActionEvent(
	body []byte,
) (string, error) {
	var authEvent struct {
		AuthDetails struct {
			Username string `json:"username"`
		} `json:"authDetails"`
	}

	if err := json.Unmarshal(body, &authEvent); err != nil {
		return "", fmt.Errorf("error parsing custom required action event: %w", err)
	}

	return authEvent.AuthDetails.Username, nil
}

func extractUsernameFromUserUpdateEvent(
	body []byte,
) (string, error) {
	var adminEvent struct {
		Representation string `json:"representation"`
	}

	if err := json.Unmarshal(body, &adminEvent); err != nil {
		return "", fmt.Errorf("error parsing user update event: %w", err)
	}

	// internal json parsing
	var userRepresentation struct {
		Username string `json:"username"`
	}

	if err := json.Unmarshal([]byte(adminEvent.Representation), &userRepresentation); err != nil {
		return "", fmt.Errorf("error parsing representation JSON: %w", err)
	}

	return userRepresentation.Username, nil
}
