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
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("MyDrive management", func() {
	pvcNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mydrive-pvcs",
		},
	}

	personalNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-" + tnName,
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector": "test",
			},
		},
	}

	BeforeEach(func() {
		// Set namespace status to created
		tnResource.Status.PersonalNamespace = clv1alpha2.NameCreated{
			Name:    "tenant-" + tnName,
			Created: true,
		}

		addObjToObjectsList(pvcNs)
		addObjToObjectsList(personalNs)
	})

	AfterEach(func() {
		// Clean up the MyDrive PVC namespace
		removeObjFromObjectsList(pvcNs)
		removeObjFromObjectsList(personalNs)
	})

	Context("When tenant needs MyDrive resources", func() {
		It("Should create the PVC in the MyDrive namespace", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      forge.MyDrivePVCName(tnName),
				Namespace: "mydrive-pvcs",
			}, pvc, BeTrue(), timeout, interval)

			Expect(pvc.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(pvc.Annotations).To(HaveKeyWithValue(forge.AuthorizationAnnotationKey, strings.ReplaceAll(forge.MyDriveAuthorizationAnnotationValue, "{tenant-id}", tnName)))
			Expect(pvc.Spec.AccessModes).To(ContainElement(corev1.ReadWriteMany))
			Expect(pvc.Spec.StorageClassName).ToNot(BeNil())
			Expect(*pvc.Spec.StorageClassName).To(Equal("nfs"))

			// Verify resource requests
			Expect(pvc.Spec.Resources.Requests).ToNot(BeEmpty())
			storageRequest := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			expectedStorage := resource.MustParse("5Gi")
			Expect(storageRequest).To(Equal(expectedStorage))
		})

		Context("When PVC is bound", func() {
			// Create a PVC and set it to bound status
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      forge.MyDrivePVCName(tnName),
					Namespace: "mydrive-pvcs",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Status: corev1.PersistentVolumeClaimStatus{
					Phase: corev1.ClaimBound,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "pv-" + tnName,
				},
			}

			// Create a PV with NFS information
			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pv-" + tnName,
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "nfs.csi.k8s.io",
							VolumeAttributes: map[string]string{
								"server":    "nfs.example.com",
								"clusterID": "test-cluster",
								"share":     "/exports/user/" + tnName,
							},
						},
					},
				},
			}

			BeforeEach(func() {
				addObjToObjectsList(pvc)
				addObjToObjectsList(pv)
			})

			AfterEach(func() {
				removeObjFromObjectsList(pvc)
				removeObjFromObjectsList(pv)
			})

			It("Should create the PVC Mirror in the tenant namespace", func() {
				mirror := &corev1.PersistentVolumeClaim{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{
					Name:      forge.MyDrivePVCMirrorName(tnName),
					Namespace: "tenant-" + tnName,
				}, mirror, BeTrue(), timeout, interval)

				Expect(mirror.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
				Expect(mirror.Labels).To(HaveKeyWithValue(forge.LabelVolumeTypeKey, forge.VolumeTypeValueMirror))
				Expect(mirror.Spec.DataSourceRef.Name).To(Equal(tnName + "-drive"))
				Expect(*mirror.Spec.DataSourceRef.Namespace).To(Equal("mydrive-pvcs"))
				Expect(mirror.Spec.DataSourceRef.Kind).To(Equal("PersistentVolumeClaim"))
			})

			It("Should create the provisioning job if not already run", func() {
				job := &batchv1.Job{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{
					Name:      forge.MyDrivePVCName(tnName) + "-provision",
					Namespace: "mydrive-pvcs",
				}, job, BeTrue(), timeout, interval)

				Expect(job.Spec.Template.Spec.Containers).To(HaveLen(1))

				// Check for the updated PVC label
				pvc := &corev1.PersistentVolumeClaim{}
				Expect(cl.Get(ctx, client.ObjectKey{
					Name:      forge.MyDrivePVCName(tnName),
					Namespace: "mydrive-pvcs",
				}, pvc)).To(Succeed())

				Expect(pvc.Labels).To(HaveKeyWithValue(forge.ProvisionJobLabel, forge.ProvisionJobValuePending))
			})

			Context("When provisioning job succeeds", func() {
				// Create a PVC with provisioning label already set to pending
				pvc := &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      forge.MyDrivePVCName(tnName),
						Namespace: "mydrive-pvcs",
						Labels: map[string]string{
							"crownlabs.polito.it/operator-selector": "test",
							forge.ProvisionJobLabel:                 forge.ProvisionJobValuePending,
						},
					},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimBound,
					},
				}

				// Create a job that has succeeded
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:              forge.MyDrivePVCName(tnName) + "-provision",
						Namespace:         "mydrive-pvcs",
						CreationTimestamp: metav1.Now(),
					},
					Status: batchv1.JobStatus{
						Succeeded: 1,
					},
				}

				BeforeEach(func() {
					// Clean up the outer version of the object
					removeObjFromObjectsList(pvc)

					addObjToObjectsList(pvc)
					addObjToObjectsList(job)
				})

				AfterEach(func() {
					// Clean up the job after test
					removeObjFromObjectsList(pvc)
					removeObjFromObjectsList(job)
				})

				It("Should update the PVC label to ok", func() {
					Eventually(func() string {
						gotPvc := &corev1.PersistentVolumeClaim{}
						err := cl.Get(ctx, forge.NamespacedNameFromObject(pvc), gotPvc)
						if err != nil {
							return ""
						}
						return gotPvc.Labels[forge.ProvisionJobLabel]
					}, timeout, interval).Should(Equal(forge.ProvisionJobValueOk))
				})
			})

			Context("When provisioning job fails", func() {
				// Create a PVC with provisioning label already set to pending
				pvc := &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      forge.MyDrivePVCName(tnName),
						Namespace: "mydrive-pvcs",
						Labels: map[string]string{
							"crownlabs.polito.it/operator-selector": "test",
							forge.ProvisionJobLabel:                 forge.ProvisionJobValuePending,
						},
					},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimBound,
					},
				}

				// Create a job that has failed
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:              forge.MyDrivePVCName(tnName) + "-provision",
						Namespace:         "mydrive-pvcs",
						CreationTimestamp: metav1.Now(),
					},
					Status: batchv1.JobStatus{
						Failed: 1,
					},
				}

				BeforeEach(func() {
					// Clean up the outer version of the object
					removeObjFromObjectsList(pvc)

					addObjToObjectsList(pvc)
					addObjToObjectsList(job)
				})

				AfterEach(func() {
					// Clean up the job after test
					removeObjFromObjectsList(pvc)
					removeObjFromObjectsList(job)
				})

				It("Should keep the PVC label as pending", func() {
					Consistently(func() string {
						gotPvc := &corev1.PersistentVolumeClaim{}
						err := cl.Get(ctx, forge.NamespacedNameFromObject(pvc), gotPvc)
						if err != nil {
							return ""
						}
						return gotPvc.Labels[forge.ProvisionJobLabel]
					}, 2*time.Second, 250*time.Millisecond).Should(Equal(forge.ProvisionJobValuePending))
				})
			})
		})

		Context("When PVC is pending", func() {
			// Create a PVC and set it to pending status
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      forge.MyDrivePVCName(tnName),
					Namespace: "mydrive-pvcs",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Status: corev1.PersistentVolumeClaimStatus{
					Phase: corev1.ClaimPending,
				},
			}

			BeforeEach(func() {
				addObjToObjectsList(pvc)
			})

			AfterEach(func() {
				// Clean up the PVC after test
				removeObjFromObjectsList(pvc)
			})

			It("Should not create any PVC Mirror or provisioning job", func() {
				mirror := &corev1.PersistentVolumeClaim{}
				Consistently(func() bool {
					err := cl.Get(ctx, client.ObjectKey{
						Name:      forge.MyDrivePVCMirrorName(tnName),
						Namespace: "tenant-" + tnName,
					}, mirror)
					return err == nil
				}, 2*time.Second, 250*time.Millisecond).Should(BeFalse())

				job := &batchv1.Job{}
				Consistently(func() bool {
					err := cl.Get(ctx, client.ObjectKey{
						Name:      forge.MyDrivePVCName(tnName) + "-provision",
						Namespace: "mydrive-pvcs",
					}, job)
					return err == nil
				}, 2*time.Second, 250*time.Millisecond).Should(BeFalse())
			})
		})
	})

	Context("When tenant namespace does not exist", func() {
		BeforeEach(func() {
			// Set namespace status to not created
			tnResource.Status.PersonalNamespace = clv1alpha2.NameCreated{
				Created: false,
			}
			tnResource.Spec.LastLogin = timePtr(metav1.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
		})

		It("Should not create the PVC", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      forge.MyDrivePVCName(tnName),
				Namespace: "mydrive-pvcs",
			}, pvc, BeFalse(), timeout, interval)
		})
	})

	Context("When tenant is being deleted", func() {
		// Create a PVC to be deleted
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      forge.MyDrivePVCName(tnName),
				Namespace: "mydrive-pvcs",
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test",
				},
			},
			Status: corev1.PersistentVolumeClaimStatus{
				Phase: corev1.ClaimBound,
			},
		}

		BeforeEach(func() {
			// Mark tenant for deletion
			tenantBeingDeleted()

			addObjToObjectsList(pvc)
		})

		AfterEach(func() {
			// Clean up the PVC after test
			removeObjFromObjectsList(pvc)
		})

		It("Should delete the PVC", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			Eventually(func() bool {
				err := cl.Get(ctx, client.ObjectKey{
					Name:      forge.MyDrivePVCName(tnName),
					Namespace: "mydrive-pvcs",
				}, pvc)
				return err == nil
			}, timeout, interval).Should(BeFalse())
		})
	})
})
