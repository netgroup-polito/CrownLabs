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

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// createMyDrivePVC creates the PVC for tenant's personal storage in cross-namespace.
func (r *Reconciler) createMyDrivePVC(ctx context.Context, log logr.Logger, tn *v1alpha2.Tenant) error {
	// If the personal namespace does not exist, skip the PVC creation
	if !tn.Status.PersonalNamespace.Created {
		log.Info("Tenant namespace does not exist, skipping PVC creation")
		return nil
	}

	// Persistent volume claim NFS
	pvc, err := r.createOrUpdatePVC(ctx, tn)
	if err != nil {
		log.Error(err, "Unable to create or update PVC for tenant", "tenant", tn.Name)
		return err
	}
	log.Info("PVC created/updated")

	switch pvc.Status.Phase {
	case v1.ClaimBound:
		// authorize the user to access the PVC
		if created, err := r.createOrUpdatePVCSecret(ctx, log, tn, pvc); err != nil {
			log.Error(err, "Unable to create or update PVC Secret for tenant", "tenant", tn.Name)
			return err
		} else if created {
			log.Info("PVC Secret created/updated")
		} else {
			log.Info("Tenant namespace does not exist, skipping PVC secret creation")
		}

		val, found := pvc.Labels[forge.ProvisionJobLabel]
		if !found || val != forge.ProvisionJobValueOk {
			err = r.launchPVCProvisionJob(ctx, log, tn, pvc, forge.ProvisionJobValuePending)
			if err != nil {
				log.Error(err, "Unable to manage PVC Provisioning Job for tenant", "tenant", tn.Name)
				return err
			}
		}
	case v1.ClaimPending:
		log.Info("PVC pending for tenant", "tenant", tn.Name)
	default:
		log.Info("PVC for tenant is in unexpected phase", "tenant", tn.Name, "phase", pvc.Status.Phase)
	}

	return nil
}

// deleteMyDrivePVC deletes the PVC for tenant's personal storage.
func (r *Reconciler) deleteMyDrivePVC(ctx context.Context, log logr.Logger, tn *v1alpha2.Tenant) error {
	pvc := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.GetMyDrivePVCName(tn.Name),
			Namespace: r.MyDrivePVCsNamespace,
		},
	}

	if err := utils.EnforceObjectAbsence(ctx, r.Client, &pvc, "MyDrive PVC"); err != nil {
		log.Error(err, "Error deleting MyDrive PVC for tenant", "tenant", tn.Name)
		return err
	}

	log.Info("🔥 MyDrive PVC deleted for tenant", "tenant", tn.Name)
	return nil
}

func (r *Reconciler) createOrUpdatePVC(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) (*v1.PersistentVolumeClaim, error) {
	pvc := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.GetMyDrivePVCName(tn.Name),
			Namespace: r.MyDrivePVCsNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		// Configure the PVC using the forge package
		forge.ConfigureMyDrivePVC(&pvc, r.MyDrivePVCsStorageClassName, r.MyDrivePVCsSize,
			forge.UpdateTenantResourceCommonLabels(pvc.Labels, r.TargetLabel))

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
		log.Error(err, "Unable to get PV for tenant", "tenant", tn.Name)
		return false, err
	}

	// Get NFS server and path information from the PV
	serverName, exportPath := forge.NFSShVolSpec(&pv)

	pvcSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.NFSSecretName,
			Namespace: tn.Status.PersonalNamespace.Name,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &pvcSecret, func() error {
		// Configure the secret using the forge package
		forge.ConfigureMyDriveSecret(&pvcSecret, serverName, exportPath,
			forge.UpdateTenantResourceCommonLabels(pvcSecret.Labels, r.TargetLabel))

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
	pvc *v1.PersistentVolumeClaim,
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
			log.Info("PVC Provisioning Job created for tenant", "tenant", tn.Name)
			// Configure the provisioning job using the forge package
			forge.ConfigureMyDriveProvisioningJob(&chownJob, pvc)
		} else if provisionJobLabel == forge.ProvisionJobValuePending {
			if chownJob.Status.Succeeded == 1 {
				labelToSet = forge.ProvisionJobValueOk
				log.Info("PVC Provisioning Job completed for tenant", "tenant", tn.Name)
			} else if chownJob.Status.Failed == 1 {
				log.Info("PVC Provisioning Job failed for tenant", "tenant", tn.Name)
			}
		}

		return controllerutil.SetControllerReference(tn, &chownJob, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("unable to create or update PVC Provisioning Job: %w", err)
	}
	log.Info("PVC Provisioning Job launched")

	// Update the PVC label using the forge package
	forge.UpdatePVCProvisioningJobLabel(pvc, labelToSet)
	if err := r.Update(ctx, pvc); err != nil {
		return fmt.Errorf("unable to update PVC Provisioning Job label: %w", err)
	}

	return nil
}
