package instance_controller

import (
	"context"

	"k8s.io/klog/v2"
	virtv1 "kubevirt.io/client-go/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	instance_creation "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// EnforceVMEnvironment implements the logic to create all the different
// Kubernetes resources required to start a CrownLabs environment.
func (r *InstanceReconciler) EnforceVMEnvironment(ctx context.Context, instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) error {
	log := ctrl.LoggerFrom(ctx, "environment", environment.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	// Enforce the cloud-init secret
	if err := r.EnforceCloudInitSecret(ctx, instance); err != nil {
		log.Error(err, "failed to enforce the cloud-init secret existence")
		return err
	}

	// Create the service and the ingress to expose the environment.
	// nolint:dogsled // this will be fixed as part of the next refactoring.
	_, _, _, err := r.CreateInstanceExpositionEnvironment(ctx, instance, false)
	if err != nil {
		return err
	}

	// Create a VirtualMachine if the environment is persistent.
	if environment.Persistent {
		return r.enforceVirtualMachine(ctx, instance, environment)
	}

	// Create a VirtualMachineInstance if the environment is not persistent.
	return r.enforceVirtualMachineInstance(ctx, instance, environment)
}

// enforceVirtualMachine enforces the presence of the VirtualMachine, and updates the instance phase based on its status.
func (r *InstanceReconciler) enforceVirtualMachine(ctx context.Context, instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) error {
	log := ctrl.LoggerFrom(ctx)

	vm := virtv1.VirtualMachine{ObjectMeta: forge.ObjectMeta(instance)}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &vm, func() error {
		if vm.CreationTimestamp.IsZero() {
			vm.Spec.DataVolumeTemplates = []virtv1.DataVolumeTemplateSpec{
				forge.DataVolumeTemplate(vm.Name, environment),
			}
		}
		instance_creation.UpdateVirtualMachineSpec(&vm, environment, instance.Spec.Running)
		vm.Spec.Template.ObjectMeta.Labels = instance_creation.UpdateLabels(vm.Spec.Template.ObjectMeta.Labels, environment, vm.GetName())
		return ctrl.SetControllerReference(instance, &vm, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to create virtualmachine", "virtualmachine", klog.KObj(&vm))
		return err
	}
	log.V(utils.FromResult(res)).Info("virtualmachine enforced", "virtualmachine", klog.KObj(&vm), "result", res)

	phase := r.RetrievePhaseFromVM(&vm)
	if phase != instance.Status.Phase {
		log.Info("phase changed", "virtualmachine", klog.KObj(&vm),
			"previous", string(instance.Status.Phase), "current", string(phase))
		instance.Status.Phase = phase
	}

	return nil
}

// enforceVirtualMachineInstance enforces the presence of the VirtualMachineInstance, and updates the instance phase based on its status.
func (r *InstanceReconciler) enforceVirtualMachineInstance(ctx context.Context, instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) error {
	log := ctrl.LoggerFrom(ctx)

	vmi := virtv1.VirtualMachineInstance{ObjectMeta: forge.ObjectMeta(instance)}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &vmi, func() error {
		if vmi.ObjectMeta.CreationTimestamp.IsZero() {
			instance_creation.UpdateVirtualMachineInstanceSpec(&vmi, environment)
		}
		vmi.Labels = instance_creation.UpdateLabels(vmi.Labels, environment, vmi.GetName())
		return ctrl.SetControllerReference(instance, &vmi, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to create virtualmachineinstance", "virtualmachineinstance", klog.KObj(&vmi))
		return err
	}
	log.V(utils.FromResult(res)).Info("virtualmachineinstance enforced", "virtualmachineinstance", klog.KObj(&vmi), "result", res)

	phase := r.RetrievePhaseFromVMI(&vmi)
	if phase != instance.Status.Phase {
		log.Info("phase changed", "virtualmachineinstance", klog.KObj(&vmi),
			"previous", string(instance.Status.Phase), "current", string(phase))
		instance.Status.Phase = phase
	}

	return nil
}
