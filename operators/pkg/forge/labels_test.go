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

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Labels forging", func() {

	const (
		instanceName      = "kubernetes-0000"
		instanceNamespace = "tenant-tester"
		templateName      = "kubernetes"
		templateNamespace = "workspace-netgroup"
		tenantName        = "tester"
		workspaceName     = "netgroup"
		environmentName   = "control-plane"
		statusCheckURL    = "https://some/url"
		componentName     = "crownlabs-component"
		sandboxName       = "sandbox"
	)

	Describe("The forge.InstanceLabels function", func() {
		var template clv1alpha2.Template
		var instance clv1alpha2.Instance

		type InstanceLabelsCase struct {
			Input           map[string]string
			ExpectedOutput  map[string]string
			ExpectedUpdated bool
		}

		type InstancePersistentLabelCase struct {
			EnvironmentList []clv1alpha2.Environment
			ExpectedValue   string
		}

		type NodeSelectorEnabledLabelCase struct {
			EnvironmentList      []clv1alpha2.Environment
			InstanceNodeSelector map[string]string
			ExpectedValue        string
		}

		type InstanceAutomationLabelCase struct {
			Input                     map[string]string
			InstanceCustomizationUrls *clv1alpha2.InstanceCustomizationUrls
			ExpectedValue             string
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
				output, updated := forge.InstanceLabels(c.Input, &template, nil)

				Expect(output).To(Equal(c.ExpectedOutput))
				Expect(updated).To(BeIdenticalTo(c.ExpectedUpdated))
			},
			Entry("When the input labels map is nil", InstanceLabelsCase{
				Input: nil,
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by":        "instance",
					"crownlabs.polito.it/workspace":         workspaceName,
					"crownlabs.polito.it/template":          templateName,
					"crownlabs.polito.it/persistent":        "false",
					"crownlabs.polito.it/has-node-selector": "false",
				},
				ExpectedUpdated: true,
			}),
			Entry("When the input labels map already contains the expected values", InstanceLabelsCase{
				Input: map[string]string{
					"crownlabs.polito.it/managed-by":        "instance",
					"crownlabs.polito.it/workspace":         workspaceName,
					"crownlabs.polito.it/template":          templateName,
					"crownlabs.polito.it/persistent":        "false",
					"user/key":                              "user/value",
					"crownlabs.polito.it/has-node-selector": "false",
				},
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by":        "instance",
					"crownlabs.polito.it/workspace":         workspaceName,
					"crownlabs.polito.it/template":          templateName,
					"crownlabs.polito.it/persistent":        "false",
					"user/key":                              "user/value",
					"crownlabs.polito.it/has-node-selector": "false",
				},
				ExpectedUpdated: false,
			}),
			Entry("When the input labels map contains only part of the expected values", InstanceLabelsCase{
				Input: map[string]string{
					"crownlabs.polito.it/workspace": workspaceName,
					"user/key":                      "user/value",
				},
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by":        "instance",
					"crownlabs.polito.it/workspace":         workspaceName,
					"crownlabs.polito.it/template":          templateName,
					"crownlabs.polito.it/persistent":        "false",
					"user/key":                              "user/value",
					"crownlabs.polito.it/has-node-selector": "false",
				},
				ExpectedUpdated: true,
			}),
		)

		DescribeTable("Correctly configures the persistent label",
			func(c InstancePersistentLabelCase) {
				template.Spec.EnvironmentList = c.EnvironmentList
				output, _ := forge.InstanceLabels(map[string]string{}, &template, nil)
				Expect(output).To(HaveKeyWithValue("crownlabs.polito.it/persistent", c.ExpectedValue))
			},
			Entry("When a single, non-persistent environment is present", InstancePersistentLabelCase{
				EnvironmentList: []clv1alpha2.Environment{{Persistent: false}},
				ExpectedValue:   "false",
			}),
			Entry("When multiple, non-persistent environments are present", InstancePersistentLabelCase{
				EnvironmentList: []clv1alpha2.Environment{{Persistent: false}, {Persistent: false}},
				ExpectedValue:   "false",
			}),
			Entry("When a single, persistent environment is present", InstancePersistentLabelCase{
				EnvironmentList: []clv1alpha2.Environment{{Persistent: true}},
				ExpectedValue:   "true",
			}),
			Entry("When multiple, persistent environments are present", InstancePersistentLabelCase{
				EnvironmentList: []clv1alpha2.Environment{{Persistent: true}, {Persistent: true}},
				ExpectedValue:   "true",
			}),
			Entry("When multiple, mixed environments are present", InstancePersistentLabelCase{
				EnvironmentList: []clv1alpha2.Environment{{Persistent: false}, {Persistent: true}, {Persistent: false}},
				ExpectedValue:   "true",
			}),
		)

		DescribeTable("Correctly configures the node selection presence label",
			func(c NodeSelectorEnabledLabelCase) {
				template.Spec.EnvironmentList = c.EnvironmentList
				instance.Spec.NodeSelector = c.InstanceNodeSelector
				output, _ := forge.InstanceLabels(map[string]string{}, &template, &instance)
				Expect(output).To(HaveKeyWithValue("crownlabs.polito.it/has-node-selector", c.ExpectedValue))
			},
			Entry("When the node selector of the environment and the instance are present and different", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: &map[string]string{"key1": "val1"}}},
				InstanceNodeSelector: map[string]string{"key2": "val2"},
				ExpectedValue:        "true",
			}),
			Entry("When the node selector of the environment and the instance are present and equal", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: &map[string]string{"key": "val"}}},
				InstanceNodeSelector: map[string]string{"key": "val"},
				ExpectedValue:        "true",
			}),
			Entry("When the node selector of the environment is empty and the node selector of the instance is present", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: &map[string]string{}}},
				InstanceNodeSelector: map[string]string{"key": "val"},
				ExpectedValue:        "true",
			}),
			Entry("When the node selector of the environment is present and the node selector of the instance is empty", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: &map[string]string{"key": "val"}}},
				InstanceNodeSelector: map[string]string{},
				ExpectedValue:        "true",
			}),
			Entry("When the node selector of the environment is present and the node selector of the instance is nil", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: &map[string]string{"key": "val"}}},
				InstanceNodeSelector: nil,
				ExpectedValue:        "true",
			}),
			Entry("When the node selector of the environment is nil and the node selector of the instance is present", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: nil}},
				InstanceNodeSelector: map[string]string{"key": "val"},
				ExpectedValue:        "false",
			}),
			Entry("When the node selector of the environment is nil and the node selector of the instance is empty", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: &map[string]string{"key": "val"}}},
				InstanceNodeSelector: map[string]string{},
				ExpectedValue:        "true",
			}),
			Entry("When the node selector of the environment is empty and the node selector of the instance is nil", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: &map[string]string{}}},
				InstanceNodeSelector: nil,
				ExpectedValue:        "false",
			}),
			Entry("When the node selector of the environment and the node selector of the instance are empty", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: &map[string]string{}}},
				InstanceNodeSelector: map[string]string{},
				ExpectedValue:        "false",
			}),
			Entry("When the node selector of the environment and the node selector of the instance are nil", NodeSelectorEnabledLabelCase{
				EnvironmentList:      []clv1alpha2.Environment{{NodeSelector: nil}},
				InstanceNodeSelector: nil,
				ExpectedValue:        "false",
			}),
		)

		DescribeTable("Correctly configures the automation labels",
			func(c InstanceAutomationLabelCase) {
				output, _ := forge.InstanceLabels(c.Input, &template, &clv1alpha2.Instance{
					Spec: clv1alpha2.InstanceSpec{
						CustomizationUrls: c.InstanceCustomizationUrls,
					},
				})
				if c.ExpectedValue != "" {
					Expect(output).To(HaveKeyWithValue(forge.InstanceTerminationSelectorLabel, c.ExpectedValue))
				} else {
					Expect(output).NotTo(HaveKey(forge.InstanceTerminationSelectorLabel))
				}
			},
			Entry("When the Instance customizationUrls is nil", InstanceAutomationLabelCase{
				Input:                     map[string]string{},
				InstanceCustomizationUrls: nil,
				ExpectedValue:             "",
			}),
			Entry("When the Instance customizationUrls statusCheck is not set", InstanceAutomationLabelCase{
				Input:                     map[string]string{},
				InstanceCustomizationUrls: &clv1alpha2.InstanceCustomizationUrls{},
				ExpectedValue:             "",
			}),
			Entry("When the Instance customizationUrls statusCheck is set", InstanceAutomationLabelCase{
				Input:                     map[string]string{},
				InstanceCustomizationUrls: &clv1alpha2.InstanceCustomizationUrls{StatusCheck: statusCheckURL},
				ExpectedValue:             "true",
			}),
			Entry("When the Instance termination label was already set", InstanceAutomationLabelCase{
				Input: map[string]string{
					forge.InstanceTerminationSelectorLabel: "false",
				},
				InstanceCustomizationUrls: &clv1alpha2.InstanceCustomizationUrls{StatusCheck: statusCheckURL},
				ExpectedValue:             "false",
			}),
		)

		Context("Checking side effects", func() {
			var input, expectedInput map[string]string

			BeforeEach(func() {
				input = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
				expectedInput = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
			})

			JustBeforeEach(func() { forge.InstanceLabels(input, &template, nil) })
			It("The original labels map is not modified", func() { Expect(input).To(Equal(expectedInput)) })
		})
	})

	Describe("The forge.InstanceObjectLabels function", func() {
		type ObjectLabelsCase struct {
			Input          map[string]string
			ExpectedOutput map[string]string
		}

		var instance clv1alpha2.Instance

		BeforeEach(func() {
			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
				Spec: clv1alpha2.InstanceSpec{
					Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
					Tenant:   clv1alpha2.GenericRef{Name: tenantName},
				},
			}
		})

		DescribeTable("Correctly populates the labels set",
			func(c ObjectLabelsCase) {
				Expect(forge.InstanceObjectLabels(c.Input, &instance)).To(Equal(c.ExpectedOutput))
			},
			Entry("When the input labels map is nil", ObjectLabelsCase{
				Input: nil,
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/instance":   instanceName,
					"crownlabs.polito.it/template":   templateName,
					"crownlabs.polito.it/tenant":     tenantName,
				},
			}),
			Entry("When the input labels map already contains the expected values", ObjectLabelsCase{
				Input: map[string]string{
					"crownlabs.polito.it/managed-by":        "instance",
					"crownlabs.polito.it/instance":          instanceName,
					"crownlabs.polito.it/template":          templateName,
					"crownlabs.polito.it/tenant":            tenantName,
					"user/key":                              "user/value",
					"crownlabs.polito.it/has-node-selector": "false",
				},
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by":        "instance",
					"crownlabs.polito.it/instance":          instanceName,
					"crownlabs.polito.it/template":          templateName,
					"crownlabs.polito.it/tenant":            tenantName,
					"user/key":                              "user/value",
					"crownlabs.polito.it/has-node-selector": "false",
				},
			}),
			Entry("When the input labels map contains only part of the expected values", ObjectLabelsCase{
				Input: map[string]string{
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/template":   templateName,
					"user/key":                       "user/value",
				},
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/instance":   instanceName,
					"crownlabs.polito.it/template":   templateName,
					"crownlabs.polito.it/tenant":     tenantName,
					"user/key":                       "user/value",
				},
			}),
		)

		Context("Checking side effects", func() {
			var input, expectedInput map[string]string

			BeforeEach(func() {
				input = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
				expectedInput = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
			})

			JustBeforeEach(func() { forge.InstanceObjectLabels(input, &instance) })
			It("The original labels map is not modified", func() { Expect(input).To(Equal(expectedInput)) })
		})
	})

	Describe("The forge.SandboxObjectLabels function", func() {

		type ObjectLabelsCase struct {
			Input          map[string]string
			ExpectedOutput map[string]string
		}

		DescribeTable("Correctly populates the labels set",
			func(c ObjectLabelsCase) {
				Expect(forge.SandboxObjectLabels(c.Input, tenantName)).To(Equal(c.ExpectedOutput))
			},
			Entry("When the input labels map is nil", ObjectLabelsCase{
				Input: nil,
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by": "tenant",
					"crownlabs.polito.it/type":       sandboxName,
					"crownlabs.polito.it/tenant":     tenantName,
				},
			}),
			Entry("When the input labels map already contains the expected values", ObjectLabelsCase{
				Input: map[string]string{
					"crownlabs.polito.it/managed-by": "tenant",
					"crownlabs.polito.it/type":       sandboxName,
					"crownlabs.polito.it/tenant":     tenantName,
					"user/key":                       "user/value",
				},
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by": "tenant",
					"crownlabs.polito.it/type":       sandboxName,
					"crownlabs.polito.it/tenant":     tenantName,
					"user/key":                       "user/value",
				},
			}),
			Entry("When the input labels map contains only part of the expected values", ObjectLabelsCase{
				Input: map[string]string{
					"crownlabs.polito.it/managed-by": "tenant",
					"crownlabs.polito.it/type":       sandboxName,
					"user/key":                       "user/value",
				},
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by": "tenant",
					"crownlabs.polito.it/type":       sandboxName,
					"crownlabs.polito.it/tenant":     tenantName,
					"user/key":                       "user/value",
				},
			}),
		)

		Context("Checking side effects", func() {
			var input, expectedInput map[string]string

			BeforeEach(func() {
				input = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
				expectedInput = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
			})

			JustBeforeEach(func() { forge.SandboxObjectLabels(input, sandboxName) })
			It("The original labels map is not modified", func() { Expect(input).To(Equal(expectedInput)) })
		})
	})

	Describe("The forge.InstanceSelectorLabels function", func() {
		var instance clv1alpha2.Instance

		BeforeEach(func() {
			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
				Spec: clv1alpha2.InstanceSpec{
					Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
					Tenant:   clv1alpha2.GenericRef{Name: tenantName},
				},
			}
		})

		Context("The selector labels are generated", func() {
			It("Should have the correct values", func() {
				Expect(forge.InstanceSelectorLabels(&instance)).To(Equal(map[string]string{
					"crownlabs.polito.it/instance": instanceName,
					"crownlabs.polito.it/template": templateName,
					"crownlabs.polito.it/tenant":   tenantName,
				}))
			})

			It("Should be a subset of the object labels", func() {
				selectorLabels := forge.InstanceSelectorLabels(&instance)
				objectLabels := forge.InstanceObjectLabels(nil, &instance)
				for key, value := range selectorLabels {
					Expect(objectLabels).To(HaveKeyWithValue(key, value))
				}
			})
		})
	})

	Describe("The forge.InstanceAutomationLabelsOnTermination function", func() {
		type AutomationLabelsOnTerminationCase struct {
			Input                 map[string]string
			InputSubmissionNeeded bool
			ExpectedOutput        map[string]string
		}
		DescribeTable("Correctly populates the labels set",
			func(c AutomationLabelsOnTerminationCase) {
				Expect(forge.InstanceAutomationLabelsOnTermination(c.Input, c.InputSubmissionNeeded)).To(Equal(c.ExpectedOutput))
			},
			Entry("When the input labels map is nil", AutomationLabelsOnTerminationCase{
				Input:                 nil,
				InputSubmissionNeeded: true,
				ExpectedOutput: map[string]string{
					forge.InstanceTerminationSelectorLabel: "false",
					forge.InstanceSubmissionSelectorLabel:  "true",
				},
			}),
			Entry("When the submission is not needed", AutomationLabelsOnTerminationCase{
				Input:                 nil,
				InputSubmissionNeeded: false,
				ExpectedOutput: map[string]string{
					forge.InstanceTerminationSelectorLabel: "false",
					forge.InstanceSubmissionSelectorLabel:  "false",
				},
			}),
			Entry("When the input labels map contains other values", AutomationLabelsOnTerminationCase{
				Input:                 map[string]string{"some-key": "some-value"},
				InputSubmissionNeeded: true,
				ExpectedOutput: map[string]string{
					"some-key":                             "some-value",
					forge.InstanceTerminationSelectorLabel: "false",
					forge.InstanceSubmissionSelectorLabel:  "true",
				},
			}),
			Entry("When the input labels map is already compliant", AutomationLabelsOnTerminationCase{
				Input: map[string]string{
					forge.InstanceTerminationSelectorLabel: "false",
					forge.InstanceSubmissionSelectorLabel:  "true",
				},
				InputSubmissionNeeded: true,
				ExpectedOutput: map[string]string{
					forge.InstanceTerminationSelectorLabel: "false",
					forge.InstanceSubmissionSelectorLabel:  "true",
				},
			}),
		)
		Context("Checking side effects", func() {
			var input, expectedInput map[string]string
			BeforeEach(func() {
				input = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
				expectedInput = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
			})
			JustBeforeEach(func() { forge.InstanceAutomationLabelsOnTermination(input, true) })
			It("The original labels map is not modified", func() { Expect(input).To(Equal(expectedInput)) })
		})
	})

	Describe("The forge.InstanceAutomationLabelsOnSubmission function", func() {
		type AutomationLabelsOnSubmissionCase struct {
			Input                  map[string]string
			InputSubmissionSuccess bool
			ExpectedOutput         map[string]string
		}

		DescribeTable("Correctly populates the labels set",
			func(c AutomationLabelsOnSubmissionCase) {
				Expect(forge.InstanceAutomationLabelsOnSubmission(c.Input, c.InputSubmissionSuccess)).To(Equal(c.ExpectedOutput))
			},
			Entry("When the input labels map contains other values", AutomationLabelsOnSubmissionCase{
				Input:                  map[string]string{"some-key": "some-value"},
				InputSubmissionSuccess: true,
				ExpectedOutput: map[string]string{
					"some-key":                             "some-value",
					forge.InstanceSubmissionSelectorLabel:  "false",
					forge.InstanceSubmissionCompletedLabel: "true",
				},
			}),
			Entry("When the input labels map is nil", AutomationLabelsOnSubmissionCase{
				Input:                  nil,
				InputSubmissionSuccess: true,
				ExpectedOutput: map[string]string{
					forge.InstanceSubmissionSelectorLabel:  "false",
					forge.InstanceSubmissionCompletedLabel: "true",
				},
			}),
			Entry("When the input labels map is already compliant", AutomationLabelsOnSubmissionCase{
				Input: map[string]string{
					forge.InstanceSubmissionSelectorLabel:  "false",
					forge.InstanceSubmissionCompletedLabel: "true",
				},
				InputSubmissionSuccess: true,
				ExpectedOutput: map[string]string{
					forge.InstanceSubmissionSelectorLabel:  "false",
					forge.InstanceSubmissionCompletedLabel: "true",
				},
			}),
			Entry("When the submission is not successful", AutomationLabelsOnSubmissionCase{
				Input:                  nil,
				InputSubmissionSuccess: false,
				ExpectedOutput: map[string]string{
					forge.InstanceSubmissionSelectorLabel:  "false",
					forge.InstanceSubmissionCompletedLabel: "false",
				},
			}),
		)

		Context("Checking side effects", func() {
			var input, expectedInput map[string]string

			BeforeEach(func() {
				input = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
				expectedInput = map[string]string{"crownlabs.polito.it/managed-by": "whatever"}
			})

			JustBeforeEach(func() { forge.InstanceAutomationLabelsOnSubmission(input, false) })
			It("The original labels map is not modified", func() { Expect(input).To(Equal(expectedInput)) })
		})
	})

	Describe("The forge.InstanceComponentLabels function", func() {
		var instance clv1alpha2.Instance

		BeforeEach(func() {
			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
				Spec: clv1alpha2.InstanceSpec{
					Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
					Tenant:   clv1alpha2.GenericRef{Name: tenantName},
				},
			}
		})

		Context("The component labels are generated", func() {
			It("Should be a superset of the object labels", func() {
				componentLabels := forge.InstanceComponentLabels(&instance, componentName)
				objectLabels := forge.InstanceObjectLabels(nil, &instance)
				for key, value := range objectLabels {
					Expect(componentLabels).To(HaveKeyWithValue(key, value))
				}
			})

			It("Should contain the component label", func() {
				Expect(forge.InstanceComponentLabels(&instance, componentName)).To(HaveKeyWithValue("crownlabs.polito.it/component", componentName))
			})
		})
	})

	Describe("The forge.MonitorableServiceLabels function", func() {
		const (
			externalLabel = "external-annotation"
			externalValue = "external-value"
		)
		expectedResult := map[string]string{
			"crownlabs.polito.it/metrics-enabled": "true",
		}
		It("Should set the correct values", func() {
			Expect(forge.MonitorableServiceLabels(nil)).To(Equal(expectedResult))
		})
		It("Should override incorrect values", func() {
			Expect(forge.MonitorableServiceLabels(map[string]string{
				"crownlabs.polito.it/metrics-enabled": "NOP",
			})).To(Equal(expectedResult))
		})
		It("Should not alter original values", func() {
			Expect(forge.MonitorableServiceLabels(map[string]string{
				externalLabel: externalValue,
			})).To(Equal(map[string]string{
				"crownlabs.polito.it/metrics-enabled": "true",
				externalLabel:                         externalValue,
			}))
		})

		Context("Checking side effects", func() {
			var input, expectedInput map[string]string
			BeforeEach(func() {
				input = map[string]string{externalLabel: externalValue}
				expectedInput = map[string]string{externalLabel: externalValue}
			})

			JustBeforeEach(func() { forge.MonitorableServiceLabels(input) })
			It("The original annotations map is not modified", func() { Expect(input).To(Equal(expectedInput)) })
		})
	})

	Describe("The forge.SharedVolumeObjectLabels function", func() {
		var (
			input    map[string]string
			output   map[string]string
			expected = map[string]string{
				"key":                             "value",
				"crownlabs.polito.it/managed-by":  "instance",
				"crownlabs.polito.it/volume-type": "sharedvolume",
			}
		)

		BeforeEach(func() {
			input = map[string]string{
				"key": "value",
			}
		})
		JustBeforeEach(func() {
			output = forge.SharedVolumeObjectLabels(input)
		})

		It("Should have the same labels", func() {
			Expect(output).To(Equal(expected))
		})
	})
})
