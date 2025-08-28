# Defining quotas per workspace and per user

## Defining quotas per workspace

Each workspace has the maximum resource quota defined in the `spec` of the CrownLabs workspace itself, as in the following example which refers to the `test` workspace:

    admin@k8s-master:~$ kubectl get workspace test -o yaml
    apiVersion: crownlabs.polito.it/v1alpha1
    kind: Workspace
    ...
    spec:
      prettyName: CrownLabs workspace dedicated to testing
      quota:
        cpu: "10"
        instances: 3
        memory: 64G

In the above example, users belonging to this workspace can launch up to 3 instances (either VMs or containers), but globally cannot consume more than 10CPUs and 64GB of memory.

Resource quotas can be changed by simply updating the `quota` section of the workspace specification.

## Defining quotas per user

While each CrownLabs workspace has its own resource limits, the values associated to each user depends on the sum of the resource quota of each workspace it belongs to.
For instance, if a user belongs to a first workspace whose CPU limit is 10, and a second workspace whose CPU limit is 8, the total CPU limit for the user is 18, no matter which workspace it is currently using.

In other words, with this (simple) resource quota algorithm, the above user can consume up to 18 CPU even when launching instances (VMs or containers) all within the same workspace, as the user quota has the precedence over the workspace quota.
