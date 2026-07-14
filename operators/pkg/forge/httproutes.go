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
	"strings"

	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// ExpositionInstancePrefix -> the prefix prepended to the path of any HTTPRoute
	// targeting the instance or its subresources.
	ExpositionInstancePrefix = "/instance"

	// DefaultTimeoutSeconds -> the default timeout as a Gateway API Duration string (GEP-2257 format).
	DefaultTimeoutSeconds = "3600s"

	// StandaloneRewriteEndpoint -> the endpoint used to rewrite standalone GUI URLs.
	StandaloneRewriteEndpoint = ""

	// GUIRewriteEndpoint -> the endpoint used to rewrite CloudVM/VM GUI URLs.
	GUIRewriteEndpoint = ""
)

// HTTPRouteTemplate groups the minimal parameters required to forge an HTTPRouteSpec.
type HTTPRouteTemplate struct {
	Path        string
	ServiceName string
}

// ExpositionConfig holds gateway information used by HTTPRouteSpec.
type ExpositionConfig struct {
	WebsiteBaseURL       string
	InstancesAuthURL     string
	EnableAuthentication bool
	GatewayAPIMode       bool
	GatewayName          string
	GatewayNamespace     string
}

// ParseGatewayParent parses a gateway parent reference of the form
// "namespace/name" and returns the two components. Returns an error on invalid input.
func ParseGatewayParent(raw string) (namespace, name string, err error) {
	raw = strings.TrimSpace(raw)
	parts := strings.Split(raw, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid gateway parent reference: %q", raw)
	}
	trim := strings.TrimSpace
	namespace, name = trim(parts[0]), trim(parts[1])
	if namespace == "" || name == "" {
		return "", "", fmt.Errorf("invalid gateway parent reference, empty namespace or name: %q", raw)
	}
	return namespace, name, nil
}

// HTTPRouteSpec forges the specification of a Kubernetes HTTPRoute resource.
func HTTPRouteSpec(tpl *HTTPRouteTemplate, expo *ExpositionConfig, environment *clv1alpha2.Environment, servicePort int32) gatewayv1.HTTPRouteSpec {
	parentRef := ParentReference(expo.GatewayName, expo.GatewayNamespace)
	rule := HTTPRouteRule(tpl.Path, tpl.ServiceName, servicePort, environment)

	spec := gatewayv1.HTTPRouteSpec{
		CommonRouteSpec: gatewayv1.CommonRouteSpec{
			ParentRefs: []gatewayv1.ParentReference{parentRef},
		},
		Rules: []gatewayv1.HTTPRouteRule{rule},
	}

	return spec
}

// ParentReference creates the appropriate ParentReference for the provided gateway.
func ParentReference(gatewayName, gatewayNamespace string) gatewayv1.ParentReference {
	parent := gatewayv1.ParentReference{}

	parent.Name = gatewayv1.ObjectName(gatewayName)
	parent.Namespace = ptr.To(gatewayv1.Namespace(gatewayNamespace))

	return parent
}

// HTTPRouteRule constructs a complete HTTPRouteRule including match, backend reference and timeout filters.
func HTTPRouteRule(path, serviceName string, servicePort int32, environment *clv1alpha2.Environment) gatewayv1.HTTPRouteRule {
	match := HTTPRouteMatch(path)
	backendRef := HTTPBackendRef(serviceName, servicePort)
	timeout := HTTPRouteTimeouts()

	rule := gatewayv1.HTTPRouteRule{
		Matches:     []gatewayv1.HTTPRouteMatch{match},
		BackendRefs: []gatewayv1.HTTPBackendRef{backendRef},
		Timeouts:    timeout,
		Filters:     HTTPRouteRuleFilters(environment),
	}

	return rule
}

// HTTPRouteMatch prepares a path match for an HTTPRouteRule.
func HTTPRouteMatch(path string) gatewayv1.HTTPRouteMatch {
	pathMatchType := gatewayv1.PathMatchPathPrefix
	pathValue := path

	match := gatewayv1.HTTPRouteMatch{
		Path: &gatewayv1.HTTPPathMatch{Type: ptr.To(pathMatchType), Value: ptr.To(pathValue)},
	}

	return match
}

