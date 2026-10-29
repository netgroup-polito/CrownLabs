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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/crownjob"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Saving the log of the job pods (P10)", func() {
	sh := []string{"sh", "-c", "seq 1 100"}

	BeforeEach(func() { drainEvents() })

	// createNodePod simulates the Job controller creating a pod of the Job, the scheduler binding it to a node, and the
	// kubelet writing its status.
	createNodePod := func(job *batchv1.Job, name, node string, containers ...corev1.ContainerStatus) *corev1.Pod {
		ExpectWithOffset(1, k8sClient.Get(ctx, client.ObjectKeyFromObject(job), job)).To(Succeed())
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: name, Namespace: job.Namespace,
				Labels: map[string]string{batchv1.ControllerUidLabel: string(job.UID), batchv1.JobNameLabel: job.Name},
			},
			Spec: *job.Spec.Template.Spec.DeepCopy(),
		}
		pod.Spec.NodeName = node
		ExpectWithOffset(1, k8sClient.Create(ctx, pod)).To(Succeed())
		pod.Status = corev1.PodStatus{Phase: corev1.PodRunning, StartTime: ptr.To(metav1.Now()), ContainerStatuses: containers,
			InitContainerStatuses: []corev1.ContainerStatus{terminated(forge.CrownJobRunFolderContainerName, 0)}}
		ExpectWithOffset(1, k8sClient.Status().Update(ctx, pod)).To(Succeed())
		return pod
	}

	setReady := func(job *batchv1.Job, ready int32) {
		updateJobStatus(job, func(status *batchv1.JobStatus) {
			// Set once, as the Job controller does: the API refuses to change it on an unsuspended Job.
			if status.StartTime == nil {
				status.StartTime = ptr.To(metav1.Now())
			}
			status.Active = 1
			status.Ready = ptr.To(ready)
		})
	}

	helpersOf := func(cj *clv1alpha2.CrownJob) []batchv1.Job {
		var jobs batchv1.JobList
		ExpectWithOffset(1, k8sClient.List(ctx, &jobs, client.InNamespace(tenantNamespace),
			client.MatchingLabels{forge.LabelCrownJobKey: cj.Name, forge.LabelCrownJobComponentKey: forge.CrownJobComponentLogHelper})).To(Succeed())
		return jobs.Items
	}

	current := func(cj *clv1alpha2.CrownJob) *clv1alpha2.CrownJob {
		var c clv1alpha2.CrownJob
		ExpectWithOffset(1, k8sClient.Get(ctx, client.ObjectKeyFromObject(cj), &c)).To(Succeed())
		return &c
	}

	// failHelper and completeHelper simulate the Job controller ending a helper Job.
	failHelper := func(helper *batchv1.Job) {
		updateJobStatus(helper, func(status *batchv1.JobStatus) {
			status.StartTime = ptr.To(metav1.Now())
			status.Failed = 3
			status.Conditions = []batchv1.JobCondition{
				jobCondition(batchv1.JobFailureTarget, "BackoffLimitExceeded", "Job has reached the specified backoff limit"),
				jobCondition(batchv1.JobFailed, "BackoffLimitExceeded", "Job has reached the specified backoff limit"),
			}
		})
	}
	completeHelper := func(helper *batchv1.Job) {
		updateJobStatus(helper, func(status *batchv1.JobStatus) {
			status.StartTime = ptr.To(metav1.Now())
			status.CompletionTime = ptr.To(metav1.Now())
			status.Succeeded = 1
			status.Conditions = []batchv1.JobCondition{
				jobCondition(batchv1.JobSuccessCriteriaMet, "CompletionsReached", "Reached expected number of succeeded pods"),
				jobCondition(batchv1.JobComplete, "CompletionsReached", "Reached expected number of succeeded pods"),
			}
		})
	}

	It("Should create one helper for a running pod, with the permissions to read its log", func() {
		cj, job := startJob("inst", sh)
		createNodePod(job, job.Name+"-a", "node-1", running(envName))
		setReady(job, 1)
		reconcileCrownJob(cj)
		Expect(crownJobStatus(cj).Phase).To(Equal(clv1alpha2.CrownJobPhaseRunning))

		helpers := helpersOf(cj)
		Expect(helpers).To(HaveLen(1))
		helper := helpers[0]
		Expect(helper.Name).To(Equal(forge.CrownJobLogHelperName(current(cj), 1)))
		Expect(metav1.IsControlledBy(&helper, current(cj))).To(BeTrue())
		Expect(helper.Annotations).To(HaveKeyWithValue(forge.AnnotationCrownJobPodKey, job.Name+"-a"))
		Expect(helper.Annotations).To(HaveKeyWithValue(forge.AnnotationCrownJobLogFileKey, "runs/"+cj.Name+"/log.txt"))
		Expect(helper.Labels).ToNot(HaveKey(forge.KueueQueueNameLabel))
		Expect(helper.Spec.Template.Spec.Containers[0].Image).To(Equal(logHelperImage))
		Expect(helper.Spec.Template.Spec.Affinity).To(BeNil(), "the disk of the fixture is ReadWriteMany")
		Expect(helper.Spec.TTLSecondsAfterFinished).To(Equal(ptr.To[int32](600)))
		mainJob, err := getJob(cj.Name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mainJob.Annotations).To(HaveKeyWithValue(forge.AnnotationCrownJobLogFilesKey, "1"))

		var sa corev1.ServiceAccount
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: tenantNamespace, Name: forge.CrownJobLogReaderName}, &sa)).To(Succeed())
		Expect(sa.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "crownjob"))
		var role rbacv1.Role
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: tenantNamespace, Name: forge.CrownJobLogReaderName}, &role)).To(Succeed())
		Expect(role.Rules).To(Equal(forge.CrownJobLogReaderRole(tenantNamespace).Rules))
		var rb rbacv1.RoleBinding
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: tenantNamespace, Name: forge.CrownJobLogReaderName}, &rb)).To(Succeed())
		Expect(rb.Subjects).To(Equal(forge.CrownJobLogReaderRoleBinding(tenantNamespace).Subjects))

		By("Reconciling again: no second helper for the same pod")
		reconcileCrownJob(cj)
		Expect(helpersOf(cj)).To(HaveLen(1))
		Expect(drainEvents()).To(ConsistOf("Normal InQueue Your job is waiting in the queue of workspace ws.", "Normal Running Your job is running."))

		By("The helper ends: the CrownJob is not affected (the status reads the main Job only)")
		completeHelper(&helper)
		reconcileCrownJob(cj)
		expectPhase(cj, clv1alpha2.CrownJobPhaseRunning, crownjob.ReasonRunning, "Your job is running.")

		By("The main Job is deleted: the job was deleted before it ended, and its output was saved")
		// Background propagation, as kubectl: the default for a Job orphans, and envtest has no garbage collector (P9-6).
		Expect(k8sClient.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground))).To(Succeed())
		reconcileCrownJob(cj)
		expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonError,
			fmt.Sprintf("The job was deleted before it ended. The output is saved in runs/%s/log.txt.", cj.Name))
	})

	It("Should run the helper on the node of the pod when the disk is ReadWriteOnce", func() {
		createInstance(tenantNamespace, "inst-log-rwo", "tpl", tenantName, false)
		inst, err := getInstance("inst-log-rwo")
		Expect(err).ToNot(HaveOccurred())
		createDeployment(inst, "tpl")
		createDisk(inst, corev1.ReadWriteOnce)
		stopInstanceContainer(inst)

		cj, job := startJob("inst-log-rwo", sh)
		createNodePod(job, job.Name+"-a", "node-7", running(envName))
		setReady(job, 1)
		reconcileCrownJob(cj)
		helpers := helpersOf(cj)
		Expect(helpers).To(HaveLen(1))
		Expect(helpers[0].Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(Equal(
			[]corev1.NodeSelectorTerm{{MatchFields: []corev1.NodeSelectorRequirement{{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"node-7"}}}}}))
	})

	It("Should give a replacement pod its own helper (log-2.txt), and mention log.txt when the job ends", func() {
		cj, job := startJob("inst", sh)
		first := createNodePod(job, job.Name+"-a", "node-1", running(envName))
		setReady(job, 1)
		reconcileCrownJob(cj)
		Expect(helpersOf(cj)).To(HaveLen(1))

		By("The first pod is evicted and replaced")
		first.Status.ContainerStatuses = []corev1.ContainerStatus{terminated(envName, 137)}
		first.Status.Phase = corev1.PodFailed
		Expect(k8sClient.Status().Update(ctx, first)).To(Succeed())
		setReady(job, 0)
		replacement := createNodePod(job, job.Name+"-b", "node-2", waiting(envName, "ContainerCreating"))
		// The replacement starts later than the first pod (the deadline tells the pods apart).
		replacement.Status.StartTime = ptr.To(metav1.NewTime(time.Now().Add(time.Minute)))
		Expect(k8sClient.Status().Update(ctx, replacement)).To(Succeed())
		reconcileCrownJob(cj)
		Expect(crownJobStatus(cj).Phase).To(Equal(clv1alpha2.CrownJobPhaseStarting))
		Expect(helpersOf(cj)).To(HaveLen(1), "the replacement has not started yet")

		var second corev1.Pod
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: tenantNamespace, Name: job.Name + "-b"}, &second)).To(Succeed())
		second.Status.ContainerStatuses = []corev1.ContainerStatus{running(envName)}
		Expect(k8sClient.Status().Update(ctx, &second)).To(Succeed())
		setReady(job, 1)
		reconcileCrownJob(cj)
		helpers := helpersOf(cj)
		Expect(helpers).To(HaveLen(2))
		files := map[string]string{}
		for i := range helpers {
			files[helpers[i].Annotations[forge.AnnotationCrownJobPodKey]] = helpers[i].Annotations[forge.AnnotationCrownJobLogFileKey]
		}
		Expect(files).To(Equal(map[string]string{
			job.Name + "-a": "runs/" + cj.Name + "/log.txt",
			job.Name + "-b": "runs/" + cj.Name + "/log-2.txt",
		}))

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
		reconcileCrownJob(cj)
		expectPhase(cj, clv1alpha2.CrownJobPhaseSucceeded, crownjob.ReasonCompleted,
			fmt.Sprintf("Your job finished successfully. The output is saved in runs/%s/log.txt.", cj.Name))
	})

	It("Should save the log of a pod that ended before it was seen running, and say so in the terminal message", func() {
		cj, job := startJob("inst", sh)
		createNodePod(job, job.Name+"-a", "node-1", terminated(envName, 3))
		failJob(job, "PodFailurePolicy", "Container "+envName+" for pod tenant-student/x failed with exit code 3 matching FailJob rule at index 1")
		reconcileCrownJob(cj)
		expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonExitCode,
			fmt.Sprintf("Your command ended with exit code 3. The output is saved in runs/%s/log.txt.", cj.Name))
		Expect(helpersOf(cj)).To(HaveLen(1))
	})

	It("Should create no helper for a pod that never started, and not mention a log", func() {
		cj, job := startJob("inst", sh)
		createNodePod(job, job.Name+"-a", "node-1", waiting(envName, "ContainerCreating"))
		reconcileCrownJob(cj)
		Expect(crownJobStatus(cj).Phase).To(Equal(clv1alpha2.CrownJobPhaseStarting))
		failJob(job, "DeadlineExceeded", "Job was active longer than specified deadline")
		reconcileCrownJob(cj)
		expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonTimeLimit, "Your job was stopped because it reached its time limit (1h).")
		Expect(helpersOf(cj)).To(BeEmpty())
	})

	It("Should report a failed helper once, with LogsNotSaved, without changing the phase, also after the job ended", func() {
		cj, job := startJob("inst", sh)
		createNodePod(job, job.Name+"-a", "node-1", running(envName))
		setReady(job, 1)
		reconcileCrownJob(cj)
		helpers := helpersOf(cj)
		Expect(helpers).To(HaveLen(1))
		drainEvents()

		failHelper(&helpers[0])
		reconcileCrownJob(cj)
		expectPhase(cj, clv1alpha2.CrownJobPhaseRunning, crownjob.ReasonRunning, "Your job is running.")
		Expect(drainEvents()).To(ConsistOf(fmt.Sprintf(
			"Warning LogsNotSaved The output of the job could not be saved in runs/%s/log.txt: Job has reached the specified backoff limit.", cj.Name)))
		reconcileCrownJob(cj)
		Expect(drainEvents()).To(BeEmpty())

		By("A helper of a job that has already ended")
		cj2, job2 := startJob("inst", sh)
		createNodePod(job2, job2.Name+"-a", "node-1", terminated(envName, 0))
		updateJobStatus(job2, func(status *batchv1.JobStatus) {
			status.StartTime = ptr.To(metav1.Now())
			status.CompletionTime = ptr.To(metav1.Now())
			status.Succeeded = 1
			status.Conditions = []batchv1.JobCondition{
				jobCondition(batchv1.JobSuccessCriteriaMet, "CompletionsReached", "Reached expected number of succeeded pods"),
				jobCondition(batchv1.JobComplete, "CompletionsReached", "Reached expected number of succeeded pods"),
			}
		})
		reconcileCrownJob(cj2)
		Expect(crownJobStatus(cj2).Phase).To(Equal(clv1alpha2.CrownJobPhaseSucceeded))
		helpers2 := helpersOf(cj2)
		Expect(helpers2).To(HaveLen(1))
		drainEvents()
		failHelper(&helpers2[0])
		reconcileCrownJob(cj2)
		Expect(crownJobStatus(cj2).Phase).To(Equal(clv1alpha2.CrownJobPhaseSucceeded))
		Expect(drainEvents()).To(ConsistOf(fmt.Sprintf(
			"Warning LogsNotSaved The output of the job could not be saved in runs/%s/log.txt: Job has reached the specified backoff limit.", cj2.Name)))
	})

	Context("When the finished Jobs are deleted after their TTL (P11)", func() {
		completeJob := func(job *batchv1.Job) {
			updateJobStatus(job, func(status *batchv1.JobStatus) {
				if status.StartTime == nil {
					status.StartTime = ptr.To(metav1.Now())
				}
				status.Active = 0
				status.Ready = ptr.To[int32](0)
				status.Succeeded = 1
				status.CompletionTime = ptr.To(metav1.Now())
				status.Conditions = []batchv1.JobCondition{
					jobCondition(batchv1.JobSuccessCriteriaMet, "CompletionsReached", "Reached expected number of succeeded pods"),
					jobCondition(batchv1.JobComplete, "CompletionsReached", "Reached expected number of succeeded pods"),
				}
			})
		}

		It("Should keep the terminal phase, message and log sentence once the Job, its pod and its helper are gone", func() {
			cj, job := startJob("inst", sh)
			pod := createNodePod(job, job.Name+"-a", "node-1", running(envName))
			setReady(job, 1)
			reconcileCrownJob(cj)
			helpers := helpersOf(cj)
			Expect(helpers).To(HaveLen(1))

			completeHelper(&helpers[0])
			pod.Status.Phase = corev1.PodSucceeded
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{terminated(envName, 0)}
			Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())
			completeJob(job)
			reconcileCrownJob(cj)
			message := fmt.Sprintf("Your job finished successfully. The output is saved in runs/%s/log.txt.", cj.Name)
			expectPhase(cj, clv1alpha2.CrownJobPhaseSucceeded, crownjob.ReasonCompleted, message)
			drainEvents()

			By("The TTL controller deletes the Job with foreground propagation: first the Job is only marked")
			Expect(k8sClient.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationForeground))).To(Succeed())
			marked, err := getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(marked.DeletionTimestamp).ToNot(BeNil())
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseSucceeded, crownjob.ReasonCompleted, message)

			By("The pod goes, then the Job (envtest has no garbage collector), and the helper after its own TTL")
			Expect(k8sClient.Delete(ctx, pod, client.GracePeriodSeconds(0))).To(Succeed())
			marked.Finalizers = nil
			Expect(k8sClient.Update(ctx, marked)).To(Succeed())
			Expect(k8sClient.Delete(ctx, &helpers[0], client.PropagationPolicy(metav1.DeletePropagationBackground))).To(Succeed())
			Eventually(func() bool { _, err := getJob(cj.Name); return kerrors.IsNotFound(err) }).Should(BeTrue())
			Expect(helpersOf(cj)).To(BeEmpty())

			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseSucceeded, crownjob.ReasonCompleted, message)
			Expect(crownJobStatus(cj).RunFolder).To(Equal("runs/" + cj.Name))
			Expect(drainEvents()).To(BeEmpty())
			Expect(helpersOf(cj)).To(BeEmpty())
			_, err = getJob(cj.Name)
			Expect(kerrors.IsNotFound(err)).To(BeTrue(), "the Job is never created again")
		})

		It("Should number the log of a replacement pod, and keep the log sentence, after the first helper was deleted", func() {
			cj, job := startJob("inst", sh)
			first := createNodePod(job, job.Name+"-a", "node-1", running(envName))
			setReady(job, 1)
			reconcileCrownJob(cj)
			helpers := helpersOf(cj)
			Expect(helpers).To(HaveLen(1))

			By("The first pod is evicted; its helper ends and is deleted after its TTL, before the replacement starts")
			first.Status.ContainerStatuses = []corev1.ContainerStatus{terminated(envName, 137)}
			first.Status.Phase = corev1.PodFailed
			Expect(k8sClient.Status().Update(ctx, first)).To(Succeed())
			setReady(job, 0)
			completeHelper(&helpers[0])
			Expect(k8sClient.Delete(ctx, &helpers[0], client.PropagationPolicy(metav1.DeletePropagationBackground))).To(Succeed())
			Expect(helpersOf(cj)).To(BeEmpty())

			replacement := createNodePod(job, job.Name+"-b", "node-2", running(envName))
			replacement.Status.StartTime = ptr.To(metav1.NewTime(time.Now().Add(time.Minute)))
			Expect(k8sClient.Status().Update(ctx, replacement)).To(Succeed())
			setReady(job, 1)
			reconcileCrownJob(cj)
			helpers = helpersOf(cj)
			Expect(helpers).To(HaveLen(1))
			Expect(helpers[0].Name).To(Equal(forge.CrownJobLogHelperName(current(cj), 2)))
			Expect(helpers[0].Annotations).To(HaveKeyWithValue(forge.AnnotationCrownJobPodKey, job.Name+"-b"))
			Expect(helpers[0].Annotations).To(HaveKeyWithValue(forge.AnnotationCrownJobLogFileKey, "runs/"+cj.Name+"/log-2.txt"))
			mainJob, err := getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mainJob.Annotations).To(HaveKeyWithValue(forge.AnnotationCrownJobLogFilesKey, "2"))

			completeJob(job)
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseSucceeded, crownjob.ReasonCompleted,
				fmt.Sprintf("Your job finished successfully. The output is saved in runs/%s/log.txt.", cj.Name))
		})

		It("Should create no helper once the Job is gone, even for a pod left behind (the operator missed the end)", func() {
			cj, job := startJob("inst", sh)
			createNodePod(job, job.Name+"-a", "node-1", terminated(envName, 0))
			completeJob(job)
			Expect(k8sClient.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground))).To(Succeed())
			reconcileCrownJob(cj)
			// The known limit: with the operator down for longer than the TTL across the end of the Job, the end is lost.
			expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonError, "The job was deleted before it ended.")
			Expect(helpersOf(cj)).To(BeEmpty())
		})
	})

	It("Should not take over a Job with the name of the helper that it does not control", func() {
		cj, job := startJob("inst", sh)
		name := forge.CrownJobLogHelperName(current(cj), 1)
		foreign := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: tenantNamespace},
			Spec: batchv1.JobSpec{Suspend: ptr.To(true), Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever, Containers: []corev1.Container{{Name: "c", Image: "busybox"}},
			}}},
		}
		create(foreign)
		createNodePod(job, job.Name+"-a", "node-1", running(envName))
		setReady(job, 1)
		drainEvents()
		reconcileCrownJob(cj)
		Expect(crownJobStatus(cj).Phase).To(Equal(clv1alpha2.CrownJobPhaseRunning))
		Expect(drainEvents()).To(ContainElement(fmt.Sprintf(
			"Warning LogsNotSaved The output of the job could not be saved in runs/%s/log.txt: Another job named %s already exists in this namespace.", cj.Name, name)))
		Expect(helpersOf(cj)).To(BeEmpty())
	})
})
