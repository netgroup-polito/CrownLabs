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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/instance/webhook"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("InstanceValidator LocalVM volume sources", func() {
	const (
		publicNamespace      = "public-snapshots"
		otherTenantNamespace = "tenant-other-tester"
		bypassGroup          = "system:masters"
		snapshotPVC          = "published-snapshot"
		liveDiskPVC          = "live-disk"
	)

	// volume returns a PVC in the given namespace, marked as a snapshot artifact when published is set.
	volume := func(namespace, name string, published bool) *corev1.PersistentVolumeClaim {
		pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
		if published {
			pvc.Labels = map[string]string{forge.LabelSnapshotArtifactKey: forge.LabelSnapshotArtifactValue}
		}
		return pvc
	}

	type SourceCase struct {
		Image  string
		Groups []string
		// ExpectedError is empty when the request must be admitted.
		ExpectedError string
	}

	DescribeTable("Correctly decides whether the tenant can boot from the referenced volume",
		func(c SourceCase) {
			tenant := &clv1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{Name: testTenant},
				Spec: clv1alpha2.TenantSpec{
					Workspaces: []clv1alpha2.TenantWorkspaceEntry{{Name: testWorkspace, Role: clv1alpha2.User}},
				},
			}
			// A workspace with no quota imposes no limit, so only the source check can reject the request.
			ws := &clv1alpha1.Workspace{ObjectMeta: metav1.ObjectMeta{Name: testWorkspace}}
			tmpl := &clv1alpha2.Template{
				ObjectMeta: metav1.ObjectMeta{Name: testTemplate, Namespace: testWorkspaceNamespace},
				Spec: clv1alpha2.TemplateSpec{
					EnvironmentList: []clv1alpha2.Environment{{
						Name:            testEnvironment,
						EnvironmentType: clv1alpha2.ClassLocalVM,
						Image:           c.Image,
					}},
					WorkspaceRef: clv1alpha2.GenericRef{Name: testWorkspace},
				},
			}
			inst := &clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testNewInstance,
					Namespace: testTenantNamespace,
					Labels:    map[string]string{forge.LabelWorkspaceKey: testWorkspace},
				},
				Spec: clv1alpha2.InstanceSpec{
					Template: clv1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
				},
			}

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				tenant, ws, tmpl,
				volume(testTenantNamespace, snapshotPVC, true),
				volume(testTenantNamespace, liveDiskPVC, false),
				volume(testWorkspaceNamespace, snapshotPVC, true),
				volume(otherTenantNamespace, snapshotPVC, true),
				volume(publicNamespace, liveDiskPVC, false),
			).Build()

			validator := &webhook.InstanceValidator{
				Client:                  fakeClient,
				APIReader:               fakeClient,
				PublicSnapshotNamespace: publicNamespace,
				BypassGroups:            []string{bypassGroup},
			}

			ctx := admission.NewContextWithRequest(context.Background(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UserInfo: authenticationv1.UserInfo{Username: testTenant, Groups: c.Groups},
				},
			})

			_, err := validator.ValidateCreate(ctx, inst)

			if c.ExpectedError == "" {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(err).To(MatchError(ContainSubstring(c.ExpectedError)))
		},
		Entry("When booting a published snapshot of the tenant", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC,
		}),
		Entry("When booting a live disk of the tenant", SourceCase{
			Image: testTenantNamespace + "/" + liveDiskPVC, ExpectedError: "not a published snapshot",
		}),
		Entry("When booting a published snapshot of a workspace the tenant is enrolled in", SourceCase{
			Image: testWorkspaceNamespace + "/" + snapshotPVC,
		}),
		Entry("When booting a published snapshot of another tenant", SourceCase{
			Image: otherTenantNamespace + "/" + snapshotPVC, ExpectedError: "cannot use volume",
		}),
		Entry("When booting an unlabeled volume from the public catalog", SourceCase{
			Image: publicNamespace + "/" + liveDiskPVC,
		}),
		Entry("When booting a volume that does not exist", SourceCase{
			Image: testTenantNamespace + "/missing", ExpectedError: "does not exist",
		}),
		Entry("When the image is malformed", SourceCase{
			Image: "not-a-pvc-reference", ExpectedError: "invalid LocalVM image",
		}),
		Entry("When the requester belongs to a bypass group", SourceCase{
			Image: otherTenantNamespace + "/" + snapshotPVC, Groups: []string{bypassGroup},
		}),
	)
})
