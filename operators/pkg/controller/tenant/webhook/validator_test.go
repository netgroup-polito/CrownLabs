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
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant/webhook"
)

var _ = Describe("Validator webhook", func() {
	forgeTenantWithWorkspace := func(ws v1alpha2.TenantWorkspaceEntry) *v1alpha2.Tenant {
		return &v1alpha2.Tenant{
			Spec: v1alpha2.TenantSpec{
				Workspaces: []v1alpha2.TenantWorkspaceEntry{ws},
			},
		}
	}

	forgeTenantWithWorkspaceUser := func(wsName string) *v1alpha2.Tenant {
		return forgeTenantWithWorkspace(v1alpha2.TenantWorkspaceEntry{
			Name: wsName,
			Role: v1alpha2.User,
		})
	}

	var (
		tnValidator *webhook.TenantValidator
		tnWebhook   *admission.Webhook
		request     admission.Request
		response    admission.Response
		manager     *v1alpha2.Tenant
		fakeManager *v1alpha2.Tenant

		workspaceWA     *v1alpha1.Workspace
		workspaceWAName = "test-workspace-withApproval"
		workspaceNA     *v1alpha1.Workspace
		workspaceNAName = "test-workspace-noAutoenroll"
		workspaceIM     *v1alpha1.Workspace
		workspaceIMName = "test-workspace-immediate"
	)

	BeforeEach(func() {

		manager = &v1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{Name: "some-manager"},
			Spec: v1alpha2.TenantSpec{
				Workspaces: []v1alpha2.TenantWorkspaceEntry{{
					Name: testWorkspace,
					Role: v1alpha2.Manager,
				}},
			},
		}

		fakeManager = &v1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{Name: "some-fake-manager"},
		}

		workspaceWA = &v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: workspaceWAName},
			Spec: v1alpha1.WorkspaceSpec{
				PrettyName: "test-workspace",
				Quota: v1alpha1.WorkspaceResourceQuota{
					Instances: 1,
				},
				AutoEnroll: v1alpha1.AutoenrollWithApproval,
			},
		}

		workspaceNA = &v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: workspaceNAName},
			Spec: v1alpha1.WorkspaceSpec{
				PrettyName: "test-workspace",
				Quota: v1alpha1.WorkspaceResourceQuota{
					Instances: 1,
				},
			},
		}

		workspaceIM = &v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: workspaceIMName},
			Spec: v1alpha1.WorkspaceSpec{
				PrettyName: "test-workspace",
				Quota: v1alpha1.WorkspaceResourceQuota{
					Instances: 1,
				},
				AutoEnroll: v1alpha1.AutoenrollImmediate,
			},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			manager,
			fakeManager,
			workspaceWA,
			workspaceNA,
			workspaceIM,
		).Build()

		tnValidator = &webhook.TenantValidator{
			TenantWebhook: webhook.TenantWebhook{
				Client:       fakeClient,
				BypassGroups: bypassGroups,
			}}

		tnWebhook = admission.WithCustomValidator(
			scheme,
			&v1alpha2.Tenant{},
			tnValidator,
		)
	})

	Describe("The TenantValidator.Handle method", func() {
		JustBeforeEach(func() {
			response = tnWebhook.Handle(ctx, request)
		})

		When("the user is an admin/operator", func() {
			BeforeEach(func() {
				request = forgeRequest(admissionv1.Create, &v1alpha2.Tenant{}, nil)
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
				request = forgeRequest(admissionv1.Update, &v1alpha2.Tenant{}, nil)
			})
			It("Should return an error response", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusBadRequest))
				Expect(response.Result.Message).ToNot(BeEmpty())
			})
		})

		When("Tenant is being edited by a non-existent user", func() {
			BeforeEach(func() {
				request = forgeRequest(admissionv1.Update, &v1alpha2.Tenant{}, &v1alpha2.Tenant{})
				request.UserInfo.Username = "invalid-manager"
			})
			It("Should return an error response", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusNotFound))
				Expect(response.Result.Message).ToNot(BeEmpty())
			})
		})

		When("Tenant is being edited by a non-manager user", func() {
			BeforeEach(func() {
				request = forgeRequest(admissionv1.Update, &v1alpha2.Tenant{
					Spec: v1alpha2.TenantSpec{
						Workspaces: []v1alpha2.TenantWorkspaceEntry{{
							Name: testWorkspace,
							Role: v1alpha2.User,
						}},
					}}, &v1alpha2.Tenant{})
				request.UserInfo.Username = fakeManager.Name
				request.UserInfo.Groups = []string{"other-group"}
			})

			It("Should return an error response", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
				Expect(response.Result.Message).ToNot(BeEmpty())
			})
		})
	})

	Describe("The TenantValidator.HandleSelfEdit method", func() {
		var newTenant, oldTenant *v1alpha2.Tenant
		JustBeforeEach(func() {
			response = forgeResponse(tnValidator.HandleSelfEdit(ctx, newTenant, oldTenant))
		})

		When("only public keys are changed", func() {
			BeforeEach(func() {
				oldTenant = &v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{PublicKeys: []string{"some-key"}}}
				newTenant = &v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{PublicKeys: []string{"other-key"}}}
			})
			It("should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("lastLogin is changed within the LastLoginToleration", func() {
			BeforeEach(func() {
				newTenant = &v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{
					LastLogin: metav1.Time{
						Time: time.Now().Add(webhook.LastLoginToleration / 2),
					},
				}}
			})
			It("should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("lastLogin too far from now (abs(lastLogin-now)) > LastLoginToleration)", func() {
			BeforeEach(func() {
				newTenant = &v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{
					LastLogin: metav1.Time{
						Time: time.Now().Add(webhook.LastLoginToleration + time.Millisecond),
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
			WhenBody := func(nt *v1alpha2.Tenant, shouldSucceed bool) func() {
				return func() {
					BeforeEach(func() {
						oldTenant = &v1alpha2.Tenant{}
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
				forgeTenantWithWorkspace(v1alpha2.TenantWorkspaceEntry{
					Name: workspaceNAName,
					Role: v1alpha2.Candidate,
				}),
				false,
			))

			When("autoenroll withApproval and user role", WhenBody(
				forgeTenantWithWorkspaceUser(workspaceWAName),
				false,
			))

			When("autoenroll withApproval and candidate role", WhenBody(
				forgeTenantWithWorkspace(v1alpha2.TenantWorkspaceEntry{
					Name: workspaceWAName,
					Role: v1alpha2.Candidate,
				}),
				true,
			))

			When("autoenroll immediate and user role", WhenBody(
				forgeTenantWithWorkspaceUser(workspaceIMName),
				true,
			))

			When("autoenroll immediate and candidate role", WhenBody(
				forgeTenantWithWorkspace(v1alpha2.TenantWorkspaceEntry{
					Name: workspaceIMName,
					Role: v1alpha2.Candidate,
				}),
				false,
			))
		})

		When("other fields are changed", func() {
			BeforeEach(func() {
				oldTenant = &v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{LastName: "test"}}
				newTenant = &v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{Email: "other"}}
			})
			It("should deny the change", func() {
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
				Expect(response.Result.Reason).NotTo(BeEmpty())
			})
		})
	})

	Describe("The TenantValidator.HandleWorkspaceEdit", func() {
		var newTenant, oldTenant *v1alpha2.Tenant
		var operation admissionv1.Operation
		JustBeforeEach(func() {
			response = forgeResponse(tnValidator.HandleWorkspaceEdit(ctx, newTenant, oldTenant, manager, operation))
		})

		When("manager adds a workspace he manages", func() {
			BeforeEach(func() {
				oldTenant = &v1alpha2.Tenant{}
				newTenant = forgeTenantWithWorkspaceUser(testWorkspace)
				operation = admissionv1.Update
			})
			It("Should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("manager removes a workspace he manages", func() {
			BeforeEach(func() {
				newTenant = &v1alpha2.Tenant{}
				oldTenant = forgeTenantWithWorkspaceUser(testWorkspace)
				operation = admissionv1.Update
			})
			It("Should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})

		When("manager adds a workspace he doesn't manage", func() {
			BeforeEach(func() {
				oldTenant = &v1alpha2.Tenant{}
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
				newTenant = &v1alpha2.Tenant{}
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
				oldTenant = &v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{LastName: "test"}}
				newTenant = &v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{Email: "other"}}
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
				oldTenant = &v1alpha2.Tenant{}
				newTenant = &v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{Email: "other"}}
				operation = admissionv1.Create
			})
			It("should allow the change", func() {
				Expect(response.Allowed).To(BeTrue())
			})
		})
	})

	Describe("The CalculateWorkspacesDiff function", func() {
		type CalcWsDiffCase struct {
			a, b            []v1alpha2.TenantWorkspaceEntry
			expectedChanges []string
		}
		WhenBody := func(cwdc CalcWsDiffCase) func() {
			return func() {
				var actuals []string
				result := webhook.CalculateWorkspacesDiff(
					&v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{Workspaces: cwdc.a}},
					&v1alpha2.Tenant{Spec: v1alpha2.TenantSpec{Workspaces: cwdc.b}},
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
		makeWs := func(name string, role v1alpha2.WorkspaceUserRole) v1alpha2.TenantWorkspaceEntry {
			return v1alpha2.TenantWorkspaceEntry{Name: name, Role: role}
		}

		When("The two tenants have no different workspaces", WhenBody(CalcWsDiffCase{
			a:               []v1alpha2.TenantWorkspaceEntry{makeWs("test-a", v1alpha2.Manager)},
			b:               []v1alpha2.TenantWorkspaceEntry{makeWs("test-a", v1alpha2.Manager)},
			expectedChanges: nil,
		}))

		When("The first tenant has one more workspace", WhenBody(CalcWsDiffCase{
			a:               []v1alpha2.TenantWorkspaceEntry{makeWs("test-a", v1alpha2.Manager)},
			b:               []v1alpha2.TenantWorkspaceEntry{},
			expectedChanges: []string{"test-a"},
		}))

		When("The second tenant has one more workspace", WhenBody(CalcWsDiffCase{
			a:               []v1alpha2.TenantWorkspaceEntry{makeWs("test-a", v1alpha2.Manager)},
			b:               []v1alpha2.TenantWorkspaceEntry{makeWs("test-a", v1alpha2.Manager), makeWs("test-b", v1alpha2.User)},
			expectedChanges: []string{"test-b"},
		}))

		When("There is a difference in roles", WhenBody(CalcWsDiffCase{
			a:               []v1alpha2.TenantWorkspaceEntry{makeWs("test-a", v1alpha2.User)},
			b:               []v1alpha2.TenantWorkspaceEntry{makeWs("test-a", v1alpha2.Manager)},
			expectedChanges: []string{"test-a"},
		}))

	})
	When("Modifying the createPersonalWorkspace spec", func() {
		var oldTenant, newTenant *v1alpha2.Tenant
		BeforeEach(func() {
			oldTenant = &v1alpha2.Tenant{}
			newTenant = oldTenant.DeepCopy()
		})
		When("The user is in bypass group", func() {
			JustBeforeEach(func() {
				request = forgeRequest(admissionv1.Update, newTenant, oldTenant)
				request.UserInfo.Groups = bypassGroups
				response = tnWebhook.Handle(ctx, request)
			})
			When("The personal workspace is being enabled", func() {
				BeforeEach(func() {
					newTenant.Spec.CreatePersonalWorkspace = true
				})
				When("It was disabled before", func() {
					BeforeEach(func() {
						oldTenant.Spec.CreatePersonalWorkspace = false
					})
					It("should allow the change", func() {
						Expect(response.Allowed).To(BeTrue())
					})
				})
				When("It was enabled before", func() {
					BeforeEach(func() {
						oldTenant.Spec.CreatePersonalWorkspace = true
					})
					It("should allow the change", func() {
						Expect(response.Allowed).To(BeTrue())
					})
				})
			})
			When("The personal workspace is being disabled", func() {
				BeforeEach(func() {
					newTenant.Spec.CreatePersonalWorkspace = false
				})
				When("It was disabled before", func() {
					BeforeEach(func() {
						oldTenant.Spec.CreatePersonalWorkspace = false
					})
					It("should allow the change", func() {
						Expect(response.Allowed).To(BeTrue())
					})
				})
				When("It was enabled before", func() {
					BeforeEach(func() {
						oldTenant.Spec.CreatePersonalWorkspace = true
						oldTenant.Status.PersonalNamespace.Created = true
						oldTenant.Status.PersonalNamespace.Name = testTenantPersonalNamespace
					})
					When("The personal workspace has active instances", func() {
						BeforeEach(func() {
							instance := &v1alpha2.Instance{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "test-instance",
									Namespace: testTenantPersonalNamespace,
								},
								Spec: v1alpha2.InstanceSpec{
									Template: v1alpha2.GenericRef{
										Name:      "test-template",
										Namespace: testTenantPersonalNamespace,
									},
								},
							}
							err := tnValidator.Client.Create(ctx, instance)
							Expect(err).NotTo(HaveOccurred())
						})
						It("should deny the change", func() {
							Expect(response.Allowed).To(BeFalse())
							Expect(response.Result.Code).To(BeNumerically("==", http.StatusConflict))
							Expect(response.Result.Reason).NotTo(BeEmpty())
						})
						When("The personal workspace is failing", func() {
							BeforeEach(func() {
								oldTenant.Status.FailingWorkspaces = []string{"personal-workspace"}
								oldTenant.Status.PersonalWorkspaceCreated = false
							})
							It("should deny the change", func() {
								Expect(response.Allowed).To(BeFalse())
								Expect(response.Result.Code).To(BeNumerically("==", http.StatusConflict))
								Expect(response.Result.Reason).NotTo(BeEmpty())
							})
						})
					})
					When("The personal workspace has no active instances", func() {
						It("should allow the change", func() {
							Expect(response.Allowed).To(BeTrue())
						})
					})
					When("The Instances listing errors", func() {
						BeforeEach(func() {
							interceptorClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
								manager,
								fakeManager,
								workspaceWA,
								workspaceNA,
								workspaceIM,
							).WithInterceptorFuncs(interceptor.Funcs{
								List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
									if _, ok := list.(*v1alpha2.InstanceList); ok {
										return fmt.Errorf("error listing templates")
									}
									return client.List(ctx, list, opts...)
								},
							}).Build()
							tnValidator.Client = interceptorClient
						})
						It("Should deny the change", func() {
							Expect(response.Allowed).To(BeFalse())
							Expect(response.Result.Code).To(BeNumerically("==", http.StatusInternalServerError))
							Expect(response.Result.Reason).NotTo(BeEmpty())
						})
					})
				})
			})
		})
		When("The user is in not bypass group", func() {
			JustBeforeEach(func() {
				request = forgeRequest(admissionv1.Update, newTenant, oldTenant)
				response = tnWebhook.Handle(ctx, request)
			})
			When("The personal workspace is being enabled", func() {
				BeforeEach(func() {
					oldTenant.Spec.CreatePersonalWorkspace = false
					newTenant.Spec.CreatePersonalWorkspace = true
				})
				It("should deny the change", func() {
					Expect(response.Allowed).To(BeFalse())
					Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
					Expect(response.Result.Reason).NotTo(BeEmpty())
				})
			})
			When("The personal workspace is being disabled", func() {
				BeforeEach(func() {
					oldTenant.Spec.CreatePersonalWorkspace = true
					newTenant.Spec.CreatePersonalWorkspace = false
				})
				It("should deny the change", func() {
					Expect(response.Allowed).To(BeFalse())
					Expect(response.Result.Code).To(BeNumerically("==", http.StatusForbidden))
					Expect(response.Result.Reason).NotTo(BeEmpty())
				})
			})
		})
	})
})
