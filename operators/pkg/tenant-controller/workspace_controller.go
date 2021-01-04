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

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	KcA    *KcActor
}

// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=workspaces/status,verbs=get;update;patch

// Reconcile reconciles the state of a workspace resource
func (r *WorkspaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	var ws crownlabsv1alpha1.Workspace

	if err := r.Get(ctx, req.NamespacedName, &ws); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting workspace %s before starting reconcile", ws.Name)
		klog.Error(err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Workspace %s deleted", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var retrigErr error = nil
	if !ws.ObjectMeta.DeletionTimestamp.IsZero() {
		klog.Infof("Processing deletion of workspace %s", ws.Name)
		// workspace is being deleted
		if containsString(ws.ObjectMeta.Finalizers, crownlabsv1alpha1.TnOperatorFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			rolesToDelete := genWorkspaceRolesData(ws.Name, ws.Spec.PrettyName)
			if err := r.KcA.deleteKcRoles(ctx, rolesToDelete); err != nil {
				klog.Errorf("Error when deleting roles of workspace %s", ws.Name)
				klog.Error(err)
				retrigErr = err
			}

			// unsubscribe tenants from workspace to delete
			var tenantsToUpdate crownlabsv1alpha1.TenantList
			targetLabel := fmt.Sprintf("%s%s", crownlabsv1alpha1.WorkspaceLabelPrefix, ws.Name)
			switch err := r.List(ctx, &tenantsToUpdate, &client.HasLabels{targetLabel}); {
			case client.IgnoreNotFound(err) != nil:
				klog.Errorf("Error when listing tenants subscribed to workspace %s upon deletion", ws.Name)
				klog.Error(err)
				retrigErr = err
			case err != nil:
				klog.Infof("No tenants subscribed to workspace %s", ws.Name)
			default:
				for i := range tenantsToUpdate.Items {
					deleteWorkspace(&tenantsToUpdate.Items[i].Spec.Workspaces, ws.Name)
					if err := r.Update(ctx, &tenantsToUpdate.Items[i]); err != nil {
						klog.Errorf("Error when unsubscribing tenant %s from workspace %s", tenantsToUpdate.Items[i].Name, ws.Name)
						klog.Error(err)
						retrigErr = err
					}
				}
			}

			// remove finalizer from the workspace
			ws.ObjectMeta.Finalizers = removeString(ws.ObjectMeta.Finalizers, crownlabsv1alpha1.TnOperatorFinalizerName)
			if err := r.Update(context.Background(), &ws); err != nil {
				klog.Errorf("Error when removing tenant operator finalizer from workspace %s", ws.Name)
				klog.Error(err)
				retrigErr = err
			}
		}
		if retrigErr == nil {
			klog.Infof("Workspace %s ready for deletion", ws.Name)
		} else {
			klog.Errorf("Error when preparing workspace %s for deletion, need to retry", ws.Name)
		}
		return ctrl.Result{}, retrigErr
	}

	// workspace is NOT being deleted
	klog.Infof("Reconciling workspace %s", ws.Name)

	// add tenant operator finalizer to workspace
	if !containsString(ws.ObjectMeta.Finalizers, crownlabsv1alpha1.TnOperatorFinalizerName) {
		ws.ObjectMeta.Finalizers = append(ws.ObjectMeta.Finalizers, crownlabsv1alpha1.TnOperatorFinalizerName)
		if err := r.Update(context.Background(), &ws); err != nil {
			klog.Errorf("Error when adding finalizer to workspace %s", ws.Name)
			klog.Error(err)
			retrigErr = err
		}
	}

	nsName := fmt.Sprintf("workspace-%s", ws.Name)
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}

	nsOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		updateNamespace(&ns)
		return ctrl.SetControllerReference(&ws, &ns, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update namespace of workspace %s", ws.Name)
		klog.Error(err)
		ws.Status.Namespace.Created = false
		ws.Status.Namespace.Name = ""
		retrigErr = err
	} else {
		klog.Infof("Namespace %s for workspace %s %s", nsName, ws.Name, nsOpRes)
		ws.Status.Namespace.Created = true
		ws.Status.Namespace.Name = nsName
		if err = r.createOrUpdateWsClusterResources(ctx, &ws, nsName); err != nil {
			klog.Errorf("Error creating k8s resources for workspace %s", ws.Name)
			klog.Error(err)
			retrigErr = err
		} else {
			klog.Infof("Cluster resources for workspace %s have been correctly handled", ws.Name)
		}
	}

	if ws.Status.Subscriptions == nil {
		ws.Status.Subscriptions = make(map[string]crownlabsv1alpha1.SubscriptionStatus)
	}
	if err = r.KcA.createKcRoles(ctx, genWorkspaceRolesData(ws.Name, ws.Spec.PrettyName)); err != nil {
		ws.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrFailed
		retrigErr = err
	} else {
		ws.Status.Subscriptions["keycloak"] = crownlabsv1alpha1.SubscrOk
	}

	ws.Status.Ready = retrigErr == nil

	// update status before exiting reconcile
	if err = r.Status().Update(ctx, &ws); err != nil {
		// if status update fails, still try to reconcile later
		klog.Error("Unable to update status before exiting reconciler", err)
		retrigErr = err
	}

	if retrigErr != nil {
		klog.Errorf("Workspace %s failed to reconcile", ws.Name)
		return ctrl.Result{}, retrigErr
	}

	// no retrigErr, need to normal reconcile later, so need to create random number and exit
	nextRequeSeconds, err := randomRange(3600, 7200) // need to use seconds value for interval 1h-2h to have resultion to the second
	if err != nil {
		klog.Error("Error when generating random number for reque", err)
		return ctrl.Result{}, err
	}
	nextRequeDuration := time.Second * time.Duration(*nextRequeSeconds)
	klog.Infof("Workspace %s reconciled successfully, next in %s", ws.Name, nextRequeDuration)
	return ctrl.Result{RequeueAfter: nextRequeDuration}, nil
}

