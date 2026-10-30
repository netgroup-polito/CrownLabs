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
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/crownjob"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("The CrownJob controller, before the Job exists (P10)", func() {
	sh := []string{"sh", "-c", "echo hello"}
	finalizer := forge.CrownJobInstanceFinalizer

	BeforeEach(func() { drainEvents() })

	// newInstanceWithDisk creates a running Instance of the template tpl, with its Deployment and a disk of the given mode.
	newInstanceWithDisk := func(name string, accessMode corev1.PersistentVolumeAccessMode) *clv1alpha2.Instance {
		createInstance(tenantNamespace, name, "tpl", tenantName, false)
		inst, err := getInstance(name)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		createDeployment(inst, "tpl")
		createDisk(inst, accessMode)
		return inst
	}

	// deleteKept deletes an Instance that another finalizer keeps, so that its deletionTimestamp is set.
	deleteKept := func(inst *clv1alpha2.Instance) {
		ExpectWithOffset(1, k8sClient.Get(ctx, client.ObjectKeyFromObject(inst), inst)).To(Succeed())
		inst.Finalizers = append(inst.Finalizers, "test.crownlabs.polito.it/keep")
		ExpectWithOffset(1, k8sClient.Update(ctx, inst)).To(Succeed())
		ExpectWithOffset(1, k8sClient.Delete(ctx, inst)).To(Succeed())
	}

	Describe("Waiting for the disk of the instance", func() {
		It("Should wait with DiskInUse while a ReadWriteOnce disk is used by the instance, then start, and never check again", func() {
			inst := newInstanceWithDisk("inst-rwo", corev1.ReadWriteOnce)
			cj := newCrownJob("inst-rwo", sh, "")

			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonDiskInUse,
				"Your job will start when you stop the instance inst-rwo: its disk can be used by one machine at a time.")
			Expect(drainEvents()).To(ConsistOf("Normal DiskInUse Your job will start when you stop the instance inst-rwo: its disk can be used by one machine at a time."))
			_, err := getJob(cj.Name)
			Expect(kerrors.IsNotFound(err)).To(BeTrue())

			By("Scaling the Deployment to 0 while its pod is still stopping: still waiting")
			var deployment appsv1.Deployment
			Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(inst, envName), &deployment)).To(Succeed())
			deployment.Spec.Replicas = ptr.To[int32](0)
			Expect(k8sClient.Update(ctx, &deployment)).To(Succeed())
			deployment.Status.Replicas = 1
			Expect(k8sClient.Status().Update(ctx, &deployment)).To(Succeed())
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			Expect(crownJobStatus(cj).Reason).To(Equal(crownjob.ReasonDiskInUse))

			By("Once the pod is gone, the Job is created")
			stopInstanceContainer(inst)
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonInQueue, "Your job is waiting in the queue of workspace ws.")
			_, err = getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())

			By("Starting the instance again: the Job exists, so the disk is not checked again")
			Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(inst, envName), &deployment)).To(Succeed())
			deployment.Spec.Replicas = ptr.To[int32](1)
			Expect(k8sClient.Update(ctx, &deployment)).To(Succeed())
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			Expect(crownJobStatus(cj).Reason).To(Equal(crownjob.ReasonInQueue))
		})

		It("Should create the Job at once when the disk is ReadWriteMany, while the instance runs", func() {
			inst := newInstanceWithDisk("inst-rwx", corev1.ReadWriteMany)
			var deployment appsv1.Deployment
			Expect(k8sClient.Get(ctx, forge.NamespacedNameWithSuffix(inst, envName), &deployment)).To(Succeed())
			Expect(deployment.Spec.Replicas).To(PointTo(BeNumerically("==", 1)))

			cj := newCrownJob("inst-rwx", sh, "")
			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonInQueue, "Your job is waiting in the queue of workspace ws.")
		})
	})

	Describe("Waiting for the job queue of the workspace", func() {
		It("Should wait with QueueNotReady while the ClusterQueue does not exist, then create the Job", func() {
			create(workspace("ws-late", &clv1alpha1.QueueLimits{MaxAtOnce: map[string]resource.Quantity{"cpu": resource.MustParse("2")}}))
			createTemplate("tpl-late", "ws-late", &clv1alpha2.TemplateJob{EnvironmentName: envName}, 1, 50)
			createInstance(tenantNamespace, "inst-qlate", "tpl-late", "student-late", true)
			create(tenant("student-late", clv1alpha2.TenantWorkspaceEntry{Name: "ws-late", Role: clv1alpha2.User}))
			cj := newCrownJob("inst-qlate", sh, "")

			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonQueueNotReady,
				"The job queue of workspace ws-late is not ready yet: your job will start when it is.")
			_, err := getJob(cj.Name)
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
			localQueue := &unstructured.Unstructured{}
			localQueue.SetGroupVersionKind(forge.KueueLocalQueueGVK)
			Expect(kerrors.IsNotFound(k8sClient.Get(ctx, types.NamespacedName{Namespace: tenantNamespace, Name: "crownlabs-ws-late"}, localQueue))).To(BeTrue())
			inst, err := getInstance("inst-qlate")
			Expect(err).ToNot(HaveOccurred())
			Expect(inst.Finalizers).ToNot(ContainElement(finalizer))

			By("Creating the ClusterQueue, as the queue controller does when the Workspace gets the target label")
			_, err = queueReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: "ws-late"}})
			Expect(err).ToNot(HaveOccurred())
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{RequeueAfter: 30 * time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseWaiting, crownjob.ReasonInQueue, "Your job is waiting in the queue of workspace ws-late.")
			_, err = getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("The finalizer of the instance", func() {
		It("Should be added right before the Job is created", func() {
			newInstanceWithDisk("inst-fin", corev1.ReadWriteMany)
			inst, err := getInstance("inst-fin")
			Expect(err).ToNot(HaveOccurred())
			Expect(inst.Finalizers).To(BeEmpty())

			cj := newCrownJob("inst-fin", sh, "")
			reconcileCrownJob(cj)
			_, err = getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			inst, err = getInstance("inst-fin")
			Expect(err).ToNot(HaveOccurred())
			Expect(inst.Finalizers).To(ConsistOf(finalizer))
		})

		It("Should not be added when the job cannot start, or waits", func() {
			newInstanceWithDisk("inst-nofin", corev1.ReadWriteOnce)
			cj := newCrownJob("inst-nofin", sh, "2h")
			reconcileCrownJob(cj)
			Expect(crownJobStatus(cj).Reason).To(Equal(crownjob.ReasonCannotStart))
			cj = newCrownJob("inst-nofin", sh, "")
			reconcileCrownJob(cj)
			Expect(crownJobStatus(cj).Reason).To(Equal(crownjob.ReasonDiskInUse))
			inst, err := getInstance("inst-nofin")
			Expect(err).ToNot(HaveOccurred())
			Expect(inst.Finalizers).To(BeEmpty())
		})
	})

	Describe("An instance being deleted", func() {
		It("Should refuse a new job with CannotStart, before checking the Deployment, and add no finalizer", func() {
			createInstance(tenantNamespace, "inst-gone-nodeploy", "tpl", tenantName, false)
			inst, err := getInstance("inst-gone-nodeploy")
			Expect(err).ToNot(HaveOccurred())
			deleteKept(inst)

			cj := newCrownJob("inst-gone-nodeploy", sh, "")
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCannotStart, "The instance inst-gone-nodeploy is being deleted.")
			Expect(drainEvents()).To(ConsistOf("Warning CannotStart The instance inst-gone-nodeploy is being deleted."))
			inst, err = getInstance("inst-gone-nodeploy")
			Expect(err).ToNot(HaveOccurred())
			Expect(inst.Finalizers).ToNot(ContainElement(finalizer))
		})

		It("Should refuse the job when the instance is deleted right after the finalizer is added (the read again)", func() {
			newInstanceWithDisk("inst-race", corev1.ReadWriteMany)
			withWatch, err := client.NewWithWatch(testEnv.Config, client.Options{Scheme: k8sClient.Scheme()})
			Expect(err).ToNot(HaveOccurred())
			// The Instance is deleted as soon as the reconciler has added its finalizer to it.
			racing := interceptor.NewClient(withWatch, interceptor.Funcs{
				Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					if err := c.Patch(ctx, obj, patch, opts...); err != nil {
						return err
					}
					if inst, ok := obj.(*clv1alpha2.Instance); ok {
						deleteKept(inst.DeepCopy())
					}
					return nil
				},
			})
			reconciler := crownJobReconciler
			reconciler.Client = racing

			cj := newCrownJob("inst-race", sh, "")
			_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cj)})
			Expect(err).ToNot(HaveOccurred())
			expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCannotStart, "The instance inst-race is being deleted.")
			_, err = getJob(cj.Name)
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
			inst, err := getInstance("inst-race")
			Expect(err).ToNot(HaveOccurred())
			Expect(inst.DeletionTimestamp).ToNot(BeNil())
		})
	})

	Describe("Writing the status", func() {
		It("Should never overwrite a status written meanwhile by another controller (optimistic locking)", func() {
			newInstanceWithDisk("inst-lock", corev1.ReadWriteMany)
			withWatch, err := client.NewWithWatch(testEnv.Config, client.Options{Scheme: k8sClient.Scheme()})
			Expect(err).ToNot(HaveOccurred())
			// Right before the reconciler writes its status, the instance controller cancels the CrownJob.
			racing := interceptor.NewClient(withWatch, interceptor.Funcs{
				SubResourcePatch: func(ctx context.Context, c client.Client, subResource string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
					if cj, ok := obj.(*clv1alpha2.CrownJob); ok {
						var current clv1alpha2.CrownJob
						Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(cj), &current)).To(Succeed())
						current.Status.Phase = clv1alpha2.CrownJobPhaseFailed
						current.Status.Reason = crownjob.ReasonCanceled
						current.Status.Message = "canceled meanwhile"
						Expect(k8sClient.Status().Update(ctx, &current)).To(Succeed())
					}
					return c.SubResource(subResource).Patch(ctx, obj, patch, opts...)
				},
			})
			reconciler := crownJobReconciler
			reconciler.Client = racing

			cj := newCrownJob("inst-lock", sh, "")
			result, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cj)})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{RequeueAfter: time.Second}))
			expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCanceled, "canceled meanwhile")

			By("The next reconcile deletes the Job created before the cancellation")
			_, err = getJob(cj.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(reconcileCrownJob(cj)).To(Equal(ctrl.Result{}))
			_, err = getJob(cj.Name)
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
			expectPhase(cj, clv1alpha2.CrownJobPhaseFailed, crownjob.ReasonCanceled, "canceled meanwhile")
		})
	})
})
