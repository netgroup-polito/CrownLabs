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

	service := ForgeService(name, namespace)

	assert.Equal(t, service.ObjectMeta.Name, name)
	assert.Equal(t, service.ObjectMeta.Namespace, namespace)
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
		name             = "usertest"
		namespace        = "namespacetest"
		urlUUID          = "urlUUIDtest"
		websiteBaseURL   = "websiteBaseUrlTest"
		instancesAuthURL = "fake.com/auth"
		svc              = corev1.Service{
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

	instancesAuthAnnotations := appendInstancesAuthAnnotations(map[string]string{}, instancesAuthURL)
	ingress := ForgeIngress(name, namespace, &svc, websiteBaseURL, urlUUID, instancesAuthURL)

	assert.Equal(t, ingress.ObjectMeta.Name, name)
	assert.Equal(t, ingress.ObjectMeta.Namespace, namespace)
	assert.Equal(t, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Name, svc.Name)
	assert.Equal(t, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Port.Number, svc.Spec.Ports[0].TargetPort.IntVal)
	assert.Equal(t, ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"], `sub_filter '<head>' '<head> <base href="https://$host/`+urlUUID+`/index.html">';`)
	assert.Equal(t, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path, "/"+urlUUID+"(/|$)(.*)")
	assert.Equal(t, ingress.ObjectMeta.Annotations["crownlabs.polito.it/probe-url"], "https://"+url)
	assert.Equal(t, ingress.Spec.TLS[0].Hosts[0], websiteBaseURL)
	assert.Equal(t, ingress.Spec.Rules[0].Host, websiteBaseURL)

	for key, value := range instancesAuthAnnotations {
		assert.Contains(t, ingress.GetAnnotations(), key)
		assert.Equal(t, ingress.GetAnnotations()[key], value)
	}
}

func TestAppendInstancesAuthAnnotations(t *testing.T) {
	const (
		instancesAuthURL = "fake.com/auth"
		originalKey      = "originalKey"
		originalValue    = "originalValue"
	)

	originalAnnotations := map[string]string{
		originalKey: originalValue,
	}

	resultingAnnotations := appendInstancesAuthAnnotations(originalAnnotations, instancesAuthURL)

	// The original annotations are unmodified
	assert.Contains(t, resultingAnnotations, originalKey)
	assert.Equal(t, resultingAnnotations[originalKey], originalValue)

	// The new annotations are added correctly
	assert.Contains(t, resultingAnnotations, "nginx.ingress.kubernetes.io/auth-url")
	assert.Contains(t, resultingAnnotations, "nginx.ingress.kubernetes.io/auth-signin")

	assert.Equal(t, resultingAnnotations["nginx.ingress.kubernetes.io/auth-url"], instancesAuthURL+"/auth")
	assert.Equal(t, resultingAnnotations["nginx.ingress.kubernetes.io/auth-signin"], instancesAuthURL+"/start?rd=$escaped_request_uri")
}
