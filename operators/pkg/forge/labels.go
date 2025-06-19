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

package forge

import (
	"strconv"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	labelManagedByKey    = "crownlabs.polito.it/managed-by"
	labelInstanceKey     = "crownlabs.polito.it/instance"
	labelWorkspaceKey    = "crownlabs.polito.it/workspace"
	labelTemplateKey     = "crownlabs.polito.it/template"
	labelTenantKey       = "crownlabs.polito.it/tenant"
	labelPersistentKey   = "crownlabs.polito.it/persistent"
	labelComponentKey    = "crownlabs.polito.it/component"
	labelMetricsEnabled  = "crownlabs.polito.it/metrics-enabled"
	labelTypeKey         = "crownlabs.polito.it/type"
	labelVolumeTypeKey   = "crownlabs.polito.it/volume-type"
	labelNodeSelectorKey = "crownlabs.polito.it/has-node-selector"

	// InstanceTerminationSelectorLabel -> label for Instances which have to be be checked for termination.
	InstanceTerminationSelectorLabel = "crownlabs.polito.it/watch-for-instance-termination"
	// InstanceSubmissionSelectorLabel -> label for Instances which have to be submitted.
	InstanceSubmissionSelectorLabel = "crownlabs.polito.it/instance-submission-requested"
	// InstanceSubmissionCompletedLabel -> label for Instances that have been submitted.
	InstanceSubmissionCompletedLabel = "crownlabs.polito.it/instance-submission-completed"
	// ProvisionJobLabel -> Key of the label added by the Provision Job to flag the PVC after it completed.
	ProvisionJobLabel = "crownlabs.polito.it/volume-provisioning"

	labelManagedByInstanceValue = "instance"
	labelManagedByTenantValue   = "tenant"
	labelTypeSandboxValue       = "sandbox"

	// ProvisionJobValueOk -> Value of the label added by the Provision Job to flag the PVC when everything worked fine.
	ProvisionJobValueOk = "completed"
	// ProvisionJobValuePending -> Value of the label added by the Provision Job to flag the PVC when it hasn't completed yet.
	ProvisionJobValuePending = "pending"

	// VolumeTypeValueShVol -> Value of the label for PVC which has been created by a Shared Volume.
	VolumeTypeValueShVol = "sharedvolume"
)

// InstanceLabels receives in input a set of labels and returns the updated set depending on the specified template,
// along with a boolean value indicating whether an update should be performed.
func InstanceLabels(labels map[string]string, template *clv1alpha2.Template, instance *clv1alpha2.Instance) (map[string]string, bool) {
	labels = deepCopyLabels(labels)
	update := false

	update = updateLabel(labels, labelManagedByKey, labelManagedByInstanceValue) || update
	update = updateLabel(labels, labelWorkspaceKey, template.Spec.WorkspaceRef.Name) || update
	update = updateLabel(labels, labelTemplateKey, template.Name) || update
	update = updateLabel(labels, labelPersistentKey, persistentLabelValue(template.Spec.EnvironmentList)) || update
	update = updateLabel(labels, labelNodeSelectorKey, nodeSelectorLabelValue(template.Spec.EnvironmentList, instance)) || update

	if instance != nil {
		instCustomizationUrls := instance.Spec.CustomizationUrls

		if instCustomizationUrls != nil && instCustomizationUrls.StatusCheck != "" && labels[InstanceTerminationSelectorLabel] == "" {
			update = updateLabel(labels, InstanceTerminationSelectorLabel, strconv.FormatBool(true))
		}
	}

	return labels, update
}

// InstanceObjectLabels receives in input a set of labels and returns the updated set depending on the specified instance.
func InstanceObjectLabels(labels map[string]string, instance *clv1alpha2.Instance) map[string]string {
	labels = deepCopyLabels(labels)

	labels[labelManagedByKey] = labelManagedByInstanceValue
	labels[labelInstanceKey] = instance.Name
	labels[labelTemplateKey] = instance.Spec.Template.Name
	labels[labelTenantKey] = instance.Spec.Tenant.Name

	return labels
}

// SandboxObjectLabels receives in input a set of labels and the tenant name, returns the updated set.
func SandboxObjectLabels(labels map[string]string, name string) map[string]string {
	labels = deepCopyLabels(labels)

	labels[labelManagedByKey] = labelManagedByTenantValue
	labels[labelTypeKey] = labelTypeSandboxValue
	labels[labelTenantKey] = name

	return labels
}

