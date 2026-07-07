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

package forge

import (
	"fmt"
	"strings"

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

	// PodBridgeNetworkLiveMigrationAnnotation enables live migration for VMs using the pod bridge network.
	PodBridgeNetworkLiveMigrationAnnotation = "kubevirt.io/allow-pod-bridge-network-live-migration"
	podBridgeNetworkLiveMigrationValue      = "true"
)

var (
	// cpuHypervisorOverhead -> the CPU overhead added to the reservation to account for the hypervisor.
	cpuHypervisorOverhead = *resource.NewScaledQuantity(500, resource.Milli)
	// memoryHypervisorOverhead -> the memory overhead added to the reservation to account for the hypervisor.
	memoryHypervisorOverhead = *resource.NewScaledQuantity(500, resource.Mega)
)

// VirtualMachineSpec forges the specification of a Kubevirt VirtualMachine object
// representing the definition of the VM corresponding to a persistent CrownLabs environment.
func VirtualMachineSpec(instance *clv1alpha2.Instance, template *clv1alpha2.Template, environment *clv1alpha2.Environment, mountInfos []corev1.VolumeMount) virtv1.VirtualMachineSpec {
	return virtv1.VirtualMachineSpec{
		Template: &virtv1.VirtualMachineInstanceTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: EnvironmentSelectorLabels(instance, environment)},
			Spec:       VirtualMachineInstanceSpec(instance, template, environment, mountInfos),
		},
	}
}

// VirtualMachineInstanceSpec forges the specification of a Kubevirt VirtualMachineInstance
// object representing the definition of the VMI corresponding to a non-persistent CrownLabs Environment.
func VirtualMachineInstanceSpec(instance *clv1alpha2.Instance, template *clv1alpha2.Template, environment *clv1alpha2.Environment, mountInfos []corev1.VolumeMount) virtv1.VirtualMachineInstanceSpec {
	return virtv1.VirtualMachineInstanceSpec{
		Domain:                        VirtualMachineDomain(environment, mountInfos),
		Volumes:                       Volumes(instance, environment, mountInfos),
		ReadinessProbe:                VirtualMachineReadinessProbe(environment),
		Networks:                      []virtv1.Network{*virtv1.DefaultPodNetwork()},
		TerminationGracePeriodSeconds: ptr.To[int64](terminationGracePeriod),
		NodeSelector:                  NodeSelectorLabels(instance, template),
	}
}

// Volumes forges the array of volumes to be mounted onto the VMI specification.
func Volumes(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, mountInfos []corev1.VolumeMount) []virtv1.Volume {
	return append([]virtv1.Volume{
		VolumeRootDisk(instance, environment),
		VolumeCloudInit(CanonicalName(instance.GetName())),
	}, AttachableVolumes(mountInfos)...)
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

// VirtualMachineDomain forges the specification of the domain of a Kubevirt VirtualMachineInstance
// object representing the definition of the VM corresponding to a given CrownLabs Environment.
func VirtualMachineDomain(environment *clv1alpha2.Environment, mountInfos []corev1.VolumeMount) virtv1.DomainSpec {
	return virtv1.DomainSpec{
		CPU:       &virtv1.CPU{Cores: environment.Resources.CPU},
		Memory:    &virtv1.Memory{Guest: &environment.Resources.Memory},
		Resources: VirtualMachineResources(environment),
		Devices: virtv1.Devices{
			Disks:       VolumeDiskTargets(environment),
			Filesystems: VirtualMachineFilesystems(mountInfos),
			Interfaces:  []virtv1.Interface{*virtv1.DefaultBridgeNetworkInterface()},
		},
	}
}

// AttachableVolumes forges the array of attachable volumes (MyDrive, SharedVolumes) to be mounted onto the VMI specification.
func AttachableVolumes(mountInfos []corev1.VolumeMount) []virtv1.Volume {
	volumes := []virtv1.Volume{}
	for _, mount := range mountInfos {
		volumes = append(volumes, VirtPVCVolume(&mount))
	}
	return volumes
}

// VirtPVCVolume forges the specification of an external volume (MyDrive, SharedVolume) to be mounted through PVC.
func VirtPVCVolume(mount *corev1.VolumeMount) virtv1.Volume {
	return virtv1.Volume{
		Name: mount.Name,
		VolumeSource: virtv1.VolumeSource{
			PersistentVolumeClaim: &virtv1.PersistentVolumeClaimVolumeSource{
				PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: mount.Name,
					ReadOnly:  mount.ReadOnly,
				},
			},
		},
	}
}

