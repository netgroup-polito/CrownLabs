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

package forge

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	virtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	urlDockerPrefix = "docker://"

	//nolint:gosec // The constant refers to the name of a secret, and it is not a secret itself.
	registryCredentialsSecretName = "registry-credentials"
	//nolint:gosec // The constant refers to the name of a secret, and it is not a secret itself.
	cdiSecretName = "registry-credentials-cdi"

	volumeRootName      = "root"
	volumeCloudInitName = "cloud-init"
	virtioDiskType      = "virtio"

	// terminationGracePeriod -> the amount of seconds before a terminating VM is forcefully deleted.
	terminationGracePeriod = 60
)

var (
	// cpuHypervisorOverhead -> the CPU overhead added to the reservation to account for the hypervisor.
	cpuHypervisorOverhead = *resource.NewScaledQuantity(500, resource.Milli)
	// memoryHypervisorOverhead -> the memory overhead added to the reservation to account for the hypervisor.
	memoryHypervisorOverhead = *resource.NewScaledQuantity(500, resource.Mega)
)

// VirtualMachineSpec forges the specification of a Kubevirt VirtualMachine object
// representing the definition of the VM corresponding to a persistent CrownLabs environment.
func VirtualMachineSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) virtv1.VirtualMachineSpec {
	return virtv1.VirtualMachineSpec{
		Template: &virtv1.VirtualMachineInstanceTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: InstanceSelectorLabels(instance)},
			Spec:       VirtualMachineInstanceSpec(instance, environment),
		},
		DataVolumeTemplates: []virtv1.DataVolumeTemplateSpec{
			DataVolumeTemplate(NamespacedName(instance).Name, environment),
		},
	}
}

// VirtualMachineInstanceSpec forges the specification of a Kubevirt VirtualMachineInstance
// object representing the definition of the VMI corresponding to a non-persistent CrownLabs Environment.
func VirtualMachineInstanceSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) virtv1.VirtualMachineInstanceSpec {
	return virtv1.VirtualMachineInstanceSpec{
		Domain:                        VirtualMachineDomain(environment),
		Volumes:                       Volumes(instance, environment),
		ReadinessProbe:                VirtualMachineReadinessProbe(environment),
		Networks:                      []virtv1.Network{*virtv1.DefaultPodNetwork()},
		TerminationGracePeriodSeconds: ptr.To[int64](terminationGracePeriod),
		NodeSelector:                  NodeSelectorLabels(instance, environment),
	}
}

// VirtualMachineDomain forges the specification of the domain of a Kubevirt VirtualMachineInstance
// object representing the definition of the VM corresponding to a given CrownLabs Environment.
func VirtualMachineDomain(environment *clv1alpha2.Environment) virtv1.DomainSpec {
	return virtv1.DomainSpec{
		CPU:       &virtv1.CPU{Cores: environment.Resources.CPU},
		Memory:    &virtv1.Memory{Guest: &environment.Resources.Memory},
		Resources: VirtualMachineResources(environment),
		Devices: virtv1.Devices{
			Disks:      VolumeDiskTargets(environment),
			Interfaces: []virtv1.Interface{*virtv1.DefaultBridgeNetworkInterface()},
		},
	}
}

// Volumes forges the array of volumes to be mounted onto the VMI specification.
func Volumes(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) []virtv1.Volume {
	volumes := []virtv1.Volume{VolumeRootDisk(instance, environment)}
	// Attach cloudinit volume on non-restricted environments
	if environment.Mode == clv1alpha2.ModeStandard {
		volumes = append(volumes, VolumeCloudInit(NamespacedName(instance).Name))
	}
	return volumes
}

// VolumeRootDisk forges the specification of the root volume, either ephemeral or persistent based on
// the environment characteristics.
func VolumeRootDisk(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) virtv1.Volume {
	if environment.Persistent {
		return VolumePersistentDisk(NamespacedName(instance).Name)
	}
	return VolumeContainerDisk(environment.Image)
}

// VolumePersistentDisk forges the specification of a volume mapping a DataVolume containing the root image.
func VolumePersistentDisk(dataVolumeName string) virtv1.Volume {
	return virtv1.Volume{
		Name: volumeRootName,
		VolumeSource: virtv1.VolumeSource{
			DataVolume: &virtv1.DataVolumeSource{
				Name: dataVolumeName,
			},
		},
	}
}

// VolumeContainerDisk forges the specification of a volume mapping an ephemeral container containing the root image.
func VolumeContainerDisk(image string) virtv1.Volume {
	return virtv1.Volume{
		Name: volumeRootName,
		VolumeSource: virtv1.VolumeSource{
			ContainerDisk: &virtv1.ContainerDiskSource{
				Image:           image,
				ImagePullSecret: registryCredentialsSecretName,
				ImagePullPolicy: corev1.PullIfNotPresent,
			},
		},
	}
}

