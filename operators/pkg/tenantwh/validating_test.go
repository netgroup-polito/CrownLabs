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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Validating webhook", func() {
	forgeTenantWithWorkspace := func(ws clv1alpha2.TenantWorkspaceEntry) *clv1alpha2.Tenant {
		return &clv1alpha2.Tenant{
			Spec: clv1alpha2.TenantSpec{
				Workspaces: []clv1alpha2.TenantWorkspaceEntry{ws},
			},
		}
	}

	forgeTenantWithWorkspaceUser := func(wsName string) *clv1alpha2.Tenant {
		return forgeTenantWithWorkspace(clv1alpha2.TenantWorkspaceEntry{
			Name: wsName,
			Role: clv1alpha2.User,
		})
	}

	var (
		validatingWH *TenantValidator
		request      admission.Request
		response     admission.Response
		manager      *clv1alpha2.Tenant

		workspaceWA     *clv1alpha1.Workspace
		workspaceWAName = "test-workspace-withApproval"
		workspaceNA     *clv1alpha1.Workspace
		workspaceNAName = "test-workspace-noAutoenroll"
		workspaceIM     *clv1alpha1.Workspace
		workspaceIMName = "test-workspace-immediate"
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

		workspaceWA = &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: workspaceWAName},
			Spec: clv1alpha1.WorkspaceSpec{
				PrettyName: "test-workspace",
				Quota: clv1alpha1.WorkspaceResourceQuota{
					Instances: 1,
				},
				AutoEnroll: clv1alpha1.AutoenrollWithApproval,
			},
		}

		workspaceNA = &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: workspaceNAName},
			Spec: clv1alpha1.WorkspaceSpec{
				PrettyName: "test-workspace",
				Quota: clv1alpha1.WorkspaceResourceQuota{
					Instances: 1,
				},
			},
		}

		workspaceIM = &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: workspaceIMName},
			Spec: clv1alpha1.WorkspaceSpec{
				PrettyName: "test-workspace",
				Quota: clv1alpha1.WorkspaceResourceQuota{
					Instances: 1,
				},
				AutoEnroll: clv1alpha1.AutoenrollImmediate,
			},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			manager,
			workspaceWA,
			workspaceNA,
			workspaceIM,
		).Build()

		validatingWH = MakeTenantValidator(fakeClient, bypassGroups, scheme).Handler.(*TenantValidator)

		Expect(validatingWH.decoder).NotTo(BeNil())
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

		When("lastLogin is changed within the LastLoginToleration", func() {
			BeforeEach(func() {
				newTenant = &clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{
					LastLogin: metav1.Time{
						Time: time.Now().Add(LastLoginToleration / 2),
					},
				}}
			})
			It("should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("lastLogin too far from now (abs(lastLogin-now)) > LastLoginToleration)", func() {
			BeforeEach(func() {
				newTenant = &clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{
					LastLogin: metav1.Time{
						Time: time.Now().Add(LastLoginToleration + time.Millisecond),
					},
				}}
			})
			It("should deny the change", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
				Expect(response.Result.Reason).NotTo(BeEmpty())
			})
		})

		Describe("a workspace is added", func() {
			WhenBody := func(nt *clv1alpha2.Tenant, shouldSucceed bool) func() {
				return func() {
					BeforeEach(func() {
						oldTenant = &clv1alpha2.Tenant{}
						newTenant = nt
					})

					if shouldSucceed {
						It("should allow the change", func() {
							Expect(response.Allowed).To(BeTrue())
						})
					} else {
						It("should deny the change", func() {
							Expect(response.Allowed).To(BeFalse())
							Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
							Expect(response.Result.Reason).ToNot(BeEmpty())
						})
					}
				}
			}

			When("no autoenroll and user role", WhenBody(
				forgeTenantWithWorkspaceUser(workspaceNAName),
				false,
			))

			When("no autoenroll and candidate role", WhenBody(
				forgeTenantWithWorkspace(clv1alpha2.TenantWorkspaceEntry{
					Name: workspaceNAName,
					Role: clv1alpha2.Candidate,
				}),
				false,
			))

			When("autoenroll withApproval and user role", WhenBody(
				forgeTenantWithWorkspaceUser(workspaceWAName),
				false,
			))

			When("autoenroll withApproval and candidate role", WhenBody(
				forgeTenantWithWorkspace(clv1alpha2.TenantWorkspaceEntry{
					Name: workspaceWAName,
					Role: clv1alpha2.Candidate,
				}),
				true,
			))

			When("autoenroll immediate and user role", WhenBody(
				forgeTenantWithWorkspaceUser(workspaceIMName),
				true,
			))

			When("autoenroll immediate and candidate role", WhenBody(
				forgeTenantWithWorkspace(clv1alpha2.TenantWorkspaceEntry{
					Name: workspaceIMName,
					Role: clv1alpha2.Candidate,
				}),
				false,
			))
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
		var operation admissionv1.Operation
		JustBeforeEach(func() {
			response = validatingWH.HandleWorkspaceEdit(ctx, newTenant, oldTenant, manager, operation)
		})

		When("manager adds a workspace he manages", func() {
			BeforeEach(func() {
				oldTenant = &clv1alpha2.Tenant{}
				newTenant = forgeTenantWithWorkspaceUser(testWorkspace)
				operation = admissionv1.Update
			})
			It("Should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("manager removes a workspace he manages", func() {
			BeforeEach(func() {
				newTenant = &clv1alpha2.Tenant{}
				oldTenant = forgeTenantWithWorkspaceUser(testWorkspace)
				operation = admissionv1.Update
			})
			It("Should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("manager adds a workspace he doesn't manage", func() {
			BeforeEach(func() {
				oldTenant = &clv1alpha2.Tenant{}
				newTenant = forgeTenantWithWorkspaceUser("invalid-workspace")
				operation = admissionv1.Update
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
				operation = admissionv1.Update
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
				operation = admissionv1.Update
			})
			It("should deny the change", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
				Expect(response.Result.Reason).NotTo(BeEmpty())
			})
		})

		When("a new user is created", func() {
			BeforeEach(func() {
				oldTenant = &clv1alpha2.Tenant{}
				newTenant = &clv1alpha2.Tenant{Spec: clv1alpha2.TenantSpec{Email: "other"}}
				operation = admissionv1.Create
			})
			It("should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
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
				result := CalculateWorkspacesDiff(
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
