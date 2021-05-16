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

// +kubebuilder:validation:Enum="VirtualMachine";"Container"

// EnvironmentType is an enumeration of the different types of environments that
// can be instantiated in CrownLabs.
type EnvironmentType string

const (
	// ClassContainer -> the environment is constituted by a Docker container.
	ClassContainer EnvironmentType = "Container"
	// ClassVM -> the environment is constituted by a Virtual Machine.
	ClassVM EnvironmentType = "VirtualMachine"
)

// TemplateSpec is the specification of the desired state of the Template.
type TemplateSpec struct {
	// The human-readable name of the Template.
	PrettyName string `json:"prettyName"`

	// A textual description of the Template.
	Description string `json:"description"`

	// The reference to the Workspace this Template belongs to.
	WorkspaceRef GenericRef `json:"workspace.crownlabs.polito.it/WorkspaceRef,omitempty"`

	// The list of environments (i.e. VMs or containers) that compose the Template.
	EnvironmentList []Environment `json:"environmentList"`

	// +kubebuilder:validation:Pattern="^[0-9]+[mhd]$"
	// +kubebuilder:default="7d"

	// The maximum lifetime of an Instance referencing the current Template.
	// Once this period is expired, the Instance may be automatically deleted
	// or stopped to save resources.
	DeleteAfter string `json:"deleteAfter,omitempty"`
}

// TemplateStatus reflects the most recently observed status of the Template.
type TemplateStatus struct {
}

// Environment defines the characteristics of an environment composing the Template.
type Environment struct {
	// The name identifying the specific environment.
	Name string `json:"name"`

	// The VM or container to be started when instantiating the environment.
	Image string `json:"image"`

	// The type of environment to be instantiated, among VirtualMachine and
	// Container.
	EnvironmentType EnvironmentType `json:"environmentType"`

	// +kubebuilder:default=true

	// Whether the environment is characterized by a graphical desktop or not.
	GuiEnabled bool `json:"guiEnabled,omitempty"`

	// +kubebuilder:default=false

	// Whether the environment should be persistent (i.e. preserved when the
	// corresponding instance is terminated) or not.
	Persistent bool `json:"persistent,omitempty"`

	// The amount of computational resources associated with the environment.
	Resources EnvironmentResources `json:"resources"`
}

// EnvironmentResources is the specification of the amount of resources
// (i.e. CPU, RAM, ...) assigned to a certain environment.
type EnvironmentResources struct {
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=8

	// The maximum number of CPU cores made available to the environment
	// (ranging between 1 and 8 cores). This maps to the 'limits' specified
	// for the actual pod representing the environment.
	CPU uint32 `json:"cpu"`

	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=100

	// The percentage of reserved CPU cores, ranging between 1 and 100, with
	// respect to the 'CPU' value. Essentially, this corresponds to the 'requests'
	// specified for the actual pod representing the environment.
	ReservedCPUPercentage uint32 `json:"reservedCPUPercentage"`

	// The amount of RAM memory assigned to the given environment. Requests and
	// limits do correspond to avoid OOMKill issues.
	Memory resource.Quantity `json:"memory"`

	// The size of the persistent disk allocated for the given environment.
	// This field is meaningful only in case of persistent or container-based
	// environments, while it is silently ignored in the other cases.
	// In case of containers, when this field is not specified, an emptyDir will be
	// attached to the pod but this could result in data loss whenever the pod dies.
	Disk resource.Quantity `json:"disk,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="tmpl"
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Pretty Name",type=string,JSONPath=`.spec.prettyName`
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.environmentList[0].image`,priority=10
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.environmentList[0].environmentType`,priority=10
// +kubebuilder:printcolumn:name="GUI",type=string,JSONPath=`.spec.environmentList[0].guiEnabled`,priority=10
// +kubebuilder:printcolumn:name="Persistent",type=string,JSONPath=`.spec.environmentList[0].persistent`,priority=10
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Template describes the template of a CrownLabs environment to be instantiated.
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TemplateList contains a list of Template objects.
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
