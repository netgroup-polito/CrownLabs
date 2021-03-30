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
	"time"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

// WorkspaceReconciler reconciles a Workspace object.
type WorkspaceReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	KcA              *KcActor
	TargetLabelKey   string
	TargetLabelValue string

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// Reconcile reconciles the state of a workspace resource.
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	var ws crownlabsv1alpha1.Workspace

	if err := r.Get(ctx, req.NamespacedName, &ws); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting workspace %s before starting reconcile -> %s", ws.Name, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Workspace %s deleted", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if ws.Labels[r.TargetLabelKey] != r.TargetLabelValue {
		// if entered here it means that is in the reconcile
		// which has been requed after
		// the last successful one with the old target label
		return ctrl.Result{}, nil
	}

	var retrigErr error = nil
	if !ws.ObjectMeta.DeletionTimestamp.IsZero() {
		klog.Infof("Processing deletion of workspace %s", ws.Name)
		// workspace is being deleted
		if ctrlUtil.ContainsFinalizer(&ws, crownlabsv1alpha1.TnOperatorFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.handleDeletion(ctx, ws.Name, ws.Spec.PrettyName); err != nil {
				klog.Errorf("Error when deleting resources handled by workspace  %s -> %s", ws.Name, err)
				retrigErr = err
			}
			// can remove the finalizer from the workspace if the eternal resources have been successfully deleted
			if retrigErr != nil {
				// remove finalizer from the workspace
				ctrlUtil.RemoveFinalizer(&ws, crownlabsv1alpha1.TnOperatorFinalizerName)
				if err := r.Update(context.Background(), &ws); err != nil {
					klog.Errorf("Error when removing tenant operator finalizer from workspace %s -> %s", ws.Name, err)
					retrigErr = err
					tnOpinternalErrors.WithLabelValues("workspace", "self-update").Inc()
				}
			}
		}
		if retrigErr == nil {
			klog.Infof("Workspace %s ready for deletion", ws.Name)
		} else {
			klog.Errorf("Error when preparing workspace %s for deletion, need to retry -> %s", ws.Name, retrigErr)
		}
		return ctrl.Result{}, retrigErr
	}

	// workspace is NOT being deleted
	klog.Infof("Reconciling workspace %s", ws.Name)

	// add tenant operator finalizer to workspace
	if !ctrlUtil.ContainsFinalizer(&ws, crownlabsv1alpha1.TnOperatorFinalizerName) {
		ctrlUtil.AddFinalizer(&ws, crownlabsv1alpha1.TnOperatorFinalizerName)
		if err := r.Update(context.Background(), &ws); err != nil {
			klog.Errorf("Error when adding finalizer to workspace %s -> %s", ws.Name, err)
			retrigErr = err
		}
	}

	nsName := fmt.Sprintf("workspace-%s", ws.Name)

	nsOk, err := r.createOrUpdateClusterResources(ctx, &ws, nsName)
	if nsOk {
		klog.Infof("Namespace %s for tenant %s updated", nsName, ws.Name)
		ws.Status.Namespace.Created = true
		ws.Status.Namespace.Name = nsName
		if err != nil {
			klog.Errorf("Unable to update cluster resource of tenant %s -> %s", ws.Name, err)
			retrigErr = err
			tnOpinternalErrors.WithLabelValues("workspace", "cluster-resources").Inc()
		}
		klog.Infof("Cluster resourcess for tenant %s updated", ws.Name)
	} else {
		klog.Errorf("Unable to update namespace of tenant %s -> %s", ws.Name, err)
		ws.Status.Namespace.Created = false
		ws.Status.Namespace.Name = ""
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("workspace", "cluster-resources").Inc()
	}

	if ws.Status.Subscriptions == nil {
		// len 1 of the map is for the number of subscriptions (keycloak)
		ws.Status.Subscriptions = make(map[string]crownlabsv1alpha1.SubscriptionStatus, 1)
	}
	// handling keycloak resources
	if err = r.KcA.createKcRoles(ctx, genWsKcRolesData(ws.Name, ws.Spec.PrettyName)); err != nil {
		klog.Errorf("Error when creating roles for workspace %s -> %s", ws.Name, err)
		ws.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrFailed
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("workspace", "keycloak").Inc()
	} else {
		klog.Infof("Roles for workspace %s created successfully", ws.Name)
		ws.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrOk
	}

	ws.Status.Ready = retrigErr == nil

	// update status before exiting reconcile
	if err = r.Status().Update(ctx, &ws); err != nil {
		// if status update fails, still try to reconcile later
		klog.Errorf("Unable to update status of workspace %s before exiting reconciler -> %s", ws.Name, err)
		tnOpinternalErrors.WithLabelValues("workspace", "self-update").Inc()
		retrigErr = err
	}

	if retrigErr != nil {
		klog.Errorf("Workspace %s failed to reconcile -> %s", ws.Name, retrigErr)
		return ctrl.Result{}, retrigErr
	}

	// no retrigErr, need to normal reconcile later, so need to create random number and exit
	nextRequeSeconds, err := randomRange(3600, 7200) // need to use seconds value for interval 1h-2h to have resultion to the second
	if err != nil {
		klog.Errorf("Error when generating random number for reque -> %s", err)
		tnOpinternalErrors.WithLabelValues("workspace", "self-update").Inc()
		return ctrl.Result{}, err
	}
	nextRequeDuration := time.Second * time.Duration(*nextRequeSeconds)
	klog.Infof("Workspace %s reconciled successfully, next in %s", ws.Name, nextRequeDuration)
	return ctrl.Result{RequeueAfter: nextRequeDuration}, nil
}

