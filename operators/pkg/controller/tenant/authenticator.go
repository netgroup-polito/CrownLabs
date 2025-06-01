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

// Package tenant_controller groups the functionalities related to the Tenant controller.
package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// CheckKeycloakUserVerified checks if the Tenant has already been created in Keycloak
// and its email address has already been verified.
// If it has not been created, it creates it.
// It returns true if the Tenant has confrmed his/her email, false otherwise.
func (r *TenantReconciler) CheckKeycloakUserVerified(
	ctx context.Context,
	tenant *crownlabsv1alpha2.Tenant,
) (bool, error) {
	actor := utils.GetKeycloakActor()
	if !actor.IsInitialized() {
		klog.Warningf("Keycloak actor not initialized, skipping Keycloak status check for tenant %s", tenant.Name)
		return true, nil
	}

	// Check if the tenant exists in Keycloak
	user, err := actor.GetUser(ctx, tenant.Name)
	if err != nil {
		if err.Error() == "404" {
			klog.Infof("Tenant %s not found in Keycloak, creating it", tenant.Name)

			// Create the tenant in Keycloak
			err = r.createTenantInKeycloak(ctx, tenant)
			if err != nil {
				klog.Errorf("Error creating tenant %s in Keycloak: %v", tenant.Name, err)
				return false, err
			}

			klog.Infof("Tenant %s created in Keycloak", tenant.Name)

			// retrive newly created user
			user, err = actor.GetUser(ctx, tenant.Name)
			if err != nil {
				klog.Errorf("Error retrieving newly created tenant %s in Keycloak: %v", tenant.Name, err)
				return false, err
			}
		} else {
			klog.Errorf("Error checking Keycloak status for tenant %s: %v", tenant.Name, err)
			return false, err
		}
	} else if tenant.Status.Keycloak.UserCreated.Name != *user.ID {
		klog.Infof("Tenant %s exists in Keycloak but with a different ID (%s), updating status", tenant.Name, *user.ID)
		// Update the tenant status in the cluster
		tenant.Status.Keycloak.UserCreated = crownlabsv1alpha2.NameCreated{
			Name:    *user.ID,
			Created: true,
		}
	}

	if *user.EmailVerified != tenant.Status.Keycloak.UserConfirmed {
		klog.Infof("Tenant %s email verification status in Keycloak: %v", tenant.Name, *user.EmailVerified)

		// Update the tenant status in the cluster
		tenant.Status.Keycloak.UserConfirmed = *user.EmailVerified
	}

	return *user.EmailVerified, nil
}

func (r *TenantReconciler) createTenantInKeycloak(
	ctx context.Context,
	tenant *crownlabsv1alpha2.Tenant,
) error {
	actor := utils.GetKeycloakActor()
	if !actor.IsInitialized() {
		klog.Warningf("Keycloak actor not initialized, skipping Keycloak creation for tenant %s", tenant.Name)
		return nil
	}

	// Create the tenant in Keycloak
	userId, err := actor.CreateUser(
		ctx,
		tenant.Name,
		tenant.Spec.Email,
		tenant.Spec.FirstName,
		tenant.Spec.LastName,
	)
	if err != nil {
		klog.Errorf("Error creating tenant %s in Keycloak: %v", tenant.Name, err)
		return err
	}

	tenant.Status.Keycloak = crownlabsv1alpha2.KeycloakStatus{
		UserCreated: crownlabsv1alpha2.NameCreated{
			Name:    userId,
			Created: true,
		},
		UserConfirmed: false,
	}

	return nil
}

func (r *TenantReconciler) KeycloakEventHandler(
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

	r.TriggerReconcileChannel <- event.GenericEvent{
		Object: &crownlabsv1alpha2.Tenant{
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
			return "", fmt.Errorf("error extracting username from custom required action event: %v", err)
		}
		return username, nil
	case "admin.USER-UPDATE":
		username, err := extractUsernameFromUserUpdateEvent(body)
		if err != nil {
			return "", fmt.Errorf("error extracting username from user update event: %v", err)
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
		return "", fmt.Errorf("error parsing custom required action event: %v", err)
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
		return "", fmt.Errorf("error parsing user update event: %v", err)
	}

	// internal json parsing
	var userRepresentation struct {
		Username string `json:"username"`
	}

	if err := json.Unmarshal([]byte(adminEvent.Representation), &userRepresentation); err != nil {
		return "", fmt.Errorf("error parsing representation JSON: %v", err)
	}

	return userRepresentation.Username, nil
}
