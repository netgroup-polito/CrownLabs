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

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Native VNC forging", func() {

	Describe("The forge.VirtualMachineLabels function", func() {
		const operatorSelectorKey = "crownlabs.polito.it/operator-selector"

		var (
			environment clv1alpha2.Environment
			labels      map[string]string
		)

		BeforeEach(func() {
			environment = clv1alpha2.Environment{}
		})

		JustBeforeEach(func() {
			labels = forge.VirtualMachineLabels(&environment, map[string]string{"foo": "bar"},
				map[string]string{operatorSelectorKey: "staging-1184"})
		})

		When("the environment has not the GUI enabled", func() {
			BeforeEach(func() { environment.GuiEnabled = false })

			It("Should mark the VMI for native VNC and scope it to the operator", func() {
				Expect(labels).To(Equal(map[string]string{
					"foo":                   "bar",
					forge.LabelNativeVNCKey: "true",
					operatorSelectorKey:     "staging-1184",
				}))
			})
		})

		When("the environment has the GUI enabled", func() {
			BeforeEach(func() { environment.GuiEnabled = true })

			It("Should leave the labels untouched", func() {
				Expect(labels).To(Equal(map[string]string{"foo": "bar"}))
			})
		})
	})
})
