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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Kueue objects forging", func() {
	var workspace clv1alpha1.Workspace

	BeforeEach(func() {
		workspace = clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: "ws", UID: "ws-uid"},
			Spec: clv1alpha1.WorkspaceSpec{QueueLimits: &clv1alpha1.QueueLimits{
				MaxAtOnce: map[string]resource.Quantity{
					"nvidia.com/gpu":   resource.MustParse("4"),
					"cpu":              resource.MustParse("8"),
					"amd.com/gpu":      resource.MustParse("1"),
					"example.com/fpga": resource.MustParse("2"),
				},
			}},
		}
	})

	workspaceOwner := metav1.OwnerReference{
		APIVersion: "crownlabs.polito.it/v1alpha1", Kind: "Workspace", Name: "ws", UID: "ws-uid",
		Controller: ptr.To(true), BlockOwnerDeletion: ptr.To(true),
	}

	resourceGroup := func(queue *unstructured.Unstructured) map[string]interface{} {
		groups, found, err := unstructured.NestedSlice(queue.Object, "spec", "resourceGroups")
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(groups).To(HaveLen(1))
		return groups[0].(map[string]interface{})
	}

	flavorResources := func(group map[string]interface{}) []interface{} {
		flavors := group["flavors"].([]interface{})
		Expect(flavors).To(HaveLen(1))
		flavor := flavors[0].(map[string]interface{})
		Expect(flavor["name"]).To(Equal("crownlabs-default"))
		return flavor["resources"].([]interface{})
	}

	Describe("The forge.KueueClusterQueue function", func() {
		When("maxAtOnce does not list memory", func() {
			It("Should cover cpu, memory, then the other keys sorted, with memory at the no-limit value", func() {
				queue := forge.KueueClusterQueue(&workspace)
				Expect(queue.GroupVersionKind()).To(Equal(forge.KueueClusterQueueGVK))
				Expect(queue.GetAPIVersion()).To(Equal("kueue.x-k8s.io/v1beta2"))
				Expect(queue.GetName()).To(Equal("crownlabs-ws"))
				Expect(queue.GetNamespace()).To(BeEmpty())
				Expect(queue.GetOwnerReferences()).To(ConsistOf(workspaceOwner))
				Expect(queue.GetLabels()).To(Equal(map[string]string{"crownlabs.polito.it/managed-by": "crownjob"}))

				group := resourceGroup(queue)
				Expect(group["coveredResources"]).To(Equal([]interface{}{"cpu", "memory", "amd.com/gpu", "example.com/fpga", "nvidia.com/gpu"}))
				Expect(flavorResources(group)).To(Equal([]interface{}{
					map[string]interface{}{"name": "cpu", "nominalQuota": "8"},
					map[string]interface{}{"name": "memory", "nominalQuota": "100Ti"},
					map[string]interface{}{"name": "amd.com/gpu", "nominalQuota": "1"},
					map[string]interface{}{"name": "example.com/fpga", "nominalQuota": "2"},
					map[string]interface{}{"name": "nvidia.com/gpu", "nominalQuota": "4"},
				}))
			})

			It("Should select the tenant namespaces, use BestEffortFIFO, and set no cohort and no preemption", func() {
				spec, _, err := unstructured.NestedMap(forge.KueueClusterQueue(&workspace).Object, "spec")
				Expect(err).ToNot(HaveOccurred())
				Expect(spec).To(HaveKeyWithValue("namespaceSelector", map[string]interface{}{
					"matchLabels": map[string]interface{}{"crownlabs.polito.it/type": "tenant"},
				}))
				Expect(spec).To(HaveKeyWithValue("queueingStrategy", "BestEffortFIFO"))
				Expect(spec).ToNot(HaveKey("cohortName"))
				Expect(spec).ToNot(HaveKey("preemption"))
				Expect(spec).To(HaveLen(3))
			})
		})

		When("maxAtOnce lists neither cpu nor memory", func() {
			BeforeEach(func() {
				workspace.Spec.QueueLimits.MaxAtOnce = map[string]resource.Quantity{"nvidia.com/gpu": resource.MustParse("2")}
			})

			It("Should still cover cpu and memory, at the no-limit values", func() {
				group := resourceGroup(forge.KueueClusterQueue(&workspace))
				Expect(group["coveredResources"]).To(Equal([]interface{}{"cpu", "memory", "nvidia.com/gpu"}))
				Expect(flavorResources(group)).To(Equal([]interface{}{
					map[string]interface{}{"name": "cpu", "nominalQuota": "10k"},
					map[string]interface{}{"name": "memory", "nominalQuota": "100Ti"},
					map[string]interface{}{"name": "nvidia.com/gpu", "nominalQuota": "2"},
				}))
			})
		})

		When("maxAtOnce lists both cpu and memory", func() {
			BeforeEach(func() {
				workspace.Spec.QueueLimits.MaxAtOnce = map[string]resource.Quantity{
					"memory": resource.MustParse("6Gi"), "cpu": resource.MustParse("1"),
				}
			})

			It("Should cover only cpu and memory, with the maxAtOnce values", func() {
				group := resourceGroup(forge.KueueClusterQueue(&workspace))
				Expect(group["coveredResources"]).To(Equal([]interface{}{"cpu", "memory"}))
				Expect(flavorResources(group)).To(Equal([]interface{}{
					map[string]interface{}{"name": "cpu", "nominalQuota": "1"},
					map[string]interface{}{"name": "memory", "nominalQuota": "6Gi"},
				}))
			})
		})

		When("maxAtOnce is empty", func() {
			BeforeEach(func() { workspace.Spec.QueueLimits.MaxAtOnce = nil })

			It("Should cover cpu and memory at the no-limit values", func() {
				group := resourceGroup(forge.KueueClusterQueue(&workspace))
				Expect(group["coveredResources"]).To(Equal([]interface{}{"cpu", "memory"}))
			})
		})
	})

	Describe("The forge.KueueLocalQueue function", func() {
		It("Should forge the LocalQueue in the given namespace, pointing to the ClusterQueue of the workspace", func() {
			queue := forge.KueueLocalQueue(&workspace, "tenant-student")
			Expect(queue.GroupVersionKind()).To(Equal(forge.KueueLocalQueueGVK))
			Expect(queue.GetName()).To(Equal("crownlabs-ws"))
			Expect(queue.GetNamespace()).To(Equal("tenant-student"))
			Expect(queue.GetOwnerReferences()).To(ConsistOf(workspaceOwner))
			Expect(queue.Object["spec"]).To(Equal(map[string]interface{}{"clusterQueue": "crownlabs-ws"}))
		})
	})

	Describe("The forge.KueueResourceFlavor function", func() {
		It("Should forge the crownlabs-default ResourceFlavor with an empty spec and no owner", func() {
			flavor := forge.KueueResourceFlavor()
			Expect(flavor.GroupVersionKind()).To(Equal(forge.KueueResourceFlavorGVK))
			Expect(flavor.GetName()).To(Equal("crownlabs-default"))
			Expect(flavor.GetOwnerReferences()).To(BeEmpty())
			Expect(flavor.Object["spec"]).To(Equal(map[string]interface{}{}))
		})
	})
})
