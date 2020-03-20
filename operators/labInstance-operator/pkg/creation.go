package pkg

import (
	"context"
	"github.com/go-logr/logr"
	virtv1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/kubeVirt/api/v1"
	templatev1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/labTemplate/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateVirtualMachineInstance(name string, namespace string, template templatev1.LabTemplate, secretName string, pvcName string) virtv1.VirtualMachineInstance {
	vm := template.Spec.Vm
	vm.Name = name + "-vmi"
	vm.Namespace = namespace
	vm.Labels = map[string]string{"name": name}

	for _, volume := range vm.Spec.Volumes {
		if volume.Name == "cloudinitdisk" {
			volume.CloudInitNoCloud.UserDataSecretRef = &corev1.LocalObjectReference{Name: secretName}
		}
		if volume.Name == "pvcdisk" {
			volume.PersistentVolumeClaim.ClaimName = pvcName
		}
	}
	return vm
}

func CreateSecret(name string, namespace string) corev1.Secret {

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-secret",
			Namespace: namespace,
		},
		Data: nil,
		StringData: map[string]string{"userdata": `
			network:
			  version: 2
			  id0:
			    dhcp4: true`},
		Type: corev1.SecretTypeOpaque,
	}

	return secret
}

func CreateService(name string, namespace string) corev1.Service {

	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-svc",
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "vnc",
					Protocol:   corev1.ProtocolTCP,
					Port:       6080,
					TargetPort: intstr.IntOrString{IntVal: 6080},
				},
				{
					Name:       "ssh",
					Protocol:   corev1.ProtocolTCP,
					Port:       22,
					TargetPort: intstr.IntOrString{IntVal: 6081},
				},
			},
			Selector:  map[string]string{"name": name},
			ClusterIP: "",
			Type:      corev1.ServiceTypeClusterIP,
		},
	}

	return service
}

func CreatePerstistentVolumeClaim(name string, namespace string, storageClassName string) corev1.PersistentVolumeClaim {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-pvc",
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Limits: nil,
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse("1Gi")},
			},
			StorageClassName: &storageClassName,
		},
	}

	return pvc
}

func CreateIngress(name string, namespace string, secretName string, svc corev1.Service) v1beta1.Ingress {

	ingress := v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name + "-ingress",
			Namespace:   namespace,
			Labels:      nil,
			Annotations: map[string]string{"cert-manager.io/cluster-issuer": "letsencrypt-production", "nginx.ingress.kubernetes.io/rewrite-target": "/$2"},
		},
		Spec: v1beta1.IngressSpec{
			Backend: nil,
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      []string{"crownlabs.polito.it"},
					SecretName: secretName,
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: "crownlabs.polito.it",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/" + name + "(/|$)(.*)",
									Backend: v1beta1.IngressBackend{
										ServiceName: svc.Name,
										ServicePort: svc.Spec.Ports[0].TargetPort,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
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
	case v1beta1.Ingress:
		var ing v1beta1.Ingress
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
	case virtv1.VirtualMachineInstance:
		var vmi virtv1.VirtualMachineInstance
		err := c.Get(ctx, types.NamespacedName{
			Namespace: obj.Namespace,
			Name:      obj.Name,
		}, &vmi)
		if err != nil {
			err = c.Create(ctx, &obj, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				log.Error(err, "unable to create virtual machine "+obj.Name)
				return err
			}
		}
	}

	return nil
}
