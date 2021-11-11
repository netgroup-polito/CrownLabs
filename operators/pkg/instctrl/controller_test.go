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

package instctrl_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	virtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	. "github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

// The following are integration tests aiming to verify that the instance controller
// reconcile method correctly creates the various resources it has to manage,
// by running several combinations of configuration. This is not exaustive
// as a more granular coverage is achieved through the unit tests in this package.
var _ = Describe("The instance-controller Reconcile method", func() {
	ctx := context.Background()
	var (
		testName       string
		prettyName     string
		runInstance    bool
		instance       clv1alpha2.Instance
		environment    clv1alpha2.Environment
		ingress        netv1.Ingress
		service        corev1.Service
		createTenant   bool
		createTemplate bool
	)

	RunReconciler := func() error {
		_, err := instanceReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: forge.NamespacedName(&instance),
		})
		if err != nil {
			return err
		}
		return k8sClient.Get(ctx, forge.NamespacedName(&instance), &instance)
	}

	BeforeEach(func() {
		createTenant = true
		createTemplate = true
		environment = clv1alpha2.Environment{
			Name:            "app",
			Image:           "some-image:v0",
			EnvironmentType: clv1alpha2.ClassVM,
			Persistent:      false,
			GuiEnabled:      true,
			Resources: clv1alpha2.EnvironmentResources{
				CPU:                   1,
				ReservedCPUPercentage: 20,
				Memory:                *resource.NewScaledQuantity(1, resource.Giga),
				Disk:                  *resource.NewScaledQuantity(10, resource.Giga),
			},
			Mode: clv1alpha2.ModeStandard,
		}
	})

	JustBeforeEach(func() {
		ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testName, Labels: whiteListMap}}
		webdavSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: webdavSecretName, Namespace: testName},
			Data:       map[string][]byte{"username": []byte(testName), "password": []byte(testName)},
		}
		tenant := clv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{Name: testName},
			Spec: clv1alpha2.TenantSpec{
				Email: "test@email.me",
			},
		}
		template := clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: testName, Namespace: testName},
			Spec: clv1alpha2.TemplateSpec{
				WorkspaceRef:    clv1alpha2.GenericRef{Name: testName},
				EnvironmentList: []clv1alpha2.Environment{environment},
			},
		}
		instance = clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: testName, Namespace: testName},
			Spec: clv1alpha2.InstanceSpec{
				Template:   clv1alpha2.GenericRef{Name: testName, Namespace: testName},
				Tenant:     clv1alpha2.GenericRef{Name: testName, Namespace: testName},
				Running:    runInstance,
				PrettyName: prettyName,
			},
		}

		Expect(k8sClient.Create(ctx, &ns)).To(Succeed())
		Expect(k8sClient.Create(ctx, &webdavSecret)).To(Succeed())
		if createTenant {
			Expect(k8sClient.Create(ctx, &tenant)).To(Succeed())
		}
		if createTemplate {
			Expect(k8sClient.Create(ctx, &template)).To(Succeed())
		}
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
	})

	Context("The instance is container based", func() {
		When("the environment is persistent", func() {
			BeforeEach(func() {
				testName = "test-container-persistent"
				environment.Persistent = true
				environment.EnvironmentType = clv1alpha2.ClassContainer
				runInstance = false
			})

			It("Should correctly reconcile the instance", func() {
				Expect(RunReconciler()).To(Succeed())

				Expect(instance.Status.Phase).To(Equal(clv1alpha2.EnvironmentPhaseOff))

				By("Asserting the deployment has been created", func() {
					var deploy appsv1.Deployment
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &deploy)).To(Succeed())
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 0)))
				})

				By("Asserting the PVC has been created", func() {
					var pvc corev1.PersistentVolumeClaim
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &pvc)).To(Succeed())
				})

				By("Asserting the exposition resources aren't present", func() {
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &service)).To(FailBecauseNotFound())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressGUINameSuffix), &ingress)).To(FailBecauseNotFound())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressMyDriveNameSuffix), &ingress)).To(FailBecauseNotFound())
				})

				By("Setting the instance to running", func() {
					instance.Spec.Running = true
					Expect(k8sClient.Update(ctx, &instance)).To(Succeed())
					Expect(RunReconciler()).To(Succeed())
				})

				By("Asserting the right exposition resources exist", func() {
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &service)).To(Succeed())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressGUINameSuffix), &ingress)).To(Succeed())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressMyDriveNameSuffix), &ingress)).To(Succeed())
				})

				By("Asserting the state is coherent", func() {
					Expect(instance.Status.Phase).To(Equal(clv1alpha2.EnvironmentPhaseStarting))
				})

				By("Asserting the deployment has been created", func() {
					var deploy appsv1.Deployment
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &deploy)).To(Succeed())
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 1)))
				})
			})
		})
		When("the environment is NOT persistent", func() {
			BeforeEach(func() {
				testName = "test-container-not-persistent"
				environment.EnvironmentType = clv1alpha2.ClassContainer
				runInstance = false
			})

			It("Should correctly reconcile the instance", func() {
				Expect(RunReconciler()).To(Succeed())

				Expect(instance.Status.Phase).To(Equal(clv1alpha2.EnvironmentPhaseOff))

				By("Asserting the deployment has been created with no replicas", func() {
					var deploy appsv1.Deployment
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &deploy)).To(Succeed())
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 0)))
				})

				By("Asserting the exposition resources aren't present", func() {
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &service)).To(FailBecauseNotFound())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressGUINameSuffix), &ingress)).To(FailBecauseNotFound())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressMyDriveNameSuffix), &ingress)).To(FailBecauseNotFound())
				})

				By("Setting the instance to running", func() {
					instance.Spec.Running = true
					Expect(k8sClient.Update(ctx, &instance)).To(Succeed())
					Expect(RunReconciler()).To(Succeed())
				})

				By("Asserting the right exposition resources exist", func() {
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &service)).To(Succeed())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressGUINameSuffix), &ingress)).To(Succeed())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressMyDriveNameSuffix), &ingress)).To(Succeed())
				})

				By("Asserting the state is coherent", func() {
					Expect(instance.Status.Phase).To(Equal(clv1alpha2.EnvironmentPhaseStarting))
				})

				By("Asserting the deployment has been created", func() {
					var deploy appsv1.Deployment
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &deploy)).To(Succeed())
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 1)))
				})
			})
		})
	})

	Context("The instance is VM based", func() {
		When("the environment is persistent", func() {
			BeforeEach(func() {
				testName = "test-vm-persistent"
				environment.Persistent = true
				runInstance = false
			})

			It("Should correctly reconcile the instance", func() {
				Expect(RunReconciler()).To(Succeed())

				// Check the status phase is unset since it's retrieved from the VM (and the kubervirt operator is not available in the test env)
				Expect(instance.Status.Phase).To(Equal(clv1alpha2.EnvironmentPhaseUnset))

				By("Asserting the VM has been created", func() {
					var vm virtv1.VirtualMachine
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &vm)).To(Succeed())
					Expect(vm.Spec.Running).To(PointTo(BeFalse()))
				})

				By("Asserting the cloudinit secret has been created", func() {
					var secret corev1.Secret
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &secret)).To(Succeed())
				})

				By("Asserting the exposition resources aren't present", func() {
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &service)).To(FailBecauseNotFound())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressGUINameSuffix), &ingress)).To(FailBecauseNotFound())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressMyDriveNameSuffix), &ingress)).To(FailBecauseNotFound())
				})

				By("Setting the instance to running", func() {
					instance.Spec.Running = true
					Expect(k8sClient.Update(ctx, &instance)).To(Succeed())
					Expect(RunReconciler()).To(Succeed())
				})

				By("Asserting the right exposition resources exist", func() {
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &service)).To(Succeed())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressGUINameSuffix), &ingress)).To(Succeed())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressMyDriveNameSuffix), &ingress)).To(FailBecauseNotFound())
				})

				By("Asserting the state is coherent", func() {
					var vm virtv1.VirtualMachine
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &vm)).To(Succeed())
					vm.Status.PrintableStatus = virtv1.VirtualMachineStatusRunning
					Expect(k8sClient.Update(ctx, &vm))
					Expect(RunReconciler()).To(Succeed())
					Expect(instance.Status.Phase).To(Equal(clv1alpha2.EnvironmentPhaseRunning))
				})

				By("Asserting the VM spec has been changed", func() {
					var vm virtv1.VirtualMachine
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &vm)).To(Succeed())
					Expect(vm.Spec.Running).To(PointTo(BeTrue()))
				})
			})
		})

		When("the environment is NOT persistent", func() {
			BeforeEach(func() {
				testName = "test-vm-not-persistent"
				runInstance = false
			})

			It("Should correctly reconcile the instance", func() {
				Expect(RunReconciler()).To(Succeed())

				Expect(instance.Status.Phase).To(Equal(clv1alpha2.EnvironmentPhaseOff))

				By("Asserting the VM has NOT been created", func() {
					var vmi virtv1.VirtualMachineInstance
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &vmi)).To(FailBecauseNotFound())
				})

				By("Asserting the cloudinit secret has been created", func() {
					var secret corev1.Secret
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &secret)).To(Succeed())
				})

				By("Asserting the exposition resources aren't present", func() {
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &service)).To(FailBecauseNotFound())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressGUINameSuffix), &ingress)).To(FailBecauseNotFound())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressMyDriveNameSuffix), &ingress)).To(FailBecauseNotFound())
				})

				By("Setting the instance to running", func() {
					instance.Spec.Running = true
					Expect(k8sClient.Update(ctx, &instance)).To(Succeed())
					Expect(RunReconciler()).To(Succeed())
				})

				By("Asserting the right exposition resources exist", func() {
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &service)).To(Succeed())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressGUINameSuffix), &ingress)).To(Succeed())
					Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(&instance, forge.IngressMyDriveNameSuffix), &ingress)).To(FailBecauseNotFound())
				})

				By("Asserting the VM has been created", func() {
					var vmi virtv1.VirtualMachineInstance
					Expect(k8sClient.Get(ctx, forge.NamespacedName(&instance), &vmi)).To(Succeed())
					vmi.Status.Phase = virtv1.Running
					Expect(k8sClient.Update(ctx, &vmi)).To(Succeed())
				})

				By("Asserting the state is coherent", func() {
					Expect(RunReconciler()).To(Succeed())
					Expect(instance.Status.Phase).To(Equal(clv1alpha2.EnvironmentPhaseRunning))
				})
			})
		})
	})

	Context("PrettyName handling", func() {
		When("no name is set", func() {
			BeforeEach(func() {
				testName = "test-pretty-name-gen"
				prettyName = ""
			})

			It("should set a prettyName", func() {
				Expect(RunReconciler()).To(Succeed())
				Expect(instance.Spec.PrettyName).NotTo(BeEmpty())
			})
		})

		When("name is set", func() {
			const prettyNameTest = "Some pretty name!"
			BeforeEach(func() {
				testName = "test-pretty-name-set"
				prettyName = prettyNameTest
			})

			It("should not change the prettyName", func() {
				Expect(RunReconciler()).To(Succeed())
				Expect(instance.Spec.PrettyName).To(Equal(prettyNameTest))
			})
		})
	})

	Context("In case of misconfiguration", func() {
		When("the template is missing", func() {
			BeforeEach(func() {
				testName = "test-missing-template"
				createTemplate = false
			})

			It("Should fail instance reconcile", func() {
				Expect(RunReconciler()).To(HaveOccurred())
			})
		})

		When("the tenant is missing", func() {
			BeforeEach(func() {
				testName = "test-missing-tenant"
				createTenant = false
			})

			It("Should fail instance reconcile", func() {
				Expect(RunReconciler()).To(HaveOccurred())
			})
		})
	})
})
