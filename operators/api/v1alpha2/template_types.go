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

package v1alpha2

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
)

// +kubebuilder:validation:Enum="VirtualMachine";"Container";"CloudVM";"Standalone";"LocalVM"

// EnvironmentType is an enumeration of the different types of environments that
// can be instantiated in CrownLabs.
type EnvironmentType string

const (
	// ClassContainer -> the environment is constituted by a Docker container exposing a service through a VNC server.
	ClassContainer EnvironmentType = "Container"
	// ClassVM -> the environment is constituted by a Virtual Machine.
	ClassVM EnvironmentType = "VirtualMachine"
	// ClassCloudVM -> the environment is constituited by a Virtual Machine started from cloud images downloaded from HTTP URL.
	ClassCloudVM EnvironmentType = "CloudVM"
	// ClassStandalone -> the environment is constituted by a Docker Container exposing a web service through an http interface.
	ClassStandalone EnvironmentType = "Standalone"
	// ClassLocalVM -> the environment is constituted by a Virtual Machine started from a local Golden Image.
	ClassLocalVM EnvironmentType = "LocalVM"
)

// CleanupOptions defines the automatic actions to enforce termination policies.
type CleanupOptions struct {
	// +kubebuilder:validation:Pattern="^(never|[0-9]+[smhd])$"
	// +kubebuilder:default="never"
	// The maximum lifetime of an Instance referencing the current Template.
	// Once this period is expired, the Instance will be automatically deleted
	// regardless of whether it is still in use or not.
	// If set to "never", the instance will not be automatically terminated.
	DeleteAfterCreation string `json:"deleteAfterCreation"`

	// +kubebuilder:validation:Pattern="^(never|[0-9]+[smhd])$"
	// +kubebuilder:default="never"
	// The maximum period of inactivity after which an Instance referencing
	// the current Template will be automatically stopped, or deleted
	// (if not persistent) to save resources.
	StopAfterInactivity string `json:"stopAfterInactivity"`

	// +kubebuilder:validation:Pattern="^(never|[0-9]+[smhd])$"
	// +kubebuilder:default="never"
	// The maximum period of time a persistent instance can remain powered off
	// after being stopped for inactivity, before being completely deleted.
	DeleteAfterInactivity string `json:"deleteAfterInactivity"`
}

// +kubebuilder:validation:Enum=Fixed;Editable;ImageDefault

// JobCommandPolicy is an enumeration of the ways the command of a job started
// from a Template can be chosen.
type JobCommandPolicy string

const (
	// JobCommandFixed -> the job runs the command of the Template, and the user cannot change it.
	JobCommandFixed JobCommandPolicy = "Fixed"
	// JobCommandEditable -> the user types the command when starting the job; the command of the Template, if any, pre-fills it.
	JobCommandEditable JobCommandPolicy = "Editable"
	// JobCommandImageDefault -> the job has no command, and the entrypoint of the image runs.
	JobCommandImageDefault JobCommandPolicy = "ImageDefault"
)

// TemplateJob defines how a batch job is started from an Instance of the Template.
// +kubebuilder:validation:XValidation:rule="!has(self.commandPolicy) || self.commandPolicy != 'Fixed' || (has(self.command) && size(self.command) > 0)",message="commandPolicy Fixed requires a non-empty command"
// +kubebuilder:validation:XValidation:rule="!has(self.commandPolicy) || self.commandPolicy != 'ImageDefault' || !has(self.command) || size(self.command) == 0",message="commandPolicy ImageDefault requires no command: the entrypoint of the image runs"
type TemplateJob struct {
	// +kubebuilder:validation:Pattern="^[a-z\\d][a-z\\d-]{2,10}[a-z\\d]$"
	// The name of the environment (in environmentList) whose instance is copied
	// to run the job. It must be a persistent environment of type Container or
	// Standalone.
	EnvironmentName string `json:"environmentName"`

	// +kubebuilder:default=Editable
	// How the command of the job is chosen:
	// Fixed: the job runs command, and the user cannot change it;
	// Editable: the user types the command when starting the job, and command, if set, pre-fills it;
	// ImageDefault: the job has no command, and the entrypoint of the image runs
	// (it can detect the job through the environment variable CROWNLABS_JOB=true).
	CommandPolicy JobCommandPolicy `json:"commandPolicy,omitempty"`

	// +kubebuilder:validation:MaxItems:=64
	// +kubebuilder:validation:items:MaxLength:=8192
	// The command of the job, as a list of arguments (e.g., ["python", "train.py"]).
	// It is required if commandPolicy is Fixed, optional if Editable, and must
	// be omitted (or empty) if ImageDefault. An empty list is the same as no command.
	Command []string `json:"command,omitempty"`

	// The extended resources the job requests in addition to those of the
	// environment, by resource name (e.g., nvidia.com/gpu: 1).
	OtherResources map[string]resource.Quantity `json:"otherResources,omitempty"`

	// +kubebuilder:validation:Pattern="^[1-9][0-9]{0,3}[mhd]$"
	// The maximum time a job started from this Template may run, as a positive
	// number of at most four digits followed by a unit among m, h and d (e.g., 30m, 6h, 2d).
	// It cannot exceed the maxRunTime of the job queue of the Workspace (this
	// is checked by the operator). If empty, the maxRunTime of the Workspace applies.
	MaxRunTime string `json:"maxRunTime,omitempty"`
}

