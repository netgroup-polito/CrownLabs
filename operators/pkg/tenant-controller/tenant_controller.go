// Copyright 2020-2022 Politecnico di Torino
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
package tenant_controller

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

const (
	// NoWorkspacesLabel -> label to be set (to true) when no workspaces are associated to the tenant.
	NoWorkspacesLabel = "crownlabs.polito.it/no-workspaces"
)

// TenantReconciler reconciles a Tenant object.
type TenantReconciler struct {
	client.Client
	Scheme                   *runtime.Scheme
	KcA                      *KcActor
	NcA                      NcHandler
	TargetLabelKey           string
	TargetLabelValue         string
	SandboxClusterRole       string
	Concurrency              int
	RequeueTimeMinimum       time.Duration
	RequeueTimeMaximum       time.Duration
	TenantWorkspaceKeepAlive time.Duration

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// Reconcile reconciles the state of a tenant resource.
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "tenant", req.NamespacedName.Name)
	ctx = ctrl.LoggerInto(ctx, log)
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	var tn crownlabsv1alpha2.Tenant
	if err := r.Get(ctx, req.NamespacedName, &tn); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting tenant %s before starting reconcile -> %s", req.Name, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Tenant %s deleted", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if tn.Labels[r.TargetLabelKey] != r.TargetLabelValue {
		// if entered here it means that is in the reconcile
		// which has been requed after
		// the last successful one with the old target label
		return ctrl.Result{}, nil
	}

	var retrigErr error
	if tn.Status.Subscriptions == nil {
		// make initial len is 2 (keycloak and nextcloud)
		tn.Status.Subscriptions = make(map[string]crownlabsv1alpha2.SubscriptionStatus, 2)
	}

	if !tn.ObjectMeta.DeletionTimestamp.IsZero() {
		klog.Infof("Processing deletion of tenant %s", tn.Name)
		if ctrlUtil.ContainsFinalizer(&tn, crownlabsv1alpha2.TnOperatorFinalizerName) {
			// reconcile was triggered by a delete request
			if err := r.handleDeletion(ctx, tn.Name); err != nil {
				klog.Errorf("error when deleting external resources on tenant %s deletion -> %s", tn.Name, err)
				retrigErr = err
			}
			// can remove the finalizer from the tenant if the eternal resources have been successfully deleted
			if retrigErr == nil {
				// remove finalizer from the tenant
				ctrlUtil.RemoveFinalizer(&tn, crownlabsv1alpha2.TnOperatorFinalizerName)
				if err := r.Update(context.Background(), &tn); err != nil {
					klog.Errorf("Error when removing tenant operator finalizer from tenant %s -> %s", tn.Name, err)
					tnOpinternalErrors.WithLabelValues("tenant", "self-update").Inc()
					retrigErr = err
				}
			}
		}
		if retrigErr == nil {
			klog.Infof("Tenant %s ready for deletion", tn.Name)
		} else {
			klog.Errorf("Error when preparing tenant %s for deletion, need to retry -> %s", tn.Name, retrigErr)
		}
		return ctrl.Result{}, retrigErr
	}
	// tenant is NOT being deleted
	klog.Infof("Reconciling tenant %s", tn.Name)

	// convert the email to lower-case, to prevent issues with keycloak
	// This modification is not persisted (on purpose) in the tenant resource, since the
	// update is performed after an update of the status, which restores the original spec.
	tn.Spec.Email = strings.ToLower(tn.Spec.Email)

	// add tenant operator finalizer to tenant
	if !ctrlUtil.ContainsFinalizer(&tn, crownlabsv1alpha2.TnOperatorFinalizerName) {
		ctrlUtil.AddFinalizer(&tn, crownlabsv1alpha2.TnOperatorFinalizerName)
		if err := r.Update(context.Background(), &tn); err != nil {
			klog.Errorf("Error when adding finalizer to tenant %s -> %s ", tn.Name, err)
			retrigErr = err
		}
	}

	tenantExistingWorkspaces, workspaces, err := r.checkValidWorkspaces(ctx, &tn)

	if err != nil {
		retrigErr = err
	}

	// Determine if the personal namespace should be deleted

	nsName := fmt.Sprintf("tenant-%s", strings.ReplaceAll(tn.Name, ".", "-"))

	// Test if namespace has been open for too long; attempt to delete if there are no instances inside
	keepNsOpen, err := r.enforceNamespaceKeepAliveOrDelete(ctx, &tn, nsName, r.TenantWorkspaceKeepAlive)
	if err != nil {
		klog.Errorf("Error in r.List: unable to capture instances in tenant workspace %s -> %s", nsName, err)
	}

	// update resource quota in the status of the tenant after checking validity of workspaces.
	tn.Status.Quota = forge.TenantResourceList(workspaces, tn.Spec.Quota)

	var nsOk bool
	nsOk, err = r.enforceClusterResources(ctx, &tn, nsName, keepNsOpen)
	if err != nil {
		klog.Errorf("Error when enforcing cluster resources for tenant %s -> %s", tn.Name, err)
	}

	if err = r.handleKeycloakSubscription(ctx, &tn, tenantExistingWorkspaces); err != nil {
		klog.Errorf("Error when updating keycloak subscription for tenant %s -> %s", tn.Name, err)
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrFailed
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("tenant", "keycloak").Inc()
	} else {
		klog.Infof("Keycloak subscription for tenant %s updated", tn.Name)
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrOk
	}

	if keepNsOpen { // Only handle nextcloud subscription if namespace exists
		if nsOk {
			if err = r.handleNextcloudSubscription(ctx, &tn, nsName); err != nil {
				klog.Errorf("Error when updating nextcloud subscription for tenant %s -> %s", tn.Name, err)
				tn.Status.Subscriptions["nextcloud"] = crownlabsv1alpha2.SubscrFailed
				retrigErr = err
				tnOpinternalErrors.WithLabelValues("tenant", "nextcloud").Inc()
			} else {
				klog.Infof("Nextcloud subscription for tenant %s updated", tn.Name)
				tn.Status.Subscriptions["nextcloud"] = crownlabsv1alpha2.SubscrOk
			}
		} else {
			klog.Errorf("Could not handle nextcloud subscription for tenant %s -> namespace update for secret gave error", tn.Name)
			tn.Status.Subscriptions["nextcloud"] = crownlabsv1alpha2.SubscrFailed
		}
	} else {
		klog.Infof("Nextcloud subscription for tenant %s not updated because namespace has been temporarily deleted", tn.Name)
	}

	if err = r.EnforceSandboxResources(ctx, &tn); err != nil {
		klog.Error("Failed checking sandbox for tenant %s -> %s", tn.Name, err)
		tn.Status.SandboxNamespace.Created = false
		tnOpinternalErrors.WithLabelValues("tenant", "sandbox-resources").Inc()
		return ctrl.Result{}, err
	}

	// place status value to ready if everything is fine, in other words, no need to reconcile
	tn.Status.Ready = retrigErr == nil

	if err = r.Status().Update(ctx, &tn); err != nil {
		// if status update fails, still try to reconcile later
		klog.Errorf("Unable to update status of tenant %s before exiting reconciler -> %s", tn.Name, err)
		retrigErr = err
	}

	if err = updateTnLabels(&tn, tenantExistingWorkspaces); err != nil {
		klog.Errorf("Unable to update label of tenant %s -> %s", tn.Name, err)
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("tenant", "self-update").Inc()
	}

	// need to update resource to apply labels
	if err = r.Update(ctx, &tn); err != nil {
		// if status update fails, still try to reconcile later
		klog.Errorf("Unable to update tenant %s before exiting reconciler -> %s", tn.Name, err)
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("tenant", "self-update").Inc()
	}

	if retrigErr != nil {
		klog.Errorf("Tenant %s failed to reconcile -> %s", tn.Name, retrigErr)
		return ctrl.Result{}, retrigErr
	}

	// no retrigErr, need to normal reconcile later, so need to create random number and exit
	nextRequeueSeconds, err := randomRange(int(r.RequeueTimeMinimum.Seconds()), int(r.RequeueTimeMaximum.Seconds()))
	if err != nil {
		klog.Errorf("Error when generating random number for requeue -> %s", err)
		tnOpinternalErrors.WithLabelValues("tenant", "self-update").Inc()
		return ctrl.Result{}, err
	}
	nextRequeueDuration := time.Second * time.Duration(*nextRequeueSeconds)
	klog.Infof("Tenant %s reconciled successfully, next in %s", tn.Name, nextRequeueDuration)
	return ctrl.Result{RequeueAfter: nextRequeueDuration}, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha2.Tenant{}, builder.WithPredicates(labelSelectorPredicate(r.TargetLabelKey, r.TargetLabelValue))).
		// owns the secret related to the nextcloud credentials, to allow new password generation in case tenant has a problem with nextcloud
		Owns(&v1.Secret{}).
		Owns(&v1.Namespace{}).
		Owns(&v1.ResourceQuota{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&netv1.NetworkPolicy{}).
		Watches(&source.Kind{Type: &crownlabsv1alpha1.Workspace{}},
			handler.EnqueueRequestsFromMapFunc(r.workspaceToEnrolledTenants)).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Concurrency,
		}).
		Complete(r)
}

