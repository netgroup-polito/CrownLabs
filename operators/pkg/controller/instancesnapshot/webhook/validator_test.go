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

	// snapshotOf returns a snapshot, created in the tenant namespace, of the instance living in the
	// given namespace.
	snapshotOf := func(sourceNamespace string) *clv1alpha2.InstanceSnapshot {
		return &clv1alpha2.InstanceSnapshot{
			ObjectMeta: metav1.ObjectMeta{Name: testSnapshot, Namespace: testTenantNamespace},
			Spec: clv1alpha2.InstanceSnapshotSpec{
				Instance:    clv1alpha2.GenericRef{Name: testInstance, Namespace: sourceNamespace},
				Environment: testEnvironment,
			},
		}
	}

	Describe("The ValidateCreate function", func() {
		type CreateCase struct {
			Username        string
			Groups          []string
			SourceNamespace string
			// ExpectedError is empty when the request must be admitted.
			ExpectedError string
		}

		DescribeTable("Correctly decides whether the source instance can be snapshotted",
			func(c CreateCase) {
				_, err := validator.ValidateCreate(requestFrom(c.Username, c.Groups...), snapshotOf(c.SourceNamespace))

				if c.ExpectedError == "" {
					Expect(err).NotTo(HaveOccurred())
					return
				}
				Expect(err).To(MatchError(ContainSubstring(c.ExpectedError)))
			},
			Entry("When the instance lives in the tenant namespace", CreateCase{
				Username: testTenant, SourceNamespace: testTenantNamespace,
			}),
			Entry("When the instance lives in a workspace the tenant is enrolled in", CreateCase{
				Username: testTenant, SourceNamespace: testWorkspaceNamespace,
			}),
			Entry("When the instance lives in the public catalog", CreateCase{
				Username: testTenant, SourceNamespace: testPublicNamespace,
			}),
			Entry("When the instance belongs to another tenant", CreateCase{
				Username: testTenant, SourceNamespace: testOtherTenantNamespace,
				ExpectedError: "cannot snapshot instance",
			}),
			Entry("When the source namespace is left empty", CreateCase{
				Username: testTenant, SourceNamespace: "",
				ExpectedError: "must be set explicitly",
			}),
			Entry("When the requester has no Tenant behind it", CreateCase{
				Username: testServiceAccount, SourceNamespace: testTenantNamespace,
				ExpectedError: "failed to get tenant",
			}),
			Entry("When the requester belongs to a bypass group", CreateCase{
				Username: testTenant, Groups: []string{testBypassGroup}, SourceNamespace: testOtherTenantNamespace,
			}),
		)
	})

	Describe("The ValidateUpdate function", func() {
		It("Should accept updates leaving the source untouched, even from identities with no Tenant", func() {
			// This is what the snapshot controller does when it adds its finalizer: it acts as a
			// service account, which would be rejected were the scope checked again.
			oldSnapshot := snapshotOf(testTenantNamespace)
			newSnapshot := oldSnapshot.DeepCopy()
			newSnapshot.Finalizers = []string{"instancesnapshot.crownlabs.polito.it/finalizer"}

			_, err := validator.ValidateUpdate(requestFrom(testServiceAccount), oldSnapshot, newSnapshot)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should reject repointing the source at somebody else's instance", func() {
			_, err := validator.ValidateUpdate(requestFrom(testTenant),
				snapshotOf(testTenantNamespace), snapshotOf(testOtherTenantNamespace))
			Expect(err).To(MatchError(ContainSubstring("cannot snapshot instance")))
		})
	})
})
