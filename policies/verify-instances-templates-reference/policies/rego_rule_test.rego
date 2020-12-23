package crownlabs_instance_template_reference

test_namespace_notexist {
	namet := "temp2"
	nst := "notexist"
	ns1 := "test-space1"
	ns2 := "test-space1"
	ns3 := "test-space2"
	input := {"review": input_review(namet, nst)}
	my_data := data_inventory(ns1, ns2, ns3)
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

test_namespace_empty {
	namet := "temp2"
	nst := "test-space3"
	ns1 := "test-space1"
	ns2 := "test-space3"
	ns3 := "test-space2"
	input := {"review": input_review(namet, nst)}
	my_data := data_inventory(ns1, ns2, ns3)
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

test_nametemplate_notexist {
	namet := "notexist"
	nst := "test-space2"
	ns1 := "test-space1"
	ns2 := "test-space1"
	ns3 := "test-space2"
	input := {"review": input_review(namet, nst)}
	my_data := data_inventory(ns1, ns2, ns3)
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

test_namespace_exist {
	namet := "temp1"
	nst := "test-space1"
	ns1 := "test-space1"
	ns2 := "test-space3"
	ns3 := "test-space2"
	input := {"review": input_review(namet, nst)}
	my_data := data_inventory(ns1, ns2, ns3)
	result := violation with input as input with data.inventory as my_data
	count(result) == 0
}

test_namespace_default_fail {
	namet := "temp2"
	ns1 := "default"
	ns2 := "test-space1"
	ns3 := "test-space2"
	input := {"review": input_review1(namet)}
	my_data := data_inventory(ns1, ns2, ns3)
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

test_namespace_default {
	namet := "temp1"
	ns1 := "test-space3"
	ns2 := "test-space1"
	ns3 := "test-space2"
	input := {"review": input_review1(namet)}
	my_data := data_inventory(ns1, ns2, ns3)
	result := violation with input as input with data.inventory as my_data
	count(result) == 0
}

input_review(namet, nst) = output {
	output = {"object": {
		"metadata": {
			"name": "test-name",
			"namespace": "test-namespace",
		},
		"spec": {
			"template.crownlabs.polito.it/TemplateRef": {
				"name": namet,
				"namespace": nst,
			},
			"tenant.crownlabs.polito.it/TenantRef": {"name": "test-tenant-name"},
		},
	}}
}

input_review1(namet) = output {
	output = {"object": {
		"metadata": {
			"name": "test-name",
			"namespace": "test-namespace",
		},
		"spec": {
			"template.crownlabs.polito.it/TemplateRef": {"name": namet},
			"tenant.crownlabs.polito.it/TenantRef": {"name": "test-tenant-name"},
		},
	}}
}

data_inventory(ns1, ns2, ns3) = output {
	output = {"namespace": {
		ns1: {"crownlabs.polito.it/v1alpha2": {"Template": {"temp1": {
			"apiVersion": "crownlabs.polito.it/v1alpha2",
			"kind": "Template",
			"metadata": {
				"name": "temp1",
				"namespace": ns1,
			},
		}}}},
		ns3: {"crownlabs.polito.it/v1alpha2": {"Template": {"temp2": {
			"apiVersion": "crownlabs.polito.it/v1alpha2",
			"kind": "Template",
			"metadata": {
				"name": "temp2",
				"namespace": ns3,
			},
		}}}},
		ns2: {"crownlabs.polito.it/v1alpha2": {"Template": {}}},
		"default": {"crownlabs.polito.it/v1alpha2": {"Template": {"temp1": {
			"apiVersion": "crownlabs.polito.it/v1alpha2",
			"kind": "Template",
			"metadata": {
				"name": "temp1",
				"namespace": "default",
			},
		}}}},
	}}
}