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
	"strings"

	"github.com/go-logr/logr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
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
	errInvalidConfig      = errors.New("invalid configuration for pmp")
	errPendingPVC         = errors.New("cannot mirror a pending PVC")
	errStopProvision      = errors.New("stop provisioning")
	errSlowRetry          = errors.New("provisioning error, will be slowly retried")
)

const (
	// MirroredPvLabel is the key of the label identifying which PV is mirroring.
	MirroredPvLabel = "crownlabs.polito.it/mirrored-pv"
	// MirroredPvcLabel is the key of the label identifying which PVC is mirroring.
	MirroredPvcLabel = "crownlabs.polito.it/mirrored-pvc"

	// AuthorizationAnnotationKey is the key of the annotation that shows which labels are requested on the target ns to mirror the pvc.
	AuthorizationAnnotationKey = "pmp.crownlabs.polito.it/required-target-ns-labels"
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
// It can return three types of errors: IgnoreError (provisioning will never be retried again), Infeasible (will be slowly retried with backoff),
// and other errors (will be rescheduled normally); the type of returned error depends on if the error can be recovered or not.
func (p *PvcMirrorProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*corev1.PersistentVolume, controller.ProvisioningState, error) {
	mirrorPVC := options.PVC

	// Check TargetLabel
	if proceed, err := utils.CheckNamespaceTargetLabel(ctx, p.Client, mirrorPVC.Namespace, p.TargetLabel); !proceed {
		if err != nil {
			p.Logger.Error(err, "failed checking target label")
			return nil, controller.ProvisioningFinished, err
		}

		p.Logger.Info("PVC is not responsibility of this controller, skipping provision")
		return nil, controller.ProvisioningFinished, &controller.IgnoredError{Reason: "PVC is not responsibility of this controller"}
	}

	// Check mirror PVC's spec
	if mirrorPVC.Spec.DataSourceRef == nil {
		p.Logger.Error(errStopProvision, "cannot retrieve origin PVC (DataSourceRef is missing)")
		return nil, controller.ProvisioningFinished, &controller.IgnoredError{Reason: "DataSourceRef is missing"}
	}
	if mirrorPVC.Spec.DataSourceRef.APIGroup != nil || mirrorPVC.Spec.DataSourceRef.Kind != "PersistentVolumeClaim" {
		p.Logger.Error(errStopProvision, "cannot retrieve origin PVC (DataSourceRef is not a PVC)")
		return nil, controller.ProvisioningFinished, &controller.IgnoredError{Reason: "DataSourceRef is not a PVC"}
	}

	// Get origin PVC
	originKey := types.NamespacedName{
		Name:      mirrorPVC.Spec.DataSourceRef.Name,
		Namespace: *mirrorPVC.Spec.DataSourceRef.Namespace,
	}

	var originPVC corev1.PersistentVolumeClaim
	if err := p.Client.Get(ctx, originKey, &originPVC); err != nil {
		if kerrors.IsNotFound(err) {
			p.Logger.Error(err, "origin PVC does not exist")
			return nil, controller.ProvisioningFinished, status.Error(codes.InvalidArgument, "Origin PVC does not exist")
		}

		p.Logger.Error(err, "could not retrieve origin PVC")
		return nil, controller.ProvisioningFinished, err
	}

	// Check Authorization
	requiredLabels, present := originPVC.Annotations[AuthorizationAnnotationKey]
	if !present {
		// Default: deny-all
		p.Logger.Error(errSlowRetry, "no required labels specified on origin PVC, access denied")
		return nil, controller.ProvisioningFinished, status.Error(codes.InvalidArgument, "Unauthorized")
	}
	if proceed, err := utils.CheckNamespaceWithSelector(ctx, p.Client, mirrorPVC.Namespace, requiredLabels); !present || !proceed {
		if err != nil {
			p.Logger.Error(err, "failed checking namespace selector")
			return nil, controller.ProvisioningFinished, err
		}

		p.Logger.Error(errSlowRetry, "namespace does not match the selector on the origin PVC, access denied")
		return nil, controller.ProvisioningFinished, status.Error(codes.InvalidArgument, "Unauthorized")
	}

	// Check origin PVC's phase
	switch originPVC.Status.Phase {
	case corev1.ClaimPending:
		p.Logger.Error(errSlowRetry, errPendingPVC.Error())
		return nil, controller.ProvisioningFinished, errPendingPVC
	case corev1.ClaimLost:
		p.Logger.Error(errStopProvision, "PV has been lost and cannot be mirrored")
		return nil, controller.ProvisioningFinished, &controller.IgnoredError{Reason: "PVC is in Lost phase"}
	case corev1.ClaimBound:
		// continue
	default:
		p.Logger.Error(errSlowRetry, strings.ReplaceAll(errPendingPVC.Error(), "pending", "no-phase"))
		return nil, controller.ProvisioningFinished, errPendingPVC
	}

	// Check origin PV's CSI spec
	var originPV corev1.PersistentVolume
	if err := p.Client.Get(ctx, types.NamespacedName{Name: originPVC.Spec.VolumeName}, &originPV); err != nil {
		// Surely it's not a NotFound error
		p.Logger.Error(err, "could not retrieve origin PV")
		return nil, controller.ProvisioningFinished, err
	}
	if originPV.Spec.CSI == nil {
		p.Logger.Error(errStopProvision, "Cannot retrieve CSI spec from PV (CSI is missing)")
		return nil, controller.ProvisioningFinished, &controller.IgnoredError{Reason: "PV's CSI spec is missing"}
	}

	// Create mirror PV
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName, // By default: "pvc-" + claim.UID
			Labels: map[string]string{
				MirroredPvLabel:  originPV.Name,
				MirroredPvcLabel: originPVC.Namespace + "---" + originPVC.Name,
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *options.StorageClass.ReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: defaultMirrorCapacity,
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: originPV.Spec.CSI,
			},
			MountOptions:     originPV.Spec.MountOptions,
			StorageClassName: options.StorageClass.Name,
		},
	}

	return pv, controller.ProvisioningFinished, nil
}

// Delete is the Provisioner interface function called when a PVC has to be deleted.
func (p *PvcMirrorProvisioner) Delete(_ context.Context, _ *corev1.PersistentVolume) error {
	// Nothing to be done here, since there is no "backed" volume
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

	if outputSc.Provisioner != p.MirrorProvisionerName || *outputSc.AllowVolumeExpansion {
		p.Logger.Error(errInvalidConfig, "mismatching Provisioner or invalid storageclass AllowVolumeExpansion")
		return errInvalidConfig
	}

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

// ShouldDelete is the DeletionGuard interface function that returns whether deleting the PV should be attempted.
func (p *PvcMirrorProvisioner) ShouldDelete(context.Context, *corev1.PersistentVolume) bool {
	// If needed, checks to be done before deleting a PV can be added here
	return true
}
