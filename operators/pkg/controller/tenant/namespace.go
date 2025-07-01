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

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	netv1 "k8s.io/api/networking/v1"

	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func (r *Reconciler) createResourcesRelatedToPersonalNamespace(
	ctx context.Context,
	log logr.Logger,
	tn *crownlabsv1alpha2.Tenant,
) error {
	// Create the personal namespace for the tenant
	if err := r.createPersonalNamespace(ctx, tn); err != nil {
		return fmt.Errorf("error when creating personal namespace for tenant %s: %w", tn.Name, err)
	}
	log.Info("Personal namespace created", "namespace", getNamespaceName(tn))

	// manage resource quota
	if err := r.createResourceQuota(ctx, tn); err != nil {
		return fmt.Errorf("error when creating resource quota for tenant %s: %w", tn.Name, err)
	}
	log.Info("Resource quota created", "namespace", getNamespaceName(tn))

	// TODO: tutte le cose che partono da enforceClusterResources

	 // manage role binding for instance management
    if err := r.createInstanceRoleBinding(ctx, tn); err != nil {
        return fmt.Errorf("error when creating role binding for tenant %s: %w", tn.Name, err)
    }
    log.Info("Role binding created", "namespace", getNamespaceName(tn))

	// Network Policies
    if err := r.createDenyNetworkPolicy(ctx, tn); err != nil {
        return fmt.Errorf("error when creating deny network policy for tenant %s: %w", tn.Name, err)
    }
    log.Info("Deny network policy created", "namespace", getNamespaceName(tn))

    if err := r.createAllowNetworkPolicy(ctx, tn); err != nil {
        return fmt.Errorf("error when creating allow network policy for tenant %s: %w", tn.Name, err)
    }
    log.Info("Allow network policy created", "namespace", getNamespaceName(tn))


	return nil
}

func (r *Reconciler) deleteResourcesRelatedToPersonalNamespace(
	ctx context.Context,
	log logr.Logger,
	tn *crownlabsv1alpha2.Tenant,
) error {
	// TODO: tutte le cose che partono da enforceClusterResources


	// Delete Network Policies
    if err := r.deleteDenyNetworkPolicy(ctx, tn); err != nil {
        return fmt.Errorf("error when deleting deny network policy for tenant %s: %w", tn.Name, err)
    }
    log.Info("Deny network policy deleted", "namespace", getNamespaceName(tn))

    if err := r.deleteAllowNetworkPolicy(ctx, tn); err != nil {
        return fmt.Errorf("error when deleting allow network policy for tenant %s: %w", tn.Name, err)
    }
    log.Info("Allow network policy deleted", "namespace", getNamespaceName(tn))
	 // Delete the role binding for instance management
    if err := r.deleteInstanceRoleBinding(ctx, tn); err != nil {
        return fmt.Errorf("error when deleting role binding for tenant %s: %w", tn.Name, err)
    }
    log.Info("Role binding deleted", "namespace", getNamespaceName(tn))

	// Delete the resource quota for the personal namespace
	if err := r.deleteResourceQuota(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting resource quota for tenant %s: %w", tn.Name, err)
	}
	log.Info("Resource quota deleted", "namespace", getNamespaceName(tn))

	// Delete the personal namespace for the tenant
	if err := r.deletePersonalNamespace(ctx, tn); err != nil {
		return fmt.Errorf("error when deleting personal namespace for tenant %s: %w", tn.Name, err)
	}
	log.Info("Personal namespace deleted", "namespace", getNamespaceName(tn))

	return nil
}

