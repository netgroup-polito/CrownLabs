package tenant

import (
    "context"
    "fmt"
    "strings"

    batchv1 "k8s.io/api/batch/v1"
    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/klog/v2"
    "k8s.io/utils/ptr"

    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

    crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
    "github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
    "github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// Constants for MyDrive 
// const (
//     NFSSecretName           = "mydrive-info"
//     NFSSecretServerNameKey  = "server-name"
//     NFSSecretPathKey        = "path"
//     ProvisionJobBaseImage   = "busybox"
//     ProvisionJobMaxRetries  = 3
//     ProvisionJobTTLSeconds  = 3600 * 24 * 7  // 1 settimana
// )

// createMyDrivePVC creates the PVC for tenant's personal storage in cross-namespace
func (r *Reconciler) createMyDrivePVC(ctx context.Context, tn *crownlabsv1alpha2.Tenant) error {
    nsName := getNamespaceName(tn)
    return r.createOrUpdateTnPersonalNFSVolume(ctx, tn, nsName)
}

// deleteMyDrivePVC deletes the PVC for tenant's personal storage
func (r *Reconciler) deleteMyDrivePVC(ctx context.Context, tn *crownlabsv1alpha2.Tenant) error {
    pvc := v1.PersistentVolumeClaim{
        ObjectMeta: metav1.ObjectMeta{
            Name:      myDrivePVCName(tn.Name),
            Namespace: r.MyDrivePVCsNamespace,
        },
    }

    if err := utils.EnforceObjectAbsence(ctx, r.Client, &pvc, "MyDrive PVC"); err != nil {
        klog.Errorf("Error deleting MyDrive PVC for tenant %s: %v", tn.Name, err)
        return err
    }

    klog.Infof("🔥 MyDrive PVC deleted for tenant %s", tn.Name)
    return nil
}

// createOrUpdateTnPersonalNFSVolume 
func (r *Reconciler) createOrUpdateTnPersonalNFSVolume(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) error {
    // Persistent volume claim NFS
    pvc := v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: myDrivePVCName(tn.Name), Namespace: r.MyDrivePVCsNamespace}}

    pvcOpRes, err := controllerutil.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
        r.updateTnPersistentVolumeClaim(&pvc)
        return controllerutil.SetControllerReference(tn, &pvc, r.Scheme)
    })
    if err != nil {
        klog.Errorf("Unable to create or update PVC for tenant %s -> %s", tn.Name, err)
        return err
    }
    klog.Infof("PVC for tenant %s %s", tn.Name, pvcOpRes)

    if pvc.Status.Phase == v1.ClaimBound {
			  //add this one in order to manage secret if the tenant namespace doesn't exist
 				namespace := v1.Namespace{}
        if err := r.Get(ctx, types.NamespacedName{Name: nsName}, &namespace); err != nil {
            klog.Warningf("Namespace %s for tenant %s does not exist yet, skipping secret creation", nsName, tn.Name)
            return nil 
        }

        pv := v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: pvc.Spec.VolumeName}}
        if err := r.Get(ctx, types.NamespacedName{Name: pv.Name}, &pv); err != nil {
            klog.Errorf("Unable to get PV for tenant %s -> %s", tn.Name, err)
            return err
        }
        pvcSecret := v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: NFSSecretName, Namespace: nsName}}
        pvcSecOpRes, err := controllerutil.CreateOrUpdate(ctx, r.Client, &pvcSecret, func() error {
            r.updateTnPVCSecret(&pvcSecret, fmt.Sprintf("%s.%s", pv.Spec.CSI.VolumeAttributes["server"], pv.Spec.CSI.VolumeAttributes["clusterID"]), pv.Spec.CSI.VolumeAttributes["share"])
            return controllerutil.SetControllerReference(tn, &pvcSecret, r.Scheme)
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

            chownJobOpRes, err := controllerutil.CreateOrUpdate(ctx, r.Client, &chownJob, func() error {
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

                return controllerutil.SetControllerReference(tn, &chownJob, r.Scheme)
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

// Helper functions 
func myDrivePVCName(tnName string) string {
    return fmt.Sprintf("%s-drive", strings.ReplaceAll(tnName, ".", "-"))
}

func (r *Reconciler) updateTnPersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) {
    scName := r.MyDrivePVCsStorageClassName
    pvc.Labels = r.updateTnResourceCommonLabels(pvc.Labels)

    pvc.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteMany}
    pvc.Spec.StorageClassName = &scName

    oldSize := *pvc.Spec.Resources.Requests.Storage()
    if sizeDiff := r.MyDrivePVCsSize.Cmp(oldSize); sizeDiff > 0 || oldSize.IsZero() {
        pvc.Spec.Resources.Requests = v1.ResourceList{v1.ResourceStorage: r.MyDrivePVCsSize}
    }
}

func (r *Reconciler) updateTnPVCSecret(sec *v1.Secret, dnsName, path string) {
    sec.Labels = r.updateTnResourceCommonLabels(sec.Labels)

    sec.Type = v1.SecretTypeOpaque
    sec.Data = make(map[string][]byte, 2)
    sec.Data[NFSSecretServerNameKey] = []byte(dnsName)
    sec.Data[NFSSecretPathKey] = []byte(path)
}

func (r *Reconciler) updateTnProvisioningJob(chownJob *batchv1.Job, pvc *v1.PersistentVolumeClaim) {
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
            }},
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
        }}
        chownJob.Spec.Template.Spec.Volumes = []v1.Volume{{
            Name: "mydrive",
            VolumeSource: v1.VolumeSource{
                PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
                    ClaimName: pvc.Name,
                },
            },
        }}
    }
}