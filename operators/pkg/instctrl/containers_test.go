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

package instctrl_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
	tntctrl "github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller"
)

var _ = Describe("Generation of the container based instances", func() {
	// deploymentSpecCleanup removes resources which may conflict because of reordering
	// or internal representation differences.
	deploymentSpecCleanup := func(ds *appsv1.DeploymentSpec) {
		for i := range ds.Template.Spec.Containers {
			ds.Template.Spec.Containers[i].Resources = corev1.ResourceRequirements{}
			ds.Template.Spec.Containers[i].Env = []corev1.EnvVar{}
		}
		for i := range ds.Template.Spec.InitContainers {
			ds.Template.Spec.InitContainers[i].Resources = corev1.ResourceRequirements{}
			ds.Template.Spec.InitContainers[i].Env = []corev1.EnvVar{}
		}
		ds.Selector = nil
		ds.Template.ObjectMeta.Labels = nil
		ds.Template.Spec.NodeSelector = nil
	}

	var (
		ctx           context.Context
		clientBuilder fake.ClientBuilder
		reconciler    instctrl.InstanceReconciler

		instance    clv1alpha2.Instance
		environment clv1alpha2.Environment

		objectName types.NamespacedName
		svc        corev1.Service
		deploy     appsv1.Deployment
		pvc        corev1.PersistentVolumeClaim

		myDriveSecret corev1.Secret

		ownerRef metav1.OwnerReference

		shvol       clv1alpha2.SharedVolume
		shvolMounts []clv1alpha2.SharedVolumeMountInfo
		mountInfos  []forge.NFSVolumeMountInfo

		containerOpts forge.ContainerEnvOpts

		err      error
		errShVol error
	)

	const (
		instanceName      = "kubernetes-0000"
		instanceNamespace = "tenant-tester"
		templateName      = "kubernetes"
		templateNamespace = "workspace-netgroup"
		environmentName   = "control-plane"
		tenantName        = "tester"

		image       = "internal/registry/image:v1.0"
		cpu         = 2
		cpuReserved = 25
		memory      = "1250M"
		disk        = "20Gi"

		nfsServerName     = "rook-nfs-server-name"
		nfsMyDriveExpPath = "/nfs/path"
		nfsShVolName      = "nfs0"
		nfsShVolExpPath   = "/nfs/shvol"
		nfsShVolMountPath = "/mnt/path"
		nfsShVolReadOnly  = true
		shVolName         = "myshvol"
		shVolNamespace    = "default"
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		containerOpts = forge.ContainerEnvOpts{
			ImagesTag:            "v1.2.3",
			XVncImg:              "x-vnc",
			WebsockifyImg:        "wskfy",
			ContentDownloaderImg: "archdownloader:v0.1.2",
		}
		myDriveSecret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tntctrl.NFSSecretName,
				Namespace: instanceNamespace,
			},
			Data: map[string][]byte{
				tntctrl.NFSSecretServerNameKey: []byte(nfsServerName),
				tntctrl.NFSSecretPathKey:       []byte(nfsMyDriveExpPath),
			},
		}
		clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
			&clv1alpha2.Template{ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateNamespace}},
			&myDriveSecret,
		)

		instance = clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
			Spec: clv1alpha2.InstanceSpec{
				Running:  true,
				Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
				Tenant:   clv1alpha2.GenericRef{Name: tenantName},
			},
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

		objectName = forge.NamespacedName(&instance)

		svc = corev1.Service{}
		deploy = appsv1.Deployment{}
		pvc = corev1.PersistentVolumeClaim{}

		ownerRef = metav1.OwnerReference{
			APIVersion:         clv1alpha2.GroupVersion.String(),
			Kind:               "Instance",
			Name:               instance.GetName(),
			UID:                instance.GetUID(),
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
		}

		shvol = clv1alpha2.SharedVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      shVolName,
				Namespace: shVolNamespace,
			},
			Spec: clv1alpha2.SharedVolumeSpec{
				PrettyName: "My Pretty Name",
				Size:       resource.MustParse("1Gi"),
			},
			Status: clv1alpha2.SharedVolumeStatus{},
		}
		shvolMounts = []clv1alpha2.SharedVolumeMountInfo{
			{
				SharedVolumeRef: clv1alpha2.GenericRef{
					Name:      shvol.Name,
					Namespace: shvol.Namespace,
				},
				MountPath: nfsShVolMountPath,
				ReadOnly:  nfsShVolReadOnly,
			},
		}
		mountInfos = []forge.NFSVolumeMountInfo{
			forge.MyDriveNFSVolumeMountInfo(nfsServerName, nfsMyDriveExpPath),
			forge.ShVolNFSVolumeMountInfo(0, &shvol, shvolMounts[0]),
		}
	})

	JustBeforeEach(func() {

		reconciler = instctrl.InstanceReconciler{
			Client: clientBuilder.Build(), Scheme: scheme.Scheme,
			ContainerEnvOpts: containerOpts,
		}

		ctx, _ = clctx.InstanceInto(ctx, &instance)
		ctx, _ = clctx.EnvironmentInto(ctx, &environment)
		errShVol = reconciler.Create(ctx, &shvol)
		err = reconciler.EnforceContainerEnvironment(ctx)
	})

	It("Should enforce the environment exposition objects", func() {
		// Here, we only check the service presence to assert the function execution, leaving the other assertions to the proper tests.
		Expect(reconciler.Get(ctx, objectName, &svc)).To(Succeed())
	})

	Context("The environment is not persistent", func() {
		BeforeEach(func() { environment.Persistent = false })

		When("the deployment it is not yet present", func() {
			When("the instance is running", func() {
				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The deployment should be present and have the common attributes", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					Expect(deploy.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
					Expect(deploy.GetOwnerReferences()).To(ContainElement(ownerRef))
				})

				It("The deployment should be present and have the expected specs", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					expected := forge.DeploymentSpec(&instance, &environment, nil, &containerOpts)
					expected.Replicas = forge.ReplicasCount(&instance, &environment, false)

					// These labels are checked here since it BeEquivalentTo ignores reordering. They are removed from the spec in deploymentSpecCleanup.
					Expect(deploy.Spec.Selector.MatchLabels).To(BeEquivalentTo(expected.Selector.MatchLabels))
					Expect(deploy.Spec.Template.GetLabels()).To(BeEquivalentTo(expected.Template.GetLabels()))

					// Here we skip checking of deploment resources and env, since they would have a different representation due to the
					// marshaling/unmarshaling process. Still, the correctness of the value is already checked with the
					// appropriate test case.
					deploymentSpecCleanup(&deploy.Spec)
					deploymentSpecCleanup(&expected)

					Expect(deploy.Spec).To(Equal(expected))
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 1)))
				})

				It("Should set the instance phase to starting", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseStarting))
				})
			})

			When("the instance is not running", func() {
				BeforeEach(func() {
					instance.Spec.Running = false
				})

				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The deployment replicas should be 0", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 0)))
				})

				It("Should set the instance phase to Off", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseOff))
				})
			})
		})

		When("the deployment is already present", func() {
			var existing appsv1.Deployment

			BeforeEach(func() {
				existing = appsv1.Deployment{
					ObjectMeta: forge.NamespacedNameToObjectMeta(objectName),
					Spec:       appsv1.DeploymentSpec{},
					Status:     appsv1.DeploymentStatus{ReadyReplicas: 1},
				}
				existing.SetCreationTimestamp(metav1.NewTime(time.Now()))
				clientBuilder.WithObjects(&existing)
			})

			When("the instance is running", func() {
				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The deployment should still be present and have the common attributes", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					Expect(deploy.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
					Expect(deploy.GetOwnerReferences()).To(ContainElement(ownerRef))
				})

				It("The deployment should still be present and have unmodified specs", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					// Here we overwrite the replicas value, as it is checked in a different It clause.
					deploy.Spec.Replicas = nil
					Expect(deploy.Spec).To(Equal(appsv1.DeploymentSpec{}))
				})

				It("The deployment should be present and with the replicas set to one", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 1)))
				})

				It("Should set the correct instance phase", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseReady))
				})
			})

			When("the instance is not running", func() {
				BeforeEach(func() { instance.Spec.Running = false })

				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("The deployment should still be present and have unmodified specs", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					Expect(deploy.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
					// Here we overwrite the replicas value, as it is checked in a different It clause.
					deploy.Spec.Replicas = nil
					Expect(deploy.Spec).To(Equal(appsv1.DeploymentSpec{}))
				})

				It("The deployment should be present and with the replicas set to one", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 1)))
				})

				It("Should set the instance phase to Off", func() {
					Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseOff))
				})
			})
		})
	})

	Context("The environment is persistent", func() {
		BeforeEach(func() { environment.Persistent = true })

		When("the PVC is not yet present", func() {
			It("The PVC should be present and have the common attributes", func() {
				Expect(reconciler.Get(ctx, objectName, &pvc)).To(Succeed())
				Expect(pvc.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
				Expect(pvc.GetOwnerReferences()).To(ContainElement(ownerRef))
			})

			It("The PVC should be present and have the expected specs", func() {
				Expect(reconciler.Get(ctx, objectName, &pvc)).To(Succeed())
				Expect(pvc.Spec).To(Equal(forge.InstancePVCSpec(&environment)))
			})
		})

		When("the deployment is not yet present", func() {
			It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

			It("The deployment should be present and have the common attributes", func() {
				Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
				Expect(deploy.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
				Expect(deploy.GetOwnerReferences()).To(ContainElement(ownerRef))
			})

			It("The deployment should be present and have the expected specs", func() {
				Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
				expected := forge.DeploymentSpec(&instance, &environment, nil, &containerOpts)
				expected.Replicas = forge.ReplicasCount(&instance, &environment, true)

				// These labels are checked here since it BeEquivalentTo ignores reordering. They are removed from the spec in deploymentSpecCleanup.
				Expect(deploy.Spec.Selector.MatchLabels).To(BeEquivalentTo(expected.Selector.MatchLabels))
				Expect(deploy.Spec.Template.GetLabels()).To(BeEquivalentTo(expected.Template.GetLabels()))

				// Here we skip checking of deploment resources and env, since they would have a different representation due to the
				// marshaling/unmarshaling process. Still, the correctness of the value is already checked with the
				// appropriate test case.
				deploymentSpecCleanup(&expected)
				deploymentSpecCleanup(&deploy.Spec)

				Expect(deploy.Spec).To(Equal(expected))
			})

			It("The deployment should be present and with the replicas set to one", func() {
				Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
				Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 1)))
			})

			It("Should set the instance phase as starting", func() {
				Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseStarting))
			})
		})

		When("the PVC is already present", func() {
			BeforeEach(func() {
				existing := corev1.PersistentVolumeClaim{
					ObjectMeta: forge.NamespacedNameToObjectMeta(objectName),
					Spec:       corev1.PersistentVolumeClaimSpec{},
					Status:     corev1.PersistentVolumeClaimStatus{},
				}
				existing.SetCreationTimestamp(metav1.NewTime(time.Now()))
				clientBuilder.WithObjects(&existing)
			})

			It("The PVC should be present and have the common attributes", func() {
				Expect(reconciler.Get(ctx, objectName, &pvc)).To(Succeed())
				Expect(pvc.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
				Expect(pvc.GetOwnerReferences()).To(ContainElement(ownerRef))
			})

			It("The PVC should be present and have unmodified specs", func() {
				Expect(reconciler.Get(ctx, objectName, &pvc)).To(Succeed())
				Expect(pvc.Spec).To(Equal(corev1.PersistentVolumeClaimSpec{}))
			})
		})

		When("the deployment is already present", func() {
			BeforeEach(func() {
				existing := appsv1.Deployment{
					ObjectMeta: forge.NamespacedNameToObjectMeta(objectName),
					Spec:       appsv1.DeploymentSpec{},
					Status:     appsv1.DeploymentStatus{ReadyReplicas: 0},
				}
				existing.SetCreationTimestamp(metav1.NewTime(time.Now()))
				clientBuilder.WithObjects(&existing)
			})

			It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

			It("The deployment should still be present and have the common attributes", func() {
				Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
				Expect(deploy.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
				Expect(deploy.GetOwnerReferences()).To(ContainElement(ownerRef))
			})

			It("The deployment should still be present and have unmodified specs", func() {
				Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
				// Here we overwrite the replicas value, as it is checked in a different It clause.
				deploy.Spec.Replicas = nil
				Expect(deploy.Spec).To(Equal(appsv1.DeploymentSpec{}))
			})

			It("Should set the correct instance phase", func() {
				Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseStarting))
			})

			Context("The instance is running", func() {
				BeforeEach(func() { instance.Spec.Running = true })

				It("The deployment should be present and with one replica", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 1)))
				})
			})

			Context("The instance is not running", func() {
				BeforeEach(func() { instance.Spec.Running = false })

				It("The deployment should be present and with no replicas", func() {
					Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
					Expect(deploy.Spec.Replicas).To(PointTo(BeNumerically("==", 0)))
				})
			})
		})
	})

	Context("The environment needs the personal drive", func() {
		BeforeEach(func() { environment.MountMyDriveVolume = true })

		When("the deployment is not yet present", func() {
			It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

			// The spec of the deployment is checked in forge.
		})

		When("the mydrive-info secret is not present", func() {
			BeforeEach(func() {
				// Set a different name for the mydrive-info secret so the reconciler won't find it.
				myDriveSecret.Name = "wrong-name"
			})

			It("Should return an error", func() { Expect(err).To(HaveOccurred()) })
		})
	})

	Context("The environment has the personal drive and a shared volume to mount", func() {
		BeforeEach(func() {
			environment.MountMyDriveVolume = true
			environment.SharedVolumeMounts = shvolMounts
		})

		When("the deployment is not yet present", func() {
			It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

			// The spec of the deployment is checked in forge.
		})

		When("the shvol is not yet present", func() {
			It("Creating shvol should not return an error", func() { Expect(errShVol).ToNot(HaveOccurred()) })

			It("The deployment should be present and with the correct volumes spec", func() {
				Expect(reconciler.Get(ctx, objectName, &deploy)).To(Succeed())
				expected := forge.DeploymentSpec(&instance, &environment, mountInfos, &containerOpts)

				Expect(deploy.Spec.Template.Spec.Volumes).To(Equal(expected.Template.Spec.Volumes))
			})
		})
	})
})
