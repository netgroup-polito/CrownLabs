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

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	virtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// EnforceVMEnvironment implements the logic to create all the different
// Kubernetes resources required to start a CrownLabs environment.
func (r *InstanceReconciler) EnforceVMEnvironment(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	// Enforce the cloud-init secret when environment is not restricted
	if environment.Mode == clv1alpha2.ModeStandard {
		if err := r.EnforceCloudInitSecret(ctx); err != nil {
			log.Error(err, "failed to enforce the cloud-init secret existence")
			return err
		}
	}

	// Enforce the service and the ingress to expose the environment.
	err := r.EnforceInstanceExposition(ctx)
	if err != nil {
		log.Error(err, "failed to enforce the instance exposition objects")
		return err
	}

	// Create a VirtualMachine if the environment is persistent.
	if environment.Persistent {
		return r.enforceVirtualMachine(ctx)
	}

	// Create a VirtualMachineInstance if the environment is not persistent.
	return r.enforceVirtualMachineInstance(ctx)
}

// enforceVirtualMachine enforces the presence of the VirtualMachine, and updates the instance phase based on its status.
func (r *InstanceReconciler) enforceVirtualMachine(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	vm := virtv1.VirtualMachine{ObjectMeta: forge.ObjectMeta(instance)}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &vm, func() error {
		// VirtualMachine specifications are forged only at creation time, as changing them later may be
		// either rejected by the webhook or cause the restart of the child VMI, with consequent possible data loss.
		if vm.CreationTimestamp.IsZero() {
			vm.Spec = forge.VirtualMachineSpec(instance, environment)
		}
		// Afterwards, the only modification to the specifications is performed to configure the running flag.
		vm.Spec.Running = ptr.To(instance.Spec.Running)
		vm.SetLabels(forge.InstanceObjectLabels(vm.GetLabels(), instance))
		return ctrl.SetControllerReference(instance, &vm, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to enforce virtualmachine", "virtualmachine", klog.KObj(&vm))
		return err
	}
	log.V(utils.FromResult(res)).Info("virtualmachine enforced", "virtualmachine", klog.KObj(&vm), "result", res)

	// It is necessary to retrieve the VMI object associated with the VM (if any), to correctly detect the ResourceQuotaExceeded phase.
	// VM and VMI are characterized by the same resource name.
	vmi := virtv1.VirtualMachineInstance{ObjectMeta: forge.ObjectMeta(instance)}
	if err = r.Get(ctx, client.ObjectKeyFromObject(&vmi), &vmi); client.IgnoreNotFound(err) != nil {
		log.Error(err, "failed to retrieve virtualmachineinstance", "virtualmachineinstance", klog.KObj(&vm))
		return err
	} else if err != nil {
		klog.Infof("VMI %s doesn't exist", instance.Name)
	}
	phase := r.RetrievePhaseFromVM(&vm, &vmi)

	if phase != instance.Status.Phase {
		log.Info("phase changed", "virtualmachine", klog.KObj(&vm),
			"previous", string(instance.Status.Phase), "current", string(phase))
		instance.Status.Phase = phase
	}

	return nil
}

// enforceVirtualMachineInstance enforces the presence of the VirtualMachineInstance, and updates the instance phase based on its status.
func (r *InstanceReconciler) enforceVirtualMachineInstance(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	vmi := virtv1.VirtualMachineInstance{ObjectMeta: forge.ObjectMeta(instance)}
	var phase clv1alpha2.EnvironmentPhase

	// If the Instance is not running, we do not enforce the VirtualMachineInstance presence.
	// Yet, we cannot terminate it, as it would lead to data loss. All in all, shutting down
	// an ephemeral instance only means un-exposing it.
	if instance.Spec.Running {
		res, err := ctrl.CreateOrUpdate(ctx, r.Client, &vmi, func() error {
			// VirtualMachineInstance specifications are forged only at creation time, as changing them later may be
			// either rejected by the webhook or cause the restart of the VMI itself, with consequent data loss.
			if vmi.CreationTimestamp.IsZero() {
				vmi.Spec = forge.VirtualMachineInstanceSpec(instance, environment)
			}
			vmi.SetLabels(forge.InstanceObjectLabels(vmi.GetLabels(), instance))
			return ctrl.SetControllerReference(instance, &vmi, r.Scheme)
		})

		if err != nil {
			log.Error(err, "failed to enforce virtualmachineinstance", "virtualmachineinstance", klog.KObj(&vmi))
			return err
		}
		log.V(utils.FromResult(res)).Info("virtualmachineinstance enforced", "virtualmachineinstance", klog.KObj(&vmi), "result", res)
		phase = r.RetrievePhaseFromVMI(&vmi)
	} else {
		phase = clv1alpha2.EnvironmentPhaseOff
	}

	if phase != instance.Status.Phase {
		log.Info("phase changed", "virtualmachineinstance", klog.KObj(&vmi),
			"previous", string(instance.Status.Phase), "current", string(phase))
		instance.Status.Phase = phase
	}

	return nil
}