// handleDeletion deletes external resources of a tenant using a fail-fast:false strategy.
func (r *TenantReconciler) handleDeletion(ctx context.Context, tnName string) error {
	var retErr error
	// delete keycloak user
	if userID, _, err := r.KcA.getUserInfo(ctx, tnName); err != nil {
		klog.Errorf("Error when checking if user %s existed for deletion -> %s", tnName, err)
		tnOpinternalErrors.WithLabelValues("tenant", "keycloak").Inc()
		retErr = err
	} else if userID != nil {
		// userID != nil means user exist in keycloak, so need to delete it
		if err = r.KcA.Client.DeleteUser(ctx, r.KcA.GetAccessToken(), r.KcA.TargetRealm, *userID); err != nil {
			klog.Errorf("Error when deleting user %s -> %s", tnName, err)
			tnOpinternalErrors.WithLabelValues("tenant", "keycloak").Inc()
			retErr = err
		}
	}
	// delete nextcloud user
	if err := r.NcA.DeleteUser(genNcUsername(tnName)); err != nil {
		klog.Errorf("Error when deleting nextcloud user for tenant %s -> %s", tnName, err)
		tnOpinternalErrors.WithLabelValues("tenant", "nextcloud").Inc()
		retErr = err
	}
	return retErr
}

