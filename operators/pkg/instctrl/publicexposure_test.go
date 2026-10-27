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

package instctrl_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/clcontext"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
)

// No setup function needed: all tests are pure unit tests with fake client and in-memory objects

var _ = Describe("PublicExposure", func() {
	var (
		ctx        context.Context
		fakeClient *fake.ClientBuilder
		reconciler *instctrl.InstanceReconciler
		template   *clv1alpha2.Template
		instance   *clv1alpha2.Instance
	)

	BeforeEach(func() {
		_ = clv1alpha2.AddToScheme(scheme.Scheme)
		ctx = context.Background()
		fakeClient = fake.NewClientBuilder()
		reconciler = &instctrl.InstanceReconciler{
			Client:         fakeClient.Build(),
			Scheme:         scheme.Scheme,
			EventsRecorder: record.NewFakeRecorder(1024),
			PublicExposureOpts: forge.PublicExposureOpts{
				IPPool: []string{"192.168.100.1"},
				CommonAnnotations: map[string]string{
					"metallb.universe.tf/allow-shared-ip": "public-exposure",
					"metallb.universe.tf/address-pool":    "public",
				},
				LoadBalancerIPsKey: "metallb.universe.tf/loadBalancerIPs",
			},
		}
		template = &clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: "test-template", Namespace: "test-ns"},
			Spec: clv1alpha2.TemplateSpec{
				AllowPublicExposure: true,
				EnvironmentList: []clv1alpha2.Environment{
					{
						Name:            "main",
						Image:           "test-image:latest",
						EnvironmentType: clv1alpha2.ClassContainer,
						GuiEnabled:      true,
					},
				},
			},
		}
		instance = &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: "test-instance", Namespace: "test-ns"},
			Spec: clv1alpha2.InstanceSpec{
				Running: true,
				PublicExposure: &clv1alpha2.InstancePublicExposure{
					Ports: []clv1alpha2.PublicServicePort{{Name: "http", Port: 8080, TargetPort: 80}},
				},
			},
		}
	})

	It("Should create LoadBalancer service when conditions are met (unit)", func() {
		ctx1, _ := clctx.InstanceInto(ctx, instance)
		ctx1, _ = clctx.TemplateInto(ctx1, template)
		err := reconciler.EnforcePublicExposure(ctx1)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should update instance status correctly (unit)", func() {
		ctx1, _ := clctx.InstanceInto(ctx, instance)
		ctx1, _ = clctx.TemplateInto(ctx1, template)
		err := reconciler.EnforcePublicExposure(ctx1)
		Expect(err).ToNot(HaveOccurred())
		Expect(instance.Status.PublicExposure).ToNot(BeNil())
	})

	It("Should fail with duplicate ports in public exposure", func() {
		instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
			{Name: "ssh", Port: 30022, TargetPort: 22},
			{Name: "web", Port: 30022, TargetPort: 80},
		}
		err := instctrl.ValidatePublicExposureRequest(instance.Spec.PublicExposure.Ports)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("duplicate requested port"))
	})

	It("Should fail with duplicate targetPorts in public exposure", func() {
		instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
			{Name: "ssh", Port: 30022, TargetPort: 22},
			{Name: "web", Port: 30023, TargetPort: 22},
		}
		err := instctrl.ValidatePublicExposureRequest(instance.Spec.PublicExposure.Ports)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("duplicate desired targetPort"))
	})

	It("Should allow auto-assigned port (Port=0)", func() {
		instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
			{Name: "ssh", Port: 0, TargetPort: 22},
			{Name: "web", Port: 30023, TargetPort: 80},
		}
		err := instctrl.ValidatePublicExposureRequest(instance.Spec.PublicExposure.Ports)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should allow different protocols for different ports", func() {
		instance.Spec.PublicExposure.Ports = []clv1alpha2.PublicServicePort{
			{Name: "ssh", Port: 30022, TargetPort: 22, Protocol: "TCP"},
			{Name: "web", Port: 30023, TargetPort: 80, Protocol: "UDP"},
		}
		err := instctrl.ValidatePublicExposureRequest(instance.Spec.PublicExposure.Ports)
		Expect(err).ToNot(HaveOccurred())
	})

	When("The LoadBalancer service already has the requested IP and ports", func() {
		const currentIP = "192.168.100.2"
		var existing *corev1.Service

		BeforeEach(func() {
			reconciler.PublicExposureOpts.IPPool = []string{"192.168.100.1", currentIP}
			reconciler.PublicExposureOpts.CommonLabels = map[string]string{"lb-ipam.cilium.io/pool": "public"}

			// The CRD defaults the protocol to TCP, which is also the one reported by the service.
			instance.Spec.PublicExposure.Ports[0].Protocol = corev1.ProtocolTCP
			instance.Status.PublicExposure = &clv1alpha2.InstancePublicExposureStatus{
				Ports: instance.Spec.PublicExposure.Ports,
			}

			// A service created before the current common labels and annotations were configured.
			existing = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      forge.LoadBalancerServiceName(instance),
					Namespace: instance.Namespace,
					Labels:    map[string]string{"example.com/unmanaged": "value"},
					Annotations: map[string]string{
						"metallb.universe.tf/loadBalancerIPs": currentIP,
						"example.com/unmanaged":               "value",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
					Ports: []corev1.ServicePort{{
						Name: "http", Port: 8080, TargetPort: intstr.FromInt32(80), Protocol: corev1.ProtocolTCP,
					}},
				},
			}
			reconciler.Client = fake.NewClientBuilder().WithObjects(existing).Build()
		})

		It("Should keep the IP and ports and enforce the common labels and annotations", func() {
			ctx1, _ := clctx.InstanceInto(ctx, instance)
			ctx1, _ = clctx.TemplateInto(ctx1, template)
			Expect(reconciler.EnforcePublicExposure(ctx1)).To(Succeed())

			var service corev1.Service
			Expect(reconciler.Get(ctx, client.ObjectKeyFromObject(existing), &service)).To(Succeed())

			Expect(service.Annotations).To(HaveKeyWithValue("metallb.universe.tf/loadBalancerIPs", currentIP))
			Expect(service.Annotations).To(HaveKeyWithValue("metallb.universe.tf/allow-shared-ip", "public-exposure"))
			Expect(service.Annotations).To(HaveKeyWithValue("metallb.universe.tf/address-pool", "public"))
			Expect(service.Annotations).To(HaveKeyWithValue("example.com/unmanaged", "value"))

			Expect(service.Labels).To(HaveKeyWithValue(forge.LabelComponentKey, forge.LabelPublicExposureValue))
			Expect(service.Labels).To(HaveKeyWithValue("lb-ipam.cilium.io/pool", "public"))
			Expect(service.Labels).To(HaveKeyWithValue("example.com/unmanaged", "value"))

			Expect(service.Spec.Ports).To(Equal(existing.Spec.Ports))
			Expect(instance.Status.PublicExposure.Ports).To(ConsistOf(instance.Spec.PublicExposure.Ports))
		})
	})

	It("Should not define Status if public exposure in Spec is not set", func() {
		instance.Spec.PublicExposure = nil
		ctx1, _ := clctx.InstanceInto(ctx, instance)
		ctx1, _ = clctx.TemplateInto(ctx1, template)
		err := reconciler.EnforcePublicExposure(ctx1)
		Expect(err).ToNot(HaveOccurred())
		Expect(instance.Status.PublicExposure).To(BeNil())

		// No panic, no error, and no service created
	})
})
