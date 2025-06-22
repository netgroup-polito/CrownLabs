package instctrl_test

import (
	"context"
	"fmt"
	"time"

	controlplanekamajiv1 "github.com/clastix/cluster-api-control-plane-provider-kamaji/api/v1alpha1"
	"github.com/go-logr/logr"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	infrav1 "sigs.k8s.io/cluster-api-provider-kubevirt/api/v1alpha1"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Generation of the Cluster instances", func() {
	var (
		ctx                context.Context
		clientBuilder      fake.ClientBuilder
		reconciler         instctrl.InstanceReconciler
		kamajiinfra        infrav1.KubevirtCluster
		instance           clv1alpha2.Instance
		template           clv1alpha2.Template
		environment        clv1alpha2.Environment
		tenant             clv1alpha2.Tenant
		key                types.NamespacedName
		objectName         types.NamespacedName
		secret             corev1.Secret
		cluster            capiv1.Cluster
		machinedeployment  capiv1.MachineDeployment
		kamajicontrolplane controlplanekamajiv1.KamajiControlPlane
		ownerRef           metav1.OwnerReference

		err error
	)
	const (
		instanceName      = "kubernetes-0000"
		instanceNamespace = "tenant-tester"
		templateName      = "kubernetes"
		templateNamespace = "workspace-netgroup"
		environmentName   = "control-plane"
		tenantName        = "tester"
		workspaceName     = "netgroup"
		webdavCredentials = "webdav-credentials"
		podsNet           = "10.80.0.0/16"
		servicesNet       = "10.95.0.0/16"
		image             = "internal/registry/image:v1.0"
		clusterName       = "testdemo"
		replicas          = 1
		cpu               = 2
		cpuReserved       = 25
		memory            = "1250M"
		disk              = "20Gi"
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
			// These objects are required by the EnforceCloudInitSecret function.
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: webdavCredentials, Namespace: instanceNamespace},
				Data: map[string][]byte{
					instctrl.WebdavSecretUsernameKey: []byte("username"),
					instctrl.WebdavSecretPasswordKey: []byte("password"),
				},
			},
			&clv1alpha2.Template{ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateNamespace}},
			&clv1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: tenantName}},
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
			Name:            environmentName,
			EnvironmentType: clv1alpha2.ClassCluster,
			Image:           image,
			Cluster: &clv1alpha2.ClusterTemplate{
				Name: clusterName,
				ClusterNet: clv1alpha2.ClusterNetwork{
					Pods:     podsNet,
					Services: servicesNet,
				},
				ControlPlane: clv1alpha2.ControlPlaneRef{
					Replicas: replicas,
				},
				MachineDeploy: clv1alpha2.MachineDeployment{
					Replicas: replicas,
				},
			},
			Visulizer: &clv1alpha2.VisualizationType{
				Isvisualizer: true,
			},
			Resources: clv1alpha2.EnvironmentResources{
				CPU:                   cpu,
				ReservedCPUPercentage: cpuReserved,
				Memory:                resource.MustParse(memory),
				Disk:                  resource.MustParse(disk),
			},
		}
		template = clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateNamespace},
			Spec: clv1alpha2.TemplateSpec{
				WorkspaceRef:    clv1alpha2.GenericRef{Name: workspaceName},
				EnvironmentList: []clv1alpha2.Environment{environment},
			},
		}
		tenant = clv1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: tenantName}}

		objectName = forge.NamespacedName(&instance)
		key = types.NamespacedName{
			Name:      fmt.Sprintf("%s-cluster", clusterName),
			Namespace: instance.Namespace,
		}
		secret = corev1.Secret{}
		cluster = capiv1.Cluster{}
		kamajiinfra = infrav1.KubevirtCluster{}
		kamajicontrolplane = controlplanekamajiv1.KamajiControlPlane{}
		machinedeployment = capiv1.MachineDeployment{}
		ownerRef = metav1.OwnerReference{
			APIVersion:         clv1alpha2.GroupVersion.String(),
			Kind:               "Instance",
			Name:               instance.GetName(),
			UID:                instance.GetUID(),
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
		}

	})

	JustBeforeEach(func() {
		reconciler = instctrl.InstanceReconciler{Client: clientBuilder.Build(), Scheme: scheme.Scheme}

		ctx, _ = clctx.InstanceInto(ctx, &instance)
		ctx, _ = clctx.TemplateInto(ctx, &template)
		ctx, _ = clctx.EnvironmentInto(ctx, &environment)
		ctx, _ = clctx.TenantInto(ctx, &tenant)
		err = reconciler.EnforceClusterEnvironment(ctx)
	})

	Context("The environment mode is Standard", func() {
		BeforeEach(func() {
			environment.Mode = clv1alpha2.ModeStandard
		})
		It("Should enforce the cloud-init secret", func() {
			// Here, we only check the secret presence to assert the function execution, leaving the other assertions to the proper tests.
			Expect(reconciler.Get(ctx, objectName, &secret)).To(Succeed())
		})
	})
	Context("The ControlPlaneProvider is kamaji", func() {
		BeforeEach(func() {
			environment.Cluster.ControlPlane.Provider = clv1alpha2.ProviderKamaji
		})
		When("the Cluster is not yet present", func() {
			It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
			It("The Cluster should be present and have the common attributes", func() {
				Expect(reconciler.Get(ctx, key, &cluster)).To(Succeed())
				Expect(cluster.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
				Expect(cluster.GetOwnerReferences()).To(ContainElement(ownerRef))
			})
			It("The Cluster should be present and  have the expected specs", func() {
				Expect(reconciler.Get(ctx, key, &cluster)).To(Succeed())
				cluster.Spec.ClusterNetwork = ptr.To(forge.ClusterNetworking(&environment))
				cluster.Spec.ControlPlaneRef = ptr.To(corev1.ObjectReference{
					APIVersion: "controlplane.cluster.x-k8s.io/v1alpha1",
					Kind:       "KamajiControlPlane",
					Name:       fmt.Sprintf("%s-control-plane", environment.Cluster.Name),
					Namespace:  instance.Namespace,
				})
				cluster.Spec.InfrastructureRef = ptr.To(corev1.ObjectReference{
					APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha1",
					Kind:       "KubevirtCluster",
					Name:       fmt.Sprintf("%s-infra", environment.Cluster.Name),
					Namespace:  instance.Namespace,
				})
				Expect(cluster.Spec).To(Equal(forge.ClusterSpec(&instance, &environment)))
			})
		})
		WhenClusterAlreadyPresentCase := func() {
			BeforeEach(func() {
				existing := capiv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-cluster", clusterName),
						Namespace: instance.Namespace,
					},
					Status: capiv1.ClusterStatus{Phase: capiv1.ProvisionedV1Beta2Reason},
				}
				existing.SetCreationTimestamp(metav1.NewTime(time.Now()))
				clientBuilder.WithObjects(&existing)
			})
			It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
			It("The Cluster should still be present and have the common attributes", func() {
				Expect(reconciler.Get(ctx, key, &cluster)).To(Succeed())
				Expect(cluster.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
				Expect(cluster.GetOwnerReferences()).To(ContainElement(ownerRef))
			})
			It("The Cluster should stil be present and  have the expected specs", func() {
				Expect(reconciler.Get(ctx, key, &cluster)).To(Succeed())
				cluster.Spec.ClusterNetwork = ptr.To(forge.ClusterNetworking(&environment))
				cluster.Spec.ControlPlaneRef = ptr.To(corev1.ObjectReference{
					APIVersion: "controlplane.cluster.x-k8s.io/v1alpha1",
					Kind:       "KamajiControlPlane",
					Name:       fmt.Sprintf("%s-control-plane", environment.Cluster.Name),
					Namespace:  instance.Namespace,
				})
				cluster.Spec.InfrastructureRef = ptr.To(corev1.ObjectReference{
					APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha1",
					Kind:       "KubevirtCluster",
					Name:       fmt.Sprintf("%s-infra", environment.Cluster.Name),
					Namespace:  instance.Namespace,
				})
				Expect(cluster.Spec).To(Equal(forge.ClusterSpec(&instance, &environment)))
			})
			It("Should set the correct instance phase", func() {
				Expect(instance.Status.Phase).To(BeIdenticalTo(clv1alpha2.EnvironmentPhaseRunning))
			})
			When("the Infra is not present", func() {
				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
				It("Cluster should still present", func() {
					Expect(reconciler.Get(ctx, key, &cluster)).To(Succeed())
					Expect(cluster.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
					Expect(cluster.GetOwnerReferences()).To(ContainElement(ownerRef))
				})
			})
			When("the Infra is presenet", func() {
				BeforeEach(func() {
					key = types.NamespacedName{
						Name:      fmt.Sprintf("%s-infra", clusterName),
						Namespace: instance.Namespace,
					}
				})
				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
				It("The Infra should stil be present with right spec ", func() {
					Expect(reconciler.Get(ctx, key, &kamajiinfra)).To(Succeed())
					Expect(kamajiinfra.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
				})
				When("kamajicontrolplane is not present yet", func() {
					It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
					It("The Infra should stil be present with right spec ", func() {
						Expect(reconciler.Get(ctx, key, &kamajiinfra)).To(Succeed())
						Expect(kamajiinfra.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
					})
				})
				When("kamajicontrolplane is present", func() {
					var existing controlplanekamajiv1.KamajiControlPlane
					var host string

					BeforeEach(func() {
						host = "crownlabs.polito.it"

						key = types.NamespacedName{
							Name:      fmt.Sprintf("%s-control-plane", clusterName),
							Namespace: instance.Namespace,
						}

						existing = controlplanekamajiv1.KamajiControlPlane{
							ObjectMeta: forge.NamespacedNameToObjectMeta(objectName),
							Spec:       forge.KamajiControlPlaneSpec(&environment, host),
							Status: controlplanekamajiv1.KamajiControlPlaneStatus{
								Ready:    true,
								Replicas: replicas,
							},
						}
						existing.SetCreationTimestamp(metav1.NewTime(time.Now()))
						clientBuilder.WithObjects(&existing)
					})

					It("Should not return an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					It("The kamajicontrolplane should be present with the expected key fields", func() {
						Expect(reconciler.Get(ctx, key, &kamajicontrolplane)).To(Succeed())
						expected := forge.KamajiControlPlaneSpec(&environment, host)
						Expect(kamajicontrolplane.Spec.Version).To(Equal(expected.Version))
					})
				})
			})
			When("MachineDeployment is not present yet", func() {
				When("Controlplane is not present yet", func() {
					BeforeEach(func() {
						key = types.NamespacedName{
							Name: fmt.Sprintf("%s-md", clusterName), Namespace: instance.Namespace,
						}
					})
					It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
					It("MachineDeployment should meet the right spec", func() {
						Expect(reconciler.Get(ctx, key, &machinedeployment)).To(Succeed())
						machinedeployment.Spec.Template.Spec = capiv1.MachineSpec{
							ClusterName: fmt.Sprintf("%s-cluster", environment.Cluster.Name),
							Version:     ptr.To(environment.Cluster.Version),
							Bootstrap: capiv1.Bootstrap{
								ConfigRef: ptr.To(forge.BootstrapConfigRef(&instance, &environment)),
							},
							InfrastructureRef: forge.MachineInfrastructureRef(&instance, &environment, fmt.Sprintf("%s-md-worker", environment.Cluster.Name)),
						}
						Expect(machinedeployment.Spec.Template.Spec).To(Equal(forge.MachineDeploymentSepc(&instance, &environment)))
					})
				})
				When("MachineDeployment is present yet", func() {
					var existing capiv1.MachineDeployment
					BeforeEach(func() {
						key = types.NamespacedName{
							Name: fmt.Sprintf("%s-md", clusterName), Namespace: instance.Namespace,
						}
						existing = capiv1.MachineDeployment{
							ObjectMeta: forge.NamespacedNameToObjectMeta(objectName),
							Spec: capiv1.MachineDeploymentSpec{
								Template: capiv1.MachineTemplateSpec{
									Spec: forge.MachineDeploymentSepc(&instance, &environment),
								},
							},
							Status: capiv1.MachineDeploymentStatus{
								Replicas: replicas,
							},
						}
						existing.SetCreationTimestamp(metav1.NewTime(time.Now()))
						clientBuilder.WithObjects(&existing)

					})
					It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
					It("MachineDeployment should still meet the right spec", func() {
						Expect(reconciler.Get(ctx, key, &machinedeployment)).To(Succeed())
						machinedeployment.Spec.Template.Spec = capiv1.MachineSpec{
							ClusterName: fmt.Sprintf("%s-cluster", environment.Cluster.Name),
							Version:     ptr.To(environment.Cluster.Version),
							Bootstrap: capiv1.Bootstrap{
								ConfigRef: ptr.To(forge.BootstrapConfigRef(&instance, &environment)),
							},
							InfrastructureRef: forge.MachineInfrastructureRef(&instance, &environment, fmt.Sprintf("%s-md-worker", environment.Cluster.Name)),
						}
						Expect(machinedeployment.Spec.Template.Spec).To(Equal(forge.MachineDeploymentSepc(&instance, &environment)))
					})
				})
			})
		}
		When("the cluster is already present and it is running", func() { WhenClusterAlreadyPresentCase() })

	})
})
