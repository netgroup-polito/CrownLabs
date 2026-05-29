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

package forge_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// DRY, table-driven tests for the HTTPRoute helpers in pkg/forge/httproutes.go.
// These tests maximize coverage while avoiding repetition via DescribeTable and helper builders.
var _ = Describe("HTTPRoute helpers (DRY)", func() {
	const (
		host        = "crownlabs.example.com"
		path        = "/instance/uuid/environment"
		serviceName = "service-name"
		gwName      = "crownlabs-gw"
		gwNs        = "envoy-gateway-system"
		gwSection   = "https"
		svcPort     = int32(6080)
	)

	// Helper to create environment pointers with minimal syntax.
	makeEnv := func(t clv1alpha2.EnvironmentType, rewrite bool, name string) *clv1alpha2.Environment {
		return &clv1alpha2.Environment{EnvironmentType: t, RewriteURL: rewrite, Name: name}
	}

	DescribeTable("RewriteFilterForEnvironment behavior",
		func(env *clv1alpha2.Environment, expectNil bool, expectedReplace string) {
			f := forge.RewriteFilterForEnvironment(env)
			if expectNil {
				Expect(f).To(BeNil())
				return
			}
			Expect(f).NotTo(BeNil())
			Expect(f.Type).To(Equal(gatewayv1.HTTPRouteFilterURLRewrite))
			Expect(f.URLRewrite).ToNot(BeNil())
			Expect(f.URLRewrite.Path).ToNot(BeNil())
			Expect(f.URLRewrite.Path.Type).To(Equal(gatewayv1.PrefixMatchHTTPPathModifier))
			Expect(*f.URLRewrite.Path.ReplacePrefixMatch).To(Equal(expectedReplace))
		},
		Entry("nil environment", nil, true, ""),
		Entry("rewrite disabled", makeEnv(clv1alpha2.ClassStandalone, false, ""), true, ""),
		Entry("standalone rewrite", makeEnv(clv1alpha2.ClassStandalone, true, ""), false, "/"),
		Entry("cloudvm rewrite", makeEnv(clv1alpha2.ClassCloudVM, true, ""), false, "/gui/"),
	)

	DescribeTable("HTTPRouteSpec and rule construction",
		func(gwName, gwNs, gwSection string, env *clv1alpha2.Environment, expectRewrite bool) {
			params := forge.HTTPRouteSpecParams{
				Host:               host,
				Path:               path,
				ServiceName:        serviceName,
				GatewayName:        gwName,
				GatewayNamespace:   gwNs,
				GatewaySectionName: gwSection,
			}
			spec := forge.HTTPRouteSpec(params, env, svcPort)

			// ParentRef assertions
			Expect(spec.CommonRouteSpec.ParentRefs).To(HaveLen(1))
			p := spec.ParentRefs[0]
			if gwName != "" {
				Expect(p.Name).To(Equal(gatewayv1.ObjectName(gwName)))
				if gwNs != "" {
					Expect(p.Namespace).ToNot(BeNil())
					Expect(*p.Namespace).To(Equal(gatewayv1.Namespace(gwNs)))
				}
				if gwSection != "" {
					Expect(p.SectionName).ToNot(BeNil())
					Expect(*p.SectionName).To(Equal(gatewayv1.SectionName(gwSection)))
				}
			} else {
				// empty parent when no gateway specified
				Expect(string(p.Name)).To(Equal(""))
				Expect(p.Namespace).To(BeNil())
				Expect(p.SectionName).To(BeNil())
			}

			// Rules and matches
			Expect(spec.Rules).To(HaveLen(1))
			rule := spec.Rules[0]
			Expect(rule.Matches).To(HaveLen(1))
			Expect(*rule.Matches[0].Path.Type).To(Equal(gatewayv1.PathMatchPathPrefix))
			Expect(*rule.Matches[0].Path.Value).To(Equal(path))

			// Backend numeric port
			Expect(rule.BackendRefs).To(HaveLen(1))
			br := rule.BackendRefs[0].BackendObjectReference
			Expect(br.Name).To(Equal(gatewayv1.ObjectName(serviceName)))
			Expect(br.Port).ToNot(BeNil())
			Expect(*br.Port).To(Equal(gatewayv1.PortNumber(svcPort)))

			// Timeouts always present
			Expect(rule.Timeouts).ToNot(BeNil())
			Expect(string(*rule.Timeouts.Request)).To(Equal("3600s"))
			Expect(string(*rule.Timeouts.BackendRequest)).To(Equal("3600s"))

			// Filters: either empty or contain a single URLRewrite
			if expectRewrite {
				Expect(rule.Filters).To(HaveLen(1))
				Expect(rule.Filters[0].Type).To(Equal(gatewayv1.HTTPRouteFilterURLRewrite))
			} else {
				Expect(rule.Filters).To(BeEmpty())
			}
		},
		Entry("gateway + nil env no rewrite", gwName, gwNs, gwSection, nil, false),
		Entry("no gateway + standalone rewrite", "", "", "", makeEnv(clv1alpha2.ClassStandalone, true, ""), true),
		Entry("no gateway + cloudvm rewrite", "", "", "", makeEnv(clv1alpha2.ClassCloudVM, true, ""), true),
		Entry("gateway but rewrite disabled", gwName, gwNs, gwSection, makeEnv(clv1alpha2.ClassStandalone, false, ""), false),
	)

	DescribeTable("GUI path helpers produce expected paths/URLs",
		func(instUID string, env *clv1alpha2.Environment, expectedPath, expectedClean, expectedStatus, expectedInstanceRoot, expectedFromRoot string) {
			inst := clv1alpha2.Instance{}
			inst.UID = types.UID(instUID)

			// Path
			p := forge.HTTPRouteGUIPath(&inst, env)
			Expect(p).To(Equal(expectedPath))

			// Clean
			c := forge.HTTPRouteGUICleanPath(&inst, env)
			Expect(c).To(Equal(expectedClean))

			// Status URL
			s := forge.HTTPRouteGuiStatusURL(host, env, &inst)
			Expect(s).To(Equal(expectedStatus))

			// Instance root URL
			r := forge.HTTPRouteGuiStatusInstanceURL(host, &inst)
			Expect(r).To(Equal(expectedInstanceRoot))

			// From root
			fr := forge.HTTPRouteGuiStatusFromRootURL(expectedInstanceRoot, env)
			Expect(fr).To(Equal(expectedFromRoot))
		},
		Entry("Standalone with rewrite", "abcd-1", makeEnv(clv1alpha2.ClassStandalone, true, "env"), "/instance/abcd-1/env(/|$)(.*)", "/instance/abcd-1/env", "https://"+host+"/instance/abcd-1/env/", "https://"+host+"/instance/abcd-1/", "https://"+host+"/instance/abcd-1/env/"),
		Entry("Standalone without rewrite", "abcd-2", makeEnv(clv1alpha2.ClassStandalone, false, "env2"), "/instance/abcd-2/env2", "/instance/abcd-2/env2", "https://"+host+"/instance/abcd-2/env2/", "https://"+host+"/instance/abcd-2/", "https://"+host+"/instance/abcd-2/env2/"),
		Entry("CloudVM", "uid-3", makeEnv(clv1alpha2.ClassCloudVM, false, "envvm"), "/instance/uid-3/envvm/(.*)", "/instance/uid-3/envvm", "https://"+host+"/instance/uid-3/envvm/", "https://"+host+"/instance/uid-3/", "https://"+host+"/instance/uid-3/envvm/"),
	)
})