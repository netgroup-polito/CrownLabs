# CrownLabs operator — local setup

This guide explains how to set up base RBAC, and how to run the CrownLabs main operator against the local realm.
Without this, `Tenant` and `Workspace` resources never reconcile: no personal namespace, no `mydrive`, and no Keycloak `RoleBinding`s. They just sit there, unowned.

## Prerequisites

- The base k3s cluster. See [`../base-k3s/README.md`](../base-k3s/README.md).
- [Envoy Gateway](../envoy/README.md), Keycloak, and the k3s API server OIDC integration. See [`../keycloak/README.md`](../keycloak/README.md).
  The operator needs the realm's `operator-local` client secret, and the CA certificate you extracted in that guide's step 2.
- `helm` (to render the base RBAC resources from the real chart, below).
- Go >= 1.25.

## 1. Base CrownLabs RBAC (`ClusterRole`/`ClusterRoleBinding`)

The operator's per-tenant and per-workspace `RoleBinding`s reference base `ClusterRole`s (`crownlabs-manage-tenants`, `crownlabs-manage-instances`, `crownlabs-view-workspaces`, and others).
The main Helm chart normally installs these, but they are missing here, because these local steps run the operator directly with `go run`/`make`, not through Helm.
Render them straight from the real chart (`deploy/crownlabs`), instead of keeping a hand-maintained copy that can drift from it:

```bash
cd dev-local/keycloak   # The chart's clusterroles/clusterrolebindings templates
                        # are rendered relative to this folder's location.
helm dependency build ../../deploy/crownlabs   # One-time step; only needed if deploy/crownlabs/charts is empty.
helm template -s templates/clusterroles.yaml -s templates/clusterrolebindings.yaml ../../deploy/crownlabs | kubectl apply -f -
```

These RBAC definitions are purely additive.
Creating a `ClusterRole` grants nothing by itself, until something is bound to it.
This means applying them is always safe, and safe to re-run any time the chart's RBAC templates change.

## 2. Run the CrownLabs operator

```bash
# The operator's Go HTTP client must also trust the Gateway's self-signed certificate.
# See ../keycloak/README.md step 2.
export SSL_CERT_FILE=~/certs/crownlabs-ca.crt

cd operators
make run-operator-local KEYCLOAK_CLIENT_SECRET=operator-local-dev-secret
```

`KEYCLOAK_URL` defaults to `https://keycloak.crownlabs.local` (the Gateway's `HTTPRoute` hostname, see [`../envoy/README.md`](../envoy/README.md)).

> **On WSL2**: also override `KEYCLOAK_URL` to the Gateway's port-forward bridge.
> Keycloak is not reachable from WSL2 itself either, not just from the Windows browser.
> See `../envoy/README.md` step 5 for why.
> ```bash
> make run-operator-local KEYCLOAK_URL=https://keycloak.crownlabs.local:8443 KEYCLOAK_CLIENT_SECRET=operator-local-dev-secret
> ```

`run-operator-local` (in `operators/Makefile`) already sets `KEYCLOAK_URL`, `KEYCLOAK_REALM`, `KEYCLOAK_CLIENT_ID`, and `KEYCLOAK_TARGET_CLIENT` to match the local realm.
It also sets `MYDRIVE_STORAGE_CLASS=local-path` (k3s's default storage class; override this if yours is different).
Only `KEYCLOAK_CLIENT_SECRET` needs to be passed explicitly. This value is specific to each realm install, so there is no sensible default to bake in. See `../keycloak/README.md` for the value this realm export ships with.

Applying the CRDs and `operators/samples/` is a prerequisite of this target, the same way it is for `run-instance-local`.
This means **sample `Tenant`, `Workspace`, and `Template` resources are created automatically**: you do not need a separate step for them.
If you would rather test with a real Keycloak user, instead of the samples, or alongside them, create a matching `Tenant` by hand. For example, for `john.doe`:

```bash
kubectl apply -f - <<'EOF'
apiVersion: crownlabs.polito.it/v1alpha2
kind: Tenant
metadata:
  name: john.doe
  labels:
    crownlabs.polito.it/operator-selector: local
spec:
  email: john.doe@example.com
  firstName: John
  lastName: Doe
  workspaces: []
EOF
```

The operator starts reconciling automatically.

> **Gotcha**: `run-operator-local` (and `run-instance-local`) re-apply `operators/samples/` on every run.
> `kubectl apply` does a 3-way merge against the *last-applied-configuration* it recorded, not against whatever the object currently looks like.
> So, if you changed a field out-of-band with `kubectl patch` (for example, adding a `Workspace` enrollment to `Tenant.spec.workspaces`, per step 7 of `../keycloak/apiserver-oidc-integration.md`), that change is never reflected in the sample file.
> The next time this target runs, `operators/samples/tenant.yml` **silently reverts your change**.
> Symptom: a `403 Forbidden` on a namespace or role that worked a moment ago.
> Fix: re-apply the `kubectl patch`. If it is a mapping you want to keep permanently, add it to the sample file itself, instead of patching the live object.

### Forcing a reconciliation

The operator does not watch Keycloak in real time: there is no active webhook locally.
If you change something on the Keycloak side (for example, verifying a user), and you want the operator to pick it up immediately, force a new reconciliation with a harmless annotation:

```bash
kubectl annotate tenant.crownlabs.polito.it <tenant-name> crownlabs.polito.it/force-reconcile="$(date +%s)" --overwrite
```

## Final checks

```bash
kubectl get pods -A
# Expected: everything is Running or Completed, with no CrashLoopBackOff.

kubectl get tenants
kubectl describe tenant <tenant-name>
# Expected: status.ready is true, with no reconciliation errors.
```

Expected end state:
- ✅ Base `ClusterRole`/`ClusterRoleBinding`s applied.
- ✅ CrownLabs operator running, with `Tenant` and `Workspace` reconciled.

## Next

For qlkube (which talks directly to the real API server, instead of through `kubectl proxy`) and for the frontend, see [`../local-development.md`](../local-development.md).
