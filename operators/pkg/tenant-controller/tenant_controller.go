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
	"regexp"
	"strings"
	"time"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

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
		klog.Errorf("Error when getting tenant %s before starting reconcile", req.Name)
		klog.Error(err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Tenant %s deleted", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var retrigErr error = nil
	if tn.Status.Subscriptions == nil {
		tn.Status.Subscriptions = make(map[string]crownlabsv1alpha1.SubscriptionStatus)
	}

	if !tn.ObjectMeta.DeletionTimestamp.IsZero() {
		klog.Infof("Processing deletion of tenant %s", tn.Name)
		if containsString(tn.ObjectMeta.Finalizers, crownlabsv1alpha1.TnOperatorFinalizerName) {
			// reconcile was triggered by a delete request
			if userID, _, err := r.KcA.getUserInfo(ctx, tn.Name); err != nil {
				klog.Errorf("Error when checking if user %s existed for deletion", tn.Name)
				klog.Error(err)
				retrigErr = err
			} else if userID != nil {
				// userID != nil means user exist in keycloak, so need to delete it
				if err = r.KcA.Client.DeleteUser(ctx, r.KcA.Token.AccessToken, r.KcA.TargetRealm, *userID); err != nil {
					klog.Errorf("Error when deleting user %s", tn.Name)
					klog.Error(err)
					retrigErr = err
				}
			}
			// remove finalizer from the tenant
			tn.ObjectMeta.Finalizers = removeString(tn.ObjectMeta.Finalizers, crownlabsv1alpha1.TnOperatorFinalizerName)
			if err := r.Update(context.Background(), &tn); err != nil {
				klog.Errorf("Error when removing tenant operator finalizer from tenant %s", tn.Name)
				klog.Error(err)
				retrigErr = err
			}
		}
		if retrigErr == nil {
			klog.Infof("Tenant %s ready for deletion", tn.Name)
		} else {
			klog.Errorf("Error when preparing tenant %s for deletion, need to retry", tn.Name)
		}
		return ctrl.Result{}, retrigErr
	}
	// tenant is NOT being deleted
	klog.Infof("Reconciling tenant %s", req.Name)

	// add tenant operator finalizer to tenant
	if !containsString(tn.ObjectMeta.Finalizers, crownlabsv1alpha1.TnOperatorFinalizerName) {
		tn.ObjectMeta.Finalizers = append(tn.ObjectMeta.Finalizers, crownlabsv1alpha1.TnOperatorFinalizerName)
		if err := r.Update(context.Background(), &tn); err != nil {
			klog.Errorf("Error when adding finalizer to tenant %s", tn.Name)
			klog.Error(err)
			retrigErr = err
		}
	}

	// namespace creation
	nsName := fmt.Sprintf("tenant-%s", strings.Replace(tn.Name, ".", "-", -1))

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
		if err := createOrUpdateTnClusterResources(ctx, r, &tn, nsName); err != nil {
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

	if err := updateTnLabels(&tn, tnWorkspaceLabels); err != nil {
		klog.Errorf("Unable to update label of tenant %s", tn.Name)
		klog.Error(err)
		retrigErr = err
	}

	// need to update resource to apply labels
	if err := r.Update(ctx, &tn); err != nil {
		// if status update fails, still try to reconcile later
		klog.Error("Unable to update resource before exiting reconciler", err)
		retrigErr = err
	}

	if retrigErr != nil {
		klog.Errorf("Tenant %s failed to reconcile", tn.Name)
		return ctrl.Result{}, retrigErr
	}

	// no retrigErr, need to normal reconcile later, so need to create random number and exit
	nextRequeSeconds, err := randomRange(3600, 7200) // need to use seconds value for interval 1h-2h to have resoultion to the second
	if err != nil {
		klog.Error("Error when generating random number for reque", err)
		return ctrl.Result{}, err
	}
	nextRequeDuration := time.Second * time.Duration(*nextRequeSeconds)
	klog.Infof("Tenant %s reconciled successfully, next in %s", tn.Name, nextRequeDuration)
	return ctrl.Result{RequeueAfter: nextRequeDuration}, nil
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha1.Tenant{}).
		Complete(r)
}

