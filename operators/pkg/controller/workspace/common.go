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
	"fmt"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

func (r *WorkspaceReconciler) updateWsResourceCommonLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 1)
	}
	labels[r.TargetLabel.GetKey()] = r.TargetLabel.GetValue()
	labels["crownlabs.polito.it/managed-by"] = "workspace"

	return labels
}

func workspaceRoleName(
	ws *v1alpha1.Workspace,
	role v1alpha2.WorkspaceUserRole,
) string {
	return fmt.Sprintf("workspace-%s:%s", ws.Name, role)
}
