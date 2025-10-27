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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("LoadBalancers forging", func() {

	const (
		instanceName      = "kubernetes-0000"
		instanceNamespace = "tenant-tester"
		templateName      = "kubernetes"
		tenantName        = "tester"
		externalIP        = "172.18.0.240"
		serviceName       = "test-service"
		portName          = "http"
		port              = int32(8080)
		targetPort        = int32(80)
	)

	var instance clv1alpha2.Instance

	BeforeEach(func() {
		instance = clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: instanceNamespace,
			},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: templateName},
				Tenant:   clv1alpha2.GenericRef{Name: tenantName},
			},
		}
	})

	Describe("The forge.LoadBalancerServiceSpec function", func() {
		var (
			spec  v1.ServiceSpec
			ports []clv1alpha2.PublicServicePort
		)

		BeforeEach(func() {
			ports = []clv1alpha2.PublicServicePort{
				{
					Name:       portName,
					Port:       port,
					TargetPort: targetPort,
				},
			}
		})

		JustBeforeEach(func() {
			spec = forge.LoadBalancerServiceSpec(&instance, ports)
		})

		When("Forging the LoadBalancer service spec", func() {
			It("Should set the correct service type", func() {
				Expect(spec.Type).To(Equal(v1.ServiceTypeLoadBalancer))
			})

			It("Should configure the correct selector", func() {
				expectedSelector := forge.InstanceSelectorLabels(&instance)
				Expect(spec.Selector).To(Equal(expectedSelector))
			})

			It("Should configure the correct number of ports", func() {
				Expect(spec.Ports).To(HaveLen(1))
			})

			It("Should configure the correct port details", func() {
				actual := spec.Ports[0]
				expectedPort := v1.ServicePort{
					Name:       portName,
					Port:       port,
					TargetPort: intstr.FromInt32(targetPort),
					Protocol:   v1.ProtocolTCP,
				}
				if actual.Protocol == "" {
					expectedPort.Protocol = ""
				}
				Expect(actual).To(Equal(expectedPort))
			})

			It("Should configure explicit TCP protocol", func() {
				ports[0].Protocol = v1.ProtocolTCP
				spec = forge.LoadBalancerServiceSpec(&instance, ports)
				Expect(spec.Ports[0].Protocol).To(Equal(v1.ProtocolTCP))
			})

			It("Should configure explicit UDP protocol", func() {
				ports[0].Protocol = v1.ProtocolUDP
				spec = forge.LoadBalancerServiceSpec(&instance, ports)
				Expect(spec.Ports[0].Protocol).To(Equal(v1.ProtocolUDP))
			})

			It("Should default to TCP when protocol is empty", func() {
				ports[0].Protocol = ""
				spec = forge.LoadBalancerServiceSpec(&instance, ports)
				Expect(spec.Ports[0].Protocol).To(Equal(v1.ProtocolTCP))
			})
		})

		When("Multiple ports are specified", func() {
			BeforeEach(func() {
				ports = []clv1alpha2.PublicServicePort{
					{
						Name:       "http",
						Port:       8080,
						TargetPort: 80,
					},
					{
						Name:       "https",
						Port:       8443,
						TargetPort: 443,
					},
				}
			})

			It("Should configure all ports correctly", func() {
				Expect(spec.Ports).To(HaveLen(2))

				actual0 := spec.Ports[0]
				expectedPort0 := v1.ServicePort{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt32(80),
					Protocol:   v1.ProtocolTCP,
				}
				if actual0.Protocol == "" {
					expectedPort0.Protocol = ""
				}
				Expect(actual0).To(Equal(expectedPort0))

				actual1 := spec.Ports[1]
				expectedPort1 := v1.ServicePort{
					Name:       "https",
					Port:       8443,
					TargetPort: intstr.FromInt32(443),
					Protocol:   v1.ProtocolTCP,
				}
				if actual1.Protocol == "" {
					expectedPort1.Protocol = ""
				}
				Expect(actual1).To(Equal(expectedPort1))
			})
		})
	})

	Describe("The forge.LoadBalancerServiceAnnotations function", func() {
		var (
			annotations map[string]string
			opts        forge.PublicExposureOpts
		)

		BeforeEach(func() {
			opts = forge.PublicExposureOpts{
				CommonAnnotations: map[string]string{
					"metallb.universe.tf/shared-ip":    "public-exposure",
					"metallb.universe.tf/address-pool": "public",
				},
				LoadBalancerIPsKey: "metallb.universe.tf/loadBalancerIPs",
			}
		})

		JustBeforeEach(func() {
			annotations = forge.LoadBalancerServiceAnnotations(externalIP, &opts)
		})

		When("Forging the LoadBalancer service annotations", func() {
			It("Should set the correct MetalLB address pool annotation", func() {
				Expect(annotations).To(HaveKeyWithValue(
					"metallb.universe.tf/address-pool",
					"public",
				))
			})

			It("Should set the correct MetalLB allow shared IP annotation", func() {
				Expect(annotations).To(HaveKeyWithValue(
					"metallb.universe.tf/shared-ip",
					"public-exposure",
				))
			})

			It("Should set the correct MetalLB loadBalancerIPs annotation", func() {
				Expect(annotations).To(HaveKeyWithValue(
					"metallb.universe.tf/loadBalancerIPs",
					externalIP,
				))
			})

			It("Should contain exactly 3 annotations", func() {
				Expect(annotations).To(HaveLen(3))
			})
		})

		When("Different external IP is provided", func() {
			const differentIP = "172.18.0.241"

			JustBeforeEach(func() {
				annotations = forge.LoadBalancerServiceAnnotations(differentIP, &opts)
			})

			It("Should use the provided external IP", func() {
				Expect(annotations).To(HaveKeyWithValue(
					"metallb.universe.tf/loadBalancerIPs",
					differentIP,
				))
			})
		})

		When("Custom annotation keys are used", func() {
			BeforeEach(func() {
				opts = forge.PublicExposureOpts{
					CommonAnnotations: map[string]string{
						"lbipam.cilium.io/sharing-key": "public-exposure",
					},
					LoadBalancerIPsKey: "lbipam.cilium.io/ips",
				}
			})

			It("Should use the custom LoadBalancer IPs key", func() {
				Expect(annotations).To(HaveKeyWithValue(
					"lbipam.cilium.io/ips",
					externalIP,
				))
			})

			It("Should include common annotations", func() {
				Expect(annotations).To(HaveKeyWithValue(
					"lbipam.cilium.io/sharing-key",
					"public-exposure",
				))
			})
		})
	})

	Describe("The forge.LoadBalancerServiceLabels function", func() {
		var labels map[string]string

		JustBeforeEach(func() {
			labels = forge.LoadBalancerServiceLabels()
		})

		When("Forging the LoadBalancer service labels", func() {
			It("Should set the correct component label", func() {
				Expect(labels).To(HaveKeyWithValue(
					"crownlabs.polito.it/component",
					"pe",
				))
			})

			It("Should contain exactly 1 label", func() {
				Expect(labels).To(HaveLen(1))
			})
		})
	})

	Describe("The forge.LoadBalancerServiceName function", func() {
		var serviceName string

		JustBeforeEach(func() {
			serviceName = forge.LoadBalancerServiceName(&instance)
		})

		When("Forging the LoadBalancer service name", func() {
			It("Should generate the correct service name", func() {
				expectedName := instanceName + "-pe"
				Expect(serviceName).To(Equal(expectedName))
			})
		})

		When("Different instance name is provided", func() {
			const differentInstanceName = "nginx-1234"

			BeforeEach(func() {
				instance.Name = differentInstanceName
			})

			It("Should use the instance name in the service name", func() {
				expectedName := differentInstanceName + "-pe"
				Expect(serviceName).To(Equal(expectedName))
			})
		})
	})
})
