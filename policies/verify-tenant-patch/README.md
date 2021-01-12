# VERIFY TENANT PATCH

## GOAL

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

## TESTS

Tests are available in folder [policies](./policies).

## HOW TO DEPLOY

Run the following commands inside folder [manifest](./manifest):

- kubectl create -f config_sync.yaml
- kubectl create -f template.yaml
- kubectl create -f constraint.yaml

**Severity:** Violation 

**Resources:**  [Tenant](../../operators/deploy/crds/crownlabs.polito.it_tenants.yaml)