// checkValidWorkspaces check validity of workspaces in tenant.
func (r *TenantReconciler) checkValidWorkspaces(ctx context.Context, tn *crownlabsv1alpha2.Tenant) ([]crownlabsv1alpha2.TenantWorkspaceEntry, []crownlabsv1alpha1.Workspace, error) {
	tenantExistingWorkspaces := []crownlabsv1alpha2.TenantWorkspaceEntry{}
	workspaces := []crownlabsv1alpha1.Workspace{}
	tn.Status.FailingWorkspaces = []string{}
	var err error
	// check every workspace of a tenant
	for _, tnWs := range tn.Spec.Workspaces {
		wsLookupKey := types.NamespacedName{Name: tnWs.Name}
		var ws crownlabsv1alpha1.Workspace
		if err = r.Get(ctx, wsLookupKey, &ws); err != nil {
			// if there was a problem, add the workspace to the status of the tenant
			klog.Errorf("Error when checking if workspace %s exists in tenant %s -> %s", tnWs.Name, tn.Name, err)
			tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, tnWs.Name)
			tnOpinternalErrors.WithLabelValues("tenant", "workspace-not-exist").Inc()
		} else {
			tenantExistingWorkspaces = append(tenantExistingWorkspaces, tnWs)
			workspaces = append(workspaces, ws)
		}
	}
	return tenantExistingWorkspaces, workspaces, err
}

// deleteClusterNamespace deletes the namespace for the tenant, if it fails then it returns an error.
func (r *TenantReconciler) deleteClusterNamespace(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) (err error) {
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}

	nsErr := utils.EnforceObjectAbsence(ctx, r.Client, &ns, "personal namespace")

	if nsErr != nil {
		klog.Errorf("Error when deleting namespace of tenant %s -> %s", tn.Name, nsErr)
	}

	return nsErr
}

