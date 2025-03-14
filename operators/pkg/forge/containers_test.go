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

// Package forge groups the methods used to forge the Kubernetes object definitions
package forge_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Containers and Deployment spec forging", func() {

	var (
		instance    clv1alpha2.Instance
		environment clv1alpha2.Environment
		mountInfos  []forge.NFSVolumeMountInfo
		opts        forge.ContainerEnvOpts
		container   corev1.Container
	)

	// example values
	const (
		instanceName         = "kubernetes-0000"
		envName              = "test-environment"
		instanceNamespace    = "tenant-tester"
		image                = "internal/registry/image:v1.0"
		cpu                  = 2
		expectedCPUReqMillis = 500
		expectedCPULimMillis = 2000
		cpuReserved          = 25
		memory               = "1250M"
		disk                 = "20Gi"
		volumeName           = "vol"
		volumePath           = "/path"
		claimName            = "claim"
		envVarName           = "VAR"
		envVarVal            = "VALUE"
		envVarResName        = corev1.ResourceRequestsCPU
		portName             = "some-port"
		portNum              = 1234
		httpPath             = "/some/path"
		httpPathAlternative  = "/some/different/path"
		nfsServerName        = "nfs-server-name"
		nfsMyDriveExpPath    = "/nfs/path"
		nfsShVolName         = "nfs0"
		nfsShVolExpPath      = "/nfs/shvol"
		nfsShVolMountPath    = "/mnt/path"
		nfsShVolReadOnly     = true
	)

	var (
		mountInfoMyDrive = forge.MyDriveNFSVolumeMountInfo(nfsServerName, nfsMyDriveExpPath)
		mountInfoShVol   = forge.NFSVolumeMountInfo{
			VolumeName:    nfsShVolName,
			ServerAddress: nfsServerName,
			ExportPath:    nfsShVolExpPath,
			MountPath:     nfsShVolMountPath,
			ReadOnly:      nfsShVolReadOnly,
		}
	)

	BeforeEach(func() {
		instance = clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
		}
		environment = clv1alpha2.Environment{
			Image: image,
			Name:  envName,
			Mode:  clv1alpha2.ModeStandard,
			Resources: clv1alpha2.EnvironmentResources{
				CPU:                   cpu,
				ReservedCPUPercentage: cpuReserved,
				Memory:                resource.MustParse(memory),
				Disk:                  resource.MustParse(disk),
			},
		}

		mountInfos = []forge.NFSVolumeMountInfo{
			mountInfoMyDrive,
			mountInfoShVol,
		}
		opts = forge.ContainerEnvOpts{
			ImagesTag:            "tag",
			XVncImg:              "x-vnc-img",
			WebsockifyImg:        "wsfy-img",
			ContentDownloaderImg: "cont-dler-img",
			ContentUploaderImg:   "cont-uplr-img",
		}
		container = corev1.Container{}
	})

	Describe("The forge.PVCSpec function", func() {
		var spec corev1.PersistentVolumeClaimSpec

		JustBeforeEach(func() {
			spec = forge.InstancePVCSpec(&environment)
		})

		It("Should set the correct access mode", func() {
			Expect(spec.AccessModes).To(Equal([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}))
		})
		It("Should set the correct resources", func() {
			Expect(spec.Resources.Requests.Storage()).To(PointTo(Equal(environment.Resources.Disk)))
		})
		It("Should leave the storage class name unset", func() {
			Expect(spec.StorageClassName).To(BeNil())
		})

		When("A custom storage class is specified", func() {
			BeforeEach(func() { environment.StorageClassName = "foo" })
			It("Should set the storage class name", func() {
				Expect(spec.StorageClassName).To(PointTo(Equal("foo")))
			})
		})
	})

	Describe("The forge.PodSecurityContext function", func() {
		var psc corev1.PodSecurityContext

		JustBeforeEach(func() {
			psc = *forge.PodSecurityContext()
		})

		It("Should set the correct RunAsGroup", func() {
			Expect(psc.RunAsGroup).To(PointTo(BeNumerically("==", 1010)))
		})
		It("Should set the correct RunAsUser", func() {
			Expect(psc.RunAsUser).To(PointTo(BeNumerically("==", 1010)))
		})
		It("Should set the correct FsGroup", func() {
			Expect(psc.FSGroup).To(PointTo(BeNumerically("==", 1010)))
		})
		It("Should set the correct RunAsNonRoot", func() {
			Expect(psc.RunAsNonRoot).To(PointTo(BeTrue()))
		})
	})

	Describe("The forge.ReplicasCount function", func() {
		type ReplicasCountCase struct {
			Persistent     bool
			Running        bool
			IsNew          bool
			ExpectedOutput int32
		}

		WhenBody := func(c ReplicasCountCase) func() {
			return func() {
				JustBeforeEach(func() {
					instance.Spec.Running = c.Running
					environment.Persistent = c.Persistent
				})

				It("Should return the correct Replica count", func() {
					Expect(forge.ReplicasCount(&instance, &environment, c.IsNew)).To(PointTo(Equal(c.ExpectedOutput)))
				})
			}
		}

		When("the environment is not persistent and instance is new and not running", WhenBody(ReplicasCountCase{
			Persistent:     false,
			Running:        false,
			IsNew:          true,
			ExpectedOutput: 0,
		}))

		When("the environment is not persistent and instance is not new and not running", WhenBody(ReplicasCountCase{
			Persistent:     false,
			Running:        false,
			IsNew:          false,
			ExpectedOutput: 1,
		}))

		When("the environment is not persistent and instance is running", WhenBody(ReplicasCountCase{
			Persistent:     false,
			Running:        true,
			ExpectedOutput: 1,
		}))

		When("the environment is persistent and instance is not running", WhenBody(ReplicasCountCase{
			Persistent:     true,
			Running:        false,
			ExpectedOutput: 0,
		}))

		When("the environment is persistent and instance is running", WhenBody(ReplicasCountCase{
			Persistent:     true,
			Running:        true,
			ExpectedOutput: 1,
		}))
	})

	Describe("The forge.DeploymentSpec function", func() {
		var spec appsv1.DeploymentSpec

		JustBeforeEach(func() {
			spec = forge.DeploymentSpec(&instance, &environment, mountInfos, &opts)
		})

		It("Should set the correct template labels", func() {
			Expect(spec.Template.ObjectMeta.GetLabels()).To(Equal(forge.InstanceSelectorLabels(&instance)))
		})
		It("Should set the correct template spec", func() {
			Expect(spec.Template.Spec).To(Equal(forge.PodSpec(&instance, &environment, mountInfos, &opts)))
		})
		It("Should set the correct selector", func() {
			Expect(spec.Selector.MatchLabels).To(Equal(forge.InstanceSelectorLabels(&instance)))
		})
	})

	Describe("The forge.PodSpec function", func() {
		var spec corev1.PodSpec

		type PodSpecContainersCase struct {
			Mode            clv1alpha2.EnvironmentMode
			EnvironmentType clv1alpha2.EnvironmentType
			ExpectedOutput  func(*clv1alpha2.Instance, *clv1alpha2.Environment) []corev1.Container
		}

		ContainersWhenBody := func(psc PodSpecContainersCase) func() {
			return func() {
				BeforeEach(func() {
					environment.Mode = psc.Mode
					environment.EnvironmentType = psc.EnvironmentType
				})
				It("Should set the correct containers", func() {
					Expect(spec.Containers).To(ConsistOf(psc.ExpectedOutput(&instance, &environment)))
				})
			}
		}

		JustBeforeEach(func() {
			spec = forge.PodSpec(&instance, &environment, mountInfos, &opts)
		})

		It("Should set the security context", func() {
			Expect(spec.SecurityContext).To(Equal(forge.PodSecurityContext()))
		})

		It("Should disable the service account token automount", func() {
			Expect(spec.AutomountServiceAccountToken).To(PointTo(BeFalse()))
		})

		It("Should disable service links", func() {
			Expect(spec.EnableServiceLinks).To(PointTo(BeFalse()))
		})

		It("Should set the container hostname accordingly", func() {
			Expect(spec.Hostname).To(Equal(forge.InstanceHostname(&environment)))
		})

		It("Should set the node selector labels accordingly", func() {
			Expect(spec.NodeSelector).To(Equal(forge.NodeSelectorLabels(&instance, &environment)))
		})

		When("the environment type is Standalone", func() {
			When("the environment mode is Standard", ContainersWhenBody(PodSpecContainersCase{
				Mode:            clv1alpha2.ModeStandard,
				EnvironmentType: clv1alpha2.ClassStandalone,
				ExpectedOutput: func(i *clv1alpha2.Instance, e *clv1alpha2.Environment) []corev1.Container {
					return []corev1.Container{
						forge.StandaloneContainer(i, e, forge.PersistentMountPath(e), mountInfos),
					}
				},
			}))
		})

		When("the environment type is Container", func() {
			When("the environment mode is Standard", ContainersWhenBody(PodSpecContainersCase{
				Mode:            clv1alpha2.ModeStandard,
				EnvironmentType: clv1alpha2.ClassContainer,
				ExpectedOutput: func(i *clv1alpha2.Instance, e *clv1alpha2.Environment) []corev1.Container {
					return []corev1.Container{
						forge.WebsockifyContainer(&opts, e, i),
						forge.XVncContainer(&opts),
						forge.AppContainer(e, forge.PersistentMountPath(e), mountInfos),
					}
				},
			}))

			When("the environment mode is Exercise", ContainersWhenBody(PodSpecContainersCase{
				Mode:            clv1alpha2.ModeExercise,
				EnvironmentType: clv1alpha2.ClassContainer,
				ExpectedOutput: func(i *clv1alpha2.Instance, e *clv1alpha2.Environment) []corev1.Container {
					return []corev1.Container{
						forge.WebsockifyContainer(&opts, e, i),
						forge.XVncContainer(&opts),
						forge.AppContainer(e, forge.PersistentMountPath(e), mountInfos),
					}
				},
			}))

			When("the environment mode is Exam", ContainersWhenBody(PodSpecContainersCase{
				Mode:            clv1alpha2.ModeExam,
				EnvironmentType: clv1alpha2.ClassContainer,
				ExpectedOutput: func(i *clv1alpha2.Instance, e *clv1alpha2.Environment) []corev1.Container {
					return []corev1.Container{
						forge.WebsockifyContainer(&opts, e, i),
						forge.XVncContainer(&opts),
						forge.AppContainer(e, forge.PersistentMountPath(&environment), mountInfos),
					}
				},
			}))
		})
	})

	Describe("The forge.StandaloneContainer function forges a standalone container", func() {
		var actual, expected corev1.Container

		JustBeforeEach(func() {
			actual = forge.StandaloneContainer(&instance, &environment, forge.PersistentMountPath(&environment), mountInfos)
		})

		It("Should set container port", func() {
			Expect(actual.Ports).To(Equal([]corev1.ContainerPort{{
				Name:          "gui",
				ContainerPort: int32(6080),
				Protocol:      corev1.ProtocolTCP},
			}))
		})

		When("RewriteURL is true", func() {
			var probe *corev1.Probe
			BeforeEach(func() {
				probe = forge.ContainerProbe()
				probe.HTTPGet = &corev1.HTTPGetAction{
					Port: intstr.FromString("gui"),
					Path: "/",
				}
				environment.RewriteURL = true
			})
			It("ReadinessProbe URL is /", func() {
				Expect(actual.ReadinessProbe).To(Equal(probe))
			})
		})

		When("RewriteURL is false", func() {
			var probe *corev1.Probe
			BeforeEach(func() {
				probe = forge.ContainerProbe()
				probe.HTTPGet = &corev1.HTTPGetAction{
					Port: intstr.FromString("gui"),
					Path: forge.IngressGUIPath(&instance, &environment),
				}
				environment.RewriteURL = false
			})
			It("ReadinessProbe URL is "+forge.IngressGUIPath(&instance, &environment), func() {
				Expect(actual.ReadinessProbe).To(Equal(probe))
			})

		})

		It("Should set the env variables", func() {
			expected.Name = envName
			forge.AddEnvVariableToContainer(&expected, "CROWNLABS_BASE_PATH", forge.IngressGUICleanPath(&instance))
			forge.AddEnvVariableToContainer(&expected, "CROWNLABS_LISTEN_PORT", "6080")
			forge.AddEnvVariableFromResourcesToContainer(&expected, "CROWNLABS_CPU_REQUESTS", expected.Name, corev1.ResourceRequestsCPU, forge.DefaultDivisor)
			forge.AddEnvVariableFromResourcesToContainer(&expected, "CROWNLABS_CPU_LIMITS", expected.Name, corev1.ResourceLimitsCPU, forge.DefaultDivisor)
			Expect(actual.Env).To(ConsistOf(expected.Env))
		})
	})

	Describe("The forge.WebsockifyContainer function forges a websockify sidecar container", func() {
		var actual, expected corev1.Container
		JustBeforeEach(func() {
			expected = corev1.Container{}
			actual = forge.WebsockifyContainer(&opts, &environment, &instance)
		})

		It("Should set the correct container name and image", func() {
			// PodSecurityContext setting is checked by GenericContainer specific tests
			Expect(actual.Name).To(Equal("websockify"))
			Expect(actual.Image).To(Equal(opts.WebsockifyImg + ":" + opts.ImagesTag))
		})
		It("Should set the correct resources", func() {
			forge.SetContainerResources(&expected, 0.01, 0.1, 30, 100)
			Expect(actual.Resources).To(Equal(expected.Resources))
		})
		It("Should set the tcp port exposition", func() {
			forge.AddTCPPortToContainer(&expected, "gui", 6080)
			forge.AddTCPPortToContainer(&expected, "metrics", 9090)
			Expect(actual.Ports).To(Equal(expected.Ports))
		})
		It("Should set the readiness probe", func() {
			forge.SetContainerReadinessHTTPProbe(&expected, "gui", forge.HealthzEndpoint)
			Expect(actual.ReadinessProbe).To(Equal(expected.ReadinessProbe))
		})
		It("Should set the env varibles", func() {
			expected.Name = forge.WebsockifyName
			forge.AddEnvVariableFromFieldToContainer(&expected, forge.PodNameEnvName, "metadata.name")
			forge.AddEnvVariableFromResourcesToContainer(&expected, forge.AppCPULimitsEnvName, environment.Name, corev1.ResourceLimitsCPU, forge.MilliDivisor)
			forge.AddEnvVariableFromResourcesToContainer(&expected, forge.AppMEMLimitsEnvName, environment.Name, corev1.ResourceLimitsMemory, forge.DefaultDivisor)
			Expect(actual.Env).To(ConsistOf(expected.Env))
		})

		disableCtrlsWhenBody := func(disableCtrls bool) func() {
			return func() {
				BeforeEach(func() {
					environment.DisableControls = disableCtrls
				})

				It("Should set the related argument accordingly", func() {
					Expect(actual.Args).To(ContainElement(
						fmt.Sprintf("--show-controls=%v", !disableCtrls),
					))
					Expect(actual.Args).NotTo(ContainElement(
						fmt.Sprintf("--show-controls=%v", disableCtrls),
					))
				})
			}
		}
		When("disableControls is true", disableCtrlsWhenBody(true))
		When("disableControls is false", disableCtrlsWhenBody(false))

		When("the environment mode is Standard", func() {
			BeforeEach(func() {
				instance.UID = instanceName
				environment.Mode = clv1alpha2.ModeStandard
			})
			It("Should set the correct arguments", func() {
				Expect(actual.Args).To(ConsistOf([]string{
					fmt.Sprintf("--http-addr=:%d", forge.GUIPortNumber),
					fmt.Sprintf("--base-path=%s", forge.IngressGUICleanPath(&instance)),
					fmt.Sprintf("--metrics-addr=:%d", forge.MetricsPortNumber),
					fmt.Sprintf("--show-controls=%v", !environment.DisableControls),
					fmt.Sprintf("--instmetrics-server-endpoint=%s", opts.InstMetricsEndpoint),
					fmt.Sprintf("--pod-name=$(%s)", forge.PodNameEnvName),
					fmt.Sprintf("--cpu-limit=$(%s)", forge.AppCPULimitsEnvName),
					fmt.Sprintf("--memory-limit=$(%s)", forge.AppMEMLimitsEnvName),
				}))
			})
		})

		When("the environment mode is non Standard", func() {
			BeforeEach(func() {
				instance.UID = instanceName
				environment.Mode = clv1alpha2.ModeExercise
			})
			It("Should set the correct arguments", func() {
				Expect(actual.Args).To(ConsistOf([]string{
					fmt.Sprintf("--http-addr=:%d", forge.GUIPortNumber),
					fmt.Sprintf("--base-path=%s", forge.IngressGUICleanPath(&instance)),
					fmt.Sprintf("--metrics-addr=:%d", forge.MetricsPortNumber),
					fmt.Sprintf("--show-controls=%v", !environment.DisableControls),
					fmt.Sprintf("--instmetrics-server-endpoint=%s", opts.InstMetricsEndpoint),
					fmt.Sprintf("--pod-name=$(%s)", forge.PodNameEnvName),
					fmt.Sprintf("--cpu-limit=$(%s)", forge.AppCPULimitsEnvName),
					fmt.Sprintf("--memory-limit=$(%s)", forge.AppMEMLimitsEnvName),
				}))
			})
		})
	})

	Describe("The forge.XVncContainer function forges a x-vnc sidecar container", func() {
		var actual, expected corev1.Container
		xvncName := "xvnc"
		JustBeforeEach(func() {
			actual = forge.XVncContainer(&opts)
		})

		It("Should set the correct container name and image", func() {
			// PodSecurityContext setting is checked by GenericContainer specific tests
			Expect(actual.Name).To(Equal(xvncName))
			Expect(actual.Image).To(Equal(opts.XVncImg + ":" + opts.ImagesTag))
		})
		It("Should set the correct resources", func() {
			forge.SetContainerResources(&expected, 0.05, 0.25, 200, 600)
			Expect(actual.Resources).To(Equal(expected.Resources))
		})
		It("Should set the tcp port exposition", func() {
			forge.AddTCPPortToContainer(&expected, xvncName, 5900)
			Expect(actual.Ports).To(Equal(expected.Ports))
		})
		It("Should set the readiness probe", func() {
			forge.SetContainerReadinessTCPProbe(&expected, xvncName)
			Expect(actual.ReadinessProbe).To(Equal(expected.ReadinessProbe))
		})
	})

	Describe("The forge.AppContainer function forges the main application container", func() {
		var actual, expected corev1.Container

		Context("Has to set the general parameters", func() {
			JustBeforeEach(func() {
				actual = forge.AppContainer(&environment, forge.PersistentMountPath(&environment), mountInfos)
			})

			It("Should set the correct container name and image", func() {
				// PodSecurityContext setting is checked by GenericContainer specific tests
				Expect(actual.Name).To(Equal(environment.Name))
				Expect(actual.Image).To(Equal(environment.Image))
			})
			It("Should set the correct resources", func() {
				forge.SetContainerResourcesFromEnvironment(&expected, &environment)
				Expect(actual.Resources).To(Equal(expected.Resources))
			})
			It("Should NOT set container ports", func() {
				Expect(actual.Ports).To(BeEmpty())
			})
			It("Should NOT set readiness probes", func() {
				Expect(actual.ReadinessProbe).To(BeNil())
			})
			It("Should set the env varibles", func() {
				expected.Name = envName
				forge.AddEnvVariableFromResourcesToContainer(&expected, "CROWNLABS_CPU_REQUESTS", expected.Name, corev1.ResourceRequestsCPU, forge.DefaultDivisor)
				forge.AddEnvVariableFromResourcesToContainer(&expected, "CROWNLABS_CPU_LIMITS", expected.Name, corev1.ResourceLimitsCPU, forge.DefaultDivisor)
				Expect(actual.Env).To(ConsistOf(expected.Env))
			})
		})

		Context("Adds the right volume mounts", func() {
			type VolumeMountCase struct {
				PersonalVolume bool
				MountInfos     []forge.NFSVolumeMountInfo
				ExpectedOutput func(*clv1alpha2.Environment) []corev1.VolumeMount
			}

			WhenBody := func(c VolumeMountCase) func() {
				return func() {
					BeforeEach(func() {
						environment.MountMyDriveVolume = c.PersonalVolume
					})

					JustBeforeEach(func() {
						fmt.Printf("Using mountInfos %v\n", c.MountInfos)
						actual = forge.AppContainer(&environment, forge.PersistentMountPath(&environment), c.MountInfos)
					})

					It("Should set the VolumeMounts accordingly", func() {
						Expect(actual.VolumeMounts).To(Equal(c.ExpectedOutput(&environment)))
					})
				}
			}

			When("The personal volume mount is enabled", WhenBody(VolumeMountCase{
				PersonalVolume: true,
				MountInfos: []forge.NFSVolumeMountInfo{
					mountInfoMyDrive,
				},
				ExpectedOutput: func(e *clv1alpha2.Environment) []corev1.VolumeMount {
					c := corev1.Container{}
					forge.AddContainerVolumeMount(&c, forge.PersistentVolumeName, forge.PersistentMountPath(e))
					forge.AddContainerVolumeMount(&c, forge.MyDriveVolumeName, forge.MyDriveVolumeMountPath)
					return c.VolumeMounts
				},
			}))

			When("The personal volume mount is disabled", WhenBody(VolumeMountCase{
				PersonalVolume: false,
				MountInfos:     nil,
				ExpectedOutput: func(e *clv1alpha2.Environment) []corev1.VolumeMount {
					c := corev1.Container{}
					forge.AddContainerVolumeMount(&c, forge.PersistentVolumeName, forge.PersistentMountPath(e))
					return c.VolumeMounts
				},
			}))

			When("There is a mounted shared volume and the personal volume mount is enabled", WhenBody(VolumeMountCase{
				PersonalVolume: true,
				MountInfos: []forge.NFSVolumeMountInfo{
					mountInfoMyDrive,
					mountInfoShVol,
				},
				ExpectedOutput: func(e *clv1alpha2.Environment) []corev1.VolumeMount {
					c := corev1.Container{}
					forge.AddContainerVolumeMount(&c, forge.PersistentVolumeName, forge.PersistentMountPath(e))
					forge.AddContainerVolumeMount(&c, forge.MyDriveVolumeName, forge.MyDriveVolumeMountPath)
					forge.AddContainerVolumeMount(&c, nfsShVolName, nfsShVolMountPath)
					return c.VolumeMounts
				},
			}))

			When("There is a mounted shared volume and the personal volume mount is disabled", WhenBody(VolumeMountCase{
				PersonalVolume: true,
				MountInfos: []forge.NFSVolumeMountInfo{
					mountInfoShVol,
				},
				ExpectedOutput: func(e *clv1alpha2.Environment) []corev1.VolumeMount {
					c := corev1.Container{}
					forge.AddContainerVolumeMount(&c, forge.PersistentVolumeName, forge.PersistentMountPath(e))
					forge.AddContainerVolumeMount(&c, nfsShVolName, nfsShVolMountPath)
					return c.VolumeMounts
				},
			}))
		})

		Context("Has to handle custom startup options", func() {
			testArguments := []string{"argument=yes", "another-argument"}
			type ContainerCase struct {
				StartupOpts    *clv1alpha2.ContainerStartupOpts
				ExpectedOutput func(*clv1alpha2.Environment) []string
			}

			WhenBody := func(c ContainerCase) func() {
				return func() {
					BeforeEach(func() {
						environment.ContainerStartupOptions = c.StartupOpts
					})

					JustBeforeEach(func() {
						actual = forge.AppContainer(&environment, forge.PersistentMountPath(&environment), mountInfos)
					})

					It("Should return the correct startup args", func() {
						Expect(actual.Args).To(Equal(c.ExpectedOutput(&environment)))
					})
				}
			}

			When("ContainerStartupOptions is nil", WhenBody(ContainerCase{
				StartupOpts:    nil,
				ExpectedOutput: func(_ *clv1alpha2.Environment) []string { return nil },
			}))

			When("startup argument are not set", WhenBody(ContainerCase{
				StartupOpts:    &clv1alpha2.ContainerStartupOpts{},
				ExpectedOutput: func(_ *clv1alpha2.Environment) []string { return nil },
			}))

			When("startup argument are not set", WhenBody(ContainerCase{
				StartupOpts:    &clv1alpha2.ContainerStartupOpts{StartupArgs: testArguments},
				ExpectedOutput: func(_ *clv1alpha2.Environment) []string { return testArguments },
			}))
		})

		Context("Has to handle the workdir", func() {
			type ContainerCase struct {
				StartupOpts    *clv1alpha2.ContainerStartupOpts
				ExpectedOutput func(*clv1alpha2.Environment) string
			}

			WhenBody := func(c ContainerCase) func() {
				return func() {
					BeforeEach(func() {
						environment.ContainerStartupOptions = c.StartupOpts
					})

					JustBeforeEach(func() {
						actual = forge.AppContainer(&environment, forge.PersistentMountPath(&environment), mountInfos)
					})

					It("Should set the WorkingDirectory accordingly", func() {
						Expect(actual.WorkingDir).To(Equal(c.ExpectedOutput(&environment)))
					})
				}
			}

			When("ContainerStartupOptions is nil", WhenBody(ContainerCase{
				StartupOpts:    nil,
				ExpectedOutput: func(_ *clv1alpha2.Environment) string { return "" },
			}))

			When("EnforceWorkdir is set", WhenBody(ContainerCase{
				StartupOpts:    &clv1alpha2.ContainerStartupOpts{EnforceWorkdir: true},
				ExpectedOutput: forge.PersistentMountPath,
			}))
		})
	})

	Describe("The forge.InitContainers function forges the list of init containers for the podSpec", func() {
		var actual []corev1.Container

		JustBeforeEach(func() {
			actual = forge.InitContainers(&instance, &environment, &opts)
		})

		type InitContainersCase struct {
			StartupOpts    *clv1alpha2.ContainerStartupOpts
			ExpectedOutput func(*clv1alpha2.Instance, *clv1alpha2.Environment) []corev1.Container
		}

		WhenBody := func(c InitContainersCase) func() {
			return func() {
				BeforeEach(func() {
					environment.ContainerStartupOptions = c.StartupOpts
				})

				It("Should return the correct volumeSource", func() {
					Expect(actual).To(Equal(c.ExpectedOutput(&instance, &environment)))
				})
			}
		}

		When("ContainerStartupOpts is nil", WhenBody(InitContainersCase{
			StartupOpts: nil,
			ExpectedOutput: func(_ *clv1alpha2.Instance, _ *clv1alpha2.Environment) []corev1.Container {
				return nil
			},
		}))
		When("no archive source is specified", WhenBody(InitContainersCase{
			StartupOpts: &clv1alpha2.ContainerStartupOpts{},
			ExpectedOutput: func(_ *clv1alpha2.Instance, _ *clv1alpha2.Environment) []corev1.Container {
				return nil
			},
		}))
		When("an archive source is specified", WhenBody(InitContainersCase{
			StartupOpts: &clv1alpha2.ContainerStartupOpts{SourceArchiveURL: httpPath},
			ExpectedOutput: func(i *clv1alpha2.Instance, e *clv1alpha2.Environment) []corev1.Container {
				_, val := forge.NeedsInitContainer(i, e)
				return []corev1.Container{forge.ContentDownloaderInitContainer(val, &opts)}
			},
		}))
	})

	Describe("The forge.ContentDownloaderInitContainer function forges the initContainer for volume pre-population", func() {
		const containerName = "content-downloader"
		var actual, expected corev1.Container

		JustBeforeEach(func() {
			actual = forge.ContentDownloaderInitContainer(httpPath, &opts)
		})

		It("Should set the correct container name and image", func() {
			// PodSecurityContext setting is checked by GenericContainer specific tests
			Expect(actual.Name).To(Equal(containerName))
			Expect(actual.Image).To(Equal("cont-dler-img:tag"))
		})
		It("Should set the correct resources", func() {
			forge.SetContainerResources(&expected, 0.5, 1, 256, 1024)
			Expect(actual.Resources).To(Equal(expected.Resources))
		})
		It("Should NOT set container ports", func() {
			Expect(actual.Ports).To(BeEmpty())
		})
		It("Should NOT set readiness probes", func() {
			Expect(actual.ReadinessProbe).To(BeNil())
		})
		It("Should set the volume mount", func() {
			forge.AddContainerVolumeMount(&expected, forge.PersistentVolumeName, forge.PersistentDefaultMountPath)
			Expect(actual.VolumeMounts).To(Equal(expected.VolumeMounts))
		})
		It("Should set the correct environment variables", func() {
			forge.AddEnvVariableToContainer(&expected, "SOURCE_ARCHIVE", httpPath)
			forge.AddEnvVariableToContainer(&expected, "DESTINATION_PATH", forge.PersistentDefaultMountPath)
			Expect(actual.Env).To(ConsistOf(expected.Env))
		})
	})

	Describe("The forge.SubmissionJobSpec function forges the JobSpec for a submission job", func() {
		var actual batchv1.JobSpec

		BeforeEach(func() {
			environment.Persistent = true
			instance.Spec.CustomizationUrls = &clv1alpha2.InstanceCustomizationUrls{
				ContentDestination: httpPath,
			}
		})

		JustBeforeEach(func() {
			actual = forge.SubmissionJobSpec(&instance, &environment, &opts)
		})

		It("should return the correct podSpecification", func() {
			Expect(actual).To(Equal(batchv1.JobSpec{
				BackoffLimit:            ptr.To[int32](forge.SubmissionJobMaxRetries),
				TTLSecondsAfterFinished: ptr.To[int32](forge.SubmissionJobTTLSeconds),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							forge.ContentUploaderJobContainer(httpPath, instance.Name, &opts),
						},
						Volumes:                      forge.ContainerVolumes(&instance, &environment, nil),
						SecurityContext:              forge.PodSecurityContext(),
						AutomountServiceAccountToken: ptr.To(false),
						RestartPolicy:                corev1.RestartPolicyOnFailure,
					},
				},
			}))
		})
	})

	Describe("The forge.ContentUploaderJobContainer function forges a container for content submission", func() {
		const containerName = "content-uploader"
		var actual, expected corev1.Container

		JustBeforeEach(func() {
			actual = forge.ContentUploaderJobContainer(httpPath, instanceName, &opts)
		})

		It("Should set the correct container name and image", func() {
			// PodSecurityContext setting is checked by GenericContainer specific tests
			Expect(actual.Name).To(Equal(containerName))
			Expect(actual.Image).To(Equal("cont-uplr-img:tag"))
		})
		It("Should set the correct resources", func() {
			forge.SetContainerResources(&expected, 0.5, 1, 256, 1024)
			Expect(actual.Resources).To(Equal(expected.Resources))
		})
		It("Should NOT set container ports", func() {
			Expect(actual.Ports).To(BeEmpty())
		})
		It("Should NOT set readiness probes", func() {
			Expect(actual.ReadinessProbe).To(BeNil())
		})
		It("Should set the volume mount", func() {
			forge.AddContainerVolumeMount(&expected, forge.PersistentVolumeName, forge.PersistentDefaultMountPath)
			Expect(actual.VolumeMounts).To(Equal(expected.VolumeMounts))
		})
		It("Should set the correct environment variables", func() {
			forge.AddEnvVariableToContainer(&expected, "SOURCE_PATH", forge.PersistentDefaultMountPath)
			forge.AddEnvVariableToContainer(&expected, "DESTINATION_URL", httpPath)
			forge.AddEnvVariableToContainer(&expected, "FILENAME", instanceName)
			Expect(actual.Env).To(ConsistOf(expected.Env))
		})
	})

	Describe("The forge.GenericContainer function forges a new container", func() {
		JustBeforeEach(func() {
			container = forge.GenericContainer(environment.Name, environment.Image)
		})

		It("Should set the container name", func() {
			Expect(container.Name).To(Equal(environment.Name))
		})
		It("Should set the container image", func() {
			Expect(container.Image).To(Equal(environment.Image))
		})
		It("Should set the container security context", func() {
			Expect(container.SecurityContext).To(Equal(forge.RestrictiveSecurityContext()))
		})
	})

	Describe("The forge.RestrictiveSecurityContext function forges a suitable container security context", func() {
		var secCtx corev1.SecurityContext
		BeforeEach(func() {
			secCtx = *forge.RestrictiveSecurityContext()
		})

		It("Should set the correct capabilities", func() {
			By("Setting a non null value", func() {
				Expect(secCtx.Capabilities).NotTo(BeNil())
			})
			By("Setting drop on ALL", func() {
				Expect(secCtx.Capabilities.Drop).To(ConsistOf([]corev1.Capability{
					corev1.Capability("ALL"),
				}))
			})
			By("Not adding capabilities", func() {
				Expect(secCtx.Capabilities.Add).To(BeEmpty())
			})
		})
		It("Should set not privileged", func() {
			Expect(secCtx.Privileged).To(PointTo(BeFalse()))
		})
		It("Should set not to allow privilege escalation", func() {
			Expect(secCtx.AllowPrivilegeEscalation).To(PointTo(BeFalse()))
		})
	})

	Describe("The forge.AddTCPPortToContainer", func() {
		JustBeforeEach(func() {
			forge.AddTCPPortToContainer(&container, portName, portNum)
		})
		It("Should add a single port with the specified parameters", func() {
			Expect(container.Ports).To(ConsistOf(corev1.ContainerPort{
				Name:          portName,
				ContainerPort: portNum,
				Protocol:      corev1.ProtocolTCP,
			}))
		})
	})

	Describe("The forge.AddEnvVariableToContainer", func() {
		JustBeforeEach(func() {
			forge.AddEnvVariableToContainer(&container, envVarName, envVarVal)
		})
		It("Should add a single env entry with the specified parameters", func() {
			Expect(container.Env).To(ConsistOf(corev1.EnvVar{
				Name:  envVarName,
				Value: envVarVal,
			}))
		})
	})

	Describe("The forge.AddEnvVariableFromResourcesToContainer", func() {
		JustBeforeEach(func() {
			container.Name = envName
			forge.AddEnvVariableFromResourcesToContainer(&container, envVarName, container.Name, corev1.ResourceRequestsCPU, forge.DefaultDivisor)
		})
		It("Should add a single env entry with the specified parameters", func() {
			Expect(container.Env).To(ConsistOf(corev1.EnvVar{
				Name: envVarName,
				ValueFrom: &corev1.EnvVarSource{
					ResourceFieldRef: &corev1.ResourceFieldSelector{
						ContainerName: envName,
						Resource:      envVarResName.String(),
					},
				},
			}))
		})
	})

	Describe("The forge.AddContainerVolumeMount", func() {
		JustBeforeEach(func() {
			forge.AddContainerVolumeMount(&container, volumeName, volumePath)
		})
		It("Should add a single volumeMount entry with the specified parameters", func() {
			Expect(container.VolumeMounts).To(ConsistOf(corev1.VolumeMount{
				Name:      volumeName,
				MountPath: volumePath,
			}))
		})
	})

	Describe("The forge.AddContainerArg", func() {
		JustBeforeEach(func() {
			forge.AddContainerArg(&container, "param", "val")
		})
		It("Should add a single correctly constructed argument", func() {
			Expect(container.Args).To(ConsistOf("--param=val"))
		})
	})

	Describe("The forge.SetContainerReadinessTCPProbe", func() {
		var expected *corev1.Probe
		JustBeforeEach(func() {
			expected = forge.ContainerProbe()
			expected.TCPSocket = &corev1.TCPSocketAction{
				Port: intstr.FromString(portName),
			}
			forge.SetContainerReadinessTCPProbe(&container, portName)
		})

		It("Should correctly set the readiness probe as a tcp probe", func() {
			Expect(container.ReadinessProbe).To(Equal(expected))
		})
	})

	Describe("The forge.SetContainerReadinessHTTPProbe", func() {
		var expected *corev1.Probe
		JustBeforeEach(func() {
			expected = forge.ContainerProbe()
			expected.HTTPGet = &corev1.HTTPGetAction{
				Port: intstr.FromString(portName),
				Path: httpPath,
			}
			forge.SetContainerReadinessHTTPProbe(&container, portName, httpPath)
		})

		It("Should correctly set the readiness probe as an http probe", func() {
			Expect(container.ReadinessProbe).To(Equal(expected))
		})
	})

	Describe("The forge.ContainerProbe function", func() {
		var actual, expected corev1.Probe
		JustBeforeEach(func() {
			actual = *forge.ContainerProbe()
			expected = corev1.Probe{
				InitialDelaySeconds: 10,
				PeriodSeconds:       2,
				SuccessThreshold:    2,
				FailureThreshold:    5,
			}
		})
		It("Should create a general probe with custom values", func() {
			Expect(actual).To(Equal(expected))
		})
	})

	Describe("The forge.SetContainerResources function", func() {
		var memTest1, memTest2 resource.Quantity
		BeforeEach(func() {
			memTest1 = resource.MustParse("3Mi")
			memTest2 = resource.MustParse("1Gi")
			forge.SetContainerResources(&container, 1, 0.2, 3, 1024)
		})
		It("Should correctly set resource of the passed container", func() {
			Expect(container.Resources.Requests.Cpu().MilliValue()).To(BeNumerically("==", 1000))
			Expect(container.Resources.Limits.Cpu().MilliValue()).To(BeNumerically("==", 200))
			Expect(container.Resources.Requests.Memory().Value()).To(Equal((&memTest1).Value()))
			Expect(container.Resources.Limits.Memory().Value()).To(Equal((&memTest2).Value()))
		})
	})

	Describe("The forge.SetContainerResourcesFromEnvironment function", func() {
		JustBeforeEach(func() {
			forge.SetContainerResourcesFromEnvironment(&container, &environment)
		})
		It("Should correctly set resource of the passed container", func() {
			Expect(container.Resources.Requests.Cpu().MilliValue()).To(BeNumerically("==", expectedCPUReqMillis))
			Expect(container.Resources.Limits.Cpu().MilliValue()).To(BeNumerically("==", expectedCPULimMillis))
			Expect(container.Resources.Requests.Memory()).To(PointTo(Equal(environment.Resources.Memory)))
			Expect(container.Resources.Limits.Memory()).To(PointTo(Equal(environment.Resources.Memory)))
		})
	})

	Describe("The forge.ContainerVolumes function", func() {
		var actual []corev1.Volume

		type ContainerVolumesCase struct {
			Persistent          bool
			MountPersonalVolume bool
			MountInfos          []forge.NFSVolumeMountInfo
			Mode                clv1alpha2.EnvironmentMode
			StartupOpts         *clv1alpha2.ContainerStartupOpts
			ExpectedOutputVSs   func(*clv1alpha2.Environment) []corev1.Volume
		}

		WhenBody := func(c ContainerVolumesCase) func() {
			return func() {
				BeforeEach(func() {
					environment.Persistent = c.Persistent
					environment.ContainerStartupOptions = c.StartupOpts
					environment.Mode = c.Mode
					environment.MountMyDriveVolume = c.MountPersonalVolume
				})

				JustBeforeEach(func() {
					actual = forge.ContainerVolumes(&instance, &environment, c.MountInfos)
				})

				It("Should return the correct volumeSource", func() {
					Expect(actual).To(Equal(c.ExpectedOutputVSs(&environment)))
				})
			}
		}

		When("the environment is not persistent and mode is standard", WhenBody(ContainerVolumesCase{
			Persistent: false,
			Mode:       clv1alpha2.ModeStandard,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume(forge.PersistentVolumeName, instanceName, e)}
			},
		}))

		When("the environment is not persistent and mode is exam", WhenBody(ContainerVolumesCase{
			Persistent: false,
			Mode:       clv1alpha2.ModeExam,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume(forge.PersistentVolumeName, instanceName, e)}
			},
		}))

		When("the environment is not persistent and mode is exercise", WhenBody(ContainerVolumesCase{
			Persistent: false,
			Mode:       clv1alpha2.ModeExercise,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume(forge.PersistentVolumeName, instanceName, e)}
			},
		}))

		When("the environment is persistent and mode is standard", WhenBody(ContainerVolumesCase{
			Persistent: true,
			Mode:       clv1alpha2.ModeStandard,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume(forge.PersistentVolumeName, instanceName, e)}
			},
		}))

		When("the environment is persistent and mode is exam", WhenBody(ContainerVolumesCase{
			Persistent: true,
			Mode:       clv1alpha2.ModeExam,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume(forge.PersistentVolumeName, instanceName, e)}
			},
		}))

		When("the environment is persistent and mode is exercise", WhenBody(ContainerVolumesCase{
			Persistent: true,
			Mode:       clv1alpha2.ModeExercise,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume(forge.PersistentVolumeName, instanceName, e)}
			},
		}))

		When("the environment has the source archive url option", WhenBody(ContainerVolumesCase{
			StartupOpts: &clv1alpha2.ContainerStartupOpts{SourceArchiveURL: httpPath},
			Persistent:  false,
			Mode:        clv1alpha2.ModeExam,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume(forge.PersistentVolumeName, instanceName, e)}
			},
		}))

		When("the environment has the mount personal volume option", WhenBody(ContainerVolumesCase{
			MountPersonalVolume: true,
			MountInfos: []forge.NFSVolumeMountInfo{
				forge.MyDriveNFSVolumeMountInfo(nfsServerName, nfsMyDriveExpPath),
			},
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{
					forge.ContainerVolume(forge.PersistentVolumeName, instanceName, e),
					forge.NFSVolume(mountInfos[0]),
				}
			},
		}))

		When("the environment has a mounted shared volume and the personal volume", WhenBody(ContainerVolumesCase{
			MountPersonalVolume: true,
			MountInfos: []forge.NFSVolumeMountInfo{
				forge.MyDriveNFSVolumeMountInfo(nfsServerName, nfsMyDriveExpPath),
				{
					VolumeName:    nfsShVolName,
					ServerAddress: nfsServerName,
					ExportPath:    nfsShVolExpPath,
					MountPath:     nfsShVolMountPath,
					ReadOnly:      nfsShVolReadOnly,
				},
			},
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{
					forge.ContainerVolume(forge.PersistentVolumeName, instanceName, e),
					forge.NFSVolume(mountInfos[0]),
					forge.NFSVolume(mountInfos[1]),
				}
			},
		}))
	})

	Describe("The forge.ContainerVolume function", func() {
		var actual corev1.Volume

		JustBeforeEach(func() {
			actual = forge.ContainerVolume(volumeName, claimName, &environment)
		})

		type ContainerVolumeCase struct {
			Persistent       bool
			ExpectedOutputVS corev1.VolumeSource
		}

		WhenBody := func(c ContainerVolumeCase) func() {
			return func() {
				BeforeEach(func() {
					environment.Persistent = c.Persistent
				})

				It("Should return the correct volumeSource", func() {
					Expect(actual.VolumeSource).To(Equal(c.ExpectedOutputVS))
				})

				It("Should set the correct volume name", func() {
					Expect(actual.Name).To(Equal(volumeName))
				})
			}
		}

		When("the environment is not persistent", WhenBody(ContainerVolumeCase{
			Persistent: false,
			ExpectedOutputVS: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}))

		When("the environment is persistent", WhenBody(ContainerVolumeCase{
			Persistent: true,
			ExpectedOutputVS: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claimName,
				},
			},
		}))
	})

	Describe("The forge.NeedsInitContainer function", func() {
		type NeedsInitContainerCase struct {
			StartupOpts          *clv1alpha2.ContainerStartupOpts
			InstCustomOpts       *clv1alpha2.InstanceCustomizationUrls
			ExpectedOutputVal    bool
			ExpectedOutputOrigin string
		}

		WhenBody := func(c NeedsInitContainerCase) func() {
			return func() {
				BeforeEach(func() {
					environment.ContainerStartupOptions = c.StartupOpts
					instance.Spec.CustomizationUrls = c.InstCustomOpts
				})

				It("Should return the correct values", func() {
					val, origin := forge.NeedsInitContainer(&instance, &environment)
					Expect(val).To(Equal(c.ExpectedOutputVal))
					Expect(origin).To(Equal(c.ExpectedOutputOrigin))
				})
			}
		}

		When("No origin is provided", WhenBody(NeedsInitContainerCase{
			StartupOpts:          nil,
			InstCustomOpts:       nil,
			ExpectedOutputVal:    false,
			ExpectedOutputOrigin: "",
		}))
		Context("no instance custom options are provided", func() {
			When("no source archive is specified in the template", WhenBody(NeedsInitContainerCase{
				StartupOpts:          &clv1alpha2.ContainerStartupOpts{},
				InstCustomOpts:       nil,
				ExpectedOutputVal:    false,
				ExpectedOutputOrigin: "",
			}))
			When("a source archive is specified in the template", WhenBody(NeedsInitContainerCase{
				StartupOpts:          &clv1alpha2.ContainerStartupOpts{SourceArchiveURL: httpPath},
				InstCustomOpts:       nil,
				ExpectedOutputVal:    true,
				ExpectedOutputOrigin: httpPath,
			}))
		})
		Context("no template custom options are provied", func() {
			When("no source archive is specified in the instance", WhenBody(NeedsInitContainerCase{
				StartupOpts:          nil,
				InstCustomOpts:       &clv1alpha2.InstanceCustomizationUrls{},
				ExpectedOutputVal:    false,
				ExpectedOutputOrigin: "",
			}))
			When("a source archive is specified in the instance", WhenBody(NeedsInitContainerCase{
				StartupOpts:          nil,
				InstCustomOpts:       &clv1alpha2.InstanceCustomizationUrls{ContentOrigin: httpPath},
				ExpectedOutputVal:    true,
				ExpectedOutputOrigin: httpPath,
			}))
		})
		When("both template and instance custom options are provided", WhenBody(NeedsInitContainerCase{
			StartupOpts:          &clv1alpha2.ContainerStartupOpts{SourceArchiveURL: httpPath},
			InstCustomOpts:       &clv1alpha2.InstanceCustomizationUrls{ContentOrigin: httpPathAlternative},
			ExpectedOutputVal:    true,
			ExpectedOutputOrigin: httpPathAlternative,
		}))
	})

	Describe("The forge.PersistentMountPath function", func() {
		type MyDriveMountPathCase struct {
			StartupOpts    *clv1alpha2.ContainerStartupOpts
			ExpectedOutput string
		}

		WhenBody := func(c MyDriveMountPathCase) func() {
			return func() {
				BeforeEach(func() {
					environment.ContainerStartupOptions = c.StartupOpts
				})

				It("Should return the correct value", func() {
					Expect(forge.PersistentMountPath(&environment)).To(Equal(c.ExpectedOutput))
				})
			}
		}

		When("ContainerStartupOpts is nil", WhenBody(MyDriveMountPathCase{
			StartupOpts:    nil,
			ExpectedOutput: forge.PersistentDefaultMountPath,
		}))
		When("no content path is specified", WhenBody(MyDriveMountPathCase{
			StartupOpts:    &clv1alpha2.ContainerStartupOpts{},
			ExpectedOutput: forge.PersistentDefaultMountPath,
		}))
		When("content path is specified", WhenBody(MyDriveMountPathCase{
			StartupOpts:    &clv1alpha2.ContainerStartupOpts{ContentPath: volumePath},
			ExpectedOutput: volumePath,
		}))
	})

	Describe("The forge.InstanceHostname function", func() {
		var actual string

		JustBeforeEach(func() {
			actual = forge.InstanceHostname(&environment)
		})

		type EnvModeCase struct {
			EnvMode        clv1alpha2.EnvironmentMode
			ExpectedOutput string
		}

		WhenBody := func(c EnvModeCase) func() {
			return func() {
				BeforeEach(func() {
					environment.Mode = c.EnvMode
				})

				It("Should return the correct hostname", func() {
					Expect(actual).To(Equal(c.ExpectedOutput))
				})
			}
		}

		When("the environment mode is Exercise", WhenBody(EnvModeCase{
			EnvMode:        clv1alpha2.ModeExercise,
			ExpectedOutput: "exercise",
		}))

		When("the environment mode is Exam", WhenBody(EnvModeCase{
			EnvMode:        clv1alpha2.ModeExam,
			ExpectedOutput: "exam",
		}))

		When("the environment mode is Standard", WhenBody(EnvModeCase{
			EnvMode:        clv1alpha2.ModeStandard,
			ExpectedOutput: "",
		}))
	})

	Describe("The forge.NodeSelectorLabels function", func() {
		type NodeSelectorLabelsCase struct {
			TemplateLabelSelector *map[string]string
			InstanceLabelSelector map[string]string
			ExpectedOutput        map[string]string
		}

		WhenBody := func(c NodeSelectorLabelsCase, desc string) func() {
			return func() {
				BeforeEach(func() {
					environment.NodeSelector = c.TemplateLabelSelector
					instance.Spec.NodeSelector = c.InstanceLabelSelector
				})
				It("Should return the right set of labels: "+desc, func() {
					Expect(forge.NodeSelectorLabels(&instance, &environment)).To(Equal(c.ExpectedOutput))
				})
			}
		}

		Context("TemplateLabelSelector is nil", func() {
			When("InstanceLabelSelector is nil", WhenBody(NodeSelectorLabelsCase{
				TemplateLabelSelector: nil,
				InstanceLabelSelector: nil,
				ExpectedOutput:        map[string]string{},
			}, "feature disabled"))
			When("InstanceLabelSelector is present", WhenBody(NodeSelectorLabelsCase{
				TemplateLabelSelector: nil,
				InstanceLabelSelector: map[string]string{"key": "value"},
				ExpectedOutput:        map[string]string{},
			}, "feature disabled"))
		})

		Context("TemplateLabelSelector is present", func() {
			When("TemplateLabelSelector is empty and InstanceLabelSelector is nil", WhenBody(NodeSelectorLabelsCase{
				TemplateLabelSelector: &map[string]string{},
				InstanceLabelSelector: nil,
				ExpectedOutput:        nil,
			}, "template and instance labels empty"))
			When("TemplateLabelSelector is empty and InstanceLabelSelector is present", WhenBody(NodeSelectorLabelsCase{
				TemplateLabelSelector: &map[string]string{},
				InstanceLabelSelector: map[string]string{"key": "value"},
				ExpectedOutput:        map[string]string{"key": "value"},
			}, "instance labels are chosen over empty template labels"))
			When("TemplateLabelSelector is present and InstanceLabelSelector is present", WhenBody(NodeSelectorLabelsCase{
				TemplateLabelSelector: &map[string]string{"templateKey": "templateValue"},
				InstanceLabelSelector: map[string]string{"instanceKey": "instanceValue"},
				ExpectedOutput:        map[string]string{"templateKey": "templateValue"},
			}, "template labels are chosen over instance labels"))
		})
	})
})
