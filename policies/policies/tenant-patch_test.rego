package crownlabs_tenant_patch

#creator does not have any tenant
test_no_tenant_present {
	input := {"review": input_review_invalid_creation_creator_has_not_tenant}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

#creator cannot create a tenant for himself
test_creator_already_has_a_tenant {
	input := {"review": input_review_invalid_creation_creator_has_a_tenant}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

#creator cannot create a tenant in workspace he is not enrolled in
test_creator_not_enrolled {
	input := {"review": input_review_invalid_creation_creator_not_enrolled}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

#creator cannot create a tenant where he is not manager
test_creator_not_manager {
	input := {"review": input_review_invalid_creation_creator_not_manager}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

#correct creation of new tenant
test_correct_creation_of_tenant {
	input := {"review": input_review_valid_creation_tenant}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) == 0
}

#try to change the email
test_invalid_patch_on_wrong_fields {
	input := {"review": input_review_invalid_patch_field}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

#update a workspace in which the creator is not enrolled
test_invalid_patch_for_notenrolled_workspace {
	input := {"review": input_review_invalid_patch_creator_not_enrolled}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

#update a workspace in which the creator is not manager
test_invalid_patch_for_not_manager_creator {
	input := {"review": input_review_invalid_patch_creator_not_manager}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

#update a workspace in which the creator is the manager
test_valid_patch_for_manager_creator {
	input := {"review": input_review_valid_patch_creator_is_manager}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) == 0
}

#update public key not on your own tenant
test_invalid_patch_publicKey {
	input := {"review": input_review_invalid_patch_publicKey}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) > 0
}

#update public key on your own tenant
test_valid_patch_publicKey {
	input := {"review": input_review_valid_patch_publicKey}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) == 0
}

#create tenant with cluster-admin priviledges
test_valid_creation_cluster_admin {
	input := {"review": input_review_valid_creation_cluster_admin}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) == 0
}

#update tenant with serviceaccounts priviledges
test_valid_creation_serviceaccounts {
	input := {"review": input_review_valid_patch_serviceaccounts}
	my_data := data_review1
	result := violation with input as input with data.inventory as my_data
	count(result) == 0
}


data_review1 = output {
	output = {"cluster": {"crownlabs.polito.it/v1alpha1": {"Tenant": {
		"mariocalza": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"name": "mariocalza",
				"resourceVersion": "2580284",
				"selfLink": "/apis/crownlabs.polito.it/v1alpha1/tenants/mariocalza",
				"uid": "f9fe46e2-e8bc-4442-acff-30f6bb812bf2",
				"creationTimestamp": "2020-12-25T15:22:44Z",
				"generation": 1,
			},
			"spec": {
				"workspaces": [
					{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}},
					{"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11},
				],
				"email": "mario.calza@email.com",
				"firstName": "mario",
				"lastName": "calza",
				"publicKeys": ["key1", "key2", "key3"],
			},
		},
		"mariorossi": {
			"metadata": {
				"creationTimestamp": "2020-12-25T14:57:30Z",
				"generation": 1,
				"name": "mariorossi",
				"resourceVersion": "2576050",
				"selfLink": "/apis/crownlabs.polito.it/v1alpha1/tenants/mariorossi",
				"uid": "ecc18c03-c3e9-4a95-9339-fdd7bec3df4d",
			},
			"spec": {
				"email": "mario.rossi@email.com",
				"firstName": "mario",
				"lastName": "rossi",
				"publicKeys": ["key1", "key2", "key3"],
				"workspaces": [
					{"groupNumber": 12, "role": "manager", "workspaceRef": {"name": "landc"}},
					{"groupNumber": 11, "role": "user", "workspaceRef": {"name": "landa"}},
				],
			},
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
		},
	}}}}
}

