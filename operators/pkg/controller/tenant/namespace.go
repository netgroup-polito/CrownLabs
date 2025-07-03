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
	"strings"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	v1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func (r *Reconciler) createResourcesRelatedToPersonalNamespace(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	// Create the personal namespace for the tenant
	if err := r.createPersonalNamespace(ctx, tn); err != nil {
		return fmt.Errorf("error when creating personal namespace for tenant %s: %w", tn.Name, err)
	}
	klog.Info("Personal namespace created ", "namespace", getNamespaceName(tn))

	// manage resource quota
	if err := r.createResourceQuota(ctx, tn); err != nil {
		return fmt.Errorf("error when creating resource quota for tenant %s: %w", tn.Name, err)
	}
	klog.Info("Resource quota created ", "namespace", getNamespaceName(tn))

	// manage role binding for instance management
	if err := r.createInstanceRoleBinding(ctx, tn); err != nil {
		return fmt.Errorf("error when creating role binding for tenant %s: %w", tn.Name, err)
	}
	klog.Info("Role binding created ", "namespace", getNamespaceName(tn))

	// Network Policies
	if err := r.createDenyNetworkPolicy(ctx, tn); err != nil {
		return fmt.Errorf("error when creating deny network policy for tenant %s: %w", tn.Name, err)
	}
	klog.Info("Deny network policy created ", "namespace", getNamespaceName(tn))

	if err := r.createAllowNetworkPolicy(ctx, tn); err != nil {
		return fmt.Errorf("error when creating allow network policy for tenant %s: %w", tn.Name, err)
	}
	klog.Info("Allow network policy created ", "namespace", getNamespaceName(tn))

	return nil
}

func (r *Reconciler) deleteResourcesRelatedToPersonalNamespace(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	// Delete Network Policies
	if err := r.deleteDenyNetworkPolicy(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting deny network policy for tenant %s: %w", tn.Name, err)
	}
	klog.Info("🔥 Deny network policy deleted", "namespace", getNamespaceName(tn))

	if err := r.deleteAllowNetworkPolicy(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting allow network policy for tenant %s: %w", tn.Name, err)
	}
	klog.Info("🔥 Allow network policy deleted", "namespace", getNamespaceName(tn))
	// Delete the role binding for instance management
	if err := r.deleteInstanceRoleBinding(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting role binding for tenant %s: %w", tn.Name, err)
	}
	klog.Info("🔥 Role binding deleted", "namespace", getNamespaceName(tn))

	// Delete the resource quota for the personal namespace
	if err := r.deleteResourceQuota(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting resource quota for tenant %s: %w", tn.Name, err)
	}
	klog.Info("🔥 Resource quota deleted", "namespace", getNamespaceName(tn))

	// Delete the personal namespace for the tenant
	if err := r.deletePersonalNamespace(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting personal namespace for tenant %s: %w", tn.Name, err)
	}
	klog.Info("🔥 Personal namespace deleted", "namespace", getNamespaceName(tn))

	return nil
}

func (r *Reconciler) createPersonalNamespace(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: getNamespaceName(tn),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		ns.Labels = r.updateTnResourceCommonLabels(ns.Labels)
		ns.Labels["crownlabs.polito.it/type"] = "tenant"
		ns.Labels["crownlabs.polito.it/name"] = tn.Name
		ns.Labels["crownlabs.polito.it/instance-resources-replication"] = "true"

		return controllerutil.SetControllerReference(tn, &ns, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error when creating namespace for tenant %s: %w", tn.Name, err)
	}

	return nil
}

// deleteClusterNamespace deletes the namespace for the tenant, if it fails then it returns an error.
func (r *Reconciler) deletePersonalNamespace(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: getNamespaceName(tn),
		},
	}

	err := utils.EnforceObjectAbsence(ctx, r.Client, &ns, "personal namespace")

	if err != nil {
		klog.Errorf("Error when deleting namespace of tenant %s -> %s", tn.Name, err)
	}

	return err
}

