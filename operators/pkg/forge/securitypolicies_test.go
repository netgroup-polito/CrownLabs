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
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("SecurityPolicy forging", func() {
	Describe("ParseAuthServiceAnnotation", func() {
		const (
			tenantNamespace = "tenant-test"
			defService      = "oauth2-proxy"
			defNamespace    = "crownlabs"
			defPort         = int32(4180)
			defPath         = "/auth"
		)

		var (
			defaultAuth = &forge.AuthServiceInfo{
				ServiceName: defService,
				Namespace:   defNamespace,
				Port:        defPort,
				Path:        defPath,
				Mode:        "default",
			}
		)

		It("Should return default values for empty string", func() {
			info, err := forge.ParseAuthServiceAnnotation("", tenantNamespace, defaultAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode).To(Equal("default"))
			Expect(info.ServiceName).To(Equal(defService))
			Expect(info.Namespace).To(Equal(defNamespace))
			Expect(info.Port).To(Equal(defPort))
			Expect(info.Path).To(Equal(defPath))
		})

		It("Should return default values for 'default' string", func() {
			info, err := forge.ParseAuthServiceAnnotation("default", tenantNamespace, defaultAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode).To(Equal("default"))
			Expect(info.ServiceName).To(Equal(defService))
			Expect(info.Namespace).To(Equal(defNamespace))
			Expect(info.Port).To(Equal(defPort))
			Expect(info.Path).To(Equal(defPath))
		})

		It("Should return none for 'none' string", func() {
			info, err := forge.ParseAuthServiceAnnotation("none", tenantNamespace, defaultAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode).To(Equal("none"))
			Expect(info.ServiceName).To(BeEmpty())
			Expect(info.Namespace).To(BeEmpty())
			Expect(info.Port).To(BeZero())
			Expect(info.Path).To(BeEmpty())
		})

		It("Should return none for 'disabled' string", func() {
			info, err := forge.ParseAuthServiceAnnotation("disabled", tenantNamespace, defaultAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode).To(Equal("none"))
			Expect(info.ServiceName).To(BeEmpty())
			Expect(info.Namespace).To(BeEmpty())
			Expect(info.Port).To(BeZero())
			Expect(info.Path).To(BeEmpty())
		})

		It("Should parse full custom service", func() {
			info, err := forge.ParseAuthServiceAnnotation("my-auth.custom-ns:8080/check-auth", tenantNamespace, defaultAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode).To(Equal("custom"))
			Expect(info.ServiceName).To(Equal("my-auth"))
			Expect(info.Namespace).To(Equal("custom-ns"))
			Expect(info.Port).To(Equal(int32(8080)))
			Expect(info.Path).To(Equal("/check-auth"))
		})

		It("Should parse custom service without namespace, using tenant namespace", func() {
			info, err := forge.ParseAuthServiceAnnotation("exam-agent:9000/auth", tenantNamespace, defaultAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode).To(Equal("custom"))
			Expect(info.ServiceName).To(Equal("exam-agent"))
			Expect(info.Namespace).To(Equal(tenantNamespace))
			Expect(info.Port).To(Equal(int32(9000)))
			Expect(info.Path).To(Equal("/auth"))
		})

		It("Should parse custom service without port, defaulting to 80", func() {
			info, err := forge.ParseAuthServiceAnnotation("exam-agent.custom-ns/auth", tenantNamespace, defaultAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode).To(Equal("custom"))
			Expect(info.ServiceName).To(Equal("exam-agent"))
			Expect(info.Namespace).To(Equal("custom-ns"))
			Expect(info.Port).To(Equal(int32(80)))
			Expect(info.Path).To(Equal("/auth"))
		})

		It("Should parse custom service without path, defaulting to /check", func() {
			info, err := forge.ParseAuthServiceAnnotation("exam-agent.custom-ns:9000", tenantNamespace, defaultAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode).To(Equal("custom"))
			Expect(info.ServiceName).To(Equal("exam-agent"))
			Expect(info.Namespace).To(Equal("custom-ns"))
			Expect(info.Port).To(Equal(int32(9000)))
			Expect(info.Path).To(Equal("/check"))
		})

		It("Should parse minimalist custom service", func() {
			info, err := forge.ParseAuthServiceAnnotation("simple-agent", tenantNamespace, defaultAuth)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode).To(Equal("custom"))
			Expect(info.ServiceName).To(Equal("simple-agent"))
			Expect(info.Namespace).To(Equal(tenantNamespace))
			Expect(info.Port).To(Equal(int32(80)))
			Expect(info.Path).To(Equal("/check"))
		})

		It("Should fail on invalid port", func() {
			_, err := forge.ParseAuthServiceAnnotation("exam-agent:invalid/auth", tenantNamespace, defaultAuth)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("SecurityPolicy", func() {
		var (
			instance    *clv1alpha2.Instance
			environment *clv1alpha2.Environment
			sp          egv1alpha1.SecurityPolicySpec
		)

		BeforeEach(func() {
			instance = &clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: "test-instance", Namespace: "tenant-test", UID: "1234"},
			}
			environment = &clv1alpha2.Environment{Name: "gui", EnvironmentType: clv1alpha2.ClassVM}
			sp = forge.SecurityPolicySpec(instance, environment, &forge.AuthServiceInfo{
				ServiceName: "my-auth",
				Namespace:   "custom-ns",
				Port:        8080,
				Path:        "/check",
			})
		})

		It("Should configure the TargetRefs correctly", func() {
			Expect(sp.TargetRefs).To(HaveLen(1))
			ref := sp.TargetRefs[0]
			Expect(ref.Group).To(Equal(gwapiv1a2.Group("gateway.networking.k8s.io")))
			Expect(ref.Kind).To(Equal(gwapiv1a2.Kind("HTTPRoute")))
			Expect(ref.Name).To(Equal(gwapiv1a2.ObjectName("test-instance-gui")))
		})

		It("Should configure the ExtAuth HTTP service correctly", func() {
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
