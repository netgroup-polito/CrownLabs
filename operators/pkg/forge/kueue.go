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

package forge

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

// The Kueue objects are handled as unstructured objects, to avoid a Go dependency on Kueue
// (whose current versions require newer Go, Kubernetes and controller-runtime versions).

const (
	// KueueGroup -> the API group of Kueue.
	KueueGroup = "kueue.x-k8s.io"
	// KueueVersion -> the version of the Kueue API used by CrownLabs.
	KueueVersion = "v1beta2"

	// KueueQueueNameLabel -> the label of a Job naming the LocalQueue it is submitted to.
	KueueQueueNameLabel = "kueue.x-k8s.io/queue-name"
	// KueuePriorityClassLabel -> the label of a Job naming its WorkloadPriorityClass.
	KueuePriorityClassLabel = "kueue.x-k8s.io/priority-class"
	// KueueJobUIDLabel -> the label set by Kueue on a Workload, with the UID of its Job.
	KueueJobUIDLabel = "kueue.x-k8s.io/job-uid"

	// KueueConditionQuotaReserved -> the Workload condition telling whether quota is reserved in a ClusterQueue.
	KueueConditionQuotaReserved = "QuotaReserved"

	// CrownJobFieldManager -> the field manager used to apply the Kueue objects.
	CrownJobFieldManager = "crownlabs-crownjob"
	// CrownJobQueuePrefix -> the prefix of the name of the ClusterQueue and LocalQueue of a Workspace.
	CrownJobQueuePrefix = "crownlabs-"
	// CrownJobResourceFlavor -> the name of the ResourceFlavor used by every job queue.
	CrownJobResourceFlavor = "crownlabs-default"
	// CrownJobTenantNamespaceSelectorValue -> the value of the type label of the namespaces that can use a job queue.
	CrownJobTenantNamespaceSelectorValue = labelTypeTenantValue

	labelManagedByCrownJobValue = "crownjob"
)

var (
	// KueueClusterQueueGVK -> the GroupVersionKind of a Kueue ClusterQueue.
	KueueClusterQueueGVK = schema.GroupVersionKind{Group: KueueGroup, Version: KueueVersion, Kind: "ClusterQueue"}
	// KueueLocalQueueGVK -> the GroupVersionKind of a Kueue LocalQueue.
	KueueLocalQueueGVK = schema.GroupVersionKind{Group: KueueGroup, Version: KueueVersion, Kind: "LocalQueue"}
	// KueueResourceFlavorGVK -> the GroupVersionKind of a Kueue ResourceFlavor.
	KueueResourceFlavorGVK = schema.GroupVersionKind{Group: KueueGroup, Version: KueueVersion, Kind: "ResourceFlavor"}
	// KueueWorkloadPriorityClassGVK -> the GroupVersionKind of a Kueue WorkloadPriorityClass.
	KueueWorkloadPriorityClassGVK = schema.GroupVersionKind{Group: KueueGroup, Version: KueueVersion, Kind: "WorkloadPriorityClass"}
	// KueueWorkloadGVK -> the GroupVersionKind of a Kueue Workload.
	KueueWorkloadGVK = schema.GroupVersionKind{Group: KueueGroup, Version: KueueVersion, Kind: "Workload"}

	// KueueNoLimitCPU -> the nominal quota of cpu in a job queue whose maxAtOnce does not limit it.
	// Kueue holds for ever a workload that requests a resource its ClusterQueue does not cover,
	// so cpu and memory are always covered, with a value that no job reaches.
	KueueNoLimitCPU = resource.MustParse("10000")
	// KueueNoLimitMemory -> the nominal quota of memory in a job queue whose maxAtOnce does not limit it (see KueueNoLimitCPU).
	KueueNoLimitMemory = resource.MustParse("100Ti")
)

// KueueQueueQuota is the nominal quota of one resource in the job queue of a Workspace.
type KueueQueueQuota struct {
	Name  corev1.ResourceName
	Quota resource.Quantity
}

// CrownJobQueueName returns the name of the ClusterQueue and of the LocalQueues of the given Workspace.
func CrownJobQueueName(workspace string) string {
	return CrownJobQueuePrefix + workspace
}

