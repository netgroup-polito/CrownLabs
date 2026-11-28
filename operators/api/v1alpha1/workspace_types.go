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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
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
	Quota apicommon.WorkspaceResourceQuota `json:"quota"`

	// The batch jobs ("Run job") of the members of this Workspace: their
	// priority and the limits of their queue.
	// If omitted, no job can be run in this Workspace.
	JobOptions *JobOptions `json:"jobOptions,omitempty"`
}

// +kubebuilder:validation:Enum=Mon;Tue;Wed;Thu;Fri;Sat;Sun

// WeekDay is a day of the week, as used by the job windows of a Workspace.
type WeekDay string

const (
	// WeekDayMonday -> Monday.
	WeekDayMonday WeekDay = "Mon"
	// WeekDayTuesday -> Tuesday.
	WeekDayTuesday WeekDay = "Tue"
	// WeekDayWednesday -> Wednesday.
	WeekDayWednesday WeekDay = "Wed"
	// WeekDayThursday -> Thursday.
	WeekDayThursday WeekDay = "Thu"
	// WeekDayFriday -> Friday.
	WeekDayFriday WeekDay = "Fri"
	// WeekDaySaturday -> Saturday.
	WeekDaySaturday WeekDay = "Sat"
	// WeekDaySunday -> Sunday.
	WeekDaySunday WeekDay = "Sun"
)

// JobOptions defines the batch jobs run by the members of a Workspace.
type JobOptions struct {
	// +kubebuilder:default=100
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	// The priority of the jobs of the members of this Workspace, from 0 to
	// 10000 (default 100). All the jobs of the Workspace have this priority.
	// When jobs of several Workspaces wait for the same shared resources, the
	// job with the higher value is admitted first. The operator creates one
	// Kueue WorkloadPriorityClass per Workspace, named workspace-<name of the
	// Workspace>, with this value: it orders the jobs in the queues of Kueue,
	// and does not change the priority of their pods. The job windows cannot
	// change it.
	Priority *int32 `json:"priority,omitempty"`

	// The limits of the job queue of this Workspace.
	QueueLimits QueueLimits `json:"queueLimits"`

	// +kubebuilder:validation:MaxItems:=20
	// The periods of the week during which some limits of the job queue
	// differ from those in queueLimits (e.g., more GPUs at night).
	// Their days and times are in UTC.
	OverrideWindows []JobWindow `json:"overrideWindows,omitempty"`
}

// QueueLimits defines the limits of the job queue of a Workspace.
// The same type holds the defaults (spec.jobOptions.queueLimits) and the
// values that replace them during a window
// (spec.jobOptions.overrideWindows[].overrides).
type QueueLimits struct {
	// The maximum amount of resources that all the jobs of the Workspace may
	// use at the same time, by resource name (e.g., cpu, memory, nvidia.com/gpu).
	// If absent in a window's overrides, the value in queueLimits applies.
	MaxAtOnce map[string]resource.Quantity `json:"maxAtOnce,omitempty"`

	// The amount of resources reserved for the jobs of the Workspace, by
	// resource name. If absent in queueLimits, nothing is reserved, and the
	// jobs only use resources that other Workspaces are not using; if absent
	// in a window's overrides, the value in queueLimits applies.
	// The operator checks that it does not exceed maxAtOnce, on the values in
	// force at each moment.
	Guaranteed map[string]resource.Quantity `json:"guaranteed,omitempty"`

	// The most that one job may run (maxRunTime) and ask for (resources).
	apicommon.JobLimits `json:",inline"`

	// The limits that apply to each member of the Workspace. If set in a
	// window's overrides, it replaces the whole perUser of queueLimits.
	PerUser *PerUserLimits `json:"perUser,omitempty"`
}

// PerUserLimits defines the job queue limits that apply to each member of a Workspace.
type PerUserLimits struct {
	// The maximum amount of resources that one member may use in total across
	// all of their running jobs, by resource name.
	Resources map[string]resource.Quantity `json:"resources,omitempty"`

	// +kubebuilder:validation:Minimum:=1
	// The maximum number of jobs (waiting and running) that one member may
	// have at the same time. This limit is not enforced yet: enforcement will
	// be added in a later version of the operator.
	MaxActiveJobs *int32 `json:"maxActiveJobs,omitempty"`
}

// JobWindow defines a period of the week during which some limits of the job
// queue of a Workspace differ from the defaults. On each of its days, a window
// covers [start, end): start is included and end is excluded. Days and times
// are in UTC.
// +kubebuilder:validation:XValidation:rule="self.start < self.end",message="the window start must be earlier than its end: a window cannot cross midnight"
type JobWindow struct {
	// +kubebuilder:validation:MinItems:=1
	// +kubebuilder:validation:MaxItems:=7
	// +listType=set
	// The days of the week on which the window applies (Mon, Tue, Wed, Thu,
	// Fri, Sat, Sun), in UTC.
	Days []WeekDay `json:"days"`

	// +kubebuilder:validation:Pattern="^([01][0-9]|2[0-3]):[0-5][0-9]$"
	// +kubebuilder:validation:MaxLength:=5
	// The time of the day at which the window starts (included), as "HH:MM"
	// (24-hour clock, from "00:00" to "23:59"), in UTC. Quote the value in
	// YAML.
	Start string `json:"start"`

	// +kubebuilder:validation:Pattern="^(([01][0-9]|2[0-3]):[0-5][0-9]|24:00)$"
	// +kubebuilder:validation:MaxLength:=5
	// The time of the day at which the window ends (excluded), as "HH:MM"
	// (24-hour clock, from "00:01" to "24:00"), in UTC. Quote the value in
	// YAML. It must be later than start on the same day: a window cannot
	// cross midnight (UTC), so a night needs two windows (e.g.,
	// "20:00"-"24:00" and "00:00"-"06:00").
	End string `json:"end"`

	// The limits in force during the window. Each field set here (maxAtOnce,
	// guaranteed, maxRunTime, resources, perUser) replaces the corresponding default of
	// spec.jobOptions.queueLimits as a whole, without merging the content; the
	// fields omitted here keep the default value.
	Overrides QueueLimits `json:"overrides"`
}

// WorkspaceStatus reflects the most recently observed status of the Workspace.
type WorkspaceStatus struct {
	// The namespace containing all CrownLabs related objects of the Workspace.
	// This is the namespace that groups multiple related templates, together
	// with all the accessory resources (e.g. RBACs) created by the tenant
	// operator.
	Namespace clv1alpha2.NameCreated `json:"namespace,omitempty"`

	// The list of the subscriptions to external services (e.g. Keycloak,
	// ...), indicating for each one whether it succeeded or an error
	// occurred.
	Subscriptions map[string]clv1alpha2.SubscriptionStatus `json:"subscription,omitempty"`

	// Whether all subscriptions and resource creations succeeded or an error
	// occurred. In case of errors, the other status fields provide additional
	// information about which problem occurred.
	Ready bool `json:"ready,omitempty"`
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
