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

	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
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
	labelEnvironmentKey  = "crownlabs.polito.it/environment"

	// InstanceTerminationSelectorLabel -> label for Instances which have to be be checked for termination.
	InstanceTerminationSelectorLabel = "crownlabs.polito.it/watch-for-instance-termination"
	// InstanceSubmissionSelectorLabel -> label for Instances which have to be submitted.
	InstanceSubmissionSelectorLabel = "crownlabs.polito.it/instance-submission-requested"
	// InstanceSubmissionCompletedLabel -> label for Instances that have been submitted.
	InstanceSubmissionCompletedLabel = "crownlabs.polito.it/instance-submission-completed"
	// ProvisionJobLabel -> Key of the label added by the Provision Job to flag the PVC after it completed.
	ProvisionJobLabel = "crownlabs.polito.it/volume-provisioning"
	// InstanceInactivityIgnoreNamespace -> label added to the Namespace to ignore inactivity termination for Instances in it.
	InstanceInactivityIgnoreNamespace = "crownlabs.polito.it/instance-inactivity-ignore"
	// ExpirationIgnoreNamespace -> label added to the Namespace to ignore expiration termination for Instances in it.
	ExpirationIgnoreNamespace = "crownlabs.polito.it/expiration-ignore"

	// EnvironmentNameLabel -> Key of the label used to store the environment name.
	EnvironmentNameLabel = "crownlabs.polito.it/environment-name"

	labelManagedByInstanceValue  = "instance"
	labelManagedByTenantValue    = "tenant"
	labelManagedByWorkspaceValue = "workspace"
	labelTypeWorkspaceValue      = "workspace"
	labelTypeSandboxValue        = "sandbox"

	labelAllowInstanceAccessKey   = "crownlabs.polito.it/allow-instance-access"
	labelAllowInstanceAccessValue = "true"

	// ProvisionJobValueOk -> Value of the label added by the Provision Job to flag the PVC when everything worked fine.
	ProvisionJobValueOk = "completed"
	// ProvisionJobValuePending -> Value of the label added by the Provision Job to flag the PVC when it hasn't completed yet.
	ProvisionJobValuePending = "pending"

	// VolumeTypeValueShVol -> Value of the label for PVC which has been created by a Shared Volume.
	VolumeTypeValueShVol = "sharedvolume"

	labelFirstNameKey = "crownlabs.polito.it/first-name"
	labelLastNameKey  = "crownlabs.polito.it/last-name"

	// AlertAnnotationNum -> the number of mail sent to the tenant to inform that the instance will be stopped/removed.
	AlertAnnotationNum = "crownlabs.polito.it/number-alerts-sent"

	// LastActivityAnnotation -> timestamp of the last access detected to the instance.
	LastActivityAnnotation = "crownlabs.polito.it/last-activity"

	// LastNotificationTimestampAnnotation -> timestamp of the last notification sent to the tenant.
	LastNotificationTimestampAnnotation = "crownlabs.polito.it/last-notification-timestamp"

	// LastRunningAnnotation ->  previous value of the `Running` field of the Instance.
	LastRunningAnnotation = "crownlabs.polito.it/last-running"

	// NoWorkspacesLabelKey -> label to be set when no workspaces are associated to the tenant.
	NoWorkspacesLabelKey = "crownlabs.polito.it/no-workspaces"
	// NoWorkspacesLabelValue -> value of the label to be set when no workspaces are associated to the tenant.
	NoWorkspacesLabelValue = "true"

	// CustomNumberOfAlertsAnnotation -> annotation to mark an Instance as having a custom number of alerts.
	CustomNumberOfAlertsAnnotation = "crownlabs.polito.it/custom-number-alerts"

	// ExpiringWarningNotificationAnnotation -> annotation to mark an Instance as having sent the expiring warning notification.
	ExpiringWarningNotificationAnnotation = "crownlabs.polito.it/sent-expiring-warning-notification"
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
	update = updateLabel(labels, labelNodeSelectorKey, nodeSelectorLabelValue(instance, template)) || update

	if instance != nil {
		if instance.Spec.StatusCheckURL != "" && labels[InstanceTerminationSelectorLabel] == "" {
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

// EnvironmentObjectLabels receives in input a set of labels and returns the updated set depending on the specified environment.
func EnvironmentObjectLabels(labels map[string]string, instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) map[string]string {
	labels = deepCopyLabels(labels)

	labels[labelManagedByKey] = labelManagedByInstanceValue
	labels[labelInstanceKey] = instance.Name
	labels[labelEnvironmentKey] = environment.Name
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

// EnvironmentSelectorLabels returns a set of selector labels depending on the specified environment.
func EnvironmentSelectorLabels(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) map[string]string {
	return map[string]string{
		labelInstanceKey:    instance.Name,
		labelEnvironmentKey: environment.Name,
		labelTemplateKey:    instance.Spec.Template.Name,
		labelTenantKey:      instance.Spec.Tenant.Name,
	}
}

// InstanceAutomationLabelsOnTermination returns a set of labels to be set on an instance when it is terminated.
func InstanceAutomationLabelsOnTermination(labels map[string]string, envName string, submissionRequired bool) map[string]string {
	labels = deepCopyLabels(labels)
	labels[EnvironmentNameLabel] = envName
	labels[InstanceTerminationSelectorLabel] = strconv.FormatBool(false)
	labels[InstanceSubmissionSelectorLabel] = strconv.FormatBool(submissionRequired)
	return labels
}

// InstanceAutomationLabelsOnSubmission returns a set of labels to be set on an instance when it is submitted.
func InstanceAutomationLabelsOnSubmission(labels map[string]string, envName string, submissionSucceded bool) map[string]string {
	labels = deepCopyLabels(labels)
	labels[EnvironmentNameLabel] = envName
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

// TenantLabels receives in input a set of labels and returns the updated set depending on the specified tenant.
func TenantLabels(labels map[string]string, tenant *clv1alpha2.Tenant, targetLabel common.KVLabel) map[string]string {
	labels = deepCopyLabels(labels)

	labels = UpdateTenantResourceCommonLabels(labels, targetLabel)
	labels[labelFirstNameKey] = CleanTenantName(tenant.Spec.FirstName)
	labels[labelLastNameKey] = CleanTenantName(tenant.Spec.LastName)

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
func nodeSelectorLabelValue(instance *clv1alpha2.Instance, template *clv1alpha2.Template) string {
	if instance != nil && template != nil {
		nodeSel := NodeSelectorLabels(instance, template)
		if len(nodeSel) > 0 {
			return strconv.FormatBool(true)
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

// UpdateWorkspaceResourceCommonLabels updates the common labels for resources managed by the workspace controller.
func UpdateWorkspaceResourceCommonLabels(labels map[string]string, targetLabel common.KVLabel) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 1)
	}
	labels[targetLabel.GetKey()] = targetLabel.GetValue()
	labels[labelManagedByKey] = labelManagedByWorkspaceValue

	return labels
}

// TemplateLabelSelector returns a label selector for a specific template name.
func TemplateLabelSelector(templateName string) client.MatchingLabels {
	return client.MatchingLabels{
		labelTemplateKey: templateName,
	}
}
