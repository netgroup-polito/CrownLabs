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
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("CrownJob log helpers forging", func() {
	const (
		namespace = "tenant-student"
		envName   = "app"
		nodeName  = "node-2"
		podName   = "train-1-abcde"
	)

	var (
		crownJob clv1alpha2.CrownJob
		job      *batchv1.Job
		pod      corev1.Pod
		opts     forge.CrownJobLogHelperOpts
		index    int
		helper   *batchv1.Job
		err      error
	)

	BeforeEach(func() {
		crownJob = clv1alpha2.CrownJob{ObjectMeta: metav1.ObjectMeta{Name: "train-1", Namespace: namespace, UID: "cj-uid"}}
		environment := clv1alpha2.Environment{
			Name: envName, Image: "harbor.example.com/standalone:v1", EnvironmentType: clv1alpha2.ClassStandalone, Persistent: true,
			Resources: clv1alpha2.EnvironmentResources{
				ResourceSpec:          apicommon.ResourceSpec{CPU: 1, Memory: resource.MustParse("1Gi"), Disk: resource.MustParse("2Gi")},
				ReservedCPUPercentage: 50,
			},
		}
		template := clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: "tpl", Namespace: "workspace-ws"},
			Spec:       clv1alpha2.TemplateSpec{WorkspaceRef: clv1alpha2.GenericRef{Name: "ws"}, EnvironmentList: []clv1alpha2.Environment{environment}},
		}
		instance := clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: "inst", Namespace: namespace},
			Spec:       clv1alpha2.InstanceSpec{Template: clv1alpha2.GenericRef{Name: "tpl", Namespace: "workspace-ws"}, Running: true},
		}
		mountInfos := []corev1.VolumeMount{{Name: "shvol", MountPath: "/mnt/shared", ReadOnly: true}}
		podTemplate := forge.DeploymentSpec(&instance, &template, &environment, mountInfos,
			&forge.ContainerEnvOpts{ImagesTag: "v1", ContentToolsImg: "tools"}).Template
		job, err = forge.CrownJobBatchJob(&crownJob, &forge.CrownJobJobOpts{
			Workspace: "ws", EnvironmentName: envName, PodTemplate: &podTemplate,
			CommandPolicy: clv1alpha2.JobCommandEditable, Command: []string{"python", "train.py"}, MaxRunTime: time.Hour,
		})
		Expect(err).ToNot(HaveOccurred())

		pod = corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: namespace}, Spec: corev1.PodSpec{NodeName: nodeName}}
		opts = forge.CrownJobLogHelperOpts{Image: "registry.k8s.io/kubectl:v1.34.1"}
		index = 1
	})

	JustBeforeEach(func() {
		helper, err = forge.CrownJobLogHelperJob(&crownJob, job, &pod, index, &opts)
	})

	Describe("The forge.CrownJobLogHelperJob function", func() {
		It("Should not fail", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should set the name, the namespace and the controller ownerReference to the CrownJob", func() {
			Expect(helper.Name).To(Equal(forge.CrownJobLogHelperName(&crownJob, 1)))
			Expect(helper.Namespace).To(Equal(namespace))
			Expect(helper.OwnerReferences).To(HaveLen(1))
			Expect(helper.OwnerReferences[0].Name).To(Equal(crownJob.Name))
			Expect(helper.OwnerReferences[0].Controller).To(PointTo(BeTrue()))
		})

		It("Should set only the CrownJob and component labels, on the Job and on its pods: no Kueue and no CrownLabs label", func() {
			expected := map[string]string{forge.LabelCrownJobKey: "train-1", forge.LabelCrownJobComponentKey: forge.CrownJobComponentLogHelper}
			Expect(helper.Labels).To(Equal(expected))
			Expect(helper.Spec.Template.Labels).To(Equal(expected))
			for key := range helper.Labels {
				Expect(key).ToNot(HavePrefix("kueue.x-k8s.io/"))
				Expect(key).ToNot(HavePrefix("crownlabs.polito.it/"))
			}
		})

		It("Should name the pod and the log file in its annotations", func() {
			Expect(helper.Annotations).To(Equal(map[string]string{
				forge.AnnotationCrownJobPodKey:     podName,
				forge.AnnotationCrownJobLogFileKey: "runs/train-1/log.txt",
			}))
		})

		It("Should retry twice, and never restart the container", func() {
			Expect(helper.Spec.BackoffLimit).To(PointTo(BeNumerically("==", 2)))
			Expect(helper.Spec.Template.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
			Expect(helper.Spec.Suspend).To(BeNil())
		})

		It("Should set no ttlSecondsAfterFinished when no TTL is given", func() {
			Expect(helper.Spec.TTLSecondsAfterFinished).To(BeNil())
		})

		When("a finished-job TTL is given", func() {
			BeforeEach(func() { opts.FinishedJobTTL = 90 * time.Minute })

			It("Should set ttlSecondsAfterFinished to the TTL, in seconds", func() {
				Expect(helper.Spec.TTLSecondsAfterFinished).To(PointTo(BeNumerically("==", 5400)))
			})
		})

		It("Should use the log reader ServiceAccount, with its token mounted", func() {
			Expect(helper.Spec.Template.Spec.ServiceAccountName).To(Equal("crownjob-log-reader"))
			Expect(helper.Spec.Template.Spec.AutomountServiceAccountToken).To(PointTo(BeTrue()))
		})

		It("Should use the security context of the job pod (UID, GID and fsGroup 1010)", func() {
			Expect(helper.Spec.Template.Spec.SecurityContext).To(Equal(job.Spec.Template.Spec.SecurityContext))
			Expect(helper.Spec.Template.Spec.SecurityContext.RunAsUser).To(PointTo(BeNumerically("==", 1010)))
			Expect(helper.Spec.Template.Spec.SecurityContext.RunAsGroup).To(PointTo(BeNumerically("==", 1010)))
			Expect(helper.Spec.Template.Spec.SecurityContext.FSGroup).To(PointTo(BeNumerically("==", 1010)))
		})

		It("Should mount only the persistent disk of the job, at the same path", func() {
			var jobVolume corev1.Volume
			for _, v := range job.Spec.Template.Spec.Volumes {
				if v.Name == forge.PersistentVolumeName {
					jobVolume = v
				}
			}
			Expect(helper.Spec.Template.Spec.Volumes).To(Equal([]corev1.Volume{jobVolume}))
			Expect(helper.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(helper.Spec.Template.Spec.Containers[0].VolumeMounts).To(Equal([]corev1.VolumeMount{
				{Name: forge.PersistentVolumeName, MountPath: forge.PersistentDefaultMountPath},
			}))
			Expect(helper.Spec.Template.Spec.InitContainers).To(BeEmpty())
		})

		It("Should follow the log of the app container through go-runner, writing it into the run folder", func() {
			c := helper.Spec.Template.Spec.Containers[0]
			Expect(c.Name).To(Equal("save-log"))
			Expect(c.Image).To(Equal("registry.k8s.io/kubectl:v1.34.1"))
			Expect(c.ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
			Expect(c.Command).To(Equal([]string{
				"/go-runner", "-log-file=" + forge.PersistentDefaultMountPath + "/runs/train-1/log.txt", "-redirect-stderr=false",
				"/bin/kubectl", "logs", "-f", "-n", namespace, podName, "-c", envName,
			}))
			Expect(c.Args).To(BeEmpty())
		})

		It("Should request 10m/100m CPU and 32Mi/64Mi memory", func() {
			r := helper.Spec.Template.Spec.Containers[0].Resources
			Expect(r.Requests.Cpu().String()).To(Equal("10m"))
			Expect(r.Limits.Cpu().String()).To(Equal("100m"))
			Expect(r.Requests.Memory().String()).To(Equal("32Mi"))
			Expect(r.Limits.Memory().String()).To(Equal("64Mi"))
		})

		It("Should carry a pod-level deadline of maxRunTime + 600 s, so that the instances quota does not count it", func() {
			Expect(helper.Spec.Template.Spec.ActiveDeadlineSeconds).To(PointTo(BeNumerically("==", 3600+600)))
			Expect(helper.Spec.ActiveDeadlineSeconds).To(BeNil())
		})

		It("Should run on the node of the job pod, selected by name, when the disk is not ReadWriteMany", func() {
			Expect(helper.Spec.Template.Spec.Affinity).To(Equal(&corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchFields: []corev1.NodeSelectorRequirement{{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{nodeName}}},
				}}},
			}}))
		})

		When("the disk is ReadWriteMany", func() {
			BeforeEach(func() { opts.ReadWriteMany = true })
			It("Should have no affinity", func() {
				Expect(helper.Spec.Template.Spec.Affinity).To(BeNil())
			})
		})

		When("the pod replaces the first one", func() {
			BeforeEach(func() { index = 3 })
			It("Should write log-3.txt, with its own name", func() {
				Expect(helper.Annotations[forge.AnnotationCrownJobLogFileKey]).To(Equal("runs/train-1/log-3.txt"))
				Expect(helper.Spec.Template.Spec.Containers[0].Command[1]).To(HaveSuffix("/runs/train-1/log-3.txt"))
				Expect(helper.Name).To(Equal(forge.CrownJobLogHelperName(&crownJob, 3)))
				Expect(helper.Name).ToNot(Equal(forge.CrownJobLogHelperName(&crownJob, 1)))
			})
		})

		When("the job has no persistent disk", func() {
			BeforeEach(func() { job.Spec.Template.Spec.Volumes = nil })
			It("Should fail", func() {
				Expect(err).To(MatchError("the job train-1 has no persistent disk"))
			})
		})
	})

	Describe("The forge.CrownJobLogHelperName function", func() {
		It("Should append the index and a hash of the UID, and be a valid label value", func() {
			name := forge.CrownJobLogHelperName(&crownJob, 1)
			Expect(name).To(MatchRegexp(`^train-1-log1-[0-9a-f]{5}$`))
			Expect(validation.IsValidLabelValue(name)).To(BeEmpty())
			Expect(validation.IsDNS1123Subdomain(name)).To(BeEmpty())
		})

		It("Should differ for two CrownJobs with the same name and different UIDs", func() {
			other := crownJob.DeepCopy()
			other.UID = "other-uid"
			Expect(forge.CrownJobLogHelperName(other, 1)).ToNot(Equal(forge.CrownJobLogHelperName(&crownJob, 1)))
		})

		It("Should never equal the name of a CrownJob called <name>-log", func() {
			Expect(forge.CrownJobLogHelperName(&crownJob, 1)).ToNot(Equal("train-1-log"))
		})

		It("Should cut a 63-character name to fit, without ending a DNS label with a dot or a dash", func() {
			for _, name := range []string{strings.Repeat("a", 63), strings.Repeat("b", 54) + "." + strings.Repeat("c", 8), strings.Repeat("d", 54) + "-x"} {
				long := crownJob.DeepCopy()
				long.Name = name
				helperName := forge.CrownJobLogHelperName(long, 12)
				Expect(len(helperName)).To(BeNumerically("<=", 63), helperName)
				Expect(validation.IsValidLabelValue(helperName)).To(BeEmpty(), helperName)
				Expect(validation.IsDNS1123Subdomain(helperName)).To(BeEmpty(), helperName)
			}
		})
	})

	Describe("The forge.CrownJobLogFile function", func() {
		It("Should return log.txt for the first pod, and log-N.txt for the others", func() {
			Expect(forge.CrownJobLogFile("train-1", 1)).To(Equal("runs/train-1/log.txt"))
			Expect(forge.CrownJobLogFile("train-1", 2)).To(Equal("runs/train-1/log-2.txt"))
		})
	})

	Describe("The permissions of the log helpers", func() {
		It("Should forge a ServiceAccount with the managed-by label", func() {
			sa := forge.CrownJobLogReaderServiceAccount(namespace)
			Expect(sa.Name).To(Equal("crownjob-log-reader"))
			Expect(sa.Namespace).To(Equal(namespace))
			Expect(sa.Labels).To(Equal(map[string]string{"crownlabs.polito.it/managed-by": "crownjob"}))
			Expect(sa.APIVersion).To(Equal("v1"))
			Expect(sa.Kind).To(Equal("ServiceAccount"))
		})

		It("Should forge a Role that can only get and list pods, and get their logs", func() {
			role := forge.CrownJobLogReaderRole(namespace)
			Expect(role.Name).To(Equal("crownjob-log-reader"))
			Expect(role.Labels).To(Equal(map[string]string{"crownlabs.polito.it/managed-by": "crownjob"}))
			Expect(role.Rules).To(Equal([]rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get", "list"}},
				{APIGroups: []string{""}, Resources: []string{"pods/log"}, Verbs: []string{"get"}},
			}))
		})

		It("Should forge a RoleBinding of that Role to that ServiceAccount", func() {
			rb := forge.CrownJobLogReaderRoleBinding(namespace)
			Expect(rb.Name).To(Equal("crownjob-log-reader"))
			Expect(rb.Labels).To(Equal(map[string]string{"crownlabs.polito.it/managed-by": "crownjob"}))
			Expect(rb.RoleRef).To(Equal(rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "Role", Name: "crownjob-log-reader"}))
			Expect(rb.Subjects).To(Equal([]rbacv1.Subject{{Kind: "ServiceAccount", Name: "crownjob-log-reader", Namespace: namespace}}))
		})
	})

	Describe("The forge.IsReadWriteMany function", func() {
		It("Should tell whether ReadWriteMany is among the access modes", func() {
			Expect(forge.IsReadWriteMany([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce})).To(BeFalse())
			Expect(forge.IsReadWriteMany(nil)).To(BeFalse())
			Expect(forge.IsReadWriteMany([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce, corev1.ReadWriteMany})).To(BeTrue())
		})
	})

	It("Should keep the job pod deadline as the base of the helper deadline", func() {
		job.Spec.Template.Spec.ActiveDeadlineSeconds = ptr.To[int64](60)
		helper, err = forge.CrownJobLogHelperJob(&crownJob, job, &pod, 1, &opts)
		Expect(err).ToNot(HaveOccurred())
		Expect(helper.Spec.Template.Spec.ActiveDeadlineSeconds).To(PointTo(BeNumerically("==", 660)))
	})
})
