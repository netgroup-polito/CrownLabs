# Local Keycloak + OIDC — end-to-end setup

Everything needed to bring up Keycloak (realm, clients, scopes, base users) and the
Kubernetes apiserver validating Keycloak-issued tokens (instead of every local request
silently running as cluster-admin under `kubectl proxy`), on top of an already-existing
k3s cluster with [Envoy Gateway](../envoy/README.md) set up.

- [`manifests/`](manifests/) — applied with a single
  `kubectl apply -f dev-local/keycloak/manifests`: Keycloak + Postgres, its `HTTPRoute`,
  the realm import, and the `mydrive-pvcs` namespace.
- [`apiserver-oidc-integration.md`](apiserver-oidc-integration.md) — the
  rationale, verification steps and troubleshooting table behind the OIDC
  setup below. Read it if something in this guide doesn't behave as
  expected; this README only has the commands.

## 0. Prerequisites

- The base k3s cluster — see [`../base-k3s/README.md`](../base-k3s/README.md).
- [Envoy Gateway](../envoy/README.md) set up (Gateway API CRDs, the controller, the
  shared `crownlabs` Gateway, the `*.crownlabs.local` wildcard cert) — Keycloak's own
  manifest has no HTTPS/NodePort of its own anymore, it's exposed entirely through the
  `HTTPRoute` in `manifests/keycloak.yaml`, which needs that `Gateway` to already exist.

```bash
kubectl get nodes
# expected: your node, Ready
kubectl get gateway crownlabs -n default
# expected: PROGRAMMED = True
```

## 1. Keycloak and the `mydrive-pvcs` namespace

```bash
kubectl apply -f dev-local/keycloak/manifests
kubectl rollout status statefulset/keycloak
```

This single apply covers:
- Keycloak + Postgres (`manifests/keycloak.yaml`), with `--import-realm` enabled and the
  realm ConfigMap mounted. Keycloak itself only ever speaks plain HTTP — TLS terminates
  at the Gateway (see [`../envoy/README.md`](../envoy/README.md)), not here, so there's
  no certificate/HTTPS config on Keycloak's own side at all.
- An `HTTPRoute` (same file) routing `keycloak.crownlabs.local` to that Service through
  the shared `crownlabs` Gateway.
- The realm import itself (`manifests/crownlabs-realm-configmap.yaml`) — the
  `crownlabs` realm export (clients `k8s`/`operator-local`, scopes
  `groups`/`k8s-audience`, base users), wrapped in a `ConfigMap` so it's
  applied declaratively instead of via an easy-to-forget imperative
  `kubectl create configmap` step.
- The `mydrive-pvcs` namespace (`manifests/mydrive-pvcs-namespace.yaml`),
  used by the operator for tenants' personal-drive PVCs.

Keycloak is now reachable at `https://keycloak.crownlabs.local` — no `kubectl
port-forward` needed on native Linux (see [`../envoy/README.md`](../envoy/README.md) for
how that's wired: one shared Gateway, `ServiceLB`-backed on 80/443, instead of a
per-service `NodePort`).

