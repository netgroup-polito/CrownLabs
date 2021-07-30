package forge_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
			Entry("When the Environment is of type Container", ServiceSpecCase{
				Mutator: func(env *clv1alpha2.Environment) *clv1alpha2.Environment {
					env.EnvironmentType = clv1alpha2.ClassContainer
					env.GuiEnabled = true
					return env
				},
				Expected: []corev1.ServicePort{
					{Name: forge.GUIPortName, Protocol: corev1.ProtocolTCP, Port: forge.GUIPortNumber, TargetPort: intstr.FromInt(forge.GUIPortNumber)},
					{Name: forge.MyDrivePortName, Protocol: corev1.ProtocolTCP, Port: forge.MyDrivePortNumber, TargetPort: intstr.FromInt(forge.MyDrivePortNumber)},
				},
			}),
		)
	})
})
