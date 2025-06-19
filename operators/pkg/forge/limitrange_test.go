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
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Limit range spec forging", func() {
	Describe("The forge.SandboxLimitRangeSpec function", func() {
		var spec corev1.LimitRangeSpec

		JustBeforeEach(func() { spec = forge.SandboxLimitRangeSpec() })

		It("Should have a single limit entry of type container", func() {
			Expect(spec.Limits).To(HaveLen(1))
			Expect(spec.Limits[0].Type).To(BeIdenticalTo(corev1.LimitTypeContainer))
		})

		It("Should set the correct default requests", func() {
			Expect(spec.Limits[0].DefaultRequest[corev1.ResourceCPU]).To(WithTransform(
				func(q resource.Quantity) string { return q.String() }, Equal("10m")))
			Expect(spec.Limits[0].DefaultRequest[corev1.ResourceMemory]).To(WithTransform(
				func(q resource.Quantity) string { return q.String() }, Equal("50M")))
		})

		It("Should set the correct default limits", func() {
			Expect(spec.Limits[0].Default[corev1.ResourceCPU]).To(WithTransform(
				func(q resource.Quantity) string { return q.String() }, Equal("100m")))
			Expect(spec.Limits[0].Default[corev1.ResourceMemory]).To(WithTransform(
				func(q resource.Quantity) string { return q.String() }, Equal("250M")))
		})
	})
})
