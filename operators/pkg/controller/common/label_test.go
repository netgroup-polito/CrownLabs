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

package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Label", func() {
	Describe("NewLabel", func() {
		It("should create a new Label instance with the given key and value", func() {
			label := NewLabel("test-key", "test-value")
			Expect(label.key).To(Equal("test-key"))
			Expect(label.value).To(Equal("test-value"))
		})
	})

	Describe("ParseLabel", func() {
		Context("with a valid label string", func() {
			It("should parse correctly the key=value format", func() {
				label, err := ParseLabel("test-key=test-value")
				Expect(err).NotTo(HaveOccurred())
				Expect(label.key).To(Equal("test-key"))
				Expect(label.value).To(Equal("test-value"))
			})
		})

		Context("with an invalid label string", func() {
			It("should return an error for invalid format", func() {
				_, err := ParseLabel("invalid-format")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid label format"))
			})
		})
	})

	Describe("GetKey", func() {
		It("should return the label key", func() {
			label := NewLabel("test-key", "test-value")
			Expect(label.GetKey()).To(Equal("test-key"))
		})
	})

	Describe("GetValue", func() {
		It("should return the label value", func() {
			label := NewLabel("test-key", "test-value")
			Expect(label.GetValue()).To(Equal("test-value"))
		})
	})

	Describe("GetPredicate", func() {
		It("should create a valid predicate with the label", func() {
			label := NewLabel("test-key", "test-value")
			pred, err := label.GetPredicate()
			Expect(err).NotTo(HaveOccurred())
			Expect(pred).NotTo(BeNil())
		})
	})

	Describe("IsIncluded", func() {
		var label KVLabel

		BeforeEach(func() {
			label = NewLabel("test-key", "test-value")
		})

		Context("when labels map is nil", func() {
			It("should return false", func() {
				Expect(label.IsIncluded(nil)).To(BeFalse())
			})
		})

		Context("when label key is not in map", func() {
			It("should return false", func() {
				labels := map[string]string{
					"other-key": "other-value",
				}
				Expect(label.IsIncluded(labels)).To(BeFalse())
			})
		})

		Context("when label is in map but value doesn't match", func() {
			It("should return false", func() {
				labels := map[string]string{
					"test-key": "different-value",
				}
				Expect(label.IsIncluded(labels)).To(BeFalse())
			})
		})

		Context("when label is in map with matching value", func() {
			It("should return true", func() {
				labels := map[string]string{
					"test-key": "test-value",
				}
				Expect(label.IsIncluded(labels)).To(BeTrue())
			})
		})
	})
})
