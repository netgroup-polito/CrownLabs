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
	ctrl "sigs.k8s.io/controller-runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum=manager;user;candidate

// WorkspaceUserRole is an enumeration of the different roles that can be
// associated to a Tenant in a Workspace.
type WorkspaceUserRole string

const (
	// Manager -> a Tenant with Manager role can interact with all the environments
	// (i.e. VMs) in a Workspace, as well as add new Tenants to the Workspace.
	Manager WorkspaceUserRole = "manager"
	// User -> a Tenant with User role can only interact with his/her own
	// environments (e.g. VMs) within that Workspace.
	User WorkspaceUserRole = "user"
	// Candidate -> a Tenant with Candidate role wants to be added to the
	// Workspace, but he/she is not yet enrolled.
	Candidate WorkspaceUserRole = "candidate"

	// SVCTenantName -> name of a system/service tenant to which other resources might belong.
	SVCTenantName string = "service-tenant"
)

// TenantWorkspaceEntry contains the information regarding one of the Workspaces
// the Tenant is subscribed to, including his/her role.
type TenantWorkspaceEntry struct {
	// The Workspace the Tenant is subscribed to.
	Name string `json:"name"`

	// The role of the Tenant in the context of the Workspace.
	Role WorkspaceUserRole `json:"role"`
}

// TenantSpec is the specification of the desired state of the Tenant.
type TenantSpec struct {
	// The first name of the Tenant.
	FirstName string `json:"firstName"`

	// The last name of the Tenant.
	LastName string `json:"lastName"`

	// The last login timestamp.
	LastLogin metav1.Time `json:"lastLogin,omitempty"`

	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

	// The email associated with the Tenant, which will be used to log-in
	// into the system.
	Email string `json:"email"`

	// The list of the Workspaces the Tenant is subscribed to, along with his/her
	// role in each of them.
	// +listType=map
	// +listMapKey=name
	Workspaces []TenantWorkspaceEntry `json:"workspaces,omitempty"`

	// The list of the SSH public keys associated with the Tenant. These will be
	// used to enable to access the remote environments through the SSH protocol.
	PublicKeys []string `json:"publicKeys,omitempty"`

	// +kubebuilder:default=false

	// Whether a sandbox namespace should be created to allow the Tenant play
	// with Kubernetes.
	CreateSandbox bool `json:"createSandbox,omitempty"`

	// The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in.
	Quota *TenantResourceQuota `json:"quota,omitempty"`
}

// TenantResourceQuota defines resource quota for each Tenant.
type TenantResourceQuota struct {
	// The maximum amount of CPU which can be used by this Tenant.
	CPU resource.Quantity `json:"cpu"`

	// The maximum amount of RAM memory which can be used by this Tenant.
	Memory resource.Quantity `json:"memory"`

	// +kubebuilder:validation:Minimum:=0
	// The maximum number of concurrent instances which can be created by this Tenant.
	Instances uint32 `json:"instances"`
}

// TenantStatus reflects the most recently observed status of the Tenant.
type TenantStatus struct {
	// The namespace containing all CrownLabs related objects of the Tenant.
	// This is the namespace that groups his/her own Instances, together with
	// all the accessory resources (e.g. RBACs, resource quota, network policies,
	// ...) created by the tenant-operator.
	PersonalNamespace NameCreated `json:"personalNamespace"`

	// The namespace that can be freely used by the Tenant to play with Kubernetes.
	// This namespace is created only if the .spec.CreateSandbox flag is true.
	SandboxNamespace NameCreated `json:"sandboxNamespace"`

	// The list of Workspaces that are throwing errors during subscription.
	// This mainly happens if .spec.Workspaces contains references to Workspaces
	// which do not exist.
	FailingWorkspaces []string `json:"failingWorkspaces"`

	// The list of the subscriptions to external services (e.g. Keycloak,
	// ...), indicating for each one whether it succeeded or an error
	// occurred.
	Subscriptions map[string]SubscriptionStatus `json:"subscriptions"`

	// Whether all subscriptions and resource creations succeeded or an error
	// occurred. In case of errors, the other status fields provide additional
	// information about which problem occurred.
	// Will be set to true even when personal workspace is intentionally deleted.
	Ready bool `json:"ready"`

	// The amount of resources associated with this Tenant, either inherited from the Workspaces in which he/she is enrolled, or manually overridden.
	Quota TenantResourceQuota `json:"quota,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
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

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
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
