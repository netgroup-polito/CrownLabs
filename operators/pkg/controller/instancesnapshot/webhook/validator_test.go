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

package webhook_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/instancesnapshot/webhook"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("InstanceSnapshotValidator", func() {
	var validator *webhook.InstanceSnapshotValidator

	BeforeEach(func() {
		tenant := &clv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{Name: testTenant},
			Spec: clv1alpha2.TenantSpec{
				Workspaces: []clv1alpha2.TenantWorkspaceEntry{{Name: testWorkspace, Role: clv1alpha2.User}},
			},
		}

		validator = &webhook.InstanceSnapshotValidator{
			Client:                  fake.NewClientBuilder().WithScheme(scheme).WithObjects(tenant).Build(),
			PublicSnapshotNamespace: testPublicNamespace,
			PublisherGroup:          testPublisherGroup,
			BypassGroups:            []string{testBypassGroup},
		}
	})

	// requestFrom returns a context carrying an admission request issued by the given identity.
	requestFrom := func(username string, groups ...string) context.Context {
		return admission.NewContextWithRequest(context.Background(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				UserInfo: authenticationv1.UserInfo{Username: username, Groups: groups},
			},
		})
	}

	// snapshotOf returns a snapshot, created in the tenant namespace and labeled with the given owner
	// (none if empty), of the instance living in the given namespace.
	snapshotOf := func(owner, sourceNamespace string) *clv1alpha2.InstanceSnapshot {
		snapshot := &clv1alpha2.InstanceSnapshot{
			ObjectMeta: metav1.ObjectMeta{Name: testSnapshot, Namespace: testTenantNamespace},
			Spec: clv1alpha2.InstanceSnapshotSpec{
				Instance:    clv1alpha2.GenericRef{Name: testInstance, Namespace: sourceNamespace},
				Environment: testEnvironment,
			},
		}
		if owner != "" {
			snapshot.Labels = map[string]string{forge.LabelTenantKey: owner}
		}
		return snapshot
	}

	Describe("The ValidateCreate function", func() {
		type CreateCase struct {
			Username        string
			Groups          []string
			Owner           string
			SourceNamespace string
			// Destination is the namespace the snapshot is created in, the tenant one if empty.
			Destination string
			// ExpectedError is empty when the request must be admitted.
			ExpectedError string
		}

		DescribeTable("Correctly decides whether the source instance can be snapshotted",
			func(c CreateCase) {
				snapshot := snapshotOf(c.Owner, c.SourceNamespace)
				if c.Destination != "" {
					snapshot.Namespace = c.Destination
				}

				_, err := validator.ValidateCreate(requestFrom(c.Username, c.Groups...), snapshot)

				if c.ExpectedError == "" {
					Expect(err).NotTo(HaveOccurred())
					return
				}
				Expect(err).To(MatchError(ContainSubstring(c.ExpectedError)))
			},
			Entry("When the instance lives in the tenant namespace", CreateCase{
				Username: testTenant, Owner: testTenant, SourceNamespace: testTenantNamespace,
			}),
			Entry("When the instance lives in a workspace the tenant is enrolled in", CreateCase{
				Username: testTenant, Owner: testTenant, SourceNamespace: testWorkspaceNamespace,
			}),
			Entry("When the instance lives in the public catalog", CreateCase{
				Username: testTenant, Owner: testTenant, SourceNamespace: testPublicNamespace,
			}),
			Entry("When the instance belongs to another tenant", CreateCase{
				Username: testTenant, Owner: testTenant, SourceNamespace: testOtherTenantNamespace,
				ExpectedError: "cannot snapshot instance",
			}),
			Entry("When the source namespace is left empty", CreateCase{
				Username: testTenant, Owner: testTenant, SourceNamespace: "",
				ExpectedError: "must be set explicitly",
			}),
			Entry("When the tenant label is missing", CreateCase{
				Username: testTenant, SourceNamespace: testTenantNamespace,
				ExpectedError: "must be set to the tenant creating the snapshot",
			}),
			Entry("When the tenant label names somebody else", CreateCase{
				Username: testTenant, Owner: testOtherTenant, SourceNamespace: testTenantNamespace,
				ExpectedError: "but the snapshot is being created by",
			}),
			Entry("When the requester has no Tenant behind it", CreateCase{
				Username: testOtherTenant, Owner: testOtherTenant, SourceNamespace: testTenantNamespace,
				ExpectedError: "failed to get tenant",
			}),
			Entry("When a publisher publishes into the public catalog", CreateCase{
				Username: testTenant, Groups: []string{testPublisherGroup}, Owner: testTenant,
				SourceNamespace: testTenantNamespace, Destination: testPublicNamespace,
			}),
			Entry("When somebody else publishes into the public catalog", CreateCase{
				Username: testTenant, Owner: testTenant,
				SourceNamespace: testTenantNamespace, Destination: testPublicNamespace,
				ExpectedError: "only snapshot publishers",
			}),
			Entry("When the requester belongs to a bypass group", CreateCase{
				Username: testTenant, Groups: []string{testBypassGroup}, SourceNamespace: testOtherTenantNamespace,
			}),
			Entry("When a bypass group publishes into the public catalog", CreateCase{
				Username: testTenant, Groups: []string{testBypassGroup},
				SourceNamespace: testOtherTenantNamespace, Destination: testPublicNamespace,
			}),
		)
	})

	Describe("The ValidateUpdate function", func() {
		It("Should accept updates leaving the source untouched, even from identities with no Tenant", func() {
			// This is what the snapshot controller does when it adds its finalizer: it acts as a
			// service account, which would be rejected were the scope checked again.
			oldSnapshot := snapshotOf(testTenant, testTenantNamespace)
			newSnapshot := oldSnapshot.DeepCopy()
			newSnapshot.Finalizers = []string{"instancesnapshot.crownlabs.polito.it/finalizer"}

			_, err := validator.ValidateUpdate(requestFrom(testServiceAccount), oldSnapshot, newSnapshot)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should reject repointing the source at somebody else's instance", func() {
			_, err := validator.ValidateUpdate(requestFrom(testTenant),
				snapshotOf(testTenant, testTenantNamespace), snapshotOf(testTenant, testOtherTenantNamespace))
			Expect(err).To(MatchError("InstanceSnapshot spec is immutable"))
		})

		It("Should reject changing snapshot metadata", func() {
			oldSnapshot := snapshotOf(testTenant, testTenantNamespace)
			newSnapshot := oldSnapshot.DeepCopy()
			newSnapshot.Spec.Description = "changed after creation"

			_, err := validator.ValidateUpdate(requestFrom(testTenant), oldSnapshot, newSnapshot)
			Expect(err).To(MatchError("InstanceSnapshot spec is immutable"))
		})

		It("Should reject handing the snapshot over to another tenant", func() {
			_, err := validator.ValidateUpdate(requestFrom(testTenant),
				snapshotOf(testTenant, testTenantNamespace), snapshotOf(testOtherTenant, testTenantNamespace))
			Expect(err).To(MatchError(ContainSubstring("is immutable")))
		})

		It("Should reject removing the tenant label", func() {
			_, err := validator.ValidateUpdate(requestFrom(testTenant),
				snapshotOf(testTenant, testTenantNamespace), snapshotOf("", testTenantNamespace))
			Expect(err).To(MatchError(ContainSubstring("is immutable")))
		})
	})
})
