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
	"fmt"
	"regexp"
	"strconv"
	"strings"

	egv1alpha1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// AuthServiceAnnotation is the annotation used to specify the authentication service for an instance.
	AuthServiceAnnotation = "crownlabs.polito.it/auth-service"
	// AuthAnnotationDisabled indicates that authentication should be disabled.
	AuthAnnotationDisabled = "none"
	// AuthAnnotationDefault indicates that default authentication should be used.
	AuthAnnotationDefault = "default"
)

// AuthServiceInfo contains the details of a parsed authentication service.
type AuthServiceInfo struct {
	ServiceName string
	Namespace   string
	Port        int32
	Path        string
	Mode        string
}

// ParseAuthServiceAnnotation parses the crownlabs.polito.it/auth-service annotation.
// Expected formats:
// - "" or "default" -> defaults (parses the defaultAuthString)
// - "none" or "disabled" -> authentication disabled
// - "serviceName.namespace:port/path" -> custom service (namespace, port, path are optional)
// Returns AuthServiceInfo and error.
func ParseAuthServiceAnnotation(raw, tenantNamespace, defaultAuthString string) (*AuthServiceInfo, error) {
	raw = strings.TrimSpace(raw)

	if raw == "" || raw == AuthAnnotationDefault {
		info, err := ParseAuthServiceAnnotation(defaultAuthString, tenantNamespace, "")
		if err != nil {
			return nil, err
		}
		info.Mode = "default"
		return info, nil
	}

	if raw == AuthAnnotationDisabled || raw == "disabled" {
		return &AuthServiceInfo{Mode: "none"}, nil
	}

	// Regex to parse [serviceName][.namespace][:port][/path]
	// Example: my-service.my-ns:8080/my-path
	// Group 1: serviceName
	// Group 3: namespace (optional)
	// Group 5: port (optional)
	// Group 6: /path (optional)
	re := regexp.MustCompile(`^([a-zA-Z0-9-]+)(\.([a-zA-Z0-9-]+))?(:(\d+))?(/.*)?$`)
	matches := re.FindStringSubmatch(raw)

	if matches == nil {
		return nil, fmt.Errorf("invalid auth-service annotation format: %q", raw)
	}

	serviceName := matches[1]
	namespace := matches[3]
	if namespace == "" {
		namespace = tenantNamespace
	}

	var port int32
	portStr := matches[5]
	if portStr != "" {
		p, e := strconv.ParseUint(portStr, 10, 16)
		if e != nil {
			return nil, fmt.Errorf("invalid port in auth-service annotation: %q", portStr)
		}
		port = int32(p)
	} else {
		port = 80 // Default port for custom service
	}

	path := matches[6]
	if path == "" {
		path = "/check" // Default path for custom service if not specified
	}

	return &AuthServiceInfo{
		ServiceName: serviceName,
		Namespace:   namespace,
		Port:        port,
		Path:        path,
		Mode:        "custom",
	}, nil
}

// SecurityPolicy forges the SecurityPolicy required to expose the HTTPRoute to an external auth service.
func SecurityPolicy(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, serviceName, namespace string, port int32, path string) egv1alpha1.SecurityPolicySpec {
	// targetRouteName is the HTTPRoute that exposes the environment.
	targetRouteName := fmt.Sprintf("%v-%v", instance.Name, environment.Name)

	sp := egv1alpha1.SecurityPolicySpec{
		PolicyTargetReferences: egv1alpha1.PolicyTargetReferences{
			TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
				{
					LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
						Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
						Kind:  gwapiv1a2.Kind("HTTPRoute"),
						Name:  gwapiv1a2.ObjectName(targetRouteName),
					},
				},
			},
		},
		ExtAuth: &egv1alpha1.ExtAuth{
			HeadersToExtAuth: []string{
				"Cookie",
				"Authorization",
			},
			HTTP: &egv1alpha1.HTTPExtAuthService{
				BackendRefs: []egv1alpha1.BackendRef{
					{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Group:     ptr.To(gatewayv1.Group("")),
							Kind:      ptr.To(gatewayv1.Kind("Service")),
							Name:      gatewayv1.ObjectName(serviceName),
							Namespace: ptr.To(gatewayv1.Namespace(namespace)),
							Port:      ptr.To(gatewayv1.PortNumber(port)),
						},
					},
				},
				Path: ptr.To(path),
			},
		},
	}

	return sp
}
