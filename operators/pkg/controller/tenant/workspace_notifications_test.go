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
	"maps"
	"slices"
	"strings"
	"time"

	gocloak13 "github.com/Nerzal/gocloak/v13"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/mock"
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
	// oldTimestamp is the value of the annotations of the workspaces notified before each test.
	const oldTimestamp = "2026-01-01T00:00:00Z"

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

	// notified returns the notified workspaces, with the values of their annotations.
	notified := func() map[string]string {
		workspaces := map[string]string{}
		for key, value := range getTenant().Annotations {
			if name, ok := strings.CutPrefix(key, forge.NotifiedWorkspaceAnnotationPrefix); ok {
				workspaces[name] = value
			}
		}
		return workspaces
	}

	// haveWorkspaces matches the notified workspaces by name, whatever the values of their annotations.
	haveWorkspaces := func(names ...string) gomegatypes.GomegaMatcher {
		return WithTransform(func(workspaces map[string]string) []string {
			return slices.Sorted(maps.Keys(workspaces))
		}, Equal(names))
	}

	// recentTimestamp matches an RFC 3339 timestamp of the last minute, i.e., of an addition recorded by the test.
	recentTimestamp := func() gomegatypes.GomegaMatcher {
		return WithTransform(func(value string) (time.Time, error) {
			return time.Parse(time.RFC3339, value)
		}, BeTemporally("~", time.Now(), time.Minute))
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

	// markAsNotified marks the tenant as already notified about the given workspaces, before each test.
	markAsNotified := func(notifiedWorkspaces ...string) {
		tnResource.Annotations = map[string]string{}
		for _, name := range notifiedWorkspaces {
			tnResource.Annotations[forge.NotifiedWorkspaceAnnotationPrefix+name] = oldTimestamp
		}
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

	Context("When a new tenant is enrolled in some workspaces", func() {
		BeforeEach(func() {
			tnResource.Annotations = map[string]string{"other": "value"}
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

		It("Should record the notified workspaces with the time of the email", func() {
			Expect(notified()).To(SatisfyAll(
				HaveLen(2),
				HaveKeyWithValue("ws1", recentTimestamp()),
				HaveKeyWithValue("ws2", recentTimestamp()),
			))
		})

		It("Should not send the emails again at the next reconcile", func() {
			Expect(reconcileTenant()).To(Succeed())
			Expect(mailer.sent).To(HaveLen(2))
		})

		It("Should preserve the other annotations", func() {
			Expect(getTenant().Annotations).To(HaveKeyWithValue("other", "value"))
		})
	})

	Context("When the workspaces of an existing tenant have been marked as notified before enabling notifications", func() {
		BeforeEach(func() {
			markAsNotified("ws1", "ws2") // as recorded while notifications are disabled
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
				{Name: "ws2", Role: clv1alpha2.Manager},
			}
		})

		It("Should not send any email", func() {
			Expect(mailer.sent).To(BeEmpty())
			Expect(notified()).To(Equal(map[string]string{"ws1": oldTimestamp, "ws2": oldTimestamp}))
		})

		It("Should notify the later changes", func() {
			setWorkspaces(clv1alpha2.TenantWorkspaceEntry{Name: "ws1", Role: clv1alpha2.User})
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(ConsistOf(removedNotification("Workspace Two")))
			Expect(notified()).To(Equal(map[string]string{"ws1": oldTimestamp}))
		})
	})

	Context("When the tenant is a candidate for a workspace", func() {
		BeforeEach(func() {
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.Candidate},
			}
		})

		It("Should not send any email until the candidature is approved", func() {
			Expect(mailer.sent).To(BeEmpty())
			Expect(notified()).To(BeEmpty())

			setWorkspaces(clv1alpha2.TenantWorkspaceEntry{Name: "ws1", Role: clv1alpha2.User})
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(ConsistOf(addedNotification("Workspace One", clv1alpha2.User)))
			Expect(notified()).To(SatisfyAll(HaveLen(1), HaveKeyWithValue("ws1", recentTimestamp())))
		})

		It("Should not send any email if the candidature is rejected", func() {
			setWorkspaces()
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(BeEmpty())
			Expect(notified()).To(BeEmpty())
		})
	})

	Context("When the tenant is removed from a notified workspace", func() {
		BeforeEach(func() {
			markAsNotified("ws1", "ws2")
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws2", Role: clv1alpha2.User},
			}
		})

		It("Should notify the removal and delete the annotation of the workspace", func() {
			Expect(mailer.sent).To(ConsistOf(removedNotification("Workspace One")))
			Expect(notified()).To(Equal(map[string]string{"ws2": oldTimestamp}))
		})

		It("Should not notify the removal again at the next reconcile", func() {
			Expect(reconcileTenant()).To(Succeed())
			Expect(mailer.sent).To(HaveLen(1))
		})

		It("Should send a new email when the tenant is added back, recording its time", func() {
			setWorkspaces(
				clv1alpha2.TenantWorkspaceEntry{Name: "ws1", Role: clv1alpha2.User},
				clv1alpha2.TenantWorkspaceEntry{Name: "ws2", Role: clv1alpha2.User},
			)
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(ConsistOf(
				removedNotification("Workspace One"),
				addedNotification("Workspace One", clv1alpha2.User),
			))
			Expect(notified()).To(SatisfyAll(
				HaveLen(2),
				HaveKeyWithValue("ws1", recentTimestamp()),
				HaveKeyWithValue("ws2", oldTimestamp),
			))
		})
	})

	Context("When the tenant is removed from a workspace that has been deleted", func() {
		BeforeEach(func() {
			markAsNotified("ws1", "ws9") // ws9 does not exist anymore
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
			}
		})

		It("Should notify the removal using the workspace name", func() {
			Expect(mailer.sent).To(ConsistOf(removedNotification("ws9")))
			Expect(notified()).To(haveWorkspaces("ws1"))
		})
	})

	Context("When notifying a removal fails", func() {
		BeforeEach(func() {
			markAsNotified("ws1", "ws2")
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws2", Role: clv1alpha2.User},
			}
			mailer.failFor["Workspace One"] = true
			tnReconcileErrExpected = HaveOccurred()
		})

		It("Should keep the annotation of the workspace", func() {
			Expect(mailer.sent).To(BeEmpty())
			Expect(notified()).To(Equal(map[string]string{"ws1": oldTimestamp, "ws2": oldTimestamp}))
		})

		It("Should notify the removal at the next reconcile", func() {
			mailer.failFor = map[string]bool{}
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(ConsistOf(removedNotification("Workspace One")))
			Expect(notified()).To(haveWorkspaces("ws2"))
		})
	})

	Context("When a workspace is failing", func() {
		BeforeEach(func() {
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws3", Role: clv1alpha2.Manager}, // ws3 does not exist
			}
		})

		Context("and it has not been notified yet", func() {
			It("Should neither send an email nor record the workspace", func() {
				Expect(mailer.sent).To(BeEmpty())
				Expect(notified()).To(BeEmpty())
			})
		})

		Context("and it has already been notified", func() {
			BeforeEach(func() { markAsNotified("ws3") })

			It("Should keep the annotation of the workspace, without notifying a removal", func() {
				Expect(mailer.sent).To(BeEmpty())
				Expect(notified()).To(Equal(map[string]string{"ws3": oldTimestamp}))
			})
		})
	})

	Context("When sending an email fails", func() {
		BeforeEach(func() {
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
				{Name: "ws2", Role: clv1alpha2.User},
			}
			mailer.failFor["Workspace Two"] = true
			tnReconcileErrExpected = HaveOccurred()
		})

		It("Should record only the workspaces whose email was sent", func() {
			Expect(mailer.sent).To(ConsistOf(addedNotification("Workspace One", clv1alpha2.User)))
			Expect(notified()).To(haveWorkspaces("ws1"))
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
			Expect(notified()).To(haveWorkspaces("ws1", "ws2"))
		})
	})

	Context("When notifications are disabled", func() {
		BeforeEach(func() {
			notificationMailer = nil
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
			}
		})

		It("Should record the workspaces with the time they were added, without sending any email", func() {
			Expect(mailer.sent).To(BeEmpty())
			Expect(notified()).To(SatisfyAll(HaveLen(1), HaveKeyWithValue("ws1", recentTimestamp())))
		})

		It("Should not send emails for those workspaces once notifications are enabled", func() {
			tenantReconciler.Mailer = mailer
			Expect(reconcileTenant()).To(Succeed())
			Expect(mailer.sent).To(BeEmpty())
		})

		It("Should delete the annotations of removed workspaces without sending any email", func() {
			setWorkspaces()
			Expect(reconcileTenant()).To(Succeed())

			Expect(mailer.sent).To(BeEmpty())
			Expect(notified()).To(BeEmpty())
		})
	})

	Context("When the tenant has not verified its email in Keycloak", func() {
		var emailVerified bool

		BeforeEach(func() {
			emailVerified = false
			tnResource.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{
				{Name: "ws1", Role: clv1alpha2.User},
				{Name: "ws2", Role: clv1alpha2.Manager},
			}

			keycloakActor = mock.NewMockKeycloakActorIface(mockCtrl)
			keycloakActor.EXPECT().IsInitialized().Return(true).AnyTimes()
			keycloakActor.EXPECT().GetUser(gomock.Any(), tnName).DoAndReturn(
				func(context.Context, string) (*gocloak13.User, error) {
					return &gocloak13.User{ID: gocloak13.StringP("user-id"), EmailVerified: gocloak13.BoolP(emailVerified)}, nil
				}).AnyTimes()
			keycloakActor.EXPECT().GetUserRoles(gomock.Any(), "user-id").Return(nil, nil).AnyTimes()
			keycloakActor.EXPECT().GetRole(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, name string) (*gocloak13.Role, error) {
					return &gocloak13.Role{Name: gocloak13.StringP(name)}, nil
				}).AnyTimes()
			keycloakActor.EXPECT().AddUserToRoles(gomock.Any(), "user-id", gomock.Any()).Return(nil).AnyTimes()

			// The reconcile is run by the test, after disabling WaitUserVerification: otherwise,
			// it would stop before the notifications for unverified tenants anyway.
			runReconcile = false
		})

		AfterEach(func() {
			runReconcile = true
		})

		It("Should notify the workspaces only once the email is verified, even without WaitUserVerification", func() {
			tenantReconciler.WaitUserVerification = false

			Expect(reconcileTenant()).To(Succeed())
			Expect(mailer.sent).To(BeEmpty())
			Expect(notified()).To(BeEmpty())

			emailVerified = true
			Expect(reconcileTenant()).To(Succeed())
			Expect(mailer.sent).To(ConsistOf(
				addedNotification("Workspace One", clv1alpha2.User),
				addedNotification("Workspace Two", clv1alpha2.Manager),
			))
			Expect(notified()).To(haveWorkspaces("ws1", "ws2"))
		})
	})
})
