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

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v13/controller"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

var (
	defaultMirrorCapacity = resource.MustParse("1")
	errPendingPVC         = errors.New("cannot mirror a pending PVC")
	errStopProvision      = errors.New("stop provisioning")
)

const (
	// MirroredPvLabel is the key of the label identifying which PV is mirroring.
	MirroredPvLabel = "crownlabs.polito.it/mirrored-pv"
	// MirroredPvcLabel is the key of the label identifying which PVC is mirroring.
	MirroredPvcLabel = "crownlabs.polito.it/mirrored-pvc"

	// Key1 is the key of the annotation. //TODO.
	Key1 = "pmp.crownlabs.polito.it/required-target-ns-labels"
	// Key2 is the key of the annotation. //TODO.
	Key2 = "pmp.crownlabs.polito.it/required-target-tn-labels"
)

// PvcMirrorProvisioner provisions PVCs with MirrorStorageClass.
type PvcMirrorProvisioner struct {
	Ctx                   context.Context
	Client                client.Client
	Config                *rest.Config
	Logger                logr.Logger
	TargetLabel           common.KVLabel
	MirrorStorageClass    string
	MirrorProvisionerName string
}

// Provision is the Provisioner interface function called when a PVC has to be provisioned, and returns the PV to be created on the cluster.
func (p *PvcMirrorProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	mirrorPVC := options.PVC

	// Check TargetLabel
	if proceed, err := utils.CheckNamespaceTargetLabel(ctx, p.Client, mirrorPVC.Namespace, p.TargetLabel); !proceed {
		if err != nil {
			p.Logger.Error(err, "Failed checking target label")
			return nil, controller.ProvisioningFinished, err
		}

		p.Logger.Info("PVC is not responsibility of this controller, skipping provision")
		return nil, controller.ProvisioningFinished, errStopProvision
	}

	// Check mirror PVC's spec
	if mirrorPVC.Spec.DataSourceRef == nil {
		p.Logger.Error(errStopProvision, "Cannot retrieve origin PVC (DataSourceRef is missing)")
		return nil, controller.ProvisioningFinished, errStopProvision
	}
	if mirrorPVC.Spec.DataSourceRef.APIGroup != nil || mirrorPVC.Spec.DataSourceRef.Kind != "PersistentVolumeClaim" {
		p.Logger.Error(errStopProvision, "Cannot retrieve origin PVC (DataSourceRef is not a PVC)")
		return nil, controller.ProvisioningFinished, errStopProvision
	}

	// Get origin PVC
	originKey := types.NamespacedName{
		Name:      mirrorPVC.Spec.DataSourceRef.Name,
		Namespace: *mirrorPVC.Spec.DataSourceRef.Namespace,
	}

	var originPVC v1.PersistentVolumeClaim
	if err := p.Client.Get(ctx, originKey, &originPVC); err != nil {
		p.Logger.Error(err, "Could not retrieve origin PVC")
		return nil, controller.ProvisioningFinished, err
	}

	// Check Authorization
	requiredLabelsNs, presentNs := originPVC.Annotations[Key1]
	_, presentTn := originPVC.Annotations[Key2] // requiredLabelsTn

	if !presentNs && !presentTn {
		// Default: deny-all
		p.Logger.Error(errStopProvision, "No required labels specified on origin PVC, access denied")
		return nil, controller.ProvisioningFinished, errStopProvision
	}
	if proceed, err := utils.CheckNamespaceWithSelector(ctx, p.Client, mirrorPVC.Namespace, requiredLabelsNs); !presentNs || !proceed {
		if err != nil {
			p.Logger.Error(err, "Failed checking namespace selector")
			return nil, controller.ProvisioningFinished, err
		}

		p.Logger.Info("Namespace does not match the selector on the origin PVC, access denied")
		return nil, controller.ProvisioningFinished, errStopProvision
	}
	// if proceed, err := utils.CheckTenantWithSelector(ctx, p.Client, mirrorPVC.Namespace, requiredLabelsTn); !presentTn || !proceed {
	// 	if err != nil {
	// 		p.Logger.Error(err, "Failed checking tenant selector")
	// 		return nil, controller.ProvisioningFinished, err
	// 	}

	// 	p.Logger.Info("Tenant does not match the selector on the origin PVC, access denied")
	// 	return nil, controller.ProvisioningFinished, errStopProvision
	// }

	// Check origin PVC's phase
	switch originPVC.Status.Phase {
	case v1.ClaimPending:
		return nil, controller.ProvisioningFinished, errPendingPVC
	case v1.ClaimLost:
		return nil, controller.ProvisioningFinished, errStopProvision
	case v1.ClaimBound:
		// continue
	}

	// Check origin PV's CSI spec
	var originPV v1.PersistentVolume
	if err := p.Client.Get(ctx, types.NamespacedName{Name: originPVC.Spec.VolumeName}, &originPV); err != nil {
		// Sicuramente non è perché non c'è
		p.Logger.Error(err, "Could not retrieve origin PV")
		return nil, controller.ProvisioningFinished, err
	}
	if originPV.Spec.CSI == nil {
		p.Logger.Error(errStopProvision, "Cannot retrieve CSI spec from PV (CSI is missing)")
		return nil, controller.ProvisioningFinished, errStopProvision
	}

	// Create mirror PV
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName, // By default: "pvc-" + claim.UID
			Labels: map[string]string{
				MirroredPvLabel:  originPV.Name,
				MirroredPvcLabel: originPVC.Namespace + "---" + originPVC.Name,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *options.StorageClass.ReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: defaultMirrorCapacity,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: originPV.Spec.CSI,
			},
			MountOptions:     originPV.Spec.MountOptions,
			StorageClassName: options.StorageClass.Name,
		},
	}

	//TODO: Add annotation on the mirror PVC (cannot be done here!)

	return pv, controller.ProvisioningFinished, nil
}

//TODO: Potremmo restituire errori più bloccanti di errStopProvision come ignore o infeasible

// Delete is the Provisioner interface function called when a PVC has to be deleted.
func (p *PvcMirrorProvisioner) Delete(_ context.Context, _ *v1.PersistentVolume) error {
	return nil
}

// Start is the Runnable interface function called when the Provisioner has to be run by the Manager.
func (p *PvcMirrorProvisioner) Start(ctx context.Context) error {
	// Check if StorageClass is available
	p.Logger.Info("Checking if mirror storage class is available")
	var outputSc storagev1.StorageClass
	if err := p.Client.Get(ctx, types.NamespacedName{Name: p.MirrorStorageClass}, &outputSc); err != nil {
		if kerrors.IsNotFound(err) {
			p.Logger.Error(err, "Cannot proceed without a StorageClass")
		} else {
			p.Logger.Error(err, "Could not retrieve StorageClass")
		}
		return err
	}

	// TODO: Check outputSc.Provisioner == p.MirrorProvisionerName and outputSc.AllowVolumeExpansion == false

	// Create k8s client and run Provision controller
	clientset, err := kubernetes.NewForConfig(p.Config)
	if err != nil {
		p.Logger.Error(err, "Failed to create client")
		return err
	}
	pc := controller.NewProvisionController(ctx, clientset, p.MirrorProvisionerName, p)
	pc.Run(ctx)

	return nil
}

// ShouldDelete returns whether deleting the PV should be attempted.
func ShouldDelete(context.Context, *v1.PersistentVolume) bool {
	//TODO: If needed, add here checks to be done before deleting a PV
	return true
}