func (r *Reconciler) createPersonalNamespace(
	ctx context.Context,
	tn *crownlabsv1alpha2.Tenant,
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
	tn *crownlabsv1alpha2.Tenant,
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
func (r *Reconciler) checkNamespaceKeepAlive(ctx context.Context, tn *crownlabsv1alpha2.Tenant) (keepNsOpen bool, err error) {
	// We check to see if last login was more than r.TenantNSKeepAlive in the past:
	// if so, temporarily delete the namespace. We assume that a lastLogin of 0 occurs when a user is first created

	// Calculate time elapsed since lastLogin (now minus lastLogin in seconds)
	sPassed := time.Since(tn.Spec.LastLogin.Time)

	klog.Infof("Last login of tenant %s was %s ago", tn.Name, sPassed)

	// Attempt to get instances in current namespace
	list := &crownlabsv1alpha2.InstanceList{}

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
func getNamespaceName(tn *crownlabsv1alpha2.Tenant) string {
	return fmt.Sprintf("tenant-%s", strings.ReplaceAll(tn.Name, ".", "-"))
}

func (r *Reconciler) createResourceQuota(
	ctx context.Context,
	tn *crownlabsv1alpha2.Tenant,
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
	tn *crownlabsv1alpha2.Tenant,
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
    tn *crownlabsv1alpha2.Tenant,
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
    tn *crownlabsv1alpha2.Tenant,
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
    tn *crownlabsv1alpha2.Tenant,
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
    tn *crownlabsv1alpha2.Tenant,
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
    tn *crownlabsv1alpha2.Tenant,
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
    tn *crownlabsv1alpha2.Tenant,
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

// // Deletes namespace or updates the cluster resources.
// func (r *Reconciler) enforceClusterResources(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string, keepNsOpen bool) (nsOk bool, err error) {
// 	nsOk = false // nsOk must be initialized for later use

// 	if keepNsOpen {
// 		nsOk, err = r.createOrUpdateClusterResources(ctx, tn, nsName)
// 		if nsOk {
// 			klog.Infof("Namespace %s for tenant %s updated", nsName, tn.Name)
// 			tn.Status.PersonalNamespace.Created = true
// 			tn.Status.PersonalNamespace.Name = nsName
// 			if err != nil {
// 				klog.Errorf("Unable to update cluster resource of tenant %s -> %s", tn.Name, err)
// 				tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
// 			}
// 			klog.Infof("Cluster resources for tenant %s updated", tn.Name)
// 		} else {
// 			klog.Errorf("Unable to update namespace of tenant %s -> %s", tn.Name, err)
// 			tn.Status.PersonalNamespace.Created = false
// 			tn.Status.PersonalNamespace.Name = ""
// 			tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
// 		}
// 	} else {
// 		err := r.deleteClusterNamespace(ctx, tn, nsName)
// 		if err == nil {
// 			klog.Infof("Namespace %s for tenant %s enforced to be absent", nsName, tn.Name)
// 			tn.Status.PersonalNamespace.Created = false
// 			tn.Status.PersonalNamespace.Name = ""
// 		} else {
// 			klog.Errorf("Unable to delete namespace of tenant %s -> %s", tn.Name, err)
// 			tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
// 		}
// 	}
// 	return nsOk, err
// }

// // updateTnNamespace updates the tenant namespace.
// func (r *Reconciler) updateTnNamespace(ns *v1.Namespace, tnName string) {
// 	ns.Labels = r.updateTnResourceCommonLabels(ns.Labels)
// 	ns.Labels["crownlabs.polito.it/type"] = "tenant"
// 	ns.Labels["crownlabs.polito.it/name"] = tnName
// 	ns.Labels["crownlabs.polito.it/instance-resources-replication"] = "true"
// }

// // createOrUpdateClusterResources creates the namespace for the tenant, if it succeeds it then tries to create the rest of the resources with a fail-fast:false strategy.
// func (r *Reconciler) createOrUpdateClusterResources(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) (nsOk bool, err error) {
// 	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}

// 	if _, nsErr := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
// 		r.updateTnNamespace(&ns, tn.Name)
// 		return ctrl.SetControllerReference(tn, &ns, r.Scheme)
// 	}); nsErr != nil {
// 		klog.Errorf("Error when updating namespace of tenant %s -> %s", tn.Name, nsErr)
// 		return false, nsErr
// 	}

// 	var retErr error
// 	// handle resource quota
// 	rq := v1.ResourceQuota{
// 		ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-resource-quota", Namespace: nsName},
// 	}
// 	rqOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &rq, func() error {
// 		rq.Labels = r.updateTnResourceCommonLabels(rq.Labels)
// 		rq.Spec.Hard = forge.TenantResourceQuotaSpec(&tn.Status.Quota)

// 		return ctrl.SetControllerReference(tn, &rq, r.Scheme)
// 	})
// 	if err != nil {
// 		klog.Errorf("Unable to create or update resource quota for tenant %s -> %s", tn.Name, err)
// 		retErr = err
// 	}
// 	klog.Infof("Resource quota for tenant %s %s", tn.Name, rqOpRes)

// 	// handle roleBinding (instance management)
// 	rb := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-manage-instances", Namespace: nsName}}
// 	rbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &rb, func() error {
// 		r.updateTnRb(&rb, tn.Name)
// 		return ctrl.SetControllerReference(tn, &rb, r.Scheme)
// 	})
// 	if err != nil {
// 		klog.Errorf("Unable to create or update role binding for tenant %s -> %s", tn.Name, err)
// 		retErr = err
// 	}
// 	klog.Infof("Role binding for tenant %s %s", tn.Name, rbOpRes)

// 	// handle clusterRole (tenant access)
// 	crName := fmt.Sprintf("crownlabs-manage-%s", nsName)
// 	cr := rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: crName}}
// 	crOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &cr, func() error {
// 		r.updateTnCr(&cr, tn.Name)
// 		return ctrl.SetControllerReference(tn, &cr, r.Scheme)
// 	})
// 	if err != nil {
// 		klog.Errorf("Unable to create or update cluster role for tenant %s -> %s", tn.Name, err)
// 		retErr = err
// 	}
// 	klog.Infof("Cluster role for tenant %s %s", tn.Name, crOpRes)

// 	// handle clusterRoleBinding (tenant access)
// 	crbName := crName
// 	crb := rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: crbName}}
// 	crbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &crb, func() error {
// 		r.updateTnCrb(&crb, tn.Name, crName)
// 		return ctrl.SetControllerReference(tn, &crb, r.Scheme)
// 	})
// 	if err != nil {
// 		klog.Errorf("Unable to create or update cluster role binding for tenant %s -> %s", tn.Name, err)
// 		retErr = err
// 	}
// 	klog.Infof("Cluster role binding for tenant %s %s", tn.Name, crbOpRes)

// 	netPolDeny := netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-deny-ingress-traffic", Namespace: nsName}}
// 	npDOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &netPolDeny, func() error {
// 		r.updateTnNetPolDeny(&netPolDeny)
// 		return ctrl.SetControllerReference(tn, &netPolDeny, r.Scheme)
// 	})
// 	if err != nil {
// 		klog.Errorf("Unable to create or update deny network policy for tenant %s -> %s", tn.Name, err)
// 		retErr = err
// 	}
// 	klog.Infof("Deny network policy for tenant %s %s", tn.Name, npDOpRes)

// 	netPolAllow := netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-allow-trusted-ingress-traffic", Namespace: nsName}}
// 	npAOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &netPolAllow, func() error {
// 		r.updateTnNetPolAllow(&netPolAllow)
// 		return ctrl.SetControllerReference(tn, &netPolAllow, r.Scheme)
// 	})
// 	if err != nil {
// 		klog.Errorf("Unable to create or update allow network policy for tenant %s -> %s", tn.Name, err)
// 		retErr = err
// 	}
// 	klog.Infof("Allow network policy for tenant %s %s", tn.Name, npAOpRes)

// 	err = r.createOrUpdateTnPersonalNFSVolume(ctx, tn, nsName)
// 	if err != nil {
// 		klog.Errorf("Unable to create or update personal NFS volume for tenant %s -> %s", tn.Name, err)
// 		retErr = err
// 	}
// 	klog.Infof("Personal NFS volume for tenant %s", tn.Name)

// 	return true, retErr
// }

// // other stuff that need to be moved later on
// func (r *Reconciler) updateTnResourceCommonLabels(labels map[string]string) map[string]string {
// 	if labels == nil {
// 		labels = make(map[string]string, 1)
// 	}
// 	labels[r.TargetLabel.GetKey()] = r.TargetLabel.GetValue()
// 	labels["crownlabs.polito.it/managed-by"] = "tenant"
// 	return labels
// }
