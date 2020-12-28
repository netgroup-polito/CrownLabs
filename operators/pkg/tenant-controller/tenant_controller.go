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

package tenant_controller

import (
	"context"
	"fmt"
	"strings"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"

	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	KcA    *KcActor
}

// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=tenants/status,verbs=get;update;patch

// Reconcile reconciles the state of a tenant resource
func (r *TenantReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	var tn crownlabsv1alpha1.Tenant
	var userID *string
	if err := r.Get(ctx, req.NamespacedName, &tn); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error getting tenant %s on deletion", req.Name)
		klog.Error(err)
		return ctrl.Result{}, err
	} else if err != nil {
		// reconcile was triggered by a delete request
		if userID, _, err = r.KcA.getUserInfo(ctx, req.Name); err != nil {
			klog.Errorf("Error when checking if user %s existed for deletion", req.Name)
			klog.Error(err)
			return ctrl.Result{}, err
		} else if userID != nil {
			// userID != nil means user already existed when tenant was deleted, so need to delete user in keycloak
			if err = r.KcA.Client.DeleteUser(ctx, r.KcA.Token.AccessToken, r.KcA.TargetRealm, *userID); err != nil {
				klog.Errorf("Error when deleting user %s", req.Name)
				klog.Error(err)
				return ctrl.Result{}, err
			}
		}
		klog.Infof("Tenant %s resources deleted", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var retrigErr error = nil
	if tn.Status.Subscriptions == nil {
		tn.Status.Subscriptions = make(map[string]crownlabsv1alpha1.SubscriptionStatus)
	}

	klog.Infof("Reconciling tenant %s", req.Name)

	// namespace creation
	nsName := fmt.Sprintf("tenant-%s", tn.Name)
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}

	nsOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		updateTnNamespace(&ns, tn.Name)
		return ctrl.SetControllerReference(&tn, &ns, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update namespace of tenant %s", tn.Name)
		klog.Error(err)
		tn.Status.PersonalNamespace.Created = false
		tn.Status.PersonalNamespace.Name = ""
		retrigErr = err
	} else {
		klog.Infof("Namespace %s for tenant %s %s", nsName, req.Name, nsOpRes)
		tn.Status.PersonalNamespace.Created = true
		tn.Status.PersonalNamespace.Name = nsName
		if err := createOrUpdateTnClusterResources(ctx, r, tn.Name, nsName); err != nil {
			klog.Errorf("Error creating k8s resources for tenant %s", tn.Name)
			klog.Error(err)
			retrigErr = err
		} else {
			klog.Infof("Cluster resources for tenant %s have been correctly handled", tn.Name)
		}
	}

	// check validity of workspaces in tenant
	tnWorkspaceLabels := make(map[string]string)
	tenantExistingWorkspaces := []crownlabsv1alpha1.UserWorkspaceData{}
	tn.Status.FailingWorkspaces = []string{}
	for _, tnWs := range tn.Spec.Workspaces {
		wsLookupKey := types.NamespacedName{Name: tnWs.WorkspaceRef.Name, Namespace: ""}
		var ws crownlabsv1alpha1.Workspace

		if err := r.Get(ctx, wsLookupKey, &ws); err != nil {
			klog.Errorf("Error when checking if workspace %s exists in tenant %s", tnWs.WorkspaceRef.Name, tn.Name)
			klog.Error(err)
			retrigErr = err
			tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, tnWs.WorkspaceRef.Name)
		} else {
			wsLabelKey := fmt.Sprintf("%s%s", crownlabsv1alpha1.WorkspaceLabelPrefix, tnWs.WorkspaceRef.Name)
			tnWorkspaceLabels[wsLabelKey] = string(tnWs.Role)
			tenantExistingWorkspaces = append(tenantExistingWorkspaces, tnWs)
		}
	}

	// keycloak resources creation
	userID, currentUserEmail, err := r.KcA.getUserInfo(ctx, tn.Name)
	if err != nil {
		klog.Errorf("Error when checking if user %s existed for creation/update", tn.Name)
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrFailed
		retrigErr = err
	} else {
		if userID == nil {
			userID, err = r.KcA.createKcUser(ctx, tn.Name, tn.Spec.FirstName, tn.Spec.LastName, tn.Spec.Email)
		} else {
			err = r.KcA.updateKcUser(ctx, *userID, tn.Spec.FirstName, tn.Spec.LastName, tn.Spec.Email, *currentUserEmail != tn.Spec.Email)
		}
		if err != nil {
			klog.Errorf("Error when creating or updating user %s", tn.Name)
			klog.Error(err)
			retrigErr = err
			tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrFailed
		} else if err = r.KcA.updateUserRoles(ctx, genUserRoles(tenantExistingWorkspaces), *userID, "workspace-"); err != nil {
			klog.Errorf("Error when updating user roles of user %s", tn.Name)
			klog.Error(err)
			retrigErr = err
			tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrFailed
		} else {
			klog.Infof("Keycloak resources of user %s updated", tn.Name)
			tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrOk
		}
	}

	// place status value to ready if everything is fine, in other words, no need to reconcile
	tn.Status.Ready = retrigErr == nil

	if err := r.Status().Update(ctx, &tn); err != nil {
		// if status update fails, still try to reconcile later
		klog.Error("Unable to update status before exiting reconciler", err)
		retrigErr = err
	}
	if tn.Labels == nil {
		tn.Labels = make(map[string]string)
	} else {
		// need to do this after updating status cause the update will erase non-status changes
		cleanWorkspaceLabels(&tn.Labels)
	}
	for k, v := range tnWorkspaceLabels {
		tn.Labels[k] = v
	}
	// need to update resource to apply labels
	if err := r.Update(ctx, &tn); err != nil {
		// if status update fails, still try to reconcile later
		klog.Error("Unable to update resource before exiting reconciler", err)
		retrigErr = err
	}
	klog.Infof("Tenant %s reconciled", tn.Name)
	return ctrl.Result{}, retrigErr
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha1.Tenant{}).
		Complete(r)
}

