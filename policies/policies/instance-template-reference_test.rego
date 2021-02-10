package crownlabs_instance_template_reference

test_generic_namespace_not_exists {
	input := {"review": input_review_with_namespace("workspace-not-existing", "whatever")}
	results := violation with input as input with data.inventory as data_inventory
	count(results) > 0
}

test_generic_namespace_empty {
	input1 := {"review": input_review_with_namespace("workspace-empty-1", "whatever")}
	input2 := {"review": input_review_with_namespace("workspace-empty-2", "whatever")}
	input3 := {"review": input_review_with_namespace("workspace-empty-3", "whatever")}
	input4 := {"review": input_review_with_namespace("workspace-empty-4", "whatever")}

	results1 := violation with input as input1 with data.inventory as data_inventory
	results2 := violation with input as input2 with data.inventory as data_inventory
	results3 := violation with input as input3 with data.inventory as data_inventory
	results4 := violation with input as input4 with data.inventory as data_inventory

	count(results1) > 0
	count(results2) > 0
	count(results3) > 0
	count(results4) > 0
}

test_generic_namespace_template_not_exists {
	# This template exists but in a different namespace
	input := {"review": input_review_with_namespace("workspace-coffee", "green-tea")}
	results := violation with input as input with data.inventory as data_inventory
	count(results) > 0
}

test_generic_namespace_template_exists_single {
	input := {"review": input_review_with_namespace("workspace-coffee", "dark-coffee")}
	results := violation with input as input with data.inventory as data_inventory
	count(results) == 0
}

test_generic_namespace_template_exists_multiple {
	input1 := {"review": input_review_with_namespace("workspace-tea", "green-tea")}
	input2 := {"review": input_review_with_namespace("workspace-tea", "white-tea")}

	results1 := violation with input as input1 with data.inventory as data_inventory
	results2 := violation with input as input2 with data.inventory as data_inventory

	count(results1) == 0
	count(results2) == 0
}

test_default_namespace_empty {
	input := {"review": input_review_without_namespace("whatever")}
	results := violation with input as input with data.inventory as data_inventory_empty_default_namespace
	count(results) > 0
}

test_default_namespace_template_not_exists {
	input := {"review": input_review_without_namespace("not-existing")}
	results := violation with input as input with data.inventory as data_inventory
	count(results) > 0
}

test_default_namespace_template_exists {
	input1 := {"review": input_review_without_namespace("chilly-pepper")}
	input2 := {"review": input_review_without_namespace("just-pepper")}

	results1 := violation with input as input1 with data.inventory as data_inventory
	results2 := violation with input as input2 with data.inventory as data_inventory

	count(results1) == 0
	count(results2) == 0
}

input_review_with_namespace(template_namespace, template_name) = output {
	output = {"object": {
		"metadata": {
			"name": "instance-name",
			"namespace": "instance-namespace",
		},
		"spec": {
			"template.crownlabs.polito.it/TemplateRef": {
				"name": template_name,
				"namespace": template_namespace,
			},
			"tenant.crownlabs.polito.it/TenantRef": {"name": "tenant-name"},
		},
	}}
}

input_review_without_namespace(template_name) = output {
	output = {"object": {
		"metadata": {
			"name": "instance-name",
			"namespace": "instance-namespace",
		},
		"spec": {
			"template.crownlabs.polito.it/TemplateRef": {"name": template_name},
			"tenant.crownlabs.polito.it/TenantRef": {"name": "tenant-name"},
		},
	}}
}

data_inventory = {"namespace": {
	"workspace-tea": {"crownlabs.polito.it/v1alpha2": {"Template": {
		"green-tea": {
			"apiVersion": "crownlabs.polito.it/v1alpha2",
			"kind": "Template",
			"metadata": {
				"name": "green-tea",
				"namespace": "workspace-tea",
			},
		},
		"white-tea": {
			"apiVersion": "crownlabs.polito.it/v1alpha2",
			"kind": "Template",
			"metadata": {
				"name": "white-tea",
				"namespace": "workspace-tea",
			},
		},
	}}},
	"workspace-coffee": {
		"crownlabs.polito.it/v1alpha2": {"Template": {"dark-coffee": {
			"apiVersion": "crownlabs.polito.it/v1alpha2",
			"kind": "Template",
			"metadata": {
				"name": "dark-coffee",
				"namespace": "workspace-coffee",
			},
		}}},
		"foo.bar.com/v1alpha1": {"Baz": {}},
	},
	"workspace-empty-1": {},
	"workspace-empty-2": {"foo.bar.com/v1alpha1": {"Baz": {}}},
	"workspace-empty-3": {"crownlabs.polito.it/v1alpha2": {"Baz": {}}},
	"workspace-empty-4": {"crownlabs.polito.it/v1alpha2": {"Template": {}}},
	"default": {"crownlabs.polito.it/v1alpha2": {"Template": {
		"chilly-pepper": {
			"apiVersion": "crownlabs.polito.it/v1alpha2",
			"kind": "Template",
			"metadata": {
				"name": "chilly-pepper",
				"namespace": "default",
			},
		},
		"just-pepper": {
			"apiVersion": "crownlabs.polito.it/v1alpha2",
			"kind": "Template",
			"metadata": {
				"name": "just-pepper",
				"namespace": "default",
			},
		},
	}}},
}}

data_inventory_empty_default_namespace = {"namespace": {"default": {"crownlabs.polito.it/v1alpha2": {"Template": {}}}}}
