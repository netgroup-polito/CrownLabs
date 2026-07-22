# CrownLabs operator — local setup

Base RBAC and running the CrownLabs main operator against the local realm,
so `Tenant`/`Workspace` resources actually reconcile (personal namespace,
`mydrive`, Keycloak `RoleBinding`s) instead of sitting there unowned.

## Prerequisites

- The base k3s cluster — see [`../base-k3s/README.md`](../base-k3s/README.md).
- [Envoy Gateway](../envoy/README.md), Keycloak, and the k3s apiserver OIDC
  integration — see [`../keycloak/README.md`](../keycloak/README.md). The
  operator needs the realm's `operator-local` client secret and the CA
  certificate extracted in that guide's Step 2.
- `helm` (to render the base RBAC resources from the real chart, below).
- Go >= 1.25.

## 1. Base CrownLabs RBAC (`ClusterRole`/`ClusterRoleBinding`)

The operator's per-tenant/per-workspace `RoleBinding`s reference base
`ClusterRole`s (`crownlabs-manage-tenants`, `crownlabs-manage-instances`,
`crownlabs-view-workspaces`, ...) that are normally installed by the main
Helm chart — but are missing here since these local steps run the operator
directly via `go run`/`make`, not through Helm. Render them straight from
the real chart (`deploy/crownlabs`) instead of keeping a hand-maintained
copy that can drift from it:

```bash
cd dev-local/keycloak   # the chart's clusterroles/clusterrolebindings templates
                        # are rendered relative to this folder's location
helm dependency build ../../deploy/crownlabs   # one-time; only needed if deploy/crownlabs/charts is empty
helm template -s templates/clusterroles.yaml -s templates/clusterrolebindings.yaml ../../deploy/crownlabs | kubectl apply -f -
```

Purely additive: creating a `ClusterRole` grants nothing by itself until
something is bound to it, so applying this is always safe, and safe to
re-run any time the chart's RBAC templates change.

## 2. Run the CrownLabs operator

```bash
export SSL_CERT_FILE=~/certs/crownlabs-ca.crt   # the operator's Go HTTP client must trust the Gateway's self-signed cert too (see ../keycloak/README.md Step 2)

cd operators
make run-operator-local KEYCLOAK_CLIENT_SECRET=operator-local-dev-secret
```

`KEYCLOAK_URL` defaults to `https://keycloak.crownlabs.local` (the Gateway's
`HTTPRoute` hostname, see [`../envoy/README.md`](../envoy/README.md)).

> **On WSL2**: also override `KEYCLOAK_URL` to the Gateway's port-forward bridge
> (Keycloak isn't reachable from WSL2 itself either, not just from the Windows
> browser — see `../envoy/README.md` step 5 for why):
> ```bash
> make run-operator-local KEYCLOAK_URL=https://keycloak.crownlabs.local:8443 KEYCLOAK_CLIENT_SECRET=operator-local-dev-secret
> ```

`run-operator-local` (in `operators/Makefile`) already sets
`KEYCLOAK_URL`/`KEYCLOAK_REALM`/`KEYCLOAK_CLIENT_ID`/`KEYCLOAK_TARGET_CLIENT`
to match the local realm, and `MYDRIVE_STORAGE_CLASS=local-path` (k3s's
default storage class — override if yours differs). Only
`KEYCLOAK_CLIENT_SECRET` needs to be passed explicitly (it's per-realm-install,
so there's no sensible default to bake in — see `../keycloak/README.md` for
the value this realm export ships with).

Applying CRDs and `operators/samples/` is a prerequisite of the target
(same pattern as `run-instance-local`), so **sample `Tenant`/`Workspace`/
`Template` resources are created automatically** — no separate step needed.
If you'd rather test with a real Keycloak user instead of (or in addition
to) the samples, create a matching `Tenant` by hand, e.g. for `john.doe`:

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

> **Gotcha**: `run-operator-local` (and `run-instance-local`) re-apply
> `operators/samples/` on every run. `kubectl apply` does a 3-way merge
> against the *last-applied-configuration* it recorded — not against
> whatever the object currently looks like — so any field you changed
> out-of-band with `kubectl patch` (e.g. adding a `Workspace` enrollment to
> `Tenant.spec.workspaces` per Step 7 of
> `../keycloak/apiserver-oidc-integration.md`) gets **silently reverted** to
> whatever `operators/samples/tenant.yml` says the next time this target
> runs, since that patch was never reflected in the sample file. Symptom: a
> `403 Forbidden` on a namespace/role that worked a moment ago. Fix:
> re-apply the `kubectl patch`, or (if it's a mapping you want to keep
> permanently) add it to the sample file itself instead of patching the
> live object.

### Forcing a reconciliation

The operator doesn't watch Keycloak in real time (no active webhook
locally). If you change something on the Keycloak side (e.g. verify a user)
and want the operator to pick it up immediately, force a new reconciliation
with a harmless annotation:

```bash
kubectl annotate tenant.crownlabs.polito.it <tenant-name> crownlabs.polito.it/force-reconcile="$(date +%s)" --overwrite
```

## Final checks

```bash
kubectl get pods -A
# everything Running/Completed, no CrashLoopBackOff

kubectl get tenants
kubectl describe tenant <tenant-name>
# status.ready: true, no reconciliation errors
```

Expected end state:
- ✅ Base `ClusterRole`/`ClusterRoleBinding`s applied.
- ✅ CrownLabs operator running, `Tenant`/`Workspace` reconciled.

## Next

For qlkube (talking directly to the real apiserver instead of through
`kubectl proxy`) and the frontend, see
[`../local-development.md`](../local-development.md).
