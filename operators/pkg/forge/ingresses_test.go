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
	netv1 "k8s.io/api/networking/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func addNginxProxyTimeoutAnnotations(annotations map[string]string, timeout string) map[string]string {
	annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = timeout
	annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = timeout
	return annotations
}

var _ = Describe("Ingresses", func() {

	Describe("The forge.IngressSpec function", func() {
		var (
			spec netv1.IngressSpec
		)

		const (
			host            = "crownlabs.example.com"
			path            = "/path/to/ingress"
			certificateName = "crownlabs-certificate"
			serviceName     = "service"
			servicePort     = "port"
		)

		JustBeforeEach(func() {
			spec = forge.IngressSpec(host, path, certificateName, serviceName, servicePort)
		})

		When("Forging the ingress specifications", func() {
			It("Should configure the correct TLS rule", func() {
				Expect(spec.TLS).To(Equal([]netv1.IngressTLS{{Hosts: []string{host}, SecretName: certificateName}}))
			})

			It("Should configure a single ingress rule", func() {
				Expect(spec.Rules).To(HaveLen(1))
			})

			It("Should configure the correct host name", func() {
				Expect(spec.Rules[0].Host).To(BeIdenticalTo(host))
			})

			It("Should configure a single ingress path", func() {
				Expect(spec.Rules[0].HTTP.Paths).To(HaveLen(1))
			})

			It("Should configure the correct path", func() {
				Expect(spec.Rules[0].HTTP.Paths[0].Path).To(BeIdenticalTo("/path/to/ingress"))
			})

			It("Should configure the correct service backend", func() {
				Expect(spec.Rules[0].HTTP.Paths[0].Backend.Service).To(Equal(
					&netv1.IngressServiceBackend{
						Name: serviceName,
						Port: netv1.ServiceBackendPort{Name: servicePort},
					}))
			})
		})
	})

	Describe("The forge.IngressGUIAnnotations function", func() {

		type InstanceGUIAnnotationsCase struct {
			Annotations    map[string]string
			ExpectedOutput map[string]string
			Environment    clv1alpha2.Environment
		}

		When("EnvironmentType is ClassStandalone", func() {

			DescribeTable("Correctly populates the annotations set",
				func(c InstanceGUIAnnotationsCase) {
					Expect(forge.IngressGUIAnnotations(&c.Environment, c.Annotations)).To(Equal(c.ExpectedOutput))
				},
				Entry("When the input annotations map is nil and RewriteURL false", InstanceGUIAnnotationsCase{
					Annotations:    nil,
					Environment:    clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassStandalone, RewriteURL: false},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{}, "3600"),
				}),
				Entry("When the input annotations map is nil and RewriteURL true", InstanceGUIAnnotationsCase{
					Annotations: nil,
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassStandalone, RewriteURL: true},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/$2",
					}, "3600"),
				}),
				Entry("When the input labels map already contains the expected values and RewriteURL false", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassStandalone, RewriteURL: false},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
				}),
				Entry("When the input labels map contains only part of the expected values and RewriteURL false", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassStandalone, RewriteURL: false},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
				}),
				Entry("When the input labels map already contains the expected values and RewriteURL true", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/$2",
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassStandalone, RewriteURL: true},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/$2",
						"user/key": "user/value",
					}, "3600"),
				}),
				Entry("When the input labels map contains only part of the expected values and RewriteURL true", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassStandalone, RewriteURL: true},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/$2",
						"user/key": "user/value",
					}, "3600"),
				}),
			)
		})

		When("EnvironmentType is different from ClassStandalone", func() {
			DescribeTable("Correctly populates the annotations set",
				func(c InstanceGUIAnnotationsCase) {
					Expect(forge.IngressGUIAnnotations(&c.Environment, c.Annotations)).To(Equal(c.ExpectedOutput))
				},
				Entry("When the input annotations map is nil", InstanceGUIAnnotationsCase{
					Annotations: nil,
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassCloudVM},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/$1",
					}, "3600"),
				}),
				Entry("When the input labels map already contains the expected values", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassCloudVM},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/$1",
						"user/key": "user/value",
					}, "3600"),
				}),
				Entry("When the input labels map contains ClassCloudVM", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassCloudVM},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/$1",
						"user/key": "user/value",
					}, "3600"),
				}),
				Entry("When the input labels map contains ClassVM", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassVM},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/$1",
						"user/key": "user/value",
					}, "3600"),
				}),
				Entry("When the input labels map contains ClassLocalVM", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassLocalVM},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/$1",
						"user/key": "user/value",
					}, "3600"),
				}),
			)
		})
	})

	Describe("The forge.IngressAuthenticationAnnotations function", func() {
		const authURL = "crownlabs.example.com/auth"

		type IngressAuthenticationAnnotationsCase struct {
			Input          map[string]string
			ExpectedOutput map[string]string
		}

		DescribeTable("Correctly populates the annotations set",
			func(c IngressAuthenticationAnnotationsCase) {
				Expect(forge.IngressAuthenticationAnnotations(c.Input, authURL)).To(Equal(c.ExpectedOutput))
			},
			Entry("When the input annotations map is nil", IngressAuthenticationAnnotationsCase{
				Input: nil,
				ExpectedOutput: map[string]string{
					"nginx.ingress.kubernetes.io/auth-url":    authURL + "/auth",
					"nginx.ingress.kubernetes.io/auth-signin": authURL + "/start?rd=$escaped_request_uri",
				},
			}),
			Entry("When the input labels map already contains the expected values", IngressAuthenticationAnnotationsCase{
				Input: map[string]string{
					"nginx.ingress.kubernetes.io/auth-url":    authURL + "/auth",
					"nginx.ingress.kubernetes.io/auth-signin": authURL + "/start?rd=$escaped_request_uri",
					"user/key": "user/value",
				},
				ExpectedOutput: map[string]string{
					"nginx.ingress.kubernetes.io/auth-url":    authURL + "/auth",
					"nginx.ingress.kubernetes.io/auth-signin": authURL + "/start?rd=$escaped_request_uri",
					"user/key": "user/value",
				},
			}),
			Entry("When the input labels map contains only part of the expected values", IngressAuthenticationAnnotationsCase{
				Input: map[string]string{
					"nginx.ingress.kubernetes.io/auth-url": authURL + "/auth",
					"user/key":                             "user/value",
				},
				ExpectedOutput: map[string]string{
					"nginx.ingress.kubernetes.io/auth-url":    authURL + "/auth",
					"nginx.ingress.kubernetes.io/auth-signin": authURL + "/start?rd=$escaped_request_uri",
					"user/key": "user/value",
				},
			}),
		)
	})
})
