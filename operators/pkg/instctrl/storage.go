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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func (r *InstanceReconciler) EnforceShVolMirrorPVCs(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	for _, mountInfo := range environment.SharedVolumeMounts {
		shvol := v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mountInfo.SharedVolumeRef.Name,
				Namespace: mountInfo.SharedVolumeRef.Namespace,
			},
		}

		mirrPvc := v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      forge.GetShVolPVCMirrorName(shvol.Name, instance.Name),
				Namespace: instance.Namespace,
			},
		}
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &mirrPvc, func() error {
			// Configure the mirror PVC
			if mirrPvc.CreationTimestamp.IsZero() {
				forge.ConfigureMirrorPVC(&mirrPvc, &shvol, r.MirrorPVCStorageClassName,
					forge.InstanceObjectLabels(mirrPvc.Labels, instance))
			}
			//TODO: Stesso di là

			return controllerutil.SetControllerReference(instance, &mirrPvc, r.Scheme)
		})
		if err != nil {
			log.Error(err, "failed to enforce mirror sharedvolume", "SharedVolume", shvol, "Environment", environment)
			return err
		}
	}

	return nil
}