input_review_invalid_patch_field = output {
	output = {
		"requestKind": {"kind": "Tenant", "group": "crownlabs.polito.it", "version": "v1alpha1"},
		"operation": "UPDATE",
		"userInfo": {"username": "mariorossi", "groups": ["system:masters", "system:authenticated"]},
		"oldObject": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
			},
		},
		"options": {"kind": "UpdateOptions", "apiVersion": "meta.k8s.io/v1", "fieldManager": "kubectl-patch"},
		"uid": "3771271d-9194-4924-b3b6-edccf84d8f76",
		"requestResource": {"group": "crownlabs.polito.it", "version": "v1alpha1", "resource": "tenants"},
		"name": "example.com",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario11.calza@email.it",
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
			},
		},
		"dryRun": false,
		"kind": {
			"group": "crownlabs.polito.it",
			"version": "v1alpha1",
			"kind": "Tenant",
		},
	}
}

input_review_invalid_patch_creator_not_enrolled = output {
	output = {
		"requestKind": {"kind": "Tenant", "group": "crownlabs.polito.it", "version": "v1alpha1"},
		"operation": "UPDATE",
		"userInfo": {"username": "mariorossi", "groups": ["system:masters", "system:authenticated"]},
		"oldObject": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
			},
		},
		"options": {"kind": "UpdateOptions", "apiVersion": "meta.k8s.io/v1", "fieldManager": "kubectl-patch"},
		"uid": "3771271d-9194-4924-b3b6-edccf84d8f76",
		"requestResource": {"group": "crownlabs.polito.it", "version": "v1alpha1", "resource": "tenants"},
		"name": "example.com",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "newName"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
			},
		},
		"dryRun": false,
		"kind": {
			"group": "crownlabs.polito.it",
			"version": "v1alpha1",
			"kind": "Tenant",
		},
	}
}

input_review_invalid_patch_creator_not_manager = output {
	output = {
		"requestKind": {"kind": "Tenant", "group": "crownlabs.polito.it", "version": "v1alpha1"},
		"operation": "UPDATE",
		"userInfo": {"username": "mariorossi", "groups": ["system:masters", "system:authenticated"]},
		"oldObject": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
			},
		},
		"options": {"kind": "UpdateOptions", "apiVersion": "meta.k8s.io/v1", "fieldManager": "kubectl-patch"},
		"uid": "3771271d-9194-4924-b3b6-edccf84d8f76",
		"requestResource": {"group": "crownlabs.polito.it", "version": "v1alpha1", "resource": "tenants"},
		"name": "example.com",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "manager", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
			},
		},
		"dryRun": false,
		"kind": {
			"group": "crownlabs.polito.it",
			"version": "v1alpha1",
			"kind": "Tenant",
		},
	}
}

input_review_valid_patch_creator_is_manager = output {
	output = {
		"requestKind": {"kind": "Tenant", "group": "crownlabs.polito.it", "version": "v1alpha1"},
		"operation": "UPDATE",
		"userInfo": {"username": "mariorossi", "groups": ["system:masters", "system:authenticated"]},
		"oldObject": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
			},
		},
		"options": {"kind": "UpdateOptions", "apiVersion": "meta.k8s.io/v1", "fieldManager": "kubectl-patch"},
		"uid": "3771271d-9194-4924-b3b6-edccf84d8f76",
		"requestResource": {"group": "crownlabs.polito.it", "version": "v1alpha1", "resource": "tenants"},
		"name": "example.com",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"workspaces": [{"groupNumber": 12, "role": "manager", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
			},
		},
		"dryRun": false,
		"kind": {
			"group": "crownlabs.polito.it",
			"version": "v1alpha1",
			"kind": "Tenant",
		},
	}
}

input_review_invalid_patch_publicKey = output {
	output = {
		"requestKind": {"kind": "Tenant", "group": "crownlabs.polito.it", "version": "v1alpha1"},
		"operation": "UPDATE",
		"userInfo": {"username": "mariorossi", "groups": ["system:masters", "system:authenticated"]},
		"oldObject": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
			},
		},
		"options": {"kind": "UpdateOptions", "apiVersion": "meta.k8s.io/v1", "fieldManager": "kubectl-patch"},
		"uid": "3771271d-9194-4924-b3b6-edccf84d8f76",
		"requestResource": {"group": "crownlabs.polito.it", "version": "v1alpha1", "resource": "tenants"},
		"name": "example.com",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
				"name": "mario",
				"publicKeys": ["test-key4 test-key2"],
				"surname": "calza",
			},
		},
		"dryRun": false,
		"kind": {
			"group": "crownlabs.polito.it",
			"version": "v1alpha1",
			"kind": "Tenant",
		},
	}
}