// enforceNamespaceKeepAliveOrDelete deletes the namespace if it should be deleted.
func (r *TenantReconciler) enforceNamespaceKeepAliveOrDelete(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string, tenantWorkspaceKeepAlive time.Duration) (keepNsOpen bool, err error) {
	// We check to see if last login was more than tenantWorkspaceKeepAlive in the past:
	// if so, temporarily delete the namespace. We assume that a lastLogin of 0 occurs when a user is first created

	// Calculate time elapsed since lastLogin (now minus lastLogin in seconds)
	sPassed := time.Since(tn.Spec.LastLogin.Time)

	klog.Infof("Last login of tenant %s was %s ago", tn.Name, sPassed)

	// Attempt to get instances in current namespace
	list := &crownlabsv1alpha2.InstanceList{}

	if err := r.List(context.Background(), list, client.InNamespace(nsName)); err != nil {
		return true, err
	}

	if sPassed > tenantWorkspaceKeepAlive { // seconds
		klog.Infof("Over %s elapsed since last login of tenant %s: attempting to delete tenant namespace if not already deleted", tenantWorkspaceKeepAlive, tn.Name)
		if len(list.Items) == 0 {
			klog.Infof("No instances in %s: workspace can be deleted", nsName)
			keepNsOpen = false
		} else {
			klog.Infof("Instances in namespace %s. Namespace will not be deleted", nsName)
		}
	} else {
		klog.Infof("Under %s (limit) elapsed since last login of tenant %s: namespace is left as-is", tenantWorkspaceKeepAlive, tn.Name)
	}

	return keepNsOpen, nil
}

// Deletes namespace or updates the cluster resources.
func (r *TenantReconciler) enforceClusterResources(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string, keepNsOpen bool) (nsOk bool, err error) {
	nsOk = false // nsOk must be initialized for later use

	if keepNsOpen {
		nsOk, err = r.createOrUpdateClusterResources(ctx, tn, nsName)
		if nsOk {
			klog.Infof("Namespace %s for tenant %s updated", nsName, tn.Name)
			tn.Status.PersonalNamespace.Created = true
			tn.Status.PersonalNamespace.Name = nsName
			if err != nil {
				klog.Errorf("Unable to update cluster resource of tenant %s -> %s", tn.Name, err)
				tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
			}
			klog.Infof("Cluster resources for tenant %s updated", tn.Name)
		} else {
			klog.Errorf("Unable to update namespace of tenant %s -> %s", tn.Name, err)
			tn.Status.PersonalNamespace.Created = false
			tn.Status.PersonalNamespace.Name = ""
			tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
		}
	} else {
		nsErr := r.deleteClusterNamespace(ctx, tn, nsName)
		if nsErr == nil {
			klog.Infof("Namespace %s for tenant %s deleted if not already deleted", nsName, tn.Name)
			tn.Status.PersonalNamespace.Created = false
			tn.Status.PersonalNamespace.Name = ""
		} else {
			klog.Errorf("Unable to delete namespace of tenant %s -> %s", tn.Name, err)
			tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
		}
	}
	return nsOk, err
}

