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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum="";"Importing";"Starting";"ResourceQuotaExceeded";"Running";"Ready";"Stopping";"Off";"Failed";"CreationLoopBackoff"

// EnvironmentPhase is an enumeration of the different phases associated with
// an instance of a given environment template.
type EnvironmentPhase string

const (
	// EnvironmentPhaseUnset -> the environment phase is unknown.
	EnvironmentPhaseUnset EnvironmentPhase = ""
	// EnvironmentPhaseImporting -> the image of the environment is being imported.
	EnvironmentPhaseImporting EnvironmentPhase = "Importing"
	// EnvironmentPhaseStarting -> the environment is starting.
	EnvironmentPhaseStarting EnvironmentPhase = "Starting"
	// EnvironmentPhaseResourceQuotaExceeded -> the environment could not start because the resource quota is exceeded.
	EnvironmentPhaseResourceQuotaExceeded EnvironmentPhase = "ResourceQuotaExceeded"
	// EnvironmentPhaseRunning -> the environment is running, but not yet ready.
	EnvironmentPhaseRunning EnvironmentPhase = "Running"
	// EnvironmentPhaseReady -> the environment is ready to be accessed.
	// with the current CrownLabs dashboard.
	EnvironmentPhaseReady EnvironmentPhase = "Ready"
	// EnvironmentPhaseStopping -> the environment is being stopped.
	EnvironmentPhaseStopping EnvironmentPhase = "Stopping"
	// EnvironmentPhaseOff -> the environment is currently shut down.
	// with the current CrownLabs dashboard.
	EnvironmentPhaseOff EnvironmentPhase = "Off"
	// EnvironmentPhaseFailed -> the environment has failed, and cannot be restarted.
	EnvironmentPhaseFailed EnvironmentPhase = "Failed"
	// EnvironmentPhaseCreationLoopBackoff -> the environment has encountered a temporary error during creation.
	EnvironmentPhaseCreationLoopBackoff EnvironmentPhase = "CreationLoopBackoff"
)

// InstanceCustomizationUrls specifies optional urls for advanced integration features.
type InstanceCustomizationUrls struct {
	// URL from which GET the archive to be extracted into Template.ContainerStartupOptions.ContentPath. This field, if set, OVERRIDES Template.ContainerStartupOptions.SourceArchiveURL.
	ContentOrigin string `json:"contentOrigin,omitempty"`

	// URL to which POST an archive with the contents found (at instance termination) in Template.ContainerStartupOptions.ContentPath.
	ContentDestination string `json:"contentDestination,omitempty"`

	// URL which is periodically checked (with a GET request) to determine automatic instance shutdown. Should return any 2xx status code if the instance has to keep running, any 4xx otherwise. In case of 2xx response, it should output a JSON with a `deadline` field containing a ISO_8601 compliant date/time string of the expected instance termination time. See instautoctrl.StatusCheckResponse for exact definition.
	StatusCheck string `json:"statusCheck,omitempty"`
}

// InstanceSpec is the specification of the desired state of the Instance.
type InstanceSpec struct {
	// The reference to the Template to be instantiated.
	Template GenericRef `json:"template.crownlabs.polito.it/TemplateRef"`

	// The reference to the Tenant which owns the Instance object.
	Tenant GenericRef `json:"tenant.crownlabs.polito.it/TenantRef"`

	// +kubebuilder:default=true
	// +kubebuilder:validation:Optional

	// Whether the current instance is running or not.
	// The meaning of this flag is different depending on whether the instance
	// refers to a persistent environment or not. If the first case, it allows to
	// stop the environment (e.g. the underlying VM) without deleting the associated
	// disk. Setting the flag to true will restart the environment, attaching it
	// to the same disk used previously. Differently, if the environment is not
	// persistent, it only tears down the exposition objects, making the instance
	// effectively unreachable from outside the cluster, but allowing the
	// subsequent recreation without data loss.
	Running bool `json:"running"`

	// Custom name the user can assign and change at any time
	// in order to more easily identify the instance.
	// +kubebuilder:validation:Optional
	PrettyName string `json:"prettyName"`

	// Labels that are used for the selection of the node.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Optional urls for advanced integration features.
	CustomizationUrls *InstanceCustomizationUrls `json:"customizationUrls,omitempty"`
}

// InstanceAutomationStatus reflects the status of the instance's automation (termination and submission).
type InstanceAutomationStatus struct {
	// The last time the Instance desired status was checked.
	LastCheckTime metav1.Time `json:"lastCheckTime,omitempty"`

	// The (possibly expected) termination time of the Instance.
	TerminationTime metav1.Time `json:"terminationTime,omitempty"`

	// The time the Instance content submission has been completed.
	SubmissionTime metav1.Time `json:"submissionTime,omitempty"`
}

// InstanceStatus reflects the most recently observed status of the Instance.
type InstanceStatus struct {
	// The current status Instance, with reference to the associated environment
	// (e.g. VM). This conveys which resource is being created, as well as
	// whether the associated VM is being scheduled, is running or ready to
	// accept incoming connections.
	Phase EnvironmentPhase `json:"phase,omitempty"`

	// The URL where it is possible to access the remote desktop of the instance
	// (in case of graphical environments)
	URL string `json:"url,omitempty"`

	// The internal IP address associated with the remote environment, which can
	// be used to access it through the SSH protocol (leveraging the SSH bastion
	// in case it is not contacted from another CrownLabs Instance).
	IP string `json:"ip,omitempty"`

	// The amount of time the Instance required to become ready for the first time
	// upon creation.
	InitialReadyTime string `json:"initialReadyTime,omitempty"`

	// Timestamps of the Instance automation phases (check, termination and submission).
	Automation InstanceAutomationStatus `json:"automation,omitempty"`

	// The node on which the Instance is running.
	NodeName string `json:"nodeName,omitempty"`

	// The actual nodeSelector assigned to the Instance.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="inst"
// +kubebuilder:printcolumn:name="Pretty Name",type=string,JSONPath=`.spec.prettyName`
// +kubebuilder:printcolumn:name="Running",type=string,JSONPath=`.spec.running`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.url`,priority=10
// +kubebuilder:printcolumn:name="IP Address",type=string,JSONPath=`.status.ip`,priority=10
// +kubebuilder:printcolumn:name="Ready In",type=string,JSONPath=`.status.initialReadyTime`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Instance describes the instance of a CrownLabs environment Template.
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InstanceList contains a list of Instance objects.
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
