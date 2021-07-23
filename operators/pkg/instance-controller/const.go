// Package instance_controller groups the functionalities related to the Instance controller.
package instance_controller

const (
	// EvTmplNotFound -> the event key corresponding to a not found template.
	EvTmplNotFound = "TemplateNotFound"
	// EvTmplNotFoundMsg -> the event message corresponding to a not found template.
	EvTmplNotFoundMsg = "Template %v/%v not found"

	// EvEnvironmentErr -> the event key corresponding to a failed environment enforcement.
	EvEnvironmentErr = "EnvironmentEnforcementFailed"
	// EvEnvironmentErrMsg -> the event message corresponding to a failed environment enforcement.
	EvEnvironmentErrMsg = "Failed to enforce environment %v"
)
