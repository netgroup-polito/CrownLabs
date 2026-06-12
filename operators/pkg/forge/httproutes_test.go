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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// DRY, table-driven tests for the HTTPRoute helpers in pkg/forge/httproutes.go.
// These tests maximize coverage while avoiding repetition via DescribeTable and helper builders.
var _ = Describe("HTTPRoute", func() {
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
			tpl := forge.HTTPRouteTemplate{Path: path, ServiceName: serviceName}
			expo := forge.ExpositionConfig{GatewayName: gwName, GatewayNamespace: gwNs, GatewaySectionName: gwSection}
			spec := forge.HTTPRouteSpec(&tpl, &expo, env, svcPort)

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

	Describe("The forge.Exposition*Path functions", func() {
		var (
			instance    clv1alpha2.Instance
			path        string
			statusPath  string
			environment clv1alpha2.Environment
		)

		const (
			instanceName      = "kubernetes-0000"
			instanceNamespace = "tenant-tester"
			instanceUID       = "dcc6ead1-0040-451b-ba68-787ebfb68640"
			environmentName   = "environment-name"
			host              = "crownlabs.example.com"
		)

		BeforeEach(func() {
			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace, UID: instanceUID},
			}
			environment.Name = environmentName
		})

		Describe("The forge.ExpositionGUIPath function", func() {
			JustBeforeEach(func() {
				path = forge.ExpositionGUIPath(&instance, &environment)
			})
			When("EnvironmentType is ClassStandalone", func() {
				BeforeEach(func() {
					environment.EnvironmentType = clv1alpha2.ClassStandalone
				})
				When("Rewrite is true", func() {
					BeforeEach(func() {
						environment.RewriteURL = true
					})
					Context("The instance has no special configurations", func() {
						It("Should generate a path based on the instance UID", func() {
							Expect(path).To(BeIdenticalTo("/instance/" + instanceUID + "/" + environment.Name + "(/|$)(.*)"))
						})
					})
				})
				When("Rewrite is false", func() {
					BeforeEach(func() {
						environment.RewriteURL = false
					})
					Context("The instance has no special configurations", func() {
						It("Should generate a path based on the instance UID", func() {
							Expect(path).To(BeIdenticalTo("/instance/" + instanceUID + "/" + environment.Name))
						})
					})
				})
			})

		})

		Describe("The forge.ExpositionGuiStatusURL function", func() {
			JustBeforeEach(func() {
				statusPath = forge.ExpositionGuiStatusURL(host, &environment, &instance)
			})
			It("Should generate a path based on the instance UID and /app at the end", func() {
				Expect(statusPath).To(BeIdenticalTo("https://" + host + "/instance/" + instanceUID + "/" + environment.Name + "/"))
			})
		})

	})
})
