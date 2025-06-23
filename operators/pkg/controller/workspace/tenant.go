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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *WorkspaceReconciler) handleTenantWorkspaceDeletion(
	ctx context.Context,
	log logr.Logger,
	ws *v1alpha1.Workspace,
) error {
	var tenantsToUpdate v1alpha2.TenantList
	targetLabel := fmt.Sprintf("%s%s", v1alpha2.WorkspaceLabelPrefix, ws.Name)

	if err := r.List(ctx, &tenantsToUpdate, &client.HasLabels{targetLabel}); err != nil {
		klog.Errorf("Error when listing tenants subscribed to workspace %s upon deletion -> %s", ws.Name, err)
		return err
	}

	for _, tn := range tenantsToUpdate.Items {
		removeWorkspaceFromTenant(&tn.Spec.Workspaces, ws.Name)
		if err := r.Update(ctx, &tn); err != nil {
			klog.Errorf("Error when unsubscribing tenant %s from workspace %s -> %s", tn.Name, ws.Name, err)
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