> **On WSL2**: same underlying limitation as everything else behind the Gateway — see
> [`../envoy/README.md`](../envoy/README.md) step 5 for the one bridge you need
> (`kubectl port-forward ... 8443:443`). **Every hostname in this guide and in
> `apiserver-oidc-integration.md` needs `:8443` appended on WSL2**
> (`https://keycloak.crownlabs.local:8443`) — this must exactly match whatever URL the
> browser actually reaches Keycloak through, since that's what ends up in the token's
> `iss` claim.
>
> **Also on WSL2**: Keycloak (`KC_HOSTNAME_STRICT=false`) infers its own issuer from the
> request it receives, but — tested and confirmed — it reports the issuer **without**
> a port regardless of which port the request actually arrived on (Envoy Gateway
> doesn't forward a `X-Forwarded-Port`). On native Linux that's fine (port 443 is the
> default, no bridge involved). On WSL2 it's not: the issuer needs to say `:8443`
> explicitly, or it won't byte-for-byte match the `oidc-issuer-url` the apiserver is
> configured with in step 2 below, and every token gets rejected with an issuer
> mismatch. Pin it explicitly after applying the manifests:
> ```bash
> kubectl set env statefulset/keycloak KC_HOSTNAME=https://keycloak.crownlabs.local:8443
> kubectl rollout status statefulset/keycloak
> ```
> This is a live patch, not part of `manifests/keycloak.yaml` (which stays
> environment-agnostic) — re-run it if you ever re-`kubectl apply` the base manifest on
> WSL2, same gotcha as the `Tenant`/`Workspace` patches in `../operators/README.md`.

The self-signed wildcard certificate (issued at the Gateway, see
[`../envoy/README.md`](../envoy/README.md)) triggers a browser warning
(`ERR_CERT_AUTHORITY_INVALID`) the first time you visit
`https://keycloak.crownlabs.local/...` — expected for local dev; click through it once
("Advanced" → "Proceed") and the browser trusts it for the rest of the session.

On first startup, Keycloak imports the realm from the mounted ConfigMap
automatically (log: `KC-SERVICES0032: Import JSON RealmRepresentation from
file .../crownlabs-realm.json`). **If the realm already exists in the
database, the import is silently skipped** — standard `--import-realm`
behavior, meant to avoid losing state across restarts. To re-import from
scratch, delete the realm first (Admin Console, or `DELETE
/admin/realms/crownlabs`), or wipe the Postgres database.

Credentials included in the export:

| Who | Credentials |
|---|---|
| Keycloak admin (`master` realm) | `admin` / `admin` (set via env var, not part of the export) |
| Test user (`crownlabs` realm) | `john.doe` / `johndoe123`, email verified |
| `operator-local` client (service account for the operator) | client secret: `operator-local-dev-secret` |
| `k8s` client (frontend, public) | no secret, redirect URI already set to `http://localhost:3000/*` |

> These are **known values, deliberately documented for local development**,
> exactly like `admin`/`admin` in the manifest — not meant for real
> environments.

## 2. Configure the k3s apiserver for OIDC

This is the one step that genuinely requires root and a k3s restart, so it
can't be folded into the manifests above. See
[`apiserver-oidc-integration.md`](apiserver-oidc-integration.md) for the
full rationale (why each flag, the `groups`/`aud` claim mappers, etc.) —
here's just the commands:

```bash
mkdir -p ~/certs
kubectl get secret crownlabs-tls -n default -o jsonpath='{.data.ca\.crt}' | base64 -d > ~/certs/crownlabs-ca.crt
sudo cp ~/certs/crownlabs-ca.crt /etc/rancher/k3s/crownlabs-ca.crt
```

Write a dedicated drop-in file (requires root) instead of editing the shared
`/etc/rancher/k3s/config.yaml` — k3s merges every `*.yaml` under
`/etc/rancher/k3s/config.yaml.d/` automatically, so this can't clobber
whatever [`../base-k3s/README.md`](../base-k3s/README.md) or
[`../envoy/README.md`](../envoy/README.md) already put there:

```bash
sudo mkdir -p /etc/rancher/k3s/config.yaml.d
sudo tee /etc/rancher/k3s/config.yaml.d/30-keycloak-oidc.yaml > /dev/null <<'EOF'
kube-apiserver-arg:
  - "oidc-issuer-url=https://keycloak.crownlabs.local/realms/crownlabs"
  - "oidc-client-id=k8s"
  - "oidc-username-claim=preferred_username"
  - "oidc-username-prefix=-"
  - "oidc-groups-claim=groups"
  - "oidc-groups-prefix=kubernetes:"
  - "oidc-ca-file=/etc/rancher/k3s/crownlabs-ca.crt"
EOF
```

> **On WSL2**: use `oidc-issuer-url=https://keycloak.crownlabs.local:8443/realms/crownlabs`
> instead (the Gateway bridge port from [`../envoy/README.md`](../envoy/README.md) step
> 5) — this must exactly match whatever URL the browser actually reaches Keycloak
> through, since that's what ends up in the token's `iss` claim.

```bash
sudo systemctl restart k3s
journalctl -u k3s -f   # confirm no "invalid authentication configuration" error, then Ctrl+C
```

**This restarts the whole control plane** — the apiserver is briefly
unreachable, and any process with an open watch/proxy connection to it
(`kubectl proxy`, `kubectl port-forward`, the operators) needs restarting
afterwards. Confirm `kubectl get nodes` reports `Ready` before moving on.

## Final checks

```bash
kubectl get pods -A
# keycloak-0, postgres, cert-manager, envoy-gateway-system pods all Running, no CrashLoopBackOff

curl -sk -o /dev/null -w "%{http_code}\n" https://keycloak.crownlabs.local/realms/crownlabs
# WSL2: https://keycloak.crownlabs.local:8443/realms/crownlabs
# expected: 200

kubectl get nodes
# expected: Ready — confirms the apiserver came back up after Step 2's restart
```

Expected end state:
- ✅ Keycloak up, reachable through the Gateway.
- ✅ Realm `crownlabs` imported; `k8s`/`operator-local` clients configured.
- ✅ `mydrive-pvcs` namespace present.
- ✅ Kubernetes apiserver authenticating via OIDC.

## Next

- [`../operators/README.md`](../operators/README.md) — base RBAC and running
  the CrownLabs operator (needs everything above).
- For qlkube (talking directly to the real apiserver instead of through
  `kubectl proxy`) and the frontend's OIDC configuration, see
  [`apiserver-oidc-integration.md`](apiserver-oidc-integration.md) Steps 4–6.

## Regenerating the realm export

If you manually change something in the local realm (new client, new
scope, new user, etc.) and want to freeze it for future installs, see
[`regenerating-the-realm-export.md`](regenerating-the-realm-export.md).