// HTTPBackendRef prepares a backend (i.e., a service) reference for an HTTPRouteRule.
func HTTPBackendRef(serviceName string, servicePort int32) gatewayv1.HTTPBackendRef {
	backendPort := gatewayv1.PortNumber(servicePort)

	backendRef := gatewayv1.HTTPBackendRef{
		BackendRef: gatewayv1.BackendRef{
			BackendObjectReference: gatewayv1.BackendObjectReference{
				Name: gatewayv1.ObjectName(serviceName),
				Port: ptr.To(backendPort),
			},
		},
	}

	return backendRef
}

// HTTPRouteTimeouts returns the default timeouts for an HTTPRouteRule.
func HTTPRouteTimeouts() *gatewayv1.HTTPRouteTimeouts {
	dur := gatewayv1.Duration(DefaultTimeoutSeconds)

	timeouts := &gatewayv1.HTTPRouteTimeouts{
		Request:        ptr.To(dur),
		BackendRequest: ptr.To(dur),
	}

	return timeouts
}

// HTTPRouteRuleFilters returns the slice of HTTPRouteFilters for the given environment.
func HTTPRouteRuleFilters(environment *clv1alpha2.Environment) []gatewayv1.HTTPRouteFilter {
	if environment == nil {
		return nil
	}
	switch environment.EnvironmentType {
	case clv1alpha2.ClassStandalone, clv1alpha2.ClassContainer:
		if !environment.RewriteURL {
			return nil
		}
		return []gatewayv1.HTTPRouteFilter{URLRewriteFilter(StandaloneRewriteEndpoint)}
	case clv1alpha2.ClassCloudVM, clv1alpha2.ClassLocalVM, clv1alpha2.ClassVM:
		return []gatewayv1.HTTPRouteFilter{URLRewriteFilter(GUIRewriteEndpoint)}
	default:
		return nil
	}
}

// URLRewriteFilter returns an URLRewrite filter for the given target endpoint.
func URLRewriteFilter(target string) gatewayv1.HTTPRouteFilter {
	rewriteType := gatewayv1.PrefixMatchHTTPPathModifier

	filter := gatewayv1.HTTPRouteFilter{
		Type: gatewayv1.HTTPRouteFilterURLRewrite,
		URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
			Path: &gatewayv1.HTTPPathModifier{
				Type:               rewriteType,
				ReplacePrefixMatch: ptr.To(target),
			},
		},
	}

	return filter
}

// GUI Path helpers for the environment GUI exposure.
// These functions below are used both for forging the HTTPRoute spec and for setting environment
// variables in the containers, to ensure consistency between the two.

// ExpositionGUIPath returns the path of the route targeting the environment GUI vnc or Standalone.
func ExpositionGUIPath(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) string {
	switch environment.EnvironmentType {
	case clv1alpha2.ClassStandalone, clv1alpha2.ClassContainer:
		if environment.RewriteURL {
			return fmt.Sprintf("%v/%v/%v(/|$)(.*)", ExpositionInstancePrefix, instance.UID, environment.Name)
		}
		return strings.TrimRight(fmt.Sprintf("%v/%v/%v", ExpositionInstancePrefix, instance.UID, environment.Name), "/")
	case clv1alpha2.ClassCloudVM, clv1alpha2.ClassLocalVM, clv1alpha2.ClassVM:
		return strings.TrimRight(fmt.Sprintf("%v/%v/%v/%s", ExpositionInstancePrefix, instance.UID, environment.Name, "(.*)"), "/")
	}
	return ""
}

// ExpositionGUICleanPath returns the path of the route targeting the environment GUI vnc or Standalone, without regex.
func ExpositionGUICleanPath(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) string {
	return strings.TrimRight(fmt.Sprintf("%v/%v/%v", ExpositionInstancePrefix, instance.UID, environment.Name), "/")
}

// ExpositionGuiStatusURL returns the path of the route targeting the environment.
func ExpositionGuiStatusURL(host string, environment *clv1alpha2.Environment, instance *clv1alpha2.Instance) string {
	return fmt.Sprintf("https://%v%v/%v/%v/", host, ExpositionInstancePrefix, instance.UID, environment.Name)
}

// ExpositionGuiStatusInstanceURL returns the root of the route url targeting an environment within the instance.
func ExpositionGuiStatusInstanceURL(host string, instance *clv1alpha2.Instance) string {
	return fmt.Sprintf("https://%v%v/%v/", host, ExpositionInstancePrefix, instance.UID)
}
