package instance_creation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestForgeService(t *testing.T) {
	var (
		name      = "usertest"
		namespace = "namespacetest"
	)

	ownerRef := []metav1.OwnerReference{{
		APIVersion: "crownlabs.polito.it/v1alpha2",
		Kind:       "Instance",
		Name:       "Test1",
	},
	}

	service := ForgeService(name, namespace, ownerRef)

	assert.Equal(t, service.ObjectMeta.Name, name+"-svc")
	assert.Equal(t, service.ObjectMeta.Namespace, namespace)
	assert.Equal(t, service.OwnerReferences[0].APIVersion, "crownlabs.polito.it/v1alpha2")
	assert.Equal(t, service.OwnerReferences[0].Kind, "Instance")
	assert.Equal(t, service.OwnerReferences[0].Name, "Test1")
	assert.Equal(t, service.Spec.Ports[0].Name, "vnc")
	assert.Equal(t, service.Spec.Ports[0].Port, int32(6080))
	assert.Equal(t, service.Spec.Ports[1].Name, "ssh")
	assert.Equal(t, service.Spec.Ports[1].Port, int32(22))
	assert.Equal(t, service.Spec.Ports[0].Name, "vnc")
	assert.Equal(t, service.Spec.Selector["name"], name)
	assert.Equal(t, service.Spec.ClusterIP, "")
	assert.Equal(t, service.Spec.Type, corev1.ServiceTypeClusterIP)
}

