package forge

import clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"

const (
	labelManagedByKey = "crownlabs.polito.it/managed-by"
	labelWorkspaceKey = "crownlabs.polito.it/workspace"
	labelTemplateKey  = "crownlabs.polito.it/template"

	labelManagedByValue = "instance"
)

// InstanceLabels receives in input a set of labels and returns the updated set depending on the specified template,
// along with a boolean value indicating whether an update should be performed.
func InstanceLabels(labels map[string]string, template *clv1alpha2.Template) (map[string]string, bool) {
	update := false
	if labels == nil {
		labels = map[string]string{}
		update = true
	}

	update = updateLabel(labels, labelManagedByKey, labelManagedByValue) || update
	update = updateLabel(labels, labelWorkspaceKey, template.Spec.WorkspaceRef.Name) || update
	update = updateLabel(labels, labelTemplateKey, template.Name) || update

	return labels, update
}

// updateLabel configures a map entry to a given value, and returns whether a change was performed.
func updateLabel(labels map[string]string, key, value string) bool {
	if labels[key] != value {
		labels[key] = value
		return true
	}
	return false
}
