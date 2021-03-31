/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package tenant_controller groups the functionalities related to the Tenant controller.
package tenant_controller

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"

	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// TenantReconciler reconciles a Tenant object.
type TenantReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	KcA              *KcActor
	NcA              NcHandler
	TargetLabelKey   string
	TargetLabelValue string

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// Reconcile reconciles the state of a tenant resource.
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	var tn crownlabsv1alpha1.Tenant
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

	var retrigErr error = nil
	if tn.Status.Subscriptions == nil {
		// make initial len is 2 (keycloak and nextcloud)
		tn.Status.Subscriptions = make(map[string]crownlabsv1alpha1.SubscriptionStatus, 2)
	}

	if !tn.ObjectMeta.DeletionTimestamp.IsZero() {
		klog.Infof("Processing deletion of tenant %s", tn.Name)
		if ctrlUtil.ContainsFinalizer(&tn, crownlabsv1alpha1.TnOperatorFinalizerName) {
			// reconcile was triggered by a delete request
			if err := r.handleDeletion(ctx, tn.Name); err != nil {
				klog.Errorf("error when deleting external resources on tenant %s deletion -> %s", tn.Name, err)
				retrigErr = err
			}
			// can remove the finalizer from the tenant if the eternal resources have been successfully deleted
			if retrigErr == nil {
				// remove finalizer from the tenant
				ctrlUtil.RemoveFinalizer(&tn, crownlabsv1alpha1.TnOperatorFinalizerName)
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

	// add tenant operator finalizer to tenant
	if !ctrlUtil.ContainsFinalizer(&tn, crownlabsv1alpha1.TnOperatorFinalizerName) {
		ctrlUtil.AddFinalizer(&tn, crownlabsv1alpha1.TnOperatorFinalizerName)
		if err := r.Update(context.Background(), &tn); err != nil {
			klog.Errorf("Error when adding finalizer to tenant %s -> %s ", tn.Name, err)
			retrigErr = err
		}
	}

	nsName := fmt.Sprintf("tenant-%s", strings.ReplaceAll(tn.Name, ".", "-"))
	nsOk, err := r.createOrUpdateClusterResources(ctx, &tn, nsName)
	if nsOk {
		klog.Infof("Namespace %s for tenant %s updated", nsName, tn.Name)
		tn.Status.PersonalNamespace.Created = true
		tn.Status.PersonalNamespace.Name = nsName
		if err != nil {
			klog.Errorf("Unable to update cluster resource of tenant %s -> %s", tn.Name, err)
			retrigErr = err
			tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
		}
		klog.Infof("Cluster resourcess for tenant %s updated", tn.Name)
	} else {
		klog.Errorf("Unable to update namespace of tenant %s -> %s", tn.Name, err)
		tn.Status.PersonalNamespace.Created = false
		tn.Status.PersonalNamespace.Name = ""
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
	}

	// check validity of workspaces in tenant
	tenantExistingWorkspaces := []crownlabsv1alpha1.TenantWorkspaceEntry{}
	tn.Status.FailingWorkspaces = []string{}
	// check every workspace of a tenant
	for _, tnWs := range tn.Spec.Workspaces {
		wsLookupKey := types.NamespacedName{Name: tnWs.WorkspaceRef.Name, Namespace: ""}
		var ws crownlabsv1alpha1.Workspace
		if err = r.Get(ctx, wsLookupKey, &ws); err != nil {
			// if there was a problem, add the workspace to the status of the tenant
			klog.Errorf("Error when checking if workspace %s exists in tenant %s -> %s", tnWs.WorkspaceRef.Name, tn.Name, err)
			retrigErr = err
			tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, tnWs.WorkspaceRef.Name)
			tnOpinternalErrors.WithLabelValues("tenant", "workspace-not-exist").Inc()
		} else {
			tenantExistingWorkspaces = append(tenantExistingWorkspaces, tnWs)
		}
	}

	if err = r.handleKeycloakSubscription(ctx, &tn, tenantExistingWorkspaces); err != nil {
		klog.Errorf("Error when updating keycloak subscription for tenant %s -> %s", tn.Name, err)
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrFailed
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("tenant", "keycloak").Inc()
	} else {
		klog.Infof("Keycloak subscription for tenant %s updated", tn.Name)
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrOk
	}

	if nsOk {
		if err = r.handleNextcloudSubscription(ctx, &tn, nsName); err != nil {
			klog.Errorf("Error when updating nextcloud subscription for tenant %s -> %s", tn.Name, err)
			tn.Status.Subscriptions["nextcloud"] = crownlabsv1alpha1.SubscrFailed
			retrigErr = err
			tnOpinternalErrors.WithLabelValues("tenant", "nextcloud").Inc()
		} else {
			klog.Infof("Nextcloud subscription for tenant %s updated", tn.Name)
			tn.Status.Subscriptions["nextcloud"] = crownlabsv1alpha1.SubscrOk
		}
	} else {
		klog.Errorf("Could not handle nextcloud subscription for tenant %s -> namespace update for secret gave error", tn.Name)
		tn.Status.Subscriptions["nextcloud"] = crownlabsv1alpha1.SubscrFailed
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
	nextRequeSeconds, err := randomRange(3600, 7200) // need to use seconds value for interval 1h-2h to have resolution to the second
	if err != nil {
		klog.Errorf("Error when generating random number for reque -> %s", err)
		tnOpinternalErrors.WithLabelValues("tenant", "self-update").Inc()
		return ctrl.Result{}, err
	}
	nextRequeDuration := time.Second * time.Duration(*nextRequeSeconds)
	klog.Infof("Tenant %s reconciled successfully, next in %s", tn.Name, nextRequeDuration)
	return ctrl.Result{RequeueAfter: nextRequeDuration}, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha1.Tenant{}).
		WithEventFilter(labelSelectorPredicate(r.TargetLabelKey, r.TargetLabelValue)).
		// owns the secret related to the nextcloud credentials, to allow new password generation in case tenant has a problem with nextcloud
		Owns(&v1.Secret{}).
		Owns(&v1.Namespace{}).
		Owns(&v1.ResourceQuota{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&netv1.NetworkPolicy{}).
		Complete(r)
}

// handleDeletion deletes external resources of a tenant using a fail-fast:false strategy.
func (r *TenantReconciler) handleDeletion(ctx context.Context, tnName string) error {
	var retErr error = nil
	// delete keycloak user
	if userID, _, err := r.KcA.getUserInfo(ctx, tnName); err != nil {
		klog.Errorf("Error when checking if user %s existed for deletion -> %s", tnName, err)
		tnOpinternalErrors.WithLabelValues("tenant", "keycloak").Inc()
		retErr = err
	} else if userID != nil {
		// userID != nil means user exist in keycloak, so need to delete it
		if err = r.KcA.Client.DeleteUser(ctx, r.KcA.Token.AccessToken, r.KcA.TargetRealm, *userID); err != nil {
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

// createOrUpdateClusterResources creates the namespace for the tenant, if it succeeds it then tries to create the rest of the resources with a fail-fast:false strategy.
func (r *TenantReconciler) createOrUpdateClusterResources(ctx context.Context, tn *crownlabsv1alpha1.Tenant, nsName string) (nsOk bool, err error) {
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}

	if _, nsErr := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		r.updateTnNamespace(&ns, tn.Name)
		return ctrl.SetControllerReference(tn, &ns, r.Scheme)
	}); nsErr != nil {
		klog.Errorf("Error when updating namespace of tenant %s -> %s", tn.Name, nsErr)
		return false, nsErr
	}

	var retErr error = nil
	// handle resource quota
	rq := v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-resource-quota", Namespace: nsName},
	}
	rqOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &rq, func() error {
		r.updateTnResQuota(&rq)
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
}

// updateTnResQuota updates the tenant resource quota.
func (r *TenantReconciler) updateTnResQuota(rq *v1.ResourceQuota) {
	rq.Labels = r.updateTnResourceCommonLabels(rq.Labels)

	resourceList := make(v1.ResourceList)

	resourceList["limits.cpu"] = *resource.NewQuantity(15, resource.DecimalSI)
	resourceList["limits.memory"] = *resource.NewQuantity(25*1024*1024*1024, resource.BinarySI)
	resourceList["requests.cpu"] = *resource.NewQuantity(10, resource.DecimalSI)
	resourceList["requests.memory"] = *resource.NewQuantity(25*1024*1024*1024, resource.BinarySI)
	resourceList["count/instances.crownlabs.polito.it"] = *resource.NewQuantity(5, resource.DecimalSI)

	rq.Spec.Hard = resourceList
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
		Verbs:         []string{"get", "list", "watch"},
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

func (r *TenantReconciler) handleKeycloakSubscription(ctx context.Context, tn *crownlabsv1alpha1.Tenant, tenantExistingWorkspaces []crownlabsv1alpha1.TenantWorkspaceEntry) error {
	userID, currentUserEmail, err := r.KcA.getUserInfo(ctx, tn.Name)
	if err != nil {
		klog.Errorf("Error when checking if keycloak user %s existed for creation/update -> %s", tn.Name, err)
		return err
	}
	if userID == nil {
		userID, err = r.KcA.createKcUser(ctx, tn.Name, tn.Spec.FirstName, tn.Spec.LastName, tn.Spec.Email)
	} else {
		err = r.KcA.updateKcUser(ctx, *userID, tn.Spec.FirstName, tn.Spec.LastName, tn.Spec.Email, *currentUserEmail != tn.Spec.Email)
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
func genKcUserRoleNames(workspaces []crownlabsv1alpha1.TenantWorkspaceEntry) []string {
	userRoles := make([]string, len(workspaces))
	// convert workspaces to actual keyloak role
	for i, ws := range workspaces {
		userRoles[i] = fmt.Sprintf("workspace-%s:%s", ws.WorkspaceRef.Name, ws.Role)
	}
	return userRoles
}

func (r *TenantReconciler) handleNextcloudSubscription(ctx context.Context, tn *crownlabsv1alpha1.Tenant, nsName string) error {
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

func updateTnLabels(tn *crownlabsv1alpha1.Tenant, tenantExistingWorkspaces []crownlabsv1alpha1.TenantWorkspaceEntry) error {
	if tn.Labels == nil {
		// the len is 1 for each workspace plus the 2 for firstName and lastName
		tn.Labels = make(map[string]string, len(tenantExistingWorkspaces)+2)
	} else {
		cleanWorkspaceLabels(tn.Labels)
	}
	for _, wsData := range tenantExistingWorkspaces {
		wsLabelKey := fmt.Sprintf("%s%s", crownlabsv1alpha1.WorkspaceLabelPrefix, wsData.WorkspaceRef.Name)
		tn.Labels[wsLabelKey] = string(wsData.Role)
	}

	cleanedFirstName, err := cleanName(tn.Spec.FirstName)
	if err != nil {
		klog.Errorf("Error when cleaning first name of tenant %s -> %s", tn.Name, err)
		return err
	}
	cleanedLastName, err := cleanName(tn.Spec.LastName)
	if err != nil {
		klog.Errorf("Error when cleaning last name of tenant %s -> %s", tn.Name, err)
		return err
	}
	tn.Labels["crownlabs.polito.it/first-name"] = *cleanedFirstName
	tn.Labels["crownlabs.polito.it/last-name"] = *cleanedLastName
	return nil
}

func cleanName(name string) (*string, error) {
	okRegex := "^[a-zA-Z0-9_]+$"
	name = strings.ReplaceAll(name, " ", "_")
	ok, err := regexp.MatchString(okRegex, name)
	if err != nil {
		klog.Errorf("Error when checking name %s for cleaning -> %s", name, err)
		return nil, err
	} else if !ok {
		problemChars := make([]string, 2)
		for _, c := range name {
			if ok, err := regexp.MatchString(okRegex, string(c)); err != nil {
				klog.Errorf("Error when cleaning name %s at char %s -> %s", name, string(c), err)
				return nil, err
			} else if !ok {
				problemChars = append(problemChars, string(c))
			}
		}
		for _, v := range problemChars {
			name = strings.Replace(name, v, "", 1)
		}
	}
	return &name, nil
}

// cleanWorkspaceLabels removes all the labels of a workspace from a tenant.
func cleanWorkspaceLabels(labels map[string]string) {
	for k := range labels {
		if strings.HasPrefix(k, crownlabsv1alpha1.WorkspaceLabelPrefix) {
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