// KueueQueueQuotas returns the resources covered by a job queue with the given limits, with their nominal quota:
// cpu and memory first (always covered, see KueueNoLimitCPU), then the other keys of maxAtOnce, sorted.
func KueueQueueQuotas(limits *clv1alpha1.QueueLimits) []KueueQueueQuota {
	var maxAtOnce map[string]resource.Quantity
	if limits != nil {
		maxAtOnce = limits.MaxAtOnce
	}

	quotaOrDefault := func(name corev1.ResourceName, def resource.Quantity) resource.Quantity {
		if quota, found := maxAtOnce[name.String()]; found {
			return quota.DeepCopy()
		}
		return def.DeepCopy()
	}

	quotas := []KueueQueueQuota{
		{Name: corev1.ResourceCPU, Quota: quotaOrDefault(corev1.ResourceCPU, KueueNoLimitCPU)},
		{Name: corev1.ResourceMemory, Quota: quotaOrDefault(corev1.ResourceMemory, KueueNoLimitMemory)},
	}

	others := make([]string, 0, len(maxAtOnce))
	for name := range maxAtOnce {
		if name != corev1.ResourceCPU.String() && name != corev1.ResourceMemory.String() {
			others = append(others, name)
		}
	}
	sort.Strings(others)
	for _, name := range others {
		quotas = append(quotas, KueueQueueQuota{Name: corev1.ResourceName(name), Quota: maxAtOnce[name].DeepCopy()})
	}

	return quotas
}

// KueueResourceFlavor forges the ResourceFlavor used by every job queue (with an empty spec).
func KueueResourceFlavor() *unstructured.Unstructured {
	flavor := kueueObject(KueueResourceFlavorGVK, CrownJobResourceFlavor, "")
	flavor.Object["spec"] = map[string]interface{}{}
	return flavor
}

// KueueClusterQueue forges the ClusterQueue of the job queue of the given Workspace, which must have queueLimits.
// It admits the jobs of the tenant namespaces, in BestEffortFIFO order, without cohort and without preemption.
func KueueClusterQueue(workspace *clv1alpha1.Workspace) *unstructured.Unstructured {
	quotas := KueueQueueQuotas(workspace.Spec.QueueLimits)

	covered := make([]interface{}, 0, len(quotas))
	resources := make([]interface{}, 0, len(quotas))
	for i := range quotas {
		covered = append(covered, quotas[i].Name.String())
		resources = append(resources, map[string]interface{}{
			"name":         quotas[i].Name.String(),
			"nominalQuota": quotas[i].Quota.String(),
		})
	}

	queue := kueueObject(KueueClusterQueueGVK, CrownJobQueueName(workspace.Name), "")
	queue.SetOwnerReferences([]metav1.OwnerReference{*metav1.NewControllerRef(workspace, clv1alpha1.GroupVersion.WithKind("Workspace"))})
	queue.Object["spec"] = map[string]interface{}{
		"namespaceSelector": map[string]interface{}{
			"matchLabels": map[string]interface{}{
				LabelTypeKey: CrownJobTenantNamespaceSelectorValue,
			},
		},
		"queueingStrategy": "BestEffortFIFO",
		"resourceGroups": []interface{}{
			map[string]interface{}{
				"coveredResources": covered,
				"flavors": []interface{}{
					map[string]interface{}{
						"name":      CrownJobResourceFlavor,
						"resources": resources,
					},
				},
			},
		},
	}
	return queue
}

// KueueLocalQueue forges the LocalQueue of the job queue of the given Workspace, in the given (tenant) namespace.
func KueueLocalQueue(workspace *clv1alpha1.Workspace, namespace string) *unstructured.Unstructured {
	queue := kueueObject(KueueLocalQueueGVK, CrownJobQueueName(workspace.Name), namespace)
	queue.SetOwnerReferences([]metav1.OwnerReference{*metav1.NewControllerRef(workspace, clv1alpha1.GroupVersion.WithKind("Workspace"))})
	queue.Object["spec"] = map[string]interface{}{
		"clusterQueue": CrownJobQueueName(workspace.Name),
	}
	return queue
}

// kueueObject forges an unstructured Kueue object of the given kind, with the managed-by label.
func kueueObject(gvk schema.GroupVersionKind, name, namespace string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(name)
	obj.SetNamespace(namespace)
	obj.SetLabels(map[string]string{LabelManagedByKey: labelManagedByCrownJobValue})
	return obj
}
