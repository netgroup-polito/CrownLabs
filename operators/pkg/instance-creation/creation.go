package instance_creation

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	virtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var terminationGracePeriod int64 = 30
var CPUhypervisorReserved float32 = 0.5
var memoryHypervisorReserved string = "500M"

func CreateVirtualMachineInstance(name string, namespace string, template *crownlabsv1alpha2.Environment, instanceName string, secretName string, references []metav1.OwnerReference) (*virtv1.VirtualMachineInstance, error) {
	template.Resources.Memory.Add(resource.MustParse(memoryHypervisorReserved))
	vm := virtv1.VirtualMachineInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachineInstance",
			APIVersion: "kubevirt.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name + "-vmi",
			Namespace:       namespace,
			OwnerReferences: references,
			Labels:          map[string]string{"name": name, "template-name": template.Name, "instance-name": instanceName},
		},
		Spec: virtv1.VirtualMachineInstanceSpec{
			TerminationGracePeriodSeconds: &terminationGracePeriod,
			Domain: virtv1.DomainSpec{
				Resources: virtv1.ResourceRequirements{
					Requests: v1.ResourceList{
						"cpu":    resource.MustParse(ComputeCPURequests(template.Resources.CPU, template.Resources.ReservedCPUPercentage)),
						"memory": template.Resources.Memory,
					},
					Limits: v1.ResourceList{
						"cpu":    resource.MustParse(ComputeCPULimits(template.Resources.CPU, CPUhypervisorReserved)),
						"memory": template.Resources.Memory,
					},
				},
				CPU: &virtv1.CPU{
					Cores: template.Resources.CPU,
				},
				Memory: &virtv1.Memory{
					Guest: &template.Resources.Memory,
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
			},
			Volumes: []virtv1.Volume{
				{
					Name: "containerdisk",
					VolumeSource: virtv1.VolumeSource{
						ContainerDisk: &virtv1.ContainerDiskSource{
							Image: template.Image,
						},
					},
				},
				{
					Name: "cloudinitdisk",
					VolumeSource: virtv1.VolumeSource{
						CloudInitNoCloud: &virtv1.CloudInitNoCloudSource{
							UserDataSecretRef: &corev1.LocalObjectReference{Name: secretName},
						},
					},
				},
			},
		},
	}
	return &vm, nil
}

func ComputeCPULimits(CPU uint32, HypervisorCoefficient float32) string {
	return fmt.Sprintf("%f", float32(CPU)+HypervisorCoefficient)
}

func ComputeCPURequests(CPU uint32, percentage uint32) string {
	return fmt.Sprintf("%f", float32(CPU*percentage)/100)
}

// create a resource or update it if already exists
func CreateOrUpdate(c client.Client, ctx context.Context, log logr.Logger, object interface{}) error {

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
				log.Error(err, "unable to create secret "+obj.Name)
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
				log.Error(err, "unable to create pvc "+obj.Name)
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
				log.Error(err, "unable to create service "+obj.Name)
				return err
			}
		}
	case networkingv1.Ingress:
		var ing networkingv1.Ingress
		err := c.Get(ctx, types.NamespacedName{
			Namespace: obj.Namespace,
			Name:      obj.Name,
		}, &ing)
		if err != nil {
			err = c.Create(ctx, &obj, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				log.Error(err, "unable to create ingress "+obj.Name)
				return err
			}
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
				log.Error(err, "unable to create deployment "+obj.Name)
				return err
			}
		} else {
			err = c.Update(ctx, &obj, &client.UpdateOptions{})
			if err != nil {
				log.Error(err, "unable to update deployment "+obj.Name)
				return err
			}
		}
	case *virtv1.VirtualMachineInstance:
		err := c.Create(ctx, obj, &client.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "unable to create virtual machine "+obj.Name)
			return err
		}
	default:
		return errors.NewBadRequest("No matching type for requested resource:")
	}

	return nil
}

func CheckLabels(ns v1.Namespace, matchLabels map[string]string) bool {
	for key := range matchLabels {
		if _, ok := ns.Labels[key]; !ok {
			return false
		}
	}
	return true
}
