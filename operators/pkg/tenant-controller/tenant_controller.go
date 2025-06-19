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

// Package tenant_controller groups the functionalities related to the Tenant controller.
package tenant_controller

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

const (
	// NoWorkspacesLabel -> label to be set (to true) when no workspaces are associated to the tenant.
	NoWorkspacesLabel = "crownlabs.polito.it/no-workspaces"
	// NFSSecretName -> NFS secret name.
	NFSSecretName = "mydrive-info"
	// NFSSecretServerNameKey -> NFS Server key in NFS secret.
	NFSSecretServerNameKey = "server-name"
	// NFSSecretPathKey -> NFS path key in NFS secret.
	NFSSecretPathKey = "path"
	// ProvisionJobBaseImage -> Base container image for Personal Drive provision job.
	ProvisionJobBaseImage = "busybox"
	// ProvisionJobMaxRetries -> Maximum number of retries for Provision jobs.
	ProvisionJobMaxRetries = 3
	// ProvisionJobTTLSeconds -> Seconds for Provision jobs before deletion (either failure or success).
	ProvisionJobTTLSeconds = 3600 * 24 * 7
)

// TenantReconciler reconciles a Tenant object.
type TenantReconciler struct {
	client.Client
	Scheme                      *runtime.Scheme
	KcA                         *KcActor
	TargetLabelKey              string
	TargetLabelValue            string
	SandboxClusterRole          string
	Concurrency                 int
	MyDrivePVCsSize             resource.Quantity
	MyDrivePVCsStorageClassName string
	MyDrivePVCsNamespace        string
	RequeueTimeMinimum          time.Duration
	RequeueTimeMaximum          time.Duration
	TenantNSKeepAlive           time.Duration
	BaseWorkspaces              []string

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

	if tn.Status.Subscriptions == nil {
		// make initial len is 1 (keycloak)
		tn.Status.Subscriptions = make(map[string]crownlabsv1alpha2.SubscriptionStatus, 1)
	}

	tenantExistingWorkspaces, workspaces, enrolledWorkspaces, err := r.checkValidWorkspaces(ctx, &tn)

	if err != nil {
		retrigErr = err
	}

	// Determine if the personal namespace should be deleted

	nsName := fmt.Sprintf("tenant-%s", strings.ReplaceAll(tn.Name, ".", "-"))

	// Test if namespace has been open for too long; check if it is ok to delete
	keepNsOpen, err := r.checkNamespaceKeepAlive(ctx, &tn, nsName)
	if err != nil {
		klog.Errorf("Error checking whether tenant namespace %s should be kept alive: %s", nsName, err)
		tnOpinternalErrors.WithLabelValues("tenant", "self-update").Inc()
		return ctrl.Result{}, err
	}

	// update resource quota in the status of the tenant after checking validity of workspaces.
	tn.Status.Quota = forge.TenantResourceList(workspaces, tn.Spec.Quota)

	_, err = r.enforceClusterResources(ctx, &tn, nsName, keepNsOpen)
	if err != nil {
		klog.Errorf("Error when enforcing cluster resources for tenant %s -> %s", tn.Name, err)
		tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
		return ctrl.Result{}, err
	}

	if err = r.handleKeycloakSubscription(ctx, &tn, enrolledWorkspaces); err != nil {
		klog.Errorf("Error when updating keycloak subscription for tenant %s -> %s", tn.Name, err)
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrFailed
		retrigErr = err
		tnOpinternalErrors.WithLabelValues("tenant", "keycloak").Inc()
	} else {
		klog.Infof("Keycloak subscription for tenant %s updated", tn.Name)
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrOk
	}

	if err = r.EnforceSandboxResources(ctx, &tn); err != nil {
		klog.Errorf("Failed checking sandbox for tenant %s -> %s", tn.Name, err)
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

	if err = r.updateTnLabels(&tn, tenantExistingWorkspaces); err != nil {
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
	nextRequeueDuration := randomDuration(r.RequeueTimeMinimum, r.RequeueTimeMaximum)
	klog.Infof("Tenant %s reconciled successfully, next in %s", tn.Name, nextRequeueDuration)
	return ctrl.Result{RequeueAfter: nextRequeueDuration}, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha2.Tenant{}, builder.WithPredicates(labelSelectorPredicate(r.TargetLabelKey, r.TargetLabelValue))).
		Owns(&v1.Secret{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&v1.Namespace{}).
		Owns(&v1.ResourceQuota{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&netv1.NetworkPolicy{}).
		Owns(&batchv1.Job{}).
		Watches(&crownlabsv1alpha1.Workspace{},
			handler.EnqueueRequestsFromMapFunc(r.workspaceToEnrolledTenants)).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "Tenant")).
		Complete(r)
}

// handleDeletion deletes external resources of a tenant using a fail-fast:false strategy.
func (r *TenantReconciler) handleDeletion(ctx context.Context, tnName string) error {
	var retErr error
	// delete keycloak user
	if r.KcA != nil {
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
	}
	return retErr
}

// checkValidWorkspaces check validity of workspaces in tenant.
// allWsEntry []TenantWorkspaceEntry and allWs []Workspace contains all the workspaces associated with the tenant.
// enrolledWs []TenantWorkspaceEntry contains only the workspaces the tenant is enrolled in (`user` or `manager`).
func (r *TenantReconciler) checkValidWorkspaces(ctx context.Context, tn *crownlabsv1alpha2.Tenant) (allWsEntry []crownlabsv1alpha2.TenantWorkspaceEntry, allWs []crownlabsv1alpha1.Workspace, enrolledWs []crownlabsv1alpha2.TenantWorkspaceEntry, retErr error) {
	tenantExistingWorkspaces := []crownlabsv1alpha2.TenantWorkspaceEntry{}
	enrolledWorkspaces := []crownlabsv1alpha2.TenantWorkspaceEntry{}
	workspaces := []crownlabsv1alpha1.Workspace{}
	tn.Status.FailingWorkspaces = []string{}
	var err error
	// check every workspace of a tenant
	for _, tnWs := range tn.Spec.Workspaces {
		wsLookupKey := types.NamespacedName{Name: tnWs.Name}
		var ws crownlabsv1alpha1.Workspace
		err = r.Get(ctx, wsLookupKey, &ws)
		switch {
		case err != nil:
			// if there was a problem, add the workspace to the status of the tenant
			klog.Errorf("Error when checking if workspace %s exists in tenant %s -> %s", tnWs.Name, tn.Name, err)
			tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, tnWs.Name)
			tnOpinternalErrors.WithLabelValues("tenant", "workspace-not-exist").Inc()
		case tnWs.Role == crownlabsv1alpha2.Candidate && ws.Spec.AutoEnroll != crownlabsv1alpha1.AutoenrollWithApproval:
			// Candidate role is allowed only if the workspace has autoEnroll = WithApproval
			klog.Errorf("Workspace %s has not autoEnroll with approval, Candidate role is not allowed in tenant %s", tnWs.Name, tn.Name)
			tn.Status.FailingWorkspaces = append(tn.Status.FailingWorkspaces, tnWs.Name)
		default:
			tenantExistingWorkspaces = append(tenantExistingWorkspaces, tnWs)
			workspaces = append(workspaces, ws)
			if tnWs.Role != crownlabsv1alpha2.Candidate {
				enrolledWorkspaces = append(enrolledWorkspaces, tnWs)
			}
		}
	}
	return tenantExistingWorkspaces, workspaces, enrolledWorkspaces, err
}