input_review_valid_patch_publicKey = output {
	output = {
		"requestKind": {"kind": "Tenant", "group": "crownlabs.polito.it", "version": "v1alpha1"},
		"operation": "UPDATE",
		"userInfo": {"username": "mariorossi", "groups": ["system:masters", "system:authenticated"]},
		"oldObject": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariorossi",
				"resourceVersion": "231625",
			},
			"spec": {
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "rossi",
				"workspaces": [{"groupNumber": 12, "role": "manager", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.rossi@email.it",
			},
		},
		"options": {"kind": "UpdateOptions", "apiVersion": "meta.k8s.io/v1", "fieldManager": "kubectl-patch"},
		"uid": "3771271d-9194-4924-b3b6-edccf84d8f76",
		"requestResource": {"group": "crownlabs.polito.it", "version": "v1alpha1", "resource": "tenants"},
		"name": "example.com",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariorossi",
				"resourceVersion": "231625",
			},
			"spec": {
				"workspaces": [{"groupNumber": 12, "role": "manager", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.rossi@email.it",
				"name": "mario",
				"publicKeys": ["test-key4 test-key2"],
				"surname": "rossi",
			},
		},
		"dryRun": false,
		"kind": {
			"group": "crownlabs.polito.it",
			"version": "v1alpha1",
			"kind": "Tenant",
		},
	}
}

input_review_invalid_creation_creator_has_not_tenant = output {
	output = {
		"requestKind": {"group": "crownlabs.polito.it", "version": "v1alpha1", "kind": "Tenant"}, "name": "test-any-name",
		"operation": "CREATE",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1", "kind": "Tenant",
			"metadata": {"creationTimestamp": "2020-12-25T16:16:14Z", "generation": 1, "name": "test-any-name", "uid": "c3b3a51a-1a43-47f2-a7f4-c84914b026ba"},
			"spec": {
				"email": "test.any-name@email.com", "firstName": "test-name", "lastName": "test-last", "publicKeys": ["key1", "key2", "key3"],
				"workspaces": [{"groupNumber": 12, "role": "manager", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
			},
		},
		"oldObject": null,
		"userInfo": {"username": "test-name-no-tenant", "groups": ["system:masters", "system:authenticated"]},
	}
}

input_review_invalid_creation_creator_has_a_tenant = output {
	output = {
		"requestKind": {"group": "crownlabs.polito.it", "version": "v1alpha1", "kind": "Tenant"}, "name": "mariocalza", "operation": "CREATE",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1", "kind": "Tenant",
			"metadata": {"creationTimestamp": "2020-12-25T16:16:14Z", "generation": 1, "name": "mariocalza", "uid": "c3b3a51a-1a43-47f2-a7f4-c84914b026ba"},
			"spec": {
				"email": "mario.calza@email.com", "firstName": "mario", "lastName": "calza", "publicKeys": ["key1", "key2", "key3"],
				"workspaces": [{"groupNumber": 12, "role": "manager", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
			},
		},
		"oldObject": null,
		"userInfo": {"username": "mariocalza", "groups": ["system:masters", "system:authenticated"]},
	}
}

input_review_invalid_creation_creator_not_enrolled = output {
	output = {
		"requestKind": {"group": "crownlabs.polito.it", "version": "v1alpha1", "kind": "Tenant"}, "name": "test-name", "operation": "CREATE",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1", "kind": "Tenant",
			"metadata": {"creationTimestamp": "2020-12-25T16:16:14Z", "generation": 1, "name": "test-name", "uid": "c3b3a51a-1a43-47f2-a7f4-c84914b026ba"},
			"spec": {
				"email": "test.name@email.com", "firstName": "test-name", "lastName": "test-last", "publicKeys": ["key1", "key2", "key3"],
				"workspaces": [{"groupNumber": 12, "role": "manager", "workspaceRef": {"name": "test-land"}}],
			},
		},
		"oldObject": null,
		"userInfo": {"username": "mariocalza", "groups": ["system:masters", "system:authenticated"]},
	}
}

input_review_invalid_creation_creator_not_manager = output {
	output = {
		"requestKind": {"group": "crownlabs.polito.it", "version": "v1alpha1", "kind": "Tenant"}, "name": "test-name", "operation": "CREATE",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1", "kind": "Tenant",
			"metadata": {"creationTimestamp": "2020-12-25T16:16:14Z", "generation": 1, "name": "test-name", "uid": "c3b3a51a-1a43-47f2-a7f4-c84914b026ba"},
			"spec": {
				"email": "test.name@email.com", "firstName": "test-name", "lastName": "test-last", "publicKeys": ["key1", "key2", "key3"],
				"workspaces": [{"groupNumber": 12, "role": "manager", "workspaceRef": {"name": "landc"}}],
			},
		},
		"oldObject": null,
		"userInfo": {"username": "mariocalza", "groups": ["system:masters", "system:authenticated"]},
	}
}

