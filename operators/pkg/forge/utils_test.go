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

	Describe("The forge.NamespacedNameFromObject function", func() {
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
			namespacedName = forge.NamespacedNameFromObject(&shvol)
		})

		It("Should have a matching name", func() {
			Expect(namespacedName.Name).To(Equal(shvol.Name))
		})
		It("Should have a matching namespace", func() {
			Expect(namespacedName.Namespace).To(Equal(shvol.Namespace))
		})

	})

	Describe("The forge.NamespacedNameFromGenericRef function", func() {
		var (
			ref            clv1alpha2.GenericRef
			namespacedName types.NamespacedName
		)

		BeforeEach(func() {
			ref = clv1alpha2.GenericRef{
				Name:      "name",
				Namespace: "namespace",
			}
		})
		JustBeforeEach(func() {
			namespacedName = forge.NamespacedNameFromGenericRef(ref)
		})

		It("Should have a matching name", func() {
			Expect(namespacedName.Name).To(Equal(ref.Name))
		})
		It("Should have a matching namespace", func() {
			Expect(namespacedName.Namespace).To(Equal(ref.Namespace))
		})

	})

	Describe("The forge.LastCharsOf function", func() {
		var (
			actual string
		)

		type StringTestCase struct {
			TestCaseName   string
			OriginalString string
			ManyChars      int
			ExpectedString string
		}

		WhenBody := func(c StringTestCase) {
			JustBeforeEach(func() {
				actual = forge.LastCharsOf(c.OriginalString, c.ManyChars)
			})

			It(c.TestCaseName, func() {
				Expect(actual).To(Equal(c.ExpectedString))
			})
		}

		When("Many is not valid (less than 0)", func() {
			WhenBody(StringTestCase{
				TestCaseName:   "Should return an empty string",
				OriginalString: "abcdef",
				ManyChars:      -5,
				ExpectedString: "",
			})
		})

		When("Many is zero", func() {
			WhenBody(StringTestCase{
				TestCaseName:   "Should return an empty string",
				OriginalString: "abcdef",
				ManyChars:      0,
				ExpectedString: "",
			})
		})

		When("Many is smaller than the length of the string", func() {
			WhenBody(StringTestCase{
				TestCaseName:   "Should return the right substring",
				OriginalString: "abcdef",
				ManyChars:      5,
				ExpectedString: "bcdef",
			})
		})

		When("Many is bigger than the length of the string", func() {
			WhenBody(StringTestCase{
				TestCaseName:   "Should return the whole string",
				OriginalString: "abcdef",
				ManyChars:      100,
				ExpectedString: "abcdef",
			})
		})
	})

	Describe("The forge.MapFromKVString function", func() {
		var (
			result map[string]string
			err    error
		)

		When("An empty string is provided", func() {
			JustBeforeEach(func() {
				result, err = forge.MapFromKVString("")
			})

			It("Should return an empty map", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})

		When("A single key=value pair is provided", func() {
			JustBeforeEach(func() {
				result, err = forge.MapFromKVString("metallb.universe.tf/ip-pool=public")
			})

			It("Should parse it correctly", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(HaveLen(1))
				Expect(result).To(HaveKeyWithValue("metallb.universe.tf/ip-pool", "public"))
			})
		})

		When("Multiple key=value pairs are provided", func() {
			JustBeforeEach(func() {
				result, err = forge.MapFromKVString("metallb.universe.tf/allow-shared-ip=pe,metallb.universe.tf/address-pool=public")
			})

			It("Should parse all pairs correctly", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(HaveLen(2))
				Expect(result).To(HaveKeyWithValue("metallb.universe.tf/allow-shared-ip", "pe"))
				Expect(result).To(HaveKeyWithValue("metallb.universe.tf/address-pool", "public"))
			})
		})

		When("A pair without '=' is provided", func() {
			JustBeforeEach(func() {
				result, err = forge.MapFromKVString("invalidformat")
			})

			It("Should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		When("A pair with an empty key is provided", func() {
			JustBeforeEach(func() {
				result, err = forge.MapFromKVString("=value")
			})

			It("Should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		When("A value contains '='", func() {
			JustBeforeEach(func() {
				result, err = forge.MapFromKVString("key=val=ue")
			})

			It("Should treat only the first '=' as separator", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(HaveKeyWithValue("key", "val=ue"))
			})
		})

		When("Pairs have surrounding whitespace", func() {
			JustBeforeEach(func() {
				result, err = forge.MapFromKVString(" key1 = val1 , key2 = val2 ")
			})

			It("Should trim whitespace from keys and values", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(HaveLen(2))
				Expect(result).To(HaveKeyWithValue("key1", "val1"))
				Expect(result).To(HaveKeyWithValue("key2", "val2"))
			})
		})
	})
})
