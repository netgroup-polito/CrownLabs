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

package v1alpha2

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum="VirtualMachine";"Container";"CloudVM";"Standalone"

// EnvironmentType is an enumeration of the different types of environments that
// can be instantiated in CrownLabs.
type EnvironmentType string

// +kubebuilder:validation:Enum="Standard";"Exam";"Exercise"

// EnvironmentScope is an enumeration of the scope in which the environment of the associated instance should be started:
// each scope consists in presets for exposition and deployment.
type EnvironmentScope string

const (
	// ClassContainer -> the environment is constituted by a Docker container exposing a service through a VNC server.
	ClassContainer EnvironmentType = "Container"
	// ClassVM -> the environment is constituted by a Virtual Machine.
	ClassVM EnvironmentType = "VirtualMachine"
	// ClassCloudVM -> the environment is constituited by a Virtual Machine started from cloud images downloaded from HTTP URL.
	ClassCloudVM EnvironmentType = "CloudVM"
	// ClassStandalone -> the environment is constituted by a Docker Container exposing a web service through an http interface.
	ClassStandalone EnvironmentType = "Standalone"

	// ScopeStandard -> Normal operation (authentication, ssh, files access).
	ScopeStandard EnvironmentScope = "Standard"
	// ScopeExam -> Restricted access (no authentication, no mydrive access).
	ScopeExam EnvironmentScope = "Exam"
	// ScopeExercise -> Restricted access (no authentication, no mydrive access).
	ScopeExercise EnvironmentScope = "Exercise"
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
	// Each environment must have a unique name within the Template.
	// +listType=map
	// +listMapKey=name
	EnvironmentList []Environment `json:"environmentList"`

	// +kubebuilder:validation:Pattern="^(never|[0-9]+[mhd])$"
	// +kubebuilder:default="never"

	// The maximum lifetime of an Instance referencing the current Template.
	// Once this period is expired, the Instance may be automatically deleted
	// or stopped to save resources. If set to "never", the instance will not be
	// automatically terminated.
	DeleteAfter string `json:"deleteAfter,omitempty"`

	// Labels that are used for the selection of the node.
	// They are given by means of a pointer to check the presence of the field.
	// In case it is present, the labels that are chosen are the ones present on the instance
	NodeSelector *map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:default="Standard"

	// The scope associated with the environments belonging to the template (Standard, Exam, Exercise)
	Scope EnvironmentScope `json:"scope,omitempty"`

	// +kubebuilder:validation:Pattern="^(never|[0-9]+[mhd])$"
	// +kubebuilder:default="never"
	// The maximum period of inactivity after which an Instance referencing
	// the current Template will be automatically stopped or deleted to
	// save resources.
	InactivityTimeout string `json:"inactivityTimeout,omitempty"`
}

// TemplateStatus reflects the most recently observed status of the Template.
type TemplateStatus struct {
}

// Environment defines the characteristics of an environment composing the Template.
type Environment struct {
	// The name identifying the specific environment.
	// The name must be unique within the Template and must follow the Kubernetes
	// naming conventions, i.e. it must consist of lower case alphanumeric characters,
	// '-' or '.', must start and end with an alphanumeric character.
	// +kubebuilder:validation:Pattern="^[a-z\\d][a-z\\d-]{2,10}[a-z\\d]$"
	Name string `json:"name"`

	// The VM or container to be started when instantiating the environment.
	Image string `json:"image"`

	// The type of environment to be instantiated, among VirtualMachine,
	// Container, CloudVM and Standalone.
	EnvironmentType EnvironmentType `json:"environmentType"`

	// +kubebuilder:default=true

	// Whether the environment is characterized by a graphical desktop or not.
	GuiEnabled bool `json:"guiEnabled,omitempty"`

	// +kubebuilder:default=false

	// For VNC based containers, hide the noVNC control bar when true
	DisableControls bool `json:"disableControls,omitempty"`

	// +kubebuilder:default=false

	// Whether the environment should be persistent (i.e. preserved when the
	// corresponding instance is terminated) or not.
	Persistent bool `json:"persistent,omitempty"`

	// The amount of computational resources associated with the environment.
	Resources EnvironmentResources `json:"resources"`

	// +kubebuilder:default=false
	// Whether the environment needs the URL Rewrite or not.
	RewriteURL bool `json:"rewriteURL,omitempty"`

	// Options to customize container startup
	ContainerStartupOptions *ContainerStartupOpts `json:"containerStartupOptions,omitempty"`

	// Name of the storage class to be used for the persistent volume (when needed)
	StorageClassName string `json:"storageClassName,omitempty"`

	// +kubebuilder:default=true
	// Whether the instance has to have the user's MyDrive volume
	MountMyDriveVolume bool `json:"mountMyDriveVolume"`

	// The list of information about Shared Volumes that has to be mounted to the instance.
	SharedVolumeMounts []SharedVolumeMountInfo `json:"sharedVolumeMounts,omitempty"`
}

// EnvironmentResources is the specification of the amount of resources
// (i.e. CPU, RAM, ...) assigned to a certain environment.
type EnvironmentResources struct {
	// +kubebuilder:validation:Minimum:=1

	// The maximum number of CPU cores made available to the environment
	// (at least 1 core). This maps to the 'limits' specified
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

// ContainerStartupOpts specifies custom startup options for the created container,
// including the possibility to download and extract an archive to a given destination
// and specifying the arguments that will be passed to the application container.
type ContainerStartupOpts struct {
	// URL from which GET the archive to be extracted into ContentPath
	SourceArchiveURL string `json:"sourceArchiveURL,omitempty"`
	// Path on which storage (EmptyDir/Storage) will be mounted
	// and into which, if given in SourceArchiveURL, will be extracted the archive
	ContentPath string `json:"contentPath,omitempty"`
	// Arguments to be passed to the application container on startup
	StartupArgs []string `json:"startupArgs,omitempty"`
	// Whether forcing the container working directory to be the same as the contentPath (or default mydrive path if not specified)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	EnforceWorkdir bool `json:"enforceWorkdir"`
}

// SharedVolumeMountInfo contains mount information for a Shared Volume.
type SharedVolumeMountInfo struct {
	// The reference of the Shared Volume this Mount Info is related to.
	SharedVolumeRef GenericRef `json:"sharedVolume"`

	// The path the Shared Volume will be mounted in.
	MountPath string `json:"mountPath"`

	// Whether this Shared Volume should be mounted with R/W or R/O permission.
	ReadOnly bool `json:"readOnly"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="tmpl"
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Pretty Name",type=string,JSONPath=`.spec.prettyName`
// +kubebuilder:printcolumn:name="Scope",type=string,JSONPath=`.spec.environmentList[0].scope`
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
