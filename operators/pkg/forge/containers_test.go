// Copyright 2020-2021 Politecnico di Torino
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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Containers and Deployment spec forging", func() {

	var (
		instance    clv1alpha2.Instance
		environment clv1alpha2.Environment
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
	)

	// test aware constants (are tested against code)
	const (
		myDriveMountPath = "/mydrive"
		myDriveName      = "mydrive"
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
		opts = forge.ContainerEnvOpts{
			ImagesTag:     "tag",
			XVncImg:       "x-vnc-img",
			WebsockifyImg: "wsfy-img",
			MyDriveImg:    "fb-img",
			MyDriveImgTag: "fb-tag",
		}
		container = corev1.Container{}
	})

	Describe("The forge.PVCSpec function", func() {
		var spec corev1.PersistentVolumeClaimSpec

		JustBeforeEach(func() {
			spec = forge.PVCSpec(&environment)
		})

		It("Should set the correct access mode", func() {
			Expect(spec.AccessModes).To(Equal([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}))
		})
		It("Should set the correct resources", func() {
			Expect(spec.Resources.Requests.Storage()).To(PointTo(Equal(environment.Resources.Disk)))
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
		It("Should set the correct RunAsNonRoot", func() {
			Expect(psc.RunAsNonRoot).To(PointTo(BeTrue()))
		})
	})

	Describe("The forge.ReplicasCount function", func() {
		type ReplicasCountCase struct {
			Persistent     bool
			Running        bool
			ExpectedOutput int32
		}

		WhenBody := func(c ReplicasCountCase) func() {
			return func() {
				JustBeforeEach(func() {
					instance.Spec.Running = c.Running
					environment.Persistent = c.Persistent
				})

				It("Should return the correct Replica count", func() {
					Expect(forge.ReplicasCount(&instance, &environment)).To(PointTo(Equal(c.ExpectedOutput)))
				})
			}
		}

		When("the environment is persistent and instance is not running", WhenBody(ReplicasCountCase{
			Persistent:     false,
			Running:        false,
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
			spec = forge.DeploymentSpec(&instance, &environment, &opts)
		})

		It("Should set the correct template labels", func() {
			Expect(spec.Template.ObjectMeta.GetLabels()).To(Equal(forge.InstanceSelectorLabels(&instance)))
		})
		It("Should set the correct template spec", func() {
			Expect(spec.Template.Spec).To(Equal(forge.PodSpec(&instance, &environment, &opts)))
		})

		It("Should set the correct selector", func() {
			Expect(spec.Selector.MatchLabels).To(Equal(forge.InstanceSelectorLabels(&instance)))
		})
	})

	Describe("The forge.PodSpec function", func() {
		var spec corev1.PodSpec

		type PodSpecContainersCase struct {
			Mode           clv1alpha2.EnvironmentMode
			ExpectedOutput func(*clv1alpha2.Instance, *clv1alpha2.Environment) []corev1.Container
		}

		ContainersWhenBody := func(psc PodSpecContainersCase) func() {
			return func() {
				BeforeEach(func() {
					environment.Mode = psc.Mode
				})
				It("Should set the correct containers", func() {
					Expect(spec.Containers).To(ConsistOf(psc.ExpectedOutput(&instance, &environment)))
				})
			}
		}

		JustBeforeEach(func() {
			spec = forge.PodSpec(&instance, &environment, &opts)
		})

		It("Should set the security context", func() {
			Expect(spec.SecurityContext).To(Equal(forge.PodSecurityContext()))
		})

		It("Should disable the service account token automount", func() {
			Expect(spec.AutomountServiceAccountToken).To(PointTo(BeFalse()))
		})

		When("the environment mode is Standard", ContainersWhenBody(PodSpecContainersCase{
			Mode: clv1alpha2.ModeStandard,
			ExpectedOutput: func(i *clv1alpha2.Instance, e *clv1alpha2.Environment) []corev1.Container {
				return []corev1.Container{
					forge.WebsockifyContainer(&opts),
					forge.XVncContainer(&opts),
					forge.MyDriveContainer(i, &opts, myDriveMountPath),
					forge.AppContainer(e, myDriveMountPath),
				}
			},
		}))

		When("the environment mode is Exercise", ContainersWhenBody(PodSpecContainersCase{
			Mode: clv1alpha2.ModeExercise,
			ExpectedOutput: func(i *clv1alpha2.Instance, e *clv1alpha2.Environment) []corev1.Container {
				return []corev1.Container{
					forge.WebsockifyContainer(&opts),
					forge.XVncContainer(&opts),
					forge.AppContainer(e, myDriveMountPath),
				}
			},
		}))

		When("the environment mode is Exam", ContainersWhenBody(PodSpecContainersCase{
			Mode: clv1alpha2.ModeExam,
			ExpectedOutput: func(i *clv1alpha2.Instance, e *clv1alpha2.Environment) []corev1.Container {
				return []corev1.Container{
					forge.WebsockifyContainer(&opts),
					forge.XVncContainer(&opts),
					forge.AppContainer(e, myDriveMountPath),
				}
			},
		}))
	})

	Describe("The forge.WebsockifyContainer function forges a websockify sidecar container", func() {
		var actual, expected corev1.Container
		JustBeforeEach(func() {
			actual = forge.WebsockifyContainer(&opts)
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
			forge.AddTCPPortToContainer(&expected, "vnc", 6080)
			Expect(actual.Ports).To(Equal(expected.Ports))
		})
		It("Should set the env variable", func() {
			forge.AddEnvVariableToContainer(&expected, "WS_PORT", "6080")
			Expect(actual.Env).To(Equal(expected.Env))
		})
		It("Should set the readiness probe", func() {
			forge.SetContainerReadinessTCPProbe(&expected, "vnc")
			Expect(actual.ReadinessProbe).To(Equal(expected.ReadinessProbe))
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
			forge.SetContainerResources(&expected, 0.2, 0.5, 200, 600)
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

	Describe("The forge.MyDriveContainer function forges a mydrive sidecar container", func() {
		var actual, expected corev1.Container
		JustBeforeEach(func() {
			actual = forge.MyDriveContainer(&instance, &opts, myDriveMountPath)
		})

		It("Should set the correct container name and image", func() {
			// PodSecurityContext setting is checked by GenericContainer specific tests
			Expect(actual.Name).To(Equal(myDriveName))
			Expect(actual.Image).To(Equal(opts.MyDriveImg + ":" + opts.MyDriveImgTag))
		})
		It("Should set the correct resources", func() {
			forge.SetContainerResources(&expected, 0.01, 0.25, 100, 500)
			Expect(actual.Resources).To(Equal(expected.Resources))
		})
		It("Should set the tcp port exposition", func() {
			forge.AddTCPPortToContainer(&expected, myDriveName, 8080)
			Expect(actual.Ports).To(Equal(expected.Ports))
		})
		It("Should set the readiness probe", func() {
			forge.SetContainerReadinessHTTPProbe(&expected, myDriveName, "/healthz")
			Expect(actual.ReadinessProbe).To(Equal(expected.ReadinessProbe))
		})
		It("Should set the volume mount", func() {
			forge.AddContainerVolumeMount(&expected, myDriveName, myDriveMountPath)
			Expect(actual.VolumeMounts).To(Equal(expected.VolumeMounts))
		})
		It("Should set the launch arguments", func() {
			forge.AddContainerArg(&expected, "port", "8080")
			forge.AddContainerArg(&expected, "root", myDriveMountPath)
			forge.AddContainerArg(&expected, "noauth", "true")
			forge.AddContainerArg(&expected, "database", "/tmp/database.db")
			forge.AddContainerArg(&expected, "baseurl", forge.IngressMyDrivePath(&instance))
			Expect(actual.Args).To(Equal(expected.Args))
		})
	})

	Describe("The forge.AppContainer function forges the main application container", func() {
		var actual, expected corev1.Container
		JustBeforeEach(func() {
			actual = forge.AppContainer(&environment, myDriveMountPath)
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
		It("Should set the volume mount", func() {
			forge.AddContainerVolumeMount(&expected, myDriveName, myDriveMountPath)
			Expect(actual.VolumeMounts).To(Equal(expected.VolumeMounts))
		})
		It("Should set the env varibles", func() {
			expected.Name = envName
			forge.AddEnvVariableFromResourcesToContainer(&expected, "CROWNLABS_CPU_REQUESTS", corev1.ResourceRequestsCPU)
			forge.AddEnvVariableFromResourcesToContainer(&expected, "CROWNLABS_CPU_LIMITS", corev1.ResourceLimitsCPU)
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
			forge.AddEnvVariableFromResourcesToContainer(&container, envVarName, corev1.ResourceRequestsCPU)
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
			expected.Handler.TCPSocket = &corev1.TCPSocketAction{
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
			expected.Handler.HTTPGet = &corev1.HTTPGetAction{
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

		JustBeforeEach(func() {
			actual = forge.ContainerVolumes(&instance, &environment)
		})

		type ContainerVolumesCase struct {
			Persistent        bool
			Mode              clv1alpha2.EnvironmentMode
			ExpectedOutputVSs func(*clv1alpha2.Environment) []corev1.Volume
		}

		WhenBody := func(c ContainerVolumesCase) func() {
			return func() {
				BeforeEach(func() {
					environment.Persistent = c.Persistent
					environment.Mode = c.Mode
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
				return []corev1.Volume{forge.ContainerVolume("mydrive", instanceName, e)}
			},
		}))

		When("the environment is not persistent and mode is exam", WhenBody(ContainerVolumesCase{
			Persistent:        false,
			Mode:              clv1alpha2.ModeExam,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume { return nil },
		}))

		When("the environment is not persistent and mode is exercise", WhenBody(ContainerVolumesCase{
			Persistent:        false,
			Mode:              clv1alpha2.ModeExercise,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume { return nil },
		}))

		When("the environment is persistent and mode is standard", WhenBody(ContainerVolumesCase{
			Persistent: true,
			Mode:       clv1alpha2.ModeStandard,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume("mydrive", instanceName, e)}
			},
		}))

		When("the environment is persistent and mode is exam", WhenBody(ContainerVolumesCase{
			Persistent: true,
			Mode:       clv1alpha2.ModeExam,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume("mydrive", instanceName, e)}
			},
		}))

		When("the environment is persistent and mode is exercise", WhenBody(ContainerVolumesCase{
			Persistent: true,
			Mode:       clv1alpha2.ModeExercise,
			ExpectedOutputVSs: func(e *clv1alpha2.Environment) []corev1.Volume {
				return []corev1.Volume{forge.ContainerVolume("mydrive", instanceName, e)}
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

})
