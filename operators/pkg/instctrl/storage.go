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

package instctrl

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// EnforceShVolMirrorPVCs implements the logic behind the creation of the mirror PVCs
// for a SharedVolume mounted on an Environment.
func (r *InstanceReconciler) EnforceShVolMirrorPVCs(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	for _, mountInfo := range environment.SharedVolumeMounts {
		shvol := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mountInfo.SharedVolumeRef.Name,
				Namespace: mountInfo.SharedVolumeRef.Namespace,
			},
		}

		mirrPvc := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      forge.GetShVolPVCMirrorName(shvol.Name, instance.Name),
				Namespace: instance.Namespace,
			},
		}
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &mirrPvc, func() error {
			// Configure the mirror PVC
			if mirrPvc.CreationTimestamp.IsZero() {
				mirrPvc.Spec = forge.MirrorPVCSpec(&shvol, r.MirrorPVCStorageClassName)
			}
			mirrPvc.SetLabels(forge.UpdateShVolMirrorPVCLabels(mirrPvc.Labels))

			return controllerutil.SetControllerReference(instance, &mirrPvc, r.Scheme)
		})
		if err != nil {
			log.Error(err, "failed to enforce mirror sharedvolume", "SharedVolume", shvol, "Environment", environment)
			return err
		}
	}

	return nil
}

// CheckMyDriveMirrorPVC checks if the mirror PVC of the MyDrive has been created.
// It returns an error if it hasn't been created yet, or nil if it exists.
func (r *InstanceReconciler) CheckMyDriveMirrorPVC(ctx context.Context) error {
	tenant := clctx.TenantFrom(ctx)
	key := types.NamespacedName{
		Namespace: tenant.Status.PersonalNamespace.Name,
		Name:      forge.GetMyDrivePVCMirrorName(tenant.Name),
	}

	var mirror corev1.PersistentVolumeClaim
	return r.Get(ctx, key, &mirror)
}
