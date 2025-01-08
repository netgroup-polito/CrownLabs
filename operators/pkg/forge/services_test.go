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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Services forging", func() {

	Describe("The forge.ServiceSpec function", func() {
		var (
			instance    clv1alpha2.Instance
			environment clv1alpha2.Environment
			spec        corev1.ServiceSpec
		)

		const (
			instanceName      = "kubernetes-0000"
			instanceNamespace = "tenant-tester"
			templateName      = "kubernetes"
			templateNamespace = "workspace-netgroup"
			environmentName   = "control-plane"
			tenantName        = "tester"
		)

		type ServiceSpecCase struct {
			Mutator  func(*clv1alpha2.Environment) *clv1alpha2.Environment
			Expected []corev1.ServicePort
		}

		BeforeEach(func() {
			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
				Spec: clv1alpha2.InstanceSpec{
					Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
					Tenant:   clv1alpha2.GenericRef{Name: tenantName},
				},
			}
			environment = clv1alpha2.Environment{Name: environmentName}
		})

		JustBeforeEach(func() {
			spec = forge.ServiceSpec(&instance, &environment)
		})

		Describe("Correctly populates the common fields", func() {
			It("Should set service type to ClusterID", func() {
				Expect(spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			})

			It("Should configure the expected selector labels", func() {
				Expect(spec.Selector).To(Equal(forge.InstanceSelectorLabels(&instance)))
			})
		})

		DescribeTable("Correctly configure the service ports",
			func(c ServiceSpecCase) {
				Expect(forge.ServiceSpec(&instance, c.Mutator(&environment)).Ports).To(Equal(c.Expected))
			},
			Entry("When the Environment is of type VM, without GUI", ServiceSpecCase{
				Mutator: func(env *clv1alpha2.Environment) *clv1alpha2.Environment {
					env.EnvironmentType = clv1alpha2.ClassVM
					env.GuiEnabled = false
					return env
				},
				Expected: []corev1.ServicePort{
					{Name: forge.SSHPortName, Protocol: corev1.ProtocolTCP, Port: forge.SSHPortNumber, TargetPort: intstr.FromInt(forge.SSHPortNumber)},
				},
			}),
			Entry("When the Environment is of type VM, with GUI", ServiceSpecCase{
				Mutator: func(env *clv1alpha2.Environment) *clv1alpha2.Environment {
					env.EnvironmentType = clv1alpha2.ClassVM
					env.GuiEnabled = true
					return env
				},
				Expected: []corev1.ServicePort{
					{Name: forge.SSHPortName, Protocol: corev1.ProtocolTCP, Port: forge.SSHPortNumber, TargetPort: intstr.FromInt(forge.SSHPortNumber)},
					{Name: forge.GUIPortName, Protocol: corev1.ProtocolTCP, Port: forge.GUIPortNumber, TargetPort: intstr.FromInt(forge.GUIPortNumber)},
				},
			}),
			Entry("When the Environment is of type CloudVM, without GUI", ServiceSpecCase{
				Mutator: func(env *clv1alpha2.Environment) *clv1alpha2.Environment {
					env.EnvironmentType = clv1alpha2.ClassCloudVM
					env.GuiEnabled = false
					return env
				},
				Expected: []corev1.ServicePort{
					{Name: forge.SSHPortName, Protocol: corev1.ProtocolTCP, Port: forge.SSHPortNumber, TargetPort: intstr.FromInt(forge.SSHPortNumber)},
				},
			}),
			Entry("When the Environment is of type CloudVM, with GUI", ServiceSpecCase{
				Mutator: func(env *clv1alpha2.Environment) *clv1alpha2.Environment {
					env.EnvironmentType = clv1alpha2.ClassCloudVM
					env.GuiEnabled = true
					return env
				},
				Expected: []corev1.ServicePort{
					{Name: forge.SSHPortName, Protocol: corev1.ProtocolTCP, Port: forge.SSHPortNumber, TargetPort: intstr.FromInt(forge.SSHPortNumber)},
					{Name: forge.GUIPortName, Protocol: corev1.ProtocolTCP, Port: forge.GUIPortNumber, TargetPort: intstr.FromInt(forge.GUIPortNumber)},
				},
			}),
			Entry("When the Environment is of type Container in standard mode", ServiceSpecCase{
				Mutator: func(env *clv1alpha2.Environment) *clv1alpha2.Environment {
					env.EnvironmentType = clv1alpha2.ClassContainer
					env.GuiEnabled = true
					env.Mode = clv1alpha2.ModeStandard
					return env
				},
				Expected: []corev1.ServicePort{
					{Name: forge.GUIPortName, Protocol: corev1.ProtocolTCP, Port: forge.GUIPortNumber, TargetPort: intstr.FromInt(forge.GUIPortNumber)},
					{Name: forge.MyDrivePortName, Protocol: corev1.ProtocolTCP, Port: forge.MyDrivePortNumber, TargetPort: intstr.FromInt(forge.MyDrivePortNumber)},
					{Name: forge.MetricsPortName, Protocol: corev1.ProtocolTCP, Port: forge.MetricsPortNumber, TargetPort: intstr.FromInt(forge.MetricsPortNumber)},
				},
			}),
			Entry("When the Environment is a Container", ServiceSpecCase{
				Mutator: func(env *clv1alpha2.Environment) *clv1alpha2.Environment {
					env.EnvironmentType = clv1alpha2.ClassContainer
					env.GuiEnabled = true
					return env
				},
				Expected: []corev1.ServicePort{
					{Name: forge.GUIPortName, Protocol: corev1.ProtocolTCP, Port: forge.GUIPortNumber, TargetPort: intstr.FromInt(forge.GUIPortNumber)},
					{Name: forge.MetricsPortName, Protocol: corev1.ProtocolTCP, Port: forge.MetricsPortNumber, TargetPort: intstr.FromInt(forge.MetricsPortNumber)},
				},
			}),
			Entry("When the Environment is of type Container in exam mode", ServiceSpecCase{
				Mutator: func(env *clv1alpha2.Environment) *clv1alpha2.Environment {
					env.EnvironmentType = clv1alpha2.ClassContainer
					env.GuiEnabled = true
					env.Mode = clv1alpha2.ModeExam
					return env
				},
				Expected: []corev1.ServicePort{
					{Name: forge.GUIPortName, Protocol: corev1.ProtocolTCP, Port: forge.GUIPortNumber, TargetPort: intstr.FromInt(forge.GUIPortNumber)},
					{Name: forge.MetricsPortName, Protocol: corev1.ProtocolTCP, Port: forge.MetricsPortNumber, TargetPort: intstr.FromInt(forge.MetricsPortNumber)},
				},
			}),
			Entry("When the Environment is of type Container in exercise mode", ServiceSpecCase{
				Mutator: func(env *clv1alpha2.Environment) *clv1alpha2.Environment {
					env.EnvironmentType = clv1alpha2.ClassContainer
					env.GuiEnabled = true
					env.Mode = clv1alpha2.ModeExercise
					return env
				},
				Expected: []corev1.ServicePort{
					{Name: forge.GUIPortName, Protocol: corev1.ProtocolTCP, Port: forge.GUIPortNumber, TargetPort: intstr.FromInt(forge.GUIPortNumber)},
					{Name: forge.MetricsPortName, Protocol: corev1.ProtocolTCP, Port: forge.MetricsPortNumber, TargetPort: intstr.FromInt(forge.MetricsPortNumber)},
				},
			}),
		)
	})
})
