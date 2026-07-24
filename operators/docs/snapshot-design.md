# Persistent VM snapshot design

## Context

CrownLabs needs a system to create snapshots of persistent virtual machines. This functionality must be integrated within the instance operator ecosystem, but decoupled into a dedicated controller rather than adding extra logic inside the existing instance controller reconciliation path.

Today, a persistent VM Instance is materialized by the instance operator as:

- a CrownLabs `Instance` in the tenant namespace;
- one KubeVirt `VirtualMachine` per persistent VM environment;
- one CDI `DataVolume` per persistent VM environment, used as the VM root disk;
- optional MyDrive and SharedVolume mounts, which are separate PVCs and are not part of the VM root disk.

The snapshot feature must save the persistent root disk and enough immutable metadata to create a new VM from that disk later without going back to the original template image import path.

## Goals

- Snapshot only persistent VM environments backed by a `DataVolume`.
- Keep the feature in the instance operator by adding a dedicated snapshot controller, rather than mixing the snapshot lifecycle into `instctrl.InstanceReconciler`.
- Store a logical CrownLabs snapshot object in one of three destinations:
  - tenant namespace (`tenant-<tenant>`), used as the private destination;
  - workspace namespace (`workspace-<workspace>`);
  - public snapshot namespace (configured via Helm).
- Freeze the source and destination metadata at creation time. After creation, the snapshot `spec` should be immutable.
- Support a fast restore path by creating a target `DataVolume` from the snapshot PVC via CDI smart cloning. Snapshots are primarily **reusable template bases**, not user backups.
- Enforce authorization based on the actor creating the snapshot: normal user, workspace manager, cluster admin.
- Count snapshot storage against tenant/workspace resource quotas. This applies to the number of snapshots (via `count/instancesnapshots.crownlabs.polito.it` Kubernetes ResourceQuota). Storage size is also quota-enforced through standard PVC storage quotas, since the snapshot artifact is a regular PVC.

## Non-goals

- Snapshot non-persistent VMs, containers, MyDrive, or SharedVolumes.
- Provide a full backup/export system to object storage. Snapshots are reusable template bases, not a backup solution.
- Restore by mutating an existing running VM disk in place.
- Online (live) snapshots. The VM must be powered off before a snapshot can be taken. This is a hard requirement, not a policy default.
- CSI `VolumeSnapshot`-based snapshots. The design uses PVC clones via CDI instead of CSI VolumeSnapshots.

## Main design choice

Add a new namespaced CrownLabs CRD, tentatively named `InstanceSnapshot`.

`metadata.namespace` is the destination catalog namespace. `spec.source.instanceRef` points to the source `Instance`, which normally lives in a tenant namespace.

Since the VM must be powered off before snapshotting, the source PVC is idle and the disk is quiesced by definition. This means KubeVirt `VirtualMachineSnapshot` is not needed — its value is in coordinating live snapshots with the guest agent, which does not apply here. Instead, the controller operates directly on PVCs through CDI:

- **Private (same-namespace) snapshots**: create a CDI `DataVolume` clone from the source PVC within the tenant namespace. CDI smart cloning (CSI volume cloning) ensures this is nearly instantaneous on supported storage backends.
- **Workspace/public (cross-namespace) snapshots**: create a CDI `DataVolume` clone from the source PVC into the destination namespace. CDI handles cross-namespace data movement natively.

In both cases the final artifact is a regular PVC (created through a CDI `DataVolume`). The clone is a fully independent copy at the Kubernetes level — even if the original VM and its PVC are later deleted, the snapshot PVC continues to work. At the storage level (e.g., Ceph/RBD), the CSI driver may internally use copy-on-write for the clone, but this is transparent and the driver handles parent lifecycle automatically.

