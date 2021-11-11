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

package tenantwh_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/tenantwh"
)

var _ = Describe("Validating webhook", func() {
	forgeTenantWithWorkspaceUser := func(wsName string) *clv1alpha2.Tenant {
		return &clv1alpha2.Tenant{
			Spec: clv1alpha2.TenantSpec{
				Workspaces: []clv1alpha2.TenantWorkspaceEntry{{
					Name: wsName,
					Role: clv1alpha2.User,
				}},
			},
		}
	}

	var (
		validatingWH *tenantwh.TenantValidator
		request      admission.Request
		response     admission.Response
		manager      *clv1alpha2.Tenant
	)

	BeforeEach(func() {

		manager = &clv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{Name: "some-manager"},
			Spec: clv1alpha2.TenantSpec{
				Workspaces: []clv1alpha2.TenantWorkspaceEntry{{
					Name: testWorkspace,
					Role: clv1alpha2.Manager,
				}},
			},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(manager).Build()

		validatingWH = tenantwh.MakeTenantValidator(fakeClient, bypassGroups).Handler.(*tenantwh.TenantValidator)
		Expect(validatingWH.InjectDecoder(decoder)).To(Succeed())
	})

	Describe("The TenantValidator.Handle method", func() {
		JustBeforeEach(func() {
			response = validatingWH.Handle(ctx, request)
		})

		When("the user is an admin/operator", func() {
			BeforeEach(func() {
				request = forgeRequest(admissionv1.Create, nil, nil)
				request.UserInfo.Groups = bypassGroups
			})
			It("Should admit the request", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("the new tenant is invalid", func() {
			BeforeEach(func() {
				request = forgeRequest(admissionv1.Create, nil, nil)
			})
			It("Should return an error response", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusBadRequest))
				Expect(response.Result.Message).ToNot(BeEmpty())
			})
		})

		When("the old tenant is invalid", func() {
			BeforeEach(func() {
				request = forgeRequest(admissionv1.Update, &clv1alpha2.Tenant{}, nil)
			})
			It("Should return an error response", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusBadRequest))
				Expect(response.Result.Message).ToNot(BeEmpty())
			})
		})

		When("Tenant is being edited by an invalid external user", func() {
			BeforeEach(func() {
				request = forgeRequest(admissionv1.Update, &clv1alpha2.Tenant{}, &clv1alpha2.Tenant{})
				request.UserInfo.Username = "invalid-manager"
			})
			It("Should return an error response", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusBadRequest))
				Expect(response.Result.Message).ToNot(BeEmpty())
			})
		})
	})

	Describe("The TenantValidator.HandleSelfEdit method", func() {
		var newTenant, oldTenant *clv1alpha2.Tenant
		JustBeforeEach(func() {
			response = validatingWH.HandleSelfEdit(ctx, newTenant, oldTenant)
		})

		When("only public keys are changed", func() {
			BeforeEach(func() {
				oldTenant = &clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{PublicKeys: []string{"some-key"}}}
				newTenant = &clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{PublicKeys: []string{"other-key"}}}
			})
			It("should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("other fields are changed", func() {
			BeforeEach(func() {
				oldTenant = &clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{LastName: "test"}}
				newTenant = &clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{Email: "other"}}
			})
			It("should deny the change", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
				Expect(response.Result.Reason).NotTo(BeEmpty())
			})
		})
	})

	Describe("The TenantValidator.HandleWorkspaceEdit", func() {
		var newTenant, oldTenant *clv1alpha2.Tenant
		JustBeforeEach(func() {
			response = validatingWH.HandleWorkspaceEdit(ctx, newTenant, oldTenant, manager)
		})

		When("manager adds a workspace he manages", func() {
			BeforeEach(func() {
				oldTenant = &clv1alpha2.Tenant{}
				newTenant = forgeTenantWithWorkspaceUser(testWorkspace)
			})
			It("Should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("manager removes a workspace he manages", func() {
			BeforeEach(func() {
				newTenant = &clv1alpha2.Tenant{}
				oldTenant = forgeTenantWithWorkspaceUser(testWorkspace)
			})
			It("Should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("manager adds a workspace he doesn't manage", func() {
			BeforeEach(func() {
				oldTenant = &clv1alpha2.Tenant{}
				newTenant = forgeTenantWithWorkspaceUser("invalid-workspace")
			})
			It("Should allow the change", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
				Expect(response.Result.Reason).NotTo(BeEmpty())
			})
		})

		When("manager remvoes a workspace he doesn't manage", func() {
			BeforeEach(func() {
				newTenant = &clv1alpha2.Tenant{}
				oldTenant = forgeTenantWithWorkspaceUser("invalid-workspace")
			})
			It("Should allow the change", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
				Expect(response.Result.Reason).NotTo(BeEmpty())
			})
		})

		When("other fields are changed", func() {
			BeforeEach(func() {
				oldTenant = &clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{LastName: "test"}}
				newTenant = &clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{Email: "other"}}
			})
			It("should deny the change", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
				Expect(response.Result.Reason).NotTo(BeEmpty())
			})
		})
	})

	Describe("The CalculateWorkspacesDiff function", func() {
		type CalcWsDiffCase struct {
			a, b            []clv1alpha2.TenantWorkspaceEntry
			expectedChanges []string
		}
		WhenBody := func(cwdc CalcWsDiffCase) func() {
			return func() {
				var actuals []string
				result := tenantwh.CalculateWorkspacesDiff(
					&clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{Workspaces: cwdc.a}},
					&clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{Workspaces: cwdc.b}},
				)
				for k, v := range result {
					if v {
						actuals = append(actuals, k)
					}
				}
				It("Should calculate the correct diffs", func() {
					Expect(actuals).To(BeEquivalentTo(cwdc.expectedChanges))
				})
			}
		}
		makeWs := func(name string, role clv1alpha2.WorkspaceUserRole) clv1alpha2.TenantWorkspaceEntry {
			return clv1alpha2.TenantWorkspaceEntry{Name: name, Role: role}
		}

		When("The two tenants have no different workspaces", WhenBody(CalcWsDiffCase{
			a:               []clv1alpha2.TenantWorkspaceEntry{makeWs("test-a", clv1alpha2.Manager)},
			b:               []clv1alpha2.TenantWorkspaceEntry{makeWs("test-a", clv1alpha2.Manager)},
			expectedChanges: nil,
		}))

		When("The first tenant has one more workspace", WhenBody(CalcWsDiffCase{
			a:               []clv1alpha2.TenantWorkspaceEntry{makeWs("test-a", clv1alpha2.Manager)},
			b:               []clv1alpha2.TenantWorkspaceEntry{},
			expectedChanges: []string{"test-a"},
		}))

		When("The second tenant has one more workspace", WhenBody(CalcWsDiffCase{
			a:               []clv1alpha2.TenantWorkspaceEntry{makeWs("test-a", clv1alpha2.Manager)},
			b:               []clv1alpha2.TenantWorkspaceEntry{makeWs("test-a", clv1alpha2.Manager), makeWs("test-b", clv1alpha2.User)},
			expectedChanges: []string{"test-b"},
		}))

		When("There is a difference in roles", WhenBody(CalcWsDiffCase{
			a:               []clv1alpha2.TenantWorkspaceEntry{makeWs("test-a", clv1alpha2.User)},
			b:               []clv1alpha2.TenantWorkspaceEntry{makeWs("test-a", clv1alpha2.Manager)},
			expectedChanges: []string{"test-a"},
		}))

	})
})
