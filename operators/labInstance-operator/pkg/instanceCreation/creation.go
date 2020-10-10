package instanceCreation

import (
	"context"
	"encoding/base64"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	virtv1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/kubeVirt/api/v1"
	templatev1alpha1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/labTemplate/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateVirtualMachineInstance(name string, namespace string, template templatev1alpha1.LabTemplate, instanceName string, secretName string) virtv1.VirtualMachineInstance {
	vm := template.Spec.Vm
	vm.Name = name + "-vmi"
	vm.Namespace = namespace
	vm.Labels = map[string]string{"name": name, "template-name": template.Name, "instance-name": instanceName}

	for _, volume := range vm.Spec.Volumes {
		if volume.Name == "cloudinitdisk" {
			volume.CloudInitNoCloud.UserDataSecretRef = &corev1.LocalObjectReference{Name: secretName}
		}
	}
	return vm
}

func CreateSecret(name string, namespace string, nextUsername string, nextPassword string, nextCloudBaseUrl string) corev1.Secret {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-secret",
			Namespace: namespace,
		},
		Data: nil,
		StringData: map[string]string{"userdata": `
#cloud-config
network:
  version: 2
  id0:
    dhcp4: true
mounts:
  - [ "` + nextCloudBaseUrl + `/remote.php/dav/files/` + nextUsername + `", "/media/MyDrive", "davfs", "_netdev,auto,user,rw,uid=1000,gid=1000","0","0" ]
write_files:
  - content: |
      /media/MyDrive ` + nextUsername + " " + nextPassword + `
    path: /etc/davfs2/secrets
    permissions: '0600'
`},
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
					TargetPort: intstr.IntOrString{IntVal: 22},
				},
			},
			Selector:  map[string]string{"name": name},
			ClusterIP: "",
			Type:      corev1.ServiceTypeClusterIP,
		},
	}

	return service
}

func CreatePersistentVolumeClaim(name string, namespace string, storageClassName string) corev1.PersistentVolumeClaim {
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

func CreateIngress(name string, namespace string, svc corev1.Service, urlUUID string, websiteBaseUrl string) v1beta1.Ingress {
	url := websiteBaseUrl + "/" + urlUUID

	ingress := v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-ingress",
			Namespace: namespace,
			Labels:    nil,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target":        "/$2",
				"nginx.ingress.kubernetes.io/proxy-read-timeout":    "3600",
				"nginx.ingress.kubernetes.io/proxy-send-timeout":    "3600",
				"nginx.ingress.kubernetes.io/auth-signin":           "https://$host/" + urlUUID + "/oauth2/start?rd=$escaped_request_uri",
				"nginx.ingress.kubernetes.io/auth-url":              "https://$host/" + urlUUID + "/oauth2/auth",
				"crownlabs.polito.it/probe-url":                     "https://" + url,
				"nginx.ingress.kubernetes.io/configuration-snippet": `sub_filter '<head>' '<head> <base href="https://$host/` + urlUUID + `/index.html">';`,
			},
		},
		Spec: v1beta1.IngressSpec{
			Backend: nil,
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      []string{websiteBaseUrl},
					SecretName: "crownlabs-labinstances-secret",
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: websiteBaseUrl,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/" + urlUUID + "(/|$)(.*)",
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

func CreateOauth2Deployment(name, namespace, urlUUID, image, clientSecret, providerUrl string) appsv1.Deployment {

	cookieUUID := uuid.New().String()
	id, _ := uuid.New().MarshalBinary()
	cookieSecret := base64.StdEncoding.EncodeToString(id)

	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-oauth2-deploy",
			Namespace: namespace,
			Labels:    map[string]string{"app": name},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Args: []string{
								"--http-address=0.0.0.0:4180",
								"--reverse-proxy=true",
								"--skip-provider-button=true",
								"--cookie-secret=" + cookieSecret,
								"--cookie-expire=24h",
								"--cookie-name=_oauth2_cookie_" + string([]rune(cookieUUID)[:6]),
								"--provider=keycloak",
								"--client-id=k8s",
								"--client-secret=" + clientSecret,
								"--login-url=" + providerUrl + "/protocol/openid-connect/auth",
								"--redeem-url=" + providerUrl + "/protocol/openid-connect/token",
								"--validate-url=" + providerUrl + "/protocol/openid-connect/userinfo",
								"--proxy-prefix=/" + urlUUID + "/oauth2",
								"--cookie-path=/" + urlUUID,
								"--email-domain=*",
								"--session-cookie-minimal=true",
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 4180,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("100Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("25Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	return deploy
}

func CreateOauth2Service(name string, namespace string) corev1.Service {

	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-oauth2-svc",
			Namespace: namespace,
			Labels:    map[string]string{"app": name},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       4180,
					TargetPort: intstr.IntOrString{IntVal: 4180},
				},
			},
			Selector: map[string]string{"app": name},
		},
	}

	return service
}

func CreateOauth2Ingress(name string, namespace string, svc corev1.Service, urlUUID string, websiteBaseUrl string) v1beta1.Ingress {

	ingress := v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-oauth2-ingress",
			Namespace: namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/cors-allow-credentials": "true",
				"nginx.ingress.kubernetes.io/cors-allow-headers":     "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization",
				"nginx.ingress.kubernetes.io/cors-allow-methods":     "PUT, GET, POST, OPTIONS, DELETE, PATCH",
				"nginx.ingress.kubernetes.io/cors-allow-origin":      "https://*",
				"nginx.ingress.kubernetes.io/enable-cors":            "true",
			},
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      []string{websiteBaseUrl},
					SecretName: "crownlabs-labinstances-secret",
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: websiteBaseUrl,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/" + urlUUID + "/oauth2/.*",
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

func GetWebdavCredentials(c client.Client, ctx context.Context, log logr.Logger, secretName string, namespace string, username *string, password *string) error {
	sec := corev1.Secret{}
	nsdName := types.NamespacedName{
		Namespace: namespace,
		Name:      secretName,
	}
	if err := c.Get(ctx, nsdName, &sec); err == nil {
		var ok bool
		var userBytes, passBytes []byte
		if userBytes, ok = sec.Data["username"]; !ok {
			log.Error(nil, "Unable to find username in webdav secret "+secretName)
		} else {
			*username = string(userBytes)
		}
		if passBytes, ok = sec.Data["password"]; !ok {
			log.Error(nil, "Unable to find password in webdav secret"+secretName)
		} else {
			*password = string(passBytes)
		}
		return nil
	} else {
		return err
	}
}

func CheckLabels(ns v1.Namespace, matchLabels map[string]string) bool {
	for key := range matchLabels {
		if _, ok := ns.Labels[key]; !ok {
			return false
		}
	}
	return true
}
