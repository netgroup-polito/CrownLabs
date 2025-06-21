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
package workspace

import (
	"context"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"k8s.io/klog/v2"
)

func getWorkspaceRoles(
	ws *v1alpha1.Workspace,
) map[string]string {
	return map[string]string{
		workspaceRoleName(ws, v1alpha2.Manager): ws.Spec.PrettyName + " Manager Role",
		workspaceRoleName(ws, v1alpha2.User):    ws.Spec.PrettyName + " User Role",
	}
}

func (r *WorkspaceReconciler) createKeycloakRoles(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	if !r.KeycloakActor.IsInitialized() {
		ws.Status.Subscriptions["keycloak"] = v1alpha2.SubscrFailed
		klog.Warningf("Keycloak actor is not initialized, skipping role creation for workspace %s", ws.Name)
		return nil
	}

	for roleName, roleDescription := range getWorkspaceRoles(ws) {
		if err := r.createKeycloakRole(ctx, ws, roleName, roleDescription); err != nil {
			ws.Status.Subscriptions["keycloak"] = v1alpha2.SubscrFailed
			klog.Errorf("Error when creating Keycloak role %s for workspace %s -> %s", roleName, ws.Name, err)
			return err
		}
	}

	ws.Status.Subscriptions["keycloak"] = v1alpha2.SubscrOk

	return nil
}

func (r *WorkspaceReconciler) createKeycloakRole(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	roleName string,
	roleDescription string,
) error {
	if !r.KeycloakActor.IsInitialized() {
		klog.Warningf("Keycloak actor is not initialized, skipping role creation for workspace %s", ws.Name)
		return nil
	}

	if role, err := r.KeycloakActor.GetRole(ctx, roleName); err != nil && err.Error() != "404" {
		klog.Errorf("Error when getting Keycloak role %s for workspace %s -> %s", roleName, ws.Name, err)
		return err
	} else if role != nil {
		klog.Infof("Keycloak role %s for workspace %s already exists, skipping creation", roleName, ws.Name)
		return nil
	}

	if _, err := r.KeycloakActor.CreateRole(ctx, roleName, roleDescription); err != nil {
		klog.Errorf("Error when creating Keycloak role %s for workspace %s -> %s", roleName, ws.Name, err)
		return err
	}

	klog.Infof("Successfully created Keycloak role %s for workspace %s", roleName, ws.Name)
	return nil
}

func (r *WorkspaceReconciler) deleteKeycloakRoles(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	if !r.KeycloakActor.IsInitialized() {
		klog.Warningf("Keycloak actor is not initialized, skipping role deletion for workspace %s", ws.Name)
		return nil
	}

	for roleName := range getWorkspaceRoles(ws) {
		if err := r.deleteKeycloakRole(ctx, ws, roleName); err != nil {
			klog.Errorf("Error when deleting Keycloak role %s for workspace %s -> %s", roleName, ws.Name, err)
			return err
		}
	}

	return nil
}

func (r *WorkspaceReconciler) deleteKeycloakRole(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	roleName string,
) error {
	if !r.KeycloakActor.IsInitialized() {
		klog.Warningf("Keycloak actor is not initialized, skipping role deletion for workspace %s", ws.Name)
		return nil
	}

	if err := r.KeycloakActor.DeleteRole(ctx, roleName); err != nil {
		klog.Errorf("Error when deleting Keycloak role %s for workspace %s -> %s", roleName, ws.Name, err)
		return err
	}

	klog.Infof("Successfully deleted Keycloak role %s for workspace %s", roleName, ws.Name)
	return nil
}
