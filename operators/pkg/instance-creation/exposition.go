package instance_creation

import (
	"encoding/base64"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func CreateService(name string, namespace string, references []metav1.OwnerReference) corev1.Service {

	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name + "-svc",
			Namespace:       namespace,
			OwnerReferences: references,
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

func CreateIngress(name string, namespace string, svc corev1.Service, urlUUID string, websiteBaseUrl string, references []metav1.OwnerReference) networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	url := websiteBaseUrl + "/" + urlUUID

	ingress := networkingv1.Ingress{
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
			OwnerReferences: references,
		},
		Spec: networkingv1.IngressSpec{
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{websiteBaseUrl},
					SecretName: "crownlabs-labinstances-secret",
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: websiteBaseUrl,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/" + urlUUID + "(/|$)(.*)",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: svc.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: svc.Spec.Ports[0].TargetPort.IntVal,
											},
										},
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

func CreateOauth2Deployment(name, namespace, urlUUID, image, clientSecret, providerUrl string, references []metav1.OwnerReference) appsv1.Deployment {

	cookieUUID := uuid.New().String()
	id, _ := uuid.New().MarshalBinary()
	cookieSecret := base64.StdEncoding.EncodeToString(id)

	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name + "-oauth2-deploy",
			Namespace:       namespace,
			OwnerReferences: references,
			Labels:          map[string]string{"app": name},
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

func CreateOauth2Service(name string, namespace string, references []metav1.OwnerReference) corev1.Service {

	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name + "-oauth2-svc",
			Namespace:       namespace,
			OwnerReferences: references,
			Labels:          map[string]string{"app": name},
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

func CreateOauth2Ingress(name string, namespace string, svc corev1.Service, urlUUID string, websiteBaseUrl string, references []metav1.OwnerReference) networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name + "-oauth2-ingress",
			Namespace:       namespace,
			OwnerReferences: references,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/cors-allow-credentials": "true",
				"nginx.ingress.kubernetes.io/cors-allow-headers":     "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization",
				"nginx.ingress.kubernetes.io/cors-allow-methods":     "PUT, GET, POST, OPTIONS, DELETE, PATCH",
				"nginx.ingress.kubernetes.io/cors-allow-origin":      "https://*",
				"nginx.ingress.kubernetes.io/enable-cors":            "true",
			},
		},
		Spec: networkingv1.IngressSpec{
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{websiteBaseUrl},
					SecretName: "crownlabs-labinstances-secret",
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: websiteBaseUrl,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/" + urlUUID + "/oauth2/.*",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: svc.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: svc.Spec.Ports[0].TargetPort.IntVal,
											},
										},
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
