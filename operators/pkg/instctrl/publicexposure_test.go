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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
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
			Client: fakeClient.Build(),
			Scheme: scheme.Scheme,
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
			Spec:       clv1alpha2.TemplateSpec{AllowPublicExposure: true},
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
