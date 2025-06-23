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

package tenant

import (
	"context"
	"strings"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// if there is some problem with a workspace, we add it to the related list in the tenant's status
// for each valid workspace, a label is added to the tenant
func (r *TenantReconciler) manageWorkspaces(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	// create fresh lists
	tn.Status.FailingWorkspaces = []string{}
	deleteWorkspacesRelatadLabels(tn.Labels)

	for ws := range tn.Spec.Workspaces {
		if err := r.manageSingleWorkspace(ctx, tn, &tn.Spec.Workspaces[ws]); err != nil {
			return err
		}
	}

	// update labels
	if err := r.updatePreservingStatus(ctx, tn); err != nil {
		klog.Errorf("Error updating tenant %s with workspaces labels: %v", tn.Name, err)
		// if there was a problem updating the tenant, we add all the workspaces to the failing list
		return err
	}

	klog.Infof("Updated tenant %s with workspaces labels", tn.Name)

	return nil
}

func (r *TenantReconciler) manageSingleWorkspace(
	ctx context.Context,
	tn *v1alpha2.Tenant,
	tenantWorkspace *v1alpha2.TenantWorkspaceEntry,
) error {
	workspace := &v1alpha1.Workspace{}

	err := r.Get(ctx, types.NamespacedName{
		Name: tenantWorkspace.Name,
	}, workspace)
	switch {
	case err != nil:
		// if there was a problem, add the workspace to the status of the tenant
		klog.Errorf("Error when checking if workspace %s exists in tenant %s -> %s", tenantWorkspace.Name, tn.Name, err)
		tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, tenantWorkspace.Name)
		tnOpinternalErrors.WithLabelValues("tenant", "workspace-not-exist").Inc()
	case tenantWorkspace.Role == v1alpha2.Candidate && workspace.Spec.AutoEnroll != v1alpha1.AutoenrollWithApproval:
		// Candidate role is allowed only if the workspace has autoEnroll = WithApproval
		klog.Errorf("Workspace %s has not autoEnroll with approval, Candidate role is not allowed in tenant %s", tenantWorkspace.Name, tn.Name)
		tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, tenantWorkspace.Name)
	default:
		r.addWorkspaceLabel(ctx, tn, tenantWorkspace.Name, tenantWorkspace.Role)
	}

	return nil
}

func deleteWorkspacesRelatadLabels(
	labels map[string]string,
) {
	for lab := range labels {
		if strings.HasPrefix(lab, v1alpha2.WorkspaceLabelPrefix) {
			delete(labels, lab)
		}
	}
}

// this functions only adds the label to the structure, it does not update the tenant
func (r *TenantReconciler) addWorkspaceLabel(
	ctx context.Context,
	tn *v1alpha2.Tenant,
	workspaceName string,
	role v1alpha2.WorkspaceUserRole,
) error {
	// TODO: manage base workspaces (don't add label)

	if tn.Labels == nil {
		tn.Labels = make(map[string]string, 1)
	}

	tn.Labels[v1alpha2.WorkspaceLabelPrefix+workspaceName] = string(role)

	return nil
}