func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha1.Workspace{}).
		Owns(&v1.Namespace{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}

func updateNamespace(ns *v1.Namespace) {
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels["crownlabs.polito.it/type"] = "workspace"
}

func genWorkspaceRolesData(wsName, wsPrettyName string) map[string]string {
	return map[string]string{genWorkspaceRoleName(wsName, crownlabsv1alpha1.Manager): wsPrettyName, genWorkspaceRoleName(wsName, crownlabsv1alpha1.User): wsPrettyName}
}

func genWorkspaceRoleName(wsName string, role crownlabsv1alpha1.WorkspaceUserRole) string {
	return fmt.Sprintf("workspace-%s:%s", wsName, role)
}

func deleteWorkspace(workspaces *[]crownlabsv1alpha1.UserWorkspaceData, wsToRemove string) {
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

func (r *WorkspaceReconciler) createOrUpdateWsClusterResources(ctx context.Context, ws *crownlabsv1alpha1.Workspace, nsName string) error {
	// handle clusterRoleBinding
	crb := rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("crownlabs-manage-instances-%s", ws.Name)}}
	crbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &crb, func() error {
		updateWsCrb(&crb, ws.Name)
		return ctrl.SetControllerReference(ws, &crb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update cluster role binding for workspace %s", ws.Name)
		klog.Error(err)
		return err
	}
	klog.Infof("Cluster role binding for workspace %s %s", ws.Name, crbOpRes)

	// handle roleBinding
	rb := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-view-templates", Namespace: nsName}}
	rbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &rb, func() error {
		updateWsRb(&rb, ws.Name)
		return ctrl.SetControllerReference(ws, &rb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update role binding for workspace %s", ws.Name)
		klog.Error(err)
		return err
	}
	klog.Infof("Role binding for workspace %s %s", ws.Name, rbOpRes)

	// handle manager roleBinding
	managerRb := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-manage-templates", Namespace: nsName}}
	mngRbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &managerRb, func() error {
		updateWsRbMng(&managerRb, ws.Name)
		return ctrl.SetControllerReference(ws, &managerRb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update manager role binding for workspace %s", ws.Name)
		klog.Error(err)
		return err
	}
	klog.Infof("Manager role binding for workspace %s %s", ws.Name, mngRbOpRes)

	return nil
}

func updateWsCrb(crb *rbacv1.ClusterRoleBinding, wsName string) {
	if crb.Labels == nil {
		crb.Labels = make(map[string]string, 1)
	}
	crb.Labels["crownlabs.polito.it/managed-by"] = "workspace"
	crb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-instances", APIGroup: "rbac.authorization.k8s.io"}
	crb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWorkspaceRoleName(wsName, crownlabsv1alpha1.Manager)), APIGroup: "rbac.authorization.k8s.io"}}
}

func updateWsRb(rb *rbacv1.RoleBinding, wsName string) {
	if rb.Labels == nil {
		rb.Labels = make(map[string]string, 1)
	}
	rb.Labels["crownlabs.polito.it/managed-by"] = "workspace"
	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-view-templates", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWorkspaceRoleName(wsName, crownlabsv1alpha1.User)), APIGroup: "rbac.authorization.k8s.io"}}
}

func updateWsRbMng(rb *rbacv1.RoleBinding, wsName string) {
	if rb.Labels == nil {
		rb.Labels = make(map[string]string, 1)
	}
	rb.Labels["crownlabs.polito.it/managed-by"] = "workspace"
	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-templates", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWorkspaceRoleName(wsName, crownlabsv1alpha1.Manager)), APIGroup: "rbac.authorization.k8s.io"}}
}