// SetupWithManager registers a new controller for Workspace resources.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(labelSelectorPredicate(r.TargetLabelKey, r.TargetLabelValue)).
		For(&crownlabsv1alpha1.Workspace{}).
		Owns(&v1.Namespace{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}

func (r *WorkspaceReconciler) handleDeletion(ctx context.Context, wsName, wsPrettyName string) error {
	var retErr error
	rolesToDelete := genWsKcRolesData(wsName, wsPrettyName)
	if err := r.KcA.deleteKcRoles(ctx, rolesToDelete); err != nil {
		klog.Errorf("Error when deleting roles of workspace %s -> %s", wsName, err)
		tnOpinternalErrors.WithLabelValues("workspace", "self-update").Inc()
		retErr = err
	}

	// unsubscribe tenants from workspace to delete
	var tenantsToUpdate crownlabsv1alpha1.TenantList
	targetLabel := fmt.Sprintf("%s%s", crownlabsv1alpha1.WorkspaceLabelPrefix, wsName)

	err := r.List(ctx, &tenantsToUpdate, &client.HasLabels{targetLabel})
	switch {
	case client.IgnoreNotFound(err) != nil:
		klog.Errorf("Error when listing tenants subscribed to workspace %s upon deletion -> %s", wsName, err)
		retErr = err
	case err != nil:
		klog.Infof("No tenants subscribed to workspace %s", wsName)
	default:
		for i := range tenantsToUpdate.Items {
			removeWsFromTn(&tenantsToUpdate.Items[i].Spec.Workspaces, wsName)
			if err := r.Update(ctx, &tenantsToUpdate.Items[i]); err != nil {
				klog.Errorf("Error when unsubscribing tenant %s from workspace %s -> %s", tenantsToUpdate.Items[i].Name, wsName, err)
				tnOpinternalErrors.WithLabelValues("workspace", "tenant-unsubscription").Inc()
				retErr = err
			}
		}
	}
	return retErr
}

func removeWsFromTn(workspaces *[]crownlabsv1alpha1.TenantWorkspaceEntry, wsToRemove string) {
	idxToRemove := -1
	for i, wsData := range *workspaces {
		if wsData.WorkspaceRef.Name == wsToRemove {
			idxToRemove = i
		}
	}
	if idxToRemove != -1 {
		*workspaces = append((*workspaces)[:idxToRemove], (*workspaces)[idxToRemove+1:]...) // Truncate slice.
	}
}

func (r *WorkspaceReconciler) createOrUpdateClusterResources(ctx context.Context, ws *crownlabsv1alpha1.Workspace, nsName string) (nsOk bool, err error) {
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}

	if _, nsErr := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		r.updateWsNamespace(&ns)
		return ctrl.SetControllerReference(ws, &ns, r.Scheme)
	}); nsErr != nil {
		klog.Errorf("Error when updating namespace of workspace %s -> %s", ws.Name, nsErr)
		return false, nsErr
	}

	var retErr error = nil
	// handle clusterRoleBinding
	crb := rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("crownlabs-manage-instances-%s", ws.Name)}}
	crbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &crb, func() error {
		r.updateWsCrb(&crb, ws.Name)
		return ctrl.SetControllerReference(ws, &crb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update cluster role binding for workspace %s -> %s", ws.Name, err)
		retErr = err
	}
	klog.Infof("Cluster role binding for workspace %s %s", ws.Name, crbOpRes)

	// handle roleBinding
	rb := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-view-templates", Namespace: nsName}}
	rbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &rb, func() error {
		r.updateWsRb(&rb, ws.Name)
		return ctrl.SetControllerReference(ws, &rb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update role binding for workspace %s -> %s", ws.Name, err)
		retErr = err
	}
	klog.Infof("Role binding for workspace %s %s", ws.Name, rbOpRes)

	// handle manager roleBinding
	managerRb := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-manage-templates", Namespace: nsName}}
	mngRbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &managerRb, func() error {
		r.updateWsRbMng(&managerRb, ws.Name)
		return ctrl.SetControllerReference(ws, &managerRb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update manager role binding for workspace %s -> %s", ws.Name, err)
		retErr = err
	}
	klog.Infof("Manager role binding for workspace %s %s", ws.Name, mngRbOpRes)

	return true, retErr
}

func (r *WorkspaceReconciler) updateWsNamespace(ns *v1.Namespace) {
	ns.Labels = r.updateWsResourceCommonLabels(ns.Labels)

	ns.Labels["crownlabs.polito.it/type"] = "workspace"
}

func (r *WorkspaceReconciler) updateWsCrb(crb *rbacv1.ClusterRoleBinding, wsName string) {
	crb.Labels = r.updateWsResourceCommonLabels(crb.Labels)

	crb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-instances", APIGroup: "rbac.authorization.k8s.io"}
	crb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWsKcRoleName(wsName, crownlabsv1alpha1.Manager)), APIGroup: "rbac.authorization.k8s.io"}}
}

