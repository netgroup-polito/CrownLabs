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

// Package workspace implements the workspace controller functionality.
package workspace

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func (r *Reconciler) handleTenantWorkspaceDeletion(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	log logr.Logger,
) error {
	var tenantsToUpdate v1alpha2.TenantList
	targetLabel := forge.GetWorkspaceTargetLabel(ws.Name)

	if err := r.List(ctx, &tenantsToUpdate, &client.HasLabels{targetLabel}); err != nil {
		log.Error(err, "Error when listing tenants subscribed to workspace upon deletion", "workspace", ws.Name)
		return err
	}

	for i := range tenantsToUpdate.Items {
		tn := &tenantsToUpdate.Items[i]
		if err := utils.PatchObject(ctx, r.Client, tn, func(t *v1alpha2.Tenant) *v1alpha2.Tenant {
			removeWorkspaceFromTenant(&t.Spec.Workspaces, ws.Name)
			return t
		}); err != nil {
			log.Error(err, "Error when unsubscribing tenant from workspace", "tenant", tn.Name, "workspace", ws.Name)
			return err
		}
	}

	return nil
}

func removeWorkspaceFromTenant(
	workspaces *[]v1alpha2.TenantWorkspaceEntry,
	wsToRemove string,
) {
	idxToRemove := -1
	for i, wsData := range *workspaces {
		if wsData.Name == wsToRemove {
			idxToRemove = i
		}
	}
	if idxToRemove != -1 {
		*workspaces = append((*workspaces)[:idxToRemove], (*workspaces)[idxToRemove+1:]...) // Truncate slice.
	}
}
