// Copyright 2020-2025 Politecnico di Torino
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

		When("EnvironmentType is not ClassStandalone", func() {
			DescribeTable("Correctly populates the annotations set",
				func(c InstanceGUIAnnotationsCase) {
					Expect(forge.IngressGUIAnnotations(&c.Environment, c.Annotations)).To(Equal(c.ExpectedOutput))
				},
				Entry("When the input annotations map is nil", InstanceGUIAnnotationsCase{
					Annotations:    nil,
					Environment:    clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassContainer},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{}, "3600"),
				}),
				Entry("When the input labels map already contains the expected values", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassContainer},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
				}),
				Entry("When the input labels map contains only part of the expected values", InstanceGUIAnnotationsCase{
					Annotations: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
					Environment: clv1alpha2.Environment{EnvironmentType: clv1alpha2.ClassContainer},
					ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
						"user/key": "user/value",
					}, "3600"),
				}),
			)
		})
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
				ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
				}, "600"),
			}),
			Entry("When the input labels map already contains the expected values", InstanceMyDriveAnnotationsCase{
				Input: addNginxProxyTimeoutAnnotations(map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
					"user/key": "user/value",
				}, "600"),
				ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
					"user/key": "user/value",
				}, "600"),
			}),
			Entry("When the input labels map contains only part of the expected values", InstanceMyDriveAnnotationsCase{
				Input: addNginxProxyTimeoutAnnotations(map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
					"user/key": "user/value",
				}, "600"),
				ExpectedOutput: addNginxProxyTimeoutAnnotations(map[string]string{
					"nginx.ingress.kubernetes.io/proxy-body-size":          "0",
					"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
					"user/key": "user/value",
				}, "600"),
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
			instance    clv1alpha2.Instance
			path        string
			statusPath  string
			GUIName     string
			environment clv1alpha2.Environment
		)

		const (
			instanceName      = "kubernetes-0000"
			instanceNamespace = "tenant-tester"
			instanceUID       = "dcc6ead1-0040-451b-ba68-787ebfb68640"
			host              = "crownlabs.example.com"
		)

		BeforeEach(func() {
			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace, UID: instanceUID},
			}
		})

		Describe("The forge.HostName function", func() {
			const baseHost = "crownlabs.it"
			type HostNameCase struct {
				Mode           clv1alpha2.EnvironmentMode
				ExpectedOutput string
			}
			DescribeTable("correctly generates the hostname",
				func(c HostNameCase) {
					Expect(forge.HostName(baseHost, c.Mode)).To(Equal(c.ExpectedOutput))
				},
				Entry("when the mode is Default", HostNameCase{
					Mode:           clv1alpha2.ModeStandard,
					ExpectedOutput: baseHost,
				}),
				Entry("when the mode is Exam", HostNameCase{
					Mode:           clv1alpha2.ModeExam,
					ExpectedOutput: "exam." + baseHost,
				}),
				Entry("when the mode is Exercise", HostNameCase{
					Mode:           clv1alpha2.ModeExercise,
					ExpectedOutput: "exercise." + baseHost,
				}),
				Entry("when the mode is invalid/unset", HostNameCase{
					Mode:           "",
					ExpectedOutput: baseHost,
				}),
			)
		})

		Describe("The forge.IngressGUIPath function", func() {
			JustBeforeEach(func() {
				path = forge.IngressGUIPath(&instance, &environment)
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
							Expect(path).To(BeIdenticalTo("/instance/" + instanceUID + "/app(/|$)(.*)"))
						})
					})
				})
				When("Rewrite is false", func() {
					BeforeEach(func() {
						environment.RewriteURL = false
					})
					Context("The instance has no special configurations", func() {
						It("Should generate a path based on the instance UID", func() {
							Expect(path).To(BeIdenticalTo("/instance/" + instanceUID + "/app"))
						})
					})
				})
			})

			When("EnvironmentType is not ClassStandalone", func() {
				BeforeEach(func() {
					environment.EnvironmentType = clv1alpha2.ClassContainer
				})
				Context("The instance has no special configurations", func() {
					It("Should generate a path based on the instance UID", func() {
						Expect(path).To(BeIdenticalTo("/instance/" + instanceUID + "/app"))
					})
				})
			})

		})

		Describe("The forge.IngressGuiStatusURL function", func() {
			JustBeforeEach(func() {
				statusPath = forge.IngressGuiStatusURL(host, &environment, &instance)
			})
			When("EnvironmentType is ClassStandalone", func() {
				BeforeEach(func() {
					environment.EnvironmentType = clv1alpha2.ClassStandalone
				})
				It("Should generate a path based on the instance UID and /app at the end", func() {
					Expect(statusPath).To(BeIdenticalTo("https://" + host + "/instance/" + instanceUID + "/app/"))
				})
			})
			When("EnvironmentType is not ClassStandalone", func() {
				BeforeEach(func() {
					environment.EnvironmentType = clv1alpha2.ClassContainer
				})
				It("Should generate a path based on the instance UID and /app at the end", func() {
					Expect(statusPath).To(BeIdenticalTo("https://" + host + "/instance/" + instanceUID + "/app/"))
				})
			})
		})

		Describe("The forge.IngressGUIName function", func() {
			JustBeforeEach(func() {
				GUIName = forge.IngressGUIName(&environment)
			})
			When("EnvironmentType is ClassStandalone", func() {
				BeforeEach(func() {
					environment.EnvironmentType = clv1alpha2.ClassStandalone
				})
				It("Should generate a path based on the instance UID and /app at the end", func() {
					Expect(GUIName).To(BeIdenticalTo("app"))
				})
			})
			When("EnvironmentType is not ClassStandalone", func() {
				BeforeEach(func() {
					environment.EnvironmentType = clv1alpha2.ClassContainer
				})
				It("Should generate a path based on the instance UID", func() {
					Expect(GUIName).To(BeIdenticalTo("gui"))
				})
			})
		})
	})
})