func (r *WorkspaceReconciler) updateWsRb(rb *rbacv1.RoleBinding, wsName string) {
	rb.Labels = r.updateWsResourceCommonLabels(rb.Labels)

	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-view-templates", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWsKcRoleName(wsName, crownlabsv1alpha1.User)), APIGroup: "rbac.authorization.k8s.io"}}
}

func (r *WorkspaceReconciler) updateWsRbMng(rb *rbacv1.RoleBinding, wsName string) {
	rb.Labels = r.updateWsResourceCommonLabels(rb.Labels)

	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-templates", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWsKcRoleName(wsName, crownlabsv1alpha1.Manager)), APIGroup: "rbac.authorization.k8s.io"}}
}

func genWsKcRolesData(wsName, wsPrettyName string) map[string]string {
	return map[string]string{genWsKcRoleName(wsName, crownlabsv1alpha1.Manager): wsPrettyName, genWsKcRoleName(wsName, crownlabsv1alpha1.User): wsPrettyName}
}

func genWsKcRoleName(wsName string, role crownlabsv1alpha1.WorkspaceUserRole) string {
	return fmt.Sprintf("workspace-%s:%s", wsName, role)
}

func (r *WorkspaceReconciler) updateWsResourceCommonLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 1)
	}
	labels[r.TargetLabelKey] = r.TargetLabelValue
	labels["crownlabs.polito.it/managed-by"] = "workspace"

	// don't know why the initialization of the map doesn't work, so need to return a new one
	return labels
}
