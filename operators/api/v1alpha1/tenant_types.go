// Copyright 2020-2022 Politecnico di Torino
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TenantWorkspaceEntry contains the information regarding one of the Workspaces
// the Tenant is subscribed to, including his/her role.
type TenantWorkspaceEntry struct {
	// The reference to the Workspace resource the Tenant is subscribed to.
	WorkspaceRef v1alpha2.GenericRef `json:"workspaceRef"`

	// The role of the Tenant in the context of the Workspace.
	Role v1alpha2.WorkspaceUserRole `json:"role"`

	// The number of the group the Tenant belongs to. Empty means no group.
	GroupNumber uint `json:"groupNumber,omitempty"`
}

// TenantSpec is the specification of the desired state of the Tenant.
type TenantSpec struct {
	// The first name of the Tenant.
	FirstName string `json:"firstName"`

	// The last name of the Tenant.
	LastName string `json:"lastName"`

	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

	// The email associated with the Tenant, which will be used to log-in
	// into the system.
	Email string `json:"email"`

	// The list of the Workspaces the Tenant is subscribed to, along with his/her
	// role in each of them.
	Workspaces []TenantWorkspaceEntry `json:"workspaces,omitempty"`

	// The list of the SSH public keys associated with the Tenant. These will be
	// used to enable to access the remote environments through the SSH protocol.
	PublicKeys []string `json:"publicKeys,omitempty"`

	// +kubebuilder:default=false

	// Whether a sandbox namespace should be created to allow the Tenant play
	// with Kubernetes.
	CreateSandbox bool `json:"createSandbox,omitempty"`

	// The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in.
	Quota *v1alpha2.TenantResourceQuota `json:"quota,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:printcolumn:name="First Name",type=string,JSONPath=`.spec.firstName`
// +kubebuilder:printcolumn:name="Last Name",type=string,JSONPath=`.spec.lastName`
// +kubebuilder:printcolumn:name="Email",type=string,JSONPath=`.spec.email`,priority=10
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.status.personalNamespace.name`,priority=10
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Tenant describes a user of CrownLabs.
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec            `json:"spec,omitempty"`
	Status v1alpha2.TenantStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TenantList contains a list of Tenant objects.
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}

// SetupWebhookWithManager setups the webhook with the given manager.
func (r *Tenant) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
