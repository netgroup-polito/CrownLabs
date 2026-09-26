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

package instsnapctrl_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	virtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instsnapctrl"
)

func TestInstanceSnapshotController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InstanceSnapshot Controller Suite")
}

var _ = Describe("The InstanceSnapshot controller", func() {
	const (
		tenantNamespace = "tenant-tester"
		snapshotName    = "test-snapshot"
		snapshotUID     = types.UID("snapshot-uid")
		instanceName    = "test-instance"
		templateName    = "test-template"
		environment     = "env1"
		finalizer       = "instancesnapshot.crownlabs.polito.it/finalizer"

		// The DataVolume and its PVC share the snapshot's name.
		artifactName = snapshotName
		// The source disk of the instance being snapshotted.
		sourcePVCName = instanceName + "-" + environment
	)

	var scheme *runtime.Scheme

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())
		Expect(cdiv1beta1.AddToScheme(scheme)).To(Succeed())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(virtv1.AddToScheme(scheme)).To(Succeed())
	})

	// snapshot returns a snapshot of the stopped instance below.
	snapshot := func() *clv1alpha2.InstanceSnapshot {
		return &clv1alpha2.InstanceSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Name:       snapshotName,
				Namespace:  tenantNamespace,
				UID:        snapshotUID,
				Finalizers: []string{finalizer},
			},
			Spec: clv1alpha2.InstanceSnapshotSpec{
				Instance:    clv1alpha2.GenericRef{Name: instanceName, Namespace: tenantNamespace},
				Environment: environment,
			},
		}
	}

	// deletingSnapshot returns a snapshot being deleted, whose status points at the given DataVolume.
	deletingSnapshot := func(dvNamespace, dvName string) *clv1alpha2.InstanceSnapshot {
		snap := snapshot()
		snap.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		snap.Status.Artifact.DataVolumeRef = clv1alpha2.GenericRef{Name: dvName, Namespace: dvNamespace}
		return snap
	}

	// stoppedInstance returns the source instance, in the state the controller requires to proceed.
	stoppedInstance := func() *clv1alpha2.Instance {
		return &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: tenantNamespace},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: templateName, Namespace: tenantNamespace},
				Running:  false,
			},
		}
	}

	// singleEnvTemplate returns the template of the instance: snapshots are allowed on single
	// environment templates only.
	singleEnvTemplate := func() *clv1alpha2.Template {
		return &clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: tenantNamespace},
			Spec: clv1alpha2.TemplateSpec{
				EnvironmentList: []clv1alpha2.Environment{{
					Name:            environment,
					EnvironmentType: clv1alpha2.ClassVM,
					Persistent:      true,
				}},
			},
		}
	}

	// persistentVolumeClaim returns a PVC of the given name, in the tenant namespace.
	persistentVolumeClaim := func(name string) *corev1.PersistentVolumeClaim {
		return &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: tenantNamespace},
			Spec: corev1.PersistentVolumeClaimSpec{
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
				},
			},
		}
	}

	// dataVolume returns a DataVolume controlled by the object of the given kind and UID.
	dataVolume := func(namespace, name, ownerKind string, ownerUID types.UID) *cdiv1beta1.DataVolume {
		return &cdiv1beta1.DataVolume{ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: clv1alpha2.GroupVersion.String(),
				Kind:       ownerKind,
				Name:       "owner",
				UID:        ownerUID,
				Controller: ptr.To(true),
			}},
		}}
	}

	// reconcile runs the controller against the given objects, and returns the client to inspect them.
	reconcile := func(objs ...client.Object) client.Client {
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).
			WithStatusSubresource(&clv1alpha2.InstanceSnapshot{}).Build()
		reconciler := &instsnapctrl.InstanceSnapshotReconciler{
			Client:         cl,
			Scheme:         scheme,
			EventsRecorder: record.NewFakeRecorder(10),
		}

		_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: tenantNamespace, Name: snapshotName},
		})
		Expect(err).NotTo(HaveOccurred())

		return cl
	}

	Describe("Cloning the instance disk", func() {
		It("Should create a DataVolume owned by the snapshot and marked as an artifact", func() {
			cl := reconcile(snapshot(), stoppedInstance(), singleEnvTemplate(), persistentVolumeClaim(sourcePVCName))

			var dv cdiv1beta1.DataVolume
			Expect(cl.Get(context.Background(),
				types.NamespacedName{Namespace: tenantNamespace, Name: artifactName}, &dv)).To(Succeed())
			Expect(dv.Labels).To(HaveKeyWithValue(forge.LabelSnapshotArtifactKey, forge.LabelSnapshotArtifactValue))
			Expect(dv.Spec.Source.PVC.Namespace).To(Equal(tenantNamespace))
			Expect(dv.Spec.Source.PVC.Name).To(Equal(sourcePVCName))

			var snap clv1alpha2.InstanceSnapshot
			Expect(cl.Get(context.Background(),
				types.NamespacedName{Namespace: tenantNamespace, Name: snapshotName}, &snap)).To(Succeed())
			Expect(metav1.IsControlledBy(&dv, &snap)).To(BeTrue())
		})
	})

	Describe("Resolving the resources of an instance whose name contains a dot", func() {
		// The instance controller replaces dots with dashes when naming the resources of an
		// environment: the instance "foo.bar" owns "foo-bar-<environment>", not "foo.bar-<environment>".
		const (
			dottedInstanceName = "foo.bar"
			canonicalEnvName   = "foo-bar-" + environment
		)

		dottedSnapshot := func() *clv1alpha2.InstanceSnapshot {
			snap := snapshot()
			snap.Spec.Instance = clv1alpha2.GenericRef{Name: dottedInstanceName, Namespace: tenantNamespace}
			return snap
		}

		dottedInstance := func() *clv1alpha2.Instance {
			inst := stoppedInstance()
			inst.Name = dottedInstanceName
			return inst
		}

		It("Should clone the disk the instance controller actually created", func() {
			cl := reconcile(dottedSnapshot(), dottedInstance(), singleEnvTemplate(),
				persistentVolumeClaim(canonicalEnvName))

			var dv cdiv1beta1.DataVolume
			Expect(cl.Get(context.Background(),
				types.NamespacedName{Namespace: tenantNamespace, Name: artifactName}, &dv)).To(Succeed())
			Expect(dv.Spec.Source.PVC.Name).To(Equal(canonicalEnvName))
		})

		It("Should still see the VMI of an instance being shut down, and refuse to snapshot it", func() {
			// The VMI has not terminated yet: cloning now would copy a disk still being written to.
			vmi := &virtv1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{Name: canonicalEnvName, Namespace: tenantNamespace},
			}
			cl := fake.NewClientBuilder().WithScheme(scheme).
				WithObjects(dottedSnapshot(), dottedInstance(), singleEnvTemplate(), vmi,
					persistentVolumeClaim(canonicalEnvName)).
				WithStatusSubresource(&clv1alpha2.InstanceSnapshot{}).Build()
			reconciler := &instsnapctrl.InstanceSnapshotReconciler{
				Client:         cl,
				Scheme:         scheme,
				EventsRecorder: record.NewFakeRecorder(10),
			}

			_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: tenantNamespace, Name: snapshotName},
			})
			Expect(err).To(MatchError(ContainSubstring("is still running")))

			err = cl.Get(context.Background(),
				types.NamespacedName{Namespace: tenantNamespace, Name: artifactName}, &cdiv1beta1.DataVolume{})
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("Completing the snapshot", func() {
		// The clone is over: the artifact PVC exists, and the snapshot must be usable as a VM image.
		var succeededDataVolume *cdiv1beta1.DataVolume

		BeforeEach(func() {
			succeededDataVolume = dataVolume(tenantNamespace, artifactName, "InstanceSnapshot", snapshotUID)
			succeededDataVolume.Status.Phase = cdiv1beta1.Succeeded
		})

		It("Should label the artifact PVC, so that it can be booted from", func() {
			cl := reconcile(snapshot(), stoppedInstance(), singleEnvTemplate(), succeededDataVolume, persistentVolumeClaim(artifactName))

			var pvc corev1.PersistentVolumeClaim
			Expect(cl.Get(context.Background(),
				types.NamespacedName{Namespace: tenantNamespace, Name: artifactName}, &pvc)).To(Succeed())
			Expect(pvc.Labels).To(HaveKeyWithValue(forge.LabelSnapshotArtifactKey, forge.LabelSnapshotArtifactValue))
		})

		It("Should report the snapshot as completed", func() {
			cl := reconcile(snapshot(), stoppedInstance(), singleEnvTemplate(), succeededDataVolume, persistentVolumeClaim(artifactName))

			var snap clv1alpha2.InstanceSnapshot
			Expect(cl.Get(context.Background(),
				types.NamespacedName{Namespace: tenantNamespace, Name: snapshotName}, &snap)).To(Succeed())
			Expect(snap.Status.Phase).To(Equal(clv1alpha2.SnapshotPhaseCompleted))
			Expect(snap.Status.Artifact.DataVolumeRef.Name).To(Equal(artifactName))
			Expect(snap.Status.Artifact.DataVolumeRef.Namespace).To(Equal(tenantNamespace))
		})
	})

	Describe("Deleting the snapshot", func() {
		It("Should delete the DataVolume the snapshot controls", func() {
			dv := dataVolume(tenantNamespace, artifactName, "InstanceSnapshot", snapshotUID)
			cl := reconcile(deletingSnapshot(dv.Namespace, dv.Name), dv)

			err := cl.Get(context.Background(), client.ObjectKeyFromObject(dv), &cdiv1beta1.DataVolume{})
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
		})

		It("Should leave untouched a DataVolume the snapshot does not control", func() {
			// The status points at the disk of the VM of another tenant, controlled by its Instance:
			// this is what tampering with the status subresource would look like.
			dv := dataVolume("tenant-victim", "instance-abcde-env1", "Instance", "instance-uid")
			cl := reconcile(deletingSnapshot(dv.Namespace, dv.Name), dv)

			Expect(cl.Get(context.Background(), client.ObjectKeyFromObject(dv), &cdiv1beta1.DataVolume{})).To(Succeed())
		})

		It("Should complete the deletion when the DataVolume is already gone", func() {
			reconcile(deletingSnapshot(tenantNamespace, "missing"))
		})
	})
})

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	Expect(clv1alpha2.AddToScheme(s)).To(Succeed())
	Expect(corev1.AddToScheme(s)).To(Succeed())
	Expect(virtv1.AddToScheme(s)).To(Succeed())
	Expect(cdiv1beta1.AddToScheme(s)).To(Succeed())
	return s
}