// createOrUpdateClusterResources creates the namespace for the tenant, if it succeeds it then tries to create the rest of the resources with a fail-fast:false strategy.
func (r *TenantReconciler) createOrUpdateClusterResources(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) (nsOk bool, err error) {
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}

	if _, nsErr := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		r.updateTnNamespace(&ns, tn.Name)
		return ctrl.SetControllerReference(tn, &ns, r.Scheme)
	}); nsErr != nil {
		klog.Errorf("Error when updating namespace of tenant %s -> %s", tn.Name, nsErr)
		return false, nsErr
	}

	var retErr error

	// handle resource quota
	rq := v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-resource-quota", Namespace: nsName},
	}
	rqOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &rq, func() error {
		rq.Labels = r.updateTnResourceCommonLabels(rq.Labels)
		rq.Spec.Hard = forge.TenantResourceQuotaSpec(&tn.Status.Quota)

		return ctrl.SetControllerReference(tn, &rq, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update resource quota for tenant %s -> %s", tn.Name, err)
		retErr = err
	}
	klog.Infof("Resource quota for tenant %s %s", tn.Name, rqOpRes)

	// handle roleBinding (instance management)
	rb := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-manage-instances", Namespace: nsName}}
	rbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &rb, func() error {
		r.updateTnRb(&rb, tn.Name)
		return ctrl.SetControllerReference(tn, &rb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update role binding for tenant %s -> %s", tn.Name, err)
		retErr = err
	}
	klog.Infof("Role binding for tenant %s %s", tn.Name, rbOpRes)

	// handle clusterRole (tenant access)
	crName := fmt.Sprintf("crownlabs-manage-%s", nsName)
	cr := rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: crName}}
	crOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &cr, func() error {
		r.updateTnCr(&cr, tn.Name)
		return ctrl.SetControllerReference(tn, &cr, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update cluster role for tenant %s -> %s", tn.Name, err)
		retErr = err
	}
	klog.Infof("Cluster role for tenant %s %s", tn.Name, crOpRes)

	// handle clusterRoleBinding (tenant access)
	crbName := crName
	crb := rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: crbName}}
	crbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &crb, func() error {
		r.updateTnCrb(&crb, tn.Name, crName)
		return ctrl.SetControllerReference(tn, &crb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update cluster role binding for tenant %s -> %s", tn.Name, err)
		retErr = err
	}
	klog.Infof("Cluster role binding for tenant %s %s", tn.Name, crbOpRes)

	netPolDeny := netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-deny-ingress-traffic", Namespace: nsName}}
	npDOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &netPolDeny, func() error {
		r.updateTnNetPolDeny(&netPolDeny)
		return ctrl.SetControllerReference(tn, &netPolDeny, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update deny network policy for tenant %s -> %s", tn.Name, err)
		retErr = err
	}
	klog.Infof("Deny network policy for tenant %s %s", tn.Name, npDOpRes)

	netPolAllow := netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-allow-trusted-ingress-traffic", Namespace: nsName}}
	npAOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &netPolAllow, func() error {
		r.updateTnNetPolAllow(&netPolAllow)
		return ctrl.SetControllerReference(tn, &netPolAllow, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update allow network policy for tenant %s -> %s", tn.Name, err)
		retErr = err
	}
	klog.Infof("Allow network policy for tenant %s %s", tn.Name, npAOpRes)
	return true, retErr
}

// updateTnNamespace updates the tenant namespace.
func (r *TenantReconciler) updateTnNamespace(ns *v1.Namespace, tnName string) {
	ns.Labels = r.updateTnResourceCommonLabels(ns.Labels)
	ns.Labels["crownlabs.polito.it/type"] = "tenant"
	ns.Labels["crownlabs.polito.it/name"] = tnName
	ns.Labels["crownlabs.polito.it/instance-resources-replication"] = "true"
}

func (r *TenantReconciler) updateTnRb(rb *rbacv1.RoleBinding, tnName string) {
	rb.Labels = r.updateTnResourceCommonLabels(rb.Labels)
	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-instances", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "User", Name: tnName, APIGroup: "rbac.authorization.k8s.io"}}
}

func (r *TenantReconciler) updateTnCr(rb *rbacv1.ClusterRole, tnName string) {
	rb.Labels = r.updateTnResourceCommonLabels(rb.Labels)
	rb.Rules = []rbacv1.PolicyRule{{
		APIGroups:     []string{"crownlabs.polito.it"},
		Resources:     []string{"tenants"},
		ResourceNames: []string{tnName},
		Verbs:         []string{"get", "list", "watch", "patch", "update"},
	}}
}

func (r *TenantReconciler) updateTnCrb(rb *rbacv1.ClusterRoleBinding, tnName, crName string) {
	rb.Labels = r.updateTnResourceCommonLabels(rb.Labels)
	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: crName, APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "User", Name: tnName, APIGroup: "rbac.authorization.k8s.io"}}
}

func (r *TenantReconciler) updateTnNetPolDeny(np *netv1.NetworkPolicy) {
	np.Labels = r.updateTnResourceCommonLabels(np.Labels)
	np.Spec.PodSelector.MatchLabels = make(map[string]string)
	np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}}}}
}

func (r *TenantReconciler) updateTnNetPolAllow(np *netv1.NetworkPolicy) {
	np.Labels = r.updateTnResourceCommonLabels(np.Labels)
	np.Spec.PodSelector.MatchLabels = make(map[string]string)
	np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{{NamespaceSelector: &metav1.LabelSelector{
		MatchLabels: map[string]string{"crownlabs.polito.it/allow-instance-access": "true"},
	}}}}}
}

