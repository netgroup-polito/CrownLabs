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

// GenericRef represents a reference to a generic Kubernetes resource,
// and it is composed of the resource name and (optionally) its namespace.
type GenericRef struct {
	// The name of the resource to be referenced.
	Name string `json:"name"`

	// The namespace containing the resource to be referenced. It should be left
	// empty in case of cluster-wide resources.
	Namespace string `json:"namespace,omitempty"`
}

// NameCreated contains information about the status of a resource created in
// the cluster (e.g. a namespace). Specifically, it contains the name of the
// resource and a flag indicating whether the creation succeeded.
type NameCreated struct {
	// The name of the considered resource.
	Name string `json:"name,omitempty"`

	// Whether the creation succeeded or not.
	Created bool `json:"created"`
}

// +kubebuilder:validation:Enum=Ok;Failed

// SubscriptionStatus is an enumeration of the different states that can be
// assumed by the subscription to a service (e.g. successful or failing).
type SubscriptionStatus string

const (
	// SubscrOk -> the subscription was successful.
	SubscrOk SubscriptionStatus = "Ok"
	// SubscrFailed -> the subscription has failed.
	SubscrFailed SubscriptionStatus = "Failed"
)

// TemplateLabelPrefix is the prefix of a label assigned to a sharedvolume indicating it is mounted on a template.
const TemplateLabelPrefix = "crownlabs.polito.it/template-"

// WorkspaceLabelPrefix is the prefix of a label assigned to a tenant indicating it is subscribed to a workspace.
const WorkspaceLabelPrefix = "crownlabs.polito.it/workspace-"

// WorkspaceLabelAutoenroll is the label assigned to a workspace in which autoenroll is enabled.
const WorkspaceLabelAutoenroll = "crownlabs.polito.it/autoenroll"

// TnOperatorFinalizerName is the name of the finalizer corresponding to the tenant operator.
const TnOperatorFinalizerName = "crownlabs.polito.it/tenant-operator"

// ShVolCtrlFinalizerName is the name of the finalizer for SharedVolume's PVC protection.
const ShVolCtrlFinalizerName = "crownlabs.polito.it/shvolctrl-volume-protection"
