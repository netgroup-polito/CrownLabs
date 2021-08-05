// Copyright 2020-2021 Politecnico di Torino
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
	"fmt"
	"strings"

	netv1 "k8s.io/api/networking/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// IngressGUINameSuffix -> the suffix added to the name of the ingress targeting the environment GUI.
	IngressGUINameSuffix = "gui"
	// IngressMyDriveNameSuffix -> the suffix added to the name of the ingress targeting the environment "MyDrive".
	IngressMyDriveNameSuffix = "mydrive"

	// IngressDefaultCertificateName -> the name of the secret containing the crownlabs certificate.
	IngressDefaultCertificateName = "crownlabs-ingress-secret"

	// ingressGUIPathSuffix -> the suffix appended to the path of the ingress targeting the environment GUI.
	ingressGUIPathSuffix = "" // Currently empty, to avoid issues with the dashboard ("/gui")
	// ingressMyDrivePathSuffix -> the suffix appended to the path of the ingress targeting the environment "MyDrive".
	ingressMyDrivePathSuffix = "mydrive"
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
						Path:     path + "(/|$)(.*)",
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
func IngressGUIAnnotations(annotations map[string]string, path string) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$2"
	annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = "3600"
	annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = "3600"
	annotations["nginx.ingress.kubernetes.io/configuration-snippet"] =
		`sub_filter '<head>' '<head> <base href="https://$host/` + path + `/index.html">';`

	return annotations
}

// IngressMyDriveAnnotations receives in input a set of annotations and returns the updated set including
// the ones associated with the ingress targeting the environment "MyDrive".
func IngressMyDriveAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = "0"
	annotations["nginx.ingress.kubernetes.io/proxy-max-temp-file-size"] = "0"
	annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = "600"
	annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = "600"

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

// IngressGUIPath returns the path of the ingress targeting the environment GUI.
func IngressGUIPath(instance *clv1alpha2.Instance) string {
	return strings.TrimRight(fmt.Sprintf("/%v/%v", instance.UID, ingressGUIPathSuffix), "/")
}

// IngressMyDrivePath returns the path of the ingress targeting the environment "MyDrive".
func IngressMyDrivePath(instance *clv1alpha2.Instance) string {
	return strings.TrimRight(fmt.Sprintf("/%v/%v", instance.UID, ingressMyDrivePathSuffix), "/")
}