// InstanceSelectorLabels returns a set of selector labels depending on the specified instance.
func InstanceSelectorLabels(instance *clv1alpha2.Instance) map[string]string {
	return map[string]string{
		labelInstanceKey: instance.Name,
		labelTemplateKey: instance.Spec.Template.Name,
		labelTenantKey:   instance.Spec.Tenant.Name,
	}
}

// InstanceAutomationLabelsOnTermination returns a set of labels to be set on an instance when it is terminated.
func InstanceAutomationLabelsOnTermination(labels map[string]string, submissionRequired bool) map[string]string {
	labels = deepCopyLabels(labels)
	labels[InstanceTerminationSelectorLabel] = strconv.FormatBool(false)
	labels[InstanceSubmissionSelectorLabel] = strconv.FormatBool(submissionRequired)
	return labels
}

// InstanceAutomationLabelsOnSubmission returns a set of labels to be set on an instance when it is submitted.
func InstanceAutomationLabelsOnSubmission(labels map[string]string, submissionSucceded bool) map[string]string {
	labels = deepCopyLabels(labels)
	labels[InstanceSubmissionSelectorLabel] = strconv.FormatBool(false)
	labels[InstanceSubmissionCompletedLabel] = strconv.FormatBool(submissionSucceded)
	return labels
}

// MonitorableServiceLabels returns adds a label for a service so that it is monitored by the Instance ServiceMonitor.
func MonitorableServiceLabels(labels map[string]string) map[string]string {
	labels = deepCopyLabels(labels)
	labels[labelMetricsEnabled] = strconv.FormatBool(true)
	return labels
}

// InstanceComponentLabels returns a set of labels to be set on an instance component when it is created.
func InstanceComponentLabels(instance *clv1alpha2.Instance, componentName string) map[string]string {
	return InstanceObjectLabels(map[string]string{
		labelComponentKey: componentName,
	}, instance)
}

// SharedVolumeLabels receives in input a set of labels and returns the updated set.
func SharedVolumeLabels(labels map[string]string) (map[string]string, bool) {
	labels = deepCopyLabels(labels)
	update := false

	update = updateLabel(labels, labelManagedByKey, labelManagedByInstanceValue) || update

	return labels, update
}

// SharedVolumeObjectLabels receives in input a set of labels and returns the updated set depending on the specified shared volume.
func SharedVolumeObjectLabels(labels map[string]string) map[string]string {
	labels = deepCopyLabels(labels)

	labels[labelManagedByKey] = labelManagedByInstanceValue
	labels[labelVolumeTypeKey] = VolumeTypeValueShVol

	return labels
}

// deepCopyLabels creates a copy of the labels map.
func deepCopyLabels(input map[string]string) map[string]string {
	output := map[string]string{}
	for key, value := range input {
		output[key] = value
	}
	return output
}

// updateLabel configures a map entry to a given value, and returns whether a change was performed.
func updateLabel(labels map[string]string, key, value string) bool {
	if labels[key] != value {
		labels[key] = value
		return true
	}
	return false
}

// persistentLabelValue returns the value to be assigned to the persistent label, depending on the environment list.
func persistentLabelValue(environmentList []clv1alpha2.Environment) string {
	for i := range environmentList {
		if environmentList[i].Persistent {
			return strconv.FormatBool(true)
		}
	}
	return strconv.FormatBool(false)
}

// nodeSelectorLabelValue returns the value to be assigned to the node selector label, depending on the presence or the absence of the field.
func nodeSelectorLabelValue(environmentList []clv1alpha2.Environment, instance *clv1alpha2.Instance) string {
	if instance != nil {
		for i := range environmentList {
			nodeSel := NodeSelectorLabels(instance, &environmentList[i])
			if len(nodeSel) > 0 {
				return strconv.FormatBool(true)
			}
		}
	}
	return strconv.FormatBool(false)
}

// InstanceNameFromLabels receives in input a set of labels and returns the instance name, if any.
func InstanceNameFromLabels(labels map[string]string) (string, bool) {
	// if labelInstanceKey is present instance will receive the associated value and found will be set to true.
	instance, found := labels[labelInstanceKey]
	return instance, found
}
