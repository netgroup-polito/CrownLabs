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

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/clcontext"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
)

// FakeClientWrapped sets the ClusterIP on created Services to simulate the behavior
// of a real API server that assigns ClusterIPs.
type FakeClientWrapped struct {
	client.Client
	serviceClusterIP string
}

func (c FakeClientWrapped) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if svc, ok := obj.(*corev1.Service); ok {
		svc.Spec.ClusterIP = c.serviceClusterIP
	}
	return c.Client.Create(ctx, obj, opts...)
}

var _ = Describe("Exposition helpers", func() {
	var (
		ctx           context.Context
		clientBuilder fake.ClientBuilder
		reconciler    instctrl.InstanceReconciler

		instance    clv1alpha2.Instance
		environment clv1alpha2.Environment
		template    clv1alpha2.Template
		index       int

		serviceName   types.NamespacedName
		ingressName   types.NamespacedName
		httpRouteName types.NamespacedName

		clusterIP = "1.1.1.1"
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

		environment = clv1alpha2.Environment{Name: "control-plane", EnvironmentType: clv1alpha2.ClassContainer}
		template = clv1alpha2.Template{ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: "kubernetes"}, Spec: clv1alpha2.TemplateSpec{EnvironmentList: []clv1alpha2.Environment{environment}}}
		instance = clv1alpha2.Instance{ObjectMeta: metav1.ObjectMeta{Name: "kubernetes-0000", Namespace: "tenant-tester", UID: "dcc6ead1-0040-451b-ba68-787ebfb68640"}, Spec: clv1alpha2.InstanceSpec{Template: clv1alpha2.GenericRef{Name: "kubernetes", Namespace: "kubernetes"}, Tenant: clv1alpha2.GenericRef{Name: "tester"}}, Status: clv1alpha2.InstanceStatus{Environments: []clv1alpha2.InstanceStatusEnv{{}, {}, {}}}}
		index = 0

		serviceName = forge.NamespacedNameWithSuffix(&instance, environment.Name)
		ingressName = forge.NamespacedNameWithSuffix(&instance, environment.Name)
		httpRouteName = forge.NamespacedNameWithSuffix(&instance, environment.Name)
	})

	JustBeforeEach(func() {
		// Use suite-level reconciler as a base and customize per-test as needed.
		reconciler = instanceReconciler
		reconciler.Scheme = scheme.Scheme
		reconciler.EventsRecorder = record.NewFakeRecorder(1024)

		// Inject context values common to many tests
		ctx, _ = clctx.InstanceInto(ctx, &instance)
		ctx, _ = clctx.EnvironmentInto(ctx, &environment)
		ctx, _ = clctx.TemplateInto(ctx, &template)
		ctx = clctx.EnvironmentIndexInto(ctx, index)
	})

	Context("When instance is running", func() {
		It("creates Service and updates instance status with ClusterIP", func() {
			// Arrange
			instance.Spec.Running = true
			cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
			reconciler.Client = cl

			// Act
			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Assert
			svc := corev1.Service{}
			Expect(reconciler.Client.Get(ctx, serviceName, &svc)).To(Succeed())
			Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
		})

		Context("Gateway API mode is enabled", func() {
			It("creates HTTPRoute", func() {
				// Arrange
				reconciler.ExpositionConfig.GatewayAPIMode = true
				instance.Spec.Running = true
				cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
				reconciler.Client = cl

				// Act
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Assert
				// HTTPRoute exists and instance status updated
				httpRoute := gatewayv1.HTTPRoute{}
				Expect(reconciler.Client.Get(ctx, httpRouteName, &httpRoute)).To(Succeed())
				Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
				// Ingress should not exist and Service should exist
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
			})
			It("does not create HTTPRoute if it already exists", func() {
				// Arrange
				reconciler.ExpositionConfig.GatewayAPIMode = true
				instance.Spec.Running = true
				httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
				cl := FakeClientWrapped{Client: clientBuilder.WithObjects(&httpRoute).Build(), serviceClusterIP: clusterIP}
				reconciler.Client = cl

				// Act
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Assert
				Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
			})
		})
		It("does not create HTTPRoute for GUI-less VM", func() {
			// Arrange
			reconciler.ExpositionConfig.GatewayAPIMode = true
			instance.Spec.Running = true
			// make environment a VM without GUI and re-insert in context
			environment.EnvironmentType = clv1alpha2.ClassVM
			environment.GuiEnabled = false
			ctx, _ = clctx.EnvironmentInto(ctx, &environment)

			cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
			reconciler.Client = cl

			// Act
			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Assert
			Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
			Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
			Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
		})
		It("does not create Service or HTTPRoute if the environment index is out of range", func() {
			// Arrange
			reconciler.ExpositionConfig.GatewayAPIMode = true
			instance.Spec.Running = true
			index = 5 // out of current range
			ctx = clctx.EnvironmentIndexInto(ctx, index)
			cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
			reconciler.Client = cl

			// Act
			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Assert
			Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(HaveOccurred())
			Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
		})

		Context("Gateway API mode is disabled", func() {
			It("creates Ingress", func() {
				// Arrange
				reconciler.ExpositionConfig.GatewayAPIMode = false
				instance.Spec.Running = true
				cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
				reconciler.Client = cl

				// Act
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Assert
				// Ingress exists and instance status updated
				ingress := netv1.Ingress{}
				Expect(reconciler.Client.Get(ctx, ingressName, &ingress)).To(Succeed())
				Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
				// HTTPRoute should not exist and Service should exist
				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
			})
			It("does not create Ingress if it already exists", func() {
				// Arrange
				reconciler.ExpositionConfig.GatewayAPIMode = false
				instance.Spec.Running = true
				ingress := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
				cl := FakeClientWrapped{Client: clientBuilder.WithObjects(&ingress).Build(), serviceClusterIP: clusterIP}
				reconciler.Client = cl

				// Act
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Assert
				Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
			})
			It("if autentication is enabled, it creates Ingress with auth annotations", func() {
				// Arrange
				reconciler.ExpositionConfig.GatewayAPIMode = false
				reconciler.ExpositionConfig.EnableAuthentication = true
				instance.Spec.Running = true
				cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
				reconciler.Client = cl

				// Act
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Assert
				ingress := netv1.Ingress{}
				Expect(reconciler.Client.Get(ctx, ingressName, &ingress)).To(Succeed())
				Expect(ingress.Annotations).To(HaveKey("nginx.ingress.kubernetes.io/auth-url"))
				Expect(ingress.Annotations).To(HaveKey("nginx.ingress.kubernetes.io/auth-signin"))
			})
			It("if autentication is disabled, it creates Ingress without auth annotations", func() {
				// Arrange
				reconciler.ExpositionConfig.GatewayAPIMode = false
				reconciler.ExpositionConfig.EnableAuthentication = false
				instance.Spec.Running = true
				cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
				reconciler.Client = cl

				// Act
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Assert
				ingress := netv1.Ingress{}
				Expect(reconciler.Client.Get(ctx, ingressName, &ingress)).To(Succeed())
				Expect(ingress.Annotations).ToNot(HaveKey("nginx.ingress.kubernetes.io/auth-url"))
				Expect(ingress.Annotations).ToNot(HaveKey("nginx.ingress.kubernetes.io/auth-signin"))
			})
			It("does not create Ingress for GUI-less VM", func() {
				// Arrange
				reconciler.ExpositionConfig.GatewayAPIMode = false
				instance.Spec.Running = true
				// make environment a VM without GUI and re-insert in context
				environment.EnvironmentType = clv1alpha2.ClassVM
				environment.GuiEnabled = false
				ctx, _ = clctx.EnvironmentInto(ctx, &environment)

				cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
				reconciler.Client = cl

				// Act
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Assert: Service exists, Ingress does not
				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
				Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
			})
			It("does not create Service or Ingress if the environment index is out of range", func() {
				// Arrange
				reconciler.ExpositionConfig.GatewayAPIMode = false
				instance.Spec.Running = true
				index = 5 // out of current range
				ctx = clctx.EnvironmentIndexInto(ctx, index)
				cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
				reconciler.Client = cl

				// Act
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Assert: Service and Ingress do not exist, instance status extended but IP empty
				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(HaveOccurred())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
			})
		})
	})

	Context("When instance is not running", func() {
		It("extends Status.Environments and clears IP when index is out of range", func() {
			// Arrange
			instance.Spec.Running = false
			index = 5 // out of current range
			ctx = clctx.EnvironmentIndexInto(ctx, index)
			cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
			reconciler.Client = cl

			// Act
			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Assert: status extended and IP cleared
			Expect(len(instance.Status.Environments)).To(BeNumerically(">", index))
			Expect(instance.Status.Environments[index].IP).To(Equal(""))
		})
		It("removes Service, Ingress and HTTPRoute and clears instance status", func() {
			// Arrange
			instance.Spec.Running = false
			svc := corev1.Service{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
			ingress := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
			httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
			cl := FakeClientWrapped{Client: clientBuilder.WithObjects(&svc, &ingress, &httpRoute).Build(), serviceClusterIP: clusterIP}
			reconciler.Client = cl

			// Act
			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Assert
			Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(HaveOccurred())
			Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
			Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
			Expect(instance.Status.Environments[index].IP).To(Equal(""))
		})
		It("does not fail if Service, Ingress or HTTPRoute do not exist", func() {
			// Arrange
			instance.Spec.Running = false
			cl := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
			reconciler.Client = cl

			// Act
			err := reconciler.EnforceInstanceExposition(ctx)

			// Assert
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
