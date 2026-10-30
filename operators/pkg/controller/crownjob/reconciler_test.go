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

package crownjob_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/crownjob"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var crownJobCounter = 0

// newCrownJob creates a CrownJob with a unique name in the tenant namespace.
func newCrownJob(instanceName string, command []string, maxRunTime string) *clv1alpha2.CrownJob {
	crownJobCounter++
	cj := &clv1alpha2.CrownJob{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("cj-%d", crownJobCounter), Namespace: tenantNamespace},
		Spec: clv1alpha2.CrownJobSpec{
			Instance:   clv1alpha2.GenericRef{Name: instanceName},
			Command:    command,
			MaxRunTime: maxRunTime,
		},
	}
	ExpectWithOffset(1, k8sClient.Create(ctx, cj)).To(Succeed())
	return cj
}

func reconcileCrownJob(cj *clv1alpha2.CrownJob) ctrl.Result {
	result, err := crownJobReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cj)})
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	return result
}

func crownJobStatus(cj *clv1alpha2.CrownJob) clv1alpha2.CrownJobStatus {
	var current clv1alpha2.CrownJob
	ExpectWithOffset(1, k8sClient.Get(ctx, client.ObjectKeyFromObject(cj), &current)).To(Succeed())
	return current.Status
}

func expectPhase(cj *clv1alpha2.CrownJob, phase clv1alpha2.CrownJobPhase, reason, message string) {
	status := crownJobStatus(cj)
	ExpectWithOffset(1, status.Phase).To(Equal(phase))
	ExpectWithOffset(1, status.Reason).To(Equal(reason))
	ExpectWithOffset(1, status.Message).To(Equal(message))
}

func getJob(name string) (*batchv1.Job, error) {
	var job batchv1.Job
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: tenantNamespace, Name: name}, &job)
	return &job, err
}

// drainEvents returns the events recorded since the last call.
func drainEvents() []string {
	var events []string
	for {
		select {
		case event := <-recorder.Events:
			events = append(events, event)
		default:
			return events
		}
	}
}

// admit simulates Kueue admitting the Job: it un-suspends it.
func admit(job *batchv1.Job) {
	ExpectWithOffset(1, k8sClient.Get(ctx, client.ObjectKeyFromObject(job), job)).To(Succeed())
	job.Spec.Suspend = ptr.To(false)
	ExpectWithOffset(1, k8sClient.Update(ctx, job)).To(Succeed())
}

// updateJobStatus simulates the Job controller (envtest has none).
func updateJobStatus(job *batchv1.Job, mutate func(status *batchv1.JobStatus)) {
	ExpectWithOffset(1, k8sClient.Get(ctx, client.ObjectKeyFromObject(job), job)).To(Succeed())
	mutate(&job.Status)
	ExpectWithOffset(1, k8sClient.Status().Update(ctx, job)).To(Succeed())
}

func jobCondition(conditionType batchv1.JobConditionType, reason, message string) batchv1.JobCondition {
	now := metav1.Now()
	return batchv1.JobCondition{Type: conditionType, Status: corev1.ConditionTrue, Reason: reason, Message: message,
		LastProbeTime: now, LastTransitionTime: now}
}

// failJob sets the status of a failed Job, as the Job controller does (FailureTarget, then Failed).
func failJob(job *batchv1.Job, reason, message string) {
	updateJobStatus(job, func(status *batchv1.JobStatus) {
		status.StartTime = ptr.To(metav1.Now())
		status.Failed = 1
		status.Conditions = []batchv1.JobCondition{
			jobCondition(batchv1.JobFailureTarget, reason, message),
			jobCondition(batchv1.JobFailed, reason, message),
		}
	})
}

