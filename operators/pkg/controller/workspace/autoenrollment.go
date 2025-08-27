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
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func (r *Reconciler) enforceAutoEnrollment(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	log logr.Logger,
) error {
	// check label and update if needed
	var wantedLabel string
	if utils.AutoEnrollEnabled(ws.Spec.AutoEnroll) {
		wantedLabel = string(ws.Spec.AutoEnroll)
	} else {
		wantedLabel = "disabled"
	}

	if ws.Labels[v1alpha2.WorkspaceLabelAutoenroll] != wantedLabel {
		if err := r.enforcePreservingStatus(ctx, log, ws, func(workspace *v1alpha1.Workspace) *v1alpha1.Workspace {
			workspace.Labels[v1alpha2.WorkspaceLabelAutoenroll] = wantedLabel
			return workspace
		}); err != nil {
			log.Error(err, "Error when updating workspace", "workspace", ws.Name)
			return err
		}
	}

	// if actual AutoEnroll is WithApproval, nothing left to do
	// else, need to update Tenants in candidate status
	if ws.Spec.AutoEnroll != v1alpha1.AutoenrollWithApproval {
		if err := r.autoEnrollUpdateTenants(ctx, ws, log); err != nil {
			log.Error(err, "Error when updating tenants subscribed to workspace", "workspace", ws.Name)
			return err
		}
	}

	return nil
}

// manage Tenants in candidate status
// this function shall be called only when AutoEnrollment is enabled and not WithApproval.
func (r *Reconciler) autoEnrollUpdateTenants(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	log logr.Logger,
) error {
	var tenantsToUpdate v1alpha2.TenantList
	targetLabel := fmt.Sprintf("%s%s", v1alpha2.WorkspaceLabelPrefix, ws.Name)

	err := r.List(ctx, &tenantsToUpdate, &client.MatchingLabels{targetLabel: string(v1alpha2.Candidate)})
	if err != nil {
		log.Error(err, "Error when listing tenants subscribed to workspace", "workspace", ws.Name)
		return err
	}

	for i := range tenantsToUpdate.Items {
		tenant := &tenantsToUpdate.Items[i]
		if err := r.autoEnrollUpdateSingleTenant(ctx, ws, tenant, log); err != nil {
			log.Error(err, "Error when updating tenant", "tenant", tenant.Name)
			return err
		}
	}

	return nil
}

func (r *Reconciler) autoEnrollUpdateSingleTenant(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	tenant *v1alpha2.Tenant,
	log logr.Logger,
) error {
	patch := client.MergeFrom(tenant.DeepCopy())
	removeWorkspaceFromTenant(&tenant.Spec.Workspaces, ws.Name)
	// if AutoEnrollment is Immediate, add the workspace with User role
	if ws.Spec.AutoEnroll == v1alpha1.AutoenrollImmediate {
		tenant.Spec.Workspaces = append(
			tenant.Spec.Workspaces,
			v1alpha2.TenantWorkspaceEntry{
				Name: ws.Name,
				Role: v1alpha2.User,
			},
		)
	}
	if err := r.Patch(ctx, tenant, patch); err != nil {
		log.Error(err, "Error when updating tenant", "tenant", tenant.Name)
		return err
	}
	return nil
}
