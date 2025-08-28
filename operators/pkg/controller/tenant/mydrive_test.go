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

package tenant_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("MyDrive management", func() {
	pvcNs := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mydrive-pvcs",
		},
	}

	personalNs := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-" + tnName,
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector": "test",
			},
		},
	}

	BeforeEach(func() {
		// Set namespace status to created
		tnResource.Status.PersonalNamespace = v1alpha2.NameCreated{
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
			pvc := &v1.PersistentVolumeClaim{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      forge.GetMyDrivePVCName(tnName),
				Namespace: "mydrive-pvcs",
			}, pvc, BeTrue(), timeout, interval)

			Expect(pvc.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(pvc.Spec.AccessModes).To(ContainElement(v1.ReadWriteMany))
			Expect(pvc.Spec.StorageClassName).ToNot(BeNil())
			Expect(*pvc.Spec.StorageClassName).To(Equal("nfs"))

			// Verify resource requests
			Expect(pvc.Spec.Resources.Requests).ToNot(BeEmpty())
			storageRequest := pvc.Spec.Resources.Requests[v1.ResourceStorage]
			expectedStorage := resource.MustParse("5Gi")
			Expect(storageRequest).To(Equal(expectedStorage))
		})

		Context("When PVC is bound", func() {
			// Create a PVC and set it to bound status
			pvc := &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      forge.GetMyDrivePVCName(tnName),
					Namespace: "mydrive-pvcs",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Status: v1.PersistentVolumeClaimStatus{
					Phase: v1.ClaimBound,
				},
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeName: "pv-" + tnName,
				},
			}

			// Create a PV with NFS information
			pv := &v1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pv-" + tnName,
				},
				Spec: v1.PersistentVolumeSpec{
					PersistentVolumeSource: v1.PersistentVolumeSource{
						CSI: &v1.CSIPersistentVolumeSource{
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

			It("Should create the PVC Secret in the tenant namespace", func() {
				secret := &v1.Secret{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{
					Name:      forge.NFSSecretName,
					Namespace: "tenant-" + tnName,
				}, secret, BeTrue(), timeout, interval)

				Expect(secret.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
				Expect(secret.Data).To(HaveKeyWithValue("server-name", []byte("nfs.example.com.test-cluster")))
				Expect(secret.Data).To(HaveKeyWithValue("path", []byte("/exports/user/"+tnName)))
			})

			It("Should create the provisioning job if not already run", func() {
				job := &batchv1.Job{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{
					Name:      forge.GetMyDrivePVCName(tnName) + "-provision",
					Namespace: "mydrive-pvcs",
				}, job, BeTrue(), timeout, interval)

				Expect(job.Spec.Template.Spec.Containers).To(HaveLen(1))

				// Check for the updated PVC label
				pvc := &v1.PersistentVolumeClaim{}
				Expect(cl.Get(ctx, client.ObjectKey{
					Name:      forge.GetMyDrivePVCName(tnName),
					Namespace: "mydrive-pvcs",
				}, pvc)).To(Succeed())

				Expect(pvc.Labels).To(HaveKeyWithValue(forge.ProvisionJobLabel, forge.ProvisionJobValuePending))
			})

			Context("When provisioning job succeeds", func() {
				// Create a PVC with provisioning label already set to pending
				pvc := &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      forge.GetMyDrivePVCName(tnName),
						Namespace: "mydrive-pvcs",
						Labels: map[string]string{
							"crownlabs.polito.it/operator-selector": "test",
							forge.ProvisionJobLabel:                 forge.ProvisionJobValuePending,
						},
					},
					Status: v1.PersistentVolumeClaimStatus{
						Phase: v1.ClaimBound,
					},
					Spec: v1.PersistentVolumeClaimSpec{
						VolumeName: "pv-" + tnName,
					},
				}

				// Create a job that has succeeded
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      forge.GetMyDrivePVCName(tnName) + "-provision",
						Namespace: "mydrive-pvcs",
						Labels: map[string]string{
							forge.ProvisionJobLabel: forge.ProvisionJobValuePending,
						},
						CreationTimestamp: metav1.Now(),
					},
					Status: batchv1.JobStatus{
						Succeeded: 1,
					},
				}

				BeforeEach(func() {
					addObjToObjectsList(pvc)
					addObjToObjectsList(job)
				})

				AfterEach(func() {
					// Clean up the job after test
					removeObjFromObjectsList(pvc)
					removeObjFromObjectsList(job)
				})

				It("Should update the PVC label to ok", func() {
					pvc := &v1.PersistentVolumeClaim{}
					Eventually(func() string {
						err := cl.Get(ctx, client.ObjectKey{
							Name:      forge.GetMyDrivePVCName(tnName),
							Namespace: "mydrive-pvcs",
						}, pvc)
						if err != nil {
							return ""
						}
						return pvc.Labels[forge.ProvisionJobLabel]
					}, timeout, interval).Should(Equal(forge.ProvisionJobValueOk))
				})
			})

			Context("When provisioning job fails", func() {
				// Create a PVC with provisioning label already set to pending
				pvc := &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      forge.GetMyDrivePVCName(tnName),
						Namespace: "mydrive-pvcs",
						Labels: map[string]string{
							"crownlabs.polito.it/operator-selector": "test",
							forge.ProvisionJobLabel:                 forge.ProvisionJobValuePending,
						},
					},
					Status: v1.PersistentVolumeClaimStatus{
						Phase: v1.ClaimBound,
					},
					Spec: v1.PersistentVolumeClaimSpec{
						VolumeName: "pv-" + tnName,
					},
				}

				// Create a job that has failed
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      forge.GetMyDrivePVCName(tnName) + "-provision",
						Namespace: "mydrive-pvcs",
						Labels: map[string]string{
							forge.ProvisionJobLabel: forge.ProvisionJobValuePending,
						},
						CreationTimestamp: metav1.Now(),
					},
					Status: batchv1.JobStatus{
						Failed: 1,
					},
				}

				BeforeEach(func() {
					addObjToObjectsList(pvc)
					addObjToObjectsList(job)
				})

				AfterEach(func() {
					// Clean up the job after test
					removeObjFromObjectsList(pvc)
					removeObjFromObjectsList(job)
				})

				It("Should keep the PVC label as pending", func() {
					pvc := &v1.PersistentVolumeClaim{}
					Consistently(func() string {
						err := cl.Get(ctx, client.ObjectKey{
							Name:      forge.GetMyDrivePVCName(tnName),
							Namespace: "mydrive-pvcs",
						}, pvc)
						if err != nil {
							return ""
						}
						return pvc.Labels[forge.ProvisionJobLabel]
					}, 2*time.Second, 250*time.Millisecond).Should(Equal(forge.ProvisionJobValuePending))
				})
			})
		})

		Context("When PVC is pending", func() {
			// Create a PVC and set it to pending status
			pvc := &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      forge.GetMyDrivePVCName(tnName),
					Namespace: "mydrive-pvcs",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Status: v1.PersistentVolumeClaimStatus{
					Phase: v1.ClaimPending,
				},
			}

			BeforeEach(func() {
				addObjToObjectsList(pvc)
			})

			AfterEach(func() {
				// Clean up the PVC after test
				removeObjFromObjectsList(pvc)
			})

			It("Should not create any PVC Secret or provisioning job", func() {
				secret := &v1.Secret{}
				Consistently(func() bool {
					err := cl.Get(ctx, client.ObjectKey{
						Name:      forge.NFSSecretName,
						Namespace: "tenant-" + tnName,
					}, secret)
					return err == nil
				}, 2*time.Second, 250*time.Millisecond).Should(BeFalse())

				job := &batchv1.Job{}
				Consistently(func() bool {
					err := cl.Get(ctx, client.ObjectKey{
						Name:      forge.GetMyDrivePVCName(tnName) + "-provision",
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
			tnResource.Status.PersonalNamespace = v1alpha2.NameCreated{
				Created: false,
			}
			tnResource.Spec.LastLogin = metav1.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		})

		It("Should not create the PVC", func() {
			pvc := &v1.PersistentVolumeClaim{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      forge.GetMyDrivePVCName(tnName),
				Namespace: "mydrive-pvcs",
			}, pvc, BeFalse(), timeout, interval)
		})
	})

	Context("When tenant is being deleted", func() {
		// Create a PVC to be deleted
		pvc := &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      forge.GetMyDrivePVCName(tnName),
				Namespace: "mydrive-pvcs",
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test",
				},
			},
			Status: v1.PersistentVolumeClaimStatus{
				Phase: v1.ClaimBound,
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
			pvc := &v1.PersistentVolumeClaim{}
			Eventually(func() bool {
				err := cl.Get(ctx, client.ObjectKey{
					Name:      forge.GetMyDrivePVCName(tnName),
					Namespace: "mydrive-pvcs",
				}, pvc)
				return err == nil
			}, timeout, interval).Should(BeFalse())
		})
	})

	Context("Error handling scenarios", func() {
		Context("When PV retrieval fails", func() {
			// Create a PVC and set it to bound status with invalid PV name
			pvc := &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      forge.GetMyDrivePVCName(tnName),
					Namespace: "mydrive-pvcs",
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Status: v1.PersistentVolumeClaimStatus{
					Phase: v1.ClaimBound,
				},
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeName: "nonexistent-pv",
				},
			}

			BeforeEach(func() {
				addObjToObjectsList(pvc)

				// The reconciliation will fail because the PV does not exist
				tnReconcileErrExpected = HaveOccurred()
			})

			AfterEach(func() {
				// Clean up the PVC after test
				removeObjFromObjectsList(pvc)
			})

			It("Should return an error and not create the PVC Secret", func() {
				secret := &v1.Secret{}
				Consistently(func() bool {
					err := cl.Get(ctx, client.ObjectKey{
						Name:      forge.NFSSecretName,
						Namespace: "tenant-" + tnName,
					}, secret)
					return err == nil
				}, 2*time.Second, 250*time.Millisecond).Should(BeFalse())
			})
		})
	})
})
