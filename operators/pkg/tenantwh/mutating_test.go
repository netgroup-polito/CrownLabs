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

package tenantwh

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Mutating webhook", func() {
	var (
		mutatingWH *TenantMutator
		request    admission.Request

		opSelectorKey   = "crownlabs.polito.it/op-sel"
		opSelectorValue = "prod"
		baseWorkspaces  = []string{}
	)

	forgeOpSelectorMap := func(opSel string) map[string]string {
		return map[string]string{opSelectorKey: opSel}
	}

	forgeTenantWithLabels := func(name string, labels map[string]string) *clv1alpha2.Tenant {
		return &clv1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}}
	}

	forgeTenantWithWorkspaces := func(name string, workspaces []clv1alpha2.TenantWorkspaceEntry) *clv1alpha2.Tenant {
		return &clv1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: clv1alpha2.TenantSpec{Workspaces: workspaces}}
	}

	JustBeforeEach(func() {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mutatingWH = MakeTenantMutator(fakeClient, bypassGroups, opSelectorKey, opSelectorValue, baseWorkspaces, scheme).Handler.(*TenantMutator)
		Expect(mutatingWH.decoder).NotTo(BeNil())
	})

	Describe("The TenantMutator.Handle method", func() {
		var response, expectedResponse admission.Response

		JustBeforeEach(func() {
			response = mutatingWH.Handle(ctx, request)
		})

		When("the request is invalid", func() {
			BeforeEach(func() {
				request = admission.Request{}
			})

			It("Should return an error response", func() {
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusBadRequest))
				Expect(response.Result.Message).NotTo(BeEmpty())
				Expect(response.Allowed).To(BeFalse())
			})
		})

		When("the request is valid", func() {
			BeforeEach(func() {
				testTenant := clv1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: "test-tenant"}}
				request = forgeRequest(admissionv1.Create, &testTenant, nil)
				labels, _, _ := mutatingWH.EnforceTenantLabels(ctx, &request, nil)
				testTenant.SetLabels(labels)
				expectedResponse = mutatingWH.CreatePatchResponse(ctx, &request, &testTenant)
			})

			It("Should return a valid response", func() {
				Expect(response).To(Equal(expectedResponse))
			})
		})
	})

	Describe("The TenantMutator.EnforceTenantLabels method", func() {
		type EnforceLabelsCase struct {
			newTenant, oldTenant *clv1alpha2.Tenant
			operation            admissionv1.Operation
			expectedLabels       map[string]string
			expectedWarnings     []string
			expectedError        error
			beforeEach           func(*EnforceLabelsCase)
		}

		var (
			actualLabels   map[string]string
			actualWarnings []string
			actualError    error
		)

		WhenBody := func(elc EnforceLabelsCase) {
			BeforeEach(func() {
				request = forgeRequest(elc.operation, elc.newTenant, elc.oldTenant)
				if elc.beforeEach != nil {
					elc.beforeEach(&elc)
				}
			})
			JustBeforeEach(func() {
				actualLabels, actualWarnings, actualError = mutatingWH.EnforceTenantLabels(ctx, &request, elc.newTenant.Labels)
			})
			It("Should return the expected resutls", func() {
				if elc.expectedError != nil {
					Expect(actualError).To(MatchError(elc.expectedError))
				} else {
					Expect(actualError).To(BeNil())
				}
				Expect(actualWarnings).To(Equal(elc.expectedWarnings))
				Expect(actualLabels).To(Equal(elc.expectedLabels))
			})
		}

		Context("Operation is create", func() {
			When("operation is issued against the service tenant", func() {
				WhenBody(EnforceLabelsCase{
					operation:        admissionv1.Create,
					newTenant:        forgeTenantWithLabels(clv1alpha2.SVCTenantName, map[string]string{opSelectorKey: "something-not-nil"}),
					expectedLabels:   map[string]string{opSelectorKey: ""},
					expectedWarnings: []string{"operator selector label must not be present on service tenant and has been removed"},
				})
			})

			When("operation is issued by an admin/operator", func() {
				testLabels := map[string]string{"test1": "test", opSelectorKey: opSelectorValue}
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Create,
					newTenant:      forgeTenantWithLabels(testTenantName, testLabels),
					expectedLabels: testLabels,
					beforeEach:     func(_ *EnforceLabelsCase) { request.UserInfo.Groups = bypassGroups },
				})
			})

			When("no operator selector label is set", func() {
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Create,
					newTenant:      forgeTenantWithLabels(testTenantName, nil),
					expectedLabels: forgeOpSelectorMap(opSelectorValue),
				})
			})

			When("operator selector label is present and invalid", func() {
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Create,
					newTenant:      forgeTenantWithLabels(testTenantName, map[string]string{opSelectorKey: "something-not-nil"}),
					expectedLabels: forgeOpSelectorMap(opSelectorValue),
				})
			})

			When("operator selector label is present and valid", func() {
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Create,
					newTenant:      forgeTenantWithLabels(testTenantName, map[string]string{opSelectorKey: opSelectorValue}),
					expectedLabels: forgeOpSelectorMap(opSelectorValue),
				})
			})
		})

		Context("Operation is update", func() {
			When("old tenant is invalid", func() {
				var expectedErr error
				WhenBody(EnforceLabelsCase{
					operation:     admissionv1.Update,
					newTenant:     &clv1alpha2.Tenant{},
					oldTenant:     nil,
					expectedError: expectedErr,
					beforeEach:    func(elc *EnforceLabelsCase) { _, elc.expectedError = mutatingWH.DecodeTenant(runtime.RawExtension{}) },
				})
			})

			When("operator selector label change is attempted", func() {
				WhenBody(EnforceLabelsCase{
					operation:        admissionv1.Update,
					newTenant:        forgeTenantWithLabels(testTenantName, forgeOpSelectorMap("invalid")),
					oldTenant:        forgeTenantWithLabels(testTenantName, forgeOpSelectorMap("some")),
					expectedLabels:   forgeOpSelectorMap("some"),
					expectedWarnings: []string{"operator selector label change is prohibited and has been reverted"},
				})
			})

			When("operator selector label was not present and is added", func() {
				WhenBody(EnforceLabelsCase{
					operation:        admissionv1.Update,
					newTenant:        forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(opSelectorValue)),
					oldTenant:        forgeTenantWithLabels(testTenantName, nil),
					expectedLabels:   map[string]string{},
					expectedWarnings: []string{"operator selector label change is prohibited and has been reverted"},
				})
			})

			When("operator selector label is already present, custom and new val is correct", func() {
				customVal := "custom" + opSelectorValue
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Update,
					newTenant:      forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(customVal)),
					oldTenant:      forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(customVal)),
					expectedLabels: forgeOpSelectorMap(customVal),
				})
			})

			When("operator selector label is already present, custom and new val differs", func() {
				customVal := "custom" + opSelectorValue
				WhenBody(EnforceLabelsCase{
					operation:        admissionv1.Update,
					newTenant:        forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(opSelectorValue)),
					oldTenant:        forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(customVal)),
					expectedLabels:   forgeOpSelectorMap(customVal),
					expectedWarnings: []string{"operator selector label change is prohibited and has been reverted"},
				})
			})
		})
	})

	Describe("The TenantMutator.EnforceTenantBaseWorkspaces method", func() {
		type EnforceTenantBaseWorkspacesCase struct {
			testTenant         *clv1alpha2.Tenant
			testBaseWorkspaces []string
			expectedWorkspaces []clv1alpha2.TenantWorkspaceEntry
		}

		exampleWs1 := clv1alpha2.TenantWorkspaceEntry{Name: "workspace", Role: clv1alpha2.Manager}
		testWsName := "utilities"

		WhenBody := func(elc EnforceTenantBaseWorkspacesCase) {
			BeforeEach(func() {
				baseWorkspaces = elc.testBaseWorkspaces
			})
			JustBeforeEach(func() {
				mutatingWH.EnforceTenantBaseWorkspaces(ctx, elc.testTenant)
			})
			It("Should set the expected base workspaces", func() {
				Expect(elc.testTenant.Spec.Workspaces).To(Equal(elc.expectedWorkspaces))
			})
		}

		When("the tenant has no workspaces", func() {
			WhenBody(EnforceTenantBaseWorkspacesCase{
				testTenant:         forgeTenantWithWorkspaces(clv1alpha2.SVCTenantName, nil),
				testBaseWorkspaces: []string{testWsName},
				expectedWorkspaces: []clv1alpha2.TenantWorkspaceEntry{{
					Name: testWsName,
					Role: clv1alpha2.User,
				}},
			})
		})

		When("the tenant has some workspaces", func() {
			WhenBody(EnforceTenantBaseWorkspacesCase{
				testTenant:         forgeTenantWithWorkspaces(clv1alpha2.SVCTenantName, []clv1alpha2.TenantWorkspaceEntry{exampleWs1}),
				testBaseWorkspaces: []string{testWsName},
				expectedWorkspaces: []clv1alpha2.TenantWorkspaceEntry{
					exampleWs1, {
						Name: testWsName,
						Role: clv1alpha2.User,
					}},
			})
		})

		When("the tenant already has the base workspaces already set", func() {
			WhenBody(EnforceTenantBaseWorkspacesCase{
				testTenant: forgeTenantWithWorkspaces(clv1alpha2.SVCTenantName, []clv1alpha2.TenantWorkspaceEntry{exampleWs1, {
					Name: testWsName,
					Role: clv1alpha2.Manager,
				}}),
				testBaseWorkspaces: []string{testWsName},
				expectedWorkspaces: []clv1alpha2.TenantWorkspaceEntry{
					exampleWs1, {
						Name: testWsName,
						Role: clv1alpha2.Manager,
					}},
			})
		})
	})
})
