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

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// enforceTenantClusterResources creates both ClusterRole and ClusterRoleBinding for tenant access.
func (r *Reconciler) enforceTenantClusterResources(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	// Create ClusterRole for tenant access
	if err := r.enforceTenantClusterRole(ctx, log, tn); err != nil {
		return fmt.Errorf("error when creating cluster role for tenant %s: %w", tn.Name, err)
	}
	log.Info("Tenant ClusterRole created", "tenant", tn.Name)

	// Create ClusterRoleBinding for tenant access
	if err := r.enforceTenantClusterRoleBinding(ctx, log, tn); err != nil {
		return fmt.Errorf("error when creating cluster role binding for tenant %s: %w", tn.Name, err)
	}
	log.Info("Tenant ClusterRoleBinding created", "tenant", tn.Name)

	return nil
}

// enforceTenantClusterResourcesAbsence ensures both ClusterRole and ClusterRoleBinding don't exist for tenant access.
func (r *Reconciler) enforceTenantClusterResourcesAbsence(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	// Delete ClusterRoleBinding
	if err := r.enforceTenantClusterRoleBindingAbsence(ctx, log, tn); err != nil {
		return fmt.Errorf("error when deleting cluster role binding for tenant %s: %w", tn.Name, err)
	}
	log.Info("🔥 Tenant ClusterRoleBinding deleted", "tenant", tn.Name)

	// Delete ClusterRole
	if err := r.enforceTenantClusterRoleAbsence(ctx, log, tn); err != nil {
		return fmt.Errorf("error when deleting cluster role for tenant %s: %w", tn.Name, err)
	}
	log.Info("🔥 Tenant ClusterRole deleted", "tenant", tn.Name)

	return nil
}

// enforceTenantClusterRole ensures the ClusterRole exists for accessing the specific tenant resource.
func (r *Reconciler) enforceTenantClusterRole(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	crName := forge.GetTenantClusterRoleResourceName(tn)

	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: crName,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &cr, func() error {
		// Configure the cluster role
		forge.ConfigureTenantClusterRole(&cr, tn, forge.UpdateTenantResourceCommonLabels(nil, r.TargetLabel))

		return controllerutil.SetControllerReference(tn, &cr, r.Scheme)
	})

	if err != nil {
		log.Error(err, "Unable to create or update cluster role for tenant", "tenant", tn.Name)
	} else {
		log.Info("Cluster role for tenant created/updated", "tenant", tn.Name)
	}

	return err
}

// enforceTenantClusterRoleBinding ensures the ClusterRoleBinding exists for the tenant user.
func (r *Reconciler) enforceTenantClusterRoleBinding(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	crbName := forge.GetTenantClusterRoleResourceName(tn)

	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: crbName,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &crb, func() error {
		// Configure the cluster role binding
		forge.ConfigureTenantClusterRoleBinding(&crb, tn, forge.UpdateTenantResourceCommonLabels(nil, r.TargetLabel))

		return controllerutil.SetControllerReference(tn, &crb, r.Scheme)
	})

	if err != nil {
		log.Error(err, "Unable to create or update cluster role binding for tenant", "tenant", tn.Name)
	} else {
		log.Info("Cluster role binding for tenant created/updated", "tenant", tn.Name)
	}

	return err
}

// enforceTenantClusterRoleAbsence ensures the ClusterRole doesn't exist for tenant access.
func (r *Reconciler) enforceTenantClusterRoleAbsence(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	crName := forge.GetTenantClusterRoleResourceName(tn)

	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: crName,
		},
	}

	err := utils.EnforceObjectAbsence(ctx, r.Client, &cr, "cluster role")
	if err != nil {
		log.Error(err, "Error when deleting cluster role for tenant", "tenant", tn.Name)
	}

	return err
}

// enforceTenantClusterRoleBindingAbsence ensures the ClusterRoleBinding doesn't exist for tenant access.
func (r *Reconciler) enforceTenantClusterRoleBindingAbsence(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	crbName := forge.GetTenantClusterRoleResourceName(tn)

	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: crbName,
		},
	}

	err := utils.EnforceObjectAbsence(ctx, r.Client, &crb, "cluster role binding")
	if err != nil {
		log.Error(err, "Error when deleting cluster role binding for tenant", "tenant", tn.Name)
	}

	return err
}
