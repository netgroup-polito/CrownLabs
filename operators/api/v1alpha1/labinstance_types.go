/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LabInstanceSpec defines the desired state of LabInstance
type LabInstanceSpec struct {
	LabTemplateName      string `json:"labTemplateName,omitempty"`
	LabTemplateNamespace string `json:"labTemplateNamespace,omitempty"`
	StudentID            string `json:"studentId,omitempty"`
}

// LabInstanceStatus defines the observed state of LabInstance
type LabInstanceStatus struct {
	Phase string `json:"phase,omitempty"`
	URL   string `json:"url,omitempty"`
	IP    string `json:"ip,omitempty"`
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="labi"

// LabInstance is the Schema for the labinstances API
type LabInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabInstanceSpec   `json:"spec,omitempty"`
	Status LabInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LabInstanceList contains a list of LabInstance
type LabInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabInstance{}, &LabInstanceList{})
}
