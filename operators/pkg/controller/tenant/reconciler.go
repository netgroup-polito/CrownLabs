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
package tenant

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"

	"time"
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

type TenantReconciler struct {
	client.Client
	Scheme                      *runtime.Scheme
	TargetLabel                 common.KVLabel
	TenantNSKeepAlive           time.Duration
	TriggerReconcileChannel     chan event.GenericEvent // Channel to trigger a reconciliation of the tenant resource.
	MyDrivePVCsSize             resource.Quantity
	MyDrivePVCsStorageClassName string
	MyDrivePVCsNamespace        string
	KeycloakActor               common.KeycloakActorIface
}

// Reconcile reconciles the state of a tenant resource.
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "tenant", req.NamespacedName.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	log.Info("Reconciling tenant", "name", req.NamespacedName.Name)

	var tn crownlabsv1alpha2.Tenant
	if err := r.Get(ctx, req.NamespacedName, &tn); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting tenant %s before starting reconcile -> %s", req.Name, err)
		return ctrl.Result{}, err
	} else if err != nil {
		log.Info("Tenant deleted", "name", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !r.TargetLabel.IsIncluded(tn.Labels) {
		// the actual Tenant is not responsibility of this controller
		log.Info("Tenant is not responsibility of this controller, skipping reconcile")
		return ctrl.Result{}, nil
	}

	defer func() {
		// update the Tenant status
		if err := r.Status().Update(ctx, &tn); err != nil {
			klog.Errorf("Error updating status for tenant %s: %v", tn.Name, err)
		}
	}()

	// check if the Tenant is being deleted
	if !tn.DeletionTimestamp.IsZero() {
		err := r.deleteTenant(ctx, &tn)
		if err != nil {
			klog.Errorf("Error deleting tenant %s: %v", tn.Name, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// add the finalizer if it is not already present
	if !controllerutil.ContainsFinalizer(&tn, crownlabsv1alpha2.TnOperatorFinalizerName) {
		controllerutil.AddFinalizer(&tn, crownlabsv1alpha2.TnOperatorFinalizerName)
		if err := r.Update(ctx, &tn); err != nil {
			klog.Errorf("Error adding finalizer to tenant %s: %v", tn.Name, err)
			return ctrl.Result{}, err
		}
		klog.Infof("Finalizer %s added to tenant %s", crownlabsv1alpha2.TnOperatorFinalizerName, tn.Name)
	}

	if tn.Status.Subscriptions == nil {
		tn.Status.Subscriptions = make(map[string]crownlabsv1alpha2.SubscriptionStatus)
	}

	// manage generic labels
	if err := r.updateTenantBaseLabels(ctx, &log, &tn); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating tenant base labels: %w", err)
	}

	// manage workspaces subscription (and related labels)
	if err := r.manageWorkspaces(ctx, &tn); err != nil {
		klog.Errorf("Error managing workspaces for tenant %s: %v", tn.Name, err)
		return ctrl.Result{}, fmt.Errorf("error managing workspaces for tenant %s: %w", tn.Name, err)
	}

	// check if the tenant is already been provisioned in Keycloak
	// - if not, create the tenant in Keycloak
	// - if yes, check if the tenant is verified
	verified, err := r.CheckKeycloakUserVerified(ctx, &tn)
	if err != nil {
		klog.Errorf("Error checking Keycloak status for tenant %s: %v", tn.Name, err)
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrFailed
		tn.Status.Ready = false
		return ctrl.Result{}, err
	} else {
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrOk
	}

	// manage keycloak tenant authorization for workspaces
	if err := r.updateWorkspacesAuthorizationRoles(ctx, &log, &tn); err != nil {
		klog.Errorf("Error updating tenant authorization roles for tenant %s: %v", tn.Name, err)
		tn.Status.Subscriptions["keycloak"] = crownlabsv1alpha2.SubscrFailed
		tn.Status.Ready = false
		return ctrl.Result{}, fmt.Errorf("error updating tenant authorization roles for tenant %s: %w", tn.Name, err)
	}

	if !verified {
		// if the Tenant has not been verified, we can skip the reconciliation
		// and wait for the next reconcile loop
		log.Info("Tenant not verified, skipping resource creation")
		return ctrl.Result{}, nil
	}

	// managing resources not related to the personal namespace
	//   if the Tenant has already been verified, we can proceed with the reconciliation
	//   and create related resources

	// determine the Tenant resource quota based on the Spec and the existing workspaces
	if err := r.forgeServiceQuota(ctx, &tn); err != nil {
		klog.Errorf("Error forging service quota for tenant %s: %v", tn.Name, err)
		tnOpinternalErrors.WithLabelValues("tenant", "quota-forge").Inc()
		return ctrl.Result{}, fmt.Errorf("error forging service quota for tenant %s: %w", tn.Name, err)
	}

	// managing resources related to the personal namespace

	// Test if namespace has been open for too long; check if it is ok to delete
	keepAlive, err := r.checkNamespaceKeepAlive(ctx, &tn)
	if err != nil {
		klog.Errorf("Error checking whether tenant namespace should be kept alive: %s", err)
		tnOpinternalErrors.WithLabelValues("tenant", "self-update").Inc()
		return ctrl.Result{}, err
	}

	if keepAlive {
		// Namespace should be kept open, so we proceed with the reconciliation
		// creating or updating the cluster resources

		// TODO: create or update personal namespace
		// TODO: manage resource quota
		// TODO: tutte le cose che partono da enforceClusterResources
	} else {
		// Namespace should not be kept open, so we delete all the resources related to the tenant

		// TODO: cancella le cose che arrivano da enforceClusterResources e il personal namespace
	}

	// TODO: manage sandbox

	tn.Status.Ready = true

	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {

	pred, err := r.TargetLabel.GetPredicate()
	if err != nil {
		klog.Errorf("Error creating predicate for tenant controller: %v", err)
		return fmt.Errorf("error creating predicate for tenant controller: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha2.Tenant{}, builder.WithPredicates(pred)).
		Owns(&v1.Secret{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&v1.Namespace{}).
		Owns(&v1.ResourceQuota{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&netv1.NetworkPolicy{}).
		Owns(&batchv1.Job{}).
		WatchesRawSource(
			&source.Channel{
				Source: r.TriggerReconcileChannel,
			},
			handler.Funcs{
				GenericFunc: func(_ context.Context, e event.GenericEvent, q workqueue.RateLimitingInterface) {
					q.Add(ctrl.Request{
						NamespacedName: client.ObjectKey{
							Name: e.Object.GetName(),
						},
					})
				},
			},
		).
		// TODO
		// Watches(&crownlabsv1alpha1.Workspace{},
		// 	handler.EnqueueRequestsFromMapFunc(r.workspaceToEnrolledTenants)).
		// WithOptions(controller.Options{
		// 	MaxConcurrentReconciles: r.Concurrency,
		// }).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "Tenant")).
		Complete(r)
}

func (r *TenantReconciler) deleteTenant(
	ctx context.Context,
	tn *crownlabsv1alpha2.Tenant,
) error {
	// TODO delete all resources related to the tenant

	// remove the tenant from Keycloak
	err := r.deleteTenantInKeycloak(ctx, tn)
	if err != nil {
		klog.Errorf("Error deleting tenant %s in Keycloak: %v", tn.Name, err)
		return err
	}

	// delete the finalizer
	if controllerutil.ContainsFinalizer(tn, crownlabsv1alpha2.TnOperatorFinalizerName) {
		controllerutil.RemoveFinalizer(tn, crownlabsv1alpha2.TnOperatorFinalizerName)
		if err := r.Update(ctx, tn); err != nil {
			klog.Errorf("Error removing finalizer from tenant %s: %v", tn.Name, err)
			return err
		}
		klog.Infof("Finalizer %s removed from tenant %s", crownlabsv1alpha2.TnOperatorFinalizerName, tn.Name)
	}
	return nil
}

// func (r *TenantReconciler) createOrUpdateTnPersonalNFSVolume(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) error {
// 	// Persistent volume claim NFS
// 	pvc := v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: myDrivePVCName(tn.Name), Namespace: r.MyDrivePVCsNamespace}}

// 	pvcOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
// 		r.updateTnPersistentVolumeClaim(&pvc)
// 		return ctrl.SetControllerReference(tn, &pvc, r.Scheme)
// 	})
// 	if err != nil {
// 		klog.Errorf("Unable to create or update PVC for tenant %s -> %s", tn.Name, err)
// 		return err
// 	}
// 	klog.Infof("PVC for tenant %s %s", tn.Name, pvcOpRes)

// 	if pvc.Status.Phase == v1.ClaimBound {
// 		pv := v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: pvc.Spec.VolumeName}}
// 		if err := r.Get(ctx, types.NamespacedName{Name: pv.Name}, &pv); err != nil {
// 			klog.Errorf("Unable to get PV for tenant %s -> %s", tn.Name, err)
// 			return err
// 		}
// 		pvcSecret := v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: NFSSecretName, Namespace: nsName}}
// 		pvcSecOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvcSecret, func() error {
// 			r.updateTnPVCSecret(&pvcSecret, fmt.Sprintf("%s.%s", pv.Spec.CSI.VolumeAttributes["server"], pv.Spec.CSI.VolumeAttributes["clusterID"]), pv.Spec.CSI.VolumeAttributes["share"])
// 			return ctrl.SetControllerReference(tn, &pvcSecret, r.Scheme)
// 		})
// 		if err != nil {
// 			klog.Errorf("Unable to create or update PVC Secret for tenant %s -> %s", tn.Name, err)
// 			return err
// 		}
// 		klog.Infof("PVC Secret for tenant %s %s", tn.Name, pvcSecOpRes)

// 		val, found := pvc.Labels[forge.ProvisionJobLabel]
// 		if !found || val != forge.ProvisionJobValueOk {
// 			chownJob := batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: pvc.Name + "-provision", Namespace: pvc.Namespace}}
// 			labelToSet := forge.ProvisionJobValuePending

// 			chownJobOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &chownJob, func() error {
// 				if chownJob.CreationTimestamp.IsZero() {
// 					klog.Infof("PVC Provisioning Job created for tenant %s", tn.Name)
// 					r.updateTnProvisioningJob(&chownJob, &pvc)
// 				} else if found && val == forge.ProvisionJobValuePending {
// 					if chownJob.Status.Succeeded == 1 {
// 						labelToSet = forge.ProvisionJobValueOk
// 						klog.Infof("PVC Provisioning Job completed for tenant %s", tn.Name)
// 					} else if chownJob.Status.Failed == 1 {
// 						klog.Warningf("PVC Provisioning Job failed for tenant %s", tn.Name)
// 					}
// 				}

// 				return ctrl.SetControllerReference(tn, &chownJob, r.Scheme)
// 			})
// 			if err != nil {
// 				klog.Errorf("Unable to create or update PVC Provisioning Job for tenant %s -> %s", tn.Name, err)
// 				return err
// 			}
// 			klog.Infof("PVC Provisioning Job for tenant %s %s", tn.Name, chownJobOpRes)

// 			pvc.Labels[forge.ProvisionJobLabel] = labelToSet
// 			if err := r.Update(ctx, &pvc); err != nil {
// 				klog.Errorf("PVC Provisioning Job failed to update PVC labels for tenant %s", tn.Name)
// 			}
// 			klog.Infof("PVC Provisioning Job updateded PVC label to %s for tenant %s", labelToSet, tn.Name)
// 		}
// 	} else if pvc.Status.Phase == v1.ClaimPending {
// 		klog.Infof("PVC pending for tenant %s", tn.Name)
// 	}
// 	return nil
// }

// func (r *TenantReconciler) updateTnRb(rb *rbacv1.RoleBinding, tnName string) {
// 	rb.Labels = r.updateTnResourceCommonLabels(rb.Labels)
// 	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-manage-instances", APIGroup: "rbac.authorization.k8s.io"}
// 	rb.Subjects = []rbacv1.Subject{{Kind: "User", Name: tnName, APIGroup: "rbac.authorization.k8s.io"}}
// }

// func (r *TenantReconciler) updateTnCr(rb *rbacv1.ClusterRole, tnName string) {
// 	rb.Labels = r.updateTnResourceCommonLabels(rb.Labels)
// 	rb.Rules = []rbacv1.PolicyRule{{
// 		APIGroups:     []string{"crownlabs.polito.it"},
// 		Resources:     []string{"tenants"},
// 		ResourceNames: []string{tnName},
// 		Verbs:         []string{"get", "list", "watch", "patch", "update"},
// 	}}
// }

// func (r *TenantReconciler) updateTnCrb(rb *rbacv1.ClusterRoleBinding, tnName, crName string) {
// 	rb.Labels = r.updateTnResourceCommonLabels(rb.Labels)
// 	rb.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: crName, APIGroup: "rbac.authorization.k8s.io"}
// 	rb.Subjects = []rbacv1.Subject{{Kind: "User", Name: tnName, APIGroup: "rbac.authorization.k8s.io"}}
// }

// func (r *TenantReconciler) updateTnProvisioningJob(chownJob *batchv1.Job, pvc *v1.PersistentVolumeClaim) {
// 	if chownJob.CreationTimestamp.IsZero() {
// 		chownJob.Spec.BackoffLimit = ptr.To[int32](ProvisionJobMaxRetries)
// 		chownJob.Spec.TTLSecondsAfterFinished = ptr.To[int32](ProvisionJobTTLSeconds)
// 		chownJob.Spec.Template.Spec.RestartPolicy = v1.RestartPolicyOnFailure
// 		chownJob.Spec.Template.Spec.Containers = []v1.Container{{
// 			Name:    "chown-container",
// 			Image:   ProvisionJobBaseImage,
// 			Command: []string{"chown", "-R", fmt.Sprintf("%d:%d", forge.CrownLabsUserID, forge.CrownLabsUserID), forge.MyDriveVolumeMountPath},
// 			VolumeMounts: []v1.VolumeMount{{
// 				Name:      "mydrive",
// 				MountPath: forge.MyDriveVolumeMountPath,
// 			},
// 			},
// 			Resources: v1.ResourceRequirements{
// 				Requests: v1.ResourceList{
// 					"cpu":    resource.MustParse("100m"),
// 					"memory": resource.MustParse("128Mi"),
// 				},
// 				Limits: v1.ResourceList{
// 					"cpu":    resource.MustParse("100m"),
// 					"memory": resource.MustParse("128Mi"),
// 				},
// 			},
// 		},
// 		}
// 		chownJob.Spec.Template.Spec.Volumes = []v1.Volume{{
// 			Name: "mydrive",
// 			VolumeSource: v1.VolumeSource{
// 				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
// 					ClaimName: pvc.Name,
// 				},
// 			},
// 		},
// 		}
// 	}
// }

// func (r *TenantReconciler) updateTnPVCSecret(sec *v1.Secret, dnsName, path string) {
// 	sec.Labels = r.updateTnResourceCommonLabels(sec.Labels)

// 	sec.Type = v1.SecretTypeOpaque
// 	sec.Data = make(map[string][]byte, 2)
// 	sec.Data[NFSSecretServerNameKey] = []byte(dnsName)
// 	sec.Data[NFSSecretPathKey] = []byte(path)
// }

// func (r *TenantReconciler) updateTnNetPolDeny(np *netv1.NetworkPolicy) {
// 	np.Labels = r.updateTnResourceCommonLabels(np.Labels)
// 	np.Spec.PodSelector.MatchLabels = make(map[string]string)
// 	np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}}}}
// }

// func (r *TenantReconciler) updateTnNetPolAllow(np *netv1.NetworkPolicy) {
// 	np.Labels = r.updateTnResourceCommonLabels(np.Labels)
// 	np.Spec.PodSelector.MatchLabels = make(map[string]string)
// 	np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{{NamespaceSelector: &metav1.LabelSelector{
// 		MatchLabels: map[string]string{"crownlabs.polito.it/allow-instance-access": "true"},
// 	}}}}}
// }

// func (r *TenantReconciler) updateTnPersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) {
// 	scName := r.MyDrivePVCsStorageClassName
// 	pvc.Labels = r.updateTnResourceCommonLabels(pvc.Labels)

// 	pvc.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteMany}
// 	pvc.Spec.StorageClassName = &scName

// 	oldSize := *pvc.Spec.Resources.Requests.Storage()
// 	if sizeDiff := r.MyDrivePVCsSize.Cmp(oldSize); sizeDiff > 0 || oldSize.IsZero() {
// 		pvc.Spec.Resources.Requests = v1.ResourceList{v1.ResourceStorage: r.MyDrivePVCsSize}
// 	}
// }

// func myDrivePVCName(tnName string) string {
// 	return fmt.Sprintf("%s-drive", strings.ReplaceAll(tnName, ".", "-"))
// }