const (
	testSnapshotName      = "test-snapshot"
	testSnapshotNamespace = "test-ns"
	testInstanceName      = "test-instance"
	testInstanceNamespace = "test-instance-ns"
	testTemplateName      = "test-template"
	testTemplateNamespace = "test-template-ns"
	testEnvironmentName   = "env-one"
)

func newSnapshot() *clv1alpha2.InstanceSnapshot {
	return &clv1alpha2.InstanceSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSnapshotName,
			Namespace: testSnapshotNamespace,
			UID:       "abcde-12345",
		},
		Spec: clv1alpha2.InstanceSnapshotSpec{
			Instance: clv1alpha2.GenericRef{
				Name:      testInstanceName,
				Namespace: testInstanceNamespace,
			},
			Environment: testEnvironmentName,
		},
	}
}

func newInstance(running bool) *clv1alpha2.Instance {
	return &clv1alpha2.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testInstanceName,
			Namespace: testInstanceNamespace,
		},
		Spec: clv1alpha2.InstanceSpec{
			Template: clv1alpha2.GenericRef{
				Name:      testTemplateName,
				Namespace: testTemplateNamespace,
			},
			Tenant:  clv1alpha2.GenericRef{Name: "test-tenant"},
			Running: running,
		},
	}
}

func newTemplate(envCount int) *clv1alpha2.Template {
	envList := make([]clv1alpha2.Environment, envCount)
	for i := 0; i < envCount; i++ {
		envList[i] = clv1alpha2.Environment{
			Name:            testEnvironmentName,
			Image:           "test-image",
			EnvironmentType: clv1alpha2.ClassVM,
			Persistent:      true,
			Resources:       clv1alpha2.EnvironmentResources{},
		}
	}
	return &clv1alpha2.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testTemplateName,
			Namespace: testTemplateNamespace,
		},
		Spec: clv1alpha2.TemplateSpec{
			PrettyName:      "Test Template",
			Description:     "A test template",
			EnvironmentList: envList,
		},
	}
}

