// Copyright 2020-2026 Politecnico di Torino
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
	"slices"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// EnrolledWorkspaces returns the workspaces the tenant is effectively enrolled in, leaving out the
// entries whose enrollment is not in force: candidates are still waiting for a manager to approve
// them, and failing workspaces could not be processed by the tenant controller.
func EnrolledWorkspaces(tn *clv1alpha2.Tenant) []clv1alpha2.TenantWorkspaceEntry {
	enrolled := make([]clv1alpha2.TenantWorkspaceEntry, 0, len(tn.Spec.Workspaces))

	for _, ws := range tn.Spec.Workspaces {
		// skip workspaces in Candidate status
		if ws.Role == clv1alpha2.Candidate {
			continue
		}

		// skip failing workspaces
		if slices.Contains(tn.Status.FailingWorkspaces, ws.Name) {
			continue
		}

		enrolled = append(enrolled, ws)
	}

	return enrolled
}

// TenantCanReadNamespace tells whether the tenant is entitled to consume the volumes stored in the
// given namespace: its own namespace, the namespaces of the workspaces it is enrolled in, and the
// public snapshot catalog.
func TenantCanReadNamespace(tn *clv1alpha2.Tenant, namespace, publicSnapshotNamespace string) bool {
	if namespace == "" {
		return false
	}

	if namespace == GetTenantNamespaceName(tn) {
		return true
	}

	if publicSnapshotNamespace != "" && namespace == publicSnapshotNamespace {
		return true
	}

	for _, ws := range EnrolledWorkspaces(tn) {
		if namespace == WorkspaceNamespaceName(ws.Name) {
			return true
		}
	}

	return false
}
