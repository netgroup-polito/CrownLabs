package forge_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	virtv1 "kubevirt.io/client-go/api/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("VirtualMachines forging", func() {

	Describe("The forge.VMReadinessProbe function", func() {
		type VMReadinessProbeCase struct {
			Environment clv1alpha2.Environment
			Port        int
		}

		DescribeTable("Correctly returns the expected readiness probe",
			func(c VMReadinessProbeCase) {
				output := forge.VMReadinessProbe(&c.Environment)

				Expect(output.Handler).To(Equal(virtv1.Handler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.FromInt(c.Port),
					},
				}))

			},
			Entry("When the environment has the GUI enabled", VMReadinessProbeCase{
				Environment: clv1alpha2.Environment{GuiEnabled: true},
				Port:        forge.ServiceNoVNCPort,
			}),
			Entry("When the environment has not the GUI enabled", VMReadinessProbeCase{
				Environment: clv1alpha2.Environment{GuiEnabled: false},
				Port:        forge.ServiceSSHPort,
			}),
		)
	})

	Describe("The forge.DataVolumeTemplate function", func() {
		var (
			name               string
			environment        clv1alpha2.Environment
			dataVolumeTemplate virtv1.DataVolumeTemplateSpec
		)

		BeforeEach(func() {
			name = "kubernetes-volume"
			environment = clv1alpha2.Environment{
				Image: "target/image:v1",
				Resources: clv1alpha2.EnvironmentResources{
					Disk: resource.MustParse("20Gi"),
				},
			}
		})

		JustBeforeEach(func() {
			dataVolumeTemplate = forge.DataVolumeTemplate(name, &environment)
		})

		Context("The DataVolumeTemplate is forged", func() {
			It("Should have the correct name", func() {
				Expect(dataVolumeTemplate.GetName()).To(BeIdenticalTo(name))
			})

			It("Should target the correct image registry", func() {
				Expect(dataVolumeTemplate.Spec.Source.Registry.URL).To(
					BeIdenticalTo("docker://target/image:v1"))
			})

			It("Should request the correct disk size", func() {
				Expect(dataVolumeTemplate.Spec.PVC.Resources.Requests).To(Equal(
					corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("20Gi")}))
			})
		})
	})

})
