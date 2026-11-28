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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("The CrownJob queue controller", func() {
	// These tests use their own Workspaces, since the order of the top-level containers is random.

	reconcileWorkspace := func(name string) {
		result, err := queueReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: name}})
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		ExpectWithOffset(1, result).To(Equal(ctrl.Result{}))
	}

	getKueueObject := func(obj *unstructured.Unstructured, name string) error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: name}, obj)
	}

	clusterQueue := func(name string) (*unstructured.Unstructured, error) {
		queue := &unstructured.Unstructured{}
		queue.SetGroupVersionKind(forge.KueueClusterQueueGVK)
		return queue, getKueueObject(queue, name)
	}

	priorityClass := func(name string) (*unstructured.Unstructured, error) {
		class := &unstructured.Unstructured{}
		class.SetGroupVersionKind(forge.KueueWorkloadPriorityClassGVK)
		return class, getKueueObject(class, name)
	}

	nominalQuotas := func(queue *unstructured.Unstructured) map[string]string {
		groups, _, err := unstructured.NestedSlice(queue.Object, "spec", "resourceGroups")
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		flavors := groups[0].(map[string]interface{})["flavors"].([]interface{})
		quotas := map[string]string{}
		for _, r := range flavors[0].(map[string]interface{})["resources"].([]interface{}) {
			res := r.(map[string]interface{})
			quotas[res["name"].(string)] = res["nominalQuota"].(string)
		}
		return quotas
	}

	It("Should create, update and delete the ClusterQueue with jobOptions, and never delete the ResourceFlavor", func() {
		ws := workspace("ws-queue", &clv1alpha1.QueueLimits{MaxAtOnce: map[string]resource.Quantity{
			"cpu": resource.MustParse("2"), gpu: resource.MustParse("1"),
		}})
		create(ws)

		By("Creating the ClusterQueue and the ResourceFlavor")
		reconcileWorkspace("ws-queue")
		queue, err := clusterQueue("crownlabs-ws-queue")
		Expect(err).ToNot(HaveOccurred())
		Expect(queue.GetOwnerReferences()).To(HaveLen(1))
		Expect(queue.GetOwnerReferences()[0].Name).To(Equal("ws-queue"))
		Expect(queue.GetManagedFields()).To(ContainElement(HaveField("Manager", forge.CrownJobFieldManager)))
		Expect(nominalQuotas(queue)).To(Equal(map[string]string{"cpu": "2", "memory": "100Ti", gpu: "1"}))
		covered, _, _ := unstructured.NestedSlice(queue.Object, "spec", "resourceGroups")
		Expect(covered[0].(map[string]interface{})["coveredResources"]).To(Equal([]interface{}{"cpu", "memory", gpu}))
		strategy, _, _ := unstructured.NestedString(queue.Object, "spec", "queueingStrategy")
		Expect(strategy).To(Equal("BestEffortFIFO"))
		// Fields defaulted by the CRD are kept: no preemption (Never) and no cohort.
		cohort, found, _ := unstructured.NestedString(queue.Object, "spec", "cohortName")
		Expect(found && cohort != "").To(BeFalse())

		flavor := &unstructured.Unstructured{}
		flavor.SetGroupVersionKind(forge.KueueResourceFlavorGVK)
		Expect(getKueueObject(flavor, "crownlabs-default")).To(Succeed())

		By("Updating the ClusterQueue when maxAtOnce changes")
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), ws)).To(Succeed())
		ws.Spec.JobOptions.QueueLimits.MaxAtOnce = map[string]resource.Quantity{"cpu": resource.MustParse("3"), "memory": resource.MustParse("6Gi")}
		Expect(k8sClient.Update(ctx, ws)).To(Succeed())
		reconcileWorkspace("ws-queue")
		queue, err = clusterQueue("crownlabs-ws-queue")
		Expect(err).ToNot(HaveOccurred())
		Expect(nominalQuotas(queue)).To(Equal(map[string]string{"cpu": "3", "memory": "6Gi"}))

		By("Deleting the ClusterQueue when jobOptions is removed, keeping the ResourceFlavor")
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), ws)).To(Succeed())
		ws.Spec.JobOptions = nil
		Expect(k8sClient.Update(ctx, ws)).To(Succeed())
		reconcileWorkspace("ws-queue")
		_, err = clusterQueue("crownlabs-ws-queue")
		Expect(kerrors.IsNotFound(err)).To(BeTrue())
		Expect(getKueueObject(flavor, "crownlabs-default")).To(Succeed())

		By("Re-creating the ClusterQueue when jobOptions is set again")
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), ws)).To(Succeed())
		ws.Spec.JobOptions = &clv1alpha1.JobOptions{QueueLimits: clv1alpha1.QueueLimits{MaxAtOnce: map[string]resource.Quantity{"cpu": resource.MustParse("1")}}}
		Expect(k8sClient.Update(ctx, ws)).To(Succeed())
		reconcileWorkspace("ws-queue")
		queue, err = clusterQueue("crownlabs-ws-queue")
		Expect(err).ToNot(HaveOccurred())
		Expect(nominalQuotas(queue)).To(Equal(map[string]string{"cpu": "1", "memory": "100Ti"}))
	})

	It("Should create neither a ClusterQueue nor a WorkloadPriorityClass for a Workspace without jobOptions", func() {
		create(workspace("ws-queue-none", nil))
		reconcileWorkspace("ws-queue-none")
		_, err := clusterQueue("crownlabs-ws-queue-none")
		Expect(kerrors.IsNotFound(err)).To(BeTrue())
		_, err = priorityClass("workspace-ws-queue-none")
		Expect(kerrors.IsNotFound(err)).To(BeTrue())
	})

	It("Should create, update, restore and delete the WorkloadPriorityClass of a Workspace with jobOptions", func() {
		ws := workspace("ws-class", &clv1alpha1.QueueLimits{MaxAtOnce: map[string]resource.Quantity{"cpu": resource.MustParse("2")}})
		create(ws)
		// A class of another name, which the controller must never touch.
		other := &unstructured.Unstructured{}
		other.SetGroupVersionKind(forge.KueueWorkloadPriorityClassGVK)
		other.SetName("workspace-ws-class-other")
		other.Object["value"] = int64(7)
		create(other)
		otherVersion := other.GetResourceVersion()

		By("Creating the class with the ClusterQueue: the default priority, owned by the Workspace")
		reconcileWorkspace("ws-class")
		_, err := clusterQueue("crownlabs-ws-class")
		Expect(err).ToNot(HaveOccurred())
		class, err := priorityClass("workspace-ws-class")
		Expect(err).ToNot(HaveOccurred())
		Expect(class.Object["value"]).To(BeEquivalentTo(100))
		Expect(class.Object["description"]).To(Equal("The priority of the jobs of the CrownLabs workspace ws-class."))
		Expect(class.GetLabels()).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "crownjob"))
		Expect(class.GetOwnerReferences()).To(HaveLen(1))
		Expect(class.GetOwnerReferences()[0].Kind).To(Equal("Workspace"))
		Expect(class.GetOwnerReferences()[0].Name).To(Equal("ws-class"))
		Expect(class.GetOwnerReferences()[0].UID).To(Equal(ws.UID))
		Expect(ptr.Deref(class.GetOwnerReferences()[0].Controller, false)).To(BeTrue())
		Expect(class.GetManagedFields()).To(ContainElement(HaveField("Manager", forge.CrownJobFieldManager)))
		uid := class.GetUID()

		By("Changing its value when the priority of the Workspace changes, in place")
		for _, priority := range []int32{500, 0} {
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), ws)).To(Succeed())
			ws.Spec.JobOptions.Priority = ptr.To(priority)
			Expect(k8sClient.Update(ctx, ws)).To(Succeed())
			reconcileWorkspace("ws-class")
			class, err = priorityClass("workspace-ws-class")
			Expect(err).ToNot(HaveOccurred())
			Expect(class.Object["value"]).To(BeEquivalentTo(priority))
			Expect(class.GetUID()).To(Equal(uid))
		}

		By("Undoing a change made by hand")
		class.Object["value"] = int64(9999)
		class.Object["description"] = "changed by hand"
		Expect(k8sClient.Update(ctx, class)).To(Succeed())
		reconcileWorkspace("ws-class")
		class, err = priorityClass("workspace-ws-class")
		Expect(err).ToNot(HaveOccurred())
		Expect(class.Object["value"]).To(BeEquivalentTo(0))
		Expect(class.Object["description"]).To(Equal("The priority of the jobs of the CrownLabs workspace ws-class."))

		By("Restoring it after a deletion by hand")
		Expect(k8sClient.Delete(ctx, class)).To(Succeed())
		_, err = priorityClass("workspace-ws-class")
		Expect(kerrors.IsNotFound(err)).To(BeTrue())
		reconcileWorkspace("ws-class")
		class, err = priorityClass("workspace-ws-class")
		Expect(err).ToNot(HaveOccurred())
		Expect(class.Object["value"]).To(BeEquivalentTo(0))
		Expect(class.GetUID()).ToNot(Equal(uid))

		By("Deleting it, with the ClusterQueue, when jobOptions is removed")
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), ws)).To(Succeed())
		ws.Spec.JobOptions = nil
		Expect(k8sClient.Update(ctx, ws)).To(Succeed())
		reconcileWorkspace("ws-class")
		_, err = priorityClass("workspace-ws-class")
		Expect(kerrors.IsNotFound(err)).To(BeTrue())
		_, err = clusterQueue("crownlabs-ws-class")
		Expect(kerrors.IsNotFound(err)).To(BeTrue())

		By("Never touching the class of another name")
		Expect(getKueueObject(other, "workspace-ws-class-other")).To(Succeed())
		Expect(other.GetResourceVersion()).To(Equal(otherVersion))
		Expect(other.Object["value"]).To(BeEquivalentTo(7))
	})

	It("Should not delete a WorkloadPriorityClass with the same name that the Workspace does not control", func() {
		create(workspace("ws-class-foreign", nil))
		foreign := &unstructured.Unstructured{}
		foreign.SetGroupVersionKind(forge.KueueWorkloadPriorityClassGVK)
		foreign.SetName("workspace-ws-class-foreign")
		foreign.Object["value"] = int64(42)
		create(foreign)

		reconcileWorkspace("ws-class-foreign")
		class, err := priorityClass("workspace-ws-class-foreign")
		Expect(err).ToNot(HaveOccurred())
		Expect(class.Object["value"]).To(BeEquivalentTo(42))
	})

	It("Should ignore a Workspace without the target label", func() {
		ws := workspace("ws-queue-untargeted", &clv1alpha1.QueueLimits{MaxAtOnce: map[string]resource.Quantity{"cpu": resource.MustParse("1")}})
		ws.Labels = nil
		create(ws)
		reconcileWorkspace("ws-queue-untargeted")
		_, err := clusterQueue("crownlabs-ws-queue-untargeted")
		Expect(kerrors.IsNotFound(err)).To(BeTrue())
	})

	It("Should not delete a ClusterQueue with the same name that the Workspace does not control", func() {
		create(workspace("ws-queue-foreign", nil))
		foreign := &unstructured.Unstructured{}
		foreign.SetGroupVersionKind(forge.KueueClusterQueueGVK)
		foreign.SetName("crownlabs-ws-queue-foreign")
		foreign.Object["spec"] = map[string]interface{}{}
		create(foreign)

		reconcileWorkspace("ws-queue-foreign")
		_, err := clusterQueue("crownlabs-ws-queue-foreign")
		Expect(err).ToNot(HaveOccurred())
	})
})
