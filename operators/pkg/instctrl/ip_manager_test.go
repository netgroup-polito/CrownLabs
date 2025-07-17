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

package instctrl_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
)

var _ = Describe("IP Manager Functions", func() {
	ctx := context.Background()

	var (
		reconciler *instctrl.InstanceReconciler
		instance   *clv1alpha2.Instance
		namespace  string
	)

	BeforeEach(func() {
		namespace = "test-namespace"
		reconciler = &instctrl.InstanceReconciler{
			Client:               k8sClient,
			PublicExposureIPPool: []string{"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243"},
		}

		instance = &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-instance",
				Namespace: namespace,
			},
			Spec: clv1alpha2.InstanceSpec{
				PublicExposure: &clv1alpha2.InstancePublicExposure{
					Ports: []clv1alpha2.PublicServicePort{
						{
							Name:       "http",
							Port:       8080,
							TargetPort: 80,
						},
					},
				},
			},
		}

		// Create namespace
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed())
	})

	Describe("GetMetalLBIPPool", func() {
		It("Should return the configured IP pool", func() {
			ipPool, err := reconciler.GetMetalLBIPPool(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(ipPool).To(Equal([]string{"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243"}))
		})

		It("Should return empty pool when none configured", func() {
			reconciler.PublicExposureIPPool = []string{}
			ipPool, err := reconciler.GetMetalLBIPPool(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(ipPool).To(BeEmpty())
		})
	})

	Describe("BuildPrioritizedIPPool", func() {
		It("Should prioritize used IPs over unused ones", func() {
			fullPool := []string{"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243"}
			usedPortsByIP := map[string]map[int32]bool{
				"172.18.0.242": {8080: true},
				"172.18.0.240": {9090: true},
			}

			prioritizedPool := reconciler.BuildPrioritizedIPPool(fullPool, usedPortsByIP)

			// Used IPs should come first, sorted alphabetically
			Expect(prioritizedPool).To(HaveLen(4))
			Expect(prioritizedPool[0]).To(Equal("172.18.0.240"))
			Expect(prioritizedPool[1]).To(Equal("172.18.0.242"))
			// Unused IPs should come after, sorted alphabetically
			Expect(prioritizedPool[2]).To(Equal("172.18.0.241"))
			Expect(prioritizedPool[3]).To(Equal("172.18.0.243"))
		})

		It("Should handle empty used ports map", func() {
			fullPool := []string{"172.18.0.240", "172.18.0.241"}
			usedPortsByIP := map[string]map[int32]bool{}

			prioritizedPool := reconciler.BuildPrioritizedIPPool(fullPool, usedPortsByIP)

			Expect(prioritizedPool).To(Equal([]string{"172.18.0.240", "172.18.0.241"}))
		})

		It("Should handle all IPs being used", func() {
			fullPool := []string{"172.18.0.240", "172.18.0.241"}
			usedPortsByIP := map[string]map[int32]bool{
				"172.18.0.240": {8080: true},
				"172.18.0.241": {9090: true},
			}

			prioritizedPool := reconciler.BuildPrioritizedIPPool(fullPool, usedPortsByIP)

			Expect(prioritizedPool).To(Equal([]string{"172.18.0.240", "172.18.0.241"}))
		})
	})

	Describe("FindBestIPAndAssignPorts", func() {
		Context("With specified ports", func() {
			It("Should find available IP and assign specified ports", func() {
				instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
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

				usedPortsByIP := map[string]map[int32]bool{}

				ip, assignedPorts, err := reconciler.FindBestIPAndAssignPorts(ctx, k8sClient, instance, usedPortsByIP)

				Expect(err).ToNot(HaveOccurred())
				Expect(ip).To(Equal("172.18.0.240")) // First IP in pool
				Expect(assignedPorts).To(HaveLen(2))
				Expect(assignedPorts[0].Port).To(Equal(int32(8080)))
				Expect(assignedPorts[1].Port).To(Equal(int32(8443)))
			})

			It("Should skip IPs with conflicting ports", func() {
				instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
					{
						Name:       "http",
						Port:       8080,
						TargetPort: 80,
					},
				}

				usedPortsByIP := map[string]map[int32]bool{
					"172.18.0.240": {8080: true}, // Port conflict
					"172.18.0.241": {9090: true}, // No conflict
				}

				ip, assignedPorts, err := reconciler.FindBestIPAndAssignPorts(ctx, k8sClient, instance, usedPortsByIP)

				Expect(err).ToNot(HaveOccurred())
				Expect(ip).To(Equal("172.18.0.241")) // Should skip first IP due to conflict
				Expect(assignedPorts).To(HaveLen(1))
				Expect(assignedPorts[0].Port).To(Equal(int32(8080)))
			})
		})

		Context("With automatic port assignment", func() {
			It("Should assign automatic ports starting from base port", func() {
				instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
					{
						Name:       "auto-port",
						Port:       0, // Request automatic assignment
						TargetPort: 80,
					},
				}

				usedPortsByIP := map[string]map[int32]bool{}

				ip, assignedPorts, err := reconciler.FindBestIPAndAssignPorts(ctx, k8sClient, instance, usedPortsByIP)

				Expect(err).ToNot(HaveOccurred())
				Expect(ip).To(Equal("172.18.0.240"))
				Expect(assignedPorts).To(HaveLen(1))
				Expect(assignedPorts[0].Port).To(Equal(int32(forge.BasePortForAutomaticAssignment)))
			})

			It("Should skip used automatic ports", func() {
				instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
					{
						Name:       "auto-port",
						Port:       0,
						TargetPort: 80,
					},
				}

				usedPortsByIP := map[string]map[int32]bool{
					"172.18.0.240": {int32(forge.BasePortForAutomaticAssignment): true},
				}

				ip, assignedPorts, err := reconciler.FindBestIPAndAssignPorts(ctx, k8sClient, instance, usedPortsByIP)

				Expect(err).ToNot(HaveOccurred())
				Expect(ip).To(Equal("172.18.0.240"))
				Expect(assignedPorts).To(HaveLen(1))
				Expect(assignedPorts[0].Port).To(Equal(int32(forge.BasePortForAutomaticAssignment + 1)))
			})
		})

		Context("With existing LoadBalancer service", func() {
			It("Should prefer existing service IP", func() {
				// Create an existing LoadBalancer service
				svc := &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      forge.LoadBalancerServiceName(instance),
						Namespace: namespace,
						Annotations: map[string]string{
							forge.MetallbLoadBalancerIPsAnnotation: "172.18.0.242",
						},
					},
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
						Ports: []v1.ServicePort{
							{
								Port:       9090,
								TargetPort: intstr.FromInt(90),
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, svc)).To(Succeed())

				instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
					{
						Name:       "http",
						Port:       8080,
						TargetPort: 80,
					},
				}

				usedPortsByIP := map[string]map[int32]bool{}

				ip, assignedPorts, err := reconciler.FindBestIPAndAssignPorts(ctx, k8sClient, instance, usedPortsByIP)

				Expect(err).ToNot(HaveOccurred())
				Expect(ip).To(Equal("172.18.0.242")) // Should prefer existing service IP
				Expect(assignedPorts).To(HaveLen(1))
				Expect(assignedPorts[0].Port).To(Equal(int32(8080)))
			})
		})

		Context("Error cases", func() {
			It("Should return error when no IP can support all ports", func() {
				instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
					{
						Name:       "http",
						Port:       8080,
						TargetPort: 80,
					},
				}

				// All IPs have the requested port occupied
				usedPortsByIP := map[string]map[int32]bool{
					"172.18.0.240": {8080: true},
					"172.18.0.241": {8080: true},
					"172.18.0.242": {8080: true},
					"172.18.0.243": {8080: true},
				}

				ip, assignedPorts, err := reconciler.FindBestIPAndAssignPorts(ctx, k8sClient, instance, usedPortsByIP)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no available IP can support all requested ports"))
				Expect(ip).To(BeEmpty())
				Expect(assignedPorts).To(BeNil())
			})
		})
	})

	Describe("UpdateUsedPortsByIP", func() {
		It("Should scan and return used ports by IP", func() {
			// Create LoadBalancer services
			svc1 := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-1",
					Namespace: namespace,
					Labels:    forge.LoadBalancerServiceLabels(),
					Annotations: map[string]string{
						forge.MetallbLoadBalancerIPsAnnotation: "172.18.0.240",
					},
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
					Ports: []v1.ServicePort{
						{Port: 8080, TargetPort: intstr.FromInt(80)},
						{Port: 8443, TargetPort: intstr.FromInt(443)},
					},
				},
			}

			svc2 := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-2",
					Namespace: namespace,
					Labels:    forge.LoadBalancerServiceLabels(),
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
					Ports: []v1.ServicePort{
						{Port: 9090, TargetPort: intstr.FromInt(90)},
					},
				},
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{IP: "172.18.0.241"},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, svc1)).To(Succeed())
			Expect(k8sClient.Create(ctx, svc2)).To(Succeed())

			usedPortsByIP, err := instctrl.UpdateUsedPortsByIP(ctx, k8sClient, "", "")

			Expect(err).ToNot(HaveOccurred())
			Expect(usedPortsByIP).To(HaveKey("172.18.0.240"))
			Expect(usedPortsByIP["172.18.0.240"]).To(HaveKey(int32(8080)))
			Expect(usedPortsByIP["172.18.0.240"]).To(HaveKey(int32(8443)))
			Expect(usedPortsByIP).To(HaveKey("172.18.0.241"))
			Expect(usedPortsByIP["172.18.0.241"]).To(HaveKey(int32(9090)))
		})

		It("Should exclude specified service", func() {
			// Create LoadBalancer service
			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "exclude-service",
					Namespace: namespace,
					Labels:    forge.LoadBalancerServiceLabels(),
					Annotations: map[string]string{
						forge.MetallbLoadBalancerIPsAnnotation: "172.18.0.240",
					},
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
					Ports: []v1.ServicePort{
						{Port: 8080, TargetPort: intstr.FromInt(80)},
					},
				},
			}

			Expect(k8sClient.Create(ctx, svc)).To(Succeed())

			usedPortsByIP, err := instctrl.UpdateUsedPortsByIP(ctx, k8sClient, "exclude-service", namespace)

			Expect(err).ToNot(HaveOccurred())
			Expect(usedPortsByIP).ToNot(HaveKey("172.18.0.240"))
		})

		It("Should ignore services without external IP", func() {
			// Create LoadBalancer service without IP
			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-ip-service",
					Namespace: namespace,
					Labels:    forge.LoadBalancerServiceLabels(),
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
					Ports: []v1.ServicePort{
						{Port: 8080, TargetPort: intstr.FromInt(80)},
					},
				},
			}

			Expect(k8sClient.Create(ctx, svc)).To(Succeed())

			usedPortsByIP, err := instctrl.UpdateUsedPortsByIP(ctx, k8sClient, "", "")

			Expect(err).ToNot(HaveOccurred())
			Expect(usedPortsByIP).To(BeEmpty())
		})

		It("Should ignore non-LoadBalancer services", func() {
			// Create ClusterIP service with public exposure labels
			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cluster-ip-service",
					Namespace: namespace,
					Labels:    forge.LoadBalancerServiceLabels(),
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeClusterIP,
					Ports: []v1.ServicePort{
						{Port: 8080, TargetPort: intstr.FromInt(80)},
					},
				},
			}

			Expect(k8sClient.Create(ctx, svc)).To(Succeed())

			usedPortsByIP, err := instctrl.UpdateUsedPortsByIP(ctx, k8sClient, "", "")

			Expect(err).ToNot(HaveOccurred())
			Expect(usedPortsByIP).To(BeEmpty())
		})
	})
})
