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
	"fmt"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
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

		reconciler = instanceReconciler
		reconciler.Scheme = scheme.Scheme
		reconciler.EventsRecorder = record.NewFakeRecorder(1024)

		// Default fake client for tests; individual tests may override with WithObjects.
		reconciler.Client = FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
	})

	JustBeforeEach(func() {
		// Inject context values common to many tests
		ctx, _ = clctx.InstanceInto(ctx, &instance)
		ctx, _ = clctx.EnvironmentInto(ctx, &environment)
		ctx, _ = clctx.TemplateInto(ctx, &template)
		ctx = clctx.EnvironmentIndexInto(ctx, index)
	})

	Context("When instance is running", func() {
		BeforeEach(func() { instance.Spec.Running = true })

		It("creates Service and updates instance status with ClusterIP", func() {
			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			svc := corev1.Service{}
			Expect(reconciler.Client.Get(ctx, serviceName, &svc)).To(Succeed())
			Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
		})

		It("updates instance status with empty string when Service is headless and no Pod exists", func() {
			reconciler.Client = FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: "None"}

			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			svc := corev1.Service{}
			Expect(reconciler.Client.Get(ctx, serviceName, &svc)).To(Succeed())
			Expect(instance.Status.Environments[index].IP).To(Equal(""))
		})

		It("updates instance status with Pod IP when Service is headless and Pod exists (Container)", func() {
			// Mock pod for container environment
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-xxxxx", serviceName.Name),
					Namespace: serviceName.Namespace,
					Labels:    forge.EnvironmentSelectorLabels(&instance, &environment),
				},
				Status: corev1.PodStatus{
					PodIP: "10.244.50.60",
				},
			}
			reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects(pod).Build(), serviceClusterIP: "None"}

			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(instance.Status.Environments[index].IP).To(Equal("10.244.50.60"))
		})

		It("updates instance status with Pod IP when Service is headless and VM pod exists", func() {
			// Mock environment as VM
			vmEnv := clv1alpha2.Environment{Name: "vm-env", EnvironmentType: clv1alpha2.ClassVM}
			vmInstance := instance
			vmInstance.Status.Environments = []clv1alpha2.InstanceStatusEnv{{}}
			vmServiceName := forge.NamespacedNameWithSuffix(&vmInstance, vmEnv.Name)

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("virt-launcher-%s-xxxxx", vmServiceName.Name),
					Namespace: vmServiceName.Namespace,
					Labels:    forge.EnvironmentSelectorLabels(&vmInstance, &vmEnv),
				},
				Status: corev1.PodStatus{
					PodIP: "10.244.100.200",
				},
			}
			reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects(pod).Build(), serviceClusterIP: "None"}

			vmCtx := ctrl.LoggerInto(context.Background(), logr.Discard())
			vmCtx, _ = clctx.InstanceInto(vmCtx, &vmInstance)
			vmCtx, _ = clctx.EnvironmentInto(vmCtx, &vmEnv)
			vmCtx, _ = clctx.TemplateInto(vmCtx, &template)
			vmCtx = clctx.EnvironmentIndexInto(vmCtx, 0)

			err := reconciler.EnforceInstanceExposition(vmCtx)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmInstance.Status.Environments[0].IP).To(Equal("10.244.100.200"))
		})

		Context("Gateway API mode enabled", func() {
			BeforeEach(func() { reconciler.ExpositionConfig.GatewayAPIMode = true })

			It("creates HTTPRoute", func() {
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(Succeed())
				Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
			})

			It("returns without creating resources if env index out of range", func() {
				index = 10
				ctx = clctx.EnvironmentIndexInto(ctx, index)

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(HaveOccurred())
				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
			})

			It("does not create HTTPRoute when no service is available", func() {
				reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects().Build(), serviceClusterIP: ""}

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
			})

			It("if HTTPRouted is created but not yet accepted and Ingress is present, leaves both present but ExpositionAccepted to false", func() {
				ingress := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
				httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
				reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects(&ingress, &httpRoute).Build(), serviceClusterIP: clusterIP}

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(Succeed())
				Expect(instance.Status.Environments[index].ExpositionAccepted).To(BeFalse())
			})

			It("if HTTPRoute is accepted, ensures Ingress is absent and sets ExpositionAccepted to true", func() {
				httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
				httpRoute.Status.Parents = []gatewayv1.RouteParentStatus{{ControllerName: "gateway.networking.k8s.io/gateway-controller", ParentRef: gatewayv1.ParentReference{Name: "fake-gw", Namespace: ptr.To(gatewayv1.Namespace("fake-gw-ns"))}, Conditions: []metav1.Condition{{Type: string(gatewayv1.RouteConditionAccepted), Status: metav1.ConditionTrue}}}}
				reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects(&httpRoute).Build(), serviceClusterIP: clusterIP}

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
				Expect(instance.Status.Environments[index].ExpositionAccepted).To(BeTrue())
			})

			It("skips creating HTTPRoute for GUI-less ClassVMs", func() {
				environment.EnvironmentType = clv1alpha2.ClassVM
				environment.GuiEnabled = false
				ctx, _ = clctx.EnvironmentInto(ctx, &environment)

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
			})

			It("skips creating Ingress for GUI-less CloudVMs", func() {
				environment.EnvironmentType = clv1alpha2.ClassCloudVM
				environment.GuiEnabled = false
				ctx, _ = clctx.EnvironmentInto(ctx, &environment)

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
			})
		})

		Context("Gateway API mode disabled", func() {
			BeforeEach(func() { reconciler.ExpositionConfig.GatewayAPIMode = false })

			It("creates Ingress and Service", func() {
				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
				Expect(instance.Status.Environments[index].IP).To(Equal(clusterIP))
				Expect(instance.Status.Environments[index].ExpositionAccepted).To(BeTrue())
			})

			It("deletes HTTPRoute if present and creates Ingress", func() {
				httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
				reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects(&httpRoute).Build(), serviceClusterIP: clusterIP}

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(Succeed())
			})

			It("adds authentication annotations when enabled", func() {
				reconciler.ExpositionConfig.EnableAuthentication = true

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				ing := netv1.Ingress{}
				Expect(reconciler.Client.Get(ctx, ingressName, &ing)).To(Succeed())
				Expect(ing.Annotations).To(HaveKey("nginx.ingress.kubernetes.io/auth-url"))
				Expect(ing.Annotations).To(HaveKey("nginx.ingress.kubernetes.io/auth-signin"))
			})

			It("skips creating Ingress for GUI-less ClassVMs", func() {
				environment.EnvironmentType = clv1alpha2.ClassVM
				environment.GuiEnabled = false
				ctx, _ = clctx.EnvironmentInto(ctx, &environment)

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
			})

			It("skips creating Ingress for GUI-less CloudVMs", func() {
				environment.EnvironmentType = clv1alpha2.ClassCloudVM
				environment.GuiEnabled = false
				ctx, _ = clctx.EnvironmentInto(ctx, &environment)

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
			})

			It("does not create Ingress when no service is available", func() {
				reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects().Build(), serviceClusterIP: ""}

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
			})
		})

		Context("the environment is GUI-less", func() {
			BeforeEach(func() { environment.GuiEnabled = false })

			It("skips creating HTTPRoute or Ingress for GUI-less ClassVMs", func() {
				environment.EnvironmentType = clv1alpha2.ClassVM
				ctx, _ = clctx.EnvironmentInto(ctx, &environment)

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
			})

			It("skips creating HTTPRoute or Ingress for GUI-less LocalVMs", func() {
				environment.EnvironmentType = clv1alpha2.ClassLocalVM
				ctx, _ = clctx.EnvironmentInto(ctx, &environment)

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
			})

			It("skips creating HTTPRoute or Ingress for GUI-less CloudVMs", func() {
				environment.EnvironmentType = clv1alpha2.ClassCloudVM
				ctx, _ = clctx.EnvironmentInto(ctx, &environment)

				err := reconciler.EnforceInstanceExposition(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(Succeed())
				Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
				Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
			})
		})
	})

	Context("When instance is not running", func() {
		BeforeEach(func() { instance.Spec.Running = false })

		It("extends Status.Environments and clears IP when index is out of range", func() {
			index = 5
			ctx = clctx.EnvironmentIndexInto(ctx, index)

			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(instance.Status.Environments)).To(BeNumerically(">", index))
			Expect(instance.Status.Environments[index].IP).To(Equal(""))
			Expect(instance.Status.Environments[index].ExpositionAccepted).To(BeFalse())
		})

		It("removes Service/Ingress/HTTPRoute if present and clears status", func() {
			svc := corev1.Service{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
			ingress := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
			httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
			reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects(&svc, &ingress, &httpRoute).Build(), serviceClusterIP: clusterIP}

			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(reconciler.Client.Get(ctx, serviceName, &corev1.Service{})).To(HaveOccurred())
			Expect(reconciler.Client.Get(ctx, ingressName, &netv1.Ingress{})).To(HaveOccurred())
			Expect(reconciler.Client.Get(ctx, httpRouteName, &gatewayv1.HTTPRoute{})).To(HaveOccurred())
			Expect(instance.Status.Environments[index].IP).To(Equal(""))
			Expect(instance.Status.Environments[index].ExpositionAccepted).To(BeFalse())
		})
	})

	Context("Failure cases", func() {
		It("does not error when no resources exist and leaves status untouched", func() {
			instance.Spec.Running = false

			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(instance.Status.Environments[index].IP).To(Equal(""))
		})

		It("does not error when httpRoute exists but has no status", func() {
			reconciler.ExpositionConfig.GatewayAPIMode = true
			instance.Spec.Running = true
			httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
			reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects(&httpRoute).Build(), serviceClusterIP: clusterIP}

			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(instance.Status.Environments[index].ExpositionAccepted).To(BeFalse())
		})

		It("does not error when getHTTPRouteAcceptedStatus receives HTTPRoute with empty status", func() {
			reconciler.ExpositionConfig.GatewayAPIMode = true
			instance.Spec.Running = true
			httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(&instance, environment.Name)}
			reconciler.Client = FakeClientWrapped{Client: clientBuilder.WithObjects(&httpRoute).Build(), serviceClusterIP: clusterIP}

			err := reconciler.EnforceInstanceExposition(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(instance.Status.Environments[index].ExpositionAccepted).To(BeFalse())
		})
	})
})
