// Package instance_creation groups the functionalities related to the
// creation of the different Kubernetes objects required by the Instance controller
package instance_creation

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	virtv1 "kubevirt.io/client-go/api/v1"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var terminationGracePeriod int64 = 30
var cpuHypervisorReserved float32 = 0.5
var memoryHypervisorReserved = "500M"
var registryCred = "registry-credentials"

// UpdateVirtualMachineInstanceSpec updates the specification of a Kubevirt VirtualMachineInstance
// object representing the definition of the VM corresponding to a given CrownLabs Environment.
func UpdateVirtualMachineInstanceSpec(vmi *virtv1.VirtualMachineInstance, template *crownlabsv1alpha2.Environment) {
	domain := UpdateVMdomain(template)
	containerdisk := CreateVolumeContainerdisk(template, vmi.Name)
	cloudinitdisk := CreateVolumeCloudinitdisk(vmi.Name)
	vmi.Spec = virtv1.VirtualMachineInstanceSpec{
		TerminationGracePeriodSeconds: &terminationGracePeriod,
		Domain:                        domain,
		Volumes: []virtv1.Volume{
			containerdisk,
			cloudinitdisk,
		},
		ReadinessProbe: forge.VMReadinessProbe(template),
		Networks:       []virtv1.Network{*virtv1.DefaultPodNetwork()},
	}
}

// UpdateVirtualMachineSpec updates the specification of a Kubevirt VirtualMachineInstance
// object representing the definition of the VM corresponding to a persistent Crownlabs environment.
func UpdateVirtualMachineSpec(vm *virtv1.VirtualMachine, template *crownlabsv1alpha2.Environment, running bool) {
	domain := UpdateVMdomain(template)
	containerdisk := CreateVolumeContainerdiskPVC(template, vm.Name)
	cloudinitdisk := CreateVolumeCloudinitdisk(vm.Name)
	vm.Spec = virtv1.VirtualMachineSpec{
		Running: &running,
		Template: &virtv1.VirtualMachineInstanceTemplateSpec{
			Spec: virtv1.VirtualMachineInstanceSpec{
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Domain:                        domain,
				Volumes: []virtv1.Volume{
					containerdisk,
					cloudinitdisk,
				},
				ReadinessProbe: forge.VMReadinessProbe(template),
				Networks:       []virtv1.Network{*virtv1.DefaultPodNetwork()},
			},
		},
		DataVolumeTemplates: []virtv1.DataVolumeTemplateSpec{},
	}
}

// CreateVolumeContainerdiskPVC returns containerdisk volume for persistent VM
// object pvc is linked to the VM.
func CreateVolumeContainerdiskPVC(template *crownlabsv1alpha2.Environment, name string) virtv1.Volume {
	volume := virtv1.Volume{
		Name: "containerdisk",
		VolumeSource: virtv1.VolumeSource{
			DataVolume: &virtv1.DataVolumeSource{
				Name: name,
			},
		},
	}
	return volume
}

// CreateVolumeContainerdisk returns containerdisk volume for non persistent VMs.
func CreateVolumeContainerdisk(template *crownlabsv1alpha2.Environment, name string) virtv1.Volume {
	volume := virtv1.Volume{
		Name: "containerdisk",
		VolumeSource: virtv1.VolumeSource{
			ContainerDisk: &virtv1.ContainerDiskSource{
				Image:           template.Image,
				ImagePullSecret: registryCred,
				ImagePullPolicy: corev1.PullIfNotPresent,
			},
		},
	}
	return volume
}

// CreateVolumeCloudinitdisk returns cloudinitdisk volume.
// object contains the cloud-init configuration.
func CreateVolumeCloudinitdisk(secretName string) virtv1.Volume {
	volume := virtv1.Volume{
		Name: "cloudinitdisk",
		VolumeSource: virtv1.VolumeSource{
			CloudInitNoCloud: &virtv1.CloudInitNoCloudSource{
				UserDataSecretRef: &corev1.LocalObjectReference{Name: secretName},
			},
		},
	}
	return volume
}

// UpdateVMdomain returns the updated domain field for the all the VMs.
// represents all the resources and the disks of which the vm is composed.
func UpdateVMdomain(template *crownlabsv1alpha2.Environment) virtv1.DomainSpec {
	vmMemory := template.Resources.Memory
	template.Resources.Memory.Add(resource.MustParse(memoryHypervisorReserved))
	Domain := virtv1.DomainSpec{
		Resources: virtv1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"cpu":    resource.MustParse(computeCPURequests(template.Resources.CPU, template.Resources.ReservedCPUPercentage)),
				"memory": template.Resources.Memory,
			},
			Limits: corev1.ResourceList{
				"cpu":    resource.MustParse(computeCPULimits(template.Resources.CPU, cpuHypervisorReserved)),
				"memory": template.Resources.Memory,
			},
		},
		CPU: &virtv1.CPU{
			Cores: template.Resources.CPU,
		},
		Memory: &virtv1.Memory{
			Guest: &vmMemory,
		},
		Machine: &virtv1.Machine{},
		Devices: virtv1.Devices{
			Disks: []virtv1.Disk{
				{
					Name: "containerdisk",
					DiskDevice: virtv1.DiskDevice{
						Disk: &virtv1.DiskTarget{
							Bus: "virtio",
						},
					},
				},
				{
					Name: "cloudinitdisk",
					DiskDevice: virtv1.DiskDevice{
						Disk: &virtv1.DiskTarget{
							Bus: "virtio",
						},
					},
				},
			},
			Interfaces: []virtv1.Interface{*virtv1.DefaultBridgeNetworkInterface()},
		},
	}
	return Domain
}

// UpdateLabels is a function that modifies the  labels map for VMs and VMIs.
func UpdateLabels(labels map[string]string, template *crownlabsv1alpha2.Environment, name string) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 1)
	}
	labels["name"] = name
	labels["crownlabs.polito.it/template"] = template.Name
	labels["crownlabs.polito.it/managed-by"] = "instance"
	return labels
}

func computeCPULimits(cpu uint32, hypervisorCoefficient float32) string {
	return fmt.Sprintf("%f", float32(cpu)+hypervisorCoefficient)
}

func computeCPURequests(cpu, percentage uint32) string {
	return fmt.Sprintf("%f", float32(cpu*percentage)/100)
}
