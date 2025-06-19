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

package instctrl

const (
	// EvTmplNotFound -> the event key corresponding to a not found template.
	EvTmplNotFound = "TemplateNotFound"
	// EvTmplNotFoundMsg -> the event message corresponding to a not found template.
	EvTmplNotFoundMsg = "Template %v/%v not found"

	// EvTntNotFound -> the event key corresponding to a not found tenant.
	EvTntNotFound = "TenantNotFound"
	// EvTntNotFoundMsg -> the event message corresponding to a not found tenant.
	EvTntNotFoundMsg = "Tenant %v not found"

	// EvEnvironmentErr -> the event key corresponding to a failed environment enforcement.
	EvEnvironmentErr = "EnvironmentEnforcementFailed"
	// EvEnvironmentErrMsg -> the event message corresponding to a failed environment enforcement.
	EvEnvironmentErrMsg = "Failed to enforce environment %v"
)
