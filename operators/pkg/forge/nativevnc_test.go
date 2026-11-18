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

		When("the environment is served by native VNC", func() {
			BeforeEach(func() {
				environment.EnvironmentType = clv1alpha2.ClassVM
				environment.GuiMode = clv1alpha2.GuiModeNative
			})

			It("Should mark the VMI for native VNC and scope it to the operator", func() {
				Expect(labels).To(Equal(map[string]string{
					"foo":                   "bar",
					forge.LabelNativeVNCKey: "true",
					operatorSelectorKey:     "staging-1184",
				}))
			})
		})

		When("the environment has the web desktop", func() {
			BeforeEach(func() {
				environment.EnvironmentType = clv1alpha2.ClassVM
				environment.GuiMode = clv1alpha2.GuiModeWeb
			})

			It("Should leave the labels untouched", func() {
				Expect(labels).To(Equal(map[string]string{"foo": "bar"}))
			})
		})

		When("the environment has no graphical desktop", func() {
			BeforeEach(func() {
				environment.EnvironmentType = clv1alpha2.ClassVM
			})

			It("Should leave the labels untouched", func() {
				Expect(labels).To(Equal(map[string]string{"foo": "bar"}))
			})
		})
	})

	Describe("The forge.EnvironmentGuiMode function", func() {
		DescribeTable("Returns how the graphical desktop is reached",
			func(environment clv1alpha2.Environment, expected clv1alpha2.GuiMode) {
				Expect(forge.EnvironmentGuiMode(&environment)).To(Equal(expected))
			},
			Entry("guiMode takes precedence over guiEnabled",
				clv1alpha2.Environment{GuiMode: clv1alpha2.GuiModeNative, GuiEnabled: true}, clv1alpha2.GuiModeNative),
			Entry("guiMode Web", clv1alpha2.Environment{GuiMode: clv1alpha2.GuiModeWeb}, clv1alpha2.GuiModeWeb),
			Entry("unset guiMode, guiEnabled true", clv1alpha2.Environment{GuiEnabled: true}, clv1alpha2.GuiModeWeb),
			Entry("unset guiMode, guiEnabled false", clv1alpha2.Environment{}, clv1alpha2.GuiMode("")),
		)
	})

	Describe("The forge.NativeVNCEnabled function", func() {
		DescribeTable("Tells whether QEMU's VNC server serves the desktop",
			func(envType clv1alpha2.EnvironmentType, mode clv1alpha2.GuiMode, expected bool) {
				environment := clv1alpha2.Environment{EnvironmentType: envType, GuiMode: mode}
				Expect(forge.NativeVNCEnabled(&environment)).To(Equal(expected))
			},
			Entry("VM with native VNC", clv1alpha2.ClassVM, clv1alpha2.GuiModeNative, true),
			Entry("CloudVM with native VNC", clv1alpha2.ClassCloudVM, clv1alpha2.GuiModeNative, true),
			Entry("LocalVM with native VNC", clv1alpha2.ClassLocalVM, clv1alpha2.GuiModeNative, true),
			Entry("VM with the web desktop", clv1alpha2.ClassVM, clv1alpha2.GuiModeWeb, false),
			Entry("VM with no graphical desktop", clv1alpha2.ClassVM, clv1alpha2.GuiMode(""), false),
			Entry("Container with native VNC", clv1alpha2.ClassContainer, clv1alpha2.GuiModeNative, false),
		)
	})
})
