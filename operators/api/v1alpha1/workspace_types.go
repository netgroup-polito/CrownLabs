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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum="";withApproval;immediate

// WorkspaceAutoenroll defines auto-enroll capabilities of the Workspace.
type WorkspaceAutoenroll string

const (
	// AutoenrollNone -> no auto-enroll capabilities.
	AutoenrollNone WorkspaceAutoenroll = ""

	// AutoenrollWithApproval -> auto-enroll is enabled, but requires approval by a manager.
	AutoenrollWithApproval WorkspaceAutoenroll = "withApproval"

	// AutoenrollImmediate -> auto-enroll is enabled, and the user can enroll himself/herself.
	AutoenrollImmediate WorkspaceAutoenroll = "immediate"
)

// WorkspaceSpec is the specification of the desired state of the Workspace.
type WorkspaceSpec struct {
	// The human-readable name of the Workspace.
	PrettyName string `json:"prettyName"`

	// AutoEnroll capability definition. If omitted, no autoenroll features will be added.
	AutoEnroll WorkspaceAutoenroll `json:"autoEnroll,omitempty"`

	// The amount of resources associated with this workspace, and inherited by enrolled tenants.
	Quota WorkspaceResourceQuota `json:"quota"`
}

// WorkspaceStatus reflects the most recently observed status of the Workspace.
type WorkspaceStatus struct {
	// The namespace containing all CrownLabs related objects of the Workspace.
	// This is the namespace that groups multiple related templates, together
	// with all the accessory resources (e.g. RBACs) created by the tenant
	// operator.
	Namespace v1alpha2.NameCreated `json:"namespace,omitempty"`

	// The list of the subscriptions to external services (e.g. Keycloak,
	// ...), indicating for each one whether it succeeded or an error
	// occurred.
	Subscriptions map[string]v1alpha2.SubscriptionStatus `json:"subscription,omitempty"`

	// Whether all subscriptions and resource creations succeeded or an error
	// occurred. In case of errors, the other status fields provide additional
	// information about which problem occurred.
	Ready bool `json:"ready,omitempty"`
}

// WorkspaceResourceQuota defines the resource quota for each Workspace.
type WorkspaceResourceQuota struct {
	// The maximum amount of CPU required by this Workspace.
	CPU resource.Quantity `json:"cpu"`

	// The maximum amount of RAM memory required by this Workspace.
	Memory resource.Quantity `json:"memory"`

	// +kubebuilder:validation:Minimum:=1
	// The maximum number of concurrent instances required by this Workspace.
	Instances uint32 `json:"instances"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:printcolumn:name="Pretty Name",type=string,JSONPath=`.spec.prettyName`
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.status.namespace.name`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Workspace describes a workspace in CrownLabs.
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace objects.
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