// VolumeRootDisk forges the specification of the root volume, either ephemeral or persistent based on
// the environment characteristics.
func VolumeRootDisk(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) virtv1.Volume {
	if environment.Persistent {
		return VolumePersistentDisk(NamespacedNameWithSuffix(instance, environment.Name).Name)
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

// VolumeDiskTargets forges the array of disks to be attached to the VM Domain.
func VolumeDiskTargets(_ *clv1alpha2.Environment) []virtv1.Disk {
	disks := []virtv1.Disk{VolumeDiskTarget(volumeRootName)}
	disks = append(disks, VolumeDiskTarget(volumeCloudInitName))
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

// VirtualMachineFilesystems forges the array of filesystems to be attached to the VM.
func VirtualMachineFilesystems(mountInfos []corev1.VolumeMount) []virtv1.Filesystem {
	fss := []virtv1.Filesystem{}

	for _, mount := range mountInfos {
		fss = append(fss, virtv1.Filesystem{
			Name:     mount.Name,
			Virtiofs: &virtv1.FilesystemVirtiofs{},
		})
	}
	return fss
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

// DataVolumeSourceForge forges the DataVolumeSource for DataVolume.
func DataVolumeSourceForge(environment *clv1alpha2.Environment) (*cdiv1beta1.DataVolumeSource, error) {
	// For ClassLocalVM, the DataVolume is created from a pre-existing PVC containing the golden image.
	if environment.EnvironmentType == clv1alpha2.ClassLocalVM {
		// Splitting the environment.Image
		// In case of LocalVM the string must be formatted as: namespace/pvc-name

		parts := strings.Split(environment.Image, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return &cdiv1beta1.DataVolumeSource{
				PVC: &cdiv1beta1.DataVolumeSourcePVC{
					Namespace: parts[0],
					Name:      parts[1],
				},
			}, nil
		}
		return nil, fmt.Errorf("invalid LocalVM image %q: expected namespace/pvc-name", environment.Image)
	}

	// For ClassCloudVM, the DataVolume is created from an HTTP source pointing to the image URL.
	if environment.EnvironmentType == clv1alpha2.ClassCloudVM {
		return &cdiv1beta1.DataVolumeSource{
			HTTP: &cdiv1beta1.DataVolumeSourceHTTP{
				URL: environment.Image,
			},
		}, nil
	}

	// For ClassVM, the DataVolume is created from a registry source.
	return &cdiv1beta1.DataVolumeSource{
		Registry: &cdiv1beta1.DataVolumeSourceRegistry{
			URL:       ptr.To(urlDockerPrefix + environment.Image),
			SecretRef: ptr.To(cdiSecretName),
		},
	}, nil
}

// DataVolumeSpec forges the spec of a DataVolume, which needs to be created before the VM.
func DataVolumeSpec(environment *clv1alpha2.Environment) (cdiv1beta1.DataVolumeSpec, error) {
	// Select the correct volume mode based on VM type. Defaults to FS, but for CloudVMs Block Mode is used
	volumeMode := corev1.PersistentVolumeFilesystem

	if environment.EnvironmentType == clv1alpha2.ClassCloudVM || environment.EnvironmentType == clv1alpha2.ClassLocalVM {
		volumeMode = corev1.PersistentVolumeBlock
	}

	source, err := DataVolumeSourceForge(environment)
	if err != nil {
		return cdiv1beta1.DataVolumeSpec{}, err
	}

	return cdiv1beta1.DataVolumeSpec{
		Source: source,
		PVC: &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			VolumeMode:  &volumeMode,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: environment.Resources.Disk,
				},
			},
		},
	}, nil
}
