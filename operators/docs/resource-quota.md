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
        cpu: 10
        instances: 3
        memory: 64Gi
        disk: 15Gi
        otherResources:
          nvidia.com/gpu: "1"
          amd.com/gpu: "1"

In the above example, users belonging to this workspace can launch up to 3 instances (either VMs or containers), but globally cannot consume more than 10CPUs, 64Gi of memory, 15Gi of disk, 1 NVIDIA GPU and 1 AMD GPU.

Resource quotas can be changed by simply updating the `quota` section of the workspace specification.

Types of the resources defined above are the following:

- CPU: int64
- Instances: int64
- Memory: resource.Quantity
- Disk: resource.Quantity
- otherResources: map[string]resource.Quantity

## Defining quotas per user

While each CrownLabs workspace has its own resource limits, the values associated to each user depends on the sum of the resource quota of each workspace it belongs to.
For instance, if a user belongs to a first workspace whose CPU limit is 10, and a second workspace whose CPU limit is 8, the total CPU limit for the user is 18, no matter which workspace it is currently using.

In other words, with this (simple) resource quota algorithm, the above user can consume up to 18 CPU even when launching instances (VMs or containers) all within the same workspace, as the user quota has the precedence over the workspace quota.

NOTE: Regarding CPU limits it must be taken into account that CPU Cores are essentially summed among different workspaces, but objects like Templates (which belong to a specific Workspace) own an extra property which is ReservedCPUPercentage that limits the CPU Usage of the instances belonging to a specific template with this property defined.

Beyond the accumulated limits inherited from course workspaces, each user is allocated a Personal Workspace. This quota acts as a standalone baseline for private workloads and may be configured or expanded by modifying the `personalWorkspace` field directly within the `Tenant`'s Custom Resource.

Moreover, otherResources property has a particular management with respect to the other kinds of resources: it is a map which relates a key to a specific value. This is done in order to make CrownLabs future-proof so that if a particular custom resource has to be added to the cluster then all is necessary is to add the specific custom resource to the map otherResources. During the addition of this property the reason that made CrownLabs developers to design such a kind of resource was based on the opportunity to provide both NVIDIA and AMD GPUs in the near future.

## Enforcement Mechanism: ResourceQuota vs Validation Webhook

Understanding how CrownLabs enforces these limits under the hood is critical for cluster administrators.

1. The Native Kubernetes ResourceQuota (Global Barrier):
   When a Tenant is reconciled, the Tenant Operator automatically creates a standard Kubernetes `ResourceQuota` object inside the user's personal namespace (`tenant-<username>`). However, because all instances created by a single tenant are placed inside this same namespace regardless of the CrownLabs workspace they belong to, this native quota represents only a part of the global security barrier.

2. The Validation Webhook (Fine-Grained Check):
   To enforce workspace-specific logic and handle the sum of multiple workspaces correctly, CrownLabs leverages a custom Validation Webhook. This webhook intercepts the API traffic whenever an Instance is:
   * Created from scratch.
   * Resumed or unpaused (which updates the active resource consumption).

The webhook calculates the current real-time consumption of the user, compares it against the allowed sum of their workspaces, and explicitly denies the operation if the request would lead to a quota violation.

## Runtime Behavior and Quota Violations

When a user attempts to launch or resume an environment but the required resources exceed the allowed quota, CrownLabs prevents the cluster from entering a broken state:

1. Environment Phase: The Instance Controller intercepts the quota exhaustion error during the environment enforcement phase and changes the specific environment status to `ResourceQuotaExceeded`.

2. Automatic Recovery (Requeuing): Instead of failing permanently, the controller emits a warning event and automatically schedules a new reconciliation check.

Once the tenant terminates other running instances (freeing up resources) or an administrator increases the workspace quota, the next periodic check will automatically succeed, and the environment will transition smoothly into the `Starting` and `Ready` phases.

NOTE: In case of a Disk failure, the reconciliaton check is hard-coded in instctrl/controller.go and a new attempt to recover Instance execution happens after 1 minute (cyclically)