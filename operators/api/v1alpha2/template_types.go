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

package v1alpha2

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EnvironmentType string

const (
	ClassContainer EnvironmentType = "Container"
	ClassVM        EnvironmentType = "VirtualMachine"
)

// TemplateSpec defines the desired state of Template
type TemplateSpec struct {
	WorkspaceRef    GenericRef    `json:"workspace.crownlabs.polito.it/WorkspaceRef,omitempty"`
	PrettyName      string        `json:"prettyName"`
	Description     string        `json:"description"`
	EnvironmentList []Environment `json:"environmentList"`
	// +kubebuilder:validation:Pattern="^[0-9]+[mhd]$"
	// +kubebuilder:default="7d"
	DeleteAfter string `json:"deleteAfter,omitempty"`
}

// TemplateStatus defines the observed state of Template
type TemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
}

type Environment struct {
	Name string `json:"name,omitempty"`
	// +kubebuilder:default=true
	GuiEnabled bool                 `json:"guiEnabled,omitempty"`
	Resources  EnvironmentResources `json:"resources"`
	// +kubebuilder:validation:Enum="VirtualMachine";"Container"
	EnvironmentType EnvironmentType `json:"environmentType"`
	Persistent      bool            `json:"persistent"`
	Image           string          `json:"image"`
}

type EnvironmentResources struct {
	// +kubebuilder:validation:Maximum:=8
	// +kubebuilder:validation:Minimum:=1
	CPU uint32 `json:"cpu"`
	// +kubebuilder:validation:Maximum:=100
	// +kubebuilder:validation:Minimum:=1
	ReservedCPUPercentage uint32            `json:"reservedCPUPercentage"`
	Memory                resource.Quantity `json:"memory"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="tmpl"
// +kubebuilder:storageversion

// Template is the Schema for the templates API
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
