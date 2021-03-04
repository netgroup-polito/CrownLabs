package crownlabs_instance_template_reference

tmpl_api_version := "crownlabs.polito.it/v1alpha2"

tmpl_ref_field := "template.crownlabs.polito.it/TemplateRef"

# Return the referenced template namespace, in case the field is defined, otherwise "default"
get_tmpl_namespace(review) = namespace {
	review.object.spec[tmpl_ref_field].namespace
	namespace := review.object.spec[tmpl_ref_field].namespace
} else = "default" {
	true
}

# Return the referenced template name
get_tmpl_name(review) = name {
	name := review.object.spec[tmpl_ref_field].name
}

# This violation is triggered if the referenced template does not exist
violation[{"msg": msg, "details": details}] {
	tmpl_namespace := get_tmpl_namespace(input.review)
	tmpl_name := get_tmpl_name(input.review)

	not data.inventory.namespace[tmpl_namespace][tmpl_api_version].Template[tmpl_name]
	msg := sprintf("Template %v not found in namespace %v", [tmpl_name, tmpl_namespace])
	details := {"template_name": tmpl_name, "template_namespace": tmpl_namespace}
}
