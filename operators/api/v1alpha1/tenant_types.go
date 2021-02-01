/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
See the License for the specific language governing permissions and
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum=manager;user
type WorkspaceUserRole string

const (
	// Manager allows to interact with all VMs of a workspace
	Manager WorkspaceUserRole = "manager"
	// User allows to interact with owned vms
	User WorkspaceUserRole = "user"
)

// UserWorkspaceData contains the info of the workspaces related to a user
type UserWorkspaceData struct {
	// reference to the workspace resource in the cluster
	WorkspaceRef GenericRef `json:"workspaceRef"`
	GroupNumber  uint       `json:"groupNumber,omitempty"`
	// role of the user in the context of the workspace
	Role WorkspaceUserRole `json:"role"`
}

// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	FirstName string `json:"firstName"`

	LastName string `json:"lastName"`

	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	Email string `json:"email"`

	// list of workspace the tenant subscribed to
	Workspaces []UserWorkspaceData `json:"workspaces,omitempty"`

	// list of the public keys of the tenant for SSH access
	PublicKeys []string `json:"publicKeys,omitempty"`

	// should the resource create the sandbox namespace for k8s practice environment
	// +kubebuilder:default=false
	CreateSandbox bool `json:"createSandbox,omitempty"`
}

// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	// info about the namespace for the user resources inside the cluster
	PersonalNamespace NameCreated `json:"personalNamespace"`
	SandboxNamespace  NameCreated `json:"sandboxNamespace"`

	// list of workspace that are throwing errors during subscription
	FailingWorkspaces []string `json:"failingWorkspaces"`

	// list of subscriptions to non-k8s services (keycloak, nextcloud, ..)
	Subscriptions map[string]SubscriptionStatus `json:"subscriptions"`

	// false if there have been errors within the last reconcile, true otherwise
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:printcolumn:name="First Name",type=string,JSONPath=`.spec.firstName`
// +kubebuilder:printcolumn:name="Last Name",type=string,JSONPath=`.spec.lastName`
// +kubebuilder:printcolumn:name="Email",type=string,JSONPath=`.spec.email`
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.status.personalNamespace.name`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.ready`

// Tenant is the Schema for the tenants API
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TenantList contains a list of Tenant
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}
