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

// Package workspace contains functionality related to CrownLabs workspace management.
package workspace

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func (r *Reconciler) createKeycloakRoles(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	log logr.Logger,
) error {
	if !r.KeycloakActor.IsInitialized() {
		ws.Status.Subscriptions["keycloak"] = v1alpha2.SubscrFailed
		log.Info("Keycloak actor is not initialized, skipping role creation for workspace", "workspace", ws.Name)
		return nil
	}

	// Create manager role
	managerRoleName := forge.GetWorkspaceManagerRoleName(ws)
	managerRoleDesc := forge.GetWorkspaceManagerRoleDescription(ws)
	if err := r.createKeycloakRole(ctx, ws, managerRoleName, managerRoleDesc, log); err != nil {
		ws.Status.Subscriptions["keycloak"] = v1alpha2.SubscrFailed
		log.Error(err, "Error when creating Keycloak manager role", "role", managerRoleName, "workspace", ws.Name)
		return err
	}

	// Create user role
	userRoleName := forge.GetWorkspaceUserRoleName(ws)
	userRoleDesc := forge.GetWorkspaceUserRoleDescription(ws)
	if err := r.createKeycloakRole(ctx, ws, userRoleName, userRoleDesc, log); err != nil {
		ws.Status.Subscriptions["keycloak"] = v1alpha2.SubscrFailed
		log.Error(err, "Error when creating Keycloak user role", "role", userRoleName, "workspace", ws.Name)
		return err
	}

	ws.Status.Subscriptions["keycloak"] = v1alpha2.SubscrOk

	return nil
}

func (r *Reconciler) createKeycloakRole(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	roleName string,
	roleDescription string,
	log logr.Logger,
) error {
	if !r.KeycloakActor.IsInitialized() {
		log.Info("Keycloak actor is not initialized, skipping role creation for workspace", "workspace", ws.Name)
		return nil
	}

	if role, err := r.KeycloakActor.GetRole(ctx, roleName); err != nil && err.Error() != fmt.Sprintf("%d", http.StatusNotFound) {
		log.Error(err, "Error when getting Keycloak role", "role", roleName, "workspace", ws.Name)
		return err
	} else if role != nil {
		log.Info("Keycloak role already exists, skipping creation", "role", roleName, "workspace", ws.Name)
		return nil
	}

	if _, err := r.KeycloakActor.CreateRole(ctx, roleName, roleDescription); err != nil {
		log.Error(err, "Error when creating Keycloak role", "role", roleName, "workspace", ws.Name)
		return err
	}

	log.Info("Successfully created Keycloak role", "role", roleName, "workspace", ws.Name)
	return nil
}

func (r *Reconciler) deleteKeycloakRoles(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	log logr.Logger,
) error {
	if !r.KeycloakActor.IsInitialized() {
		log.Info("Keycloak actor is not initialized, skipping role deletion for workspace", "workspace", ws.Name)
		return nil
	}

	// Delete manager role
	managerRoleName := forge.GetWorkspaceManagerRoleName(ws)
	if err := r.deleteKeycloakRole(ctx, ws, managerRoleName, log); err != nil {
		log.Error(err, "Error when deleting Keycloak manager role", "role", managerRoleName, "workspace", ws.Name)
		return err
	}

	// Delete user role
	userRoleName := forge.GetWorkspaceUserRoleName(ws)
	if err := r.deleteKeycloakRole(ctx, ws, userRoleName, log); err != nil {
		log.Error(err, "Error when deleting Keycloak user role", "role", userRoleName, "workspace", ws.Name)
		return err
	}

	return nil
}

func (r *Reconciler) deleteKeycloakRole(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	roleName string,
	log logr.Logger,
) error {
	if !r.KeycloakActor.IsInitialized() {
		log.Info("Keycloak actor is not initialized, skipping role deletion for workspace", "workspace", ws.Name)
		return nil
	}

	if err := r.KeycloakActor.DeleteRole(ctx, roleName); err != nil {
		log.Error(err, "Error when deleting Keycloak role", "role", roleName, "workspace", ws.Name)
		return err
	}

	log.Info("Successfully deleted Keycloak role", "role", roleName, "workspace", ws.Name)
	return nil
}
