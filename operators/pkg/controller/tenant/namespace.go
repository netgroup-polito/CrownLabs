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
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func (r *Reconciler) enforceResourcesRelatedToPersonalNamespace(
	ctx context.Context,
	extlog logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	log := extlog.WithValues("namespace", forge.GetTenantNamespaceName(tn))

	// Create the personal namespace for the tenant
	if err := r.createPersonalNamespace(ctx, tn); err != nil {
		return fmt.Errorf("error when creating personal namespace for tenant %s: %w", tn.Name, err)
	}
	log.Info("Personal namespace created")

	// manage resource quota
	if err := r.createResourceQuota(ctx, tn); err != nil {
		return fmt.Errorf("error when creating resource quota for tenant %s: %w", tn.Name, err)
	}
	log.Info("Resource quota created")

	// manage role binding for instance management
	if err := r.createInstanceRoleBinding(ctx, tn); err != nil {
		return fmt.Errorf("error when creating role binding for tenant %s: %w", tn.Name, err)
	}
	log.Info("Role binding created")

	// Network Policies
	if err := r.createDenyNetworkPolicy(ctx, tn); err != nil {
		return fmt.Errorf("error when creating deny network policy for tenant %s: %w", tn.Name, err)
	}
	log.Info("Deny network policy created")

	if err := r.createAllowNetworkPolicy(ctx, tn); err != nil {
		return fmt.Errorf("error when creating allow network policy for tenant %s: %w", tn.Name, err)
	}
	log.Info("Allow network policy created")

	return nil
}

func (r *Reconciler) deleteResourcesRelatedToPersonalNamespace(
	ctx context.Context,
	extlog logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	log := extlog.WithValues("namespace", forge.GetTenantNamespaceName(tn))

	// Delete Network Policies
	if err := r.deleteDenyNetworkPolicy(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting deny network policy for tenant %s: %w", tn.Name, err)
	}
	log.Info("🔥 Deny network policy deleted")

	if err := r.deleteAllowNetworkPolicy(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting allow network policy for tenant %s: %w", tn.Name, err)
	}
	log.Info("🔥 Allow network policy deleted")
	// Delete the role binding for instance management
	if err := r.deleteInstanceRoleBinding(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting role binding for tenant %s: %w", tn.Name, err)
	}
	log.Info("🔥 Role binding deleted")

	// Delete the resource quota for the personal namespace
	if err := r.deleteResourceQuota(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting resource quota for tenant %s: %w", tn.Name, err)
	}
	log.Info("🔥 Resource quota deleted")

	// Delete the personal namespace for the tenant
	if err := r.deletePersonalNamespace(ctx, log, tn); err != nil {
		return fmt.Errorf("error when deleting personal namespace for tenant %s: %w", tn.Name, err)
	}
	log.Info("🔥 Personal namespace deleted")

	return nil
}

func (r *Reconciler) createPersonalNamespace(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: forge.GetTenantNamespaceName(tn),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		// Configure the namespace using the forge package
		forge.ConfigureTenantNamespace(&ns, tn, forge.UpdateTenantResourceCommonLabels(ns.Labels, r.TargetLabel))

		return controllerutil.SetControllerReference(tn, &ns, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error when creating namespace for tenant %s: %w", tn.Name, err)
	}

	tn.Status.PersonalNamespace.Created = true
	tn.Status.PersonalNamespace.Name = ns.Name

	return nil
}

// deletePersonalNamespace deletes the namespace for the tenant, if it fails then it returns an error.
func (r *Reconciler) deletePersonalNamespace(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: forge.GetTenantNamespaceName(tn),
		},
	}

	err := utils.EnforceObjectAbsence(ctx, r.Client, &ns, "personal namespace")

	if err != nil {
		log.Error(err, "Error when deleting namespace of tenant", "tenant", tn.Name)
	}

	tn.Status.PersonalNamespace.Created = false
	tn.Status.PersonalNamespace.Name = ""

	return err
}

