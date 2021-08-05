// Copyright 2020-2021 Politecnico di Torino
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

package instance_controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	virtv1 "kubevirt.io/client-go/api/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// RetrievePhaseFromVM converts the VM phase to the corresponding one of the instance.
func (r *InstanceReconciler) RetrievePhaseFromVM(vm *virtv1.VirtualMachine) clv1alpha2.EnvironmentPhase {
	switch vm.Status.PrintableStatus {
	case virtv1.VirtualMachineStatusStarting:
		return clv1alpha2.EnvironmentPhaseStarting
	case virtv1.VirtualMachineStatusProvisioning:
		return clv1alpha2.EnvironmentPhaseImporting

	case virtv1.VirtualMachineStatusStopping:
		return clv1alpha2.EnvironmentPhaseStopping
	case virtv1.VirtualMachineStatusTerminating:
		return clv1alpha2.EnvironmentPhaseStopping
	case virtv1.VirtualMachineStatusStopped:
		return clv1alpha2.EnvironmentPhaseOff

	case virtv1.VirtualMachineStatusRunning:
		if vm.Status.Ready {
			return clv1alpha2.EnvironmentPhaseReady
		}
		return clv1alpha2.EnvironmentPhaseRunning

	default:
		return clv1alpha2.EnvironmentPhaseUnset
	}
}

// RetrievePhaseFromVMI converts the VMI phase to the corresponding one of the instance.
func (r *InstanceReconciler) RetrievePhaseFromVMI(vmi *virtv1.VirtualMachineInstance) clv1alpha2.EnvironmentPhase {
	if !vmi.DeletionTimestamp.IsZero() {
		return clv1alpha2.EnvironmentPhaseStopping
	}

	switch vmi.Status.Phase {
	case virtv1.Pending:
		return clv1alpha2.EnvironmentPhaseStarting
	case virtv1.Scheduling:
		return clv1alpha2.EnvironmentPhaseStarting
	case virtv1.Scheduled:
		return clv1alpha2.EnvironmentPhaseStarting

	case virtv1.Unknown:
		return clv1alpha2.EnvironmentPhaseFailed
	case virtv1.Failed:
		return clv1alpha2.EnvironmentPhaseFailed
	case virtv1.Succeeded:
		return clv1alpha2.EnvironmentPhaseFailed

	case virtv1.Running:
		if isVMIReady(vmi) {
			return clv1alpha2.EnvironmentPhaseReady
		}
		return clv1alpha2.EnvironmentPhaseRunning

	default:
		return clv1alpha2.EnvironmentPhaseUnset
	}
}

// RetrievePhaseFromDeployment converts the Deployment phase to the corresponding one of the instance.
func (r *InstanceReconciler) RetrievePhaseFromDeployment(deployment *appsv1.Deployment) clv1alpha2.EnvironmentPhase {
	if !deployment.DeletionTimestamp.IsZero() {
		return clv1alpha2.EnvironmentPhaseStopping
	}

	if *deployment.Spec.Replicas == 0 {
		return clv1alpha2.EnvironmentPhaseOff
	}

	switch *deployment.Spec.Replicas {
	case 0:
		return clv1alpha2.EnvironmentPhaseOff
	case deployment.Status.ReadyReplicas:
		return clv1alpha2.EnvironmentPhaseReady
	default:
		return clv1alpha2.EnvironmentPhaseStarting
	}
}

// isVMIReady checks whether a VMI is ready, depending on its conditions.
func isVMIReady(vmi *virtv1.VirtualMachineInstance) bool {
	for _, condition := range vmi.Status.Conditions {
		if condition.Type == virtv1.VirtualMachineInstanceReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}
