VERIFY INSTANCES TEMPLATES REFERENCE
=====================================
## GOALS 
This policy verifies that an instance  refers to an existing template in the correct namespace when it is created or updated.

## TESTS
Tests are available in folder [policies](./policies).

## HOW TO DEPLOY
Run the following commands inside folder [manifest](./manifest):
- kubectl create -f config-sync.yaml
- kubectl create -f template.yaml
- kubectl create -f constraint.yaml

**Severity:** Violation

**Resources:** crownlabs.polito.it/Instance[Instance](../../operators/deploy/crds/crownlabs.polito.it_instances.yaml)