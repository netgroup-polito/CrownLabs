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

package tenant

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// Constants for MyDrive.
const (
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

// createMyDrivePVC creates the PVC for tenant's personal storage in cross-namespace.
func (r *Reconciler) createMyDrivePVC(ctx context.Context, log logr.Logger, tn *v1alpha2.Tenant) error {
	// Persistent volume claim NFS
	pvc, err := r.createOrUpdatePVC(ctx, tn)
	if err != nil {
		log.Error(err, "Unable to create or update PVC for tenant %s -> %s", tn.Name, err)
		return err
	}
	log.Info("PVC created/updated")

	switch pvc.Status.Phase {
	case v1.ClaimBound:
		// authorize the user to access the PVC
		if created, err := r.createOrUpdatePVCSecret(ctx, log, tn, pvc); err != nil {
			log.Error(err, "Unable to create or update PVC Secret for tenant %s -> %s", tn.Name, err)
			return err
		} else if created {
			log.Info("PVC Secret created/updated")
		} else {
			log.Info("Tenant namespace does not exist, skipping PVC secret creation")
		}

		val, found := pvc.Labels[forge.ProvisionJobLabel]
		if !found || val != forge.ProvisionJobValueOk {
			err = r.launchPVCProvisionJob(ctx, log, tn, *pvc, forge.ProvisionJobValuePending)
			if err != nil {
				log.Error(err, "Unable to manage PVC Provisioning Job for tenant %s -> %s", tn.Name, err)
				return err
			}
		}
	case v1.ClaimPending:
		log.Info("PVC pending for tenant %s", tn.Name)
	default:
		log.Info("PVC for tenant %s is in unexpected phase: %s", tn.Name, pvc.Status.Phase)
	}

	return nil
}

// deleteMyDrivePVC deletes the PVC for tenant's personal storage.
func (r *Reconciler) deleteMyDrivePVC(ctx context.Context, log logr.Logger, tn *v1alpha2.Tenant) error {
	pvc := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myDrivePVCName(tn.Name),
			Namespace: r.MyDrivePVCsNamespace,
		},
	}

	if err := utils.EnforceObjectAbsence(ctx, r.Client, &pvc, "MyDrive PVC"); err != nil {
		log.Error(err, "Error deleting MyDrive PVC for tenant %s: %v", tn.Name, err)
		return err
	}

	log.Info("🔥 MyDrive PVC deleted for tenant %s", tn.Name)
	return nil
}

func (r *Reconciler) createOrUpdatePVC(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) (*v1.PersistentVolumeClaim, error) {
	pvc := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myDrivePVCName(tn.Name),
			Namespace: r.MyDrivePVCsNamespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			StorageClassName: &r.MyDrivePVCsStorageClassName,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		pvc.Labels = r.updateTnResourceCommonLabels(pvc.Labels)

		oldSize := *pvc.Spec.Resources.Requests.Storage()
		if sizeDiff := r.MyDrivePVCsSize.Cmp(oldSize); sizeDiff > 0 || oldSize.IsZero() {
			pvc.Spec.Resources.Requests = v1.ResourceList{v1.ResourceStorage: r.MyDrivePVCsSize}
		}
		return controllerutil.SetControllerReference(tn, &pvc, r.Scheme)
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create or update PVC for tenant %s: %w", tn.Name, err)
	}

	return &pvc, nil
}

func (r *Reconciler) createOrUpdatePVCSecret(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
	pvc *v1.PersistentVolumeClaim,
) (bool, error) {
	// if the personal namespace does not exist, skip the secret creation
	if !tn.Status.PersonalNamespace.Created {
		return false, nil
	}

	pv := v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvc.Spec.VolumeName,
		},
	}
	if err := r.Get(ctx, types.NamespacedName{Name: pv.Name}, &pv); err != nil {
		log.Error(err, "Unable to get PV for tenant %s -> %s", tn.Name, err)
		return false, err
	}

	pvcSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NFSSecretName,
			Namespace: tn.Status.PersonalNamespace.Name,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &pvcSecret, func() error {
		pvcSecret.Labels = r.updateTnResourceCommonLabels(pvcSecret.Labels)

		pvcSecret.Type = v1.SecretTypeOpaque
		pvcSecret.Data = make(map[string][]byte, 2)
		pvcSecret.Data[NFSSecretServerNameKey] = []byte(fmt.Sprintf(
			"%s.%s",
			pv.Spec.CSI.VolumeAttributes["server"],
			pv.Spec.CSI.VolumeAttributes["clusterID"],
		))
		pvcSecret.Data[NFSSecretPathKey] = []byte(pv.Spec.CSI.VolumeAttributes["share"])
		return controllerutil.SetControllerReference(tn, &pvcSecret, r.Scheme)
	})

	if err != nil {
		return false, fmt.Errorf("unable to create or update PVC Secret for tenant %s: %w", tn.Name, err)
	}

	return true, nil
}

func (r *Reconciler) launchPVCProvisionJob(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
	pvc v1.PersistentVolumeClaim,
	provisionJobLabel string,
) error {
	chownJob := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc.Name + "-provision",
			Namespace: pvc.Namespace,
		},
	}
	labelToSet := forge.ProvisionJobValuePending

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &chownJob, func() error {
		if chownJob.CreationTimestamp.IsZero() {
			log.Info("PVC Provisioning Job created for tenant %s", tn.Name)
			r.updateTnProvisioningJob(&chownJob, &pvc)
		} else if provisionJobLabel == forge.ProvisionJobValuePending {
			if chownJob.Status.Succeeded == 1 {
				labelToSet = forge.ProvisionJobValueOk
				log.Info("PVC Provisioning Job completed for tenant %s", tn.Name)
			} else if chownJob.Status.Failed == 1 {
				log.Info("PVC Provisioning Job failed for tenant %s", tn.Name)
			}
		}

		return controllerutil.SetControllerReference(tn, &chownJob, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("unable to create or update PVC Provisioning Job: %w", err)
	}
	log.Info("PVC Provisioning Job launched")

	pvc.Labels[forge.ProvisionJobLabel] = labelToSet
	if err := r.Update(ctx, &pvc); err != nil {
		return fmt.Errorf("unable to update PVC Provisioning Job label: %w", err)
	}

	return nil
}

// Helper functions.
func myDrivePVCName(tnName string) string {
	return fmt.Sprintf("%s-drive", strings.ReplaceAll(tnName, ".", "-"))
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