func (r *TenantReconciler) handleKeycloakSubscription(ctx context.Context, tn *crownlabsv1alpha2.Tenant, tenantExistingWorkspaces []crownlabsv1alpha2.TenantWorkspaceEntry) error {
	// KcA could be nil for local testing skipping the keycloak subscription
	if r.KcA == nil {
		klog.Warningf("Skipping creation/update of tenant %v in keycloak", tn.GetName())
		return nil
	}
	userID, currentUserEmail, err := r.KcA.getUserInfo(ctx, tn.Name)
	if err != nil {
		klog.Errorf("Error when checking if keycloak user %s existed for creation/update -> %s", tn.Name, err)
		return err
	}
	if userID == nil {
		userID, err = r.KcA.createKcUser(ctx, tn.Name, tn.Spec.FirstName, tn.Spec.LastName, tn.Spec.Email)
	} else {
		err = r.KcA.updateKcUser(ctx, *userID, tn.Spec.FirstName, tn.Spec.LastName, tn.Spec.Email, !strings.EqualFold(*currentUserEmail, tn.Spec.Email))
	}
	if err != nil {
		klog.Errorf("Error when creating or updating keycloak user %s -> %s", tn.Name, err)
		return err
	} else if err = r.KcA.updateUserRoles(ctx, genKcUserRoleNames(tenantExistingWorkspaces), *userID, "workspace-"); err != nil {
		klog.Errorf("Error when updating user roles of user %s -> %s", tn.Name, err)
		return err
	}
	klog.Infof("Keycloak resources of user %s updated", tn.Name)
	return nil
}

// genKcUserRoleNames maps the workspaces of a tenant to the needed roles in keycloak.
func genKcUserRoleNames(workspaces []crownlabsv1alpha2.TenantWorkspaceEntry) []string {
	userRoles := make([]string, len(workspaces))
	// convert workspaces to actual keyloak role
	for i, ws := range workspaces {
		userRoles[i] = fmt.Sprintf("workspace-%s:%s", ws.Name, ws.Role)
	}
	return userRoles
}

func (r *TenantReconciler) handleNextcloudSubscription(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) error {
	// NcA could be nil for local testing skipping the nextcloud subscription
	if r.NcA == nil {
		klog.Warningf("Skipping creation/update of tenant %v in nextcloud", tn.GetName())
		return nil
	}
	// independently of the existence of the nexctloud secret for the nextcloud credentials of the user, need to know the displayname of the user, in order to update it if necessary
	ncUsername := genNcUsername(tn.Name)
	expectedDisplayname := genNcDisplayname(tn.Spec.FirstName, tn.Spec.LastName)
	ncSecret := v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "nextcloud-credentials", Namespace: nsName}}
	ncUserFound, ncDisplayname, err := r.NcA.GetUser(ncUsername)
	switch {
	case err != nil:
		klog.Errorf("Error when getting nextcloud user for tenant %s -> %s", tn.Name, err)
		return err
	case ncUserFound:
		if err = r.Get(ctx, types.NamespacedName{Name: ncSecret.Name, Namespace: ncSecret.Namespace}, &ncSecret); client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when getting nextcloud secret for tenant %s -> %s", tn.Name, err)
			return err
		} else if err != nil {
			ncPsw, errToken := generateToken()
			if errToken != nil {
				klog.Errorf("Error when generating nextcloud password of tenant %s -> %s", tn.Name, errToken)
				return errToken
			}
			if errToken = r.NcA.UpdateUserData(ncUsername, "password", *ncPsw); errToken != nil {
				klog.Errorf("Error when updating password of tenant %s -> %s", tn.Name, errToken)
				return errToken
			}
			// nextcloud secret not found, need to create it
			ncSecOpRes, errCreateSec := ctrl.CreateOrUpdate(ctx, r.Client, &ncSecret, func() error {
				r.updateTnNcSecret(&ncSecret, ncUsername, *ncPsw)
				return ctrl.SetControllerReference(tn, &ncSecret, r.Scheme)
			})
			if errCreateSec != nil {
				klog.Errorf("Unable to create or update nexcloud secret for tenant %s -> %s", tn.Name, errCreateSec)
				return errCreateSec
			}
			klog.Infof("Nextcloud secret for tenant %s %s", tn.Name, ncSecOpRes)

			return nil
		}
		if *ncDisplayname != expectedDisplayname {
			if err = r.NcA.UpdateUserData(ncUsername, "displayname", expectedDisplayname); err != nil {
				klog.Errorf("Error when updating displayname of tenant %s -> %s", tn.Name, err)
				return err
			}
		}
		return nil
	default:
		klog.Infof("Nextcloud user for %s not found", tn.Name)
		ncPsw, err := generateToken()
		if err != nil {
			klog.Errorf("Error when generating nextcloud password of tenant %s -> %s", tn.Name, err)
			return err
		}
		if err = r.NcA.CreateUser(ncUsername, *ncPsw, genNcDisplayname(tn.Spec.FirstName, tn.Spec.LastName)); err != nil {
			klog.Errorf("Error when creating nextcloud user for tenant %s -> %s", tn.Name, err)
			return err
		}
		ncSecretOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &ncSecret, func() error {
			r.updateTnNcSecret(&ncSecret, ncUsername, *ncPsw)
			return ctrl.SetControllerReference(tn, &ncSecret, r.Scheme)
		})
		if err != nil {
			klog.Errorf("Unable to create or update nextcloud secret of tenant %s  -> %s", tn.Name, err)
			return err
		}
		klog.Infof("Nextcloud secret for tenant %s %s", tn.Name, ncSecretOpRes)
		return nil
	}
}

