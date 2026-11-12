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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CrownJobSpec defines the desired state of CrownJob.
type CrownJobSpec struct {
	// Instance is the reference to the Instance whose environment is copied to
	// run the job. If the namespace is empty, the namespace of the CrownJob is used.
	Instance GenericRef `json:"instanceRef"`

	// Command is the command of the job, as a list of arguments (e.g., ["python", "train.py"]).
	// It is used only if the commandPolicy of the job section of the Template
	// is Editable (this is checked by the operator).
	// +optional
	// +kubebuilder:validation:MaxItems=64
	// +kubebuilder:validation:items:MaxLength=8192
	Command []string `json:"command,omitempty"`

	// MaxRunTime is the maximum time this job may run, as a positive number of
	// at most four digits followed by a unit among m, h and d (e.g., 30m, 6h, 2d).
	// It can only lower the maxRunTime of the Template (this is checked by the
	// operator). If empty, the maxRunTime of the Template applies.
	// +optional
	// +kubebuilder:validation:Pattern="^[1-9][0-9]{0,3}[mhd]$"
	MaxRunTime string `json:"maxRunTime,omitempty"`
}

// CrownJobPhase describes the current phase of the CrownJob.
// +kubebuilder:validation:Enum=Waiting;Starting;Running;Succeeded;Failed
type CrownJobPhase string

const (
	// CrownJobPhaseWaiting means the job is waiting in the queue of the Workspace.
	CrownJobPhaseWaiting CrownJobPhase = "Waiting"
	// CrownJobPhaseStarting means the job has been admitted and is starting.
	CrownJobPhaseStarting CrownJobPhase = "Starting"
	// CrownJobPhaseRunning means the job is running.
	CrownJobPhaseRunning CrownJobPhase = "Running"
	// CrownJobPhaseSucceeded means the job has completed successfully.
	CrownJobPhaseSucceeded CrownJobPhase = "Succeeded"
	// CrownJobPhaseFailed means the job has ended without succeeding.
	CrownJobPhaseFailed CrownJobPhase = "Failed"
)

// CrownJobStatus defines the observed state of CrownJob.
type CrownJobStatus struct {
	// Phase is the current phase of the job.
	// +optional
	Phase CrownJobPhase `json:"phase,omitempty"`

	// Reason is the reason for the current phase, as a single CamelCase word
	// (e.g., InQueue, QuotaReached, ExitCode, TimeLimit, Canceled, CannotStart).
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a plain sentence for the user describing the current phase and its reason.
	// +optional
	Message string `json:"message,omitempty"`

	// QueuePosition is the position of the job in the queue of the Workspace while it is waiting.
	// +optional
	QueuePosition *int32 `json:"queuePosition,omitempty"`

	// Deadline is the time at which the job will be stopped if it is still running.
	// +optional
	Deadline *metav1.Time `json:"deadline,omitempty"`

	// RunFolder is the path, on the disk of the Instance, of the folder that
	// holds the logs and outputs of this run (e.g., runs/<name>).
	// +optional
	RunFolder string `json:"runFolder,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// +kubebuilder:printcolumn:name="Instance",type=string,JSONPath=`.spec.instanceRef.name`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// CrownJob is the Schema for the crownjobs API.
// Each CrownJob is one run of a batch job started from an Instance.
type CrownJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="CrownJob spec is immutable"
	Spec   CrownJobSpec   `json:"spec,omitempty"`
	Status CrownJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CrownJobList contains a list of CrownJob.
type CrownJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CrownJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CrownJob{}, &CrownJobList{})
}
