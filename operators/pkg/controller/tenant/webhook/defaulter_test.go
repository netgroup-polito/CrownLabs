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

package webhook_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant/webhook"
)

// DummyObject implements runtime.Object to test error case.
type DummyObject struct{}

func (d *DummyObject) GetObjectKind() schema.ObjectKind { return nil }
func (d *DummyObject) DeepCopyObject() runtime.Object   { return &DummyObject{} }

var _ = Describe("Defaulter webhook", func() {
	var (
		mutatingWH *webhook.TenantDefaulter
		request    admission.Request

		baseWorkspaces = []string{}
	)

	opSelector := common.NewLabel("crownlabs.polito.it/op-sel", "test")

	forgeOpSelectorMap := func(opSel string) map[string]string {
		return map[string]string{opSelector.GetKey(): opSel}
	}

	forgeTenantWithLabels := func(name string, labels map[string]string) *v1alpha2.Tenant {
		return &v1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}}
	}

	forgeTenantWithWorkspaces := func(name string, workspaces []v1alpha2.TenantWorkspaceEntry) *v1alpha2.Tenant {
		return &v1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: v1alpha2.TenantSpec{Workspaces: workspaces}}
	}

	JustBeforeEach(func() {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		mutatingWH = &webhook.TenantDefaulter{
			OpSelectorLabel: opSelector,
			BaseWorkspaces:  baseWorkspaces,
			Decoder:         admission.NewDecoder(scheme),
			TenantWebhook: webhook.TenantWebhook{
				Client:       fakeClient,
				BypassGroups: []string{"system:serviceaccounts:default"},
			},
		}
		Expect(mutatingWH.Decoder).NotTo(BeNil())
	})

	Describe("The TenantMutator.Default method", func() {
		var (
			tenant *v1alpha2.Tenant
			obj    runtime.Object
			ctx    context.Context
		)

		BeforeEach(func() {
			ctx = context.TODO()
		})

		Context("when creating a normal tenant", func() {
			BeforeEach(func() {
				tenant = forgeTenantWithLabels("test-tenant", nil)
				obj = tenant
				request = forgeRequest(admissionv1.Create, tenant, nil)
				ctx = admission.NewContextWithRequest(ctx, request)
			})
			It("should set the operator selector label and base workspaces", func() {
				err := mutatingWH.Default(ctx, obj)
				Expect(err).To(BeNil())
				labels := tenant.GetLabels()
				Expect(labels).To(HaveKeyWithValue(opSelector.GetKey(), opSelector.GetValue()))
			})
		})

		Context("when creating the service tenant", func() {
			BeforeEach(func() {
				tenant = forgeTenantWithLabels(v1alpha2.SVCTenantName, map[string]string{opSelector.GetKey(): "should-be-removed"})
				obj = tenant
				request = forgeRequest(admissionv1.Create, tenant, nil)
				ctx = admission.NewContextWithRequest(ctx, request)
			})
			It("should not set the operator selector label", func() {
				err := mutatingWH.Default(ctx, obj)
				Expect(err).To(BeNil())
				labels := tenant.GetLabels()
				Expect(labels[opSelector.GetKey()]).To(Equal(""))
			})
		})

		Context("when updating a tenant and trying to change the label", func() {
			BeforeEach(func() {
				oldTenant := forgeTenantWithLabels("test-tenant", map[string]string{opSelector.GetKey(): "old-value"})
				tenant = forgeTenantWithLabels("test-tenant", map[string]string{opSelector.GetKey(): "new-value"})
				obj = tenant
				request = forgeRequest(admissionv1.Update, tenant, oldTenant)
				ctx = admission.NewContextWithRequest(ctx, request)
			})
			It("should revert the label to the old value", func() {
				err := mutatingWH.Default(ctx, obj)
				Expect(err).To(BeNil())
				labels := tenant.GetLabels()
				Expect(labels[opSelector.GetKey()]).To(Equal("old-value"))
			})
		})

		Context("when passing a non-Tenant object", func() {
			It("should return an error", func() {
				err := mutatingWH.Default(ctx, &DummyObject{})
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("The TenantMutator.EnforceTenantLabels method", func() {
		type EnforceLabelsCase struct {
			newTenant, oldTenant *v1alpha2.Tenant
			operation            admissionv1.Operation
			expectedLabels       map[string]string
			expectedError        error
			beforeEach           func(*EnforceLabelsCase)
		}

		var (
			actualLabels map[string]string
			actualError  error
		)

		WhenBody := func(elc EnforceLabelsCase) {
			BeforeEach(func() {
				request = forgeRequest(elc.operation, elc.newTenant, elc.oldTenant)
				if elc.beforeEach != nil {
					elc.beforeEach(&elc)
				}
			})
			JustBeforeEach(func() {
				actualLabels, actualError = mutatingWH.EnforceTenantLabels(ctx, &request, elc.newTenant.Labels)
			})
			It("Should return the expected resutls", func() {
				if elc.expectedError != nil {
					Expect(actualError).To(MatchError(elc.expectedError))
				} else {
					Expect(actualError).To(BeNil())
				}
				Expect(actualLabels).To(Equal(elc.expectedLabels))
			})
		}

		Context("Operation is create", func() {
			When("operation is issued against the service tenant", func() {
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Create,
					newTenant:      forgeTenantWithLabels(v1alpha2.SVCTenantName, map[string]string{opSelector.GetKey(): "something-not-nil"}),
					expectedLabels: map[string]string{opSelector.GetKey(): ""},
				})
			})

			When("operation is issued by an admin/operator", func() {
				testLabels := map[string]string{"test1": "test", opSelector.GetKey(): opSelector.GetValue()}
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
					expectedLabels: forgeOpSelectorMap(opSelector.GetValue()),
				})
			})

			When("operator selector label is present and invalid", func() {
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Create,
					newTenant:      forgeTenantWithLabels(testTenantName, map[string]string{opSelector.GetKey(): "something-not-nil"}),
					expectedLabels: forgeOpSelectorMap(opSelector.GetValue()),
				})
			})

			When("operator selector label is present and valid", func() {
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Create,
					newTenant:      forgeTenantWithLabels(testTenantName, map[string]string{opSelector.GetKey(): opSelector.GetValue()}),
					expectedLabels: forgeOpSelectorMap(opSelector.GetValue()),
				})
			})
		})

		Context("Operation is update", func() {
			When("old tenant is invalid", func() {
				var expectedErr error
				WhenBody(EnforceLabelsCase{
					operation:     admissionv1.Update,
					newTenant:     &v1alpha2.Tenant{},
					oldTenant:     nil,
					expectedError: expectedErr,
					beforeEach:    func(elc *EnforceLabelsCase) { _, elc.expectedError = mutatingWH.DecodeTenant(runtime.RawExtension{}) },
				})
			})

			When("operator selector label change is attempted", func() {
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Update,
					newTenant:      forgeTenantWithLabels(testTenantName, forgeOpSelectorMap("invalid")),
					oldTenant:      forgeTenantWithLabels(testTenantName, forgeOpSelectorMap("some")),
					expectedLabels: forgeOpSelectorMap("some"),
				})
			})

			When("operator selector label was not present and is added", func() {
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Update,
					newTenant:      forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(opSelector.GetValue())),
					oldTenant:      forgeTenantWithLabels(testTenantName, nil),
					expectedLabels: map[string]string{},
				})
			})

			When("operator selector label is already present, custom and new val is correct", func() {
				customVal := "custom" + opSelector.GetValue()
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Update,
					newTenant:      forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(customVal)),
					oldTenant:      forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(customVal)),
					expectedLabels: forgeOpSelectorMap(customVal),
				})
			})

			When("operator selector label is already present, custom and new val differs", func() {
				customVal := "custom" + opSelector.GetValue()
				WhenBody(EnforceLabelsCase{
					operation:      admissionv1.Update,
					newTenant:      forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(opSelector.GetValue())),
					oldTenant:      forgeTenantWithLabels(testTenantName, forgeOpSelectorMap(customVal)),
					expectedLabels: forgeOpSelectorMap(customVal),
				})
			})
		})
	})

	Describe("The TenantMutator.EnforceTenantBaseWorkspaces method", func() {
		type EnforceTenantBaseWorkspacesCase struct {
			testTenant         *v1alpha2.Tenant
			testBaseWorkspaces []string
			expectedWorkspaces []v1alpha2.TenantWorkspaceEntry
		}

		exampleWs1 := v1alpha2.TenantWorkspaceEntry{Name: "workspace", Role: v1alpha2.Manager}
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
				testTenant:         forgeTenantWithWorkspaces(v1alpha2.SVCTenantName, nil),
				testBaseWorkspaces: []string{testWsName},
				expectedWorkspaces: []v1alpha2.TenantWorkspaceEntry{{
					Name: testWsName,
					Role: v1alpha2.User,
				}},
			})
		})

		When("the tenant has some workspaces", func() {
			WhenBody(EnforceTenantBaseWorkspacesCase{
				testTenant:         forgeTenantWithWorkspaces(v1alpha2.SVCTenantName, []v1alpha2.TenantWorkspaceEntry{exampleWs1}),
				testBaseWorkspaces: []string{testWsName},
				expectedWorkspaces: []v1alpha2.TenantWorkspaceEntry{
					exampleWs1, {
						Name: testWsName,
						Role: v1alpha2.User,
					}},
			})
		})

		When("the tenant already has the base workspaces already set", func() {
			WhenBody(EnforceTenantBaseWorkspacesCase{
				testTenant: forgeTenantWithWorkspaces(v1alpha2.SVCTenantName, []v1alpha2.TenantWorkspaceEntry{exampleWs1, {
					Name: testWsName,
					Role: v1alpha2.Manager,
				}}),
				testBaseWorkspaces: []string{testWsName},
				expectedWorkspaces: []v1alpha2.TenantWorkspaceEntry{
					exampleWs1, {
						Name: testWsName,
						Role: v1alpha2.Manager,
					}},
			})
		})
	})
})
