
package crownlabs_instance_template_reference

t_ns := input.review.object.spec["template.crownlabs.polito.it/TemplateRef"].namespace

ns = t_ns {
	count([t_ns]) > 0
	
}

else = "default" {
	true
}

violation[{"msg": msg, "details": {}}] {
	not data.inventory.namespace[ns]
	msg := sprintf("Namespace %v doesn't exist", [data.inventory])
}

violation[{"msg": msg, "details": {}}] {
	var := data.inventory.namespace[ns]["crownlabs.polito.it/v1alpha2"].Template
	var == {}
	msg := sprintf("Namespace %v doesn't contain any template", [ns])
}

violation[{"msg": msg, "details": {"missing_template": [missing]}}] {
	provided := {input.review.object.spec["template.crownlabs.polito.it/TemplateRef"].name}
	required := {data.inventory.namespace[ns]["crownlabs.polito.it/v1alpha2"].Template[_].metadata.name}
	missing := provided - required
	count(missing) > 0
	msg := sprintf("wrong template %v", [missing])
}

