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
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	virtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
)

var _ = Describe("Generation of the virtual machine and virtual machine instances", func() {
	var (
		ctx           context.Context
		clientBuilder fake.ClientBuilder
		reconciler    instctrl.InstanceReconciler

		instance    clv1alpha2.Instance
		template    clv1alpha2.Template
		environment clv1alpha2.Environment
		tenant      clv1alpha2.Tenant

		objectName types.NamespacedName
		svc        corev1.Service
		secret     corev1.Secret
		vm         virtv1.VirtualMachine
		vmi        virtv1.VirtualMachineInstance

		ownerRef metav1.OwnerReference

		err error
	)

	const (
		instanceName      = "kubernetes-0000"
		instanceNamespace = "tenant-tester"
		templateName      = "kubernetes"
		templateNamespace = "workspace-netgroup"
		environmentName   = "control-plane"
		tenantName        = "tester"
		workspaceName     = "netgroup"
		webdavCredentials = "webdav-credentials"

		image       = "internal/registry/image:v1.0"
		cpu         = 2
		cpuReserved = 25
		memory      = "1250M"
		disk        = "20Gi"
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
			// These objects are required by the EnforceCloudInitSecret function.
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: webdavCredentials, Namespace: instanceNamespace},
				Data: map[string][]byte{
					instctrl.WebdavSecretUsernameKey: []byte("username"),
					instctrl.WebdavSecretPasswordKey: []byte("password"),
				},
			},
			&clv1alpha2.Template{ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateNamespace}},
			&clv1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: tenantName}},
		)

		instance = clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
			Spec: clv1alpha2.InstanceSpec{
				Running:  true,
				Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
				Tenant:   clv1alpha2.GenericRef{Name: tenantName},
			},
		}
		environment = clv1alpha2.Environment{
			Name:            environmentName,
			EnvironmentType: clv1alpha2.ClassVM,
			Image:           image,
			Resources: clv1alpha2.EnvironmentResources{
				CPU:                   cpu,
				ReservedCPUPercentage: cpuReserved,
				Memory:                resource.MustParse(memory),
				Disk:                  resource.MustParse(disk),
			},
		}
		template = clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateNamespace},
			Spec: clv1alpha2.TemplateSpec{
				WorkspaceRef:    clv1alpha2.GenericRef{Name: workspaceName},
				EnvironmentList: []clv1alpha2.Environment{environment},
			},
		}

		tenant = clv1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: tenantName}}

		objectName = forge.NamespacedName(&instance)

		svc = corev1.Service{}
		secret = corev1.Secret{}
		vm = virtv1.VirtualMachine{}
		vmi = virtv1.VirtualMachineInstance{}

		ownerRef = metav1.OwnerReference{
			APIVersion:         clv1alpha2.GroupVersion.String(),
			Kind:               "Instance",
			Name:               instance.GetName(),
			UID:                instance.GetUID(),
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
		}
	})

	JustBeforeEach(func() {
		reconciler = instctrl.InstanceReconciler{Client: clientBuilder.Build(), Scheme: scheme.Scheme}

		ctx, _ = clctx.InstanceInto(ctx, &instance)
		ctx, _ = clctx.TemplateInto(ctx, &template)
		ctx, _ = clctx.EnvironmentInto(ctx, &environment)
		ctx, _ = clctx.TenantInto(ctx, &tenant)
		err = reconciler.EnforceVMEnvironment(ctx)
	})

	Context("The environment mode is Standard", func() {
		BeforeEach(func() {
			environment.Mode = clv1alpha2.ModeStandard
		})
		It("Should enforce the cloud-init secret", func() {
			// Here, we only check the secret presence to assert the function execution, leaving the other assertions to the proper tests.
			Expect(reconciler.Get(ctx, objectName, &secret)).To(Succeed())
		})
	})

	Context("The environment mode is Exam", func() {
		BeforeEach(func() {
			environment.Mode = clv1alpha2.ModeExam
		})
		It("Should not enforce the cloud-init secret", func() {
			// Here, we only check the secret absence to assert the function execution, leaving the other assertions to the proper tests.
			Expect(reconciler.Get(ctx, objectName, &secret)).To(
				MatchError(kerrors.NewNotFound(corev1.Resource("secrets"), objectName.Name)),
			)
		})
	})

	Context("The environment mode is Exercise", func() {
		BeforeEach(func() {
			environment.Mode = clv1alpha2.ModeExercise
		})
		It("Should not enforce the cloud-init secret", func() {
			// Here, we only check the secret absence to assert the function execution, leaving the other assertions to the proper tests.
			Expect(reconciler.Get(ctx, objectName, &secret)).To(
				MatchError(kerrors.NewNotFound(corev1.Resource("secrets"), objectName.Name)),
			)
		})
	})

	It("Should enforce the environment exposition objects", func() {
		// Here, we only check the service presence to assert the function execution, leaving the other assertions to the proper tests.
		Expect(reconciler.Get(ctx, objectName, &svc)).To(Succeed())
	})

	Context("The environment is not persistent", func() {
		BeforeEach(func() { environment.Persistent = false })

		When("the VMI it is not yet present", func() {
			When("the instance is running", func() {
				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The VMI should be present and have the common attributes", func() {
					Expect(reconciler.Get(ctx, objectName, &vmi)).To(Succeed())
					Expect(vmi.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
					Expect(vmi.GetOwnerReferences()).To(ContainElement(ownerRef))
				})

				It("The VMI should be present and have the expected specs", func() {
					Expect(reconciler.Get(ctx, objectName, &vmi)).To(Succeed())
					// Here we overwrite the VMI resources, since they would have a different representation due to the
					// marshaling/unmarshaling process. Still, the correctness of the value is already checked with the
					// appropriate test case.
					vmi.Spec.Domain.Resources = forge.VirtualMachineResources(&environment)
					vmi.Spec.NodeSelector = map[string]string{}
					Expect(vmi.Spec).To(Equal(forge.VirtualMachineInstanceSpec(&instance, &environment)))
				})

				It("Should leave the instance phase unset", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseUnset))
				})
			})

			When("the instance is not running", func() {
				var notFoundError error

				BeforeEach(func() {
					instance.Spec.Running = false
					notFoundError = kerrors.NewNotFound(virtv1.Resource("virtualmachineinstances"), objectName.Name)
				})

				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The VMI should not be present", func() {
					Expect(reconciler.Get(ctx, objectName, &vmi)).To(MatchError(notFoundError))
				})

				It("Should set the instance phase to Off", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseOff))
				})
			})
		})

		When("the VMI is already present", func() {
			var existing virtv1.VirtualMachineInstance

			BeforeEach(func() {
				existing = virtv1.VirtualMachineInstance{
					ObjectMeta: forge.NamespacedNameToObjectMeta(objectName),
					Status:     virtv1.VirtualMachineInstanceStatus{Phase: virtv1.Running},
				}
				existing.SetCreationTimestamp(metav1.NewTime(time.Now()))
				clientBuilder.WithObjects(&existing)
			})

			When("the instance is running", func() {
				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The VMI should still be present and have the common attributes", func() {
					Expect(reconciler.Get(ctx, objectName, &vmi)).To(Succeed())
					Expect(vmi.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
					Expect(vmi.GetOwnerReferences()).To(ContainElement(ownerRef))
				})

				It("The VMI should still be present and have unmodified specs", func() {
					Expect(reconciler.Get(ctx, objectName, &vmi)).To(Succeed())
					Expect(vmi.Spec).To(Equal(existing.Spec))
				})

				It("Should set the correct instance phase", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseRunning))
				})
			})

			When("the instance is not running", func() {
				BeforeEach(func() { instance.Spec.Running = false })

				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The VMI should still be present and have unmodified specs", func() {
					Expect(reconciler.Get(ctx, objectName, &vmi)).To(Succeed())
					Expect(vmi.ObjectMeta.Labels).To(Equal(existing.ObjectMeta.Labels))
					Expect(vmi.Spec).To(Equal(existing.Spec))
					Expect(vmi.Status).To(Equal(existing.Status))
				})

				It("Should set the instance phase to Off", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseOff))
				})
			})
		})
	})

	Context("The environment is persistent", func() {
		BeforeEach(func() { environment.Persistent = true })

		ContextBody := func(envType clv1alpha2.EnvironmentType) {
			BeforeEach(func() {
				environment.EnvironmentType = envType
			})

			When("the VM is not yet present", func() {
				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The VM should be present and have the common attributes", func() {
					Expect(reconciler.Get(ctx, objectName, &vm)).To(Succeed())
					Expect(vm.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
					Expect(vm.GetOwnerReferences()).To(ContainElement(ownerRef))
				})

				It("The VM should be present and have the expected specs", func() {
					Expect(reconciler.Get(ctx, objectName, &vm)).To(Succeed())
					// Here we overwrite the VM resources, since they would have a different representation due to the
					// marshaling/unmarshaling process. Still, the correctness of the value is already checked with the
					// appropriate test case. Additionally, we also overwrite the running value, which is checked in a
					// different It clause.
					vm.Spec.Template.Spec.Domain.Resources = forge.VirtualMachineResources(&environment)
					vm.Spec.Running = nil
					vm.Spec.Template.Spec.NodeSelector = map[string]string{}
					Expect(vm.Spec).To(Equal(forge.VirtualMachineSpec(&instance, &environment)))
				})

				It("The VM should be present and with the running flag set", func() {
					Expect(reconciler.Get(ctx, objectName, &vm)).To(Succeed())
					Expect(*vm.Spec.Running).To(BeTrue())
				})

				It("Should leave the instance phase unset", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseUnset))
				})
			})

			WhenVMAlreadyPresentCase := func(running bool) {
				BeforeEach(func() {
					existing := virtv1.VirtualMachine{
						ObjectMeta: forge.NamespacedNameToObjectMeta(objectName),
						Spec:       virtv1.VirtualMachineSpec{Running: ptr.To(running)},
						Status:     virtv1.VirtualMachineStatus{PrintableStatus: virtv1.VirtualMachineStatusRunning},
					}
					existing.SetCreationTimestamp(metav1.NewTime(time.Now()))
					clientBuilder.WithObjects(&existing)
				})

				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The VM should still be present and have the common attributes", func() {
					Expect(reconciler.Get(ctx, objectName, &vm)).To(Succeed())
					Expect(vm.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
					Expect(vm.GetOwnerReferences()).To(ContainElement(ownerRef))
				})

				It("The VM should still be present and have unmodified specs", func() {
					Expect(reconciler.Get(ctx, objectName, &vm)).To(Succeed())
					// Here we overwrite the running value, as it is checked in a different It clause.
					vm.Spec.Running = nil
					Expect(vmi.Spec).To(Equal(virtv1.VirtualMachineInstanceSpec{}))
				})

				It("Should set the correct instance phase", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseRunning))
				})

				Context("The instance is running", func() {
					BeforeEach(func() { instance.Spec.Running = true })

					It("The VM should be present and with the running flag set", func() {
						Expect(reconciler.Get(ctx, objectName, &vm)).To(Succeed())
						Expect(*vm.Spec.Running).To(BeTrue())
					})
				})

				Context("The instance is not running", func() {
					BeforeEach(func() { instance.Spec.Running = false })

					It("The VM should be present and with the running flag not set", func() {
						Expect(reconciler.Get(ctx, objectName, &vm)).To(Succeed())
						Expect(*vm.Spec.Running).To(BeFalse())
					})
				})
			}

			When("the VM is already present and it is running", func() { WhenVMAlreadyPresentCase(true) })
			When("the VM is already present and it is not running", func() { WhenVMAlreadyPresentCase(false) })
		}

		Context("The environment type is VirtualMachine", func() { ContextBody(clv1alpha2.ClassVM) })
		Context("The environment type is CloudVM", func() { ContextBody(clv1alpha2.ClassCloudVM) })
	})
})
