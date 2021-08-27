// Copyright 2020-2021 Politecnico di Torino
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
	labelManagedByKey  = "crownlabs.polito.it/managed-by"
	labelInstanceKey   = "crownlabs.polito.it/instance"
	labelWorkspaceKey  = "crownlabs.polito.it/workspace"
	labelTemplateKey   = "crownlabs.polito.it/template"
	labelTenantKey     = "crownlabs.polito.it/tenant"
	labelPersistentKey = "crownlabs.polito.it/persistent"

	labelManagedByValue = "instance"
)

// InstanceLabels receives in input a set of labels and returns the updated set depending on the specified template,
// along with a boolean value indicating whether an update should be performed.
func InstanceLabels(labels map[string]string, template *clv1alpha2.Template) (map[string]string, bool) {
	labels = deepCopyLabels(labels)
	update := false

	update = updateLabel(labels, labelManagedByKey, labelManagedByValue) || update
	update = updateLabel(labels, labelWorkspaceKey, template.Spec.WorkspaceRef.Name) || update
	update = updateLabel(labels, labelTemplateKey, template.Name) || update
	update = updateLabel(labels, labelPersistentKey, persistentLabelValue(template.Spec.EnvironmentList)) || update

	return labels, update
}

// InstanceObjectLabels receives in input a set of labels and returns the updated set depending on the specified instance.
func InstanceObjectLabels(labels map[string]string, instance *clv1alpha2.Instance) map[string]string {
	labels = deepCopyLabels(labels)

	labels[labelManagedByKey] = labelManagedByValue
	labels[labelInstanceKey] = instance.Name
	labels[labelTemplateKey] = instance.Spec.Template.Name
	labels[labelTenantKey] = instance.Spec.Tenant.Name

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
