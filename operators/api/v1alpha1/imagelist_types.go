// Copyright 2020-2025 Politecnico di Torino
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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImageListItem describes a single VM image.
type ImageListItem struct {
	// The name identifying a single image.
	Name string `json:"name"`

	// The list of versions the image is available in.
	Versions []string `json:"versions"`
}

// ImageListSpec is the specification of the desired state of the ImageList.
type ImageListSpec struct {
	// The host name that can be used to access the registry.
	RegistryName string `json:"registryName"`

	// The list of VM images currently available in CrownLabs.
	Images []ImageListItem `json:"images"`
}

// ImageListStatus reflects the most recently observed status of the ImageList.
type ImageListStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:printcolumn:name="Registry Name",type=string,JSONPath=`.spec.registryName`

// ImageList describes the available VM images in the CrownLabs registry.
type ImageList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageListSpec   `json:"spec,omitempty"`
	Status ImageListStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ImageListList contains a list of ImageList objects.
type ImageListList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageList `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageList{}, &ImageListList{})
}
