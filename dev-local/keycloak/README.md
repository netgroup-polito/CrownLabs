# Local Keycloak + OIDC — end-to-end setup

This section presents the steps needed to bring up Keycloak locally: the realm, the clients, the scopes, and the base users.
It also configures the Kubernetes API server to validate the tokens Keycloak issues.
Without this, every local request through `kubectl proxy` runs silently as cluster-admin, with no real authentication.
You need an already-existing k3s cluster with [Envoy Gateway](../envoy/README.md) set up before you start.

- [`manifests/`](manifests/): apply everything with a single command, `kubectl apply -f dev-local/keycloak/manifests`. This includes Keycloak, Postgres, its `HTTPRoute`, the realm import, and the `mydrive-pvcs` namespace.
- [`apiserver-oidc-integration.md`](apiserver-oidc-integration.md): the rationale, verification steps, and troubleshooting table behind the OIDC setup below. Read it if something in this guide does not behave as expected. This README only lists the commands.

## 0. Prerequisites

- The base k3s cluster. See [`../base-k3s/README.md`](../base-k3s/README.md).
- [Envoy Gateway](../envoy/README.md) set up: the Gateway API CRDs, the controller, the shared `crownlabs` Gateway, and the `*.crownlabs.local` wildcard certificate.
  Keycloak's own manifest no longer has its own HTTPS listener or NodePort.
  It is exposed entirely through the `HTTPRoute` in `manifests/keycloak.yaml`, which needs that `Gateway` to already exist.

```bash
kubectl get nodes
# Expected: your node (the machine running k3s) is Ready.
kubectl get gateway crownlabs -n default
# Expected: PROGRAMMED = True.
```

## 1. Keycloak and the `mydrive-pvcs` namespace

```bash
kubectl apply -f dev-local/keycloak/manifests
kubectl rollout status statefulset/keycloak
```

This single apply covers:
- Keycloak and Postgres (`manifests/keycloak.yaml`), with `--import-realm` enabled and the
  realm ConfigMap mounted. Keycloak itself only speaks plain HTTP (in fact, TLS terminates
  at the Gateway, as described in [`../envoy/README.md`](../envoy/README.md)), so there's
  no certificate or HTTPS configuration on Keycloak's own side at all.
- An `HTTPRoute` (in the same file) that routes `keycloak.crownlabs.local` to that Service, through the shared `crownlabs` Gateway.
- The realm import itself (`manifests/crownlabs-realm-configmap.yaml`): the `crownlabs` realm export, with the `k8s` and `operator-local` clients, the `groups` and `k8s-audience` scopes, and the base users. It is wrapped in a `ConfigMap`, so you apply it declaratively instead of running an easy-to-forget, manual `kubectl create configmap` command.
- The `mydrive-pvcs` namespace (`manifests/mydrive-pvcs-namespace.yaml`), used by the operator for tenants' personal-drive PVCs.

Keycloak is now reachable at `https://keycloak.crownlabs.local`.
You do not need `kubectl port-forward` on native Linux.
See [`../envoy/README.md`](../envoy/README.md) for how this works: one shared Gateway, backed by `ServiceLB` on ports 80 and 443, instead of a separate `NodePort` per service.

> **On WSL2**: this has the same underlying limitation as everything else behind the Gateway.
> See [`../envoy/README.md`](../envoy/README.md) step 5 for the one bridge you need (`kubectl port-forward ... 8443:443`).
> **Append `:8443` to every hostname in this guide and in `apiserver-oidc-integration.md` when you use WSL2** (for example, `https://keycloak.crownlabs.local:8443`).
> This must exactly match the URL the browser actually uses to reach Keycloak, because that URL ends up in the token's `iss` (issuer) claim.
>
> **Also on WSL2**: Keycloak (with `KC_HOSTNAME_STRICT=false`) infers its own issuer from the request it receives.
> We tested this: Keycloak always reports the issuer **without** a port, no matter which port the request actually arrived on, because Envoy Gateway does not forward an `X-Forwarded-Port` header.
> On native Linux, this is fine, since port 443 is the default and there is no bridge involved.
> On WSL2, it is not fine: the issuer needs to say `:8443` explicitly, or it will not match, byte-for-byte, the `oidc-issuer-url` you configure for the API server in step 2 below. If it does not match, every token gets rejected with an issuer mismatch error.
> Pin the hostname explicitly after you apply the manifests:
> ```bash
> kubectl set env statefulset/keycloak KC_HOSTNAME=https://keycloak.crownlabs.local:8443
> kubectl rollout status statefulset/keycloak
> ```
> This is a live patch. It is not part of `manifests/keycloak.yaml`, which stays environment-agnostic.
> Re-run this patch if you ever re-apply the base manifest on WSL2. This is the same gotcha as the `Tenant`/`Workspace` patches described in `../operators/README.md`.

