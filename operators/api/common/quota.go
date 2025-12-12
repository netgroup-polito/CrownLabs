package common

import "k8s.io/apimachinery/pkg/api/resource"

// WorkspaceResourceQuota defines the resource quota for each Workspace.
// +k8s:deepcopy-gen=true
type WorkspaceResourceQuota struct {
	// The maximum amount of CPU required by this Workspace.
	CPU resource.Quantity `json:"cpu"`

	// The maximum amount of RAM memory required by this Workspace.
	Memory resource.Quantity `json:"memory"`

	// +kubebuilder:validation:Minimum:=1
	// The maximum number of concurrent instances required by this Workspace.
	Instances int64 `json:"instances"`
}
