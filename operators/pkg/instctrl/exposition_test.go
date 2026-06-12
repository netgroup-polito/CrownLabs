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
	"reflect"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

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

var _ = Describe("Generation of the exposition environment", func() {
	var (
		ctx           context.Context
		clientBuilder fake.ClientBuilder
		reconciler    instctrl.InstanceReconciler

		enableAuth       bool
		instancesAuthURL string

		instance    clv1alpha2.Instance
		environment clv1alpha2.Environment
		template    clv1alpha2.Template
		index       int

		serviceName types.NamespacedName
		service     corev1.Service

		ingressGUIName types.NamespacedName
		ingress        netv1.Ingress

		httpRouteName types.NamespacedName
		httpRoute     gatewayv1.HTTPRoute

		ownerRef metav1.OwnerReference

		err error

		enableGateway        bool
		gatewayAPIRefsValues string
	)

	const (
		instanceName      = "kubernetes-0000"
		instanceNamespace = "tenant-tester"
		instanceUID       = "dcc6ead1-0040-451b-ba68-787ebfb68640"
		templateName      = "kubernetes"
		templateNamespace = "workspace-netgroup"
		environmentName   = "control-plane"
		tenantName        = "tester"
		host              = "crownlabs.example.com"
		clusterIP         = "1.1.1.1"
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

		environment = clv1alpha2.Environment{Name: environmentName, EnvironmentType: clv1alpha2.ClassContainer}
		template = clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateName},
			Spec: clv1alpha2.TemplateSpec{
				EnvironmentList: []clv1alpha2.Environment{environment},
			},
		}
		instance = clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace, UID: instanceUID},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
				Tenant:   clv1alpha2.GenericRef{Name: tenantName},
			},
			Status: clv1alpha2.InstanceStatus{
				Environments: []clv1alpha2.InstanceStatusEnv{
					{Phase: ""},
					{Phase: ""},
					{Phase: ""},
				},
			},
		}
		index = 0

		serviceName = forge.NamespacedNameWithSuffix(&instance, environment.Name)
		ingressGUIName = forge.NamespacedNameWithSuffix(&instance, environment.Name)
		httpRouteName = forge.NamespacedNameWithSuffix(&instance, environment.Name)

		service = corev1.Service{}
		ingress = netv1.Ingress{}

		ownerRef = metav1.OwnerReference{
			APIVersion:         clv1alpha2.GroupVersion.String(),
			Kind:               "Instance",
			Name:               instance.GetName(),
			UID:                instance.GetUID(),
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
		}

		enableAuth = false
		instancesAuthURL = ""
		enableGateway = false
		gatewayAPIRefsValues = ""
	})

	JustBeforeEach(func() {
		client := FakeClientWrapped{Client: clientBuilder.Build(), serviceClusterIP: clusterIP}
		expo := instctrl.ExpositionConfig{WebsiteBaseURL: host, InstancesAuthURL: instancesAuthURL}
		expo.GatewayAPIMode = enableGateway
		if gatewayAPIRefsValues != "" {
			ns, name, section := utils.ParseGatewayParent(gatewayAPIRefsValues)
			expo.GatewayNamespace = ns
			expo.GatewayName = name
			expo.GatewaySectionName = section
		}
		reconciler = instctrl.InstanceReconciler{
			Client:               client,
			Scheme:               scheme.Scheme,
			ExpositionConfig:     expo,
			EventsRecorder:       record.NewFakeRecorder(1024),
			EnableAuthentication: enableAuth,
		}

		ctx, _ = clctx.InstanceInto(ctx, &instance)
		ctx, _ = clctx.EnvironmentInto(ctx, &environment)
		ctx, _ = clctx.TemplateInto(ctx, &template)
		ctx = clctx.EnvironmentIndexInto(ctx, index)
		err = reconciler.EnforceInstanceExposition(ctx)
	})

	type DescribeBodyParameters struct {
		NamespacedName *types.NamespacedName
		Object         client.Object

		ExpectedSpecForger func(*clv1alpha2.Instance, *clv1alpha2.Environment) interface{}
		EmptySpec          interface{}

		InstanceStatusGetter   func(*clv1alpha2.Instance) string
		InstanceStatusExpected string

		GroupResource schema.GroupResource
	}

	DescribeBodyParametersService := DescribeBodyParameters{
		NamespacedName: &serviceName, Object: &service, GroupResource: corev1.Resource("services"),
		ExpectedSpecForger: func(inst *clv1alpha2.Instance, env *clv1alpha2.Environment) interface{} {
			svc := forge.ServiceSpec(inst, env)
			// Normalise empty ports slice to nil to match fake client's representation
			if len(svc.Ports) == 0 {
				svc.Ports = nil
			}
			svc.ClusterIP = clusterIP
			return svc
		},
		EmptySpec: corev1.ServiceSpec{ClusterIP: clusterIP},
		InstanceStatusGetter: func(inst *clv1alpha2.Instance) string {
			if index >= len(inst.Status.Environments) {
				return ""
			}

			return inst.Status.Environments[index].IP
		},
		InstanceStatusExpected: clusterIP,
	}

	DescribeBodyParametersIngressGUI := DescribeBodyParameters{
		NamespacedName: &ingressGUIName, Object: &ingress, GroupResource: netv1.Resource("ingresses"),
		ExpectedSpecForger: func(inst *clv1alpha2.Instance, _ *clv1alpha2.Environment) interface{} {
			return forge.IngressSpec(host, forge.ExpositionGUIPath(inst, &environment),
				forge.IngressDefaultCertificateName, serviceName.Name, forge.GUIPortName)
		},
		EmptySpec: netv1.IngressSpec{},
		InstanceStatusGetter: func(_ *clv1alpha2.Instance) string {
			return ""
		},
		InstanceStatusExpected: "",
	}

	DescribeBodyParametersIngressGUIContainer := DescribeBodyParameters{
		NamespacedName: &ingressGUIName, Object: &ingress, GroupResource: netv1.Resource("ingresses"),
		ExpectedSpecForger: func(inst *clv1alpha2.Instance, _ *clv1alpha2.Environment) interface{} {
			return forge.IngressSpec(host, forge.ExpositionGUIPath(inst, &environment),
				forge.IngressDefaultCertificateName, serviceName.Name, forge.GUIPortName)
		},
		EmptySpec: netv1.IngressSpec{},
		InstanceStatusGetter: func(_ *clv1alpha2.Instance) string {
			return ""
		},
		InstanceStatusExpected: "",
	}

	DescribeBodyParametersHTTPRoute := DescribeBodyParameters{
		NamespacedName: &httpRouteName, Object: &httpRoute, GroupResource: schema.GroupResource{Group: gatewayv1.GroupVersion.Group, Resource: "httproutes"},
		ExpectedSpecForger: func(inst *clv1alpha2.Instance, _ *clv1alpha2.Environment) interface{} {
			params := &forge.HTTPRouteSpecParams{Host: host, Path: forge.ExpositionGUIPath(inst, &environment), ServiceName: serviceName.Name}
			if gatewayAPIRefsValues != "" {
				ns, name, section := utils.ParseGatewayParent(gatewayAPIRefsValues)
				params.GatewayNamespace = ns
				params.GatewayName = name
				params.GatewaySectionName = section
			}
			return forge.HTTPRouteSpec(params, &environment, forge.GUIPortNumber)
		},
		EmptySpec:              gatewayv1.HTTPRouteSpec{},
		InstanceStatusGetter:   func(_ *clv1alpha2.Instance) string { return "" },
		InstanceStatusExpected: "",
	}

	Context("The instance is running", func() {
		BeforeEach(func() { instance.Spec.Running = true })

		ObjectToSpec := func(obj client.Object) interface{} {
			if svc, ok := obj.(*corev1.Service); ok {
				return svc.Spec
			} else if ing, ok := obj.(*netv1.Ingress); ok {
				return ing.Spec
			} else if hr, ok := obj.(*gatewayv1.HTTPRoute); ok {
				return hr.Spec
			}
			Fail("Unexpected resource type " + reflect.TypeOf(obj).String())
			return nil
		}

		Describe("Authentication annotations", func() {
			Context("when authentication is enabled and InstancesAuthURL is set", func() {
				BeforeEach(func() {
					enableAuth = true
					instancesAuthURL = "https://auth.example.com"
				})

				It("adds auth annotations to the GUI ingress", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(reconciler.Get(ctx, ingressGUIName, &ingress)).To(Succeed())
					ann := ingress.GetAnnotations()
					Expect(ann).To(HaveKeyWithValue("nginx.ingress.kubernetes.io/auth-url", "https://auth.example.com/auth"))
					Expect(ann).To(HaveKeyWithValue("nginx.ingress.kubernetes.io/auth-signin", "https://auth.example.com/start?rd=$escaped_request_uri"))
					Expect(ann).To(HaveKeyWithValue("nginx.ingress.kubernetes.io/proxy-read-timeout", "3600"))
				})
			})

			Context("when authentication is enabled and InstancesAuthURL is empty", func() {
				BeforeEach(func() {
					enableAuth = true
					instancesAuthURL = ""
				})

				It("adds relative auth annotations to the GUI ingress", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(reconciler.Get(ctx, ingressGUIName, &ingress)).To(Succeed())
					ann := ingress.GetAnnotations()
					Expect(ann).To(HaveKeyWithValue("nginx.ingress.kubernetes.io/auth-url", "/auth"))
					Expect(ann).To(HaveKeyWithValue("nginx.ingress.kubernetes.io/auth-signin", "/start?rd=$escaped_request_uri"))
				})
			})

			Context("when authentication is disabled", func() {
				BeforeEach(func() {
					enableAuth = false
					instancesAuthURL = "https://auth.example.com"
				})

				It("does not add auth annotations to the GUI ingress", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(reconciler.Get(ctx, ingressGUIName, &ingress)).To(Succeed())
					ann := ingress.GetAnnotations()
					Expect(ann).ToNot(HaveKey("nginx.ingress.kubernetes.io/auth-url"))
					Expect(ann).ToNot(HaveKey("nginx.ingress.kubernetes.io/auth-signin"))
				})
			})
		})

		DescribeBodyPresent := func(p DescribeBodyParameters) {
			When("it is not yet present", func() {
				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("Should be present and have the common attributes", func() {
					Expect(reconciler.Get(ctx, *p.NamespacedName, p.Object)).To(Succeed())
					for k, v := range forge.InstanceObjectLabels(nil, &instance) {
						Expect(p.Object.GetLabels()).To(HaveKeyWithValue(k, v))
					}
					Expect(p.Object.GetOwnerReferences()).To(ContainElement(ownerRef))
				})

				It("Should be present and have the expected specs", func() {
					Expect(reconciler.Get(ctx, *p.NamespacedName, p.Object)).To(Succeed())
					Expect(p.Object).To(WithTransform(ObjectToSpec, Equal(p.ExpectedSpecForger(&instance, &environment))))
				})

				It("Should fill the correct instance status value", func() {
					Expect(p.InstanceStatusGetter(&instance)).To(BeIdenticalTo(p.InstanceStatusExpected))
				})
			})

			When("it is already present", func() {
				BeforeEach(func() {
					svc := corev1.Service{ObjectMeta: forge.NamespacedNameToObjectMeta(serviceName)}
					svc.SetCreationTimestamp(metav1.NewTime(time.Now()))
					svc.Spec.ClusterIP = clusterIP

					// prepare objects to pre-seed in the fake client depending on the resource under test
					objs := []client.Object{&svc}

					// if testing ingresses, pre-seed the ingress object
					if p.GroupResource.Resource == "ingresses" {
						ingGUI := netv1.Ingress{ObjectMeta: forge.NamespacedNameToObjectMeta(ingressGUIName)}
						ingGUI.SetCreationTimestamp(metav1.NewTime(time.Now()))
						objs = append(objs, &ingGUI)
					}

					// if testing httproutes, pre-seed the HTTPRoute object
					if p.GroupResource.Resource == "httproutes" {
						hr := gatewayv1.HTTPRoute{ObjectMeta: forge.NamespacedNameToObjectMeta(*p.NamespacedName)}
						hr.SetCreationTimestamp(metav1.NewTime(time.Now()))
						objs = append(objs, &hr)
					}

					clientBuilder.WithObjects(objs...)
				})

				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("Should still be present and have the common attributes", func() {
					Expect(reconciler.Get(ctx, *p.NamespacedName, p.Object)).To(Succeed())
					for k, v := range forge.InstanceObjectLabels(nil, &instance) {
						Expect(p.Object.GetLabels()).To(HaveKeyWithValue(k, v))
					}
					Expect(p.Object.GetOwnerReferences()).To(ContainElement(ownerRef))
				})

				It("Should still be present and have unmodified specs", func() {
					Expect(reconciler.Get(ctx, *p.NamespacedName, p.Object)).To(Succeed())
					Expect(p.Object).To(WithTransform(ObjectToSpec, Equal(p.EmptySpec)))
				})

				It("Should fill the correct instance status value", func() {
					Expect(p.InstanceStatusGetter(&instance)).To(BeIdenticalTo(p.InstanceStatusExpected))
				})
			})
		}

		DescribeBodyAbsent := func(p DescribeBodyParameters) {
			When("it is not yet present", func() {
				var notFoundError error

				BeforeEach(func() {
					notFoundError = kerrors.NewNotFound(p.GroupResource, p.NamespacedName.Name)
				})

				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("Should not be created", func() {
					Expect(reconciler.Get(ctx, *p.NamespacedName, p.Object)).To(MatchError(notFoundError))
				})

				It("Should set the instance status empty", func() {
					Expect(p.InstanceStatusGetter(&instance)).To(BeIdenticalTo(""))
				})
			})
		}

		Context("The environment is VM-based", func() {
			BeforeEach(func() { environment.EnvironmentType = clv1alpha2.ClassVM })

			Context("The environment has a GUI", func() {
				BeforeEach(func() { environment.GuiEnabled = true })

				Describe("Assessing the service presence", func() { DescribeBodyPresent(DescribeBodyParametersService) })
				Describe("Assessing the GUI ingress presence", func() { DescribeBodyPresent(DescribeBodyParametersIngressGUI) })
			})

			Context("The environment has not a GUI", func() {
				BeforeEach(func() { environment.GuiEnabled = false })

				Describe("Assessing the service presence", func() { DescribeBodyPresent(DescribeBodyParametersService) })
				Describe("Assessing the GUI ingress absence", func() { DescribeBodyAbsent(DescribeBodyParametersIngressGUI) })
			})
		})

		Context("The environment is CloudVM-based", func() {
			BeforeEach(func() { environment.EnvironmentType = clv1alpha2.ClassCloudVM })

			Context("The environment has a GUI", func() {
				BeforeEach(func() { environment.GuiEnabled = true })

				Describe("Assessing the service presence", func() { DescribeBodyPresent(DescribeBodyParametersService) })
				Describe("Assessing the GUI ingress presence", func() { DescribeBodyPresent(DescribeBodyParametersIngressGUI) })
			})

			Context("The environment has not a GUI", func() {
				BeforeEach(func() { environment.GuiEnabled = false })

				Describe("Assessing the service presence", func() { DescribeBodyPresent(DescribeBodyParametersService) })
				Describe("Assessing the GUI ingress absence", func() { DescribeBodyAbsent(DescribeBodyParametersIngressGUI) })
			})
		})

		Context("The environment is Container-based", func() {
			BeforeEach(func() {
				environment.EnvironmentType = clv1alpha2.ClassContainer
				environment.GuiEnabled = true
			})

			Describe("Assessing the service presence", func() { DescribeBodyPresent(DescribeBodyParametersService) })
			Describe("Assessing the GUI ingress presence", func() { DescribeBodyPresent(DescribeBodyParametersIngressGUIContainer) })

			Context("When Gateway API is enabled", func() {
				BeforeEach(func() { enableGateway = true })
				BeforeEach(func() { gatewayAPIRefsValues = "" })
				BeforeEach(func() {
					environment.EnvironmentType = clv1alpha2.ClassContainer
					environment.GuiEnabled = true
				})

				BeforeEach(func() {
					gatewayAPIRefsValues = templateNamespace + "/test-gateway/section"
				})

				Describe("Assessing the service presence", func() { DescribeBodyPresent(DescribeBodyParametersService) })
				Describe("Assessing the GUI HTTPRoute presence", func() { DescribeBodyPresent(DescribeBodyParametersHTTPRoute) })
			})
		})

		Context("The instance is multi-environment", func() {

			Context("The environment has an index greater than 1", func() {
				BeforeEach(func() {
					index = 1
				})
				Describe("Assessing the service presence", func() { DescribeBodyPresent(DescribeBodyParametersService) })
			})

			Context("The index is out of range", func() {
				BeforeEach(func() {
					index = 3
				})
				Describe("Assessing the service absence", func() { DescribeBodyAbsent(DescribeBodyParametersService) })
			})
		})

	})

	Context("The instance is not running", func() {
		BeforeEach(func() { instance.Spec.Running = false })

		DescribeBody := func(p DescribeBodyParameters) {
			var notFoundError error

			BeforeEach(func() {
				notFoundError = kerrors.NewNotFound(p.GroupResource, p.NamespacedName.Name)
			})

			When("it has not yet been deleted", func() {
				BeforeEach(func() {
					svc := corev1.Service{ObjectMeta: forge.NamespacedNameToObjectMeta(serviceName)}
					ingGUI := netv1.Ingress{ObjectMeta: forge.NamespacedNameToObjectMeta(ingressGUIName)}
					clientBuilder.WithObjects(&svc, &ingGUI)
				})

				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
				It("Should not be present", func() {
					Expect(reconciler.Get(ctx, *p.NamespacedName, p.Object)).To(MatchError(notFoundError))
				})
				It("Should set the instance status empty", func() {
					Expect(p.InstanceStatusGetter(&instance)).To(BeIdenticalTo(""))
				})
			})

			When("it has already been deleted", func() {
				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
				It("Should not be present", func() {
					Expect(reconciler.Get(ctx, *p.NamespacedName, p.Object)).To(MatchError(notFoundError))
				})
				It("Should set the instance status empty", func() {
					Expect(p.InstanceStatusGetter(&instance)).To(BeIdenticalTo(""))
				})
			})
		}

		Describe("Assessing the service deletion", func() { DescribeBody(DescribeBodyParametersService) })
		Describe("Assessing the GUI ingress deletion", func() { DescribeBody(DescribeBodyParametersIngressGUI) })

		Context("When Gateway API is enabled for deletion", func() {
			BeforeEach(func() { enableGateway = true })

			BeforeEach(func() {
				// prepare objects: service and HTTPRoute
				svc := corev1.Service{ObjectMeta: forge.NamespacedNameToObjectMeta(serviceName)}
				hr := gatewayv1.HTTPRoute{ObjectMeta: forge.NamespacedNameToObjectMeta(httpRouteName)}
				svc.SetCreationTimestamp(metav1.NewTime(time.Now()))
				hr.SetCreationTimestamp(metav1.NewTime(time.Now()))
				clientBuilder.WithObjects(&svc, &hr)
			})

			It("Should delete the HTTPRoute and clear status", func() {
				Expect(err).ToNot(HaveOccurred())
				notFoundError := kerrors.NewNotFound(schema.GroupResource{Group: gatewayv1.GroupVersion.Group, Resource: "httproutes"}, httpRouteName.Name)
				Expect(reconciler.Get(ctx, httpRouteName, &httpRoute)).To(MatchError(notFoundError))
				Expect(instance.Status.Environments[0].IP).To(BeIdenticalTo(""))
			})
		})
	})
})
