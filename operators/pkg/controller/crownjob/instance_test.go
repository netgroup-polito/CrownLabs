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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/crownjob"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("The instance controller of the CrownJobs (crownjob-instance)", func() {
	sh := []string{"sh", "-c", "sleep 60"}
	finalizer := forge.CrownJobInstanceFinalizer
	const (
		timeout  = 10 * time.Second
		interval = 100 * time.Millisecond
	)

	BeforeEach(func() { drainEvents() })

	// reconcileInstance reconciles the Instance until the given condition holds: the reconciler reads through a cache.
	reconcileInstanceUntil := func(name string, condition func() bool) {
		EventuallyWithOffset(1, func() bool {
			_, err := instanceReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: tenantNamespace, Name: name}})
			Expect(err).ToNot(HaveOccurred())
			return condition()
		}, timeout, interval).Should(BeTrue())
	}

	hasFinalizer := func(name string) func() bool {
		return func() bool {
			inst, err := getInstance(name)
			return err == nil && len(inst.Finalizers) > 0 && inst.Finalizers[0] == finalizer
		}
	}

	// completeJob simulates the Job controller ending the Job successfully, and lets the CrownJob reconciler see it.
	completeJob := func(cj *clv1alpha2.CrownJob) {
		job, err := getJob(cj.Name)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		updateJobStatus(job, func(status *batchv1.JobStatus) {
			status.StartTime = ptr.To(metav1.Now())
			status.CompletionTime = ptr.To(metav1.Now())
			status.Succeeded = 1
			status.Conditions = []batchv1.JobCondition{
				jobCondition(batchv1.JobSuccessCriteriaMet, "CompletionsReached", "Reached expected number of succeeded pods"),
				jobCondition(batchv1.JobComplete, "CompletionsReached", "Reached expected number of succeeded pods"),
			}
		})
		reconcileCrownJob(cj)
		ExpectWithOffset(1, crownJobStatus(cj).Phase).To(Equal(clv1alpha2.CrownJobPhaseSucceeded))
	}

	It("Should keep the finalizer while a job is not terminal, and remove it once every job is", func() {
		createInstance(tenantNamespace, "inst-keep", "tpl", tenantName, true)
		cj1, _ := startJob("inst-keep", sh)
		cj2 := newCrownJob("inst-keep", sh, "")
		reconcileCrownJob(cj2)
		Expect(hasFinalizer("inst-keep")()).To(BeTrue())

		By("Reconciling the instance while both jobs run: the finalizer stays")
		reconcileInstanceUntil("inst-keep", func() bool { return true })
		Consistently(hasFinalizer("inst-keep"), time.Second, interval).Should(BeTrue())

		By("Ending one job: the finalizer stays")
		completeJob(cj1)
		for range 5 {
			_, err := instanceReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: tenantNamespace, Name: "inst-keep"}})
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(interval)
		}
		Expect(hasFinalizer("inst-keep")()).To(BeTrue())

		By("Ending the other: the finalizer is removed, and the instance stays (it is not being deleted)")
		completeJob(cj2)
		reconcileInstanceUntil("inst-keep", func() bool { return !hasFinalizer("inst-keep")() })
		_, err := getInstance("inst-keep")
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should cancel the jobs not admitted yet of an instance being deleted, and let the admitted ones run until they end", func() {
		createInstance(tenantNamespace, "inst-del", "tpl", tenantName, true)
		admitted, _ := startJob("inst-del", sh)
		reconcileCrownJob(admitted)
		Expect(crownJobStatus(admitted).Phase).To(Equal(clv1alpha2.CrownJobPhaseStarting))
		queued := newCrownJob("inst-del", sh, "")
		reconcileCrownJob(queued)
		Expect(crownJobStatus(queued).Reason).To(Equal(crownjob.ReasonInQueue))
		// A job just admitted by Kueue, whose CrownJob still reads Waiting: it is not canceled.
		justAdmitted := newCrownJob("inst-del", sh, "")
		reconcileCrownJob(justAdmitted)
		job, err := getJob(justAdmitted.Name)
		Expect(err).ToNot(HaveOccurred())
		admit(job)
		Expect(crownJobStatus(justAdmitted).Phase).To(Equal(clv1alpha2.CrownJobPhaseWaiting))
		// A job waiting before its Job exists (e.g., for the disk).
		noJob := newCrownJob("inst-del", sh, "")
		var current clv1alpha2.CrownJob
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(noJob), &current)).To(Succeed())
		current.Status = clv1alpha2.CrownJobStatus{Phase: clv1alpha2.CrownJobPhaseWaiting, Reason: crownjob.ReasonDiskInUse, Message: "waiting"}
		Expect(k8sClient.Status().Update(ctx, &current)).To(Succeed())
		// A job its reconciler has not seen yet: left to it.
		unseen := newCrownJob("inst-del", sh, "")
		Expect(hasFinalizer("inst-del")()).To(BeTrue())

		By("Deleting the instance: it is kept by the finalizer")
		drainEvents()
		inst, err := getInstance("inst-del")
		Expect(err).ToNot(HaveOccurred())
		Expect(k8sClient.Delete(ctx, inst)).To(Succeed())
		inst, err = getInstance("inst-del")
		Expect(err).ToNot(HaveOccurred())
		Expect(inst.DeletionTimestamp).ToNot(BeNil())

		reconcileInstanceUntil("inst-del", func() bool {
			return crownJobStatus(queued).Phase == clv1alpha2.CrownJobPhaseFailed && crownJobStatus(noJob).Phase == clv1alpha2.CrownJobPhaseFailed
		})
		message := "The instance inst-del was deleted before your job started."
		expectPhase(queued, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCanceled, message)
		expectPhase(noJob, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCanceled, message)
		_, err = getJob(queued.Name)
		Expect(kerrors.IsNotFound(err)).To(BeTrue(), "the Job of the canceled job is deleted")
		Expect(drainEvents()).To(ConsistOf("Warning Canceled "+message, "Warning Canceled "+message))

		Expect(crownJobStatus(admitted).Phase).To(Equal(clv1alpha2.CrownJobPhaseStarting))
		Expect(crownJobStatus(justAdmitted).Phase).To(Equal(clv1alpha2.CrownJobPhaseWaiting))
		_, err = getJob(justAdmitted.Name)
		Expect(err).ToNot(HaveOccurred())
		Expect(crownJobStatus(unseen).Phase).To(BeEmpty())
		Expect(hasFinalizer("inst-del")()).To(BeTrue())

		By("The job its reconciler had not seen is refused by it")
		reconcileCrownJob(unseen)
		expectPhase(unseen, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCannotStart, "The instance inst-del is being deleted.")
		reconcileCrownJob(justAdmitted)
		Expect(crownJobStatus(justAdmitted).Phase).To(Equal(clv1alpha2.CrownJobPhaseStarting))

		By("Ending the admitted jobs: the finalizer is removed, and the instance is gone")
		completeJob(admitted)
		for range 5 {
			_, err := instanceReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: tenantNamespace, Name: "inst-del"}})
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(interval)
		}
		Expect(hasFinalizer("inst-del")()).To(BeTrue(), "one job is still running")
		completeJob(justAdmitted)
		reconcileInstanceUntil("inst-del", func() bool {
			_, err := getInstance("inst-del")
			return kerrors.IsNotFound(err)
		})
	})

	It("Should remove the finalizer when the running jobs are deleted (the force path)", func() {
		createInstance(tenantNamespace, "inst-force", "tpl", tenantName, true)
		cj, _ := startJob("inst-force", sh)
		reconcileCrownJob(cj)
		inst, err := getInstance("inst-force")
		Expect(err).ToNot(HaveOccurred())
		Expect(k8sClient.Delete(ctx, inst)).To(Succeed())
		reconcileInstanceUntil("inst-force", func() bool { return true })
		Consistently(hasFinalizer("inst-force"), time.Second, interval).Should(BeTrue())

		Expect(k8sClient.Delete(ctx, cj)).To(Succeed())
		reconcileInstanceUntil("inst-force", func() bool {
			_, err := getInstance("inst-force")
			return kerrors.IsNotFound(err)
		})
	})

	It("Should remove a finalizer left on an instance without jobs, and leave untargeted namespaces alone", func() {
		createInstance(tenantNamespace, "inst-leftover", "tpl", tenantName, true)
		inst, err := getInstance("inst-leftover")
		Expect(err).ToNot(HaveOccurred())
		inst.Finalizers = []string{finalizer}
		Expect(k8sClient.Update(ctx, inst)).To(Succeed())
		reconcileInstanceUntil("inst-leftover", func() bool { return !hasFinalizer("inst-leftover")() })

		var untargeted clv1alpha2.Instance
		key := types.NamespacedName{Namespace: untargetedNamespace, Name: "inst"}
		Expect(k8sClient.Get(ctx, key, &untargeted)).To(Succeed())
		untargeted.Finalizers = []string{finalizer}
		Expect(k8sClient.Update(ctx, &untargeted)).To(Succeed())
		Eventually(func() error {
			_, err := instanceReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: key})
			return err
		}, timeout, interval).Should(Succeed())
		Consistently(func() []string {
			_, _ = instanceReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: key})
			Expect(k8sClient.Get(ctx, key, &untargeted)).To(Succeed())
			return untargeted.Finalizers
		}, time.Second, interval).Should(ConsistOf(finalizer))
		untargeted.Finalizers = nil
		Expect(k8sClient.Update(ctx, &untargeted)).To(Succeed())
	})

	It("Should map a CrownJob to its instance, in the namespace of the CrownJob", func() {
		cj := &clv1alpha2.CrownJob{
			ObjectMeta: metav1.ObjectMeta{Name: "cj", Namespace: tenantNamespace},
			Spec:       clv1alpha2.CrownJobSpec{Instance: clv1alpha2.GenericRef{Name: "inst-x", Namespace: otherNamespace}},
		}
		Expect(crownjob.CrownJobToInstance(ctx, cj)).To(Equal([]reconcile.Request{
			{NamespacedName: types.NamespacedName{Namespace: tenantNamespace, Name: "inst-x"}},
		}))
		Expect(crownjob.CrownJobToInstance(ctx, &clv1alpha2.CrownJob{})).To(BeEmpty())
	})
})