*Why use a PVC clone instead of a CSI VolumeSnapshot?*
- **Independence**: A cloned PVC is a standalone Kubernetes resource with no dependency on the source PVC lifecycle and no risk of dangling snapshot references.
- **Simplicity**: The controller only manages `DataVolume` resources (already used throughout CrownLabs). No need for `VolumeSnapshot`, `VolumeSnapshotClass`, or `VolumeSnapshotContent` resources and their associated RBAC.
- **Fewer prerequisites**: No CSI snapshot support required. Only CDI (already a hard dependency) and CSI volume cloning (supported by Rook-Ceph and most modern CSI drivers).
- **Restore performance**: Creating student VMs from a snapshot PVC leverages CDI smart cloning (PVC-to-PVC), which is nearly instantaneous on backends that support CSI volume cloning.
- **Storage cost trade-off**: A PVC clone allocates full storage (e.g., 30GB), unlike a `VolumeSnapshot` which would use copy-on-write at the storage level. This is an acceptable trade-off for the simplicity and independence guarantees.
- **Immutability trade-off**: Unlike a `VolumeSnapshot`, a PVC is read-write. To prevent accidental corruption of the snapshot template, the snapshot PVC should not be mounted directly by any VM. The controller should add a protective label (e.g., `crownlabs.polito.it/snapshot-artifact: true`), and the restore flow always creates a *new* clone from the snapshot PVC, never using it directly.

This approach avoids both the KubeVirt snapshot API and the CSI VolumeSnapshot API, reduces the number of resource types involved, and simplifies both the controller logic and the RBAC surface.

## API sketch

```yaml
apiVersion: crownlabs.polito.it/v1alpha2
kind: InstanceSnapshot
metadata:
  name: linux-vm-lab1
  namespace: tenant-mario-rossi
spec:
  displayName: "Lab 1 baseline"
  source:
    instanceRef:
      name: linux-vm
      namespace: tenant-mario-rossi
    environmentName: desktop
    environmentType: VirtualMachine
    tenantRef:
      name: mario.rossi
    templateRef:
      name: ubuntu
      namespace: workspace-os
    workspaceRef:
      name: os
    pvcName: linux-vm-desktop
    disk:
      storageClassName: rook-ceph-block
      volumeMode: Block
      accessModes:
        - ReadWriteOnce
      size: 30Gi
  destination:
    scope: Private
    namespace: tenant-mario-rossi
status:
  phase: Ready
  artifact:
    dataVolumeRef:
      name: linux-vm-lab1-a1b2c3
      namespace: tenant-mario-rossi
  conditions: []
```

The user-facing create request can be smaller:

```yaml
spec:
  displayName: "Lab 1 baseline"
  source:
    instanceRef:
      name: linux-vm
      namespace: tenant-mario-rossi
    environmentName: desktop
    environmentType: VirtualMachine
  destination:
    scope: Private
```

A mutating admission webhook should resolve and freeze the remaining `spec.source` and `spec.destination.namespace` fields at creation time. A validating webhook should reject every later `spec` change.

*Note on the trade-off*: The webhook resolves all necessary metadata (Template, Environment, PVC) to freeze it in the `spec`. While this introduces external API calls during admission (potentially increasing latency or transient failures), it guarantees atomic immutability: the `spec` is completely frozen before the CR is persisted, removing the need to handle partially-populated `spec`s in the controller.
The API targets a single environment per `InstanceSnapshot` request (`spec.source.environmentName`). If an Instance contains multiple environments (e.g., multiple VMs), the user must create a separate `InstanceSnapshot` for each environment they wish to snapshot.

## Snapshot lifecycle

1. Admission receives an `InstanceSnapshot` create request.
2. The webhook resolves the source `Instance`, `Template`, `Tenant`, `Workspace`, and PVC.
3. The webhook validates:
   - the environment exists;
   - the environment type is VM-like;
   - `persistent: true`;
   - the source DataVolume/PVC exists and is already populated;
   - **the VM is not running** (`Instance.spec.running=false`). If the VM is running, the webhook rejects the request and the user is notified;
   - the destination namespace matches the selected scope;
   - the actor is authorized.

### Private (same-namespace) flow