input_review_valid_creation_tenant = output {
	output = {
		"requestKind": {"group": "crownlabs.polito.it", "version": "v1alpha1", "kind": "Tenant"}, "name": "test-name", "operation": "CREATE",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1", "kind": "Tenant",
			"metadata": {"creationTimestamp": "2020-12-25T16:16:14Z", "generation": 1, "name": "test-name", "uid": "c3b3a51a-1a43-47f2-a7f4-c84914b026ba"},
			"spec": {
				"email": "test.name@email.com", "firstName": "test-name", "lastName": "test-last", "publicKeys": ["key1", "key2", "key3"],
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}],
			},
		},
		"oldObject": null,
		"userInfo": {"username": "mariorossi", "groups": ["system:masters", "system:authenticated"]},
	}
}


input_review_valid_creation_cluster_admin=output{
        output = {
		"requestKind": {"group": "crownlabs.polito.it", "version": "v1alpha1", "kind": "Tenant"}, "name": "test-name", "operation": "CREATE",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1", "kind": "Tenant",
			"metadata": {"creationTimestamp": "2020-12-25T16:16:14Z", "generation": 1, "name": "test-name", "uid": "c3b3a51a-1a43-47f2-a7f4-c84914b026ba"},
			"spec": {
				"email": "test.name@email.com", "firstName": "test-name", "lastName": "test-last", "publicKeys": ["key1", "key2", "key3"],
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landa"}}],
			},
		},
		"oldObject": null,
		"userInfo": {"username": "giovannibianchi", "groups": ["system:masters", "system:authenticated","kubernetes:admin"]},
	}
}



input_review_valid_patch_serviceaccounts = output {
	output = {
		"requestKind": {"kind": "Tenant", "group": "crownlabs.polito.it", "version": "v1alpha1"},
		"operation": "UPDATE",
		"userInfo": {"username": "giovannibianchi", "groups": ["system:masters", "system:authenticated","system:serviceaccounts"]},
		"oldObject": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario.calza@email.it",
			},
		},
		"options": {"kind": "UpdateOptions", "apiVersion": "meta.k8s.io/v1", "fieldManager": "kubectl-patch"},
		"uid": "3771271d-9194-4924-b3b6-edccf84d8f76",
		"requestResource": {"group": "crownlabs.polito.it", "version": "v1alpha1", "resource": "tenants"},
		"name": "example.com",
		"object": {
			"apiVersion": "crownlabs.polito.it/v1alpha1",
			"kind": "Tenant",
			"metadata": {
				"uid": "664a606d-0f45-4ca3-9ae5-629bc86495eb",
				"creationTimestamp": "2020-12-20T11:14:59Z",
				"generation": 1,
				"name": "mariocalza",
				"resourceVersion": "231625",
			},
			"spec": {
				"workspaces": [{"groupNumber": 12, "role": "user", "workspaceRef": {"name": "landc"}}, {"role": "user", "workspaceRef": {"name": "landa"}, "groupNumber": 11}],
				"createSandbox": false,
				"email": "mario11.calza@email.it",
				"name": "mario",
				"publicKeys": ["test-key1 test-key2"],
				"surname": "calza",
			},
		},
		"dryRun": false,
		"kind": {
			"group": "crownlabs.polito.it",
			"version": "v1alpha1",
			"kind": "Tenant",
		},
	}
}
