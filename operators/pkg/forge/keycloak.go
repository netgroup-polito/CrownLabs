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

package forge

import (
	"fmt"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// WorkspaceRoleName generates the role name for a workspace.
func WorkspaceRoleName(
	wsName string,
	role v1alpha2.WorkspaceUserRole,
) string {
	return fmt.Sprintf("workspace-%s:%s", wsName, role)
}

// GetWorkspaceManagerRoleName returns the Keycloak role name for workspace managers.
func GetWorkspaceManagerRoleName(ws *v1alpha1.Workspace) string {
	return WorkspaceRoleName(ws.Name, v1alpha2.Manager)
}

// GetWorkspaceManagerRoleDescription returns the Keycloak role description for workspace managers.
func GetWorkspaceManagerRoleDescription(ws *v1alpha1.Workspace) string {
	return fmt.Sprintf("%s Manager Role", ws.Spec.PrettyName)
}

// GetWorkspaceUserRoleName returns the Keycloak role name for workspace users.
func GetWorkspaceUserRoleName(ws *v1alpha1.Workspace) string {
	return WorkspaceRoleName(ws.Name, v1alpha2.User)
}

// GetWorkspaceUserRoleDescription returns the Keycloak role description for workspace users.
func GetWorkspaceUserRoleDescription(ws *v1alpha1.Workspace) string {
	return fmt.Sprintf("%s User Role", ws.Spec.PrettyName)
}
