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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

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
})