// createPod simulates the Job controller creating the pod of the Job, and the kubelet writing its status.
func createPod(job *batchv1.Job, status *corev1.PodStatus) *corev1.Pod {
	ExpectWithOffset(1, k8sClient.Get(ctx, client.ObjectKeyFromObject(job), job)).To(Succeed())
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: job.Name + "-pod", Namespace: job.Namespace,
			Labels: map[string]string{batchv1.ControllerUidLabel: string(job.UID), batchv1.JobNameLabel: job.Name},
		},
		Spec: *job.Spec.Template.Spec.DeepCopy(),
	}
	ExpectWithOffset(1, k8sClient.Create(ctx, pod)).To(Succeed())
	pod.Status = *status
	ExpectWithOffset(1, k8sClient.Status().Update(ctx, pod)).To(Succeed())
	return pod
}

func waiting(name, reason string) corev1.ContainerStatus {
	return corev1.ContainerStatus{Name: name, Image: image, State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason}}}
}

func terminated(name string, exitCode int32) corev1.ContainerStatus {
	return corev1.ContainerStatus{Name: name, Image: image, ImageID: "sha256:1",
		State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: exitCode}}}
}

func running(name string) corev1.ContainerStatus {
	return corev1.ContainerStatus{Name: name, Image: image, ImageID: "sha256:1", Ready: true, Started: ptr.To(true),
		State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Now()}}}
}

// startJob creates a CrownJob, reconciles it until its Job exists, and admits the Job.
func startJob(instanceName string, command []string) (*clv1alpha2.CrownJob, *batchv1.Job) {
	cj := newCrownJob(instanceName, command, "")
	reconcileCrownJob(cj)
	job, err := getJob(cj.Name)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	admit(job)
	return cj, job
}

