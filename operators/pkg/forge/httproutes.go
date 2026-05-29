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

// TODO not here
//   - TLS: configure on the Gateway;
//   - Auth (auth-url/auth-signin): use SecurityPolicy;

package forge

import (
	"fmt"
	"strings"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// HTTPRouteInstancePrefix -> the prefix prepended to the path of any HTTPRoute
	// targeting the instance or its subresources.
	HTTPRouteInstancePrefix = "/instance"

	// DefaultTimeoutSeconds -> the default timeout as a Gateway API Duration string (GEP-2257 format).
	DefaultTimeoutSeconds = "3600s"

	// StandaloneRewriteEndpointHTTPRoute -> the endpoint used to rewrite standalone GUI URLs.
	// TODO: change name.
	StandaloneRewriteEndpointHTTPRoute = "/"

	// GUIRewriteEndpointHTTPRoute -> the endpoint used to rewrite CloudVM/VM GUI URLs.
	// TODO: change name.
	GUIRewriteEndpointHTTPRoute = "/gui/"
)

// HTTPRouteSpecParams groups the string parameters required to forge an HTTPRouteSpec.
type HTTPRouteSpecParams struct {
	Host               string
	Path               string
	ServiceName        string
	GatewayName        string
	GatewayNamespace   string
	GatewaySectionName string
}

// HTTPRouteSpec forges the specification of a Kubernetes HTTPRoute resource.
func HTTPRouteSpec(params *HTTPRouteSpecParams, environment *clv1alpha2.Environment, servicePort int32) gatewayv1.HTTPRouteSpec {
	parentRef := buildParentReference(params.GatewayName, params.GatewayNamespace, params.GatewaySectionName)
	rule := buildRouteRule(params.Path, params.ServiceName, servicePort, environment)

	spec := gatewayv1.HTTPRouteSpec{
		Hostnames: []gatewayv1.Hostname{gatewayv1.Hostname(params.Host)},
		CommonRouteSpec: gatewayv1.CommonRouteSpec{
			ParentRefs: []gatewayv1.ParentReference{parentRef},
		},
		Rules: []gatewayv1.HTTPRouteRule{rule},
	}

	return spec
}

// buildParentReference creates the appropriate ParentReference for the provided gateway.
func buildParentReference(gatewayName, gatewayNamespace, gatewaySectionName string) gatewayv1.ParentReference {
	if gatewayName != "" {
		parent := gatewayv1.ParentReference{Name: gatewayv1.ObjectName(gatewayName)}
		if gatewayNamespace != "" {
			namespace := gatewayv1.Namespace(gatewayNamespace)
			parent.Namespace = &namespace
		}
		if gatewaySectionName != "" {
			sectionName := gatewayv1.SectionName(gatewaySectionName)
			parent.SectionName = &sectionName
		}
		return parent
	}
	// No gateway specified, return empty parent refs to create an "orphan" HTTPRoute.
	return gatewayv1.ParentReference{}
}

// buildRouteRule constructs a complete HTTPRouteRule including match, backend reference and timeout filters.
func buildRouteRule(path, serviceName string, servicePort int32, environment *clv1alpha2.Environment) gatewayv1.HTTPRouteRule {
	pathMatchType := gatewayv1.PathMatchPathPrefix
	pathValue := path
	backendPort := gatewayv1.PortNumber(servicePort)
	rewrite := RewriteFilterForEnvironment(environment)

	// Prepare filters
	var filters []gatewayv1.HTTPRouteFilter
	if rewrite != nil {
		filters = append(filters, *rewrite)
	}

	rule := gatewayv1.HTTPRouteRule{
		Matches: []gatewayv1.HTTPRouteMatch{{
			Path: &gatewayv1.HTTPPathMatch{Type: &pathMatchType, Value: &pathValue},
		}},
		BackendRefs: []gatewayv1.HTTPBackendRef{{
			BackendRef: gatewayv1.BackendRef{
				BackendObjectReference: gatewayv1.BackendObjectReference{
					Name: gatewayv1.ObjectName(serviceName),
					Port: &backendPort,
				},
			},
		}},
		Timeouts: func() *gatewayv1.HTTPRouteTimeouts {
			d := gatewayv1.Duration(DefaultTimeoutSeconds)
			return &gatewayv1.HTTPRouteTimeouts{
				Request:        &d,
				BackendRequest: &d,
			}
		}(),
		Filters: filters,
	}

	return rule
}

// RewriteFilterForEnvironment returns an URLRewrite filter for the given environment.
func RewriteFilterForEnvironment(environment *clv1alpha2.Environment) *gatewayv1.HTTPRouteFilter {
	if environment == nil || !environment.RewriteURL {
		return nil
	}

	var target string
	switch environment.EnvironmentType {
	case clv1alpha2.ClassStandalone, clv1alpha2.ClassContainer:
		target = StandaloneRewriteEndpointHTTPRoute
	case clv1alpha2.ClassCloudVM, clv1alpha2.ClassVM:
		target = GUIRewriteEndpointHTTPRoute
	}

	rewriteType := gatewayv1.PrefixMatchHTTPPathModifier
	return &gatewayv1.HTTPRouteFilter{
		Type: gatewayv1.HTTPRouteFilterURLRewrite,
		URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
			Path: &gatewayv1.HTTPPathModifier{
				Type:               rewriteType,
				ReplacePrefixMatch: &target,
			},
		},
	}
}

// HTTPRouteGUIPath returns the path of the route targeting the environment GUI vnc or Standalone.
func HTTPRouteGUIPath(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) string {
	switch environment.EnvironmentType {
	case clv1alpha2.ClassStandalone, clv1alpha2.ClassContainer:
		if environment.RewriteURL {
			return fmt.Sprintf("%v/%v/%v(/|$)(.*)", HTTPRouteInstancePrefix, instance.UID, environment.Name)
		}
		return strings.TrimRight(fmt.Sprintf("%v/%v/%v", HTTPRouteInstancePrefix, instance.UID, environment.Name), "/")
	case clv1alpha2.ClassCloudVM, clv1alpha2.ClassVM:
		return strings.TrimRight(fmt.Sprintf("%v/%v/%v/%s", HTTPRouteInstancePrefix, instance.UID, environment.Name, "(.*)"), "/")
	}
	return ""
}

// HTTPRouteGUICleanPath returns the path of the route targeting the environment GUI vnc or Standalone, without regex.
func HTTPRouteGUICleanPath(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) string {
	return strings.TrimRight(fmt.Sprintf("%v/%v/%v", HTTPRouteInstancePrefix, instance.UID, environment.Name), "/")
}

// HTTPRouteGuiStatusURL returns the path of the route targeting the environment.
func HTTPRouteGuiStatusURL(host string, environment *clv1alpha2.Environment, instance *clv1alpha2.Instance) string {
	return fmt.Sprintf("https://%v%v/%v/%v/", host, HTTPRouteInstancePrefix, instance.UID, environment.Name)
}

// HTTPRouteGuiStatusInstanceURL returns the root of the route url targeting an environment within the instance.
func HTTPRouteGuiStatusInstanceURL(host string, instance *clv1alpha2.Instance) string {
	return fmt.Sprintf("https://%v%v/%v/", host, HTTPRouteInstancePrefix, instance.UID)
}

// HTTPRouteGuiStatusFromRootURL returns the path of the route targeting the environment given the root url (url of the instance).
func HTTPRouteGuiStatusFromRootURL(rootURL string, environment *clv1alpha2.Environment) string {
	return rootURL + fmt.Sprintf("%v/", environment.Name)
}
