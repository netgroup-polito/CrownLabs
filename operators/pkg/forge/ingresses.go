// Copyright 2020-2025 Politecnico di Torino
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
	"k8s.io/utils/ptr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// IngressInstancePrefix -> the prefix prepended to the path of any ingresses targeting the instance or its subresources.
	IngressInstancePrefix = "/instance"

	// IngressGUINameSuffix -> the suffix added to the name of the ingress targeting the environment GUI.
	IngressGUINameSuffix = "gui"

	// IngressAppSuffix -> the suffix added to the path of the ingress targeting standalone and container environments.
	IngressAppSuffix = "app"

	// IngressDefaultCertificateName -> the name of the secret containing the crownlabs certificate.
	IngressDefaultCertificateName = "crownlabs-ingress-secret"

	// IngressVNCGUIPathSuffix -> the suffix appended to the path of the ingress targeting the environment GUI websocketed vnc endpoint.
	IngressVNCGUIPathSuffix = "vnc"

	// IngressdashboardPathSuffix -> the suffix appended to the path of the ingress targeting the environment dashboard endpoint.
	IngressdashboardPathSuffix = "dashboard"

	// WebsockifyRewriteEndpoint -> endpoint of the websocketed vnc server.
	WebsockifyRewriteEndpoint = "/websockify"
	// StandaloneRewriteEndpoint -> endpoint of the standalone application.
	StandaloneRewriteEndpoint = "/$2"
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

// IngressSpec forges the specification of a Kubernetes Ingress resource only for cluster
func IngressClusterSpec(host, path, certificateName, serviceName, servicePort string) netv1.IngressSpec {
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
		IngressClassName: ptr.To("nginx-ssl"),
	}
}

// IngressGUIAnnotations receives in input a set of annotations and returns the updated set including
// the ones associated with the ingress targeting the environment GUI.
func IngressGUIAnnotations(environment *clv1alpha2.Environment, annotations map[string]string) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}
	if environment.EnvironmentType == clv1alpha2.ClassStandalone && environment.RewriteURL {
		annotations["nginx.ingress.kubernetes.io/rewrite-target"] = StandaloneRewriteEndpoint
	}
	annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = "3600"
	annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = "3600"
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

// HostName returns the hostname based on the given EnvironmentMode.
func HostName(baseHostName string, mode clv1alpha2.EnvironmentMode) string {
	switch mode {
	case clv1alpha2.ModeStandard:
		return baseHostName
	case clv1alpha2.ModeExam:
		return "exam." + baseHostName
	case clv1alpha2.ModeExercise:
		return "exercise." + baseHostName
	}

	return baseHostName
}

// IngressGUIPath returns the path of the ingress targeting the environment GUI vnc or Standalone.
func IngressGUIPath(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) string {
	switch environment.EnvironmentType {
	case clv1alpha2.ClassStandalone:
		if environment.RewriteURL {
			return strings.TrimRight(fmt.Sprintf("%v/%v/%v", IngressInstancePrefix, instance.UID, IngressAppSuffix+"(/|$)(.*)"), "/")
		}
		return strings.TrimRight(fmt.Sprintf("%v/%v/%v", IngressInstancePrefix, instance.UID, IngressAppSuffix), "/")
	case clv1alpha2.ClassContainer:
		return strings.TrimRight(fmt.Sprintf("%v/%v/%v", IngressInstancePrefix, instance.UID, IngressAppSuffix), "/")
	case clv1alpha2.ClassCloudVM, clv1alpha2.ClassVM:
		return strings.TrimRight(fmt.Sprintf("%v/%v/%v", IngressInstancePrefix, instance.UID, IngressVNCGUIPathSuffix), "/")
	case clv1alpha2.ClassCluster:
		return strings.TrimRight(fmt.Sprintf("%v/%v/%v", IngressInstancePrefix, instance.UID, IngressdashboardPathSuffix), "/")
	}
	return ""
}

// IngressGUICleanPath returns the path of the ingress targeting the environment GUI vnc or Standalone, without the url-rewrite's regex.
func IngressGUICleanPath(instance *clv1alpha2.Instance) string {
	return strings.TrimRight(fmt.Sprintf("%v/%v/%v", IngressInstancePrefix, instance.UID, IngressAppSuffix), "/")
}

// IngressGuiStatusURL returns the path of the ingress targeting the environment.
func IngressGuiStatusURL(host string, environment *clv1alpha2.Environment, instance *clv1alpha2.Instance) string {
	switch environment.EnvironmentType {
	case clv1alpha2.ClassStandalone, clv1alpha2.ClassContainer:
		return fmt.Sprintf("https://%v%v/%v/%v/", host, IngressInstancePrefix, instance.UID, IngressAppSuffix)
	case clv1alpha2.ClassVM, clv1alpha2.ClassCloudVM:
		return fmt.Sprintf("https://%v%v/%v/", host, IngressInstancePrefix, instance.UID)
	case clv1alpha2.ClassCluster:
		return fmt.Sprintf("https://%v:%v/", host, environment.Cluster.ClusterNet.NginxTargetPort)
	}
	return ""
}

// IngressGUIName returns the name of the ingress resource.
func IngressGUIName(environment *clv1alpha2.Environment) string {
	switch environment.EnvironmentType {
	case clv1alpha2.ClassStandalone:
		return IngressAppSuffix
	case clv1alpha2.ClassContainer, clv1alpha2.ClassVM, clv1alpha2.ClassCloudVM:
		return IngressGUINameSuffix
	case clv1alpha2.ClassCluster:
		return IngressdashboardPathSuffix
	}
	return ""
}