// deleteClusterNamespace deletes the namespace for the tenant, if it fails then it returns an error.
func (r *TenantReconciler) deleteClusterNamespace(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) error {
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}

	err := utils.EnforceObjectAbsence(ctx, r.Client, &ns, "personal namespace")

	if err != nil {
		klog.Errorf("Error when deleting namespace of tenant %s -> %s", tn.Name, err)
	}

	return err
}

// checkNamespaceKeepAlive checks to see if the namespace should be deleted.
func (r *TenantReconciler) checkNamespaceKeepAlive(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) (keepNsOpen bool, err error) {
	// We check to see if last login was more than r.TenantNSKeepAlive in the past:
	// if so, temporarily delete the namespace. We assume that a lastLogin of 0 occurs when a user is first created

	// Calculate time elapsed since lastLogin (now minus lastLogin in seconds)
	sPassed := time.Since(tn.Spec.LastLogin.Time)

	klog.Infof("Last login of tenant %s was %s ago", tn.Name, sPassed)

	// Attempt to get instances in current namespace
	list := &crownlabsv1alpha2.InstanceList{}

	if err := r.List(ctx, list, client.InNamespace(nsName)); err != nil {
		return true, err
	}

	if sPassed > r.TenantNSKeepAlive { // seconds
		klog.Infof("Over %s elapsed since last login of tenant %s: tenant namespace shall be absent", r.TenantNSKeepAlive, tn.Name)
		if len(list.Items) == 0 {
			klog.Infof("No instances found in %s: namespace can be deleted", nsName)
			return false, nil
		}
		klog.Infof("Instances found in namespace %s. Namespace will not be deleted", nsName)
	} else {
		klog.Infof("Under %s (limit) elapsed since last login of tenant %s: tenant namespace shall be present", r.TenantNSKeepAlive, tn.Name)
	}

	return true, nil
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
		err := r.deleteClusterNamespace(ctx, tn, nsName)
		if err == nil {
			klog.Infof("Namespace %s for tenant %s enforced to be absent", nsName, tn.Name)
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

	err = r.createOrUpdateTnPersonalNFSVolume(ctx, tn, nsName)
	if err != nil {
		klog.Errorf("Unable to create or update personal NFS volume for tenant %s -> %s", tn.Name, err)
		retErr = err
	}
	klog.Infof("Personal NFS volume for tenant %s", tn.Name)

	return true, retErr
}

// Creates or updates the user's personal NFS volume.
func (r *TenantReconciler) createOrUpdateTnPersonalNFSVolume(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) error {
	// Persistent volume claim NFS
	pvc := v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: myDrivePVCName(tn.Name), Namespace: r.MyDrivePVCsNamespace}}

	pvcOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		r.updateTnPersistentVolumeClaim(&pvc)
		return ctrl.SetControllerReference(tn, &pvc, r.Scheme)
	})
	if err != nil {
		klog.Errorf("Unable to create or update PVC for tenant %s -> %s", tn.Name, err)
		return err
	}
	klog.Infof("PVC for tenant %s %s", tn.Name, pvcOpRes)

	if pvc.Status.Phase == v1.ClaimBound {
		pv := v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: pvc.Spec.VolumeName}}
		if err := r.Get(ctx, types.NamespacedName{Name: pv.Name}, &pv); err != nil {
			klog.Errorf("Unable to get PV for tenant %s -> %s", tn.Name, err)
			return err
		}
		pvcSecret := v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: NFSSecretName, Namespace: nsName}}
		pvcSecOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvcSecret, func() error {
			r.updateTnPVCSecret(&pvcSecret, fmt.Sprintf("%s.%s", pv.Spec.CSI.VolumeAttributes["server"], pv.Spec.CSI.VolumeAttributes["clusterID"]), pv.Spec.CSI.VolumeAttributes["share"])
			return ctrl.SetControllerReference(tn, &pvcSecret, r.Scheme)
		})
		if err != nil {
			klog.Errorf("Unable to create or update PVC Secret for tenant %s -> %s", tn.Name, err)
			return err
		}
		klog.Infof("PVC Secret for tenant %s %s", tn.Name, pvcSecOpRes)

		val, found := pvc.Labels[forge.ProvisionJobLabel]
		if !found || val != forge.ProvisionJobValueOk {
			chownJob := batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: pvc.Name + "-provision", Namespace: pvc.Namespace}}
			labelToSet := forge.ProvisionJobValuePending

			chownJobOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &chownJob, func() error {
				if chownJob.CreationTimestamp.IsZero() {
					klog.Infof("PVC Provisioning Job created for tenant %s", tn.Name)
					r.updateTnProvisioningJob(&chownJob, &pvc)
				} else if found && val == forge.ProvisionJobValuePending {
					if chownJob.Status.Succeeded == 1 {
						labelToSet = forge.ProvisionJobValueOk
						klog.Infof("PVC Provisioning Job completed for tenant %s", tn.Name)
					} else if chownJob.Status.Failed == 1 {
						klog.Warningf("PVC Provisioning Job failed for tenant %s", tn.Name)
					}
				}

				return ctrl.SetControllerReference(tn, &chownJob, r.Scheme)
			})
			if err != nil {
				klog.Errorf("Unable to create or update PVC Provisioning Job for tenant %s -> %s", tn.Name, err)
				return err
			}
			klog.Infof("PVC Provisioning Job for tenant %s %s", tn.Name, chownJobOpRes)

			pvc.Labels[forge.ProvisionJobLabel] = labelToSet
			if err := r.Update(ctx, &pvc); err != nil {
				klog.Errorf("PVC Provisioning Job failed to update PVC labels for tenant %s", tn.Name)
			}
			klog.Infof("PVC Provisioning Job updateded PVC label to %s for tenant %s", labelToSet, tn.Name)
		}
	} else if pvc.Status.Phase == v1.ClaimPending {
		klog.Infof("PVC pending for tenant %s", tn.Name)
	}
	return nil
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

