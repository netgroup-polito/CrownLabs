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

// Package common contains common API types used across different CrownLabs operator groups.
package common

import (
	"encoding/json"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

// ResourceSpec contains the common resource fields shared across CrownLabs API objects.
// +k8s:deepcopy-gen=true
type ResourceSpec struct {
	// The maximum amount of CPU required by this resource set.
	// +kubebuilder:validation:Minimum:=1
	CPU int64 `json:"cpu"`

	// The maximum amount of RAM memory required by this resource set.
	// +kubebuilder:validation:XValidation:rule="quantity(self).compareTo(quantity('1Gi')) >= 0",message="Minimum 1 GB of RAM is required"
	Memory resource.Quantity `json:"memory"`

	// The maximum amount of disk occupancy required by this resource set.
	// +kubebuilder:validation:Optional
	Disk resource.Quantity `json:"disk,omitempty"`

	// Generic map to handle any extended hardware resources (e.g., nvidia.com/gpu, amd.com/gpu)
	// without hardcoding specific vendor keys.
	OtherResources map[string]resource.Quantity `json:"otherResources,omitempty"`
}

// TO DO : Remove this alias int64Flexible conversion method and only leave int64 once the production mutation webhook is updated.
// this is necessary because right now the mutation webhook changes integers into strings.

// UnmarshalJSON implements custom JSON unmarshaling to handle both integer and string representations of CPU.
func (s *ResourceSpec) UnmarshalJSON(data []byte) error {
	type Alias ResourceSpec

	aux := &struct {
		CPU int64Flexible `json:"cpu"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.CPU = int64(aux.CPU)
	return nil
}

// int64Flexible is an internal helper type that unmarshals both integers (e.g., 8) and strings (e.g., "8" or "8000m") into int64.
type int64Flexible int64

func (f *int64Flexible) UnmarshalJSON(b []byte) error {
	// 1. Try decoding directly as int64
	var i int64
	if err := json.Unmarshal(b, &i); err == nil {
		*f = int64Flexible(i)
		return nil
	}

	// 2. Otherwise, try decoding as string
	var strVal string
	if err := json.Unmarshal(b, &strVal); err != nil {
		return err
	}

	strVal = strings.TrimSpace(strVal)
	if strVal == "" {
		*f = 0
		return nil
	}

	// 3. Handle millicores if "m" suffix is present (e.g., "8000m" -> 8 cores)
	if strings.HasSuffix(strVal, "m") {
		trimmed := strings.TrimSuffix(strVal, "m")
		milli, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return err
		}
		*f = int64Flexible(milli / 1000)
		return nil
	}

	// 4. Parse string as regular integer
	val, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		return err
	}

	*f = int64Flexible(val)
	return nil
}

// Accumulate merges and adds the resources from one or more ResourceSpec objects into the receiver.
func (s *ResourceSpec) Accumulate(other *ResourceSpec) {
	if other == nil {
		return
	}

	s.CPU += other.CPU
	s.Memory.Add(other.Memory)
	s.Disk.Add(other.Disk)

	// Dynamically merge and sum extended/custom resources
	if other.OtherResources != nil {
		if s.OtherResources == nil {
			s.OtherResources = make(map[string]resource.Quantity)
		}
		for resName, resQty := range other.OtherResources {
			if currentQty, exists := s.OtherResources[resName]; exists {
				currentQty.Add(resQty)
				s.OtherResources[resName] = currentQty
			} else {
				s.OtherResources[resName] = resQty.DeepCopy()
			}
		}
	}
}

// WorkspaceResourceQuota defines the resource quota for each Workspace.
// +k8s:deepcopy-gen=true
type WorkspaceResourceQuota struct {
	ResourceSpec `json:",inline"`

	// The maximum number of concurrent instances required by this Workspace.
	// +kubebuilder:validation:Minimum:=1
	Instances int64 `json:"instances"`
}

// UnmarshalJSON implements custom unmarshaling to prevent method promotion issues.
// TODO: Remove this workaround once the production mutation webhook is updated.
func (w *WorkspaceResourceQuota) UnmarshalJSON(data []byte) error {
	// 1. Unmarshal embedded ResourceSpec (handles cpu string/int, memory, disk, etc.)
	if err := json.Unmarshal(data, &w.ResourceSpec); err != nil {
		return err
	}

	// 2. Unmarshal WorkspaceResourceQuota specific fields
	var aux struct {
		Instances int64 `json:"instances"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	w.Instances = aux.Instances
	return nil
}
