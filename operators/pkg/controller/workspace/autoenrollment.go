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

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *WorkspaceReconciler) manageAutoEnrollment(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	// check label and update if needed
	var wantedLabel string
	if utils.AutoEnrollEnabled(ws.Spec.AutoEnroll) {
		wantedLabel = string(ws.Spec.AutoEnroll)
	} else {
		wantedLabel = "disabled"
	}

	if ws.Labels[v1alpha2.WorkspaceLabelAutoenroll] != wantedLabel {
		ws.Labels[v1alpha2.WorkspaceLabelAutoenroll] = wantedLabel

		if err := r.Update(ctx, ws); err != nil {
			klog.Errorf("Error when updating workspace %s -> %s", ws.Name, err)
			return err
		}
	}

	// if actual AutoEnroll is WithApproval, nothing left to do
	// else, need to update Tenants in candidate status
	if ws.Spec.AutoEnroll != v1alpha1.AutoenrollWithApproval {
		if err := r.autoEnrollUpdateTenants(ctx, ws); err != nil {
			klog.Errorf("Error when updating tenants subscribed to workspace %s -> %s", ws.Name, err)
			return err
		}
	}

	return nil
}

// manage Tenants in candidate status
// this function shall be called only when AutoEnrollment is enabled and not WithApproval
func (r *WorkspaceReconciler) autoEnrollUpdateTenants(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	var tenantsToUpdate v1alpha2.TenantList
	targetLabel := fmt.Sprintf("%s%s", v1alpha2.WorkspaceLabelPrefix, ws.Name)

	err := r.List(ctx, &tenantsToUpdate, &client.MatchingLabels{targetLabel: string(v1alpha2.Candidate)})
	if err != nil {
		klog.Errorf("Error when listing tenants subscribed to workspace %s -> %s", ws.Name, err)
		return err
	}

	for _, tenant := range tenantsToUpdate.Items {
		if err := r.autoEnrollUpdateSingleTenant(ctx, ws, &tenant); err != nil {
			klog.Errorf("Error when updating tenant %s -> %s", tenant.Name, err)
			return err
		}
	}

	return nil
}

func (r *WorkspaceReconciler) autoEnrollUpdateSingleTenant(
	ctx context.Context,
	ws *v1alpha1.Workspace,
	tenant *v1alpha2.Tenant,
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
		klog.Errorf("Error when updating tenant %s -> %s", tenant.Name, err)
		return err
	}
	return nil
}
