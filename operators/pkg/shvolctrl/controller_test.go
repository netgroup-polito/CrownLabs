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

package shvolctrl_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// The following are integration tests aiming to verify that the sharedvolume controller
// reconcile method correctly creates the various resources it has to manage,
// by running some combinations of configuration. This is not exaustive
// as a more granular coverage is achieved through the unit tests.
var _ = Describe("The sharedvolume-controller Reconcile method", Ordered, func() {
	ctx := context.Background()
	var (
		shvol         clv1alpha2.SharedVolume
		environment   clv1alpha2.Environment
		environmentSV clv1alpha2.Environment
		pv            corev1.PersistentVolume
		sc            storagev1.StorageClass
	)

	const (
		testName = "test-reconciler"
	)

	RunReconciler := func() error {
		_, err := shvolReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: forge.NamespacedNameFromSharedVolume(&shvol),
		})
		if err != nil {
			return err
		}
		return k8sClient.Get(ctx, forge.NamespacedNameFromSharedVolume(&shvol), &shvol)
	}

	PVCNamespacedName := func(shvol *clv1alpha2.SharedVolume) types.NamespacedName {
		return types.NamespacedName{
			Name:      fmt.Sprintf("shvol-%s", shvol.Name),
			Namespace: shvol.Namespace,
		}
	}

	PVCObjectMeta := func(shvol *clv1alpha2.SharedVolume) metav1.ObjectMeta {
		nsn := PVCNamespacedName(shvol)
		return metav1.ObjectMeta{
			Name:      nsn.Name,
			Namespace: nsn.Namespace,
		}
	}

	JobNamespacedName := func(shvol *clv1alpha2.SharedVolume) types.NamespacedName {
		pvcnsn := PVCNamespacedName(shvol)
		return types.NamespacedName{
			Name:      fmt.Sprintf("%s-provision", pvcnsn.Name),
			Namespace: pvcnsn.Namespace,
		}
	}

	BeforeAll(func() {
		ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testName, Labels: whiteListMap}}
		sc = storagev1.StorageClass{
			ObjectMeta:           metav1.ObjectMeta{Name: pvcStorageClass},
			AllowVolumeExpansion: ptr.To(true),
			Provisioner:          "my-nfs-a",
		}
		Expect(k8sClient.Create(ctx, &ns)).To(Succeed())
		Expect(k8sClient.Create(ctx, &sc)).To(Succeed())
	})

	JustBeforeEach(func() {
		shvol = clv1alpha2.SharedVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName,
				Namespace: testName,
			},
			Spec: clv1alpha2.SharedVolumeSpec{
				PrettyName: "My Test Drive",
				Size:       *resource.NewScaledQuantity(1, resource.Giga),
			},
		}
		environment = clv1alpha2.Environment{
			Name:            "app",
			Image:           "some-image:v0",
			EnvironmentType: clv1alpha2.ClassVM,
			Persistent:      false,
			GuiEnabled:      true,
			Resources: clv1alpha2.EnvironmentResources{
				CPU:                   1,
				ReservedCPUPercentage: 20,
				Memory:                *resource.NewScaledQuantity(1, resource.Giga),
				Disk:                  *resource.NewScaledQuantity(10, resource.Giga),
			},
			Mode:               clv1alpha2.ModeStandard,
			MountMyDriveVolume: false,
		}
		environmentSV = clv1alpha2.Environment{
			Name:            "app",
			Image:           "some-image:v0",
			EnvironmentType: clv1alpha2.ClassVM,
			Persistent:      false,
			GuiEnabled:      true,
			Resources: clv1alpha2.EnvironmentResources{
				CPU:                   1,
				ReservedCPUPercentage: 20,
				Memory:                *resource.NewScaledQuantity(1, resource.Giga),
				Disk:                  *resource.NewScaledQuantity(10, resource.Giga),
			},
			Mode:               clv1alpha2.ModeStandard,
			MountMyDriveVolume: false,
			SharedVolumeMounts: []clv1alpha2.SharedVolumeMountInfo{
				{
					SharedVolumeRef: clv1alpha2.GenericRef{
						Name:      shvol.Name,
						Namespace: shvol.Namespace,
					},
					MountPath: "/mnt/path",
					ReadOnly:  true,
				},
			},
		}
		pv = corev1.PersistentVolume{
			ObjectMeta: PVCObjectMeta(&shvol),
			Spec: corev1.PersistentVolumeSpec{
				Capacity: corev1.ResourceList{
					"storage": *resource.NewScaledQuantity(1, resource.Giga),
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteMany,
				},
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{
						Driver: "nfs.example.com",
						VolumeAttributes: map[string]string{
							"server":    "my-nfs-cluster",
							"clusterID": "local",
							"share":     "/nfs/path",
						},
						VolumeHandle: "/nfs/path",
					},
				},
			},
		}

		Expect(k8sClient.Create(ctx, &shvol)).To(Succeed())
		Expect(RunReconciler()).To(Succeed())
	})

	Context("The shared volume has valid parameters", func() {
		var pvc corev1.PersistentVolumeClaim

		It("Should create PVC", func() {
			Expect(shvol.Status.Phase).To(Equal(clv1alpha2.SharedVolumePhasePending))

			Expect(k8sClient.Get(ctx, PVCNamespacedName(&shvol), &pvc)).To(Succeed())
			Expect(pvc.Status.Phase).To(Equal(corev1.ClaimPending))
		})

		When("the PV has been created, and the PVC is bound", func() {
			var job batchv1.Job

			JustBeforeEach(func() {
				// Create PV, Update PVC Status
				Expect(k8sClient.Create(ctx, &pv)).To(Succeed())

				Expect(k8sClient.Get(ctx, PVCNamespacedName(&shvol), &pvc)).To(Succeed())
				pvc.Status.Phase = corev1.ClaimBound
				Expect(k8sClient.Status().Update(ctx, &pvc)).To(Succeed())
				pvc.Spec.VolumeName = pv.Name
				Expect(k8sClient.Update(ctx, &pvc)).To(Succeed())

				Expect(RunReconciler()).To(Succeed())
			})

			It("Should create the Provisioning Job", func() {
				Expect(shvol.Status.Phase).To(Equal(clv1alpha2.SharedVolumePhaseProvisioning))

				Expect(k8sClient.Get(ctx, JobNamespacedName(&shvol), &job)).To(Succeed())
			})

			When("the Job has successfully completed", func() {
				JustBeforeEach(func() {
					// Update Job Status
					Expect(k8sClient.Get(ctx, JobNamespacedName(&shvol), &job)).To(Succeed())
					job.Status.Succeeded = 1
					Expect(k8sClient.Status().Update(ctx, &job)).To(Succeed())

					Expect(RunReconciler()).To(Succeed())
				})

				It("Should correctly set labels and finalizers", func() {
					Expect(shvol.Status.Phase).To(Equal(clv1alpha2.SharedVolumePhaseReady))

					// Labels on PVC
					Expect(k8sClient.Get(ctx, PVCNamespacedName(&shvol), &pvc)).To(Succeed())
					Expect(pvc.Labels[forge.ProvisionJobLabel]).To(Equal(forge.ProvisionJobValueOk))

					// Labels and Finalizers on ShVol
					Expect(shvol.Labels).To(HaveKeyWithValue(ContainSubstring("managed-by"), ContainSubstring("instance")))
					Expect(ctrlUtil.ContainsFinalizer(&shvol, clv1alpha2.ShVolCtrlFinalizerName))
				})

				Context("The shared volume deletion should be handled correctly", func() {
					JustBeforeEach(func() {
						// Try to Delete ShVol
						Expect(k8sClient.Delete(ctx, &shvol)).To(Succeed())
						_ = RunReconciler()
					})

					When("there is no template that mounts the shared volume", func() {
						var template clv1alpha2.Template

						BeforeEach(func() {
							// Create Template with no ShVol
							template = clv1alpha2.Template{
								ObjectMeta: metav1.ObjectMeta{Name: testName, Namespace: testName},
								Spec: clv1alpha2.TemplateSpec{
									WorkspaceRef:    clv1alpha2.GenericRef{Name: testName},
									EnvironmentList: []clv1alpha2.Environment{environment},
								},
							}
							Expect(k8sClient.Create(ctx, &template)).To(Succeed())
						})

						It("Should smoothly delete the shared volume", func() {
							err := k8sClient.Get(ctx, forge.NamespacedNameFromSharedVolume(&shvol), &shvol)
							Expect(kerrors.IsNotFound(err)).To(BeTrue())
						})

						JustAfterEach(func() {
							// Delete Template with no ShVol
							_ = k8sClient.Delete(ctx, &template)
						})
					})

					When("there is one template that mounts the shared volume", func() {
						var templateSV clv1alpha2.Template

						BeforeEach(func() {
							// Create Template with ShVol
							templateSV = clv1alpha2.Template{
								ObjectMeta: metav1.ObjectMeta{Name: testName, Namespace: testName},
								Spec: clv1alpha2.TemplateSpec{
									WorkspaceRef:    clv1alpha2.GenericRef{Name: testName},
									EnvironmentList: []clv1alpha2.Environment{environmentSV},
								},
							}
							Expect(k8sClient.Create(ctx, &templateSV)).To(Succeed())
						})

						It("Should not delete the shared volume and transition to Deleting", func() {
							err := k8sClient.Get(ctx, forge.NamespacedNameFromSharedVolume(&shvol), &shvol)
							Expect(err).ToNot(HaveOccurred())
							Expect(shvol.Status.Phase).To(Equal(clv1alpha2.SharedVolumePhaseDeleting))
						})

						JustAfterEach(func() {
							// Delete Template with ShVol
							_ = k8sClient.Delete(ctx, &templateSV)
							_ = RunReconciler()
						})
					})
				})

				Context("The shared volume update should be handled correctly", func() {
					When("Its size is increased", func() {
						JustBeforeEach(func() {
							// Increase ShVol size
							shvol.Spec.Size = *resource.NewScaledQuantity(2, resource.Giga)
							Expect(k8sClient.Update(ctx, &shvol)).To(Succeed())
							Expect(RunReconciler()).To(Succeed())
						})

						It("Should enforce the PVC accordingly", func() {
							Expect(shvol.Status.Phase).To(Equal(clv1alpha2.SharedVolumePhaseReady))

							Expect(k8sClient.Get(ctx, PVCNamespacedName(&shvol), &pvc)).To(Succeed())
							Expect(pvc.Spec.Resources.Requests.Storage().String()).To(Equal(resource.NewScaledQuantity(2, resource.Giga).String()))
						})
					})

					When("Its size is decreased", func() {
						JustBeforeEach(func() {
							// Decrease ShVol size
							shvol.Spec.Size = *resource.NewScaledQuantity(500, resource.Mega)
							Expect(k8sClient.Update(ctx, &shvol)).To(Succeed())
							Expect(RunReconciler()).To(Succeed())
						})

						It("Should transition to phase Error", func() {
							Expect(k8sClient.Get(ctx, forge.NamespacedNameFromSharedVolume(&shvol), &shvol)).To(Succeed())
							Expect(shvol.Status.Phase).To(Equal(clv1alpha2.SharedVolumePhaseError))

							Expect(k8sClient.Get(ctx, PVCNamespacedName(&shvol), &pvc)).To(Succeed())
							Expect(pvc.Spec.Resources.Requests.Storage().String()).To(Equal(resource.NewScaledQuantity(1, resource.Giga).String()))
						})
					})
				})
			})

			AfterEach(func() {
				Expect(k8sClient.Get(ctx, PVCNamespacedName(&shvol), &pv)).To(Succeed())
				ctrlUtil.RemoveFinalizer(&pv, "kubernetes.io/pv-protection")
				_ = k8sClient.Update(ctx, &pv)
				_ = k8sClient.Delete(ctx, &pv)
			})
		})

		AfterEach(func() {
			Expect(k8sClient.Get(ctx, PVCNamespacedName(&shvol), &pvc)).To(Succeed())
			ctrlUtil.RemoveFinalizer(&pvc, "kubernetes.io/pvc-protection")
			_ = k8sClient.Update(ctx, &pvc)
			_ = k8sClient.Delete(ctx, &pvc)
		})
	})

	/**
	* This test case is being skipped since the ResourceQuota admission webhook seems not to be enabled.
	 */
	Context("The shared volume exceeds an existing ResourceQuota", func() {
		var quota, quotaSkip corev1.ResourceQuota
		var pvcSkip corev1.PersistentVolumeClaim

		BeforeEach(func() {
			// Test if ResourceQuotas are being enforced
			quotaSkip = corev1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{Name: testName + "-skiptest", Namespace: testName},
				Spec: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{
						corev1.ResourceRequestsStorage: *resource.NewScaledQuantity(0, resource.Giga),
					},
				},
			}
			pvcSkip = corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Name: testName + "-skiptest", Namespace: testName},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: *resource.NewScaledQuantity(1, resource.Giga),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, &quotaSkip)).To(Succeed())
			errSkip := k8sClient.Create(ctx, &pvcSkip)
			if errSkip == nil {
				Skip("ResourceQuota is not being enforced")
			}

			// Create ResourceQuota
			quota = corev1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{Name: testName, Namespace: testName},
				Spec: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{
						corev1.ResourceName(shvolReconciler.PVCStorageClass + ".storageclass.storage.k8s.io/requests.storage"): *resource.NewScaledQuantity(0, resource.Giga),
					},
				},
			}
			Expect(k8sClient.Create(ctx, &quota)).To(Succeed())
		})

		It("Should not create PVC and transition to phase ResourceQuotaExceeded", func() {
			var pvc corev1.PersistentVolumeClaim
			err := k8sClient.Get(ctx, PVCNamespacedName(&shvol), &pvc)
			fmt.Printf("-- GOT PVC: %s | %s\n", *pvc.Spec.StorageClassName, pvc.Spec.Resources.Requests.Storage().String())
			Expect(kerrors.IsNotFound(err)).To(BeTrue())

			Expect(shvol.Status.Phase).To(Equal(clv1alpha2.SharedVolumePhaseResourceQuotaExceeded))
		})

		JustAfterEach(func() {
			_ = k8sClient.Delete(ctx, &pvcSkip)
			_ = k8sClient.Delete(ctx, &quotaSkip)

			// Delete ResourceQuota
			_ = k8sClient.Delete(ctx, &quota)
		})
	})

	JustAfterEach(func() {
		_ = k8sClient.Delete(ctx, &shvol)
		_ = RunReconciler()
	})
})