func (r *TenantReconciler) updateTnPersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) {
	scName := r.MyDrivePVCsStorageClassName
	pvc.Labels = r.updateTnResourceCommonLabels(pvc.Labels)

	pvc.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteMany}
	pvc.Spec.StorageClassName = &scName

	oldSize := *pvc.Spec.Resources.Requests.Storage()
	if sizeDiff := r.MyDrivePVCsSize.Cmp(oldSize); sizeDiff > 0 || oldSize.IsZero() {
		pvc.Spec.Resources.Requests = v1.ResourceList{v1.ResourceStorage: r.MyDrivePVCsSize}
	}
}

func (r *TenantReconciler) updateTnProvisioningJob(chownJob *batchv1.Job, pvc *v1.PersistentVolumeClaim) {
	if chownJob.CreationTimestamp.IsZero() {
		chownJob.Spec.BackoffLimit = ptr.To[int32](ProvisionJobMaxRetries)
		chownJob.Spec.TTLSecondsAfterFinished = ptr.To[int32](ProvisionJobTTLSeconds)
		chownJob.Spec.Template.Spec.RestartPolicy = v1.RestartPolicyOnFailure
		chownJob.Spec.Template.Spec.Containers = []v1.Container{{
			Name:    "chown-container",
			Image:   ProvisionJobBaseImage,
			Command: []string{"chown", "-R", fmt.Sprintf("%d:%d", forge.CrownLabsUserID, forge.CrownLabsUserID), forge.MyDriveVolumeMountPath},
			VolumeMounts: []v1.VolumeMount{{
				Name:      "mydrive",
				MountPath: forge.MyDriveVolumeMountPath,
			},
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"cpu":    resource.MustParse("100m"),
					"memory": resource.MustParse("128Mi"),
				},
				Limits: v1.ResourceList{
					"cpu":    resource.MustParse("100m"),
					"memory": resource.MustParse("128Mi"),
				},
			},
		},
		}
		chownJob.Spec.Template.Spec.Volumes = []v1.Volume{{
			Name: "mydrive",
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
				},
			},
		},
		}
	}
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

