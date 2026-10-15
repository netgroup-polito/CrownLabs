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
	"k8s.io/apimachinery/pkg/api/resource"
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
	volume := func(namespace, name string, published bool, requested, capacity string) *corev1.PersistentVolumeClaim {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Spec: corev1.PersistentVolumeClaimSpec{
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(requested)},
				},
			},
		}
		if capacity != "" {
			pvc.Status.Capacity = corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(capacity)}
		}
		if published {
			pvc.Labels = map[string]string{forge.LabelSnapshotArtifactKey: forge.LabelSnapshotArtifactValue}
		}
		return pvc
	}

	type SourceCase struct {
		Image           string
		Groups          []string
		Disk            string
		SourceRequested string
		SourceCapacity  string
		// ExpectedError is empty when the request must be admitted.
		ExpectedError string
	}

	DescribeTable("Correctly decides whether the tenant can boot from the referenced volume",
		func(c SourceCase) {
			if c.Disk == "" {
				c.Disk = "10Gi"
			}
			if c.SourceRequested == "" {
				c.SourceRequested = "10Gi"
			}
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
			tmpl.Spec.EnvironmentList[0].Resources.Disk = resource.MustParse(c.Disk)
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
				volume(testTenantNamespace, snapshotPVC, true, c.SourceRequested, c.SourceCapacity),
				volume(testTenantNamespace, liveDiskPVC, false, c.SourceRequested, c.SourceCapacity),
				volume(testWorkspaceNamespace, snapshotPVC, true, c.SourceRequested, c.SourceCapacity),
				volume(otherTenantNamespace, snapshotPVC, true, c.SourceRequested, c.SourceCapacity),
				volume(publicNamespace, liveDiskPVC, false, c.SourceRequested, c.SourceCapacity),
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

			// Authorization bypass must not require an administrator to have a Tenant resource.
			if len(c.Groups) > 0 {
				Expect(fakeClient.Delete(ctx, tenant)).To(Succeed())
			}

			By("validating creation, whether initially running or stopped")
			for _, running := range []bool{false, true} {
				inst.Spec.Running = running
				_, err := validator.ValidateCreate(ctx, inst)
				if c.ExpectedError == "" {
					Expect(err).NotTo(HaveOccurred())
				} else {
					Expect(err).To(MatchError(ContainSubstring(c.ExpectedError)))
				}
			}

			By("validating the transition from stopped to running")
			oldInst := inst.DeepCopy()
			oldInst.Spec.Running = false
			_, err := validator.ValidateUpdate(ctx, oldInst, inst)
			if c.ExpectedError == "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(MatchError(ContainSubstring(c.ExpectedError)))
			}

			By("allowing updates that do not start the instance")
			_, err = validator.ValidateUpdate(ctx, inst, inst.DeepCopy())
			Expect(err).NotTo(HaveOccurred())
			_, err = validator.ValidateUpdate(ctx, inst, oldInst)
			Expect(err).NotTo(HaveOccurred())
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
		Entry("When the requested disk is smaller than the snapshot", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, Disk: "5Gi",
			ExpectedError: "requests disk size 5Gi, but source volume " + testTenantNamespace + "/" + snapshotPVC + " requires at least 10Gi",
		}),
		Entry("When the requested disk is larger than the snapshot", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, Disk: "20Gi",
		}),
		Entry("When the requested disk equals the snapshot in different units", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, Disk: "10240Mi",
		}),
		Entry("When decimal units make the requested disk smaller", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, Disk: "10G", ExpectedError: "requires at least 10Gi",
		}),
		Entry("When the requested disk is zero", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, Disk: "0", ExpectedError: "requires at least 10Gi",
		}),
		Entry("When the snapshot capacity exceeds its storage request", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, SourceCapacity: "20Gi", ExpectedError: "requires at least 20Gi",
		}),
		Entry("When the requested disk equals the snapshot capacity", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, Disk: "20Gi", SourceCapacity: "20Gi",
		}),
		Entry("When the snapshot has a pending expansion", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, SourceRequested: "20Gi", SourceCapacity: "10Gi",
			ExpectedError: "requires at least 20Gi",
		}),
		Entry("When only the snapshot capacity is available", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, SourceRequested: "0", SourceCapacity: "10Gi",
		}),
		Entry("When the snapshot size is unknown", SourceCase{
			Image: testTenantNamespace + "/" + snapshotPVC, SourceRequested: "0", ExpectedError: "cannot determine the disk size",
		}),
		Entry("When the disk is too small for a workspace snapshot", SourceCase{
			Image: testWorkspaceNamespace + "/" + snapshotPVC, Disk: "5Gi", ExpectedError: "requires at least 10Gi",
		}),
		Entry("When the disk is too small for a public image", SourceCase{
			Image: publicNamespace + "/" + liveDiskPVC, Disk: "5Gi", ExpectedError: "requires at least 10Gi",
		}),
		Entry("When a public image is missing", SourceCase{
			Image: publicNamespace + "/missing", ExpectedError: "does not exist",
		}),
		Entry("When authorization is bypassed but the disk is too small", SourceCase{
			Image: otherTenantNamespace + "/" + snapshotPVC, Groups: []string{bypassGroup}, Disk: "5Gi",
			ExpectedError: "requires at least 10Gi",
		}),
		Entry("When authorization is bypassed for an unlabeled volume", SourceCase{
			Image: testTenantNamespace + "/" + liveDiskPVC, Groups: []string{bypassGroup},
		}),
	)
})
