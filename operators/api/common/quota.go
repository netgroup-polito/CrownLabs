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

package common

import "k8s.io/apimachinery/pkg/api/resource"

// WorkspaceResourceQuota defines the resource quota for each Workspace.
// +k8s:deepcopy-gen=true
type WorkspaceResourceQuota struct {
	// The maximum amount of CPU required by this Workspace.
	// +kubebuilder:validation:XValidation:rule="quantity(self).compareTo(quantity('1')) >= 0",message="Minimum 1 CPU core is required"
	CPU resource.Quantity `json:"cpu"`

	// The maximum amount of RAM memory required by this Workspace.
	// +kubebuilder:validation:XValidation:rule="quantity(self).compareTo(quantity('1Gi')) >= 0",message="Minimum 1 GB of RAM is required"
	Memory resource.Quantity `json:"memory"`

	// The maximum number of concurrent instances required by this Workspace.
	// +kubebuilder:validation:Minimum:=1
	Instances int64 `json:"instances"`
}