// updateTnNamespace updates the tenant namespace
func updateTnNamespace(ns *v1.Namespace, tnName string) {
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels["crownlabs.polito.it/type"] = "tenant"
	ns.Labels["crownlabs.polito.it/name"] = tnName
	ns.Labels["crownlabs.polito.it/operator-selector"] = "production"
}

// genUserRoles maps the workspaces of a tenant to the need roles in keycloak
func genUserRoles(workspaces []crownlabsv1alpha1.UserWorkspaceData) []string {
	userRoles := make([]string, len(workspaces))
	// convert workspaces to actual keyloak role
	for i, ws := range workspaces {
		userRoles[i] = fmt.Sprintf("workspace-%s:%s", ws.WorkspaceRef.Name, ws.Role)
	}
	return userRoles
}

// cleanWorkspaceLabels removes all the labels of a workspace from a tenant
func cleanWorkspaceLabels(labels *map[string]string) {
	for k := range *labels {
		if strings.HasPrefix(k, crownlabsv1alpha1.WorkspaceLabelPrefix) {
			delete(*labels, k)
		}
	}
}

func createOrUpdateTnClusterResources(ctx context.Context, r *TenantReconciler, tnName, nsName string) error {
	// handle resource quota
	rq := v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-resource-quota", Namespace: nsName},
	}
	rqOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &rq, func() error {
		updateTnResQuota(&rq)
		return nil
	})
	if err != nil {
		klog.Errorf("Unable to create or update resource quota for tenant %s", tnName)
		klog.Error(err)
		return err
	}
	klog.Infof("Resource quota for tenant %s %s", tnName, rqOpRes)

	// handle roleBinding
	rb := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-manage-instances", Namespace: nsName}}
	rbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &rb, func() error {
		updateTnRb(&rb, tnName)
		return nil
	})
	if err != nil {
		klog.Errorf("Unable to create or update role binding for tenant %s", tnName)
		klog.Error(err)
		return err
	}
	klog.Infof("Role binding for tenant %s %s", tnName, rbOpRes)

	netPolDeny := netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-deny-ingress-traffic", Namespace: nsName}}
	npDOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &netPolDeny, func() error {
		updateTnNetPolDeny(&netPolDeny)
		return nil
	})
	if err != nil {
		klog.Errorf("Unable to create or update deny network policy for tenant %s", tnName)
		klog.Error(err)
		return err
	}
	klog.Infof("Deny network policy for tenant %s %s", tnName, npDOpRes)

	netPolAllow := netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-allow-trusted-ingress-traffic", Namespace: nsName}}
	npAOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &netPolAllow, func() error {
		updateTnNetPolAllow(&netPolAllow)
		return nil
	})
	if err != nil {
		klog.Errorf("Unable to create or update allow network policy for tenant %s", tnName)
		klog.Error(err)
		return err
	}
	klog.Infof("Allow network policy for tenant %s %s", tnName, npAOpRes)
	return nil
}

// updateTnResQuota updates the tenant resource quota
func updateTnResQuota(rq *v1.ResourceQuota) {

	resourceList := make(v1.ResourceList)

	resourceList["limits.cpu"] = *resource.NewQuantity(10, resource.DecimalSI)
	resourceList["limits.memory"] = *resource.NewQuantity(25*1024*1024*1024, resource.BinarySI)
	resourceList["requests.cpu"] = *resource.NewQuantity(10, resource.DecimalSI)
	resourceList["requests.memory"] = *resource.NewQuantity(25*1024*1024*1024, resource.BinarySI)
	resourceList["count/instances.crownlabs.polito.it"] = *resource.NewQuantity(5, resource.DecimalSI)

	rq.Spec.Hard = resourceList
}

func updateTnRb(rb *rbacv1.RoleBinding, tnName string) {
	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-instances", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "User", Name: tnName, APIGroup: "rbac.authorization.k8s.io"}}
}

func updateTnNetPolDeny(np *netv1.NetworkPolicy) {
	np.Spec.PodSelector.MatchLabels = make(map[string]string)
	np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}}}}
}

func updateTnNetPolAllow(np *netv1.NetworkPolicy) {
	np.Spec.PodSelector.MatchLabels = make(map[string]string)
	np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{{NamespaceSelector: &metav1.LabelSelector{
		MatchLabels: map[string]string{"crownlabs.polito.it/allow-instance-access": "true"},
	}}}}}
}
