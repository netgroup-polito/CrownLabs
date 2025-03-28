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

	Describe("The forge.ObjectMetaWithSuffix function", func() {
		const Suffix = "prime"

		type ObjectMetaCaseWithSuffix struct {
			InstanceNamespace string
			InstanceName      string
			ExpectedOutput    metav1.ObjectMeta
		}

		DescribeTable("Correctly returns the expected object meta",
			func(c ObjectMetaCaseWithSuffix) {
				Expect(forge.ObjectMetaWithSuffix(ForgeInstance(c.InstanceNamespace, c.InstanceName), Suffix)).To(Equal(c.ExpectedOutput))
			},
			Entry("When the instance name does not contain dots", ObjectMetaCaseWithSuffix{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kubernetes-1234",
				ExpectedOutput:    metav1.ObjectMeta{Namespace: "workspace-netgroup", Name: "kubernetes-1234-prime"},
			}),
			Entry("When the instance name does contain dots", ObjectMetaCaseWithSuffix{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kuber.netes.1234",
				ExpectedOutput:    metav1.ObjectMeta{Namespace: "workspace-netgroup", Name: "kuber-netes-1234-prime"},
			}),
		)
	})

	Describe("The forge.NamespacedName function", func() {
		type NamespaceNameCase struct {
			InstanceNamespace string
			InstanceName      string
			ExpectedOutput    types.NamespacedName
		}

		DescribeTable("Correctly returns the expected object meta",
			func(c NamespaceNameCase) {
				Expect(forge.NamespacedName(ForgeInstance(c.InstanceNamespace, c.InstanceName))).To(Equal(c.ExpectedOutput))
			},
			Entry("When the instance name does not contain dots", NamespaceNameCase{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kubernetes-1234",
				ExpectedOutput:    types.NamespacedName{Namespace: "workspace-netgroup", Name: "kubernetes-1234"},
			}),
			Entry("When the instance name does contain dots", NamespaceNameCase{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kuber.netes.1234",
				ExpectedOutput:    types.NamespacedName{Namespace: "workspace-netgroup", Name: "kuber-netes-1234"},
			}),
		)
	})

	Describe("The forge.NamespacedNameWithSuffix function", func() {
		const Suffix = "prime"

		type NamespaceNameWithSuffixCase struct {
			InstanceNamespace string
			InstanceName      string
			ExpectedOutput    types.NamespacedName
		}

		DescribeTable("Correctly returns the expected object meta",
			func(c NamespaceNameWithSuffixCase) {
				Expect(forge.NamespacedNameWithSuffix(ForgeInstance(c.InstanceNamespace, c.InstanceName), Suffix)).To(Equal(c.ExpectedOutput))
			},
			Entry("When the instance name does not contain dots", NamespaceNameWithSuffixCase{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kubernetes-1234",
				ExpectedOutput:    types.NamespacedName{Namespace: "workspace-netgroup", Name: "kubernetes-1234-prime"},
			}),
			Entry("When the instance name does contain dots", NamespaceNameWithSuffixCase{
				InstanceNamespace: "workspace-netgroup",
				InstanceName:      "kuber.netes.1234",
				ExpectedOutput:    types.NamespacedName{Namespace: "workspace-netgroup", Name: "kuber-netes-1234-prime"},
			}),
		)
	})

	Describe("The forge.NamespacedNameToObjectMeta function", func() {
		var (
			namespacedName types.NamespacedName
			objectMeta     metav1.ObjectMeta
		)

		BeforeEach(func() {
			namespacedName = types.NamespacedName{Name: "kubernetes-0000", Namespace: "workspace-netgroup"}
		})
		JustBeforeEach(func() { objectMeta = forge.NamespacedNameToObjectMeta(namespacedName) })

		It("Should have a matching name", func() { Expect(objectMeta.Name).To(BeIdenticalTo(namespacedName.Name)) })
		It("Should have a matching namespace", func() { Expect(objectMeta.Namespace).To(BeIdenticalTo(namespacedName.Namespace)) })
	})

	Describe("The forge.NamespacedNameFromSharedVolume function", func() {
		var (
			shvol          clv1alpha2.SharedVolume
			namespacedName types.NamespacedName
		)

		BeforeEach(func() {
			shvol = clv1alpha2.SharedVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
			}
		})
		JustBeforeEach(func() {
			namespacedName = forge.NamespacedNameFromSharedVolume(&shvol)
		})

		It("Should have a matching name", func() {
			Expect(namespacedName.Name).To(Equal(shvol.Name))
		})
		It("Should have a matching namespace", func() {
			Expect(namespacedName.Namespace).To(Equal(shvol.Namespace))
		})

	})

	Describe("The forge.NamespacedNameFromMount function", func() {
		var (
			mountInfo      clv1alpha2.SharedVolumeMountInfo
			namespacedName types.NamespacedName
		)

		BeforeEach(func() {
			mountInfo = clv1alpha2.SharedVolumeMountInfo{
				SharedVolumeRef: clv1alpha2.GenericRef{
					Name:      "name",
					Namespace: "namespace",
				},
			}
		})
		JustBeforeEach(func() {
			namespacedName = forge.NamespacedNameFromMount(mountInfo)
		})

		It("Should have a matching name", func() {
			Expect(namespacedName.Name).To(Equal(mountInfo.SharedVolumeRef.Name))
		})
		It("Should have a matching namespace", func() {
			Expect(namespacedName.Namespace).To(Equal(mountInfo.SharedVolumeRef.Namespace))
		})

	})
})
