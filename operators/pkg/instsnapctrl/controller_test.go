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

package instsnapctrl

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	virtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

func TestInstanceSnapshotController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InstanceSnapshot Controller Suite")
}

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

		reconciler := &InstanceSnapshotReconciler{
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

		reconciler := &InstanceSnapshotReconciler{
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

		reconciler := &InstanceSnapshotReconciler{
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

		reconciler := &InstanceSnapshotReconciler{
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