var _ = Describe("The CrownJob controller", func() {
	sh := []string{"sh", "-c", "echo hello > $CROWNLABS_RUN_FOLDER/out.txt"}

	BeforeEach(func() { drainEvents() })

	Describe("Validating a CrownJob before its Job exists", func() {
		type cannotStartCase struct {
			instanceRef clv1alpha2.GenericRef
			name        string
			command     []string
			maxRunTime  string
			message     string
		}

		DescribeTable("Refusing a job that can never start: Failed, CannotStart, an event, and no Job",
			func(c cannotStartCase) {
				crownJobCounter++
				name := c.name
				if name == "" {
					name = fmt.Sprintf("cj-%d", crownJobCounter)
				}
				cj := &clv1alpha2.CrownJob{
					ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: tenantNamespace},
					Spec:       clv1alpha2.CrownJobSpec{Instance: c.instanceRef, Command: c.command, MaxRunTime: c.maxRunTime},
				}
				Expect(k8sClient.Create(ctx, cj)).To(Succeed())

				Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{}))
				expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCannotStart, c.message)
				Expect(crownJobStatus(cj).RunFolder).To(BeEmpty())
				Expect(drainEvents()).To(ConsistOf("Warning CannotStart " + c.message))

				_, err := getJob(cj.Name)
				Expect(kerrors.IsNotFound(err)).To(BeTrue(), "the Job must not exist")
			},
			Entry("1. an instance in another namespace (a tenant must not copy another tenant's disk)", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst", Namespace: otherNamespace}, command: []string{"true"},
				message: "The job must use an instance in its own namespace (tenant-student), not in tenant-other."}),
			Entry("2. a name longer than 63 characters", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst"}, command: []string{"true"}, name: strings.Repeat("a", 64),
				message: "The job name is too long: it can have at most 63 characters."}),
			Entry("3. a missing instance", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "ghost"}, command: []string{"true"},
				message: "The instance ghost does not exist."}),
			Entry("3. a missing template", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-ghosttpl"}, command: []string{"true"},
				message: "The template ghost of the instance inst-ghosttpl does not exist."}),
			Entry("3. a template without a job section", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-nojob"}, command: []string{"true"},
				message: "Jobs cannot be started from the template tpl-nojob."}),
			Entry("4. a template without a workspace", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-nows"}, command: []string{"true"},
				message: "The template tpl-nows does not belong to a workspace."}),
			Entry("4. a missing workspace", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-ghostws"}, command: []string{"true"},
				message: "The workspace ghost does not exist."}),
			Entry("4. a workspace without queueLimits", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-noqueue"}, command: []string{"true"},
				message: "The workspace ws-noqueue has no job queue."}),
			Entry("5. a missing tenant", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-ghosttenant"}, command: []string{"true"},
				message: "The tenant ghost does not exist."}),
			Entry("5. a tenant that is only a candidate of the workspace", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-candidate"}, command: []string{"true"},
				message: "The tenant candidate is not a member of the workspace ws."}),
			Entry("6. a command with commandPolicy Fixed", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-fixed"}, command: []string{"true"},
				message: "The command of this job is fixed by its template: the job cannot set its own command."}),
			Entry("6. no command with commandPolicy Editable and no template command", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst"},
				message:     "The job has no command: write the command to run."}),
			Entry("6. a command with commandPolicy ImageDefault", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-imagedefault"}, command: []string{"true"},
				message: "This job runs the command of its image: the job cannot set a command."}),
			Entry("7. a maxRunTime above the template's", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst"}, command: []string{"true"}, maxRunTime: "2h",
				message: "The job asks for a time limit of 2h, but at most 1h is allowed."}),
			Entry("8. a resource the queue does not cover", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-amd"}, command: []string{"true"},
				message: "The job queue of workspace ws does not allow amd.com/gpu."}),
			Entry("8. a request above maxAtOnce", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-bigcpu"}, command: []string{"true"},
				message: "The job asks for 8 of cpu, but the job queue of workspace ws allows at most 4 at once: it could never start."}),
			Entry("8. resources mapped to two different runtime classes", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-conflict"}, command: []string{"true"},
				message: "The job asks for resources that need different runtimes (fpga, nvidia): it cannot run."}),
			Entry("9. a priority class that does not exist", cannotStartCase{
				instanceRef: clv1alpha2.GenericRef{Name: "inst-prio-missing"}, command: []string{"true"},
				message: "The job priority class ghost does not exist."}),
		)

		It("Should refuse when a Job with the same name, not owned by the CrownJob, already exists", func() {
			crownJobCounter++
			name := fmt.Sprintf("cj-%d", crownJobCounter)
			Expect(k8sClient.Create(ctx, &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: tenantNamespace},
				Spec: batchv1.JobSpec{Suspend: ptr.To(true), Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever, Containers: []corev1.Container{{Name: "c", Image: "busybox"}},
				}}},
			})).To(Succeed())
			cj := &clv1alpha2.CrownJob{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: tenantNamespace},
				Spec:       clv1alpha2.CrownJobSpec{Instance: clv1alpha2.GenericRef{Name: "inst"}, Command: []string{"true"}},
			}
			Expect(k8sClient.Create(ctx, cj)).To(Succeed())
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCannotStart,
				fmt.Sprintf("Another job named %s already exists in this namespace.", name))
		})

		It("Should never change a CannotStart phase, even once the cause is gone", func() {
			cj := newCrownJob("inst-late-terminal", sh, "")
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCannotStart, "The instance inst-late-terminal does not exist.")

			createInstance(tenantNamespace, "inst-late-terminal", "tpl", tenantName, true)
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCannotStart, "The instance inst-late-terminal does not exist.")
			_, err := getJob(cj.Name)
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
		})

		DescribeTable("Checking the job environment of an existing Template (also one created before the CEL rules)",
			func(env clv1alpha2.Environment, message string) {
				template := &clv1alpha2.Template{ObjectMeta: metav1.ObjectMeta{Name: "tpl-x"}, Spec: clv1alpha2.TemplateSpec{
					EnvironmentList: []clv1alpha2.Environment{env}, Job: &clv1alpha2.TemplateJob{EnvironmentName: envName}}}
				_, err := crownjob.JobEnvironment(template)
				Expect(err).To(MatchError(message))
			},
			Entry("not persistent", clv1alpha2.Environment{Name: envName, EnvironmentType: clv1alpha2.ClassContainer},
				"Jobs need a persistent environment, and the environment work of the template tpl-x is not."),
			Entry("a virtual machine", clv1alpha2.Environment{Name: envName, EnvironmentType: clv1alpha2.ClassVM, Persistent: true},
				"Jobs need an environment of type Container or Standalone, and the environment work of the template tpl-x is of type VirtualMachine."),
			Entry("missing", clv1alpha2.Environment{Name: "other", EnvironmentType: clv1alpha2.ClassContainer, Persistent: true},
				"The template tpl-x has no environment named work."),
		)
	})

	Describe("Waiting for the Deployment of the instance", func() {
		It("Should wait with InstanceNotReady (not terminal), then create the Job once the Deployment appears", func() {
			createInstance(tenantNamespace, "inst-late", "tpl", tenantName, false)
			cj := newCrownJob("inst-late", sh, "")

			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 15 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonInstanceNotReady,
				"The instance inst-late is not ready yet: the job will be created when it is.")
			_, err := getJob(cj.Name)
			Expect(kerrors.IsNotFound(err)).To(BeTrue())

			var inst clv1alpha2.Instance
			Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: tenantNamespace, Name: "inst-late"}, &inst)).To(Succeed())
			createDeployment(&inst, "tpl")

			// The Deployment exists, its disk not yet: still not ready.
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 15 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonInstanceNotReady,
				"The instance inst-late is not ready yet: the job will be created when it is.")
			_, err = getJob(cj.Name)
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
			createDisk(&inst, corev1.ReadWriteOnce)
			// The disk is ReadWriteOnce: the job waits until the instance is stopped.
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonDiskInUse,
				"Your job will start when you stop the instance inst-late: its disk can be used by one machine at a time.")
			stopInstanceContainer(&inst)

			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonInQueue, "Your job is waiting in the queue of workspace ws.")
			_, err = getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			// Waiting -> Waiting is not a phase change.
			Expect(drainEvents()).To(ConsistOf("Normal InstanceNotReady The instance inst-late is not ready yet: the job will be created when it is."))
		})
	})

	Describe("Creating the Job", func() {
		It("Should create the Job and the LocalQueue with the expected fields, and wait in the queue", func() {
			cj := newCrownJob("inst", sh, "30m")
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))

			status := crownJobStatus(cj)
			Expect(status.Phase).To(Equal(clv1alpha2.CrownJobPhaseWaiting))
			Expect(status.Reason).To(Equal(crownjob.ReasonInQueue))
			Expect(status.Message).To(Equal("Your job is waiting in the queue of workspace ws."))
			Expect(status.RunFolder).To(Equal("runs/" + cj.Name))
			Expect(status.QueuePosition).To(BeNil())
			Expect(drainEvents()).To(ConsistOf("Normal InQueue Your job is waiting in the queue of workspace ws."))

			job, err := getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(job.Labels).To(HaveKeyWithValue(forge.KueueQueueNameLabel, "crownlabs-ws"))
			Expect(job.Labels).To(HaveKeyWithValue(forge.LabelCrownJobKey, cj.Name))
			Expect(job.Labels).ToNot(HaveKey(forge.KueuePriorityClassLabel))
			Expect(job.OwnerReferences).To(HaveLen(1))
			Expect(job.OwnerReferences[0].UID).To(Equal(crownJobUID(cj)))
			Expect(ptr.Deref(job.OwnerReferences[0].Controller, false)).To(BeTrue())
			Expect(job.Spec.Suspend).To(Equal(ptr.To(true)))
			Expect(job.Spec.BackoffLimit).To(Equal(ptr.To[int32](0)))
			Expect(job.Spec.ActiveDeadlineSeconds).To(Equal(ptr.To[int64](30*60 + 600)))
			Expect(job.Spec.PodFailurePolicy.Rules).To(HaveLen(2))
			Expect(job.Spec.TTLSecondsAfterFinished).To(Equal(ptr.To[int32](600)))

			pod := job.Spec.Template
			Expect(pod.Spec.ActiveDeadlineSeconds).To(Equal(ptr.To[int64](30 * 60)))
			Expect(pod.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
			Expect(pod.Labels).To(HaveKeyWithValue("job.crownlabs.polito.it/instance", "inst"))
			Expect(pod.Labels).To(HaveKeyWithValue("job.crownlabs.polito.it/workspace", workspaceName))
			Expect(pod.Labels).ToNot(HaveKey(forge.LabelInstanceKey))
			Expect(pod.Labels).ToNot(HaveKey(forge.KueueQueueNameLabel))
			Expect(pod.Spec.Containers).To(HaveLen(1))
			Expect(pod.Spec.Containers[0].Command).To(Equal(sh))
			Expect(pod.Spec.Containers[0].WorkingDir).To(Equal(forge.PersistentDefaultMountPath))
			Expect(pod.Spec.Containers[0].ReadinessProbe).To(BeNil())
			Expect(pod.Spec.InitContainers).To(HaveLen(1))
			Expect(pod.Spec.InitContainers[0].Command).To(Equal([]string{"mkdir", "-p", "/media/data/runs/" + cj.Name}))
			Expect(pod.Spec.RuntimeClassName).To(BeNil())

			localQueue := &unstructured.Unstructured{}
			localQueue.SetGroupVersionKind(forge.KueueLocalQueueGVK)
			Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: tenantNamespace, Name: "crownlabs-ws"}, localQueue)).To(Succeed())
			Expect(localQueue.Object["spec"]).To(HaveKeyWithValue("clusterQueue", "crownlabs-ws"))
			Expect(localQueue.GetOwnerReferences()).To(HaveLen(1))
			Expect(localQueue.GetOwnerReferences()[0].Kind).To(Equal("Workspace"))
			Expect(localQueue.GetOwnerReferences()[0].Name).To(Equal(workspaceName))

			// The Workload kind is not installed in envtest (the read fails with NoMatch): the job stays InQueue, without errors.
			workloads := &unstructured.UnstructuredList{}
			workloads.SetGroupVersionKind(forge.KueueWorkloadGVK.GroupVersion().WithKind("WorkloadList"))
			Expect(meta.IsNoMatchError(k8sClient.List(ctx, workloads, client.InNamespace(tenantNamespace)))).To(BeTrue())
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonInQueue, "Your job is waiting in the queue of workspace ws.")
		})

		It("Should set the priority-class label from the tenant entry", func() {
			cj := newCrownJob("inst-prio", sh, "")
			reconcileCrownJob(cj)
			job, err := getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(job.Labels).To(HaveKeyWithValue(forge.KueuePriorityClassLabel, highPriorityClass))
		})

		It("Should add other resources, set the GPU count and map the runtime class", func() {
			cj := newCrownJob("inst-gpu", sh, "")
			reconcileCrownJob(cj)
			job, err := getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			app := job.Spec.Template.Spec.Containers[0]
			Expect(app.Resources.Requests).To(HaveKeyWithValue(corev1.ResourceName(gpu), resource.MustParse("1")))
			Expect(app.Resources.Limits).To(HaveKeyWithValue(corev1.ResourceName(gpu), resource.MustParse("1")))
			Expect(app.Env).To(ContainElement(corev1.EnvVar{Name: forge.CrownJobEnvGPUs, Value: "1"}))
			Expect(job.Spec.Template.Spec.RuntimeClassName).To(Equal(ptr.To("nvidia")))
		})

		It("Should use the template command with commandPolicy Fixed, and no command with ImageDefault", func() {
			cj := newCrownJob("inst-fixed", nil, "")
			reconcileCrownJob(cj)
			job, err := getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(job.Spec.Template.Spec.Containers[0].Command).To(Equal([]string{"./run.sh"}))

			cj = newCrownJob("inst-imagedefault", nil, "")
			reconcileCrownJob(cj)
			job, err = getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(job.Spec.Template.Spec.Containers[0].Command).To(BeEmpty())
		})

		It("Should ignore a CrownJob whose namespace does not have the target label", func() {
			cj := &clv1alpha2.CrownJob{
				ObjectMeta: metav1.ObjectMeta{Name: "untargeted", Namespace: untargetedNamespace},
				Spec:       clv1alpha2.CrownJobSpec{Instance: clv1alpha2.GenericRef{Name: "inst"}, Command: []string{"true"}},
			}
			Expect(k8sClient.Create(ctx, cj)).To(Succeed())
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{}))
			Expect(crownJobStatus(cj)).To(Equal(clv1alpha2.CrownJobStatus{}))
			var job batchv1.Job
			err := k8sClient.Get(ctx, types.NamespacedName{Namespace: untargetedNamespace, Name: "untargeted"}, &job)
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("Reading the state of the Job", func() {
		It("Should go through Waiting, Starting (downloading the image), Running and Succeeded, with an event per phase", func() {
			cj, job := startJob("inst", sh)

			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseStarting, crownjob.ReasonAdmitted, "Your job is starting.")

			pod := createPod(job, &corev1.PodStatus{
				Phase:                 corev1.PodPending,
				Conditions:            []corev1.PodCondition{{Type: corev1.PodReadyToStartContainers, Status: corev1.ConditionTrue}},
				InitContainerStatuses: []corev1.ContainerStatus{waiting(forge.CrownJobRunFolderContainerName, "PodInitializing")},
				ContainerStatuses:     []corev1.ContainerStatus{waiting(envName, "PodInitializing")},
			})
			// Containers waiting to be created are not a download by themselves (P10: the image was already on the node).
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseStarting, crownjob.ReasonAdmitted, "Your job is starting.")
			createPodEvent(pod, forge.CrownJobRunFolderContainerName, "Pulling", "Pulling image \""+image+"\"", time.Now())
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseStarting, crownjob.ReasonAdmitted, "Your job is starting (downloading the image).")

			// A failing pull is retried by the kubelet, and may never succeed: it does not read as "downloading".
			pod.Status.InitContainerStatuses = []corev1.ContainerStatus{waiting(forge.CrownJobRunFolderContainerName, "ImagePullBackOff")}
			Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseStarting, crownjob.ReasonAdmitted, "Your job is starting (the image cannot be downloaded: retrying).")

			startTime := metav1.NewTime(time.Now().Add(-time.Minute).Truncate(time.Second))
			pod.Status = corev1.PodStatus{
				Phase:                 corev1.PodRunning,
				StartTime:             &startTime,
				Conditions:            []corev1.PodCondition{{Type: corev1.PodReadyToStartContainers, Status: corev1.ConditionTrue}},
				InitContainerStatuses: []corev1.ContainerStatus{terminated(forge.CrownJobRunFolderContainerName, 0)},
				ContainerStatuses:     []corev1.ContainerStatus{running(envName)},
			}
			Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())
			// Still starting, until the Job counts the pod as ready.
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseStarting, crownjob.ReasonAdmitted, "Your job is starting.")

			updateJobStatus(job, func(status *batchv1.JobStatus) {
				status.StartTime = ptr.To(metav1.Now())
				status.Active = 1
				status.Ready = ptr.To[int32](1)
			})
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseRunning, crownjob.ReasonRunning, "Your job is running.")
			Expect(crownJobStatus(cj).Deadline.Time).To(BeTemporally("==", startTime.Add(time.Hour)))

			// The running pod stops being ready while it is killed (e.g., at its time limit): still Running, not Starting.
			updateJobStatus(job, func(status *batchv1.JobStatus) { status.Ready = ptr.To[int32](0) })
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseRunning, crownjob.ReasonRunning, "Your job is running.")
			Expect(crownJobStatus(cj).Deadline.Time).To(BeTemporally("==", startTime.Add(time.Hour)))
			updateJobStatus(job, func(status *batchv1.JobStatus) { status.Ready = ptr.To[int32](1) })
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{}))

			updateJobStatus(job, func(status *batchv1.JobStatus) {
				status.Active = 0
				status.Ready = ptr.To[int32](0)
				status.Succeeded = 1
				status.CompletionTime = ptr.To(metav1.Now())
				status.Conditions = []batchv1.JobCondition{
					jobCondition(batchv1.JobSuccessCriteriaMet, "CompletionsReached", "Reached expected number of succeeded pods"),
					jobCondition(batchv1.JobComplete, "CompletionsReached", "Reached expected number of succeeded pods"),
				}
			})
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseSucceeded, crownjob.ReasonCompleted, "Your job finished successfully.")

			Expect(drainEvents()).To(Equal([]string{
				"Normal InQueue Your job is waiting in the queue of workspace ws.",
				"Normal Admitted Your job is starting.",
				"Normal Running Your job is running.",
				"Normal Completed Your job finished successfully.",
			}))

			// A terminal phase never changes: not even when the Job disappears.
			Expect(k8sClient.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground))).To(Succeed())
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseSucceeded, crownjob.ReasonCompleted, "Your job finished successfully.")
			Expect(drainEvents()).To(BeEmpty())
		})

		It("Should go back to Starting when a new pod replaces the running one (e.g., after an eviction)", func() {
			cj, job := startJob("inst", sh)
			startTime := metav1.NewTime(time.Now().Add(-time.Minute).Truncate(time.Second))
			createPod(job, &corev1.PodStatus{Phase: corev1.PodRunning, StartTime: &startTime,
				InitContainerStatuses: []corev1.ContainerStatus{terminated(forge.CrownJobRunFolderContainerName, 0)},
				ContainerStatuses:     []corev1.ContainerStatus{running(envName)}})
			updateJobStatus(job, func(status *batchv1.JobStatus) {
				status.StartTime = ptr.To(metav1.Now())
				status.Active = 1
				status.Ready = ptr.To[int32](1)
			})
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseRunning, crownjob.ReasonRunning, "Your job is running.")

			// The replacement pod is created later (the newest pod is the one that matters), and has not started yet.
			time.Sleep(1100 * time.Millisecond)
			replacement := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: job.Name + "-replacement", Namespace: job.Namespace,
					Labels: map[string]string{batchv1.ControllerUidLabel: string(job.UID), batchv1.JobNameLabel: job.Name}},
				Spec: *job.Spec.Template.Spec.DeepCopy(),
			}
			Expect(k8sClient.Create(ctx, replacement)).To(Succeed())
			updateJobStatus(job, func(status *batchv1.JobStatus) { status.Ready = ptr.To[int32](0) })

			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseStarting, crownjob.ReasonAdmitted, "Your job is starting.")
			Expect(crownJobStatus(cj).Deadline).To(BeNil())
		})

		It("Should report the deletion of the Job before it ended", func() {
			cj := newCrownJob("inst", sh, "")
			reconcileCrownJob(cj)
			job, err := getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			// Background propagation, as kubectl does: without a policy, a Job is deleted with "orphan", whose finalizer
			// only the garbage collector (absent in envtest) removes.
			Expect(k8sClient.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground))).To(Succeed())
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonError, "The job was deleted before it ended.")
		})

		type failureCase struct {
			jobReason, jobMessage string
			pod                   *corev1.PodStatus
			reason, message       string
		}

		DescribeTable("Mapping a failed Job to a reason and a message",
			func(c failureCase) {
				cj, job := startJob("inst", sh)
				if c.pod != nil {
					createPod(job, c.pod)
				}
				failJob(job, c.jobReason, c.jobMessage)

				Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{}))
				message := strings.ReplaceAll(c.message, "<name>", cj.Name)
				expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, c.reason, message)
				Expect(drainEvents()).To(ContainElement("Warning " + c.reason + " " + message))

				// A terminal phase never changes.
				reconcileCrownJob(cj)
				expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, c.reason, message)
			},
			Entry("the Job reached its activeDeadlineSeconds (no pod left)", failureCase{
				jobReason: batchv1.JobReasonDeadlineExceeded, jobMessage: "Job was active longer than specified deadline",
				reason: crownjob.ReasonTimeLimit, message: "Your job was stopped because it reached its time limit (1h)."}),
			Entry("the pod reached its activeDeadlineSeconds (its container killed with 137)", failureCase{
				jobReason: batchv1.JobReasonPodFailurePolicy, jobMessage: "Container app for pod x failed with exit code 137 matching FailJob rule at index 1",
				pod: &corev1.PodStatus{Phase: corev1.PodFailed, Reason: "DeadlineExceeded", Message: "Pod was active on the node longer than the specified deadline",
					InitContainerStatuses: []corev1.ContainerStatus{terminated(forge.CrownJobRunFolderContainerName, 0)},
					ContainerStatuses:     []corev1.ContainerStatus{terminated(envName, 137)}},
				reason: crownjob.ReasonTimeLimit, message: "Your job was stopped because it reached its time limit (1h)."}),
			Entry("the pod was evicted (its container killed with 137): not an exit code of the command (P2-5)", failureCase{
				jobReason: batchv1.JobReasonPodFailurePolicy, jobMessage: "Container app for pod x failed with exit code 137 matching FailJob rule at index 1",
				pod: &corev1.PodStatus{Phase: corev1.PodFailed, Reason: "Evicted", Message: "The node was low on resource: memory.",
					InitContainerStatuses: []corev1.ContainerStatus{terminated(forge.CrownJobRunFolderContainerName, 0)},
					ContainerStatuses:     []corev1.ContainerStatus{terminated(envName, 137)}},
				reason: crownjob.ReasonCannotStart, message: "The job could not start: Evicted: The node was low on resource: memory."}),
			Entry("the pod was rejected by the kubelet", failureCase{
				jobReason: batchv1.JobReasonBackoffLimitExceeded, jobMessage: "Job has reached the specified backoff limit",
				pod: &corev1.PodStatus{Phase: corev1.PodFailed, Reason: "UnexpectedAdmissionError",
					Message: "Allocate failed due to requested number of devices unavailable for nvidia.com/gpu"},
				reason:  crownjob.ReasonCannotStart,
				message: "The job could not start: UnexpectedAdmissionError: Allocate failed due to requested number of devices unavailable for nvidia.com/gpu."}),
			Entry("the command ended with a non-zero exit code", failureCase{
				jobReason: batchv1.JobReasonPodFailurePolicy, jobMessage: "Container app for pod x failed with exit code 3 matching FailJob rule at index 1",
				pod: &corev1.PodStatus{Phase: corev1.PodFailed,
					InitContainerStatuses: []corev1.ContainerStatus{terminated(forge.CrownJobRunFolderContainerName, 0)},
					ContainerStatuses:     []corev1.ContainerStatus{terminated(envName, 3)}},
				reason: crownjob.ReasonExitCode, message: "Your command ended with exit code 3."}),
			Entry("the run folder could not be created", failureCase{
				jobReason: batchv1.JobReasonBackoffLimitExceeded, jobMessage: "Job has reached the specified backoff limit",
				pod: &corev1.PodStatus{Phase: corev1.PodFailed,
					InitContainerStatuses: []corev1.ContainerStatus{terminated(forge.CrownJobRunFolderContainerName, 1)},
					ContainerStatuses:     []corev1.ContainerStatus{waiting(envName, "PodInitializing")}},
				reason: crownjob.ReasonCannotStart, message: "The folder runs/<name> could not be created on the disk of the instance (exit code 1)."}),
			Entry("anything else (no pod left)", failureCase{
				jobReason: batchv1.JobReasonBackoffLimitExceeded, jobMessage: "Job has reached the specified backoff limit",
				reason: crownjob.ReasonError, message: "The job failed: Job has reached the specified backoff limit."}),
		)
	})
})

func crownJobUID(cj *clv1alpha2.CrownJob) types.UID {
	var current clv1alpha2.CrownJob
	ExpectWithOffset(1, k8sClient.Get(ctx, client.ObjectKeyFromObject(cj), &current)).To(Succeed())
	return current.UID
}
