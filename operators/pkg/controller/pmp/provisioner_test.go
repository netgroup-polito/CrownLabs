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

package pmp_test

import (
	"context"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v13/controller"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/pmp"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// The following are unit test aimed to check the PVC Mirror Provisioner functions.
var _ = Describe("The PVC Mirror Provisioner methods", func() {
	AddObject := func(obj client.Object) {
		err := pmprov.Client.Create(ctx, obj)
		if apierrors.IsAlreadyExists(err) {
			Expect(pmprov.Client.Update(ctx, obj)).To(Succeed())
		} else {
			Expect(err).To(BeNil())
		}
	}

	UpdateObject := func(obj client.Object) {
		Expect(pmprov.Client.Update(ctx, obj)).To(Succeed())
	}
	UpdateObjectStatus := func(obj client.Object) {
		Expect(pmprov.Client.Status().Update(ctx, obj)).To(Succeed())
	}

	RemoveObject := func(obj client.Object) {
		_ = pmprov.Client.Delete(ctx, obj)
	}

	RemovePVC := func(pvc *corev1.PersistentVolumeClaim) {
		ctrlUtil.RemoveFinalizer(pvc, "kubernetes.io/pvc-protection")
		Expect(pmprov.Client.Update(ctx, pvc.DeepCopy())).To(Succeed())
		RemoveObject(pvc)
	}

	RemovePV := func(pv *corev1.PersistentVolume) {
		ctrlUtil.RemoveFinalizer(pv, "kubernetes.io/pv-protection")
		Expect(pmprov.Client.Update(ctx, pv.DeepCopy())).To(Succeed())
		RemoveObject(pv)
	}

	var _ = Describe("The PMP Provision method", Ordered, func() {
		var (
			nsOrigin         corev1.Namespace
			nsTarget         corev1.Namespace
			sc               storagev1.StorageClass
			pvcOrig          corev1.PersistentVolumeClaim
			pvOrig           corev1.PersistentVolume
			pvcMirr          corev1.PersistentVolumeClaim
			pvMirr           *corev1.PersistentVolume
			provisionOptions controller.ProvisionOptions
			provisioningErr  error

			scName          = "pmp-mirror"
			pvcAnnotations  map[string]string
			nsTgtLabels     map[string]string
			nsOrigLabels    map[string]string
			pvSource        corev1.PersistentVolumeSource
			resRequirements corev1.VolumeResourceRequirements
			origVolumeName  string
			pvcOrigPhase    corev1.PersistentVolumeClaimPhase
			mirrDataSrcRef  *corev1.TypedObjectReference

			isPVCOrigCreated bool
			isPVOrigCreated  bool
		)

		const (
			targetLabelKey = "crownlabs.polito.it/operator-selector"
			targetLabelVal = "test"
			provisioner    = "pmp.crownlabs.polito.it"
			pvOrigName     = "pv-origin"
			pvMirrName     = "pv-mirror"
			pvcOrigName    = "pvc-origin"
			pvcMirrName    = "pvc-mirror"
			originNsName   = "origin-ns"
			targetNsName   = "target-ns"
			tenantName     = "s123456"
		)

		isIgnoredError := func(err error) bool {
			// Ignored errors
			var ignoredErr *controller.IgnoredError
			return errors.As(err, &ignoredErr)
		}

		isStatusError := func(err error) bool {
			// Infeasible errors: gRPC status errors with codes
			status, ok := status.FromError(err)
			return ok && status.Code() == codes.InvalidArgument
		}

		isOtherError := func(err error) bool {
			// Standard errors: all others
			return !isIgnoredError(err) && !isStatusError(err)
		}

		ExpectIgnoredError := func() {
			Expect(provisioningErr).ToNot(BeNil())
			Expect(isIgnoredError(provisioningErr)).To(BeTrue())
		}

		ExpectInfeasibleError := func() {
			Expect(provisioningErr).ToNot(BeNil())
			Expect(isStatusError(provisioningErr)).To(BeTrue())
		}

		ExpectOtherError := func() {
			Expect(provisioningErr).ToNot(BeNil())
			Expect(isOtherError(provisioningErr)).To(BeTrue())
		}

		ExpectNoError := func() {
			Expect(provisioningErr).To(BeNil())
		}

		BeforeAll(func() {
			pmprov.TargetLabel = common.NewLabel(targetLabelKey, targetLabelVal)

			nsOrigin = corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: originNsName,
				},
			}
			nsTarget = corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: targetNsName,
				},
			}
			sc = storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: scName,
				},
				Provisioner:          provisioner,
				AllowVolumeExpansion: ptr.To(false),
				ReclaimPolicy:        ptr.To(corev1.PersistentVolumeReclaimDelete),
			}
			resRequirements = corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewScaledQuantity(1, resource.Giga),
				},
			}

			isPVCOrigCreated = false
			isPVOrigCreated = false

			AddObject(&nsOrigin)
			AddObject(&nsTarget)
			AddObject(&sc)
		})

		JustBeforeEach(func() {
			nsOrigin.Labels = nsOrigLabels
			nsTarget.Labels = nsTgtLabels
			UpdateObject(&nsOrigin)
			UpdateObject(&nsTarget)

			pvOrig = corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: pvOrigName,
				},
				Spec: corev1.PersistentVolumeSpec{
					Capacity: resRequirements.Requests,
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteMany,
					},
					PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
					PersistentVolumeSource:        pvSource,
					StorageClassName:              scName,
				},
			}
			pvcOrig = corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pvcOrigName,
					Namespace:   originNsName,
					Annotations: pvcAnnotations,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteMany,
					},
					StorageClassName: &scName,
					Resources:        resRequirements,
					VolumeName:       origVolumeName,
				},
			}
			pvcMirr = corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcMirrName,
					Namespace: targetNsName,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteMany,
					},
					DataSourceRef:    mirrDataSrcRef,
					StorageClassName: &scName,
					Resources:        resRequirements,
				},
			}
			provisionOptions = controller.ProvisionOptions{
				StorageClass: &sc,
				PVC:          &pvcMirr,
				PVName:       pvMirrName,
			}

			if isPVOrigCreated {
				AddObject(&pvOrig)
			}
			if isPVCOrigCreated {
				AddObject(&pvcOrig)

				pvcOrig.Status.Phase = pvcOrigPhase
				UpdateObjectStatus(&pvcOrig)
			}
			AddObject(&pvcMirr)

			// Call function under test
			pvMirr, _, provisioningErr = pmprov.Provision(ctx, provisionOptions)
		})

		JustAfterEach(func() {
			if isPVOrigCreated {
				RemovePV(&pvOrig)
			}
			if isPVCOrigCreated {
				RemovePVC(&pvcOrig)
			}
			RemovePVC(&pvcMirr)
		})

		Context("TargetLabel is correctly set on both target namespace and PMP", func() {
			BeforeEach(func() {
				nsTgtLabels = map[string]string{
					targetLabelKey: targetLabelVal,
				}
				pmprov.TargetLabel = common.NewLabel(targetLabelKey, targetLabelVal)
			})

			When("DataSourceRef is present", func() {
				When("DataSourceRef is a PersistentVolumeClaim", func() {
					BeforeEach(func() {
						mirrDataSrcRef = &corev1.TypedObjectReference{
							APIGroup:  nil,
							Kind:      "PersistentVolumeClaim",
							Name:      pvcOrigName,
							Namespace: ptr.To(originNsName),
						}
					})

					Context("the origin PVC exists", func() {
						BeforeEach(func() {
							isPVCOrigCreated = true
						})

						Context("Namespace is authorized (shvol annotation)", func() {
							BeforeEach(func() {
								pvcAnnotations = map[string]string{
									pmp.AuthorizationAnnotationKey: forge.ShVolAuthorizationAnnotationValue,
								}
								nsTgtLabels = map[string]string{
									targetLabelKey:     targetLabelVal,
									forge.LabelTypeKey: "tenant",
								}
							})

							When("the origin PVC is Bound", func() {
								BeforeEach(func() {
									isPVOrigCreated = true
									origVolumeName = pvOrigName
									pvcOrigPhase = corev1.ClaimBound
								})

								When("the origin PV has CSI spec", func() {
									BeforeEach(func() {
										pvSource = corev1.PersistentVolumeSource{
											CSI: &corev1.CSIPersistentVolumeSource{
												Driver:       "nfs.example.com",
												VolumeHandle: "/nfs/path",
												VolumeAttributes: map[string]string{
													"server": "my-nfs-cluster",
												},
											},
										}
									})

									It("should successfully provision the mirror PV", func() {
										ExpectNoError()
										Expect(pvMirr).ToNot(BeNil())
										Expect(pvMirr.Spec.CSI).ToNot(BeNil())
										Expect(pvMirr.Spec.CSI.VolumeHandle).To(Equal("/nfs/path"))
									})
								})

								When("the origin PV does not have CSI spec", func() {
									BeforeEach(func() {
										pvSource = corev1.PersistentVolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/tmp/pv",
											},
										}
									})

									It("should return an ignored error", func() {
										ExpectIgnoredError()
									})
								})
							})

							When("the origin PVC is Pending", func() {
								BeforeEach(func() {
									isPVOrigCreated = false
									pvcOrigPhase = corev1.ClaimPending
								})

								It("should return an error", func() {
									ExpectOtherError()
								})
							})

							When("the origin PVC is Lost", func() {
								BeforeEach(func() {
									isPVOrigCreated = false
									pvcOrigPhase = corev1.ClaimLost
								})

								It("should return an ignored error", func() {
									ExpectIgnoredError()
								})
							})
						})

						Context("Namespace is authorized (mydrive annotation)", func() {
							BeforeEach(func() {
								pvcAnnotations = map[string]string{
									pmp.AuthorizationAnnotationKey: strings.ReplaceAll(forge.MyDriveAuthorizationAnnotationValue, "{tenant-id}", tenantName),
								}
								nsTgtLabels = map[string]string{
									targetLabelKey:             targetLabelVal,
									forge.LabelTypeKey:         "tenant",
									"crownlabs.polito.it/name": tenantName,
								}
							})

							When("the origin PVC is Bound", func() {
								BeforeEach(func() {
									isPVOrigCreated = true
									origVolumeName = pvOrigName
									pvcOrigPhase = corev1.ClaimBound
								})

								When("the origin PV has CSI spec", func() {
									BeforeEach(func() {
										pvSource = corev1.PersistentVolumeSource{
											CSI: &corev1.CSIPersistentVolumeSource{
												Driver:       "nfs.example.com",
												VolumeHandle: "/nfs/path",
												VolumeAttributes: map[string]string{
													"server": "my-nfs-cluster",
												},
											},
										}
									})

									It("should successfully provision the mirror PV", func() {
										ExpectNoError()
										Expect(pvMirr).ToNot(BeNil())
										Expect(pvMirr.Labels).To(HaveLen(3))
										Expect(pvMirr.Labels).To(HaveKeyWithValue(pmp.MirroredPvLabel, pvOrigName))
										Expect(pvMirr.Labels).To(HaveKeyWithValue(pmp.MirroredPvcNamespaceLabel, originNsName))
										Expect(pvMirr.Labels).To(HaveKeyWithValue(pmp.MirroredPvcNameLabel, pvcOrigName))
										Expect(pvMirr.Spec.CSI).ToNot(BeNil())
										Expect(pvMirr.Spec.CSI.VolumeHandle).To(Equal("/nfs/path"))
									})
								})

								When("the origin PV does not have CSI spec", func() {
									BeforeEach(func() {
										pvSource = corev1.PersistentVolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/tmp/pv",
											},
										}
									})

									It("should return an ignored error", func() {
										ExpectIgnoredError()
									})
								})
							})

							When("the origin PVC is Pending", func() {
								BeforeEach(func() {
									isPVOrigCreated = false
									pvcOrigPhase = corev1.ClaimPending
								})

								It("should return an error", func() {
									ExpectOtherError()
								})
							})

							When("the origin PVC is Lost", func() {
								BeforeEach(func() {
									isPVOrigCreated = false
									pvcOrigPhase = corev1.ClaimLost
								})

								It("should return an ignored error", func() {
									ExpectIgnoredError()
								})
							})
						})

						When("authorization annotation is missing on origin PVC", func() {
							BeforeEach(func() {
								pvcAnnotations = map[string]string{}
								nsTgtLabels = map[string]string{
									targetLabelKey:     targetLabelVal,
									forge.LabelTypeKey: "tenant",
								}
							})

							It("should return an unauthorized error", func() {
								ExpectInfeasibleError()
							})
						})

						When("target namespace does not have required label", func() {
							BeforeEach(func() {
								pvcAnnotations = map[string]string{
									pmp.AuthorizationAnnotationKey: forge.ShVolAuthorizationAnnotationValue,
								}
								nsTgtLabels = map[string]string{
									targetLabelKey: targetLabelVal,
								}
							})

							It("should return an unauthorized error", func() {
								ExpectInfeasibleError()
							})
						})

						When("annotation-label mismatch", func() {
							BeforeEach(func() {
								pvcAnnotations = map[string]string{
									pmp.AuthorizationAnnotationKey: forge.ShVolAuthorizationAnnotationValue,
								}
								nsTgtLabels = map[string]string{
									targetLabelKey:     targetLabelVal,
									forge.LabelTypeKey: "not-tenant",
								}
							})

							It("should return an unauthorized error", func() {
								ExpectInfeasibleError()
							})
						})
					})

					Context("the origin PVC does not exist", func() {
						BeforeEach(func() {
							isPVCOrigCreated = false
							isPVOrigCreated = false
						})

						It("should return an error to slowly retry", func() {
							ExpectInfeasibleError()
						})
					})
				})

				When("DataSourceRef is not a PersistentVolumeClaim", func() {
					BeforeEach(func() {
						mirrDataSrcRef = &corev1.TypedObjectReference{
							APIGroup:  ptr.To("snapshot.storage.k8s.io"),
							Kind:      "VolumeSnapshot",
							Name:      "snap",
							Namespace: ptr.To(originNsName),
						}
					})

					It("should return an ignored error", func() {
						ExpectIgnoredError()
					})
				})
			})

			When("DataSourceRef is not present", func() {
				BeforeEach(func() {
					mirrDataSrcRef = nil
				})

				It("should return an ignored error", func() {
					ExpectIgnoredError()
				})
			})
		})

		Context("TargetLabel is not correctly set on the target namespace", func() {
			BeforeEach(func() {
				nsTgtLabels = map[string]string{
					targetLabelKey: "not-" + targetLabelVal,
				}
			})

			It("should return an ignored error", func() {
				ExpectIgnoredError()
			})
		})

		Context("TargetLabel is missing", func() {
			BeforeEach(func() {
				nsTgtLabels = map[string]string{}

				mirrDataSrcRef = &corev1.TypedObjectReference{
					APIGroup:  nil,
					Kind:      "PersistentVolumeClaim",
					Name:      pvcOrigName,
					Namespace: ptr.To(originNsName),
				}
			})

			It("should return an ignored error", func() {
				ExpectIgnoredError()
			})
		})

		AfterAll(func() {
			RemoveObject(&sc)
			RemoveObject(&nsTarget)
			RemoveObject(&nsOrigin)
		})
	})

	var _ = Describe("The PMP Delete method", func() {
		var err error

		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}
		pvMirr := corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pvc-123456",
			},
		}

		JustBeforeEach(func() {
			AddObject(&ns)
			err = pmprov.Delete(ctx, &pvMirr)
		})

		It("should never return an error", func() {
			Expect(err).To(BeNil())
		})

		JustAfterEach(func() {
			RemoveObject(&ns)
		})
	})

	var _ = Describe("The PMP ShouldDelete method", func() {
		var result bool

		pvMirr := corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pvc-123456",
			},
		}

		JustBeforeEach(func() {
			result = pmprov.ShouldDelete(ctx, &pvMirr)
		})

		It("should always return true", func() {
			Expect(result).To(BeTrue())
		})
	})

	var _ = Describe("The PMP Start method", func() {
		var (
			err    error
			innCtx context.Context
			cancel context.CancelFunc
			sc     storagev1.StorageClass
		)

		JustBeforeEach(func() {
			innCtx, cancel = context.WithCancel(ctx)
			err = pmprov.Start(innCtx)
		})

		AfterEach(func() {
			cancel()
		})

		Context("The configured StorageClass exists", func() {
			When("the configuration is correct", func() {
				BeforeEach(func() {
					sc = storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: pmprov.MirrorStorageClass,
						},
						Provisioner:          pmprov.MirrorProvisionerName,
						AllowVolumeExpansion: ptr.To(false),
					}
					AddObject(&sc)
				})

				// It("Should not return")
				// Impossible to create a test for this
			})

			When("the SC.Provisioner does not match the one in the PMP", func() {
				BeforeEach(func() {
					sc = storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: pmprov.MirrorStorageClass,
						},
						Provisioner:          pmprov.MirrorProvisionerName + "/mismatch",
						AllowVolumeExpansion: ptr.To(false),
					}
					AddObject(&sc)
				})

				It("Should return an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("the SC.VolumeExpansion is set to true", func() {
				BeforeEach(func() {
					sc = storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: pmprov.MirrorStorageClass,
						},
						Provisioner:          pmprov.MirrorProvisionerName,
						AllowVolumeExpansion: ptr.To(true),
					}
					AddObject(&sc)
				})

				It("Should return an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			AfterEach(func() {
				RemoveObject(&sc)
			})
		})

		Context("The configured StorageClass does not exist", func() {
			It("Should return an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
	})
})
