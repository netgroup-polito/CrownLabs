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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	// EnvironmentPhaseRunning -> the environment is running, but not yet ready.
	EnvironmentPhaseRunning EnvironmentPhase = "Running"
	// EnvironmentPhaseReady -> the environment is ready to be accessed.
	// TODO: update the value to Ready once no more needed for backward compatibility
	// with the current CrownLabs dashboard.
	EnvironmentPhaseReady EnvironmentPhase = "VmiReady"
	// EnvironmentPhaseStopping -> the environment is being stopped.
	EnvironmentPhaseStopping EnvironmentPhase = "Stopping"
	// EnvironmentPhaseOff -> the environment is currently shut down.
	// TODO: update the value to Off once no more needed for backward compatibility
	// with the current CrownLabs dashboard.
	EnvironmentPhaseOff EnvironmentPhase = "VmiOff"
	// EnvironmentPhaseFailed -> the environment has failed, and cannot be restarted.
	EnvironmentPhaseFailed EnvironmentPhase = "Failed"
	// EnvironmentPhaseCreationLoopBackoff -> the environment has encountered a temporary error during creation.
	EnvironmentPhaseCreationLoopBackoff EnvironmentPhase = "CreationLoopBackoff"
)

// InstanceSpec is the specification of the desired state of the Instance.
type InstanceSpec struct {
	// The reference to the Template to be instantiated.
	Template GenericRef `json:"template.crownlabs.polito.it/TemplateRef"`

	// The reference to the Tenant which owns the Instance object.
	Tenant GenericRef `json:"tenant.crownlabs.polito.it/TenantRef"`

	// +kubebuilder:default=true
	// +kubebuilder:validation:Optional

	// Whether the current instance is running or not. This field is meaningful
	// only in case the Instance refers to persistent environments, and it allows
	// to stop the environment (e.g. the underlying VM) without deleting the
	// associated disk. Setting the flag to true will restart the environment,
	// attaching it to the same disk used previously. The flag, on the other hand,
	// is silently ignored in case of non-persistent environments, as the state
	// cannot be preserved among reboots.
	Running bool `json:"running"`
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
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="inst"
// +kubebuilder:printcolumn:name="Running",type=string,JSONPath=`.spec.running`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.url`,priority=10
// +kubebuilder:printcolumn:name="IP Address",type=string,JSONPath=`.status.ip`,priority=10
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