4. The controller verifies again that the VM is not running (`Instance.spec.running=false` and no `VirtualMachineInstance` object exists). If running, the reconciliation fails and requeues. If off, the controller creates a CDI `DataVolume` clone from the source PVC, in the tenant namespace. CDI smart cloning (CSI volume cloning) ensures this is nearly instantaneous on supported storage backends.
5. Once the `DataVolume` reports a ready condition and the underlying PVC is bound, the controller sets `InstanceSnapshot.status.phase=Ready`.

### Cross-namespace (workspace/public) flow

4. The controller verifies again that the VM is not running. If off, the controller creates a CDI `DataVolume` in the destination namespace with `spec.source.pvc` pointing to the source PVC (cross-namespace clone).
5. Once the `DataVolume` is ready and the underlying PVC is bound, the controller sets `InstanceSnapshot.status.phase=Ready`.

### Cleanup

6. A finalizer on the `InstanceSnapshot` cleans up the `DataVolume` (and its underlying PVC) artifact when the `InstanceSnapshot` is deleted, unless a future retention policy explicitly says otherwise.

## Restore lifecycle

Use a separate CRD, tentatively named `InstanceSnapshotRestore`, instead of changing the existing Instance reconciliation path heavily.

The restore controller should:

1. validate that the actor can consume the referenced `InstanceSnapshot`;
2. create the target `Instance` or require it to exist but not yet be running;
3. pre-create the target `DataVolume` with the exact name expected by `instctrl` for that Instance/environment. The restore controller **must not** set any owner reference on this `DataVolume`, allowing `instctrl` to adopt it later;
4. set the `DataVolume` source to the snapshot PVC (via `spec.source.pvc` pointing to the `InstanceSnapshot` artifact PVC);
5. let the existing `InstanceReconciler` create the VM normally after the DataVolume is ready.

This works with current instance-controller behavior because `instctrl` explicitly checks `dv.CreationTimestamp.IsZero()` before forging the `DataVolume` spec. To ensure this contract is preserved in the future, integration tests must explicitly verify that `instctrl` does not overwrite pre-existing `DataVolume` specs. If the restore controller creates the DataVolume first, `instctrl` keeps that source instead of recreating it from the original template image.

*Note on `instctrl` compatibility*: Currently, `instctrl` unconditionally clears the `DataVolume` owner references (`dv.OwnerReferences = nil`) during reconciliation. This is a temporary workaround that must be refactored before implementing the restore flow. `instctrl` should be updated to safely adopt a pre-existing `DataVolume` (e.g., using `controllerutil.SetControllerReference` without blindly clearing existing owners), ensuring it works seamlessly with `DataVolume`s pre-created by the restore controller.

## Authorization proposal

Kubernetes RBAC alone is not enough, because the policy depends on the source Instance owner, workspace membership, and destination scope. Use normal RBAC as the coarse permission layer, then enforce contextual rules in the validating webhook.

| Actor | Private destination | Workspace destination | Public destination |
| --- | --- | --- | --- |
| Normal user | Allowed only for own Instance and own private namespace | Denied | Denied |
| Workspace manager | Allowed for any Instance belonging to a managed workspace, including into the student's own private namespace | Allowed for Instances belonging to a managed workspace | Denied |
| Cluster admin | Allowed | Allowed | Allowed |

Workspace manager identity can use existing Keycloak/Kubernetes groups shaped as `kubernetes:workspace-<workspace>:manager`. Normal user identity can be matched against `Instance.spec.tenant.crownlabs.polito.it/TenantRef.name` and the tenant namespace.

The destination namespace should be derived, not trusted from user input:

- `Private`: `tenant-<tenant>`, matching `forge.GetTenantNamespaceName`.
- `Workspace`: `workspace-<workspace>`, matching `forge.GetWorkspaceNamespaceName`.
- `Public`: configured through a Helm value, for example `snapshotPublicNamespace`.

## RBAC changes

The instance operator service account needs permissions for:

