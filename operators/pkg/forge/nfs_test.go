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

package forge_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("External Volume Mounts and Provisioning Job forging", func() {
	var (
		mountInfo clv1alpha2.SharedVolumeMountInfo
	)

	const (
		tenantName   = "tester"
		instanceName = "kubernetes-0000"

		shVolName      = "myshvol"
		shVolNamespace = "default"
		shVolMountPath = "/mnt/path"
		shVolReadOnly  = true
	)

	JustBeforeEach(func() {
		mountInfo = clv1alpha2.SharedVolumeMountInfo{
			SharedVolumeRef: clv1alpha2.GenericRef{
				Name:      shVolName,
				Namespace: shVolNamespace,
			},
			MountPath: shVolMountPath,
			ReadOnly:  shVolReadOnly,
		}
	})

	Describe("The forge.MyDriveMountInfo and .GetMyDrivePVCMirrorName functions", func() {
		It("Should return the correct VolumeMount object", func() {
			expected := corev1.VolumeMount{
				Name:      fmt.Sprintf("%s-drive-mirror", tenantName),
				MountPath: forge.MyDriveVolumeMountPath,
				ReadOnly:  false,
			}
			Expect(forge.MyDriveMountInfo(tenantName)).To(Equal(expected))
		})
	})

	Describe("The forge.ShVolMountInfo and .GetShVolPVCMirrorName functions", func() {
		When("Instance and SharedVolume names are short (<= 122 chars)", func() {
			It("Should return the correct VolumeMount object if names are short", func() {
				expected := corev1.VolumeMount{
					Name:      fmt.Sprintf("%s-%s-mirror", mountInfo.SharedVolumeRef.Name, instanceName),
					MountPath: mountInfo.MountPath,
					ReadOnly:  mountInfo.ReadOnly,
				}
				Expect(forge.ShVolMountInfo(mountInfo, instanceName)).To(Equal(expected))
			})
		})

		When("Instance and SharedVolume names are long (> 122 chars)", func() {
			instanceNameLong := strings.Repeat("b", 150)
			shVolNameLong := strings.Repeat("a", 150)

			JustBeforeEach(func() {
				mountInfo.SharedVolumeRef.Name = shVolNameLong
			})

			It("Should return the correctly trimmed VolumeMount object if names are long", func() {
				expected := corev1.VolumeMount{
					Name:      fmt.Sprintf("%s-%s-mirror", strings.Repeat("a", 122), strings.Repeat("b", 122)),
					MountPath: mountInfo.MountPath,
					ReadOnly:  mountInfo.ReadOnly,
				}
				Expect(forge.ShVolMountInfo(mountInfo, instanceNameLong)).To(Equal(expected))
			})
		})

	})

	Describe("The forge.PVCProvisioningJobSpec function", func() {
		var pvc corev1.PersistentVolumeClaim

		JustBeforeEach(func() {
			pvc.Name = "shvol-0000"
		})

		It("Should have the right spec", func() {
			actual := forge.PVCProvisioningJobSpec(&pvc)
			Expect(actual.Template.Spec.Containers[0].Image).To(Equal("busybox"))
			Expect(actual.Template.Spec.Containers[0].Command).To(ContainElement("chown"))
			Expect(actual.Template.Spec.Volumes[0].VolumeSource.PersistentVolumeClaim.ClaimName).To(Equal("shvol-0000"))
		})
	})

	Context("Functions that interact with k8s", func() {
		var (
			ctx           context.Context
			clientBuilder fake.ClientBuilder

			template    clv1alpha2.Template
			environment clv1alpha2.Environment
			instance    clv1alpha2.Instance
			tenant      clv1alpha2.Tenant
			manager     clv1alpha2.Tenant

			err error
		)

		const (
			instanceName      = "kubernetes-0000"
			instanceNamespace = "tenant-tester"
			templateName      = "kubernetes"
			templateNamespace = "workspace-netgroup"
			workspaceName     = "netgroup"
			environmentName   = "control-plane"
			tenantName        = "tester"
			tenantMgrName     = "manager"
		)

		Expect(clv1alpha2.AddToScheme(scheme.Scheme)).To(Succeed())

		BeforeEach(func() {
			ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
			clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
				Spec: clv1alpha2.InstanceSpec{
					Running:  true,
					Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
					Tenant:   clv1alpha2.GenericRef{Name: tenantName},
				},
			}
			template = clv1alpha2.Template{
				ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateNamespace},
				Spec: clv1alpha2.TemplateSpec{
					WorkspaceRef: clv1alpha2.GenericRef{Name: workspaceName},
				},
			}
			environment = clv1alpha2.Environment{Name: environmentName, MountMyDriveVolume: true}
			tenant = clv1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: tenantName,
					Labels: map[string]string{
						clv1alpha2.WorkspaceLabelPrefix + workspaceName: string(clv1alpha2.User),
					},
				},
			}
			manager = clv1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: tenantMgrName,
					Labels: map[string]string{
						clv1alpha2.WorkspaceLabelPrefix + workspaceName: string(clv1alpha2.Manager),
					},
				},
			}

			ctx, _ = clctx.InstanceInto(ctx, &instance)
			ctx, _ = clctx.TemplateInto(ctx, &template)
			ctx, _ = clctx.TenantInto(ctx, &tenant)
			ctx, _ = clctx.EnvironmentInto(ctx, &environment)
		})

		Describe("The forge.PVCMountInfosFromEnvironment function", func() {
			var (
				shvol      clv1alpha2.SharedVolume
				shVolRef   clv1alpha2.GenericRef
				mountInfos []corev1.VolumeMount
			)

			BeforeEach(func() {
				shvol = clv1alpha2.SharedVolume{
					ObjectMeta: metav1.ObjectMeta{
						Name:      templateName + "-shvol",
						Namespace: templateNamespace,
					},
					Status: clv1alpha2.SharedVolumeStatus{
						Phase: clv1alpha2.SharedVolumePhaseReady,
					},
				}
				shVolRef = clv1alpha2.GenericRef{Name: templateName + "-shvol", Namespace: templateNamespace}

				clientBuilder = *clientBuilder.WithObjects(&shvol, &tenant, &manager)
			})

			JustBeforeEach(func() {
				client := clientBuilder.Build()
				mountInfos, _, err = forge.PVCMountInfosFromEnvironment(ctx, client)
			})

			When("The environment does not require external volumes", func() {
				BeforeEach(func() {
					environment = clv1alpha2.Environment{Name: environmentName, MountMyDriveVolume: false}
					ctx, _ = clctx.EnvironmentInto(ctx, &environment)
				})

				It("Should correctly generate the mountInfos", func() {
					Expect(err).ToNot(HaveOccurred())

					Expect(mountInfos).To(HaveLen(0))
				})
			})

			When("The environment requires the personal volume", func() {
				BeforeEach(func() {
					environment = clv1alpha2.Environment{Name: environmentName, MountMyDriveVolume: true}
					ctx, _ = clctx.EnvironmentInto(ctx, &environment)
				})

				It("Should correctly generate the mountInfos", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(mountInfos).To(HaveLen(1))

					mountInfo := mountInfos[0]
					expected := corev1.VolumeMount{
						Name:      forge.GetMyDrivePVCMirrorName(tenantName),
						MountPath: forge.MyDriveVolumeMountPath,
						ReadOnly:  false,
					}

					Expect(mountInfo).To(Equal(expected))
				})
			})

			When("The environment requires some read-write shared volumes", func() {
				BeforeEach(func() {
					environment = clv1alpha2.Environment{
						Name:               environmentName,
						MountMyDriveVolume: false,
						SharedVolumeMounts: []clv1alpha2.SharedVolumeMountInfo{
							{
								SharedVolumeRef: shVolRef,
								MountPath:       shVolMountPath,
								ReadOnly:        false,
							},
						},
					}
					ctx, _ = clctx.EnvironmentInto(ctx, &environment)
				})

				It("Should correctly generate the mountInfos", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(mountInfos).To(HaveLen(1))

					mountInfo := mountInfos[0]
					expected := corev1.VolumeMount{
						Name:      forge.GetShVolPVCMirrorName(shvol.Name, instanceName),
						MountPath: shVolMountPath,
						ReadOnly:  false,
					}

					Expect(mountInfo).To(Equal(expected))
				})
			})

			Context("The environment requires some read-only shared volumes", func() {
				BeforeEach(func() {
					environment = clv1alpha2.Environment{
						Name:               environmentName,
						MountMyDriveVolume: false,
						SharedVolumeMounts: []clv1alpha2.SharedVolumeMountInfo{
							{
								SharedVolumeRef: shVolRef,
								MountPath:       shVolMountPath,
								ReadOnly:        true,
							},
						},
					}
					ctx, _ = clctx.EnvironmentInto(ctx, &environment)
				})

				When("The current tenant is a manager of the workspace", func() {
					BeforeEach(func() {
						ctx, _ = clctx.TenantInto(ctx, &manager)
					})

					It("Should mount the shvol as read-write", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(mountInfos).To(HaveLen(1))

						mountInfo := mountInfos[0]
						expected := corev1.VolumeMount{
							Name:      forge.GetShVolPVCMirrorName(shvol.Name, instanceName),
							MountPath: shVolMountPath,
							ReadOnly:  false,
						}

						Expect(mountInfo).To(Equal(expected))
					})
				})

				When("The current tenant is not a manager of the workspace", func() {
					It("Should mount the shvol as read-only", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(mountInfos).To(HaveLen(1))

						mountInfo := mountInfos[0]
						expected := corev1.VolumeMount{
							Name:      forge.GetShVolPVCMirrorName(shvol.Name, instanceName),
							MountPath: shVolMountPath,
							ReadOnly:  true,
						}

						Expect(mountInfo).To(Equal(expected))
					})
				})
			})

			When("The environment requires both personal and shared volumes", func() {
				BeforeEach(func() {
					environment = clv1alpha2.Environment{
						Name:               environmentName,
						MountMyDriveVolume: true,
						SharedVolumeMounts: []clv1alpha2.SharedVolumeMountInfo{
							{
								SharedVolumeRef: shVolRef,
								MountPath:       shVolMountPath,
								ReadOnly:        false,
							},
						},
					}
					ctx, _ = clctx.EnvironmentInto(ctx, &environment)
				})

				It("Should correctly generate the mountInfos", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(mountInfos).To(HaveLen(2))

					expected := []corev1.VolumeMount{
						{
							Name:      forge.GetMyDrivePVCMirrorName(tenantName),
							MountPath: forge.MyDriveVolumeMountPath,
							ReadOnly:  false,
						}, {
							Name:      forge.GetShVolPVCMirrorName(shvol.Name, instanceName),
							MountPath: shVolMountPath,
							ReadOnly:  false,
						},
					}

					Expect(mountInfos).To(Equal(expected))
				})
			})

		})
	})

	Describe("GetMyDrivePVCName", func() {
		It("Should correctly format PVC name for tenant name without dots", func() {
			tenantName := "student"

			pvcName := forge.GetMyDrivePVCName(tenantName)

			Expect(pvcName).To(Equal("student-drive"))
		})

		It("Should correctly format PVC name for tenant name with dots", func() {
			tenantName := "s123456.student"

			pvcName := forge.GetMyDrivePVCName(tenantName)

			Expect(pvcName).To(Equal("s123456-student-drive"))
		})
	})

	Describe("ConfigureMyDrivePVC", func() {
		It("Should initialize labels if nil and set PVC properties", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			storageClassName := "example-storage-class"
			storageSize := resource.MustParse("10Gi")
			labels := map[string]string{
				"custom-label": "custom-value",
			}
			annotations := map[string]string{
				"custom-annotation": "custom-value",
			}

			forge.ConfigureMyDrivePVC(pvc, storageClassName, storageSize, labels, annotations)

			Expect(pvc.Labels).ToNot(BeNil())
			Expect(pvc.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(pvc.Annotations).ToNot(BeNil())
			Expect(pvc.Annotations).To(HaveKeyWithValue("custom-annotation", "custom-value"))
			Expect(pvc.Spec.AccessModes).To(ConsistOf(corev1.ReadWriteMany))
			Expect(pvc.Spec.StorageClassName).ToNot(BeNil())
			Expect(*pvc.Spec.StorageClassName).To(Equal(storageClassName))
			Expect(pvc.Spec.Resources.Requests).ToNot(BeNil())
			Expect(pvc.Spec.Resources.Requests.Storage().String()).To(Equal("10Gi"))
		})

		It("Should preserve existing labels and update storage size if larger", func() {
			existingSize := resource.MustParse("5Gi")
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: existingSize,
						},
					},
				},
			}

			storageClassName := "example-storage-class"
			storageSize := resource.MustParse("10Gi")
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureMyDrivePVC(pvc, storageClassName, storageSize, labels, nil)

			Expect(pvc.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(pvc.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(pvc.Spec.Resources.Requests).ToNot(BeNil())
			Expect(pvc.Spec.Resources.Requests.Storage().String()).To(Equal("10Gi"))
		})

		It("Should not update storage size if existing size is larger", func() {
			existingSize := resource.MustParse("20Gi")
			pvc := &corev1.PersistentVolumeClaim{
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: existingSize,
						},
					},
				},
			}

			storageClassName := "example-storage-class"
			storageSize := resource.MustParse("10Gi")
			labels := map[string]string{}

			forge.ConfigureMyDrivePVC(pvc, storageClassName, storageSize, labels, nil)

			Expect(pvc.Spec.Resources.Requests).ToNot(BeNil())
			Expect(pvc.Spec.Resources.Requests.Storage().String()).To(Equal("20Gi"))
		})
	})

	Describe("UpdatePVCProvisioningJobLabel", func() {
		It("Should initialize labels if nil and set provisioning job label", func() {
			pvc := &corev1.PersistentVolumeClaim{}

			forge.UpdatePVCProvisioningJobLabel(pvc, "job-123")

			Expect(pvc.Labels).ToNot(BeNil())
			Expect(pvc.Labels).To(HaveKeyWithValue("crownlabs.polito.it/volume-provisioning", "job-123"))
		})

		It("Should update provisioning job label on existing labels", func() {
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"crownlabs.polito.it/volume-provisioning": "old-job",
						"existing-label": "existing-value",
					},
				},
			}

			forge.UpdatePVCProvisioningJobLabel(pvc, "job-456")

			Expect(pvc.Labels).ToNot(BeNil())
			Expect(pvc.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(pvc.Labels).To(HaveKeyWithValue("crownlabs.polito.it/volume-provisioning", "job-456"))
		})
	})

	Describe("ConfigureMyDriveProvisioningJob", func() {
		It("Should configure the job spec using PVCProvisioningJobSpec", func() {
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: "test-namespace",
				},
			}
			job := &batchv1.Job{}

			forge.ConfigureMyDriveProvisioningJob(job, pvc)

			// Verify that the spec is correctly configured
			Expect(job.Spec.Template.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyOnFailure))
			Expect(job.Spec.BackoffLimit).ToNot(BeNil())
			Expect(*job.Spec.BackoffLimit).To(Equal(int32(forge.ProvisionJobMaxRetries)))
			Expect(job.Spec.TTLSecondsAfterFinished).ToNot(BeNil())
			Expect(*job.Spec.TTLSecondsAfterFinished).To(Equal(int32(forge.ProvisionJobTTLSeconds)))

			// Verify that there is a container
			Expect(job.Spec.Template.Spec.Containers).To(HaveLen(1))
		})
	})
})
