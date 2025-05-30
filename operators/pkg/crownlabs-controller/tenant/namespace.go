package tenant
import (
	"context"
	"fmt"
	
	"time"
	"strings"


	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	ctrl "sigs.k8s.io/controller-runtime"
	
	"sigs.k8s.io/controller-runtime/pkg/client"
	

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"

)
	

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


// updateTnNamespace updates the tenant namespace.
func (r *TenantReconciler) updateTnNamespace(ns *v1.Namespace, tnName string) {
	ns.Labels = r.updateTnResourceCommonLabels(ns.Labels)
	ns.Labels["crownlabs.polito.it/type"] = "tenant"
	ns.Labels["crownlabs.polito.it/name"] = tnName
	ns.Labels["crownlabs.polito.it/instance-resources-replication"] = "true"
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

//other stuff that need to be moved later on 
func (r *TenantReconciler) updateTnResourceCommonLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 1)
	}
	labels[r.TargetLabelKey] = r.TargetLabelValue
	labels["crownlabs.polito.it/managed-by"] = "tenant"
	return labels
}

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


func (r *TenantReconciler) updateTnPVCSecret(sec *v1.Secret, dnsName, path string) {
	sec.Labels = r.updateTnResourceCommonLabels(sec.Labels)

	sec.Type = v1.SecretTypeOpaque
	sec.Data = make(map[string][]byte, 2)
	sec.Data[NFSSecretServerNameKey] = []byte(dnsName)
	sec.Data[NFSSecretPathKey] = []byte(path)
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

func myDrivePVCName(tnName string) string {
	return fmt.Sprintf("%s-drive", strings.ReplaceAll(tnName, ".", "-"))
}