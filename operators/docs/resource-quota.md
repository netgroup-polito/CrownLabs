# Defining Quotas per Workspace and per User

This document outlines the architecture, data models, enforcement mechanisms, and instructions for managing resource quotas within CrownLabs.

---

## 1. Defining Quotas per Workspace

Each workspace defines its maximum allowed resource allocation in the `spec.quota` field of the CrownLabs `Workspace` Custom Resource. 

Example configuration for the `test` workspace:

```yaml
apiVersion: crownlabs.polito.it/v1alpha1
kind: Workspace
metadata:
  name: test
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
```

In the example above, users belonging to this workspace can launch up to **3 active instances** (VMs or containers) simultaneously, and globally across the workspace cannot exceed **10 CPUs**, **64Gi of RAM**, **15Gi of Persistent Disk**, **1 NVIDIA GPU**, and **1 AMD GPU**.

Resource quotas can be updated at any time by modifying the `quota` section of the `Workspace` resource.

### Resource Types Specification

The data types used within `ResourceSpec` in the Go backend are defined as follows:

| Resource Field | Go Data Type | Description |
| :--- | :--- | :--- |
| `instances` | `int64` | Maximum number of active/running instances allowed |
| `cpu` | `int64` | Maximum vCPU cores allocated |
| `memory` | `resource.Quantity` | Maximum RAM memory (e.g., `16Gi`) |
| `disk` | `resource.Quantity` | Maximum persistent storage allocated (e.g., `20Gi`) |
| `otherResources` | `map[string]resource.Quantity` | Generic key-value map for extended hardware accelerators |

---

## 2. Defining Quotas per User (Tenants)

While each CrownLabs workspace enforces its own boundaries, the total allowance assigned to a user depends on the **sum of the resource quotas** of all workspaces they belong to, plus their **Personal Workspace**.

For instance, if a user is enrolled in:
* **Workspace A:** CPU limit = `10`
* **Workspace B:** CPU limit = `8`

The total accumulated CPU limit for this user across the platform is `18`. 

> **NOTE on CPU Limits:** CPU cores are summed across workspaces. However, individual `Templates` (which belong to a specific Workspace) can define an extra property called `ReservedCPUPercentage` (e.g., `50%`), which throttles the actual vCPU allocation of instances spawned from that template.

### Personal Workspace Quota

Beyond the accumulated limits inherited from course workspaces, each user is allocated a **Personal Workspace**. This acts as a private, standalone sandbox. Its quota is configured directly within the `personalWorkspace` field of the `Tenant` Custom Resource:

```yaml
apiVersion: crownlabs.polito.it/v1alpha2
kind: Tenant
metadata:
  name: s343940
spec:
  firstName: John
  lastName: Doe
  personalWorkspace:
    cpu: 4
    memory: 8Gi
    instances: 2
    disk: 10Gi
    otherResources:
      nvidia.com/gpu: "0"
```

---

## 3. How to Add New Custom Resources (Step-by-Step Guide)

CrownLabs uses `otherResources: map[string]resource.Quantity` to avoid hardcoding specific hardware vendor keys in the backend Go codebase. Follow these steps to introduce a new custom resource (e.g., **Intel Gaudi TPU** `intel.com/gaudi` or new **AMD GPUs** `amd.com/gpu`):

### Step 1: Cluster & Device Plugin Setup
If physical hardware scheduling is required on worker nodes, ensure the hardware vendor's Device Plugin (e.g., NVIDIA GPU Operator, AMD GPU Device Plugin) is installed.

Once active, Kubernetes automatically registers the custom hardware under node allocatable resources:

```bash
kubectl get nodes -o jsonpath='{.items[*].status.allocatable}'
# Expected output on nodes with active plugins: includes something akin to "[intel.com/gaudi](https://intel.com/gaudi)": "2" or "[nvidia.com/gpu](https://nvidia.com/gpu)": "2"
```

> **Note on Logical vs Physical Quotas:** CrownLabs quota enforcement operates as an independent logical accounting layer. In test environments without physical GPU hardware or Device Plugins, CrownLabs will still track, display, and validate custom resource limits (`otherResources`) seamlessly across Workspaces and Tenants.

### Step 2: Key Normalization (Kubernetes vs GraphQL)
Kubernetes resource keys use slashes (e.g., `nvidia.com/gpu`). Because GraphQL field names cannot contain `/` or `.`, the `qlkube` middleware automatically converts keys between formats:
* **Kubernetes API:** `intel.com/gaudi`
* **GraphQL Schema (camelCase):** `intelComGaudi`

