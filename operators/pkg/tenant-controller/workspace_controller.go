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
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// WorkspaceReconciler reconciles a Workspace object.
type WorkspaceReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	KcA              *KcActor
	TargetLabelKey   string
	TargetLabelValue string

	RequeueTimeMinimum time.Duration
	RequeueTimeMaximum time.Duration

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

	var retrigErr error
	if !ws.ObjectMeta.DeletionTimestamp.IsZero() {
		klog.Infof("Processing deletion of workspace %s", ws.Name)
		// workspace is being deleted
		if ctrlUtil.ContainsFinalizer(&ws, crownlabsv1alpha2.TnOperatorFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.handleDeletion(ctx, ws.Name, ws.Spec.PrettyName); err != nil {
				klog.Errorf("Error when deleting resources handled by workspace  %s -> %s", ws.Name, err)
				retrigErr = err
			}
			// can remove the finalizer from the workspace if the eternal resources have been successfully deleted
			if retrigErr == nil {
				// remove finalizer from the workspace
				ctrlUtil.RemoveFinalizer(&ws, crownlabsv1alpha2.TnOperatorFinalizerName)
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
	if !ctrlUtil.ContainsFinalizer(&ws, crownlabsv1alpha2.TnOperatorFinalizerName) {
		ctrlUtil.AddFinalizer(&ws, crownlabsv1alpha2.TnOperatorFinalizerName)
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
		klog.Infof("Cluster resources for tenant %s updated", ws.Name)
	} else {
		klog.Errorf("Unable to update namespace of tenant %s -> %s", ws.Name, err)
		ws.Status.Namespace.Created = false
		ws.Status.Namespace.Name = ""
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("workspace", "cluster-resources").Inc()
	}

	// handling autoEnrollment
	err = r.handleAutoEnrollment(ctx, &ws)
	if err != nil {
		klog.Errorf("Error when handling autoEnrollment for workspace %s -> %s", ws.Name, err)
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("workspace", "auto-enrollment").Inc()
	}

	if ws.Status.Subscriptions == nil {
		// len 1 of the map is for the number of subscriptions (keycloak)
		ws.Status.Subscriptions = make(map[string]crownlabsv1alpha2.SubscriptionStatus, 1)
	}
	// handling keycloak resources
	if r.KcA == nil {
		// KcA could be nil for local testing skipping the keycloak subscription
		klog.Warningf("Skipping creation/update of roles in keycloak for workspace %s", ws.Name)
		ws.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrFailed
	} else if err = r.KcA.createKcRoles(ctx, genWsKcRolesData(ws.Name, ws.Spec.PrettyName)); err != nil {
		klog.Errorf("Error when creating roles for workspace %s -> %s", ws.Name, err)
		ws.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrFailed
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("workspace", "keycloak").Inc()
	} else {
		klog.Infof("Roles for workspace %s created successfully", ws.Name)
		ws.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrOk
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
	nextRequeueDuration := randomDuration(r.RequeueTimeMinimum, r.RequeueTimeMaximum)
	klog.Infof("Workspace %s reconciled successfully, next in %s", ws.Name, nextRequeueDuration)
	return ctrl.Result{RequeueAfter: nextRequeueDuration}, nil
}

// SetupWithManager registers a new controller for Workspace resources.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(labelSelectorPredicate(r.TargetLabelKey, r.TargetLabelValue)).
		For(&crownlabsv1alpha1.Workspace{}).
		Owns(&v1.Namespace{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.RoleBinding{}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "Workspace")).
		Complete(r)
}

func (r *WorkspaceReconciler) handleDeletion(ctx context.Context, wsName, wsPrettyName string) error {
	var retErr error
	rolesToDelete := genWsKcRolesData(wsName, wsPrettyName)
	if r.KcA != nil {
		if err := r.KcA.deleteKcRoles(ctx, rolesToDelete); err != nil {
			klog.Errorf("Error when deleting roles of workspace %s -> %s", wsName, err)
			tnOpinternalErrors.WithLabelValues("workspace", "self-update").Inc()
			retErr = err
		}
	}

	// unsubscribe tenants from workspace to delete
	var tenantsToUpdate crownlabsv1alpha2.TenantList
	targetLabel := fmt.Sprintf("%s%s", crownlabsv1alpha2.WorkspaceLabelPrefix, wsName)

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

func removeWsFromTn(workspaces *[]crownlabsv1alpha2.TenantWorkspaceEntry, wsToRemove string) {
	idxToRemove := -1
	for i, wsData := range *workspaces {
		if wsData.Name == wsToRemove {
			idxToRemove = i
		}
	}
	if idxToRemove != -1 {
		*workspaces = append((*workspaces)[:idxToRemove], (*workspaces)[idxToRemove+1:]...) // Truncate slice.
	}
}

func (r *WorkspaceReconciler) handleAutoEnrollment(ctx context.Context, ws *crownlabsv1alpha1.Workspace) error {
	// check label and update if needed
	var wantedLabel string
	if utils.AutoEnrollEnabled(ws.Spec.AutoEnroll) {
		wantedLabel = string(ws.Spec.AutoEnroll)
	} else {
		wantedLabel = "disabled"
	}

	if ws.Labels[crownlabsv1alpha2.WorkspaceLabelAutoenroll] != wantedLabel {
		ws.Labels[crownlabsv1alpha2.WorkspaceLabelAutoenroll] = wantedLabel

		if err := r.Update(ctx, ws); err != nil {
			klog.Errorf("Error when updating workspace %s -> %s", ws.Name, err)
			return err
		}
	}

	// if actual AutoEnroll is WithApproval, nothing left to do
	if ws.Spec.AutoEnroll == crownlabsv1alpha1.AutoenrollWithApproval {
		return nil
	}

	// if actual AutoEnroll is not WithApproval, manage Tenants in candidate status
	var tenantsToUpdate crownlabsv1alpha2.TenantList
	targetLabel := fmt.Sprintf("%s%s", crownlabsv1alpha2.WorkspaceLabelPrefix, ws.Name)
	err := r.List(ctx, &tenantsToUpdate, &client.MatchingLabels{targetLabel: string(crownlabsv1alpha2.Candidate)})
	if err != nil {
		klog.Errorf("Error when listing tenants subscribed to workspace %s -> %s", ws.Name, err)
		return err
	}
	for i := range tenantsToUpdate.Items {
		patch := client.MergeFrom(tenantsToUpdate.Items[i].DeepCopy())
		removeWsFromTn(&tenantsToUpdate.Items[i].Spec.Workspaces, ws.Name)
		// if AutoEnrollment is Immediate, add the workspace with User role
		if ws.Spec.AutoEnroll == crownlabsv1alpha1.AutoenrollImmediate {
			tenantsToUpdate.Items[i].Spec.Workspaces = append(
				tenantsToUpdate.Items[i].Spec.Workspaces,
				crownlabsv1alpha2.TenantWorkspaceEntry{
					Name: ws.Name,
					Role: crownlabsv1alpha2.User,
				},
			)
		}
		if err := r.Patch(ctx, &tenantsToUpdate.Items[i], patch); err != nil {
			klog.Errorf("Error when updating tenant %s -> %s", tenantsToUpdate.Items[i].Name, err)
			return err
		}
	}

	return nil
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

	var retErr error
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

	// handle manager roleBinding for SharedVolumes
	managerRbSV := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-manage-sharedvolumes", Namespace: nsName}}
	mngRbSVOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &managerRbSV, func() error {
		r.updateWsRbSVMng(&managerRbSV, ws.Name)
		return ctrl.SetControllerReference(ws, &managerRbSV, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update manager role binding for SharedVolumes for workspace %s -> %s", ws.Name, err)
		retErr = err
	}
	klog.Infof("Manager role binding for SharedVolume for workspace %s %s", ws.Name, mngRbSVOpRes)

	// handle manager clusterRoleBinding
	tenantEditCrb := rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crownlabs-manage-tenants-" + ws.Name}}
	tenEdCrbOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &tenantEditCrb, func() error {
		r.updateWsRbMngTnts(&tenantEditCrb, ws.Name)
		return ctrl.SetControllerReference(ws, &tenantEditCrb, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update tenant manager role binding for workspace %s -> %s", ws.Name, err)
		retErr = err
	}
	klog.Infof("Tenant manager role binding for workspace %s %s", ws.Name, tenEdCrbOpRes)

	return true, retErr
}

func (r *WorkspaceReconciler) updateWsNamespace(ns *v1.Namespace) {
	ns.Labels = r.updateWsResourceCommonLabels(ns.Labels)

	ns.Labels["crownlabs.polito.it/type"] = "workspace"
}

func (r *WorkspaceReconciler) updateWsCrb(crb *rbacv1.ClusterRoleBinding, wsName string) {
	crb.Labels = r.updateWsResourceCommonLabels(crb.Labels)

	crb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-instances", APIGroup: "rbac.authorization.k8s.io"}
	crb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWsKcRoleName(wsName, crownlabsv1alpha2.Manager)), APIGroup: "rbac.authorization.k8s.io"}}
}

func (r *WorkspaceReconciler) updateWsRb(rb *rbacv1.RoleBinding, wsName string) {
	rb.Labels = r.updateWsResourceCommonLabels(rb.Labels)

	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-view-templates", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWsKcRoleName(wsName, crownlabsv1alpha2.User)), APIGroup: "rbac.authorization.k8s.io"}}
}

func (r *WorkspaceReconciler) updateWsRbMng(rb *rbacv1.RoleBinding, wsName string) {
	rb.Labels = r.updateWsResourceCommonLabels(rb.Labels)

	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-templates", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWsKcRoleName(wsName, crownlabsv1alpha2.Manager)), APIGroup: "rbac.authorization.k8s.io"}}
}

func (r *WorkspaceReconciler) updateWsRbSVMng(rb *rbacv1.RoleBinding, wsName string) {
	rb.Labels = r.updateWsResourceCommonLabels(rb.Labels)

	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-sharedvolumes", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWsKcRoleName(wsName, crownlabsv1alpha2.Manager)), APIGroup: "rbac.authorization.k8s.io"}}
}

func (r *WorkspaceReconciler) updateWsRbMngTnts(rb *rbacv1.ClusterRoleBinding, wsName string) {
	rb.Labels = r.updateWsResourceCommonLabels(rb.Labels)

	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-tenants", APIGroup: "rbac.authorization.k8s.io"}
	rb.Subjects = []rbacv1.Subject{{Kind: "Group", Name: fmt.Sprintf("kubernetes:%s", genWsKcRoleName(wsName, crownlabsv1alpha2.Manager)), APIGroup: "rbac.authorization.k8s.io"}}
}

func genWsKcRolesData(wsName, wsPrettyName string) map[string]string {
	return map[string]string{genWsKcRoleName(wsName, crownlabsv1alpha2.Manager): wsPrettyName, genWsKcRoleName(wsName, crownlabsv1alpha2.User): wsPrettyName}
}

func genWsKcRoleName(wsName string, role crownlabsv1alpha2.WorkspaceUserRole) string {
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