// TemplateSpec is the specification of the desired state of the Template.
// +kubebuilder:validation:XValidation:rule="!has(self.job) || self.environmentList.exists(e, e.name == self.job.environmentName)",message="job.environmentName must be the name of an environment in environmentList"
// +kubebuilder:validation:XValidation:rule="!has(self.job) || self.environmentList.all(e, e.name != self.job.environmentName || (has(e.persistent) && e.persistent && (e.environmentType == 'Container' || e.environmentType == 'Standalone')))",message="job.environmentName must name a persistent environment of type Container or Standalone"
type TemplateSpec struct {
	// The human-readable name of the Template.
	PrettyName string `json:"prettyName"`

	// A textual description of the Template.
	Description string `json:"description"`

	// The reference to the Workspace this Template belongs to.
	WorkspaceRef GenericRef `json:"workspace.crownlabs.polito.it/WorkspaceRef,omitempty"`

	// The list of environments (i.e. VMs or containers) that compose the Template.
	// Each environment must have a unique name within the Template.
	// +kubebuilder:validation:MaxItems:=10
	// +listType=map
	// +listMapKey=name
	EnvironmentList []Environment `json:"environmentList"`

	// +kubebuilder:default={deleteAfterCreation: "never", stopAfterInactivity: "never", deleteAfterInactivity: "never"}
	// Automatic actions to enforce termination policies.
	Cleanup CleanupOptions `json:"cleanup"`

	// Labels that are used for the selection of the node.
	// They are given by means of a pointer to check the presence of the field.
	// In case it is present, the labels that are chosen are the ones present on the instance
	NodeSelector *map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:default=false
	// Whether the Template has the authorization to be Public Exposed or not, using a LoadBalancer service.
	AllowPublicExposure bool `json:"allowPublicExposure,omitempty"`

	// How batch jobs are started from the Instances of this Template.
	// If omitted, no job can be started from this Template.
	Job *TemplateJob `json:"job,omitempty"`
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
	// Container, CloudVM, LocalVM and Standalone.
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
	apicommon.ResourceSpec `json:",inline"`

	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=100

	// The percentage of reserved CPU cores, ranging between 1 and 100, with
	// respect to the 'CPU' value. Essentially, this corresponds to the 'requests'
	// specified for the actual pod representing the environment.
	ReservedCPUPercentage uint32 `json:"reservedCPUPercentage"`
}

// UnmarshalJSON implements custom unmarshaling to prevent method promotion issues.
// TODO: Remove this workaround once the production mutation webhook is updated.
func (e *EnvironmentResources) UnmarshalJSON(data []byte) error {
	// 1. Unmarshal embedded ResourceSpec (handles cpu string/int, memory, disk, etc.)
	if err := json.Unmarshal(data, &e.ResourceSpec); err != nil {
		return err
	}

	// 2. Unmarshal EnvironmentResources specific fields
	var aux struct {
		ReservedCPUPercentage uint32 `json:"reservedCPUPercentage,omitempty"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	e.ReservedCPUPercentage = aux.ReservedCPUPercentage
	return nil
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