func updateTnLabels(tn *crownlabsv1alpha1.Tenant, tnWorkspaceLabels map[string]string) error {
	if tn.Labels == nil {
		tn.Labels = make(map[string]string)
	} else {
		// need to do this after updating status cause the update will erase non-status changes
		cleanWorkspaceLabels(&tn.Labels)
	}
	for k, v := range tnWorkspaceLabels {
		tn.Labels[k] = v
	}

	cleanedFirstName, err := cleanName(tn.Spec.FirstName)
	if err != nil {
		klog.Errorf("Error when cleaning first name of tenant %s", tn.Name)
		return err
	}
	cleanedLastName, err := cleanName(tn.Spec.LastName)
	if err != nil {
		klog.Errorf("Error when cleaning last name of tenant %s", tn.Name)
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
		klog.Errorf("Error when checking name %s", name)
		klog.Error(err)
		return nil, err
	} else if !ok {
		problemChars := make([]string, 0)
		for _, c := range name {
			if ok, err := regexp.MatchString(okRegex, string(c)); err != nil {
				klog.Errorf("Error when cleaning name %s at char %s", name, string(c))
				klog.Error(err)
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

func createOrUpdateTnClusterResources(ctx context.Context, r *TenantReconciler, tn *crownlabsv1alpha1.Tenant, nsName string) error {
	tnName := tn.Name

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

	// handle roleBinding (instance management)
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

	// handle clusterRole (tenant access)
	crName := fmt.Sprintf("crownlabs-manage-%s", nsName)
	cr := rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: crName}}
	crOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &cr, func() error {
		updateTnCr(&cr, tnName)
		return ctrl.SetControllerReference(tn, &cr, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update cluster role for tenant %s", tnName)
		klog.Error(err)
		return err
	}
	klog.Infof("Cluster role for tenant %s %s", tnName, crOpRes)

	// handle clusterRoleBinding (tenant access)
	crbName := crName
	crb := rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: crbName}}
	crbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &crb, func() error {
		updateTnCrb(&crb, tnName, crName)
		return ctrl.SetControllerReference(tn, &crb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update cluster role binding for tenant %s", tnName)
		klog.Error(err)
		return err
	}
	klog.Infof("Cluster role binding for tenant %s %s", tnName, crbOpRes)

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
	if rq.Labels == nil {
		rq.Labels = make(map[string]string, 1)
	}
	rq.Labels["crownlabs.polito.it/managed-by"] = "tenant"

	resourceList := make(v1.ResourceList)

	resourceList["limits.cpu"] = *resource.NewQuantity(10, resource.DecimalSI)
	resourceList["limits.memory"] = *resource.NewQuantity(25*1024*1024*1024, resource.BinarySI)
	resourceList["requests.cpu"] = *resource.NewQuantity(10, resource.DecimalSI)
	resourceList["requests.memory"] = *resource.NewQuantity(25*1024*1024*1024, resource.BinarySI)
	resourceList["count/instances.crownlabs.polito.it"] = *resource.NewQuantity(5, resource.DecimalSI)

	rq.Spec.Hard = resourceList
}

func updateTnRb(rb *rbacv1.RoleBinding, tnName string) {
	if rb.Labels == nil {
		rb.Labels = make(map[string]string, 1)
	}
	rb.Labels["crownlabs.polito.it/managed-by"] = "tenant"
	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-instances", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "User", Name: tnName, APIGroup: "rbac.authorization.k8s.io"}}
}

func updateTnCr(rb *rbacv1.ClusterRole, tnName string) {
	if rb.Labels == nil {
		rb.Labels = make(map[string]string, 1)
	}
	rb.Labels["crownlabs.polito.it/managed-by"] = "tenant"
	rb.Rules = []rbacv1.PolicyRule{{
		APIGroups:     []string{"crownlabs.polito.it"},
		Resources:     []string{"tenants"},
		ResourceNames: []string{tnName},
		Verbs:         []string{"get", "list", "watch"},
	}}
}

func updateTnCrb(rb *rbacv1.ClusterRoleBinding, tnName string, crName string) {
	if rb.Labels == nil {
		rb.Labels = make(map[string]string, 1)
	}
	rb.Labels["crownlabs.polito.it/managed-by"] = "tenant"
	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: crName, APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "User", Name: tnName, APIGroup: "rbac.authorization.k8s.io"}}
}

func updateTnNetPolDeny(np *netv1.NetworkPolicy) {
	if np.Labels == nil {
		np.Labels = make(map[string]string, 1)
	}
	np.Labels["crownlabs.polito.it/managed-by"] = "tenant"
	np.Spec.PodSelector.MatchLabels = make(map[string]string)
	np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}}}}
}

func updateTnNetPolAllow(np *netv1.NetworkPolicy) {
	if np.Labels == nil {
		np.Labels = make(map[string]string, 1)
	}
	np.Labels["crownlabs.polito.it/managed-by"] = "tenant"
	np.Spec.PodSelector.MatchLabels = make(map[string]string)
	np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{{NamespaceSelector: &metav1.LabelSelector{
		MatchLabels: map[string]string{"crownlabs.polito.it/allow-instance-access": "true"},
	}}}}}
}
