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

package forge_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("NFS Mounts and Provisioning Job forging", func() {
	var (
		shvol     clv1alpha2.SharedVolume
		mountInfo clv1alpha2.SharedVolumeMountInfo
	)

	const (
		nfsServerAddress = "nfs.example.com"
		nfsExportPath    = "/nfs/path"
		nfsMountPath     = "/mnt/path"
		nfsReadOnly      = true
	)

	JustBeforeEach(func() {
		shvol = clv1alpha2.SharedVolume{
			Spec: clv1alpha2.SharedVolumeSpec{
				PrettyName: "My Shared Volume",
				Size:       resource.Quantity{},
			},
			Status: clv1alpha2.SharedVolumeStatus{
				ServerAddress: nfsServerAddress,
				ExportPath:    nfsExportPath,
				Phase:         clv1alpha2.SharedVolumePhaseReady,
			},
		}
		mountInfo = clv1alpha2.SharedVolumeMountInfo{
			SharedVolumeRef: clv1alpha2.GenericRef{
				Name:      "myshvol",
				Namespace: "default",
			},
			MountPath: nfsMountPath,
			ReadOnly:  nfsReadOnly,
		}
	})

	Describe("The forge.NFSVolumeMount function", func() {
		type NFSVolumeMountCase struct {
			ServerAddress  string
			ExportPath     string
			MountPath      string
			ReadOnly       bool
			ExpectedOutput []string
		}

		WhenBody := func(c NFSVolumeMountCase) func() {
			return func() {
				It("Should return the correct mount string array", func() {
					Expect(forge.NFSVolumeMount(c.ServerAddress, c.ExportPath, c.MountPath, c.ReadOnly)).To(Equal(c.ExpectedOutput))
				})
			}
		}

		When("the volume is read only", WhenBody(NFSVolumeMountCase{
			ServerAddress: nfsServerAddress,
			ExportPath:    nfsExportPath,
			MountPath:     nfsMountPath,
			ReadOnly:      true,
			ExpectedOutput: []string{
				"nfs.example.com:/nfs/path",
				"/mnt/path",
				"nfs",
				"ro,tcp,hard,intr,rsize=8192,wsize=8192,timeo=14,_netdev,user",
				"0",
				"0",
			},
		}))

		When("the volume is not read only", WhenBody(NFSVolumeMountCase{
			ServerAddress: nfsServerAddress,
			ExportPath:    nfsExportPath,
			MountPath:     nfsMountPath,
			ReadOnly:      false,
			ExpectedOutput: []string{
				"nfs.example.com:/nfs/path",
				"/mnt/path",
				"nfs",
				"rw,tcp,hard,intr,rsize=8192,wsize=8192,timeo=14,_netdev,user",
				"0",
				"0",
			},
		}))
	})

	Describe("The forge.MyDriveVolumeMount function", func() {
		It("Should return the correct mount string array", func() {
			expected := []string{
				"nfs.example.com:/nfs/path",
				"/media/mydrive",
				"nfs",
				"rw,tcp,hard,intr,rsize=8192,wsize=8192,timeo=14,_netdev,user",
				"0",
				"0",
			}
			Expect(forge.MyDriveVolumeMount(nfsServerAddress, nfsExportPath)).To(Equal(expected))
		})
	})

	Describe("The forge.SharedVolumeMount function", func() {
		When("the shared volume is ready", func() {
			It("Should return the correct mount string array", func() {
				expected := []string{
					"nfs.example.com:/nfs/path",
					"/mnt/path",
					"nfs",
					"ro,tcp,hard,intr,rsize=8192,wsize=8192,timeo=14,_netdev,user",
					"0",
					"0",
				}
				Expect(forge.SharedVolumeMount(&shvol, mountInfo)).To(Equal(expected))
			})
		})

		When("the shared volume is not ready or has invalid server/path", func() {
			It("Should return the correct mount string array", func() {
				shvol.Status.ServerAddress = ""
				expected := []string{
					"# Here lies an invalid SharedVolume mount",
					"",
					"",
					"",
					"",
					"",
				}
				Expect(forge.SharedVolumeMount(&shvol, mountInfo)).To(Equal(expected))
			})

			AfterEach(func() {
				shvol.Status.ServerAddress = nfsServerAddress
			})
		})
	})

	Describe("The forge.CommentMount function", func() {
		It("Should return the correct mount string array", func() {
			expected := []string{
				"# This is a test comment",
				"",
				"",
				"",
				"",
				"",
			}
			Expect(forge.CommentMount("This is a test comment")).To(Equal(expected))
		})
	})

	Describe("The forge.MyDriveNFSVolumeMountInfo function", func() {
		It("Should return the correct NFSVolumeMountInfo object", func() {
			expected := forge.NFSVolumeMountInfo{
				VolumeName:    "mydrive",
				ServerAddress: "nfs.example.com",
				ExportPath:    "/nfs/path",
				MountPath:     "/media/mydrive",
				ReadOnly:      false,
			}
			Expect(forge.MyDriveNFSVolumeMountInfo(nfsServerAddress, nfsExportPath)).To(Equal(expected))
		})
	})

	Describe("The forge.ShVolNFSVolumeMountInfo function", func() {
		It("Should return the correct NFSVolumeMountInfo object", func() {
			expected := forge.NFSVolumeMountInfo{
				VolumeName:    "nfs7",
				ServerAddress: "nfs.example.com",
				ExportPath:    "/nfs/path",
				MountPath:     "/mnt/path",
				ReadOnly:      true,
			}
			Expect(forge.ShVolNFSVolumeMountInfo(7, &shvol, mountInfo)).To(Equal(expected))
		})
	})

	Describe("The forge.NFSShVolSpec function", func() {
		When("the PV has not the CSI params", func() {
			var pv v1.PersistentVolume

			It("Should return empty strings", func() {
				server, path := forge.NFSShVolSpec(&pv)
				Expect(server).To(Equal(""))
				Expect(path).To(Equal(""))
			})
		})

		When("the PV has the CSI params", func() {
			var pv v1.PersistentVolume

			JustBeforeEach(func() {
				pv.Spec.CSI = &v1.CSIPersistentVolumeSource{
					VolumeAttributes: map[string]string{
						"server":    "my-nfs",
						"clusterID": "local",
						"share":     nfsExportPath,
					},
				}
			})

			It("Should return empty strings", func() {
				server, path := forge.NFSShVolSpec(&pv)
				Expect(server).To(Equal("my-nfs.local"))
				Expect(path).To(Equal("/nfs/path"))
			})
		})
	})

	Describe("The forge.PVCProvisioningJobSpec function", func() {
		var pvc v1.PersistentVolumeClaim

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

	var _ = Describe("GetMyDrivePVCName", func() {
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

	var _ = Describe("ConfigureMyDrivePVC", func() {
		It("Should initialize labels if nil and set PVC properties", func() {
			pvc := &v1.PersistentVolumeClaim{}
			storageClassName := "example-storage-class"
			storageSize := resource.MustParse("10Gi")
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureMyDrivePVC(pvc, storageClassName, storageSize, labels)

			Expect(pvc.Labels).ToNot(BeNil())
			Expect(pvc.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(pvc.Spec.AccessModes).To(ConsistOf(v1.ReadWriteMany))
			Expect(pvc.Spec.StorageClassName).ToNot(BeNil())
			Expect(*pvc.Spec.StorageClassName).To(Equal(storageClassName))
			Expect(pvc.Spec.Resources.Requests).ToNot(BeNil())
			Expect(pvc.Spec.Resources.Requests.Storage().String()).To(Equal("10Gi"))
		})

		It("Should preserve existing labels and update storage size if larger", func() {
			existingSize := resource.MustParse("5Gi")
			pvc := &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
				Spec: v1.PersistentVolumeClaimSpec{
					Resources: v1.VolumeResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceStorage: existingSize,
						},
					},
				},
			}

			storageClassName := "example-storage-class"
			storageSize := resource.MustParse("10Gi")
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureMyDrivePVC(pvc, storageClassName, storageSize, labels)

			Expect(pvc.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(pvc.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(pvc.Spec.Resources.Requests).ToNot(BeNil())
			Expect(pvc.Spec.Resources.Requests.Storage().String()).To(Equal("10Gi"))
		})

		It("Should not update storage size if existing size is larger", func() {
			existingSize := resource.MustParse("20Gi")
			pvc := &v1.PersistentVolumeClaim{
				Spec: v1.PersistentVolumeClaimSpec{
					Resources: v1.VolumeResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceStorage: existingSize,
						},
					},
				},
			}

			storageClassName := "example-storage-class"
			storageSize := resource.MustParse("10Gi")
			labels := map[string]string{}

			forge.ConfigureMyDrivePVC(pvc, storageClassName, storageSize, labels)

			Expect(pvc.Spec.Resources.Requests).ToNot(BeNil())
			Expect(pvc.Spec.Resources.Requests.Storage().String()).To(Equal("20Gi"))
		})
	})

	var _ = Describe("ConfigureMyDriveSecret", func() {
		It("Should initialize labels if nil and set required data", func() {
			secret := &v1.Secret{}
			serverName := "nfs-server.example.com"
			path := "/exports/tenant-home"
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureMyDriveSecret(secret, serverName, path, labels)

			Expect(secret.Labels).ToNot(BeNil())
			Expect(secret.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(secret.Type).To(Equal(v1.SecretTypeOpaque))
			Expect(secret.Data).ToNot(BeNil())
			Expect(secret.Data).To(HaveLen(2))
			Expect(secret.Data["server-name"]).To(Equal([]byte("nfs-server.example.com")))
			Expect(secret.Data["path"]).To(Equal([]byte("/exports/tenant-home")))
		})

		It("Should preserve existing labels and update data", func() {
			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
			}

			serverName := "nfs-server.example.com"
			path := "/exports/tenant-home"
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureMyDriveSecret(secret, serverName, path, labels)

			Expect(secret.Labels).ToNot(BeNil())
			Expect(secret.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(secret.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(secret.Type).To(Equal(v1.SecretTypeOpaque))
			Expect(secret.Data).ToNot(BeNil())
			Expect(secret.Data).To(HaveLen(2))
			Expect(secret.Data["server-name"]).To(Equal([]byte("nfs-server.example.com")))
			Expect(secret.Data["path"]).To(Equal([]byte("/exports/tenant-home")))
		})
	})

	var _ = Describe("UpdatePVCProvisioningJobLabel", func() {
		It("Should initialize labels if nil and set provisioning job label", func() {
			pvc := &v1.PersistentVolumeClaim{}

			forge.UpdatePVCProvisioningJobLabel(pvc, "job-123")

			Expect(pvc.Labels).ToNot(BeNil())
			Expect(pvc.Labels).To(HaveKeyWithValue("crownlabs.polito.it/volume-provisioning", "job-123"))
		})

		It("Should update provisioning job label on existing labels", func() {
			pvc := &v1.PersistentVolumeClaim{
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

	var _ = Describe("ConfigureMyDriveProvisioningJob", func() {
		It("Should configure the job spec using PVCProvisioningJobSpec", func() {
			pvc := &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: "test-namespace",
				},
			}
			job := &batchv1.Job{}

			forge.ConfigureMyDriveProvisioningJob(job, pvc)

			// Verify that the spec is correctly configured
			Expect(job.Spec.Template.Spec.RestartPolicy).To(Equal(v1.RestartPolicyOnFailure))
			Expect(job.Spec.BackoffLimit).ToNot(BeNil())
			Expect(*job.Spec.BackoffLimit).To(Equal(int32(forge.ProvisionJobMaxRetries)))
			Expect(job.Spec.TTLSecondsAfterFinished).ToNot(BeNil())
			Expect(*job.Spec.TTLSecondsAfterFinished).To(Equal(int32(forge.ProvisionJobTTLSeconds)))

			// Verify that there is a container
			Expect(job.Spec.Template.Spec.Containers).To(HaveLen(1))
		})
	})
})
