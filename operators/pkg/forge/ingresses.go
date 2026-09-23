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
	netv1 "k8s.io/api/networking/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// IngressInstancePrefix -> the prefix prepended to the path of any ingresses targeting the instance or its subresources.
	// IngressInstancePrefix = "/instance".

	// IngressDefaultCertificateName -> the name of the secret containing the crownlabs certificate.
	IngressDefaultCertificateName = "crownlabs-ingress-secret"

	// StandaloneRewriteEndpointIngress -> endpoint of the standalone application.
	StandaloneRewriteEndpointIngress = "/$2"
	// GUIRewriteEndpointIngress -> used to clean the path of the ingress targeting the environment GUI.
	GUIRewriteEndpointIngress = "/$1"
)

// IngressSpec forges the specification of a Kubernetes Ingress resource.
func IngressSpec(host, path, certificateName, serviceName, servicePort string) netv1.IngressSpec {
	pathTypePrefix := netv1.PathTypePrefix
	return netv1.IngressSpec{
		TLS: []netv1.IngressTLS{{Hosts: []string{host}, SecretName: certificateName}},
		Rules: []netv1.IngressRule{{
			Host: host,
			IngressRuleValue: netv1.IngressRuleValue{
				HTTP: &netv1.HTTPIngressRuleValue{
					Paths: []netv1.HTTPIngressPath{{
						Path:     path,
						PathType: &pathTypePrefix,
						Backend: netv1.IngressBackend{
							Service: &netv1.IngressServiceBackend{
								Name: serviceName,
								Port: netv1.ServiceBackendPort{Name: servicePort},
							},
						},
					}},
				},
			},
		}},
	}
}

// IngressGUIAnnotations receives in input a set of annotations and returns the updated set including
// the ones associated with the ingress targeting the environment GUI.
func IngressGUIAnnotations(environment *clv1alpha2.Environment, annotations map[string]string) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}

	if environment.EnvironmentType == clv1alpha2.ClassStandalone && environment.RewriteURL {
		annotations["nginx.ingress.kubernetes.io/rewrite-target"] = StandaloneRewriteEndpointIngress
	}

	if environment.EnvironmentType == clv1alpha2.ClassCloudVM || environment.EnvironmentType == clv1alpha2.ClassVM || environment.EnvironmentType == clv1alpha2.ClassLocalVM {
		annotations["nginx.ingress.kubernetes.io/rewrite-target"] = GUIRewriteEndpointIngress
	}

	annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = "3600"
	annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = "3600"

	return annotations
}

// IngressAuthenticationAnnotations receives in input a set of annotations and returns the updated set including
// the ones required to enable the authentication in front of an ingress resource. instancesAuthURL represents the
// URL of an exposed oauth2-proxy instance properly configured.
func IngressAuthenticationAnnotations(annotations map[string]string, instancesAuthURL string) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations["nginx.ingress.kubernetes.io/auth-url"] = instancesAuthURL + "/auth"
	annotations["nginx.ingress.kubernetes.io/auth-signin"] = instancesAuthURL + "/start?rd=$escaped_request_uri"

	return annotations
}
