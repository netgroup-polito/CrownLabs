// Copyright 2020-2021 Politecnico di Torino
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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
	)

	Describe("The forge.InstanceLabels function", func() {
		var template clv1alpha2.Template

		type InstanceLabelsCase struct {
			Input           map[string]string
			ExpectedOutput  map[string]string
			ExpectedUpdated bool
		}

		type InstancePersistentLabelCase struct {
			EnvironmentList []clv1alpha2.Environment
			ExpectedValue   string
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
					"crownlabs.polito.it/persistent": "false",
				},
				ExpectedUpdated: true,
			}),
			Entry("When the input labels map already contains the expected values", InstanceLabelsCase{
				Input: map[string]string{
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/workspace":  workspaceName,
					"crownlabs.polito.it/template":   templateName,
					"crownlabs.polito.it/persistent": "false",
					"user/key":                       "user/value",
				},
				ExpectedOutput: map[string]string{
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/workspace":  workspaceName,
					"crownlabs.polito.it/template":   templateName,
					"crownlabs.polito.it/persistent": "false",
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
					"crownlabs.polito.it/persistent": "false",
					"user/key":                       "user/value",
				},
				ExpectedUpdated: true,
			}),
		)

		DescribeTable("Correctly configures the persistent label",
			func(c InstancePersistentLabelCase) {
				template.Spec.EnvironmentList = c.EnvironmentList
				output, _ := forge.InstanceLabels(map[string]string{}, &template)
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
	})

	Describe("The forge.InstanceObjectLabels function", func() {
		var instance clv1alpha2.Instance

		type ObjectLabelsCase struct {
			Input          map[string]string
			ExpectedOutput map[string]string
		}

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
					"crownlabs.polito.it/managed-by": "instance",
					"crownlabs.polito.it/instance":   instanceName,
					"crownlabs.polito.it/template":   templateName,
					"crownlabs.polito.it/tenant":     tenantName,
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
})
