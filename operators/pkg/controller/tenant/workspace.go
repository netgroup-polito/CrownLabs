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
	"maps"
	"slices"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// If there is some problem with a workspace, we add it to the related list in the tenant's status
// for each valid workspace, a label is added to the tenant.
func (r *Reconciler) syncWorkspaces(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	// create fresh lists
	tn.Status.FailingWorkspaces = []string{}
	labels := make(map[string]string)
	maps.Copy(labels, tn.Labels)
	deleteWorkspacesRelatedLabels(labels)

	nonBaseWorkspaces := 0

	for ws := range tn.Spec.Workspaces {
		nonBaseWorkspaces += r.syncSingleWorkspace(ctx, log, tn, labels, &tn.Spec.Workspaces[ws])
	}

	if nonBaseWorkspaces == 0 {
		r.addNoWorkspaceLabel(labels)
	} else {
		r.removeNoWorkspaceLabel(labels)
	}

	// update labels
	if err := r.enforcePreservingStatus(ctx, log, tn, func(t *v1alpha2.Tenant) *v1alpha2.Tenant {
		t.Labels = labels
		return t
	}); err != nil {
		log.Error(err, "Error updating tenant with workspaces labels", "tenant", tn.Name)
		return err
	}

	log.Info("Updated tenant with workspaces labels", "tenant", tn.Name)

	return nil
}

func (r *Reconciler) syncSingleWorkspace(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
	labels map[string]string,
	tenantWorkspace *v1alpha2.TenantWorkspaceEntry,
) int {
	nonBaseWorkspaces := 0
	workspace := &v1alpha1.Workspace{}

	err := r.Get(ctx, types.NamespacedName{
		Name: tenantWorkspace.Name,
	}, workspace)
	switch {
	case err != nil:
		// if there was a problem, add the workspace to the status of the tenant
		log.Error(err, "Error when checking if workspace exists in tenant", "workspace", tenantWorkspace.Name, "tenant", tn.Name)
		tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, tenantWorkspace.Name)
		tnOpinternalErrors.WithLabelValues("tenant", "workspace-not-exist").Inc()
	case tenantWorkspace.Role == v1alpha2.Candidate && workspace.Spec.AutoEnroll != v1alpha1.AutoenrollWithApproval:
		// Candidate role is allowed only if the workspace has autoEnroll = WithApproval
		log.Error(err, "Workspace has not autoEnroll with approval, Candidate role is not allowed in tenant", "workspace", tenantWorkspace.Name, "tenant", tn.Name)
		tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, tenantWorkspace.Name)
	default:
		r.addWorkspaceLabel(labels, tenantWorkspace.Name, tenantWorkspace.Role)
		if slices.Index(r.BaseWorkspaces, tenantWorkspace.Name) == -1 {
			nonBaseWorkspaces++
		}
	}

	return nonBaseWorkspaces
}

func deleteWorkspacesRelatedLabels(
	labels map[string]string,
) {
	for lab := range labels {
		if strings.HasPrefix(lab, v1alpha2.WorkspaceLabelPrefix) {
			delete(labels, lab)
		}
	}
}

// this functions only adds the label to the structure, it does not update the tenant.
func (r *Reconciler) addWorkspaceLabel(
	labels map[string]string,
	workspaceName string,
	role v1alpha2.WorkspaceUserRole,
) {
	labels[v1alpha2.WorkspaceLabelPrefix+workspaceName] = string(role)
}

// this functions only adds the label to the structure, it does not update the tenant.
func (r *Reconciler) addNoWorkspaceLabel(
	labels map[string]string,
) {
	labels[forge.NoWorkspacesLabelKey] = forge.NoWorkspacesLabelValue
}

func (r *Reconciler) removeNoWorkspaceLabel(
	labels map[string]string,
) {
	delete(labels, forge.NoWorkspacesLabelKey)
}

func (r *Reconciler) enforceServiceQuota(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	// get the enrolled workspaces
	wss, err := r.getWorkspacesList(
		ctx,
		log,
		r.getEnrolledWorkspaces(tn),
	)
	if err != nil {
		return err
	}

	// update resource quota in the status of the tenant after checking validity of workspaces.
	tn.Status.Quota = forge.TenantResourceList(wss, tn.Spec.Quota)

	return nil
}

func (r *Reconciler) getEnrolledWorkspaces(
	tn *v1alpha2.Tenant,
) []v1alpha2.TenantWorkspaceEntry {
	validWorkspaces := make([]v1alpha2.TenantWorkspaceEntry, 0, len(tn.Spec.Workspaces))
	for _, ws := range tn.Spec.Workspaces {
		// skip workspaces in Candidate status
		if ws.Role == v1alpha2.Candidate {
			continue
		}

		// skip failing workspaces
		if slices.Contains(tn.Status.FailingWorkspaces, ws.Name) {
			continue
		}

		validWorkspaces = append(validWorkspaces, ws)
	}

	return validWorkspaces
}

func (r *Reconciler) getWorkspacesList(
	ctx context.Context,
	log logr.Logger,
	tnWss []v1alpha2.TenantWorkspaceEntry,
) ([]v1alpha1.Workspace, error) {
	workspaces := make([]v1alpha1.Workspace, 0, len(tnWss))

	for _, ws := range tnWss {
		workspace := v1alpha1.Workspace{}
		err := r.Get(ctx, types.NamespacedName{
			Name: ws.Name,
		}, &workspace)
		if err != nil {
			log.Error(err, "Error when getting workspace", "workspace", ws.Name)
			return nil, err
		}

		// add the workspace to the list
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}