// checkNamespaceKeepAlive checks to see if the namespace should be deleted.
func (r *Reconciler) checkNamespaceKeepAlive(ctx context.Context, tn *v1alpha2.Tenant) (keepNsOpen bool, err error) {
	// We check to see if last login was more than r.TenantNSKeepAlive in the past:
	// if so, temporarily delete the namespace. We assume that a lastLogin of 0 occurs when a user is first created

	// Calculate time elapsed since lastLogin (now minus lastLogin in seconds)
	sPassed := time.Since(tn.Spec.LastLogin.Time)

	klog.Infof("Last login of tenant %s was %s ago", tn.Name, sPassed)

	// Attempt to get instances in current namespace
	list := &v1alpha2.InstanceList{}

	if err := r.List(ctx, list, client.InNamespace(getNamespaceName(tn))); err != nil {
		return true, err
	}

	if sPassed > r.TenantNSKeepAlive { // seconds
		klog.Infof("Over %s elapsed since last login of tenant %s: tenant namespace shall be absent", r.TenantNSKeepAlive, tn.Name)
		if len(list.Items) == 0 {
			klog.Infof("No instances found for tenant %s: namespace can be deleted", tn.Name)
			return false, nil
		}
		klog.Infof("Instances found for tenant %s. Namespace will not be deleted", tn.Name)
	} else {
		klog.Infof("Under %s (limit) elapsed since last login of tenant %s: tenant namespace shall be present", r.TenantNSKeepAlive, tn.Name)
	}

	return true, nil
}

// returns the name of the namespace for the tenant.
func getNamespaceName(tn *v1alpha2.Tenant) string {
	return fmt.Sprintf("tenant-%s", strings.ReplaceAll(tn.Name, ".", "-"))
}

func (r *Reconciler) createResourceQuota(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := getNamespaceName(tn)
	rq := v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-resource-quota",
			Namespace: nsName,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &rq, func() error {
		rq.Labels = r.updateTnResourceCommonLabels(rq.Labels)
		rq.Spec.Hard = forge.TenantResourceQuotaSpec(&tn.Status.Quota)

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
	nsName := getNamespaceName(tn)
	rq := v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-resource-quota",
			Namespace: nsName,
		},
	}

	err := utils.EnforceObjectAbsence(ctx, r.Client, &rq, "resource quota")

	if err != nil {
		klog.Errorf("Error when deleting resource quota for tenant %s -> %s", tn.Name, err)
	}

	return err
}

func (r *Reconciler) createInstanceRoleBinding(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := getNamespaceName(tn)
	rb := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-manage-instances",
			Namespace: nsName,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &rb, func() error {
		rb.Labels = r.updateTnResourceCommonLabels(rb.Labels)
		rb.RoleRef = rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "crownlabs-manage-instances",
			APIGroup: "rbac.authorization.k8s.io",
		}
		rb.Subjects = []rbacv1.Subject{{
			Kind:     "User",
			Name:     tn.Name,
			APIGroup: "rbac.authorization.k8s.io",
		}}

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
	nsName := getNamespaceName(tn)
	rb := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-manage-instances",
			Namespace: nsName,
		},
	}

	err := utils.EnforceObjectAbsence(ctx, r.Client, &rb, "role binding")
	if err != nil {
		klog.Errorf("Error when deleting role binding for tenant %s -> %s", tn.Name, err)
	}

	return err
}

func (r *Reconciler) createDenyNetworkPolicy(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := getNamespaceName(tn)
	netPolDeny := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-deny-ingress-traffic",
			Namespace: nsName,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, netPolDeny, func() error {
		netPolDeny.Labels = r.updateTnResourceCommonLabels(netPolDeny.Labels)
		netPolDeny.Spec.PodSelector.MatchLabels = make(map[string]string)
		netPolDeny.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{
			From: []netv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}},
		}}
		return controllerutil.SetControllerReference(tn, netPolDeny, r.Scheme)
	})

	return err
}

func (r *Reconciler) createAllowNetworkPolicy(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := getNamespaceName(tn)
	netPolAllow := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-allow-trusted-ingress-traffic",
			Namespace: nsName,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, netPolAllow, func() error {
		netPolAllow.Labels = r.updateTnResourceCommonLabels(netPolAllow.Labels)
		netPolAllow.Spec.PodSelector.MatchLabels = make(map[string]string)
		netPolAllow.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{
			From: []netv1.NetworkPolicyPeer{{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"crownlabs.polito.it/allow-instance-access": "true",
					},
				},
			}},
		}}
		return controllerutil.SetControllerReference(tn, netPolAllow, r.Scheme)
	})

	return err
}

func (r *Reconciler) deleteDenyNetworkPolicy(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	nsName := getNamespaceName(tn)
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
	nsName := getNamespaceName(tn)
	netPolAllow := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crownlabs-allow-trusted-ingress-traffic",
			Namespace: nsName,
		},
	}

	return utils.EnforceObjectAbsence(ctx, r.Client, netPolAllow, "allow network policy")
}