The frontend helper `getOriginalK8sKey(key)` automatically maps GraphQL camelCase keys back to standard Kubernetes format so that `totalQuota` and `consumedQuota` comparisons match seamlessly.

### Step 3: Frontend Environment Variable
Add the new resource to the `VITE_APP_CUSTOM_RESOURCES` environment variable in the frontend deployment configuration (`.env`):

```env
VITE_APP_CUSTOM_RESOURCES={"nvidia.com/gpu":"NVIDIA GPU","amd.com/gpu":"AMD GPU","intel.com/gaudi":"Intel Gaudi TPU"}
```

> ### **IMPORTANT : Future Enhancement (Dynamic Frontend Template Resources):**  
> Currently, the Custom Resource options available during Template creation/editing are statically loaded from the `VITE_APP_CUSTOM_RESOURCES` environment variable. In future iterations, the frontend UI should dynamically filter these options, displaying *only* the Custom Resources that are actively allocated (quota > 0) within the specific Workspace owning the Template.

### Step 4: Update Workspace or Tenant CRD
Administrators can now assign the new resource directly inside any `Workspace` or `Tenant` YAML definition:

```yaml
otherResources:
  intel.com/gaudi: "1"
```

## 4. Enforcement Mechanism: ResourceQuota vs Validation Webhook

CrownLabs enforces resource limits using a multi-tiered security and validation pipeline:

```
[ User Request ] 
       │
       ▼
[ 1. Validating Webhook ] ──(intercepts via namespaceSelector)──► [ Rejects if Invalid / Over-quota ]
       │
       ▼
[ 2. Instance Validator ] ──(Go component)─────────────────────► [ Verifies Workspace bounds ]
       │
       ▼
[ 3. ResourceQuota ]      ──(Tenant Namespace Level)───────────► [ Passive Global Security Barrier ]
```

1. **Validating Webhook (Cluster Level Enforcement):**
   CrownLabs deploys a `ValidatingWebhookConfiguration` that intercepts `CREATE` and `UPDATE` operations on `Tenants` and `Instances`. 
   * **Scope Filtering:** Uses `namespaceSelector` to match all CrownLabs-managed namespaces (`workspace-*` and `tenant-*`). This ensures **100% of instance creations are intercepted**, preventing validation bypasses.
   * **Failure Policy:** Configured with `failurePolicy: Fail` to block unvalidated or over-quota requests before they are written to `etcd`.

2. **ResourceQuota (Tenant Namespace Barrier):**
   When a Tenant is reconciled, the Tenant Operator creates a standard Kubernetes `ResourceQuota` inside the user's namespace (`tenant-<username>`). It acts as a passive container limit preventing cluster starvation.

3. **Instance Validator (Workspace Level):**
   A dedicated Go component that evaluates the real-time consumption of the user against their allowed course workspace limits before an instance is allowed to boot.

---

## 5. Runtime Behavior, Persistent Disk Edge Cases, and Auto-Recovery

### The Persistent Storage Allocation Trap
Standard compute resources (CPU/RAM) are validated at Pod scheduling time. If quota is exceeded, the API server rejects the request immediately.

However, for **persistent environments**, the Persistent Volume Claim (PVC) is provisioned **at the beginning of the deployment chain**, before compute validation hooks complete:
* **Previous Unhandled State:** A storage quota breach caused the PVC allocation to fail, putting the environment into a `CreationLoopBackoff` crash state and marking the entire `Instance` phase as unrecoverably `Failed`.

### The Requeuing Solution
To handle storage quota violations gracefully, the Instance Controller (`pkg/instctrl/controller.go`) implements explicit error trapping:

1. **Phase Shift:** When a disk allocation error occurs due to quota exhaustion, the controller intercepts the error and updates the individual environment status to `ResourceQuotaExceeded`.
2. **Deterministic Requeue:** Instead of dropping into a terminal `Failed` state, the controller schedules a clean reconciliation requeue every **60 seconds** (`ctrl.Result{RequeueAfter: 1 * time.Minute}`).
3. **Automatic Recovery:** As soon as the user terminates other persistent instances or an administrator increases the workspace disk limit, the next periodic check succeeds automatically, transitioning the instance to `Starting` and `Ready` without requiring manual intervention.