func newSourcePVC() *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testInstanceName + "-" + testEnvironmentName,
			Namespace: testInstanceNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}
}

var _ = Describe("InstanceSnapshot Controller - Multi-Env Template Validation", func() {
	var (
		ctx context.Context
		s   *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		s = newScheme()
	})

	It("should fail snapshot on multi-env template (2 environments)", func() {
		snapshot := newSnapshot()
		instance := newInstance(false)
		template := newTemplate(2)

		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(snapshot, instance, template).
			WithStatusSubresource(snapshot).
			Build()

		reconciler := &instsnapctrl.InstanceSnapshotReconciler{
			Client:         fakeClient,
			Scheme:         s,
			EventsRecorder: record.NewFakeRecorder(10),
		}

		result, err := reconciler.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      testSnapshotName,
				Namespace: testSnapshotNamespace,
			},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))

		// Verify the snapshot was marked as Failed
		var updatedSnapshot clv1alpha2.InstanceSnapshot
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Name:      testSnapshotName,
			Namespace: testSnapshotNamespace,
		}, &updatedSnapshot)).To(Succeed())
		Expect(updatedSnapshot.Status.Phase).To(Equal(clv1alpha2.SnapshotPhaseFailed))
	})

	It("should fail snapshot on zero-env template (0 environments)", func() {
		snapshot := newSnapshot()
		instance := newInstance(false)
		template := newTemplate(0)

		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(snapshot, instance, template).
			WithStatusSubresource(snapshot).
			Build()

		reconciler := &instsnapctrl.InstanceSnapshotReconciler{
			Client:         fakeClient,
			Scheme:         s,
			EventsRecorder: record.NewFakeRecorder(10),
		}

		result, err := reconciler.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      testSnapshotName,
				Namespace: testSnapshotNamespace,
			},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))

		// Verify the snapshot was marked as Failed
		var updatedSnapshot clv1alpha2.InstanceSnapshot
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Name:      testSnapshotName,
			Namespace: testSnapshotNamespace,
		}, &updatedSnapshot)).To(Succeed())
		Expect(updatedSnapshot.Status.Phase).To(Equal(clv1alpha2.SnapshotPhaseFailed))
	})

	It("should fail snapshot when template is not found", func() {
		snapshot := newSnapshot()
		instance := newInstance(false)
		// Do not create the template

		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(snapshot, instance).
			WithStatusSubresource(snapshot).
			Build()

		reconciler := &instsnapctrl.InstanceSnapshotReconciler{
			Client:         fakeClient,
			Scheme:         s,
			EventsRecorder: record.NewFakeRecorder(10),
		}

		result, err := reconciler.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      testSnapshotName,
				Namespace: testSnapshotNamespace,
			},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))

		// Verify the snapshot was marked as Failed
		var updatedSnapshot clv1alpha2.InstanceSnapshot
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Name:      testSnapshotName,
			Namespace: testSnapshotNamespace,
		}, &updatedSnapshot)).To(Succeed())
		Expect(updatedSnapshot.Status.Phase).To(Equal(clv1alpha2.SnapshotPhaseFailed))
	})

	It("should allow snapshot on single-env template (1 environment)", func() {
		snapshot := newSnapshot()
		instance := newInstance(false)
		template := newTemplate(1)
		sourcePVC := newSourcePVC()

		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(snapshot, instance, template, sourcePVC).
			WithStatusSubresource(snapshot).
			Build()

		reconciler := &instsnapctrl.InstanceSnapshotReconciler{
			Client:         fakeClient,
			Scheme:         s,
			EventsRecorder: record.NewFakeRecorder(10),
		}

		result, err := reconciler.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      testSnapshotName,
				Namespace: testSnapshotNamespace,
			},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))

		// Verify the snapshot was NOT marked as Failed — it should proceed past the template check.
		// With a single-env template, the reconciler should continue to create the DataVolume clone.
		var updatedSnapshot clv1alpha2.InstanceSnapshot
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Name:      testSnapshotName,
			Namespace: testSnapshotNamespace,
		}, &updatedSnapshot)).To(Succeed())
		Expect(updatedSnapshot.Status.Phase).ToNot(Equal(clv1alpha2.SnapshotPhaseFailed))
	})
})
