package forge_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Utils forging", func() {

	ForgeInstance := func(namespace, name string) *clv1alpha2.Instance {
		return &clv1alpha2.Instance{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}
	}

	Describe("The forge.ObjectMeta function", func() {
		type ObjectMetaCase struct {
			InstanceNamespace string
			InstanceName      string
			ExpectedOutput    metav1.ObjectMeta
		}

		DescribeTable("Correctly returns the expected object meta",
			func(c ObjectMetaCase) {
				Expect(forge.ObjectMeta(ForgeInstance(c.InstanceNamespace, c.InstanceName))).To(Equal(c.ExpectedOutput))
			},
			Entry("When the instance name does not contain dots", ObjectMetaCase{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kubernetes-1234",
				ExpectedOutput:    metav1.ObjectMeta{Namespace: "workspace-netgroup", Name: "kubernetes-1234"},
			}),
			Entry("When the instance name does contain dots", ObjectMetaCase{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kuber.netes.1234",
				ExpectedOutput:    metav1.ObjectMeta{Namespace: "workspace-netgroup", Name: "kuber-netes-1234"},
			}),
		)
	})

	Describe("The forge.NamespaceName function", func() {

		type ObjectMetaCase struct {
			InstanceNamespace string
			InstanceName      string
			ExpectedOutput    types.NamespacedName
		}

		DescribeTable("Correctly returns the expected object meta",
			func(c ObjectMetaCase) {
				Expect(forge.NamespaceName(ForgeInstance(c.InstanceNamespace, c.InstanceName))).To(Equal(c.ExpectedOutput))
			},
			Entry("When the instance name does not contain dots", ObjectMetaCase{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kubernetes-1234",
				ExpectedOutput:    types.NamespacedName{Namespace: "workspace-netgroup", Name: "kubernetes-1234"},
			}),
			Entry("When the instance name does contain dots", ObjectMetaCase{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kuber.netes.1234",
				ExpectedOutput:    types.NamespacedName{Namespace: "workspace-netgroup", Name: "kuber-netes-1234"},
			}),
		)
	})
})