func (r *TenantReconciler) updateTnPVCSecret(sec *v1.Secret, dnsName, path string) {
	sec.Labels = r.updateTnResourceCommonLabels(sec.Labels)

	sec.Type = v1.SecretTypeOpaque
	sec.Data = make(map[string][]byte, 2)
	sec.Data[NFSSecretServerNameKey] = []byte(dnsName)
	sec.Data[NFSSecretPathKey] = []byte(path)
}

func (r *TenantReconciler) updateTnLabels(tn *crownlabsv1alpha2.Tenant, tenantExistingWorkspaces []crownlabsv1alpha2.TenantWorkspaceEntry) error {
	if tn.Labels == nil {
		tn.Labels = map[string]string{}
	} else {
		cleanWorkspaceLabels(tn.Labels)
	}

	nonBaseWorkspacesCount := 0
	for _, wsData := range tenantExistingWorkspaces {
		wsLabelKey := fmt.Sprintf("%s%s", crownlabsv1alpha2.WorkspaceLabelPrefix, wsData.Name)
		tn.Labels[wsLabelKey] = string(wsData.Role)
		if !containsString(r.BaseWorkspaces, wsData.Name) {
			nonBaseWorkspacesCount++
		}
	}
	// label for users without workspaces
	if nonBaseWorkspacesCount == 0 {
		tn.Labels[NoWorkspacesLabel] = "true"
	} else {
		delete(tn.Labels, NoWorkspacesLabel)
	}

	tn.Labels["crownlabs.polito.it/first-name"] = cleanName(tn.Spec.FirstName)
	tn.Labels["crownlabs.polito.it/last-name"] = cleanName(tn.Spec.LastName)
	return nil
}

func myDrivePVCName(tnName string) string {
	return fmt.Sprintf("%s-drive", strings.ReplaceAll(tnName, ".", "-"))
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

func (r *TenantReconciler) workspaceToEnrolledTenants(ctx context.Context, o client.Object) []reconcile.Request {
	var enqueues []reconcile.Request
	var tenants crownlabsv1alpha2.TenantList
	if err := r.List(ctx, &tenants, client.HasLabels{
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
