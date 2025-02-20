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

package instctrl

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// EnforceContainerEnvironment implements the logic to create all the different
// Kubernetes resources required to start a containerized CrownLabs environment.
func (r *InstanceReconciler) EnforceContainerEnvironment(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	// Enforce the service and the ingress to expose the environment.
	if err := r.EnforceInstanceExposition(ctx); err != nil {
		log.Error(err, "failed to enforce the instance exposition objects")
		return err
	}

	// Enforce the environment's PVC in case of a persistent environment
	if environment.Persistent {
		if err := r.enforcePVC(ctx); err != nil {
			return err
		}
	}

	return r.enforceContainer(ctx)
}

// enforcePVC enforces the presence of the instance persistent storage
// which consists in a PVC.
func (r *InstanceReconciler) enforcePVC(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	pvc := v1.PersistentVolumeClaim{ObjectMeta: forge.ObjectMeta(instance)}

	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		// PVC's spec is immutable, it has to be set at creation
		if pvc.ObjectMeta.CreationTimestamp.IsZero() {
			pvc.Spec = forge.InstancePVCSpec(environment)
		}
		pvc.SetLabels(forge.InstanceObjectLabels(pvc.GetLabels(), instance))
		return ctrl.SetControllerReference(instance, &pvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to enforce object", "pvc", klog.KObj(&pvc))
		return err
	}
	log.V(utils.FromResult(res)).Info("object enforced", "pvc", klog.KObj(&pvc), "result", res)
	return nil
}

// enforceContainer enforces the actual deployment
// which contains all the container based instance components.
func (r *InstanceReconciler) enforceContainer(ctx context.Context) error {
	var nfsServerName, nfsPath string
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	depl := appsv1.Deployment{ObjectMeta: forge.ObjectMeta(instance)}

	mountInfos := []forge.NFSVolumeMountInfo{}

	if environment.MountMyDriveVolume {
		var err error
		nfsServerName, nfsPath, err = r.GetNFSSpecs(ctx)
		if err != nil {
			log.Error(err, "can't get NFS spec")
			return err
		}

		mountInfos = append(mountInfos, forge.MyDriveNFSVolumeMountInfo(nfsServerName, nfsPath))
	}

	for i, mount := range environment.SharedVolumeMounts {
		var shvol clv1alpha2.SharedVolume
		if err := r.Get(ctx, forge.NamespacedNameFromMount(mount), &shvol); err != nil {
			log.Error(err, "unable to retrieve shvol to mount")
			return err
		}

		mountInfos = append(mountInfos, forge.ShVolNFSVolumeMountInfo(i, &shvol, mount))
	}

	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &depl, func() error {
		// Deployment specifications are forged only at creation time, as changing them later may be
		// either rejected or cause the restart of the Pod, with consequent possible data loss.
		if depl.CreationTimestamp.IsZero() {
			depl.Spec = forge.DeploymentSpec(instance, environment, mountInfos, &r.ContainerEnvOpts)
		}

		depl.Spec.Replicas = forge.ReplicasCount(instance, environment, depl.CreationTimestamp.IsZero())

		depl.SetLabels(forge.InstanceObjectLabels(depl.GetLabels(), instance))
		return ctrl.SetControllerReference(instance, &depl, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to enforce deployment existence", "deployment", klog.KObj(&depl))
		return err
	}

	log.V(utils.FromResult(res)).Info("object enforced", "deployment", klog.KObj(&depl), "result", res)

	phase := r.RetrievePhaseFromDeployment(&depl)

	// in case of non-running, non-persistent instances, just exposition is teared down
	// we consider them off even if the container is still on, to avoid data loss
	if !instance.Spec.Running && !environment.Persistent {
		phase = clv1alpha2.EnvironmentPhaseOff
	}

	if phase != instance.Status.Phase {
		log.Info("phase changed", "deployment", klog.KObj(&depl),
			"previous", string(instance.Status.Phase), "current", string(phase))
		instance.Status.Phase = phase
	}

	return nil
}
