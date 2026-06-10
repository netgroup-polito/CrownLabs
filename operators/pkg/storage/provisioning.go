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

// Package storage collects all the logic behind the provisioning of attachable storage
package storage

import (
	"context"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// RunPVCProvisioning enforces a job to provision the passed PVC, changing its owner and adding it a label when done.
// Returns a bool as first parameter that represents whether the job has completed successfully.
func RunPVCProvisioning(ctx context.Context, log logr.Logger, c client.Client, pvc *corev1.PersistentVolumeClaim, owner metav1.Object) (bool, error) {
	log = log.WithName("provisioning-job")

	val, found := pvc.Labels[forge.ProvisionJobLabel]
	if !found || val != forge.ProvisionJobValueOk {
		chownJob := batchv1.Job{
			ObjectMeta: forge.ObjectMetaWithSuffix(pvc, "provision"),
		}
		labelToSet := forge.ProvisionJobValuePending

		chownJobOpRes, err := ctrl.CreateOrUpdate(ctx, c, &chownJob, func() error {
			if chownJob.CreationTimestamp.IsZero() {
				log.Info("Created")
				chownJob.Spec = forge.PVCProvisioningJobSpec(pvc)
			} else {
				if chownJob.Status.Succeeded == 1 {
					labelToSet = forge.ProvisionJobValueOk
					log.Info("Completed")
				} else if chownJob.Status.Failed == 1 {
					log.Info("Failed")
				}
			}

			return ctrl.SetControllerReference(owner, &chownJob, c.Scheme())
		})
		if err != nil {
			log.Error(err, "Unable to create or update Job")
			return false, err
		}
		log.Info("Job enforced", "result", chownJobOpRes)

		// Update the PVC label
		if err := utils.PatchObject(ctx, c, pvc, func(pvc *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
			pvc.SetLabels(forge.UpdatePVCProvisioningJobLabel(pvc.Labels, labelToSet))
			return pvc
		}); err != nil {
			log.Error(err, "unable to update PVC Provisioning Job label")
			return false, err
		}

		log.Info("PVC labels changed",
			"previous", val, "current", labelToSet)
	}

	if pvc.Labels[forge.ProvisionJobLabel] == forge.ProvisionJobValueOk {
		return true, nil
	}

	return false, nil
}
