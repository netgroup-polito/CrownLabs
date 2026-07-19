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
	egv1alpha1 "github.com/envoyproxy/gateway/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("SecurityPolicy forging", func() {
	Describe("SecurityPolicySpec", func() {
		var (
			instance     *clv1alpha2.Instance
			environment  *clv1alpha2.Environment
			templateSpec *egv1alpha1.SecurityPolicySpec
			sp           egv1alpha1.SecurityPolicySpec
		)

		BeforeEach(func() {
			instance = &clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: "test-instance", Namespace: "tenant-test", UID: "1234"},
			}
			environment = &clv1alpha2.Environment{Name: "gui", EnvironmentType: clv1alpha2.ClassVM}
			templateSpec = &egv1alpha1.SecurityPolicySpec{
				ExtAuth: &egv1alpha1.ExtAuth{
					HeadersToExtAuth: []string{"Cookie", "Authorization"},
					HTTP: &egv1alpha1.HTTPExtAuthService{
						BackendRefs: []egv1alpha1.BackendRef{
							{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Group:     ptr.To(gatewayv1.Group("")),
									Kind:      ptr.To(gatewayv1.Kind("Service")),
									Name:      gatewayv1.ObjectName("my-auth"),
									Namespace: ptr.To(gatewayv1.Namespace("custom-ns")),
									Port:      ptr.To(gatewayv1.PortNumber(8080)),
								},
							},
						},
						Path: ptr.To("/check"),
					},
				},
			}
			targetRouteName := forge.ObjectMetaWithSuffix(instance, environment.Name).Name
			sp = forge.SecurityPolicySpec(targetRouteName, templateSpec)
		})

		It("Should configure the TargetRefs correctly", func() {
			Expect(sp.TargetRefs).To(HaveLen(1))
			ref := sp.TargetRefs[0]
			Expect(ref.Group).To(Equal(gwapiv1a2.Group("gateway.networking.k8s.io")))
			Expect(ref.Kind).To(Equal(gwapiv1a2.Kind("HTTPRoute")))
			Expect(ref.Name).To(Equal(gwapiv1a2.ObjectName("test-instance-gui")))
		})

		It("Should configure the ExtAuth HTTP service correctly based on templateSpec", func() {
			Expect(sp.ExtAuth).ToNot(BeNil())
			Expect(sp.ExtAuth.HTTP).ToNot(BeNil())
			httpAuth := sp.ExtAuth.HTTP

			Expect(*httpAuth.Path).To(Equal("/check"))
			Expect(sp.ExtAuth.HeadersToExtAuth).To(ContainElements("Cookie", "Authorization"))

			Expect(httpAuth.BackendRefs).To(HaveLen(1))
			backend := httpAuth.BackendRefs[0]
			Expect(*backend.Group).To(Equal(gatewayv1.Group("")))
			Expect(*backend.Kind).To(Equal(gatewayv1.Kind("Service")))
			Expect(backend.Name).To(Equal(gatewayv1.ObjectName("my-auth")))
			Expect(*backend.Namespace).To(Equal(gatewayv1.Namespace("custom-ns")))
			Expect(*backend.Port).To(Equal(gatewayv1.PortNumber(8080)))
		})
	})
})
