package forge_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Labels forging", func() {

	Describe("The forge.InstanceLabels function", func() {
		var template clv1alpha2.Template

		const (
			templateName      = "kubernetes"
			templateNamespace = "workspace-netgroup"
			workspaceName     = "netgroup"
		)

		type InstanceLabelsCase struct {
			Input           map[string]string
			ExpectedOutput  map[string]string
			ExpectedUpdated bool
		}

		BeforeEach(func() {
			template = clv1alpha2.Template{
				ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateNamespace},
				Spec: clv1alpha2.TemplateSpec{
					WorkspaceRef: clv1alpha2.GenericRef{Name: workspaceName},
				},
			}
		})

		DescribeTable("Correctly populates the labels set",
			func(c InstanceLabelsCase) {
				output, updated := forge.InstanceLabels(c.Input, &template)

				Expect(output).To(Equal(c.ExpectedOutput))
				Expect(updated).To(BeIdenticalTo(c.ExpectedUpdated))
			},
			Entry("When the input labels map is nil", InstanceLabelsCase{
				Input: nil,
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/workspace":  workspaceName,
					"crownlabs.polito.it/template":   templateName,
				},
				ExpectedUpdated: true,
			}),
			Entry("When the input labels map already contains the expected values", InstanceLabelsCase{
				Input: map[string]string{
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/workspace":  workspaceName,
					"crownlabs.polito.it/template":   templateName,
					"user/key":                       "user/value",
				},
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/workspace":  workspaceName,
					"crownlabs.polito.it/template":   templateName,
					"user/key":                       "user/value",
				},
				ExpectedUpdated: false,
			}),
			Entry("When the input labels map contains only part of the expected values", InstanceLabelsCase{
				Input: map[string]string{
					"crownlabs.polito.it/workspace": workspaceName,
					"user/key":                      "user/value",
				},
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/workspace":  workspaceName,
					"crownlabs.polito.it/template":   templateName,
					"user/key":                       "user/value",
				},
				ExpectedUpdated: true,
			}),
		)
	})
})
