// Copyright 2020-2026 Politecnico di Torino
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/storage"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// enforcePersonalStorage creates the MyDrive PVC (and connected resources) for tenant's personal storage in cross-namespace.
func (r *Reconciler) enforcePersonalStorage(ctx context.Context, log logr.Logger, tn *clv1alpha2.Tenant) error {
	// If the personal namespace does not exist, skip the PVC creation
	if !tn.Status.PersonalNamespace.Created {
		log.Info("Tenant namespace does not exist, skipping PVC creation")
		return nil
	}

	// Enforce the presence of the Personal PVC
	pvc, err := r.enforceMyDrivePVC(ctx, tn)
	if err != nil {
		log.Error(err, "Unable to create or update PVC for tenant")
		return err
	}
	log.Info("PVC enforced")

	switch pvc.Status.Phase {
	case corev1.ClaimBound:
		// Authorize the user to access the PVC by creating a Mirror inside their namespace
		if created, err := r.enforceMyDrivePVCMirror(ctx, tn, pvc); err != nil {
			log.Error(err, "Unable to create or update PVC Mirror for tenant")
			return err
		} else if created {
			log.Info("PVC Mirror enforced")
		} else {
			log.Info("Tenant namespace does not exist, skipping PVC Mirror creation")
		}

		if _, err := storage.RunPVCProvisioning(ctx, log, r.Client, pvc, tn); err != nil {
			return err
		}
	case corev1.ClaimPending:
		log.Info("PVC pending for tenant")
	default:
		log.Info("PVC for tenant is in unexpected phase", "phase", pvc.Status.Phase)
	}

	return nil
}

// enforceMyDrivePVCAbsence deletes the PVC for tenant's personal storage.
func (r *Reconciler) enforceMyDrivePVCAbsence(ctx context.Context, log logr.Logger, tn *clv1alpha2.Tenant) error {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.MyDrivePVCName(tn.Name),
			Namespace: r.MyDrivePVCsNamespace,
		},
	}

	if err := utils.EnforceObjectAbsence(ctx, r.Client, &pvc, "MyDrive PVC"); err != nil {
		log.Error(err, "Error deleting MyDrive PVC for tenant")
		return err
	}

	log.Info("MyDrive PVC deleted for tenant")
	return nil
}

func (r *Reconciler) enforceMyDrivePVC(
	ctx context.Context,
	tn *clv1alpha2.Tenant,
) (*corev1.PersistentVolumeClaim, error) {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.MyDrivePVCName(tn.Name),
			Namespace: r.MyDrivePVCsNamespace,
		},
	}

	_, err := ctrlutil.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		// Configure the PVC
		if pvc.CreationTimestamp.IsZero() {
			pvc.Spec = forge.MyDrivePVCSpec(r.MyDrivePVCsStorageClassName, r.MyDrivePVCsSize)
		}
		pvc.SetLabels(forge.UpdateTenantResourceCommonLabels(pvc.Labels, r.TargetLabel))
		pvc.SetAnnotations(forge.UpdateMyDrivePVCAnnotations(pvc.Annotations, tn.Name))

		// Update size only if it needs to be bigger
		oldSize := *pvc.Spec.Resources.Requests.Storage()
		if sizeDiff := r.MyDrivePVCsSize.Cmp(oldSize); sizeDiff > 0 || oldSize.IsZero() {
			pvc.Spec.Resources.Requests = corev1.ResourceList{corev1.ResourceStorage: r.MyDrivePVCsSize}
		}

		return ctrlutil.SetControllerReference(tn, &pvc, r.Scheme)
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create or update PVC for tenant %s: %w", tn.Name, err)
	}

	return &pvc, nil
}

func (r *Reconciler) enforceMyDrivePVCMirror(
	ctx context.Context,
	tn *clv1alpha2.Tenant,
	pvc *corev1.PersistentVolumeClaim,
) (bool, error) {
	// if the personal namespace does not exist, skip the secret creation
	if !tn.Status.PersonalNamespace.Created {
		return false, nil
	}

	mirrPvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.MyDrivePVCMirrorName(tn.Name),
			Namespace: tn.Status.PersonalNamespace.Name,
		},
	}
	_, err := ctrlutil.CreateOrUpdate(ctx, r.Client, &mirrPvc, func() error {
		// Configure the mirror PVC
		if mirrPvc.CreationTimestamp.IsZero() {
			mirrPvc.Spec = forge.MirrorPVCSpec(pvc, r.MirrorPVCStorageClassName)
		}
		mirrPvc.SetLabels(forge.UpdateMyDriveMirrorPVCLabels(mirrPvc.Labels, r.TargetLabel))

		return ctrlutil.SetControllerReference(tn, &mirrPvc, r.Scheme)
	})
	if err != nil {
		return false, fmt.Errorf("unable to create or update mirror PVC for tenant %s: %w", tn.Name, err)
	}

	return true, nil
}
