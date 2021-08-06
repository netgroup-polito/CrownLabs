// Copyright 2020-2021 Politecnico di Torino
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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

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
				Expect(spec.Rules[0].HTTP.Paths[0].Path).To(BeIdenticalTo("/path/to/ingress(/|$)(.*)"))
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
		const path = "/path/to/ingress"

		type InstanceGUIAnnotationsCase struct {
			Input          map[string]string
			ExpectedOutput map[string]string
		}

		DescribeTable("Correctly populates the annotations set",
			func(c InstanceGUIAnnotationsCase) {
				Expect(forge.IngressGUIAnnotations(c.Input, path)).To(Equal(c.ExpectedOutput))
			},
			Entry("When the input annotations map is nil", InstanceGUIAnnotationsCase{
				Input: nil,
				ExpectedOutput: map[string]string{
					"nginx.ingress.kubernetes.io/rewrite-target":        "/$2",
					"nginx.ingress.kubernetes.io/proxy-read-timeout":    "3600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout":    "3600",
					"nginx.ingress.kubernetes.io/configuration-snippet": `sub_filter '<head>' '<head> <base href="https://$host/` + path + `/index.html">';`,
				},
			}),
			Entry("When the input labels map already contains the expected values", InstanceGUIAnnotationsCase{
				Input: map[string]string{
					"nginx.ingress.kubernetes.io/rewrite-target":        "/$2",
					"nginx.ingress.kubernetes.io/proxy-read-timeout":    "3600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout":    "3600",
					"nginx.ingress.kubernetes.io/configuration-snippet": `sub_filter '<head>' '<head> <base href="https://$host/` + path + `/index.html">';`,
					"user/key": "user/value",
				},
				ExpectedOutput: map[string]string{
					"nginx.ingress.kubernetes.io/rewrite-target":        "/$2",
					"nginx.ingress.kubernetes.io/proxy-read-timeout":    "3600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout":    "3600",
					"nginx.ingress.kubernetes.io/configuration-snippet": `sub_filter '<head>' '<head> <base href="https://$host/` + path + `/index.html">';`,
					"user/key": "user/value",
				},
			}),
			Entry("When the input labels map contains only part of the expected values", InstanceGUIAnnotationsCase{
				Input: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-read-timeout": "3600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout": "3600",
					"user/key": "user/value",
				},
				ExpectedOutput: map[string]string{
					"nginx.ingress.kubernetes.io/rewrite-target":        "/$2",
					"nginx.ingress.kubernetes.io/proxy-read-timeout":    "3600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout":    "3600",
					"nginx.ingress.kubernetes.io/configuration-snippet": `sub_filter '<head>' '<head> <base href="https://$host/` + path + `/index.html">';`,
					"user/key": "user/value",
				},
			}),
		)
	})

	Describe("The forge.IngressMyDriveAnnotations function", func() {
		type InstanceMyDriveAnnotationsCase struct {
			Input          map[string]string
			ExpectedOutput map[string]string
		}

		DescribeTable("Correctly populates the annotations set",
			func(c InstanceMyDriveAnnotationsCase) {
				Expect(forge.IngressMyDriveAnnotations(c.Input)).To(Equal(c.ExpectedOutput))
			},
			Entry("When the input annotations map is nil", InstanceMyDriveAnnotationsCase{
				Input: nil,
				ExpectedOutput: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
					"nginx.ingress.kubernetes.io/proxy-read-timeout":       "600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout":       "600",
				},
			}),
			Entry("When the input labels map already contains the expected values", InstanceMyDriveAnnotationsCase{
				Input: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
					"nginx.ingress.kubernetes.io/proxy-read-timeout":       "600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout":       "600",
					"user/key": "user/value",
				},
				ExpectedOutput: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
					"nginx.ingress.kubernetes.io/proxy-read-timeout":       "600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout":       "600",
					"user/key": "user/value",
				},
			}),
			Entry("When the input labels map contains only part of the expected values", InstanceMyDriveAnnotationsCase{
				Input: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
					"nginx.ingress.kubernetes.io/proxy-read-timeout":       "600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout":       "600",
					"user/key": "user/value",
				},
				ExpectedOutput: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
					"nginx.ingress.kubernetes.io/proxy-read-timeout":       "600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout":       "600",
					"user/key": "user/value",
				},
			}),
		)
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

	Describe("The forge.Ingress*Path functions", func() {
		var (
			instance clv1alpha2.Instance
			path     string
		)

		const (
			instanceName      = "kubernetes-0000"
			instanceNamespace = "tenant-tester"
			instanceUID       = "dcc6ead1-0040-451b-ba68-787ebfb68640"
		)

		BeforeEach(func() {
			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace, UID: instanceUID},
			}
		})

		Describe("The forge.IngressGUIPath function", func() {
			JustBeforeEach(func() {
				path = forge.IngressGUIPath(&instance)
			})

			Context("The instance has no special configurations", func() {
				It("Should generate a path based on the instance UID", func() {
					Expect(path).To(BeIdenticalTo("/" + instanceUID))
				})
			})
		})

		Describe("The forge.IngressMyDrivePath function", func() {
			JustBeforeEach(func() {
				path = forge.IngressMyDrivePath(&instance)
			})

			Context("The instance has no special configurations", func() {
				It("Should generate a path based on the instance UID", func() {
					Expect(path).To(BeIdenticalTo("/" + instanceUID + "/mydrive"))
				})
			})
		})
	})
})
