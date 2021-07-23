package instance_creation

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// ForgeService creates and returns a Kubernetes Service resource providing
// access to a CrownLabs environment.
func ForgeService(name, namespace string) corev1.Service {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "vnc",
					Protocol:   corev1.ProtocolTCP,
					Port:       forge.ServiceNoVNCPort,
					TargetPort: intstr.FromInt(forge.ServiceNoVNCPort),
				},
				{
					Name:       "ssh",
					Protocol:   corev1.ProtocolTCP,
					Port:       forge.ServiceSSHPort,
					TargetPort: intstr.FromInt(forge.ServiceSSHPort),
				},
			},
			Selector:  map[string]string{"name": name},
			ClusterIP: "",
			Type:      corev1.ServiceTypeClusterIP,
		},
	}

	return service
}

// ForgeIngress creates and returns a Kubernetes Ingress resource providing
// exposing the remote desktop of a CrownLabs environment.
func ForgeIngress(name, namespace string, svc *corev1.Service, websiteBaseURL, urlUUID, instancesAuthURL string) networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	url := websiteBaseURL + "/" + urlUUID

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/rewrite-target":        "/$2",
		"nginx.ingress.kubernetes.io/proxy-read-timeout":    "3600",
		"nginx.ingress.kubernetes.io/proxy-send-timeout":    "3600",
		"crownlabs.polito.it/probe-url":                     "https://" + url,
		"crownlabs.polito.it/url-uuid":                      urlUUID,
		"nginx.ingress.kubernetes.io/configuration-snippet": `sub_filter '<head>' '<head> <base href="https://$host/` + urlUUID + `/index.html">';`,
	}
	annotations = appendInstancesAuthAnnotations(annotations, instancesAuthURL)

	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{websiteBaseURL},
					SecretName: "crownlabs-ingress-secret",
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: websiteBaseURL,
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

// ForgeFileBrowserIngress creates and returns a Kubernetes Ingress resource
// exposing FileBrowser for a CrownLabs container environment.
func ForgeFileBrowserIngress(
	name, namespace string, svc *corev1.Service,
	urlUUID, websiteBaseURL, fileBrowserPortName, instancesAuthURL string,
) networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
		"nginx.ingress.kubernetes.io/proxy-read-timeout":       "600",
		"nginx.ingress.kubernetes.io/proxy-send-timeout":       "600",
		"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
	}
	annotations = appendInstancesAuthAnnotations(annotations, instancesAuthURL)

	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name + "-filebrowser",
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{websiteBaseURL},
					SecretName: "crownlabs-ingress-secret",
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: websiteBaseURL,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/" + urlUUID + "/mydrive(/|$)(.*)",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: svc.Name,
											Port: networkingv1.ServiceBackendPort{
												Name: fileBrowserPortName,
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

// appendInstancesAuthAnnotations appends the set of nginx annotations required to enable
// the authentication in front of an ingress resource. instancesAuthURL represents the
// URL of an exposed oauth2-proxy instance properly configured.
func appendInstancesAuthAnnotations(annotations map[string]string, instancesAuthURL string) map[string]string {
	annotations["nginx.ingress.kubernetes.io/auth-url"] = instancesAuthURL + "/auth"
	annotations["nginx.ingress.kubernetes.io/auth-signin"] = instancesAuthURL + "/start?rd=$escaped_request_uri"

	return annotations
}