func genNcUsername(tnName string) string {
	return fmt.Sprintf("keycloak-%s", tnName)
}

func genNcDisplayname(firstName, lastName string) string {
	return fmt.Sprintf("%s %s", firstName, lastName)
}

func (r *TenantReconciler) updateTnNcSecret(sec *v1.Secret, username, password string) {
	sec.Labels = r.updateTnResourceCommonLabels(sec.Labels)

	sec.Type = v1.SecretTypeOpaque
	sec.Data = make(map[string][]byte, 2)
	sec.Data["username"] = []byte(username)
	sec.Data["password"] = []byte(password)
}

func updateTnLabels(tn *crownlabsv1alpha2.Tenant, tenantExistingWorkspaces []crownlabsv1alpha2.TenantWorkspaceEntry) error {
	if tn.Labels == nil {
		tn.Labels = map[string]string{}
	} else {
		cleanWorkspaceLabels(tn.Labels)
	}
	for _, wsData := range tenantExistingWorkspaces {
		wsLabelKey := fmt.Sprintf("%s%s", crownlabsv1alpha2.WorkspaceLabelPrefix, wsData.Name)
		tn.Labels[wsLabelKey] = string(wsData.Role)
	}
	// label for users without workspaces
	if len(tenantExistingWorkspaces) == 0 {
		tn.Labels[NoWorkspacesLabel] = "true"
	} else {
		delete(tn.Labels, NoWorkspacesLabel)
	}

	tn.Labels["crownlabs.polito.it/first-name"] = cleanName(tn.Spec.FirstName)
	tn.Labels["crownlabs.polito.it/last-name"] = cleanName(tn.Spec.LastName)
	return nil
}

func cleanName(name string) string {
	okRegex := regexp.MustCompile("^[a-zA-Z0-9_]+$")
	name = strings.ReplaceAll(name, " ", "_")

	if !okRegex.MatchString(name) {
		problemChars := make([]string, 0)
		for _, c := range name {
			if !okRegex.MatchString(string(c)) {
				problemChars = append(problemChars, string(c))
			}
		}
		for _, v := range problemChars {
			name = strings.Replace(name, v, "", 1)
		}
	}

	return strings.Trim(name, "_")
}

// cleanWorkspaceLabels removes all the labels of a workspace from a tenant.
func cleanWorkspaceLabels(labels map[string]string) {
	for k := range labels {
		if strings.HasPrefix(k, crownlabsv1alpha2.WorkspaceLabelPrefix) {
			delete(labels, k)
		}
	}
}

func (r *TenantReconciler) updateTnResourceCommonLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 1)
	}
	labels[r.TargetLabelKey] = r.TargetLabelValue
	labels["crownlabs.polito.it/managed-by"] = "tenant"
	return labels
}

func (r *TenantReconciler) workspaceToEnrolledTenants(o client.Object) []reconcile.Request {
	var enqueues []reconcile.Request
	var tenants crownlabsv1alpha2.TenantList
	if err := r.List(context.Background(), &tenants, client.HasLabels{
		fmt.Sprintf("%s%s", crownlabsv1alpha2.WorkspaceLabelPrefix, o.GetName()),
	}); err != nil {
		klog.Errorf("Error when retrieving tenants enrolled in %s -> %s", o.GetName(), err)
		return nil
	}
	for idx := range tenants.Items {
		enqueues = append(enqueues, reconcile.Request{NamespacedName: types.NamespacedName{Name: tenants.Items[idx].GetName()}})
	}
	return enqueues
}
