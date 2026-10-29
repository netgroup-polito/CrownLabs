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

	It("Should create, update and delete the ClusterQueue with queueLimits, and never delete the ResourceFlavor", func() {
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
		ws.Spec.QueueLimits.MaxAtOnce = map[string]resource.Quantity{"cpu": resource.MustParse("3"), "memory": resource.MustParse("6Gi")}
		Expect(k8sClient.Update(ctx, ws)).To(Succeed())
		reconcileWorkspace("ws-queue")
		queue, err = clusterQueue("crownlabs-ws-queue")
		Expect(err).ToNot(HaveOccurred())
		Expect(nominalQuotas(queue)).To(Equal(map[string]string{"cpu": "3", "memory": "6Gi"}))

		By("Deleting the ClusterQueue when queueLimits is removed, keeping the ResourceFlavor")
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), ws)).To(Succeed())
		ws.Spec.QueueLimits = nil
		Expect(k8sClient.Update(ctx, ws)).To(Succeed())
		reconcileWorkspace("ws-queue")
		_, err = clusterQueue("crownlabs-ws-queue")
		Expect(kerrors.IsNotFound(err)).To(BeTrue())
		Expect(getKueueObject(flavor, "crownlabs-default")).To(Succeed())

		By("Re-creating the ClusterQueue when queueLimits is set again")
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), ws)).To(Succeed())
		ws.Spec.QueueLimits = &clv1alpha1.QueueLimits{MaxAtOnce: map[string]resource.Quantity{"cpu": resource.MustParse("1")}}
		Expect(k8sClient.Update(ctx, ws)).To(Succeed())
		reconcileWorkspace("ws-queue")
		queue, err = clusterQueue("crownlabs-ws-queue")
		Expect(err).ToNot(HaveOccurred())
		Expect(nominalQuotas(queue)).To(Equal(map[string]string{"cpu": "1", "memory": "100Ti"}))
	})

	It("Should not create a ClusterQueue for a Workspace without queueLimits", func() {
		create(workspace("ws-queue-none", nil))
		reconcileWorkspace("ws-queue-none")
		_, err := clusterQueue("crownlabs-ws-queue-none")
		Expect(kerrors.IsNotFound(err)).To(BeTrue())
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
