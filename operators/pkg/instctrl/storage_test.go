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

package instctrl_test

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
)

var _ = Describe("Storage enforcement", func() {
	Describe("The EnforceShVolMirrorPVCs function", func() {
		const (
			workspaceName     = "netgroup"
			tenantName        = "tester-mirror"
			templateName      = "kubernetes"
			templateNamespace = "workspace-netgroup"
			environmentName   = "control-plane"
			instanceName      = "kubernetes-0000"
			instanceNamespace = "tenant-tester-mirror"

			image       = "internal/registry/image:v1.0"
			cpu         = 2
			cpuReserved = 25
			memory      = "1250M"
			disk        = "20Gi"

			shvolName1    = "shvol-1"
			shvolPVCName1 = "shvol-shvol-1"
			shvolName2    = "shvol-2"
			shvolPVCName2 = "shvol-shvol-2"
			mountPath     = "/mnt/path"
		)

		var (
			ctx           context.Context
			clientBuilder fake.ClientBuilder
			reconciler    instctrl.InstanceReconciler
			containerOpts forge.ContainerEnvOpts

			environment clv1alpha2.Environment
			instance    clv1alpha2.Instance
			shvol1      clv1alpha2.SharedVolume
			shvol2      clv1alpha2.SharedVolume
			shvolMounts []clv1alpha2.SharedVolumeMountInfo

			errReconciler error
			errList       error
			pvcs          corev1.PersistentVolumeClaimList
		)

		RemovePVC := func(pvc *corev1.PersistentVolumeClaim) {
			ctrlUtil.RemoveFinalizer(pvc, "kubernetes.io/pvc-protection")
			Expect(reconciler.Update(ctx, pvc.DeepCopy())).To(Succeed())
			_ = reconciler.Delete(ctx, pvc)
		}

		BeforeEach(func() {
			ctx = ctrl.LoggerInto(context.Background(), logr.Discard())

			ns := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: instanceNamespace},
			}

			containerOpts = forge.ContainerEnvOpts{
				ImagesTag:            "v1.2.3",
				XVncImg:              "x-vnc",
				WebsockifyImg:        "wskfy",
				ContentDownloaderImg: "archdownloader:v0.1.2",
			}

			environment = clv1alpha2.Environment{
				Name:               environmentName,
				EnvironmentType:    clv1alpha2.ClassContainer,
				Image:              image,
				MountMyDriveVolume: false,
				Resources: clv1alpha2.EnvironmentResources{
					CPU:                   cpu,
					ReservedCPUPercentage: cpuReserved,
					Memory:                resource.MustParse(memory),
					Disk:                  resource.MustParse(disk),
				},
			}
			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
				Spec: clv1alpha2.InstanceSpec{
					Running:  true,
					Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
					Tenant:   clv1alpha2.GenericRef{Name: tenantName},
				},
				Status: clv1alpha2.InstanceStatus{
					Environments: []clv1alpha2.InstanceStatusEnv{
						{Name: environmentName, Phase: ""},
					},
				},
			}

			shvol1 = clv1alpha2.SharedVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:      shvolName1,
					Namespace: templateNamespace,
				},
			}
			shvol2 = clv1alpha2.SharedVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:      shvolName2,
					Namespace: templateNamespace,
				},
			}

			clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
				&ns,
				&instance,
				&shvol1,
				&shvol2,
			)
		})

		JustBeforeEach(func() {
			environment.SharedVolumeMounts = shvolMounts

			reconciler = instctrl.InstanceReconciler{
				Client: clientBuilder.Build(), Scheme: scheme.Scheme,
				ContainerEnvOpts: containerOpts,
			}

			ctx, _ = clctx.InstanceInto(ctx, &instance)
			ctx, _ = clctx.EnvironmentInto(ctx, &environment)

			errReconciler = reconciler.EnforceShVolMirrorPVCs(ctx)
			errList = reconciler.List(ctx, &pvcs,
				client.InNamespace(instanceNamespace),
				client.MatchingLabels{forge.LabelVolumeTypeKey: forge.VolumeTypeValueMirror},
			)
		})

		When("No SharedVolumes are required", func() {
			BeforeEach(func() {
				shvolMounts = []clv1alpha2.SharedVolumeMountInfo{}
			})

			It("should enforce without error", func() {
				Expect(errReconciler).ToNot(HaveOccurred())
			})

			It("should enforce the correct PVCs", func() {
				Expect(errList).ToNot(HaveOccurred())
				Expect(len(pvcs.Items)).To(Equal(0))
			})
		})

		When("1 R/W SharedVolume is required", func() {
			BeforeEach(func() {
				shvolMounts = []clv1alpha2.SharedVolumeMountInfo{{
					SharedVolumeRef: clv1alpha2.GenericRef{
						Name:      shvolName1,
						Namespace: templateNamespace,
					},
					MountPath: mountPath,
					ReadOnly:  false,
				}}
			})

			It("should enforce without error", func() {
				Expect(errReconciler).ToNot(HaveOccurred())
			})

			It("should enforce the correct PVCs", func() {
				Expect(errList).ToNot(HaveOccurred())
				Expect(len(pvcs.Items)).To(Equal(1))
				Expect(pvcs.Items[0].Name).To(Equal(shvolName1 + "-" + instanceName + "-mirror"))
			})
		})

		When("1 R/O SharedVolume is required", func() {
			BeforeEach(func() {
				shvolMounts = []clv1alpha2.SharedVolumeMountInfo{{
					SharedVolumeRef: clv1alpha2.GenericRef{
						Name:      shvolName1,
						Namespace: templateNamespace,
					},
					MountPath: mountPath,
					ReadOnly:  true,
				}}
			})

			It("should enforce without error", func() {
				Expect(errReconciler).ToNot(HaveOccurred())
			})

			It("should enforce the correct PVCs", func() {
				Expect(errList).ToNot(HaveOccurred())
				Expect(len(pvcs.Items)).To(Equal(1))
				Expect(pvcs.Items[0].Name).To(Equal(shvolName1 + "-" + instanceName + "-mirror"))
			})
		})

		When("2 SharedVolumes are required", func() {
			BeforeEach(func() {
				shvolMounts = []clv1alpha2.SharedVolumeMountInfo{
					{
						SharedVolumeRef: clv1alpha2.GenericRef{
							Name:      shvolName1,
							Namespace: templateNamespace,
						},
						MountPath: mountPath + "1",
						ReadOnly:  true,
					},
					{
						SharedVolumeRef: clv1alpha2.GenericRef{
							Name:      shvolName2,
							Namespace: templateNamespace,
						},
						MountPath: mountPath + "2",
						ReadOnly:  false,
					},
				}
			})

			It("should enforce without error", func() {
				Expect(errReconciler).ToNot(HaveOccurred())
			})

			It("should enforce the correct PVCs", func() {
				Expect(errList).ToNot(HaveOccurred())
				Expect(len(pvcs.Items)).To(Equal(2))

				expectedNames := []string{
					shvolName1 + "-" + instanceName + "-mirror",
					shvolName2 + "-" + instanceName + "-mirror",
				}
				Expect(expectedNames).To(ContainElement(pvcs.Items[0].Name))
				Expect(expectedNames).To(ContainElement(pvcs.Items[1].Name))
			})
		})

		AfterEach(func() {
			for i := range pvcs.Items {
				RemovePVC(&pvcs.Items[i])
			}
		})
	})
})
