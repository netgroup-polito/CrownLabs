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

var _ = Describe("HTTPRoute helpers", func() {
	const (
		instanceUID = "dcc6ead1-0040-451b-ba68-787ebfb68640"
		host        = "crownlabs.example.com"
	)

	Describe("ParseGatewayParent behavior", func() {
		It("ParseGatewayParent parses valid parent references", func() {
			ns, name, err := forge.ParseGatewayParent("my-ns/my-gateway")
			Expect(err).ToNot(HaveOccurred())
			Expect(ns).To(Equal("my-ns"))
			Expect(name).To(Equal("my-gateway"))
		})

		Context("ParseGatewayParent errors on invalid input", func() {
			It("errors on empty string", func() {
				_, _, err := forge.ParseGatewayParent("")
				Expect(err).To(HaveOccurred())
			})

			It("errors on missing namespace", func() {
				_, _, err := forge.ParseGatewayParent("/my-gateway")
				Expect(err).To(HaveOccurred())
			})

			It("errors on missing name", func() {
				_, _, err := forge.ParseGatewayParent("my-ns/")
				Expect(err).To(HaveOccurred())
			})

			It("errors on extra slashes", func() {
				_, _, err := forge.ParseGatewayParent("my-ns/my-gateway/extra")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("HTTPRoute creation behavior", func() {
		It("HTTPRouteSpec builds correctly a route", func() {
			tpl := forge.HTTPRouteTemplate{Path: "/instance/uid/env", ServiceName: "svc"}
			expo := forge.ExpositionConfig{GatewayName: "gw", GatewayNamespace: "gw-ns"}
			env := &clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassCloudVM, RewriteURL: true}
			spec := forge.HTTPRouteSpec(&tpl, &expo, env, 8080)

			Expect(spec).ToNot(BeNil())
		})

		It("HTTPRouteSpec does not include empty filters when rewrite is disabled", func() {
			tpl := forge.HTTPRouteTemplate{Path: "/instance/uid/env", ServiceName: "svc"}
			expo := forge.ExpositionConfig{GatewayName: "gw", GatewayNamespace: "gw-ns"}
			env := &clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassStandalone, RewriteURL: false}
			spec := forge.HTTPRouteSpec(&tpl, &expo, env, 8080)

			Expect(spec).ToNot(BeNil())
			Expect(spec.Rules).To(HaveLen(1))
			Expect(spec.Rules[0].Filters).To(BeEmpty())
		})

		It("ParentReference sets name and namespace", func() {
			p := forge.ParentReference("gw", "gw-ns")
			Expect(p.Name).To(Equal(gatewayv1.ObjectName("gw")))
			Expect(*p.Namespace).To(Equal(gatewayv1.Namespace("gw-ns")))
		})

		It("HTTPRouteMatch constructs prefix path match", func() {
			m := forge.HTTPRouteMatch("/instance/uid/env")
			Expect(*m.Path.Type).To(Equal(gatewayv1.PathMatchPathPrefix))
			Expect(*m.Path.Value).To(Equal("/instance/uid/env"))
		})

		It("HTTPBackendRef sets the backend name and numeric port", func() {
			br := forge.HTTPBackendRef("svc", 8080)
			bo := br.BackendObjectReference
			Expect(bo.Name).To(Equal(gatewayv1.ObjectName("svc")))
			Expect(bo.Port).ToNot(BeNil())
			Expect(*bo.Port).To(Equal(gatewayv1.PortNumber(8080)))
		})

		It("HTTPRouteTimeouts returns the default durations", func() {
			t := forge.HTTPRouteTimeouts()
			Expect(string(*t.Request)).To(Equal(forge.DefaultTimeoutSeconds))
			Expect(string(*t.BackendRequest)).To(Equal(forge.DefaultTimeoutSeconds))
		})

		It("URLRewriteFilter constructs a URLRewrite filter", func() {
			f := forge.URLRewriteFilter("/")
			Expect(f.Type).To(Equal(gatewayv1.HTTPRouteFilterURLRewrite))
			Expect(f.URLRewrite).ToNot(BeNil())
			Expect(f.URLRewrite.Path).ToNot(BeNil())
			Expect(f.URLRewrite.Path.Type).To(Equal(gatewayv1.PrefixMatchHTTPPathModifier))
			Expect(*f.URLRewrite.Path.ReplacePrefixMatch).To(Equal("/"))
		})

		Context("HTTPRouteRuleFilters cases", func() {
			makeEnv := func(t clv1alpha2.EnvironmentType, rewrite bool, name string) *clv1alpha2.Environment {
				return &clv1alpha2.Environment{EnvironmentType: t, RewriteURL: rewrite, Name: name}
			}

			It("returns nil for nil environment", func() {
				filters := forge.HTTPRouteRuleFilters(nil)
				Expect(filters).To(BeNil())
			})

			It("returns nil when rewrite disabled", func() {
				filters := forge.HTTPRouteRuleFilters(makeEnv(clv1alpha2.ClassStandalone, false, ""))
				Expect(filters).To(BeNil())
			})

			It("returns standalone rewrite when requested and environment is standalone", func() {
				filters := forge.HTTPRouteRuleFilters(makeEnv(clv1alpha2.ClassStandalone, true, ""))
				Expect(filters).To(HaveLen(1))
				f := filters[0]
				Expect(f.Type).To(Equal(gatewayv1.HTTPRouteFilterURLRewrite))
				Expect(*f.URLRewrite.Path.ReplacePrefixMatch).To(Equal(forge.StandaloneRewriteEndpoint))
			})

			It("returns GUI rewrite when requested and environment is cloud VM", func() {
				filters := forge.HTTPRouteRuleFilters(makeEnv(clv1alpha2.ClassCloudVM, true, ""))
				Expect(filters).To(HaveLen(1))
				f := filters[0]
				Expect(f.Type).To(Equal(gatewayv1.HTTPRouteFilterURLRewrite))
				Expect(*f.URLRewrite.Path.ReplacePrefixMatch).To(Equal(forge.GUIRewriteEndpoint))
			})

			It("returns GUI rewrite when requested and environment is local VM", func() {
				filters := forge.HTTPRouteRuleFilters(makeEnv(clv1alpha2.ClassLocalVM, true, ""))
				Expect(filters).To(HaveLen(1))
				f := filters[0]
				Expect(f.Type).To(Equal(gatewayv1.HTTPRouteFilterURLRewrite))
				Expect(*f.URLRewrite.Path.ReplacePrefixMatch).To(Equal(forge.GUIRewriteEndpoint))
			})

			It("returns GUI rewrite when requested and environment is VM", func() {
				filters := forge.HTTPRouteRuleFilters(makeEnv(clv1alpha2.ClassVM, true, ""))
				Expect(filters).To(HaveLen(1))
				f := filters[0]
				Expect(f.Type).To(Equal(gatewayv1.HTTPRouteFilterURLRewrite))
				Expect(*f.URLRewrite.Path.ReplacePrefixMatch).To(Equal(forge.GUIRewriteEndpoint))
			})

			It("returns nil for unknown environment type", func() {
				filters := forge.HTTPRouteRuleFilters(makeEnv(clv1alpha2.EnvironmentType("unknown"), true, ""))
				Expect(filters).To(BeNil())
			})
		})

		It("HTTPRouteSpec builds a route with expected parent, match, backend and timeouts for a cloud VM environment", func() {
			tpl := forge.HTTPRouteTemplate{Path: "/instance/uid/env", ServiceName: "svc"}
			expo := forge.ExpositionConfig{GatewayName: "gw", GatewayNamespace: "gw-ns"}
			env := &clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassCloudVM, RewriteURL: true}
			spec := forge.HTTPRouteSpec(&tpl, &expo, env, 6080)

			// parent
			Expect(spec.ParentRefs).To(HaveLen(1))
			Expect(spec.ParentRefs[0].Name).To(Equal(gatewayv1.ObjectName("gw")))
			Expect(*spec.ParentRefs[0].Namespace).To(Equal(gatewayv1.Namespace("gw-ns")))

			// rule/match
			Expect(spec.Rules).To(HaveLen(1))
			r := spec.Rules[0]
			Expect(*r.Matches[0].Path.Type).To(Equal(gatewayv1.PathMatchPathPrefix))
			Expect(*r.Matches[0].Path.Value).To(Equal("/instance/uid/env"))

			// backend
			bo := r.BackendRefs[0].BackendObjectReference
			Expect(bo.Name).To(Equal(gatewayv1.ObjectName("svc")))
			Expect(*bo.Port).To(Equal(gatewayv1.PortNumber(6080)))

			// timeouts
			Expect(r.Timeouts).ToNot(BeNil())
			Expect(string(*r.Timeouts.Request)).To(Equal(forge.DefaultTimeoutSeconds))
		})
	})

	Describe("Exposition GUI path helpers", func() {
		var instance clv1alpha2.Instance
		var env clv1alpha2.Environment

		BeforeEach(func() {
			instance = clv1alpha2.Instance{ObjectMeta: metav1.ObjectMeta{UID: instanceUID}}
			env.Name = "env-name"
		})

		It("ExpositionGUIPath returns regex path when rewrite enabled for standalone", func() {
			env.EnvironmentType = clv1alpha2.ClassStandalone
			env.RewriteURL = true
			p := forge.ExpositionGUIPath(&instance, &env)
			Expect(p).To(Equal("/instance/" + instanceUID + "/" + env.Name + "(/|$)(.*)"))
		})

		It("ExpositionGUIPath returns clean path when rewrite disabled for standalone", func() {
			env.EnvironmentType = clv1alpha2.ClassStandalone
			env.RewriteURL = false
			p := forge.ExpositionGUIPath(&instance, &env)
			Expect(p).To(Equal("/instance/" + instanceUID + "/" + env.Name))
		})

		It("ExpositionGUIPath returns the GUI path for local VMs", func() {
			env.EnvironmentType = clv1alpha2.ClassLocalVM
			env.RewriteURL = true
			p := forge.ExpositionGUIPath(&instance, &env)
			Expect(p).To(Equal("/instance/" + instanceUID + "/" + env.Name + "/(.*)"))
		})

		It("ExpositionGUIPath returns the GUI path for cloud VMs", func() {
			env.EnvironmentType = clv1alpha2.ClassCloudVM
			env.RewriteURL = true
			p := forge.ExpositionGUIPath(&instance, &env)
			Expect(p).To(Equal("/instance/" + instanceUID + "/" + env.Name + "/(.*)"))
		})

		It("ExpositionGUIPath returns the GUI path for VMs", func() {
			env.EnvironmentType = clv1alpha2.ClassVM
			env.RewriteURL = true
			p := forge.ExpositionGUIPath(&instance, &env)
			Expect(p).To(Equal("/instance/" + instanceUID + "/" + env.Name + "/(.*)"))
		})

		It("ExpositionGuiStatusURL composes the expected status URL", func() {
			u := forge.ExpositionGuiStatusURL(host, &env, &instance)
			Expect(u).To(Equal("https://" + host + "/instance/" + instanceUID + "/" + env.Name + "/"))
		})

		It("ExpositionGuiStatusInstanceURL composes the expected instance root URL", func() {
			u := forge.ExpositionGuiStatusInstanceURL(host, &instance)
			Expect(u).To(Equal("https://" + host + "/instance/" + instanceUID + "/"))
		})
	})
})
