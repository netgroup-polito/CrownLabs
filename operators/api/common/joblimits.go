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

package common

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// JobLimits defines how long a batch job may run and which resources it may use.
// The same type is embedded in the job section of a Template (spec.job) and
// is the type of the jobLimits of the job queue of a Workspace
// (spec.jobOptions.queueLimits.jobLimits and
// spec.jobOptions.overrideWindows[].overrides.jobLimits), so that both are
// written and validated in the same way.
// +k8s:deepcopy-gen=true
type JobLimits struct {
	// +kubebuilder:validation:Pattern="^[1-9][0-9]{0,3}[mhd]$"
	// The maximum time a job may run, as a positive number of at most four
	// digits followed by a unit among m, h and d (e.g., 30m, 6h, 2d). It is
	// counted from the start of the pod of the job, so its set-up (mounting
	// the disk, downloading the image) is included. A job still running after
	// this time is stopped.
	// In the job section of a Template, it is the time limit of the jobs
	// started from the Template: the operator applies the smaller of this
	// value and the maxRunTime of the Workspace, and the maxRunTime of the
	// Workspace if this one is absent.
	// In the jobLimits of a Workspace, it is the most that one job may run.
	// If absent in both the Template and the Workspace, the default of the
	// operator applies.
	MaxRunTime string `json:"maxRunTime,omitempty"`

	// The resources of a job, by resource name (e.g., cpu, memory,
	// nvidia.com/gpu).
	// In the job section of a Template, they are the resources of each job
	// started from the Template: a listed resource replaces the value of the
	// Instance for the job, with requests equal to limits; the resources not
	// listed keep the values of the Instance; a quantity of 0 of an extended
	// resource (a name with a "/", e.g., nvidia.com/gpu: 0) means that the job
	// does not ask for it, even if the Instance does.
	// In the jobLimits of a Workspace, they are the most that one job may ask
	// for (this is checked by the operator).
	Resources map[string]resource.Quantity `json:"resources,omitempty"`
}
