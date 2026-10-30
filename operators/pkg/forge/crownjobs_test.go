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
	"errors"
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

var _ = Describe("CrownJobs forging", func() {
	const (
		crownJobName  = "train-1"
		namespace     = "tenant-student"
		workspaceName = "ws"
		envName       = "app"
		image         = "harbor.example.com/standalone:v1"
		gpu           = "nvidia.com/gpu"
	)

	// expectCannotStart checks that the error is a CannotStartError with the given message.
	expectCannotStart := func(err error, message string) {
		var cannotStart *forge.CannotStartError
		ExpectWithOffset(1, errors.As(err, &cannotStart)).To(BeTrue(), "error %v is not a CannotStartError", err)
		ExpectWithOffset(1, err.Error()).To(Equal(message))
	}

	findContainer := func(containers []corev1.Container, name string) *corev1.Container {
		for i := range containers {
			if containers[i].Name == name {
				return &containers[i]
			}
		}
		return nil
	}

	findEnv := func(c *corev1.Container, name string) []corev1.EnvVar {
		var found []corev1.EnvVar
		for _, env := range c.Env {
			if env.Name == name {
				found = append(found, env)
			}
		}
		return found
	}

	Describe("The forge.CrownJobBatchJob function", func() {
		var (
			crownJob    clv1alpha2.CrownJob
			instance    clv1alpha2.Instance
			template    clv1alpha2.Template
			environment clv1alpha2.Environment
			podTemplate corev1.PodTemplateSpec
			opts        forge.CrownJobJobOpts
			job         *batchv1.Job
			err         error
		)

		BeforeEach(func() {
			crownJob = clv1alpha2.CrownJob{
				ObjectMeta: metav1.ObjectMeta{Name: crownJobName, Namespace: namespace, UID: "cj-uid"},
				Spec:       clv1alpha2.CrownJobSpec{Instance: clv1alpha2.GenericRef{Name: "inst"}},
			}
			environment = clv1alpha2.Environment{
				Name:            envName,
				Image:           image,
				EnvironmentType: clv1alpha2.ClassStandalone,
				Persistent:      true,
				Resources: clv1alpha2.EnvironmentResources{
					ResourceSpec: apicommon.ResourceSpec{
						CPU:    2,
						Memory: resource.MustParse("2Gi"),
						Disk:   resource.MustParse("10Gi"),
					},
					ReservedCPUPercentage: 50,
				},
				ContainerStartupOptions: &clv1alpha2.ContainerStartupOpts{
					StartupArgs:    []string{"--serve"},
					ContentPath:    "/home/user/work",
					EnforceWorkdir: false,
				},
			}
			template = clv1alpha2.Template{
				ObjectMeta: metav1.ObjectMeta{Name: "tpl", Namespace: "workspace-ws"},
				Spec: clv1alpha2.TemplateSpec{
					WorkspaceRef:    clv1alpha2.GenericRef{Name: workspaceName},
					EnvironmentList: []clv1alpha2.Environment{environment},
					NodeSelector:    &map[string]string{"zone": "a"},
				},
			}
			instance = clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{Name: "inst", Namespace: namespace},
				Spec: clv1alpha2.InstanceSpec{
					Template: clv1alpha2.GenericRef{Name: "tpl", Namespace: "workspace-ws"},
					Tenant:   clv1alpha2.GenericRef{Name: "student"},
					Running:  true,
					// A content origin makes forge add the content-downloader init container.
					ContentUrls: map[string]clv1alpha2.InstanceContentUrls{envName: {Origin: "https://example.com/a.zip"}},
				},
			}
			mountInfos := []corev1.VolumeMount{{Name: "shvol", MountPath: "/mnt/shared", ReadOnly: true}}
			podTemplate = forge.DeploymentSpec(&instance, &template, &environment, mountInfos,
				&forge.ContainerEnvOpts{ImagesTag: "v1", ContentToolsImg: "tools"}).Template
			// Another container, as the instance pod of another version could hold.
			podTemplate.Spec.Containers = append(podTemplate.Spec.Containers, corev1.Container{Name: "sidecar", Image: "sidecar:v1"})
			// A value the job must override.
			forge.AddEnvVariableToContainer(&podTemplate.Spec.Containers[0], forge.CrownJobEnvJob, "false")
			// A label that is not of CrownLabs, kept as it is.
			podTemplate.Labels["app.kubernetes.io/name"] = "x"

			opts = forge.CrownJobJobOpts{
				Workspace:       workspaceName,
				EnvironmentName: envName,
				PodTemplate:     &podTemplate,
				CommandPolicy:   clv1alpha2.JobCommandEditable,
				Command:         []string{"python", "train.py"},
				MaxRunTime:      6 * time.Hour,
			}
		})

		JustBeforeEach(func() {
			job, err = forge.CrownJobBatchJob(&crownJob, &opts)
		})

		Context("With an Editable command and no other resources", func() {
			It("Should not fail and should not modify the pod template of the instance", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(podTemplate.Spec.Containers).To(HaveLen(2))
				Expect(podTemplate.Spec.InitContainers).To(HaveLen(1))
				Expect(podTemplate.Labels).To(HaveKey(forge.LabelInstanceKey))
			})

			It("Should set name, namespace and the controller ownerReference to the CrownJob", func() {
				Expect(job.Name).To(Equal(crownJobName))
				Expect(job.Namespace).To(Equal(namespace))
				Expect(job.OwnerReferences).To(ConsistOf(metav1.OwnerReference{
					APIVersion: clv1alpha2.GroupVersion.String(), Kind: "CrownJob", Name: crownJobName, UID: "cj-uid",
					Controller: ptr.To(true), BlockOwnerDeletion: ptr.To(true),
				}))
			})

			It("Should set the Kueue queue label and the crownjob label, and no priority-class label", func() {
				Expect(job.Labels).To(Equal(map[string]string{
					forge.KueueQueueNameLabel: "crownlabs-ws",
					forge.LabelCrownJobKey:    crownJobName,
				}))
			})

			It("Should rewrite every crownlabs.polito.it/ pod label into job.crownlabs.polito.it/, and add crownjob and workspace", func() {
				Expect(job.Spec.Template.Labels).To(Equal(map[string]string{
					"job.crownlabs.polito.it/instance":    "inst",
					"job.crownlabs.polito.it/environment": envName,
					"job.crownlabs.polito.it/template":    "tpl",
					"job.crownlabs.polito.it/tenant":      "student",
					"job.crownlabs.polito.it/crownjob":    crownJobName,
					"job.crownlabs.polito.it/workspace":   workspaceName,
					"app.kubernetes.io/name":              "x",
				}))
				for key := range job.Spec.Template.Labels {
					Expect(key).ToNot(HavePrefix("crownlabs.polito.it/"))
				}
			})

			It("Should keep only the app container", func() {
				Expect(job.Spec.Template.Spec.Containers).To(HaveLen(1))
				Expect(job.Spec.Template.Spec.Containers[0].Name).To(Equal(envName))
				Expect(job.Spec.Template.Spec.Containers[0].Image).To(Equal(image))
			})

			It("Should drop every instance init container, and add only the run-folder one", func() {
				Expect(job.Spec.Template.Spec.InitContainers).To(HaveLen(1))
				Expect(job.Spec.Template.Spec.InitContainers[0].Name).To(Equal(forge.CrownJobRunFolderContainerName))
				Expect(findContainer(job.Spec.Template.Spec.InitContainers, forge.ContentDownloaderName)).To(BeNil())
			})

			It("Should forge the run-folder init container with the app image, mkdir, the persistent mount and the app security context", func() {
				app := job.Spec.Template.Spec.Containers[0]
				runFolder := job.Spec.Template.Spec.InitContainers[0]
				Expect(runFolder.Image).To(Equal(app.Image))
				Expect(runFolder.Command).To(Equal([]string{"mkdir", "-p", "/home/user/work/runs/" + crownJobName}))
				Expect(runFolder.Args).To(BeEmpty())
				Expect(runFolder.VolumeMounts).To(ConsistOf(corev1.VolumeMount{Name: forge.PersistentVolumeName, MountPath: "/home/user/work"}))
				Expect(runFolder.SecurityContext).To(Equal(app.SecurityContext))
				Expect(runFolder.Resources.Requests.Cpu().String()).To(Equal("10m"))
				Expect(runFolder.Resources.Limits.Cpu().String()).To(Equal("100m"))
				Expect(runFolder.Resources.Requests.Memory().String()).To(Equal("16Mi"))
				Expect(runFolder.Resources.Limits.Memory().String()).To(Equal("64Mi"))
			})

			It("Should set the command, clear the args and set the working directory to the persistent mount path", func() {
				app := job.Spec.Template.Spec.Containers[0]
				Expect(podTemplate.Spec.Containers[0].Args).To(Equal([]string{"--serve"}))
				Expect(app.Command).To(Equal([]string{"python", "train.py"}))
				Expect(app.Args).To(BeNil())
				Expect(app.WorkingDir).To(Equal("/home/user/work"))
			})

			It("Should remove every probe and every port", func() {
				Expect(podTemplate.Spec.Containers[0].ReadinessProbe).ToNot(BeNil())
				Expect(podTemplate.Spec.Containers[0].Ports).ToNot(BeEmpty())
				app := job.Spec.Template.Spec.Containers[0]
				Expect(app.ReadinessProbe).To(BeNil())
				Expect(app.LivenessProbe).To(BeNil())
				Expect(app.StartupProbe).To(BeNil())
				Expect(app.Ports).To(BeEmpty())
			})

			It("Should set the four job environment variables once each, overriding a value already present", func() {
				app := &job.Spec.Template.Spec.Containers[0]
				Expect(findEnv(app, forge.CrownJobEnvJob)).To(ConsistOf(corev1.EnvVar{Name: forge.CrownJobEnvJob, Value: "true"}))
				Expect(findEnv(app, forge.CrownJobEnvJobName)).To(ConsistOf(corev1.EnvVar{Name: forge.CrownJobEnvJobName, Value: crownJobName}))
				Expect(findEnv(app, forge.CrownJobEnvGPUs)).To(ConsistOf(corev1.EnvVar{Name: forge.CrownJobEnvGPUs, Value: "0"}))
				Expect(findEnv(app, forge.CrownJobEnvRunFolder)).To(ConsistOf(corev1.EnvVar{
					Name: forge.CrownJobEnvRunFolder, Value: "/home/user/work/runs/" + crownJobName}))
				// The variables of the instance are kept.
				Expect(findEnv(app, "CROWNLABS_CPU_LIMITS")).To(HaveLen(1))
			})

			It("Should keep the resources of the environment", func() {
				app := job.Spec.Template.Spec.Containers[0]
				Expect(app.Resources).To(Equal(podTemplate.Spec.Containers[0].Resources))
				Expect(app.Resources.Requests.Cpu().String()).To(Equal("1"))
			})

			It("Should set restartPolicy Never and the pod activeDeadlineSeconds to maxRunTime", func() {
				Expect(job.Spec.Template.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
				Expect(job.Spec.Template.Spec.ActiveDeadlineSeconds).To(Equal(ptr.To[int64](6 * 3600)))
			})

			It("Should set the Job activeDeadlineSeconds to maxRunTime + 600 s", func() {
				Expect(job.Spec.ActiveDeadlineSeconds).To(Equal(ptr.To[int64](6*3600 + 600)))
			})

			It("Should create the Job suspended, with backoffLimit 0", func() {
				Expect(job.Spec.Suspend).To(Equal(ptr.To(true)))
				Expect(job.Spec.BackoffLimit).To(Equal(ptr.To[int32](0)))
			})

			It("Should set no ttlSecondsAfterFinished when no TTL is given", func() {
				Expect(job.Spec.TTLSecondsAfterFinished).To(BeNil())
			})

			It("Should set the podFailurePolicy: Ignore on DisruptionTarget first, then FailJob on app exit codes NotIn [0]", func() {
				Expect(job.Spec.PodFailurePolicy).To(Equal(&batchv1.PodFailurePolicy{Rules: []batchv1.PodFailurePolicyRule{
					{
						Action: batchv1.PodFailurePolicyActionIgnore,
						OnPodConditions: []batchv1.PodFailurePolicyOnPodConditionsPattern{{
							Type: corev1.DisruptionTarget, Status: corev1.ConditionTrue,
						}},
					},
					{
						Action: batchv1.PodFailurePolicyActionFailJob,
						OnExitCodes: &batchv1.PodFailurePolicyOnExitCodesRequirement{
							ContainerName: ptr.To(envName),
							Operator:      batchv1.PodFailurePolicyOnExitCodesOpNotIn,
							Values:        []int32{0},
						},
					},
				}}))
			})

			It("Should keep volumes, nodeSelector, security context, service-account settings and grace period", func() {
				spec := job.Spec.Template.Spec
				Expect(spec.Volumes).To(Equal(podTemplate.Spec.Volumes))
				Expect(spec.Volumes).To(HaveLen(2))
				Expect(spec.NodeSelector).To(Equal(map[string]string{"zone": "a"}))
				Expect(spec.SecurityContext).To(Equal(forge.PodSecurityContext()))
				Expect(spec.AutomountServiceAccountToken).To(Equal(ptr.To(false)))
				Expect(spec.EnableServiceLinks).To(Equal(ptr.To(false)))
				Expect(spec.TerminationGracePeriodSeconds).To(Equal(podTemplate.Spec.TerminationGracePeriodSeconds))
				Expect(spec.RuntimeClassName).To(BeNil())
				app := job.Spec.Template.Spec.Containers[0]
				Expect(app.VolumeMounts).To(Equal(podTemplate.Spec.Containers[0].VolumeMounts))
			})
		})

		Context("With a finished-job TTL", func() {
			BeforeEach(func() { opts.FinishedJobTTL = 10 * time.Minute })

			It("Should set ttlSecondsAfterFinished to the TTL, in seconds", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(job.Spec.TTLSecondsAfterFinished).To(Equal(ptr.To[int32](600)))
			})
		})

		Context("With a priority class", func() {
			BeforeEach(func() { opts.PriorityClass = "high" })

			It("Should set the Kueue priority-class label", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(job.Labels).To(HaveKeyWithValue(forge.KueuePriorityClassLabel, "high"))
				Expect(job.Labels).To(HaveKeyWithValue(forge.KueueQueueNameLabel, "crownlabs-ws"))
			})
		})

		Context("With a Fixed command", func() {
			BeforeEach(func() {
				opts.CommandPolicy = clv1alpha2.JobCommandFixed
				opts.Command = []string{"./run.sh"}
			})

			It("Should set the command, clear the args and set the working directory", func() {
				Expect(err).ToNot(HaveOccurred())
				app := job.Spec.Template.Spec.Containers[0]
				Expect(app.Command).To(Equal([]string{"./run.sh"}))
				Expect(app.Args).To(BeNil())
				Expect(app.WorkingDir).To(Equal("/home/user/work"))
			})
		})

		Context("With the ImageDefault command policy", func() {
			BeforeEach(func() {
				opts.CommandPolicy = clv1alpha2.JobCommandImageDefault
				opts.Command = nil
				podTemplate.Spec.Containers[0].Command = []string{"/entrypoint"}
				podTemplate.Spec.Containers[0].WorkingDir = "/opt"
			})

			It("Should keep command, args and working directory as they are", func() {
				Expect(err).ToNot(HaveOccurred())
				app := job.Spec.Template.Spec.Containers[0]
				Expect(app.Command).To(Equal([]string{"/entrypoint"}))
				Expect(app.Args).To(Equal([]string{"--serve"}))
				Expect(app.WorkingDir).To(Equal("/opt"))
			})
		})

		Context("With other resources, also requested by the environment", func() {
			BeforeEach(func() {
				forge.InjectOtherResources(map[string]resource.Quantity{gpu: resource.MustParse("1")}, podTemplate.Spec.Containers[0].Resources.Requests)
				forge.InjectOtherResources(map[string]resource.Quantity{gpu: resource.MustParse("1")}, podTemplate.Spec.Containers[0].Resources.Limits)
				opts.OtherResources = map[string]resource.Quantity{gpu: resource.MustParse("2"), "example.com/fpga": resource.MustParse("1")}
			})

			It("Should add them to requests and limits, summed per name", func() {
				Expect(err).ToNot(HaveOccurred())
				app := job.Spec.Template.Spec.Containers[0]
				for _, list := range []corev1.ResourceList{app.Resources.Requests, app.Resources.Limits} {
					Expect(list).To(HaveLen(4))
					Expect(list.Name(gpu, resource.DecimalSI).String()).To(Equal("3"))
					Expect(list.Name("example.com/fpga", resource.DecimalSI).String()).To(Equal("1"))
					Expect(list.Memory().String()).To(Equal("2Gi"))
				}
				Expect(app.Resources.Requests.Cpu().String()).To(Equal("1"))
				Expect(app.Resources.Limits.Cpu().String()).To(Equal("2"))
				// The instance's container is unchanged.
				Expect(podTemplate.Spec.Containers[0].Resources.Requests.Name(gpu, resource.DecimalSI).String()).To(Equal("1"))
			})

			It("Should set CROWNLABS_JOB_GPUS to the sum of the requested resources whose name ends with /gpu", func() {
				app := &job.Spec.Template.Spec.Containers[0]
				Expect(findEnv(app, forge.CrownJobEnvGPUs)).To(ConsistOf(corev1.EnvVar{Name: forge.CrownJobEnvGPUs, Value: "3"}))
			})

			When("A requested resource matches the runtime class mapping", func() {
				BeforeEach(func() { opts.RuntimeClasses = map[string]string{gpu: "nvidia"} })

				It("Should set the runtime class", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(job.Spec.Template.Spec.RuntimeClassName).To(Equal(ptr.To("nvidia")))
				})
			})

			When("Two requested resources match the same runtime class", func() {
				BeforeEach(func() { opts.RuntimeClasses = map[string]string{gpu: "nvidia", "example.com/fpga": "nvidia"} })

				It("Should set that runtime class", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(job.Spec.Template.Spec.RuntimeClassName).To(Equal(ptr.To("nvidia")))
				})
			})

			When("No requested resource matches the runtime class mapping", func() {
				BeforeEach(func() {
					opts.RuntimeClasses = map[string]string{"amd.com/gpu": "rocm"}
					podTemplate.Spec.RuntimeClassName = ptr.To("copied")
				})

				It("Should keep the copied runtime class", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(job.Spec.Template.Spec.RuntimeClassName).To(Equal(ptr.To("copied")))
				})
			})

			When("Two requested resources match different runtime classes", func() {
				BeforeEach(func() { opts.RuntimeClasses = map[string]string{gpu: "nvidia", "example.com/fpga": "fpga"} })

				It("Should fail with CannotStart", func() {
					Expect(job).To(BeNil())
					expectCannotStart(err, "The job asks for resources that need different runtimes (fpga, nvidia): it cannot run.")
				})
			})
		})

		Context("When the pod template has no container named like the environment", func() {
			BeforeEach(func() { opts.EnvironmentName = "other" })

			It("Should fail with CannotStart", func() {
				expectCannotStart(err, "The instance has no container named other.")
			})
		})

		Context("When the app container has no persistent volume mount", func() {
			BeforeEach(func() { podTemplate.Spec.Containers[0].VolumeMounts = nil })

			It("Should fail with CannotStart", func() {
				expectCannotStart(err, "The instance has no persistent disk mounted in its container app.")
			})
		})
	})

	Describe("The forge.CrownJobFinishedJobTTL function", func() {
		It("Should return the TTL in seconds, nil for no TTL, and stay within the int32 range", func() {
			Expect(forge.CrownJobFinishedJobTTL(10 * time.Minute)).To(Equal(ptr.To[int32](600)))
			Expect(forge.CrownJobFinishedJobTTL(9999 * 24 * time.Hour)).To(Equal(ptr.To[int32](863913600)))
			Expect(forge.CrownJobFinishedJobTTL(0)).To(BeNil())
			Expect(forge.CrownJobFinishedJobTTL(-time.Minute)).To(BeNil())
			Expect(forge.CrownJobFinishedJobTTL(500 * time.Millisecond)).To(BeNil())
			Expect(forge.CrownJobFinishedJobTTL(100 * 365 * 24 * time.Hour)).To(Equal(ptr.To[int32](math.MaxInt32)))
		})
	})

	Describe("The forge.CrownJobCommand function", func() {
		type commandCase struct {
			policy          clv1alpha2.JobCommandPolicy
			templateCommand []string
			crownJobCommand []string
			expected        []string
			cannotStart     string
		}

		DescribeTable("Choosing the command by commandPolicy",
			func(c commandCase) {
				command, err := forge.CrownJobCommand(&clv1alpha2.TemplateJob{CommandPolicy: c.policy, Command: c.templateCommand}, c.crownJobCommand)
				if c.cannotStart != "" {
					expectCannotStart(err, c.cannotStart)
					return
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(c.expected))
			},
			Entry("Fixed uses the template command", commandCase{
				policy: clv1alpha2.JobCommandFixed, templateCommand: []string{"t"}, expected: []string{"t"}}),
			Entry("Fixed refuses a CrownJob command", commandCase{
				policy: clv1alpha2.JobCommandFixed, templateCommand: []string{"t"}, crownJobCommand: []string{"c"},
				cannotStart: "The command of this job is fixed by its template: the job cannot set its own command."}),
			Entry("Editable uses the CrownJob command", commandCase{
				policy: clv1alpha2.JobCommandEditable, templateCommand: []string{"t"}, crownJobCommand: []string{"c"}, expected: []string{"c"}}),
			Entry("Editable falls back to the template command", commandCase{
				policy: clv1alpha2.JobCommandEditable, templateCommand: []string{"t"}, expected: []string{"t"}}),
			Entry("Editable refuses when both are empty", commandCase{
				policy: clv1alpha2.JobCommandEditable, crownJobCommand: []string{},
				cannotStart: "The job has no command: write the command to run."}),
			Entry("An empty policy is Editable", commandCase{
				crownJobCommand: []string{"c"}, expected: []string{"c"}}),
			Entry("ImageDefault uses no command", commandCase{
				policy: clv1alpha2.JobCommandImageDefault, expected: nil}),
			Entry("ImageDefault refuses a CrownJob command", commandCase{
				policy: clv1alpha2.JobCommandImageDefault, crownJobCommand: []string{"c"},
				cannotStart: "This job runs the command of its image: the job cannot set a command."}),
		)
	})

	Describe("The forge.CrownJobMaxRunTime function", func() {
		type maxRunTimeCase struct {
			crownJob, template, workspace string
			expected                      time.Duration
			cannotStart                   string
		}

		DescribeTable("Choosing the maximum run time",
			func(c maxRunTimeCase) {
				maxRunTime, err := forge.CrownJobMaxRunTime(c.crownJob, c.template, c.workspace, 2*time.Hour)
				if c.cannotStart != "" {
					expectCannotStart(err, c.cannotStart)
					return
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(maxRunTime).To(Equal(c.expected))
			},
			Entry("None set: the operator default", maxRunTimeCase{expected: 2 * time.Hour}),
			Entry("Only the workspace", maxRunTimeCase{workspace: "6h", expected: 6 * time.Hour}),
			Entry("Template below workspace", maxRunTimeCase{template: "1h", workspace: "6h", expected: time.Hour}),
			Entry("Template above workspace: the minimum", maxRunTimeCase{template: "1d", workspace: "6h", expected: 6 * time.Hour}),
			Entry("CrownJob below the template", maxRunTimeCase{crownJob: "30m", template: "1h", workspace: "6h", expected: 30 * time.Minute}),
			Entry("CrownJob equal to the template", maxRunTimeCase{crownJob: "60m", template: "1h", expected: time.Hour}),
			Entry("CrownJob above the template", maxRunTimeCase{crownJob: "2h", template: "1h", workspace: "6h",
				cannotStart: "The job asks for a time limit of 2h, but at most 1h is allowed."}),
			Entry("CrownJob above the workspace, no template", maxRunTimeCase{crownJob: "1d", workspace: "6h",
				cannotStart: "The job asks for a time limit of 1d, but at most 6h is allowed."}),
			Entry("CrownJob below the template but above the workspace", maxRunTimeCase{crownJob: "8h", template: "1d", workspace: "6h",
				cannotStart: "The job asks for a time limit of 8h, but at most 6h is allowed."}),
			Entry("CrownJob above the operator default, nothing else set", maxRunTimeCase{crownJob: "3h",
				cannotStart: "The job asks for a time limit of 3h, but at most 2h is allowed."}),
			Entry("CrownJob below the operator default, nothing else set", maxRunTimeCase{crownJob: "90m", expected: 90 * time.Minute}),
		)
	})

	Describe("The utils.ParseDurationWithDays and utils.FormatDurationWithDays functions", func() {
		DescribeTable("Parsing the shapes the API pattern admits, and its extremes",
			func(input string, expected time.Duration) {
				Expect(utils.ParseDurationWithDays(input)).To(Equal(expected))
			},
			Entry("1m", "1m", time.Minute),
			Entry("9999m", "9999m", 9999*time.Minute),
			Entry("1h", "1h", time.Hour),
			Entry("9999h", "9999h", 9999*time.Hour),
			Entry("1d", "1d", 24*time.Hour),
			Entry("9999d", "9999d", 9999*24*time.Hour),
		)

		DescribeTable("Refusing what is not a positive duration that fits",
			func(input string) {
				_, err := utils.ParseDurationWithDays(input)
				Expect(err).To(HaveOccurred())
			},
			Entry("empty", ""),
			Entry("zero", "0h"),
			Entry("seconds", "30s"),
			Entry("no unit", "30"),
			Entry("negative", "-1h"),
			Entry("106752d overflows int64 nanoseconds (P8-1)", "106752d"),
			Entry("a value larger than int64", "99999999999999999999m"),
		)

		DescribeTable("Formatting with the largest unit that divides the duration",
			func(input time.Duration, expected string) {
				Expect(utils.FormatDurationWithDays(input)).To(Equal(expected))
			},
			Entry("days", 48*time.Hour, "2d"),
			Entry("hours", 6*time.Hour, "6h"),
			Entry("minutes", 90*time.Minute, "90m"),
			Entry("seconds are rounded up", 61*time.Second, "2m"),
		)
	})

	Describe("The forge.PodEffectiveRequests function", func() {
		It("Should take, per resource, the largest between each init container and the sum of the containers", func() {
			spec := corev1.PodSpec{
				InitContainers: []corev1.Container{
					{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("10m"), "memory": resource.MustParse("8Gi")}}},
				},
				Containers: []corev1.Container{
					{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("500m"), "memory": resource.MustParse("1Gi")}}},
					{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("250m"), gpu: resource.MustParse("1")}}},
				},
			}
			requests := forge.PodEffectiveRequests(&spec)
			Expect(requests).To(HaveLen(3))
			Expect(requests.Cpu().String()).To(Equal("750m"))
			Expect(requests.Memory().String()).To(Equal("8Gi"))
			Expect(requests.Name(gpu, resource.DecimalSI).String()).To(Equal("1"))
		})
	})

	Describe("The forge.CrownJobCheckQueueLimits function", func() {
		var limits clv1alpha1.QueueLimits

		BeforeEach(func() {
			limits = clv1alpha1.QueueLimits{MaxAtOnce: map[string]resource.Quantity{"cpu": resource.MustParse("4"), gpu: resource.MustParse("2")}}
		})

		It("Should accept requests within maxAtOnce, and memory that maxAtOnce does not list", func() {
			Expect(forge.CrownJobCheckQueueLimits(workspaceName, &limits, corev1.ResourceList{
				"cpu": resource.MustParse("4"), "memory": resource.MustParse("64Gi"), gpu: resource.MustParse("2"),
			})).To(Succeed())
		})

		It("Should refuse a resource that the queue does not cover", func() {
			expectCannotStart(forge.CrownJobCheckQueueLimits(workspaceName, &limits, corev1.ResourceList{
				"cpu": resource.MustParse("1"), "amd.com/gpu": resource.MustParse("1"),
			}), "The job queue of workspace ws does not allow amd.com/gpu.")
		})

		It("Should refuse a request above maxAtOnce", func() {
			expectCannotStart(forge.CrownJobCheckQueueLimits(workspaceName, &limits, corev1.ResourceList{
				"cpu": resource.MustParse("1"), gpu: resource.MustParse("3"),
			}), "The job asks for 3 of nvidia.com/gpu, but the job queue of workspace ws allows at most 2 at once: it could never start.")
		})

		It("Should refuse a GPU when maxAtOnce lists none", func() {
			expectCannotStart(forge.CrownJobCheckQueueLimits(workspaceName, &clv1alpha1.QueueLimits{}, corev1.ResourceList{
				gpu: resource.MustParse("1"),
			}), "The job queue of workspace ws does not allow nvidia.com/gpu.")
		})
	})
})