// checkNamespaceKeepAlive checks to see if the namespace should be deleted.
func (r *Reconciler) checkNamespaceKeepAlive(ctx context.Context, log logr.Logger, tn *v1alpha2.Tenant) (keepNsOpen bool, err error) {
	// We check to see if last login was more than r.TenantNSKeepAlive in the past:
	// if so, temporarily delete the namespace. We assume that a lastLogin of 0 occurs when a user is first created

	// Calculate time elapsed since lastLogin (now minus lastLogin in seconds)
	sPassed := time.Since(tn.Spec.LastLogin.Time)

	log.Info("Last login checked", "tenant", tn.Name, "elapsed", sPassed)

	// Attempt to get instances in current namespace
	list := &v1alpha2.InstanceList{}

	if err := r.List(ctx, list, client.InNamespace(forge.GetTenantNamespaceName(tn))); err != nil {
		return true, err
	}

	if sPassed > r.TenantNSKeepAlive { // seconds
		log.Info("Over elapsed since last login of tenant: tenant namespace shall be absent", "elapsed", r.TenantNSKeepAlive, "tenant", tn.Name)
		if len(list.Items) == 0 {
			log.Info("No instances found for tenant: namespace can be deleted", "tenant", tn.Name)
			return false, nil
		}
		log.Info("Instances found for tenant: namespace will not be deleted", "tenant", tn.Name)
	} else {
		log.Info("Under (limit) elapsed since last login of tenant: tenant namespace shall be present", "limit", r.TenantNSKeepAlive, "tenant", tn.Name)
	}

	return true, nil
}

func (r *Reconciler) createResourceQuota(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := forge.GetTenantNamespaceName(tn)
	rq := v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-resource-quota",
			Namespace: nsName,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &rq, func() error {
		// Configure the resource quota using the forge package
		forge.ConfigureTenantResourceQuota(&rq, &tn.Status.Quota, forge.UpdateTenantResourceCommonLabels(rq.Labels, r.TargetLabel))

		return controllerutil.SetControllerReference(tn, &rq, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error when creating resource quota for tenant %s: %w", tn.Name, err)
	}

	return nil
}

func (r *Reconciler) deleteResourceQuota(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := forge.GetTenantNamespaceName(tn)
	rq := v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-resource-quota",
			Namespace: nsName,
		},
	}

	return utils.EnforceObjectAbsence(ctx, r.Client, &rq, "resource quota")
}

func (r *Reconciler) createInstanceRoleBinding(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := forge.GetTenantNamespaceName(tn)
	rb := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-manage-instances",
			Namespace: nsName,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &rb, func() error {
		// Configure the role binding using the forge package
		forge.ConfigureTenantInstancesRoleBinding(&rb, tn, forge.UpdateTenantResourceCommonLabels(rb.Labels, r.TargetLabel))

		return controllerutil.SetControllerReference(tn, &rb, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error when creating role binding for tenant %s: %w", tn.Name, err)
	}

	return nil
}

func (r *Reconciler) deleteInstanceRoleBinding(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := forge.GetTenantNamespaceName(tn)
	rb := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-manage-instances",
			Namespace: nsName,
		},
	}

	return utils.EnforceObjectAbsence(ctx, r.Client, &rb, "role binding")
}

func (r *Reconciler) createDenyNetworkPolicy(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := forge.GetTenantNamespaceName(tn)
	netPolDeny := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-deny-ingress-traffic",
			Namespace: nsName,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, netPolDeny, func() error {
		// Configure the network policy using the forge package
		forge.ConfigureTenantDenyNetworkPolicy(netPolDeny, forge.UpdateTenantResourceCommonLabels(netPolDeny.Labels, r.TargetLabel))

		return controllerutil.SetControllerReference(tn, netPolDeny, r.Scheme)
	})

	return err
}

func (r *Reconciler) createAllowNetworkPolicy(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := forge.GetTenantNamespaceName(tn)
	netPolAllow := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-allow-trusted-ingress-traffic",
			Namespace: nsName,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, netPolAllow, func() error {
		// Configure the network policy using the forge package
		forge.ConfigureTenantAllowNetworkPolicy(netPolAllow, forge.UpdateTenantResourceCommonLabels(netPolAllow.Labels, r.TargetLabel))

		return controllerutil.SetControllerReference(tn, netPolAllow, r.Scheme)
	})

	return err
}

func (r *Reconciler) deleteDenyNetworkPolicy(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := forge.GetTenantNamespaceName(tn)
	netPolDeny := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-deny-ingress-traffic",
			Namespace: nsName,
		},
	}

	return utils.EnforceObjectAbsence(ctx, r.Client, netPolDeny, "deny network policy")
}

func (r *Reconciler) deleteAllowNetworkPolicy(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := forge.GetTenantNamespaceName(tn)
	netPolAllow := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-allow-trusted-ingress-traffic",
			Namespace: nsName,
		},
	}

	return utils.EnforceObjectAbsence(ctx, r.Client, netPolAllow, "allow network policy")
}
