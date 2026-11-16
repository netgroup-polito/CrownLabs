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

package tenant_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/mail"
)

type sentNotification struct {
	template     string
	placeholders mail.Placeholders
}

// fakeMailer records the emails it is asked to send, and fails for the configured workspaces.
type fakeMailer struct {
	sent    []sentNotification
	failFor map[string]bool // keyed by workspace pretty name
}

func (m *fakeMailer) SendCrownLabsMail(_ context.Context, template string, ph *mail.Placeholders) error {
	if m.failFor[ph.WorkspacePrettyName] {
		return errors.New("smtp server unavailable")
	}
	m.sent = append(m.sent, sentNotification{template: template, placeholders: *ph})
	return nil
}

var _ = Describe("Workspace notifications", func() {
	ws1 := &clv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "ws1"},
		Spec: clv1alpha1.WorkspaceSpec{
			PrettyName: "Workspace One",
			AutoEnroll: clv1alpha1.AutoenrollWithApproval,
		},
	}
	ws2 := &clv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "ws2"},
		Spec:       clv1alpha1.WorkspaceSpec{PrettyName: "Workspace Two"},
	}
	bws1 := &clv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "base-ws1"},
		Spec:       clv1alpha1.WorkspaceSpec{PrettyName: "Base Workspace"},
	}

	var mailer *fakeMailer

	addedNotification := func(prettyName string, role clv1alpha2.WorkspaceUserRole) sentNotification {
		return sentNotification{
			template: tenant.WorkspaceAddedMailTemplatePath,
			placeholders: mail.Placeholders{
				TenantName:          tnName,
				TenantEmail:         "test@tenant.example",
				WorkspacePrettyName: prettyName,
				WorkspaceRole:       string(role),
			},
		}
	}

	removedNotification := func(prettyName string) sentNotification {
		return sentNotification{
			template: tenant.WorkspaceRemovedMailTemplatePath,
			placeholders: mail.Placeholders{
				TenantName:          tnName,
				TenantEmail:         "test@tenant.example",
				WorkspacePrettyName: prettyName,
			},
		}
	}

	getTenant := func() *clv1alpha2.Tenant {
		tn := &clv1alpha2.Tenant{}
		Expect(cl.Get(ctx, client.ObjectKey{Name: tnName}, tn)).To(Succeed())
		return tn
	}

	// notifiedAnnotation returns the notified workspaces annotation, failing if it is missing.
	notifiedAnnotation := func() string {
		annotations := getTenant().Annotations
		ExpectWithOffset(1, annotations).To(HaveKey(forge.NotifiedWorkspacesAnnotation))
		return annotations[forge.NotifiedWorkspacesAnnotation]
	}

	setWorkspaces := func(workspaces ...clv1alpha2.TenantWorkspaceEntry) {
		tn := getTenant()
		tn.Spec.Workspaces = workspaces
		Expect(cl.Update(ctx, tn)).To(Succeed())
	}

	reconcileTenant := func() error {
		_, err := tenantReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: tnName},
		})
		return err
	}

	markAsNew := func(notified string) {
		tnResource.Annotations = map[string]string{forge.NotifiedWorkspacesAnnotation: notified}
	}

	BeforeEach(func() {
		mailer = &fakeMailer{failFor: map[string]bool{}}
		notificationMailer = mailer

		addObjToObjectsList(ws1)
		addObjToObjectsList(ws2)
		addObjToObjectsList(bws1)
	})

	AfterEach(func() {
		removeObjFromObjectsList(ws1)
		removeObjFromObjectsList(ws2)
		removeObjFromObjectsList(bws1)
	})

	Context("When the tenant was created before notifications were introduced", func() {
		BeforeEach(func() {
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
				{Name: "ws2", Role: clv1alpha2.Manager},
				{Name: "base-ws1", Role: clv1alpha2.User},
			}
		})

		It("Should record its current workspaces without sending any email", func() {
			Expect(mailer.sent).To(BeEmpty())
			Expect(notifiedAnnotation()).To(Equal("ws1,ws2"))
		})
	})

	Context("When a new tenant is enrolled in some workspaces", func() {
		BeforeEach(func() {
			markAsNew("")
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
				{Name: "ws2", Role: clv1alpha2.Manager},
				{Name: "base-ws1", Role: clv1alpha2.User},
			}
		})

		It("Should send one email per workspace, except the base ones", func() {
			Expect(mailer.sent).To(ConsistOf(
				addedNotification("Workspace One", clv1alpha2.User),
				addedNotification("Workspace Two", clv1alpha2.Manager),
			))
		})

		It("Should record the notified workspaces", func() {
			Expect(notifiedAnnotation()).To(Equal("ws1,ws2"))
		})

		It("Should not send the emails again at the next reconcile", func() {
			Expect(reconcileTenant()).To(Succeed())
			Expect(mailer.sent).To(HaveLen(2))
		})
	})

	Context("When the tenant is a candidate for a workspace", func() {
		BeforeEach(func() {
			markAsNew("")
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.Candidate},
			}
		})

		It("Should not send any email until the candidature is approved", func() {
			Expect(mailer.sent).To(BeEmpty())
			Expect(notifiedAnnotation()).To(Equal(""))

			setWorkspaces(clv1alpha2.TenantWorkspaceEntry{Name: "ws1", Role: clv1alpha2.User})
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(ConsistOf(addedNotification("Workspace One", clv1alpha2.User)))
			Expect(notifiedAnnotation()).To(Equal("ws1"))
		})

		It("Should not send any email if the candidature is rejected", func() {
			setWorkspaces()
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(BeEmpty())
			Expect(notifiedAnnotation()).To(Equal(""))
		})
	})

	Context("When the tenant is removed from a notified workspace", func() {
		BeforeEach(func() {
			markAsNew("ws1,ws2")
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws2", Role: clv1alpha2.User},
			}
		})

		It("Should notify the removal and drop the workspace from the annotation", func() {
			Expect(mailer.sent).To(ConsistOf(removedNotification("Workspace One")))
			Expect(notifiedAnnotation()).To(Equal("ws2"))
		})

		It("Should not notify the removal again at the next reconcile", func() {
			Expect(reconcileTenant()).To(Succeed())
			Expect(mailer.sent).To(HaveLen(1))
		})

		It("Should send a new email when the tenant is added back", func() {
			setWorkspaces(
				clv1alpha2.TenantWorkspaceEntry{Name: "ws1", Role: clv1alpha2.User},
				clv1alpha2.TenantWorkspaceEntry{Name: "ws2", Role: clv1alpha2.User},
			)
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(ConsistOf(
				removedNotification("Workspace One"),
				addedNotification("Workspace One", clv1alpha2.User),
			))
			Expect(notifiedAnnotation()).To(Equal("ws1,ws2"))
		})
	})

	Context("When the tenant is removed from a workspace that has been deleted", func() {
		BeforeEach(func() {
			markAsNew("ws1,ws9") // ws9 does not exist anymore
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
			}
		})

		It("Should notify the removal using the workspace name", func() {
			Expect(mailer.sent).To(ConsistOf(removedNotification("ws9")))
			Expect(notifiedAnnotation()).To(Equal("ws1"))
		})
	})

	Context("When notifying a removal fails", func() {
		BeforeEach(func() {
			markAsNew("ws1,ws2")
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws2", Role: clv1alpha2.User},
			}
			mailer.failFor["Workspace One"] = true
			tnReconcileErrExpected = HaveOccurred()
		})

		It("Should keep the workspace in the annotation", func() {
			Expect(mailer.sent).To(BeEmpty())
			Expect(notifiedAnnotation()).To(Equal("ws1,ws2"))
		})

		It("Should notify the removal at the next reconcile", func() {
			mailer.failFor = map[string]bool{}
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(ConsistOf(removedNotification("Workspace One")))
			Expect(notifiedAnnotation()).To(Equal("ws2"))
		})
	})

	Context("When a workspace is failing", func() {
		BeforeEach(func() {
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws3", Role: clv1alpha2.Manager}, // ws3 does not exist
			}
		})

		Context("and it has not been notified yet", func() {
			BeforeEach(func() { markAsNew("") })

			It("Should neither send an email nor record the workspace", func() {
				Expect(mailer.sent).To(BeEmpty())
				Expect(notifiedAnnotation()).To(Equal(""))
			})
		})

		Context("and it has already been notified", func() {
			BeforeEach(func() { markAsNew("ws3") })

			It("Should keep the workspace in the annotation, without notifying a removal", func() {
				Expect(mailer.sent).To(BeEmpty())
				Expect(notifiedAnnotation()).To(Equal("ws3"))
			})
		})
	})

	Context("When sending an email fails", func() {
		BeforeEach(func() {
			markAsNew("")
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
				{Name: "ws2", Role: clv1alpha2.User},
			}
			mailer.failFor["Workspace Two"] = true
			tnReconcileErrExpected = HaveOccurred()
		})

		It("Should record only the workspaces whose email was sent", func() {
			Expect(mailer.sent).To(ConsistOf(addedNotification("Workspace One", clv1alpha2.User)))
			Expect(notifiedAnnotation()).To(Equal("ws1"))
		})

		It("Should not affect the tenant readiness", func() {
			Expect(getTenant().Status.Ready).To(BeTrue())
		})

		It("Should send the failed email at the next reconcile", func() {
			mailer.failFor = map[string]bool{}
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(ConsistOf(
				addedNotification("Workspace One", clv1alpha2.User),
				addedNotification("Workspace Two", clv1alpha2.User),
			))
			Expect(notifiedAnnotation()).To(Equal("ws1,ws2"))
		})
	})

	Context("When notifications are disabled", func() {
		BeforeEach(func() {
			notificationMailer = nil
			markAsNew("")
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
			}
		})

		It("Should record the workspaces without sending any email", func() {
			Expect(mailer.sent).To(BeEmpty())
			Expect(notifiedAnnotation()).To(Equal("ws1"))
		})

		It("Should not send emails for those workspaces once notifications are enabled", func() {
			tenantReconciler.Mailer = mailer
			Expect(reconcileTenant()).To(Succeed())
			Expect(mailer.sent).To(BeEmpty())
		})

		It("Should drop removed workspaces without sending any email", func() {
			setWorkspaces()
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(BeEmpty())
			Expect(notifiedAnnotation()).To(Equal(""))
		})
	})
})
