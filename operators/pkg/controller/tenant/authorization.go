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

// Package tenant implements the tenant controller functionality.
package tenant

import (
	"context"
	"slices"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-logr/logr"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func (r *Reconciler) syncWorkspacesAuthorizationRoles(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	if !r.KeycloakActor.IsInitialized() {
		log.Info("Keycloak actor is not initialized, skipping workspace authorization roles update")
		return nil
	}

	if !tn.Status.Keycloak.UserCreated.Created {
		log.Info("Keycloak user has not yet been created, skipping workspace authorization roles update")
		return nil
	}

	wantedRoles := r.obtainWantedRoles(tn)

	currentRoles, err := r.obtainCurrentRoles(ctx, log, tn)
	if err != nil {
		return err
	}

	// generate the roles wanted by the tenant and not present in Keycloak
	addRoles, err := r.getRolesToAdd(ctx, wantedRoles, currentRoles)
	if err != nil {
		log.Error(err, "Error obtaining roles to add for tenant", "tenant", tn.Name)
		return err
	}

	if len(addRoles) > 0 {
		// add the missing roles to Keycloak
		if err := r.KeycloakActor.AddUserToRoles(ctx, tn.Status.Keycloak.UserCreated.Name, addRoles); err != nil {
			log.Error(err, "Error adding roles to Keycloak for tenant", "tenant", tn.Name)
			return err
		}
	}

	// generate the roles present in Keycloak but not wanted by the tenant
	deleteRoles := r.getRolesToDelete(wantedRoles, currentRoles)

	if len(deleteRoles) > 0 {
		// remove the unwanted roles from Keycloak
		if err := r.KeycloakActor.RemoveUserFromRoles(ctx, tn.Status.Keycloak.UserCreated.Name, deleteRoles); err != nil {
			log.Error(err, "Error removing roles from Keycloak for tenant", "tenant", tn.Name)
			return err
		}
	}

	return nil
}

func (r *Reconciler) obtainWantedRoles(
	tn *v1alpha2.Tenant,
) []string {
	wantedRoles := make([]string, 0, len(tn.Spec.Workspaces))

	for _, ws := range r.getEnrolledWorkspaces(tn) {
		wantedRoles = append(wantedRoles, workspaceRoleName(ws))
	}

	return wantedRoles
}

func (r *Reconciler) obtainCurrentRoles(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) ([]*gocloak.Role, error) {
	currentRoles, err := r.KeycloakActor.GetUserRoles(ctx, tn.Status.Keycloak.UserCreated.Name)
	if err != nil {
		log.Error(err, "Error getting roles from Keycloak")
		return nil, err
	}

	// filter roles to only those related to workspaces
	filteredRoles := make([]*gocloak.Role, 0)
	for _, role := range currentRoles {
		if strings.HasPrefix(*role.Name, "workspace-") {
			filteredRoles = append(filteredRoles, role)
		}
	}

	return filteredRoles, nil
}

func (r *Reconciler) getRolesToAdd(
	ctx context.Context,
	wantedRoles []string,
	currentRoles []*gocloak.Role,
) ([]*gocloak.Role, error) {
	rolesToAdd := make([]string, 0)

	for _, wantedRole := range wantedRoles {
		if !slices.ContainsFunc(currentRoles, func(role *gocloak.Role) bool {
			return *role.Name == wantedRole
		}) {
			rolesToAdd = append(rolesToAdd, wantedRole)
		}
	}

	return r.convertRoleNamesToRoles(ctx, rolesToAdd)
}

func (r *Reconciler) convertRoleNamesToRoles(
	ctx context.Context,
	roles []string,
) ([]*gocloak.Role, error) {
	gocloakRoles := make([]*gocloak.Role, len(roles))
	for i, roleName := range roles {
		role, err := r.KeycloakActor.GetRole(ctx, roleName)
		if err != nil {
			return nil, err
		}
		gocloakRoles[i] = role
	}
	return gocloakRoles, nil
}

func (r *Reconciler) getRolesToDelete(
	wantedRoles []string,
	currentRoles []*gocloak.Role,
) []*gocloak.Role {
	rolesToDelete := make([]*gocloak.Role, 0)

	for _, currentRole := range currentRoles {
		if !slices.Contains(wantedRoles, *currentRole.Name) {
			// if the current role is not in the wanted roles, we add it to the list to delete
			rolesToDelete = append(rolesToDelete, currentRole)
		}
	}

	return rolesToDelete
}

func workspaceRoleName(data v1alpha2.TenantWorkspaceEntry) string {
	return forge.WorkspaceRoleName(data.Name, data.Role)
}
