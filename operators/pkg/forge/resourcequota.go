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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

const (
	// InstancesCountKey -> The key for accessing at the total number of instances in the corev1.ResourceList map.
	InstancesCountKey = "count/instances.crownlabs.polito.it"
)

var (
	// CapInstance -> The maximum number of instances that can be started by a Tenant.
	CapInstance int64

	// CapCPU -> The total amount of CPU cores that can be requested by a Tenant.
	CapCPU int

	// CapMemoryGiga -> The total amount of RAM (in gigabytes) that can be requested by a Tenant.
	CapMemoryGiga int

	// SandboxCPUQuota -> The maximum amount of CPU cores that can be used by a sandbox namespace.
	SandboxCPUQuota = *resource.NewQuantity(4, resource.DecimalSI)

	// SandboxRequestCPUQuota -> The maximum amount of CPU cores that can be requested by a sandbox namespace.
	SandboxRequestCPUQuota = *resource.NewQuantity(2, resource.DecimalSI)

	// SandboxMemoryQuota -> The maximum amount of RAM memory that can be used by a sandbox namespace.
	SandboxMemoryQuota = *resource.NewScaledQuantity(8, resource.Giga)
)

// TenantResourceList forges the WorkspaceResourceQuota as the sum of all quota for each workspace plus the personal workspace quota.
func TenantResourceList(workspaces []clv1alpha1.Workspace, personalWorkspaceQuota *apicommon.WorkspaceResourceQuota) apicommon.WorkspaceResourceQuota {
	var quota apicommon.WorkspaceResourceQuota
	quota.OtherResources = make(map[string]resource.Quantity)

	// sum all quota for each existing workspace
	for i := range workspaces {
		quota.Accumulate(&workspaces[i].Spec.Quota.ResourceSpec)
		quota.Instances += workspaces[i].Spec.Quota.Instances
	}

	// add personal workspace quota if defined
	if personalWorkspaceQuota != nil {
		quota.Accumulate(&personalWorkspaceQuota.ResourceSpec)
		quota.Instances += personalWorkspaceQuota.Instances
	}

	// cap the quota if needed
	if CapCPU > 0 {
		quota.CPU = CapIntegerQuantity(quota.CPU, int64(CapCPU))
	}
	if CapMemoryGiga > 0 {
		quota.Memory = CapResourceQuantity(quota.Memory, *resource.NewScaledQuantity(int64(CapMemoryGiga), resource.Giga))
	}
	if CapInstance > 0 {
		quota.Instances = CapIntegerQuantity(quota.Instances, CapInstance)
	}

	return quota
}

// TenantResourceQuotaSpec converts a WorkspaceResourceQuota to a ResourceQuota's resource list.
func TenantResourceQuotaSpec(quota *apicommon.WorkspaceResourceQuota) corev1.ResourceList {
	resList := corev1.ResourceList{
		corev1.ResourceLimitsCPU:       *resource.NewQuantity(quota.CPU, resource.DecimalSI),
		corev1.ResourceLimitsMemory:    quota.Memory,
		corev1.ResourceRequestsCPU:     *resource.NewQuantity(quota.CPU, resource.DecimalSI),
		corev1.ResourceRequestsMemory:  quota.Memory,
		InstancesCountKey:              *resource.NewQuantity(quota.Instances, resource.DecimalSI),
		corev1.ResourceRequestsStorage: quota.Disk,
	}

	// Inject dynamic extended resources (e.g., nvidia.com/gpu)
	InjectOtherResources(quota.OtherResources, resList)

	return resList
}

// SandboxResourceQuotaSpec forges the Resource Quota spec for sandbox namespaces.
func SandboxResourceQuotaSpec() corev1.ResourceList {
	return corev1.ResourceList{
		corev1.ResourceLimitsCPU:      SandboxCPUQuota,
		corev1.ResourceLimitsMemory:   SandboxMemoryQuota,
		corev1.ResourceRequestsCPU:    SandboxRequestCPUQuota,
		corev1.ResourceRequestsMemory: SandboxMemoryQuota,
	}
}
