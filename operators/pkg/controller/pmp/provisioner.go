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

// Package pmp groups the functionalities related to the PVC Mirror Provisioner.
package pmp

import (
	"context"
	"errors"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v13/controller"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	defaultMirrorCapacity = resource.MustParse("1")
	errPendingPVC         = errors.New("Cannot mirror a Pending PVC")
	errStopProvision      = errors.New("stop provisioning")
)

const (
	MirroredPvLabel  = "crownlabs.polito.it/mirrored-pv"
	MirroredPvcLabel = "crownlabs.polito.it/mirrored-pvc"
)

type PvcMirrorProvisioner struct {
	// identity string useful? do we really care about the node in which the provisioner is working?
	Client client.Client
}

func (p *PvcMirrorProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	//TODO: Check if authorized!!!!!!!!!!!!!!!!!! can't mirror all pvs in any namespace

	// Check origin PV's phase
	originPVC := options.PVC
	if originPVC.Status.Phase == v1.ClaimPending {
		return nil, controller.ProvisioningFinished, errPendingPVC
	} else if originPVC.Status.Phase == v1.ClaimLost {
		return nil, controller.ProvisioningFinished, errStopProvision
	}

	// Get original PV
	var originPV v1.PersistentVolume
	if err := p.Client.Get(ctx, types.NamespacedName{Name: originPVC.Spec.VolumeName}, &originPV); err != nil {
		// Sicuramente non è perché non c'è
		//TODO: logger
		return nil, controller.ProvisioningFinished, err
	}

	if originPV.Spec.CSI == nil {
		//TODO: logger
		return nil, controller.ProvisioningFinished, errStopProvision
	}

	// Create mirror PV
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName, // By default: "pvc-" + claim.UID
			Labels: map[string]string{
				MirroredPvLabel:  originPV.Name,
				MirroredPvcLabel: originPVC.Namespace + "---" + originPVC.Name,
				//TODO: do we want a label provisioned by?
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *options.StorageClass.ReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): defaultMirrorCapacity,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: originPV.Spec.CSI,
			},
			MountOptions:     originPV.Spec.MountOptions,
			StorageClassName: options.StorageClass.Name,
		},
	}

	//TODO: Add annotation on the mirror PVC

	return pv, controller.ProvisioningFinished, nil
}

func (p *PvcMirrorProvisioner) Delete(ctx context.Context, volume *v1.PersistentVolume) error {
	return nil
}

// ShouldDelete returns whether deleting the PV should be attempted.
func ShouldDelete(context.Context, *v1.PersistentVolume) bool {
	return true
}
