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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InstanceSpec is the specification of the desired state of the Instance.
type InstanceSpec struct {
	// The reference to the Template to be instantiated.
	Template GenericRef `json:"template.crownlabs.polito.it/TemplateRef"`

	// The reference to the Tenant which owns the Instance object.
	Tenant GenericRef `json:"tenant.crownlabs.polito.it/TenantRef"`
}

// InstanceStatus reflects the most recently observed status of the Instance.
type InstanceStatus struct {
	// The current status Instance, with reference to the associated environment
	// (e.g. VM). This conveys which resource is being created, as well as
	// whether the associated VM is being scheduled, is running or ready to
	// accept incoming connections.
	Phase string `json:"phase,omitempty"`

	// The URL where it is possible to access the remote desktop of the instance
	// (in case of graphical environments)
	URL string `json:"url,omitempty"`

	// The internal IP address associated with the remote environment, which can
	// be used to access it through the SSH protocol (leveraging the SSH bastion
	// in case it is not contacted from another CrownLabs Instance).
	IP string `json:"ip,omitempty"`

	// The generation of the object which has been observed. This is used to
	// filter out the reconciliation events referring only to status modifications.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="inst"
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.url`
// +kubebuilder:printcolumn:name="IP Address",type=string,JSONPath=`.status.ip`

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
