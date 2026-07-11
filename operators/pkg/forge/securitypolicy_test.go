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
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
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

		It("Should return default values for empty string", func() {
			service, ns, port, path, mode, err := forge.ParseAuthServiceAnnotation("", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal("default"))
			Expect(service).To(Equal(defService))
			Expect(ns).To(Equal(defNamespace))
			Expect(port).To(Equal(defPort))
			Expect(path).To(Equal(defPath))
		})

		It("Should return default values for 'default' string", func() {
			service, ns, port, path, mode, err := forge.ParseAuthServiceAnnotation("default", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal("default"))
			Expect(service).To(Equal(defService))
			Expect(ns).To(Equal(defNamespace))
			Expect(port).To(Equal(defPort))
			Expect(path).To(Equal(defPath))
		})

		It("Should return none for 'none' string", func() {
			service, ns, port, path, mode, err := forge.ParseAuthServiceAnnotation("none", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal("none"))
			Expect(service).To(BeEmpty())
			Expect(ns).To(BeEmpty())
			Expect(port).To(BeZero())
			Expect(path).To(BeEmpty())
		})

		It("Should return none for 'disabled' string", func() {
			service, ns, port, path, mode, err := forge.ParseAuthServiceAnnotation("disabled", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal("none"))
			Expect(service).To(BeEmpty())
			Expect(ns).To(BeEmpty())
			Expect(port).To(BeZero())
			Expect(path).To(BeEmpty())
		})

		It("Should parse full custom service", func() {
			service, ns, port, path, mode, err := forge.ParseAuthServiceAnnotation("my-auth.custom-ns:8080/check-auth", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal("custom"))
			Expect(service).To(Equal("my-auth"))
			Expect(ns).To(Equal("custom-ns"))
			Expect(port).To(Equal(int32(8080)))
			Expect(path).To(Equal("/check-auth"))
		})

		It("Should parse custom service without namespace, using tenant namespace", func() {
			service, ns, port, path, mode, err := forge.ParseAuthServiceAnnotation("exam-agent:9000/auth", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal("custom"))
			Expect(service).To(Equal("exam-agent"))
			Expect(ns).To(Equal(tenantNamespace))
			Expect(port).To(Equal(int32(9000)))
			Expect(path).To(Equal("/auth"))
		})

		It("Should parse custom service without port, defaulting to 80", func() {
			service, ns, port, path, mode, err := forge.ParseAuthServiceAnnotation("exam-agent.custom-ns/auth", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal("custom"))
			Expect(service).To(Equal("exam-agent"))
			Expect(ns).To(Equal("custom-ns"))
			Expect(port).To(Equal(int32(80)))
			Expect(path).To(Equal("/auth"))
		})

		It("Should parse custom service without path, defaulting to /check", func() {
			service, ns, port, path, mode, err := forge.ParseAuthServiceAnnotation("exam-agent.custom-ns:9000", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal("custom"))
			Expect(service).To(Equal("exam-agent"))
			Expect(ns).To(Equal("custom-ns"))
			Expect(port).To(Equal(int32(9000)))
			Expect(path).To(Equal("/check"))
		})

		It("Should parse minimalist custom service", func() {
			service, ns, port, path, mode, err := forge.ParseAuthServiceAnnotation("simple-agent", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal("custom"))
			Expect(service).To(Equal("simple-agent"))
			Expect(ns).To(Equal(tenantNamespace))
			Expect(port).To(Equal(int32(80)))
			Expect(path).To(Equal("/check"))
		})

		It("Should fail on invalid port", func() {
			_, _, _, _, _, err := forge.ParseAuthServiceAnnotation("exam-agent:invalid/auth", tenantNamespace, defService, defNamespace, defPort, defPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("SecurityPolicy", func() {
		var (
			instance    *clv1alpha2.Instance
			environment *clv1alpha2.Environment
			sp          egv1alpha1.SecurityPolicy
		)

		BeforeEach(func() {
			instance = &clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: "test-instance", Namespace: "tenant-test", UID: "1234"},
			}
			environment = &clv1alpha2.Environment{Name: "gui", EnvironmentType: clv1alpha2.ClassVM}
			sp = forge.SecurityPolicy(instance, environment, "my-auth", "custom-ns", 8080, "/check")
		})

		It("Should configure the TargetRefs correctly", func() {
			Expect(sp.Spec.TargetRefs).To(HaveLen(1))
			ref := sp.Spec.TargetRefs[0]
			Expect(ref.Group).To(Equal(gwapiv1a2.Group("gateway.networking.k8s.io")))
			Expect(ref.Kind).To(Equal(gwapiv1a2.Kind("HTTPRoute")))
			Expect(ref.Name).To(Equal(gwapiv1a2.ObjectName("test-instance-gui")))
		})

		It("Should configure the ExtAuth HTTP service correctly", func() {
			Expect(sp.Spec.ExtAuth).ToNot(BeNil())
			Expect(sp.Spec.ExtAuth.HTTP).ToNot(BeNil())
			httpAuth := sp.Spec.ExtAuth.HTTP

			Expect(*httpAuth.Path).To(Equal("/check"))
			Expect(sp.Spec.ExtAuth.HeadersToExtAuth).To(ContainElements("Cookie", "Authorization"))

			Expect(httpAuth.BackendRefs).To(HaveLen(1))
			bRef := httpAuth.BackendRefs[0]
			Expect(bRef.Group).To(Equal(ptr.To(gatewayv1.Group(""))))
			Expect(bRef.Kind).To(Equal(ptr.To(gatewayv1.Kind("Service"))))
			Expect(bRef.Name).To(Equal(gatewayv1.ObjectName("my-auth")))
			Expect(bRef.Namespace).To(Equal(ptr.To(gatewayv1.Namespace("custom-ns"))))
			Expect(bRef.Port).To(Equal(ptr.To(gatewayv1.PortNumber(8080))))
		})

		It("Should have the correct labels and name", func() {
			Expect(sp.Name).To(Equal("test-instance-gui"))
			Expect(sp.Namespace).To(Equal("tenant-test"))
			Expect(sp.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "instance"))
		})
	})
})