The self-signed wildcard certificate (issued at the Gateway, see [`../envoy/README.md`](../envoy/README.md)) triggers a browser warning (`ERR_CERT_AUTHORITY_INVALID`) the first time you visit `https://keycloak.crownlabs.local/...`.
This warning is expected for local development.
Click through it once ("Advanced" → "Proceed"), and the browser trusts the certificate for the rest of the session.

On first startup, Keycloak automatically imports the realm from the mounted ConfigMap.
You will see this in the logs: `KC-SERVICES0032: Import JSON RealmRepresentation from file .../crownlabs-realm.json`.
**If the realm already exists in the database, Keycloak silently skips the import.**
This is the standard `--import-realm` behavior, meant to avoid losing state across restarts.
To re-import from scratch, first delete the realm (through the Admin Console, or with `DELETE /admin/realms/crownlabs`), or wipe the Postgres database.

Credentials included in the export:

| Who | Credentials |
|---|---|
| Keycloak admin (`master` realm) | `admin` / `admin` (set through an environment variable, not part of the export) |
| Test user (`crownlabs` realm) | `john.doe` / `johndoe123`, email verified |
| `operator-local` client (service account for the operator) | client secret: `operator-local-dev-secret` |
| `k8s` client (frontend, public) | no secret; redirect URI already set to `http://localhost:3000/*` |

> These are **known values, deliberately documented for local development**, exactly like `admin`/`admin` in the manifest. Do not reuse them in real environments.

## 2. Configure the k3s API server for OIDC

This step requires root privileges and a k3s restart, so it cannot be folded into the manifests above.
See [`apiserver-oidc-integration.md`](apiserver-oidc-integration.md) for the full rationale: why each flag is needed, the `groups`/`aud` claim mappers, and so on.
Here are just the commands:

```bash
mkdir -p ~/certs
kubectl get secret crownlabs-tls -n default -o jsonpath='{.data.ca\.crt}' | base64 -d > ~/certs/crownlabs-ca.crt
sudo cp ~/certs/crownlabs-ca.crt /etc/rancher/k3s/crownlabs-ca.crt
```

Write a dedicated drop-in file (this requires root) instead of editing the shared `/etc/rancher/k3s/config.yaml` file directly.
k3s automatically merges every `*.yaml` file under `/etc/rancher/k3s/config.yaml.d/`, so this new file cannot clobber anything that [`../base-k3s/README.md`](../base-k3s/README.md) or [`../envoy/README.md`](../envoy/README.md) already added there:

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

> **On WSL2**: use `oidc-issuer-url=https://keycloak.crownlabs.local:8443/realms/crownlabs` instead (the Gateway bridge port from [`../envoy/README.md`](../envoy/README.md) step 5).
> This must exactly match the URL the browser actually uses to reach Keycloak, because that URL ends up in the token's `iss` (issuer) claim.

```bash
sudo systemctl restart k3s
journalctl -u k3s -f   # Confirm there is no "invalid authentication configuration" error, then press Ctrl+C.
```

Restarting k3s restarts the whole control plane.
The API server is briefly unreachable, and you need to restart any process with an open watch or proxy connection to it (`kubectl proxy`, `kubectl port-forward`, the operators) afterwards.
Confirm `kubectl get nodes` reports `Ready` before you move on.

## Final checks

```bash
kubectl get pods -A
# Expected: keycloak-0, postgres, cert-manager, and envoy-gateway-system pods are all Running, with no CrashLoopBackOff.

curl -sk -o /dev/null -w "%{http_code}\n" https://keycloak.crownlabs.local/realms/crownlabs
# On WSL2: https://keycloak.crownlabs.local:8443/realms/crownlabs
# Expected: 200.

kubectl get nodes
# Expected: Ready. This confirms the API server came back up after the restart in step 2.
```

Expected end state:
- ✅ Keycloak is up, and reachable through the Gateway.
- ✅ The `crownlabs` realm is imported, and the `k8s`/`operator-local` clients are configured.
- ✅ The `mydrive-pvcs` namespace exists.
- ✅ The Kubernetes API server authenticates through OIDC.

## Next

- [`../operators/README.md`](../operators/README.md): set up base RBAC and run the CrownLabs operator. This needs everything above.
- For qlkube (which talks directly to the real API server instead of through `kubectl proxy`) and for the frontend's OIDC configuration, see steps 4 to 6 in [`apiserver-oidc-integration.md`](apiserver-oidc-integration.md).

## Regenerating the realm export

If you manually change something in the local realm (a new client, a new scope, a new user, and so on) and you want to freeze it for future installs, see [`regenerating-the-realm-export.md`](regenerating-the-realm-export.md).
