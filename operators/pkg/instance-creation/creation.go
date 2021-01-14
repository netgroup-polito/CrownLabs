// Package instance_creation groups the functionalities related to the
// creation of the different Kubernetes objects required by the Instance controller
package instance_creation

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	virtv1 "kubevirt.io/client-go/api/v1"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var terminationGracePeriod int64 = 30
var cpuHypervisorReserved float32 = 0.5
var memoryHypervisorReserved string = "500M"
var registryCred string = "registry-credentials"

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
			},
		},
	}
}

// CreateVolumeContainerdiskPVC returns containerdisk volume for persistent VM
// object pvc is linked to the VM.
func CreateVolumeContainerdiskPVC(template *crownlabsv1alpha2.Environment, name string) virtv1.Volume {
	volume := virtv1.Volume{
		Name: "containerdisk",
		VolumeSource: virtv1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: name,
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
		Machine: virtv1.Machine{},
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

// UpdateDataVolumeSpec create datavolume specification.
// object allows the creation of the pvc and the import of the virtual machine image.
func UpdateDataVolumeSpec(dv *cdiv1.DataVolume, template *crownlabsv1alpha2.Environment) {
	dv.Spec = cdiv1.DataVolumeSpec{
		Source: cdiv1.DataVolumeSource{
			Registry: &cdiv1.DataVolumeSourceRegistry{
				URL:       "docker://" + template.Image,
				SecretRef: "registry-credentials-cdi",
			},
		},
		PVC: &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: template.Resources.Disk,
				},
			},
		},
	}
}

func computeCPULimits(cpu uint32, hypervisorCoefficient float32) string {
	return fmt.Sprintf("%f", float32(cpu)+hypervisorCoefficient)
}

func computeCPURequests(cpu, percentage uint32) string {
	return fmt.Sprintf("%f", float32(cpu*percentage)/100)
}

// CreateOrUpdate creates a resource or updates it if already exists.
// Deprecated: use https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
func CreateOrUpdate(ctx context.Context, c client.Client, object interface{}) error {
	switch obj := object.(type) {
	case corev1.Secret:
		var sec corev1.Secret
		err := c.Get(ctx, types.NamespacedName{
			Namespace: obj.Namespace,
			Name:      obj.Name,
		}, &sec)
		if err != nil {
			err = c.Create(ctx, &obj, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				klog.Error("Unable to create secret " + obj.Name)
				klog.Error(err)
				return err
			}
		}
	case corev1.PersistentVolumeClaim:
		var pvc corev1.PersistentVolumeClaim
		err := c.Get(ctx, types.NamespacedName{
			Namespace: obj.Namespace,
			Name:      obj.Name,
		}, &pvc)
		if err != nil {
			err = c.Create(ctx, &obj, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				klog.Error("Unable to create pvc " + obj.Name)
				klog.Error(err)
				return err
			}
		} else {
			return errors.NewBadRequest("ALREADY EXISTS")
		}
	case corev1.Service:
		var svc corev1.Service
		err := c.Get(ctx, types.NamespacedName{
			Namespace: obj.Namespace,
			Name:      obj.Name,
		}, &svc)
		if err != nil {
			err = c.Create(ctx, &obj, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				klog.Error("Unable to create service " + obj.Name)
				klog.Error(err)
				return err
			}
		}
	case networkingv1.Ingress:
		var ing networkingv1.Ingress
		err := c.Get(ctx, types.NamespacedName{
			Namespace: obj.Namespace,
			Name:      obj.Name,
		}, &ing)
		if err != nil && errors.IsNotFound(err) {
			err = c.Create(ctx, &obj, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				klog.Error("Unable to create Ingress " + obj.Name)
				klog.Error(err)
			}
		} else {
			klog.Error("Unable to create an Ingress " + obj.Name)
			klog.Error(err)
			return err
		}

	case appsv1.Deployment:
		var deploy appsv1.Deployment
		err := c.Get(ctx, types.NamespacedName{
			Namespace: obj.Namespace,
			Name:      obj.Name,
		}, &deploy)
		if err != nil {
			err = c.Create(ctx, &obj, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				klog.Error("Unable to create deployment " + obj.Name)
				klog.Error(err)
				return err
			}
		} else {
			err = c.Update(ctx, &obj, &client.UpdateOptions{})
			if err != nil {
				klog.Error("unable to update deployment " + obj.Name)
				klog.Error(err)
				return err
			}
		}
	case *virtv1.VirtualMachineInstance:
		err := c.Create(ctx, obj, &client.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			klog.Error("Unable to create virtual machine " + obj.Name)
			klog.Error(err)
			return err
		}
	default:
		return errors.NewBadRequest("No matching type for requested resource:")
	}

	return nil
}

// CheckLabels verifies whether a namespace is characterized by a set of
// required labels.
func CheckLabels(ns *corev1.Namespace, matchLabels map[string]string) bool {
	for key, value := range matchLabels {
		if v1, ok := ns.Labels[key]; !ok || v1 != value {
			return false
		}
	}
	return true
}
