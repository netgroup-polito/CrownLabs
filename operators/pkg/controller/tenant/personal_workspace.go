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
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// handlePersonalWorkspace handles the personal workspace for the tenant.
func (r *Reconciler) handlePersonalWorkspace(ctx context.Context, tn *v1alpha2.Tenant) error {
	log := ctrl.LoggerFrom(ctx)
	if !tn.Status.PersonalNamespace.Created {
		// if the personal namespace is not created skip the rest.
		tn.Status.PersonalWorkspaceCreated = false
		log.Info("Tenant namespace does not exist, skipping personal workspace handling")
		return nil
	}
	manageTemplatesRB := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: forge.ManageTemplatesRoleName, Namespace: tn.Status.PersonalNamespace.Name}}
	if tn.Spec.CreatePersonalWorkspace {
		forge.ConfigurePersonalWorkspaceManageTemplatesBinding(tn, &manageTemplatesRB, forge.UpdateTenantResourceCommonLabels(manageTemplatesRB.Labels, r.TargetLabel))
		res, err := ctrl.CreateOrUpdate(ctx, r.Client, &manageTemplatesRB, func() error {
			return ctrl.SetControllerReference(tn, &manageTemplatesRB, r.Scheme)
		})
		if err != nil {
			tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, "personal-workspace")
			tn.Status.PersonalWorkspaceCreated = false
			return err
		}
		tn.Status.PersonalWorkspaceCreated = true
		log.Info(fmt.Sprintf("Personal Workspace role binding %s", res))
	} else {
		tn.Status.PersonalWorkspaceCreated = false
		if err := utils.EnforceObjectAbsence(ctx, r.Client, &manageTemplatesRB, "personal workspace role binding"); err != nil {
			return err
		}
	}
	return nil
}

// checkPersonalWorkspaceKeepAlive checks if the personal workspace should be kept alive because it has templates.
func (r *Reconciler) checkPersonalWorkspaceKeepAlive(ctx context.Context, tn *v1alpha2.Tenant) (bool, error) {
	log := ctrl.LoggerFrom(ctx)
	templates := &v1alpha2.TemplateList{}
	if !tn.Spec.CreatePersonalWorkspace {
		return false, nil
	}
	if err := r.List(ctx, templates, client.InNamespace(forge.GetTenantNamespaceName(tn))); err != nil {
		return true, err
	}
	if len(templates.Items) > 0 {
		log.Info("Templates found for tenant", "tenant", tn.Name)
		return true, nil
	}
	return false, nil
}