func TestForgeIngress(t *testing.T) {
	var (
		name           = "usertest"
		namespace      = "namespacetest"
		urlUUID        = "urlUUIDtest"
		websiteBaseURL = "websiteBaseUrlTest"
		svc            = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "svc-test",
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						TargetPort: intstr.IntOrString{IntVal: 22},
					},
				},
			},
		}
		url = websiteBaseURL + "/" + urlUUID
	)

	ownerRef := []metav1.OwnerReference{{
		APIVersion: "crownlabs.polito.it/v1alpha2",
		Kind:       "Instance",
		Name:       "Test1",
	},
	}

	ingress := ForgeIngress(name, namespace, &svc, urlUUID, websiteBaseURL, ownerRef)

	assert.Equal(t, ingress.ObjectMeta.Name, name+"-ingress")
	assert.Equal(t, ingress.ObjectMeta.Namespace, namespace)
	assert.Equal(t, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Name, svc.Name)
	assert.Equal(t, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Port.Number, svc.Spec.Ports[0].TargetPort.IntVal)
	assert.Equal(t, ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/auth-signin"], "https://$host/"+urlUUID+"/oauth2/start?rd=$escaped_request_uri")
	assert.Equal(t, ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/auth-url"], "https://$host/"+urlUUID+"/oauth2/auth")
	assert.Equal(t, ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"], `sub_filter '<head>' '<head> <base href="https://$host/`+urlUUID+`/index.html">';`)
	assert.Equal(t, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path, "/"+urlUUID+"(/|$)(.*)")
	assert.Equal(t, ingress.ObjectMeta.Annotations["crownlabs.polito.it/probe-url"], "https://"+url)
	assert.Equal(t, ingress.Spec.TLS[0].Hosts[0], websiteBaseURL)
	assert.Equal(t, ingress.Spec.Rules[0].Host, websiteBaseURL)
	assert.Equal(t, ingress.OwnerReferences[0].APIVersion, "crownlabs.polito.it/v1alpha2")
	assert.Equal(t, ingress.OwnerReferences[0].Kind, "Instance")
	assert.Equal(t, ingress.OwnerReferences[0].Name, "Test1")
}

func TestForgeOauth2Deployment(t *testing.T) {
	var (
		name         = "usertest"
		namespace    = "namespacetest"
		urlUUID      = "urlUUIDtest"
		image        = "imagetest"
		clientSecret = "secrettest"
		providerURL  = "urltest"
	)
	ownerRef := []metav1.OwnerReference{{
		APIVersion: "crownlabs.polito.it/v1alpha2",
		Kind:       "Instance",
		Name:       "Test1",
	},
	}

	deploy := ForgeOauth2Deployment(name, namespace, urlUUID, image, clientSecret, providerURL, ownerRef)

	assert.Equal(t, deploy.ObjectMeta.Name, name+"-oauth2-deploy")
	assert.Equal(t, deploy.ObjectMeta.Namespace, namespace)
	assert.Equal(t, deploy.OwnerReferences[0].APIVersion, "crownlabs.polito.it/v1alpha2")
	assert.Equal(t, deploy.OwnerReferences[0].Kind, "Instance")
	assert.Equal(t, deploy.OwnerReferences[0].Name, "Test1")
	assert.Equal(t, deploy.Spec.Template.Spec.Containers[0].Image, image)
	assert.Contains(t, deploy.Spec.Template.Spec.Containers[0].Args, "--proxy-prefix=/"+urlUUID+"/oauth2")
	assert.Contains(t, deploy.Spec.Template.Spec.Containers[0].Args, "--cookie-path=/"+urlUUID)
	assert.Contains(t, deploy.Spec.Template.Spec.Containers[0].Args, "--client-secret="+clientSecret)
	assert.Contains(t, deploy.Spec.Template.Spec.Containers[0].Args, "--login-url="+providerURL+"/protocol/openid-connect/auth")
	assert.Contains(t, deploy.Spec.Template.Spec.Containers[0].Args, "--redeem-url="+providerURL+"/protocol/openid-connect/token")
	assert.Contains(t, deploy.Spec.Template.Spec.Containers[0].Args, "--validate-url="+providerURL+"/protocol/openid-connect/userinfo")
}

func TestForgeOauth2Service(t *testing.T) {
	var (
		name      = "usertest"
		namespace = "namespacetest"
	)
	ownerRef := []metav1.OwnerReference{{
		APIVersion: "crownlabs.polito.it/v1alpha2",
		Kind:       "Instance",
		Name:       "Test1",
	},
	}
	service := ForgeOauth2Service(name, namespace, ownerRef)

	assert.Equal(t, service.ObjectMeta.Name, name+"-oauth2-svc")
	assert.Equal(t, service.ObjectMeta.Namespace, namespace)
	assert.Equal(t, service.OwnerReferences[0].APIVersion, "crownlabs.polito.it/v1alpha2")
	assert.Equal(t, service.OwnerReferences[0].Kind, "Instance")
	assert.Equal(t, service.OwnerReferences[0].Name, "Test1")
	assert.Equal(t, service.Spec.Selector, generateOauth2Labels(name))
}

func TestForgeOauth2Ingress(t *testing.T) {
	var (
		name           = "usertest"
		namespace      = "namespacetest"
		urlUUID        = "urlUUIDtest"
		websiteBaseURL = "websiteBaseUrlTest"
		svc            = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "svc-test",
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						TargetPort: intstr.IntOrString{IntVal: 22},
					},
				},
			},
		}
	)
	ownerRef := []metav1.OwnerReference{{
		APIVersion: "crownlabs.polito.it/v1alpha2",
		Kind:       "Instance",
		Name:       "Test1",
	},
	}

	ingress := ForgeOauth2Ingress(name, namespace, &svc, urlUUID, websiteBaseURL, ownerRef)

	assert.Equal(t, ingress.ObjectMeta.Name, name+"-oauth2-ingress")
	assert.Equal(t, ingress.ObjectMeta.Namespace, namespace)
	assert.Equal(t, ingress.OwnerReferences[0].APIVersion, "crownlabs.polito.it/v1alpha2")
	assert.Equal(t, ingress.OwnerReferences[0].Kind, "Instance")
	assert.Equal(t, ingress.OwnerReferences[0].Name, "Test1")
	assert.Equal(t, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Name, svc.Name)
	assert.Equal(t, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Port.Number, svc.Spec.Ports[0].TargetPort.IntVal)
	assert.Equal(t, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path, "/"+urlUUID+"/oauth2/.*")
	assert.Equal(t, ingress.Spec.TLS[0].Hosts[0], websiteBaseURL)
	assert.Equal(t, ingress.Spec.Rules[0].Host, websiteBaseURL)
}

func TestGenerateOauth2Labels(t *testing.T) {
	instanceName := "oauth2-foo"
	labels := generateOauth2Labels(instanceName)

	assert.Contains(t, labels, "app.kubernetes.io/part-of")
	assert.Contains(t, labels, "app.kubernetes.io/component")
	assert.Equal(t, labels["app.kubernetes.io/part-of"], instanceName)
	assert.Equal(t, labels["app.kubernetes.io/component"], "oauth2-proxy")
}
