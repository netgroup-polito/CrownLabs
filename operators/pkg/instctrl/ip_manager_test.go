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
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
)

var _ = Describe("IPManager", func() {
	var (
		reconciler *instctrl.InstanceReconciler
		instance   *clv1alpha2.Instance
		ctx        context.Context
	)

	BeforeEach(func() {
		reconciler = &instctrl.InstanceReconciler{
			PublicExposureOpts: forge.PublicExposureOpts{
				IPPool: []string{"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243"},
			},
		}
		instance = &clv1alpha2.Instance{
			Spec: clv1alpha2.InstanceSpec{
				PublicExposure: &clv1alpha2.InstancePublicExposure{
					Ports: []clv1alpha2.PublicServicePort{
						{Name: "http", Port: 8080, TargetPort: 80},
					},
				},
			},
		}
		ctx = context.Background()
	})

	It("BuildPrioritizedIPPool prioritizes used IPs over unused ones", func() {
		fullPool := []string{"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243"}
		usedPortsByIP := map[string]map[int32]bool{
			"172.18.0.242": {8080: true},
			"172.18.0.240": {9090: true},
		}
		prioritizedPool := reconciler.BuildPrioritizedIPPool(fullPool, usedPortsByIP)
		expected := []string{"172.18.0.240", "172.18.0.242", "172.18.0.241", "172.18.0.243"}
		Expect(prioritizedPool).To(Equal(expected))
	})

	It("BuildPrioritizedIPPool returns the full pool if no IPs are used", func() {
		fullPool := []string{"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243"}
		usedPortsByIP := map[string]map[int32]bool{}
		prioritizedPool := reconciler.BuildPrioritizedIPPool(fullPool, usedPortsByIP)
		expected := []string{"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243"}
		Expect(prioritizedPool).To(Equal(expected))
	})

	It("FindBestIPAndAssignPorts finds available IP and assigns specified ports", func() {
		usedPortsByIP := map[string]map[int32]bool{}
		ip, ports, err := reconciler.FindBestIPAndAssignPorts(ctx, instance, usedPortsByIP, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(ip).ToNot(BeEmpty())
		Expect(ports).To(HaveLen(1))
		Expect(ports[0].Port).To(Equal(int32(8080)))
	})

	It("FindBestIPAndAssignPorts skips IPs with conflicting ports", func() {
		usedPortsByIP := map[string]map[int32]bool{
			"172.18.0.240": {8080: true},
		}
		ip, _, err := reconciler.FindBestIPAndAssignPorts(ctx, instance, usedPortsByIP, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(ip).To(Equal("172.18.0.241"))
	})

	It("FindBestIPAndAssignPorts returns error when no IP can support all ports", func() {
		usedPortsByIP := map[string]map[int32]bool{
			"172.18.0.240": {8080: true},
			"172.18.0.241": {8080: true},
			"172.18.0.242": {8080: true},
			"172.18.0.243": {8080: true},
		}
		_, _, err := reconciler.FindBestIPAndAssignPorts(ctx, instance, usedPortsByIP, "")
		Expect(err).To(HaveOccurred())
	})

	It("UpdateUsedPortsByIP scans and returns used ports by IP", func() {
		svc1 := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service-1",
				Namespace: "test-ns",
				Labels:    forge.LoadBalancerServiceLabels(),
				Annotations: map[string]string{
					"metallb.universe.tf/loadBalancerIPs": "172.18.0.240",
				},
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeLoadBalancer,
				Ports: []v1.ServicePort{
					{Name: "http", Port: 8080, TargetPort: intstr.FromInt(80)},
					{Name: "https", Port: 8443, TargetPort: intstr.FromInt(443)},
				},
			},
		}
		svc2 := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service-2",
				Namespace: "test-ns",
				Labels:    forge.LoadBalancerServiceLabels(),
				Annotations: map[string]string{
					"metallb.universe.tf/loadBalancerIPs": "172.18.0.241",
				},
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeLoadBalancer,
				Ports: []v1.ServicePort{
					{Name: "api", Port: 9090, TargetPort: intstr.FromInt(90)},
				},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(svc1, svc2).Build()
		opts := &forge.PublicExposureOpts{
			CommonAnnotations: map[string]string{
				"metallb.universe.tf/shared-ip":    "public-exposure",
				"metallb.universe.tf/address-pool": "public",
			},
			LoadBalancerIPsKey: "metallb.universe.tf/loadBalancerIPs",
		}
		usedPortsByIP, err := instctrl.UpdateUsedPortsByIP(ctx, fakeClient, "", "", opts)
		Expect(err).ToNot(HaveOccurred())
		Expect(usedPortsByIP).To(HaveKey("172.18.0.240"))
		Expect(usedPortsByIP["172.18.0.240"]).To(HaveKey(int32(8080)))
		Expect(usedPortsByIP["172.18.0.240"]).To(HaveKey(int32(8443)))
		Expect(usedPortsByIP).To(HaveKey("172.18.0.241"))
		Expect(usedPortsByIP["172.18.0.241"]).To(HaveKey(int32(9090)))
	})

	It("UpdateUsedPortsByIP excludes specified service", func() {
		uniquePort := int32(9999)
		uniqueIP := "172.18.0.243"
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "exclude-service",
				Namespace: "test-ns",
				Labels:    forge.LoadBalancerServiceLabels(),
				Annotations: map[string]string{
					"metallb.universe.tf/loadBalancerIPs": uniqueIP,
				},
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeLoadBalancer,
				Ports: []v1.ServicePort{
					{Name: "unique-port", Port: uniquePort, TargetPort: intstr.FromInt(80)},
				},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(svc).Build()
		opts := &forge.PublicExposureOpts{
			CommonAnnotations: map[string]string{
				"metallb.universe.tf/shared-ip":    "public-exposure",
				"metallb.universe.tf/address-pool": "public",
			},
			LoadBalancerIPsKey: "metallb.universe.tf/loadBalancerIPs",
		}
		usedPortsByIPWithService, err := instctrl.UpdateUsedPortsByIP(ctx, fakeClient, "", "", opts)
		Expect(err).ToNot(HaveOccurred())
		Expect(usedPortsByIPWithService).To(HaveKey(uniqueIP))
		Expect(usedPortsByIPWithService[uniqueIP]).To(HaveKey(uniquePort))
		usedPortsByIPWithoutService, err := instctrl.UpdateUsedPortsByIP(ctx, fakeClient, "exclude-service", "test-ns", opts)
		Expect(err).ToNot(HaveOccurred())
		Expect(usedPortsByIPWithoutService).ToNot(HaveKey(uniqueIP))
	})
})