// VolumeCloudInit forges the specification of a volume mapping to a secret containing the cloud-init configuration.
func VolumeCloudInit(secretName string) virtv1.Volume {
	return virtv1.Volume{
		Name: volumeCloudInitName,
		VolumeSource: virtv1.VolumeSource{
			CloudInitNoCloud: &virtv1.CloudInitNoCloudSource{
				UserDataSecretRef: &corev1.LocalObjectReference{Name: secretName},
			},
		},
	}
}

// VolumeDiskTargets forges the array of disks to be attached to the VM Domain.
func VolumeDiskTargets(environment *clv1alpha2.Environment) []virtv1.Disk {
	disks := []virtv1.Disk{VolumeDiskTarget(volumeRootName)}
	// Attach cloudinit disk on non-restricted environments
	if environment.Mode == clv1alpha2.ModeStandard {
		disks = append(disks, VolumeDiskTarget(volumeCloudInitName))
	}
	return disks
}

// VolumeDiskTarget forges the specification of a KVM disk attached to volume.
func VolumeDiskTarget(name string) virtv1.Disk {
	return virtv1.Disk{
		Name: name,
		DiskDevice: virtv1.DiskDevice{
			Disk: &virtv1.DiskTarget{
				Bus: virtioDiskType,
			},
		},
	}
}

// VirtualMachineResources forges the resource requirements for a given VM environment.
func VirtualMachineResources(environment *clv1alpha2.Environment) virtv1.ResourceRequirements {
	return virtv1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    VirtualMachineCPURequests(environment),
			corev1.ResourceMemory: VirtualMachineMemoryRequirements(environment),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    VirtualMachineCPULimits(environment),
			corev1.ResourceMemory: VirtualMachineMemoryRequirements(environment),
		},
	}
}

// VirtualMachineCPURequests computes the CPU requests based on a given environment.
func VirtualMachineCPURequests(environment *clv1alpha2.Environment) resource.Quantity {
	cpu := int64(10 * environment.Resources.CPU * environment.Resources.ReservedCPUPercentage)
	return *resource.NewScaledQuantity(cpu, resource.Milli)
}

// VirtualMachineCPULimits computes the CPU limits based on a given environment.
func VirtualMachineCPULimits(environment *clv1alpha2.Environment) resource.Quantity {
	cpu := resource.NewQuantity(int64(environment.Resources.CPU), resource.DecimalSI)
	cpu.Add(cpuHypervisorOverhead)
	return *cpu
}

// VirtualMachineMemoryRequirements computes the memory requirements based on a given environment.
func VirtualMachineMemoryRequirements(environment *clv1alpha2.Environment) resource.Quantity {
	memory := environment.Resources.Memory.DeepCopy()
	memory.Add(memoryHypervisorOverhead)
	return memory
}

// VirtualMachineReadinessProbe forges the readiness probe for a given VM environment.
func VirtualMachineReadinessProbe(environment *clv1alpha2.Environment) *virtv1.Probe {
	port := SSHPortNumber
	if environment.GuiEnabled {
		port = GUIPortNumber
	}

	return &virtv1.Probe{
		InitialDelaySeconds: 10,
		PeriodSeconds:       2,
		FailureThreshold:    5,
		SuccessThreshold:    2,
		Handler: virtv1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(port),
			},
		},
	}
}

// DataVolumeSourceForge forges the DataVolumeSource for DataVolumeTemplate.
func DataVolumeSourceForge(environment *clv1alpha2.Environment) *cdiv1beta1.DataVolumeSource {
	if environment.EnvironmentType == clv1alpha2.ClassCloudVM {
		return &cdiv1beta1.DataVolumeSource{
			HTTP: &cdiv1beta1.DataVolumeSourceHTTP{
				URL: environment.Image,
			},
		}
	}
	return &cdiv1beta1.DataVolumeSource{
		Registry: &cdiv1beta1.DataVolumeSourceRegistry{
			URL:       ptr.To(urlDockerPrefix + environment.Image),
			SecretRef: ptr.To(cdiSecretName),
		},
	}
}

// DataVolumeTemplate forges the DataVolume template associated with a given environment.
func DataVolumeTemplate(name string, environment *clv1alpha2.Environment) virtv1.DataVolumeTemplateSpec {
	return virtv1.DataVolumeTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: cdiv1beta1.DataVolumeSpec{
			Source: DataVolumeSourceForge(environment),
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: environment.Resources.Disk,
					},
				},
			},
		},
	}
}
