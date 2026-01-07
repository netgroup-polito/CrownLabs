package common

import "k8s.io/apimachinery/pkg/api/resource"

// WorkspaceResourceQuota defines the resource quota for each Workspace.
// +k8s:deepcopy-gen=true
type WorkspaceResourceQuota struct {
	// The maximum amount of CPU required by this Workspace.
	// +kubebuilder:validation:XValidation:rule="quantity(string(self)).compareTo(quantity('1')) >= 0",message="Minimum 1 CPU core is required"
	CPU resource.Quantity `json:"cpu"`

	// The maximum amount of RAM memory required by this Workspace.
	// +kubebuilder:validation:XValidation:rule="quantity(string(self)).compareTo(quantity('1G')) >= 0",message="Minimum 1 GB memory is required"
	Memory resource.Quantity `json:"memory"`

	// The maximum number of concurrent instances required by this Workspace.
	// +kubebuilder:validation:Minimum:=1
	Instances int64 `json:"instances"`
}