- `crownlabs.polito.it`: `instancesnapshots`, `instancesnapshots/status`, optional `instancesnapshotrestores`, `instancesnapshotrestores/status`;
- `crownlabs.polito.it`: read `instances`, `templates`, `tenants`, `workspaces`;
- `kubevirt.io`: read `virtualmachines`, `virtualmachineinstances` (to verify VM running state);
- `cdi.kubevirt.io`: manage `datavolumes` (for same-namespace and cross-namespace cloning);
- core API: read/manage the involved PVCs and create events.

Note: Neither `snapshot.storage.k8s.io` nor `snapshot.kubevirt.io` permissions are needed because the design uses CDI `DataVolume` clones instead of CSI `VolumeSnapshot`s or KubeVirt `VirtualMachineSnapshot`s.

User-facing ClusterRoles should also be updated or split:

- extend `crownlabs-manage-instances` with `instancesnapshots` create/get/list/watch/delete where appropriate;
- add a view-only role for consuming public/workspace snapshots if the frontend needs browse-only access;
- keep `instancesnapshots/status` reserved to controllers.

## Consistency policy

The VM **must** be powered off (`Instance.spec.running=false`) before a snapshot can be taken. This is a hard requirement enforced at admission time, not a configurable policy.

If the VM is running when the snapshot is requested, the validating webhook rejects the create request. The user must stop the VM first and then retry. This guarantees filesystem consistency without requiring a guest agent.

Online snapshots are explicitly out of scope. The `spec.policy` section is removed from the API. If online snapshots are needed in the future, they can be added as a new API field with appropriate safety checks.

## Metadata to freeze

The immutable `spec.source` should store at least:

- source Instance name and namespace;
- source Tenant ref;
- source Workspace ref;
- source Template ref;
- environment name, type, persistence flag;
- PVC name;
- disk size, storage class, access modes, volume mode;
- CrownLabs labels needed for filtering: tenant, workspace, template, environment;
- creation actor and requested destination scope.

Mutable execution details belong in `status`:

- controller phase;
- final artifact references (`DataVolume`/PVC name and namespace);
- error messages and conditions;
- timestamps.

## Failure handling

- If source validation fails (including VM running check), reject creation in the webhook.
- If the source `Instance` is deleted after the webhook admission but before the controller starts processing the snapshot, the controller will fail to retrieve it (`NotFound`). The controller should mark the `InstanceSnapshot` as `Failed` and emit a `SourceInstanceDeleted` event.
- If the CDI `DataVolume` clone fails (same-namespace or cross-namespace), keep the `DataVolume` for debugging; the controller can retry on the next reconciliation since the source PVC is still intact. Set `status.phase=Failed` after exhausting retries.
- If the source Instance (and its DataVolume) is deleted while a CDI clone is in progress, the clone operation will fail. This is a known limitation. The controller will detect the failure, mark the `InstanceSnapshot` as `Failed` with a clear event, and clean up any partial artifacts.
- Use finalizers for cleanup. Do not set owner references from destination CRs to source-namespace resources, because cross-namespace ownership is not valid.
- Reconciliation must be idempotent: each phase should be recoverable from already-existing resources by labels and owner/finalizer metadata.

## Naming convention

The generated artifacts follow a deterministic naming convention to avoid collisions, especially in shared destination namespaces (Workspace and Public). The names are derived from the `InstanceSnapshot` CR name and UID:

- Destination `DataVolume`: `<instancesnapshot-name>-<short-uid>`

## Prerequisites

To support the cross-namespace cloning flow, the cluster must satisfy the following requirements:
- CDI is installed and configured to allow cross-namespace data movement. The `CDIConfig` must permit it, and appropriate RBAC `CloneRole`s must be granted to the CDI controller in the source namespaces.
- The CSI driver supports volume cloning (for CDI smart cloning). This is supported by Rook-Ceph and most modern CSI drivers. CSI snapshot support (`VolumeSnapshot`) is **not** required.

## Observability

