# OPA Policies

## Available policies

This section details and documents the different policies available in CrownLabs:
* **Verify Tenant Patch**: verifies that a tenant creation or patch is allowed
* **Verify Instance-Template Reference**: this policy verifies that an instance refers to an existing template in the correct namespace when it is created or updated.

### Verify Tenant Patch

This policy verifies that a tenant creation or patch is allowed. In particular :

- for creation
  - is allowed only if the creator is manager in all workspaces for which he creates the new tenant resource.
  - is never allowed for a user.
  - the creator must already have his own tenant in the cluster.
  - is not allowed to create a tenant that already exists

- for update
  - a user can modify any tenant workspaces for which he is manager
  - a user can always update the publicKeys on his own tenant
  - is not allowed to modify any other field

- for a cluster-admin or an operator all operations are allowed

## How to deploy

The creation of the Gatekeeper resources and their deployment is automated through an Helm Chart.

This allows the creators of the policies to focus on writing the appropriate constraints in the `rego` language (together with the corresponding tests), without dealing with the Gatekeeper manifests.
Additionally, it also prevents the duplication of the policies themselves in two files (i.e. the `rego` policy, which is used to run the tests, and the Gatekeeper ConstraintTemplate), avoiding possible inconsistencies between the two.

In order to add a new policy to the Helm chart, it is sufficient to an add an entry to the `policies` array of the [values.yaml](values.yaml) file. Specifically, each entry contains the following fields:

* **name**: the name assigned to the policy resources (in camel-case);
* **file**: the path of the `rego` file implementing the policy;
* **dryRun**: whether the policy is enforced or violations are only logged;
* **resources**: the list of resources the policy applies to;
* **sync**: the list of resources that need to be accessed from the policy (optional).
