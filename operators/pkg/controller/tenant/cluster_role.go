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

    crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
    "github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// createTenantClusterResources creates both ClusterRole and ClusterRoleBinding for tenant access
func (r *Reconciler) createTenantClusterResources(
    ctx context.Context,
    log logr.Logger,
    tn *crownlabsv1alpha2.Tenant,
) error {
    // Create ClusterRole for tenant access
    if err := r.createTenantClusterRole(ctx, log, tn); err != nil {
        return fmt.Errorf("error when creating cluster role for tenant %s: %w", tn.Name, err)
    }
    log.Info("Tenant ClusterRole created", "tenant", tn.Name)

    // Create ClusterRoleBinding for tenant access
    if err := r.createTenantClusterRoleBinding(ctx, log, tn); err != nil {
        return fmt.Errorf("error when creating cluster role binding for tenant %s: %w", tn.Name, err)
    }
    log.Info("Tenant ClusterRoleBinding created", "tenant", tn.Name)

    return nil
}

// deleteTenantClusterResources deletes both ClusterRole and ClusterRoleBinding for tenant access
func (r *Reconciler) deleteTenantClusterResources(
    ctx context.Context,
    log logr.Logger,
    tn *crownlabsv1alpha2.Tenant,
) error {
    // Delete ClusterRoleBinding 
    if err := r.deleteTenantClusterRoleBinding(ctx, log, tn); err != nil {
        return fmt.Errorf("error when deleting cluster role binding for tenant %s: %w", tn.Name, err)
    }
    log.Info("🔥 Tenant ClusterRoleBinding deleted", "tenant", tn.Name)

    // Delete ClusterRole
    if err := r.deleteTenantClusterRole(ctx, log, tn); err != nil {
        return fmt.Errorf("error when deleting cluster role for tenant %s: %w", tn.Name, err)
    }
    log.Info("🔥 Tenant ClusterRole deleted", "tenant", tn.Name)

    return nil
}

// createTenantClusterRole creates the ClusterRole for accessing the specific tenant resource
func (r *Reconciler) createTenantClusterRole(
    ctx context.Context,
    log logr.Logger,
    tn *crownlabsv1alpha2.Tenant,
) error {
    nsName := getNamespaceName(tn)
    crName := fmt.Sprintf("crownlabs-manage-%s", nsName)

    cr := rbacv1.ClusterRole{
        ObjectMeta: metav1.ObjectMeta{
            Name: crName,
        },
    }

    _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &cr, func() error {
        cr.Labels = r.updateTnResourceCommonLabels(cr.Labels)
        cr.Rules = []rbacv1.PolicyRule{{
            APIGroups:     []string{"crownlabs.polito.it"},
            Resources:     []string{"tenants"},
            ResourceNames: []string{tn.Name}, 
            Verbs:         []string{"get", "list", "watch", "patch", "update"},
        }}

        return controllerutil.SetControllerReference(tn, &cr, r.Scheme)
    })

    if err != nil {
        log.Error(err, "Unable to create or update cluster role for tenant %s -> %s", tn.Name, err)
    } else {
        log.Info("Cluster role for tenant %s created/updated", tn.Name)
    }

    return err
}

// createTenantClusterRoleBinding creates the ClusterRoleBinding for the tenant user
func (r *Reconciler) createTenantClusterRoleBinding(
    ctx context.Context,
    log logr.Logger,
    tn *crownlabsv1alpha2.Tenant,
) error {
    nsName := getNamespaceName(tn)
    crbName := fmt.Sprintf("crownlabs-manage-%s", nsName)

    crb := rbacv1.ClusterRoleBinding{
        ObjectMeta: metav1.ObjectMeta{
            Name: crbName,
        },
    }

    _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &crb, func() error {
        crb.Labels = r.updateTnResourceCommonLabels(crb.Labels)
        crb.RoleRef = rbacv1.RoleRef{
            Kind:     "ClusterRole",
            Name:     crbName, 
            APIGroup: "rbac.authorization.k8s.io",
        }
        crb.Subjects = []rbacv1.Subject{{
            Kind:     "User",
            Name:     tn.Name, 
            APIGroup: "rbac.authorization.k8s.io",
        }}

        return controllerutil.SetControllerReference(tn, &crb, r.Scheme)
    })

    if err != nil {
        log.Error(err, "Unable to create or update cluster role binding for tenant %s -> %s", tn.Name, err)
    } else {
        log.Info("Cluster role binding for tenant %s created/updated", tn.Name)
    }

    return err
}

// deleteTenantClusterRole deletes the ClusterRole for tenant access
func (r *Reconciler) deleteTenantClusterRole(
    ctx context.Context,
    log logr.Logger,
    tn *crownlabsv1alpha2.Tenant,
) error {
    nsName := getNamespaceName(tn)
    crName := fmt.Sprintf("crownlabs-manage-%s", nsName)

    cr := rbacv1.ClusterRole{
        ObjectMeta: metav1.ObjectMeta{
            Name: crName,
        },
    }

    err := utils.EnforceObjectAbsence(ctx, r.Client, &cr, "cluster role")
    if err != nil {
        log.Error(err, "Error when deleting cluster role for tenant %s -> %s", tn.Name, err)
    }

    return err
}

// deleteTenantClusterRoleBinding deletes the ClusterRoleBinding for tenant access
func (r *Reconciler) deleteTenantClusterRoleBinding(
    ctx context.Context,
    log logr.Logger,
    tn *crownlabsv1alpha2.Tenant,
) error {
    nsName := getNamespaceName(tn)
    crbName := fmt.Sprintf("crownlabs-manage-%s", nsName)

    crb := rbacv1.ClusterRoleBinding{
        ObjectMeta: metav1.ObjectMeta{
            Name: crbName,
        },
    }

    err := utils.EnforceObjectAbsence(ctx, r.Client, &crb, "cluster role binding")
    if err != nil {
        log.Error(err, "Error when deleting cluster role binding for tenant %s -> %s", tn.Name, err)
    }

    return err
}