The snapshot controller must expose standard Prometheus metrics:
- Number of snapshots in progress, completed, and failed.
- Duration of the snapshot operations.

It must also emit Kubernetes `Events` on the `InstanceSnapshot` object for key lifecycle transitions (e.g., `SnapshotStarted`, `CloneStarted`, `SnapshotReady`, `SnapshotFailed`), which the CrownLabs frontend can use to provide user feedback.

## Implementation plan

1. Add `InstanceSnapshot` API types under `operators/api/v1alpha2`, CRD generation, deepcopy generation, and samples.
2. Add the snapshot controller package under `operators/pkg/instsnapshotctrl` (rename the empty `operators/pkg/instancesnapshot-controller`, following the `instctrl`/`instautoctrl` naming convention).
3. Add Helm values for `snapshotPublicNamespace` and snapshot failure deadline.
4. Add mutating/validating admission for metadata freezing, immutable spec, VM-off enforcement, and authorization matrix.
5. Add controller reconciliation for:
   - same-namespace flow: PVC → CDI DataVolume clone;
   - cross-namespace flow: PVC → CDI DataVolume clone (cross-namespace).
6. Add restore CRD/controller only after the snapshot artifact is stable.
7. Update RBAC manifests and user-facing ClusterRoles.
8. Add snapshot quota enforcement (`count/instancesnapshots.crownlabs.polito.it`).
9. Add frontend/API support after the CRD and authorization semantics are fixed.

## Test plan

- Unit tests for source metadata resolution.
- Unit tests for the authorization matrix (including workspace manager → student private namespace).
- Unit tests for immutable `spec` validation.
- Unit tests for VM-running rejection at admission.
- Controller tests for each lifecycle phase (same-namespace and cross-namespace flows) by faking CDI statuses.
- Controller tests verifying the re-check of VM state (must be off and VMI missing) before snapshotting.
- Failure tests for missing source Instance (both at admission and during controller reconciliation), non-persistent environment, missing DataVolume, unauthorized destination, VM still running, failed CDI clone, and missing CDI configuration.
- E2E test on a cluster with CDI smart cloning enabled, because fake clients cannot prove that the storage clone path actually works.

## Resolved decisions

The following questions were resolved during design review and their answers are integrated into the sections above.

1. **Private namespace**: the tenant namespace (`tenant-<tenant>`), matching `forge.GetTenantNamespaceName`. *(integrated into Goals and Authorization)*
2. **Offline-only**: the VM must be powered off before snapshotting. If the VM is running, the webhook rejects the request and notifies the user. No online snapshot support in v1. *(integrated into Non-goals, Consistency policy, and Lifecycle)*
3. **Cross-namespace copy**: yes, copy the data into the destination namespace to decouple from the source tenant lifecycle.
4. **Quota enforcement**: yes, snapshot storage counts against tenant/workspace quotas. *(integrated into Goals and Implementation plan)*
5. **Workspace manager → student private namespace**: allowed. Any user with workspace manager privileges can snapshot a student's Instance, including into the student's private namespace. *(integrated into Authorization)*
6. **Template bases, not backups**: snapshots are primarily reusable template bases. The restore flow creates new Instances from snapshot artifacts rather than restoring in-place. *(integrated into Goals and Non-goals)*
7. **PVC clone only (no VolumeSnapshot)**: the final artifact is a CDI `DataVolume` (PVC). CSI `VolumeSnapshot` is not used; the design relies on CDI smart cloning (PVC-to-PVC) for both the snapshot and restore flows. The storage cost trade-off (full PVC allocation vs. COW snapshot) is accepted for simplicity and independence from the source VM lifecycle. *(integrated into Main design choice and Non-goals)*
8. **No KubeVirt VirtualMachineSnapshot**: since the VM is guaranteed to be off, the source PVC is idle and can be cloned directly via CDI. KubeVirt's snapshot API adds no value in this scenario and is not used. *(integrated into Main design choice, Lifecycle, and RBAC)*
