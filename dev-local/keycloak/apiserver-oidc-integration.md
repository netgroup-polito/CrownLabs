# Authenticating the local Kubernetes API server with Keycloak (OIDC)

This document explains how to make the **local k3s API server itself** trust Keycloak-issued tokens, and enforce per-user RBAC.
This mirrors how a real CrownLabs cluster authenticates users.
Without this setup, every local request runs silently as cluster-admin.

This is a separate, optional layer on top of the basic local setup described in `dev-local/local-development.md` and [`README.md`](README.md).
Read those first.
This document assumes [Envoy Gateway](../envoy/README.md), Keycloak, qlkube, the frontend, and the operators are already working locally with `kubectl proxy` and admin credentials.

## Why this is needed

In the basic local setup, qlkube (in `IN_CLUSTER=false` mode) talks to the API server through `kubectl proxy` (`http://localhost:8001`).
**`kubectl proxy` always authenticates using its own kubeconfig credentials, and completely ignores whatever `Authorization` header the client sends.**
We confirmed this by sending a garbage bearer token through it, and getting back a successful, fully-authorized response anyway.
This means every request qlkube makes runs as cluster-admin, no matter which user is logged into the frontend.
Any RBAC rule configured by the operator has no real effect locally.

To actually exercise per-user RBAC locally, and catch RBAC bugs before they reach a real cluster, the API server needs to validate the user's own OIDC token directly.
qlkube also needs to forward that token to the API server, instead of routing through `kubectl proxy`.

## Overview of what changes

1. Two extra Keycloak protocol mappers are added: one exposes the user's client roles as a `groups` claim, and the other makes sure the `k8s` client is present in its own token's `aud` claim.
2. qlkube gets a new mode (`USE_LOCAL_CLUSTER=true`) that talks to the real API server directly, instead of going through `kubectl proxy`.
3. k3s's API server gets `--oidc-*` flags pointing at Keycloak's `HTTPRoute` hostname (`keycloak.crownlabs.local`, behind the shared Gateway). TLS is required for k3s to accept the issuer at all, and the Gateway terminates that TLS, not Keycloak itself.
4. The base CrownLabs `ClusterRole` and `ClusterRoleBinding` resources get applied, so RBAC actually has something to grant. The Helm chart normally installs these, but they are never applied when you run the operators directly with `go run`.

> **If you installed Keycloak from this folder's `manifests/`** (see [`README.md`](README.md)), step 1 below is already done: the `groups` and `k8s-audience` client scopes are part of the imported realm.
> Step 4 (base RBAC) lives in [`../operators/README.md`](../operators/README.md), step 1, instead, because it is really a prerequisite for running the operator, not for Keycloak itself.
> You can skip ahead to step 3.
> Steps 1 and 4 stay here in full, for reference and rationale.

## Step 1 — Keycloak realm configuration

### 1.1 `groups` claim from client roles

The CrownLabs operator binds `ClusterRoleBinding` and `RoleBinding` subjects of `Kind: Group`, with names like `kubernetes:workspace-tea:manager` (see `operators/pkg/forge/rbac.go` and `rolebinding.go`).
For the API server to resolve a user into these groups, the token must carry a `groups` claim that lists the user's raw Keycloak client roles.
The API server itself adds the `kubernetes:` prefix, through the `--oidc-groups-prefix` flag. Keycloak does not add it.

```bash
TOKEN=$(curl -sk -X POST https://keycloak.crownlabs.local/realms/master/protocol/openid-connect/token \
  -d "client_id=admin-cli" -d "username=admin" -d "password=admin" -d "grant_type=password" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['access_token'])")

curl -sk -X POST "https://keycloak.crownlabs.local/admin/realms/crownlabs/client-scopes" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"name":"groups","protocol":"openid-connect","attributes":{"include.in.token.scope":"true","display.on.consent.screen":"false"}}'

GROUPS_SCOPE_ID=$(curl -sk "https://keycloak.crownlabs.local/admin/realms/crownlabs/client-scopes" \
  -H "Authorization: Bearer $TOKEN" | python3 -c "import json,sys; print([s['id'] for s in json.load(sys.stdin) if s['name']=='groups'][0])")

# Note: "usermodel.clientRoleMapping.clientId" must be the client that actually holds the
# workspace-*:role client roles. This is whatever you passed as --keycloak-roles-client-id
# to the operator ("operator-local" in this local setup).
curl -sk -X POST "https://keycloak.crownlabs.local/admin/realms/crownlabs/client-scopes/$GROUPS_SCOPE_ID/protocol-mappers/models" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{
    "name": "operator-local-client-roles-to-groups",
    "protocol": "openid-connect",
    "protocolMapper": "oidc-usermodel-client-role-mapper",
    "config": {
      "usermodel.clientRoleMapping.clientId": "operator-local",
      "claim.name": "groups",
      "jsonType.label": "String",
      "multivalued": "true",
      "id.token.claim": "true",
      "access.token.claim": "true",
      "userinfo.token.claim": "true"
    }
  }'

K8S_UUID=$(curl -sk "https://keycloak.crownlabs.local/admin/realms/crownlabs/clients?clientId=k8s" \
  -H "Authorization: Bearer $TOKEN" | python3 -c "import json,sys; print(json.load(sys.stdin)[0]['id'])")

curl -sk -X PUT "https://keycloak.crownlabs.local/admin/realms/crownlabs/clients/$K8S_UUID/default-client-scopes/$GROUPS_SCOPE_ID" \
  -H "Authorization: Bearer $TOKEN"
```

### 1.2 `aud` (audience) fix

By default, Keycloak does **not** include a client's own `clientId` in the `aud` claim of the tokens it issues to that client.
Only clients that have a role-mapping pointing at them (like `operator-local`, added implicitly by the mapper above), or the built-in `account` client, end up in `aud`.
Kubernetes' OIDC authenticator requires `--oidc-client-id` to be present in the token's `aud`.
Without this fix, every request would fail with an audience mismatch, even though the token is otherwise perfectly valid.

```bash
curl -sk -X POST "https://keycloak.crownlabs.local/admin/realms/crownlabs/client-scopes" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"name":"k8s-audience","protocol":"openid-connect","attributes":{"include.in.token.scope":"true","display.on.consent.screen":"false"}}'

AUD_SCOPE_ID=$(curl -sk "https://keycloak.crownlabs.local/admin/realms/crownlabs/client-scopes" \
  -H "Authorization: Bearer $TOKEN" | python3 -c "import json,sys; print([s['id'] for s in json.load(sys.stdin) if s['name']=='k8s-audience'][0])")

curl -sk -X POST "https://keycloak.crownlabs.local/admin/realms/crownlabs/client-scopes/$AUD_SCOPE_ID/protocol-mappers/models" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{
    "name": "k8s-self-audience",
    "protocol": "openid-connect",
    "protocolMapper": "oidc-audience-mapper",
    "config": {"included.client.audience": "k8s", "id.token.claim": "false", "access.token.claim": "true"}
  }'

curl -sk -X PUT "https://keycloak.crownlabs.local/admin/realms/crownlabs/clients/$K8S_UUID/default-client-scopes/$AUD_SCOPE_ID" \
  -H "Authorization: Bearer $TOKEN"
```

### 1.3 Verify the token shape

Temporarily enable direct access grants on the `k8s` client, so you can pull a password-grant test token.
Remember to disable this again afterwards: it is off by default for a reason.

```bash
curl -sk -X PUT "https://keycloak.crownlabs.local/admin/realms/crownlabs/clients/$K8S_UUID" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"directAccessGrantsEnabled": true}'

JD_TOKEN=$(curl -sk -X POST https://keycloak.crownlabs.local/realms/crownlabs/protocol/openid-connect/token \
  -d "client_id=k8s" -d "username=john.doe" -d "password=johndoe123" -d "grant_type=password" -d "scope=openid profile email api groups" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['access_token'])")

echo "$JD_TOKEN" | cut -d. -f2 | python3 -c "
import sys, base64, json
s = sys.stdin.read().strip(); s += '=' * (-len(s) % 4)
p = json.loads(base64.urlsafe_b64decode(s))
print('iss:', p.get('iss')); print('aud:', p.get('aud'))
print('preferred_username:', p.get('preferred_username')); print('groups:', p.get('groups'))
"
# Expected: "iss" matches https://keycloak.crownlabs.local/realms/crownlabs exactly.
# On WSL2: https://keycloak.crownlabs.local:8443/realms/crownlabs.
# "aud" contains "k8s". "preferred_username" is the plain username. "groups" lists workspace-*:role entries.

curl -sk -X PUT "https://keycloak.crownlabs.local/admin/realms/crownlabs/clients/$K8S_UUID" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"directAccessGrantsEnabled": false}'
```

## Step 2 — Base CrownLabs RBAC resources (rationale)

The top-level Helm chart (`deploy/crownlabs/templates/clusterroles.yaml` and `clusterrolebindings.yaml`) defines the base `ClusterRole`s that the operator's per-tenant and per-workspace bindings reference (`crownlabs-manage-tenants`, `crownlabs-manage-instances`, `crownlabs-view-workspaces`, and others).
It also defines a couple of `ClusterRoleBinding`s to the built-in `system:authenticated` group, so any logged-in user can *view* workspaces and image lists. The enrollment and discovery UI needs this.
**None of this exists if you only ran the operators locally with `go run`.**
This gap stays invisible under `kubectl proxy`, which never checks RBAC, and only shows up once the API server starts doing real authorization.

Rather than hand-maintaining a de-templated copy of those two files, which can silently drift from the real chart, [`../operators/README.md`](../operators/README.md), step 1, renders them straight from `deploy/crownlabs` with `helm template -s templates/clusterroles.yaml -s templates/clusterrolebindings.yaml`, and pipes the result into `kubectl apply`.
There is nothing to run here beyond that: this section only explains why that step exists.
These RBAC definitions are purely additive: creating a `ClusterRole` grants nothing by itself, until something is bound to it. So applying them is always safe.

## Step 3 — Configure the k3s API server

Write a dedicated drop-in file under `/etc/rancher/k3s/config.yaml.d/` (this requires root), instead of editing the shared `/etc/rancher/k3s/config.yaml` file directly.
k3s automatically merges every `*.yaml` file there, so this cannot clobber anything that [`../base-k3s/README.md`](../base-k3s/README.md) (kubeconfig access) or [`../envoy/README.md`](../envoy/README.md) (disabling Traefik) already put there:

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

On WSL2, use `oidc-issuer-url=https://keycloak.crownlabs.local:8443/realms/crownlabs` instead (the Gateway bridge port, see [`../envoy/README.md`](../envoy/README.md) step 5, or `README.md` step 2).

Notes on each flag:
- `oidc-issuer-url` must match the token's `iss` claim **exactly** (scheme, host, port, path). It reflects whatever URL the browser or frontend uses to reach Keycloak, not an arbitrary internal address.
- `oidc-username-prefix=-` is required explicitly. Without it, some Kubernetes versions prefix the username with the issuer URL (`https://...#john.doe`), which will not match the plain `john.doe` used in the operator's `RoleBinding` and `ClusterRoleBinding` subjects.
- `oidc-ca-file` is required because the API server's own HTTP client does not trust a self-signed certificate by default. It needs an explicit copy of the certificate to fetch the OIDC discovery document and the JWKS. This is the Gateway's wildcard certificate (`crownlabs-tls`) now, not a Keycloak-specific one. See [`../envoy/README.md`](../envoy/README.md).

Extract the CA certificate from the `crownlabs-tls` Secret into a user-owned copy (cert-manager stores it under `ca.crt` there, see [`../envoy/README.md`](../envoy/README.md)).
You reuse this same copy below, for the operator's `SSL_CERT_FILE` too.
Then copy it where the root-owned API server configuration expects it, and restart:

```bash
mkdir -p ~/certs
kubectl get secret crownlabs-tls -n default -o jsonpath='{.data.ca\.crt}' | base64 -d > ~/certs/crownlabs-ca.crt

sudo cp ~/certs/crownlabs-ca.crt /etc/rancher/k3s/crownlabs-ca.crt
sudo systemctl restart k3s
journalctl -u k3s -f   # Confirm there is no "invalid authentication configuration" error, then press Ctrl+C.
```

**This restarts the whole control plane.**
Running pods survive, but the API server is briefly unreachable, and any process with an open watch or proxy connection to it (`kubectl proxy`, `kubectl port-forward`, the operators) needs restarting afterwards.
Confirm everything is healthy (`kubectl get nodes`) before you move on.

### If you need to roll this back

```bash
sudo rm /etc/rancher/k3s/config.yaml.d/30-keycloak-oidc.yaml
sudo systemctl restart k3s
```

## Step 4 — qlkube: talk to the real API server

`qlkube/src/index.js` now has a `USE_LOCAL_CLUSTER=true` mode that bypasses `kubectl proxy`, and connects directly to `https://localhost:6443` (or to `LOCAL_K8S_API_URL`, if you set it).
This mode forwards whatever bearer token it receives as-is, instead of silently discarding it.
It talks to the API server's own port directly, which is unrelated to the Gateway/Keycloak hostname used above.

This mode needs two things that `kubectl proxy` did not need:

**A trusted-but-self-signed TLS connection.** Set `NODE_TLS_REJECT_UNAUTHORIZED=0` when you start qlkube this way (this is already wired into the `dev-local-cluster` npm script). This is acceptable for local development only: it disables certificate verification for the whole Node process.

**A credential for the one-off OpenAPI schema discovery call at startup**, which happens before any user request exists.
`kubectl proxy` used its own kubeconfig for this, regardless of what you sent it.
Talking to the real API server means this call is genuinely unauthenticated, unless you provide something.
k3s's default kubeconfig uses **client-certificate authentication (mTLS)**, not a bearer token, so you cannot reuse it directly here.
Instead, create a small, dedicated `ServiceAccount`, scoped to exactly what discovery needs, and nothing more:

```bash
kubectl create serviceaccount qlkube-local -n default
kubectl create clusterrolebinding qlkube-local-discovery --clusterrole=system:discovery --serviceaccount=default:qlkube-local
```

Generate a long-lived token for it, and export it before starting qlkube. Never commit this token anywhere:

```bash
export LOCAL_K8S_BOOTSTRAP_TOKEN=$(kubectl create token qlkube-local -n default --duration=87600h)
cd qlkube
CROWNLABS_QLKUBE_PORT=8085 npm run dev-local-cluster
```

Verify:

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8085/graph/schema
# Expected: 200. The qlkube log should also show "OpenApi retrieved", with baseUrl https://localhost:6443.
```

## Step 5 — Frontend configuration

```bash
VITE_APP_CROWNLABS_OIDC_AUTHORITY=https://keycloak.crownlabs.local/realms/crownlabs
VITE_APP_CROWNLABS_OIDC_CLIENT_ID=k8s
VITE_APP_CROWNLABS_GRAPHQL_URL=http://localhost:8085/graph
```

On WSL2, use `https://keycloak.crownlabs.local:8443/realms/crownlabs` for `VITE_APP_CROWNLABS_OIDC_AUTHORITY` instead.

Restart the frontend dev server after changing `.env`. Vite does not hot-reload environment variables.

### Trusting the self-signed certificate in the browser

The frontend's OIDC library makes background `fetch()` calls (the discovery document, token exchange, silent renew) directly to `https://keycloak.crownlabs.local`.
A browser silently fails these calls with `ERR_CERT_AUTHORITY_INVALID`, for an untrusted self-signed certificate.
This shows up as an **infinite login redirect loop, with no visible error**, because the app cannot tell "the network call to refresh my session failed" apart from "the session is genuinely gone".
It reacts to both cases the same way, by forcing a fresh Keycloak login (see `AuthContextProvider.tsx`'s `execLogin` logic, and the troubleshooting table in `dev-local/local-development.md`).

Fix: visit `https://keycloak.crownlabs.local/realms/crownlabs/.well-known/openid-configuration` (on WSL2: `https://keycloak.crownlabs.local:8443/...`) directly, once, in the same browser.
Click through the "not private" warning ("Advanced" → "Proceed"), then reload the app.
The browser now trusts that origin for the rest of the session.

## Step 6 — Make sure every namespace tenants touch is a real, enrolled Workspace

Any namespace used for sample `Instance` or `Template` resources needs a matching `Workspace` custom resource, with the tenant actually enrolled in it.
Otherwise, the operator never creates the Keycloak `workspace-<name>:*` roles, or the `RoleBinding`s that grant access.
Any real per-user RBAC check on that namespace's `templates` or `instances` then fails with `403 Forbidden`.
This stays invisible under `kubectl proxy`, since the admin user can read anything, and only shows up once RBAC is actually enforced.

If you have ad-hoc namespaces (for example, a `workspace-standalone` namespace, created by hand for miscellaneous sample templates, with no corresponding `Workspace` CR), create the missing `Workspace` and enroll the tenant:

```bash
kubectl apply -f - <<'EOF'
apiVersion: crownlabs.polito.it/v1alpha1
kind: Workspace
metadata:
  name: standalone
  labels:
    crownlabs.polito.it/operator-selector: local
spec:
  prettyName: Standalone environments
  quota: {cpu: "8", memory: "15Gi", instances: 3}
EOF

kubectl patch tenant john.doe --type='json' \
  -p='[{"op": "add", "path": "/spec/workspaces/-", "value": {"name": "standalone", "role": "manager"}}]'
```

If the operator is running, it picks this up automatically, and creates the corresponding Keycloak roles within a few seconds. Verify with:

```bash
TOKEN=$(curl -sk -X POST https://keycloak.crownlabs.local/realms/master/protocol/openid-connect/token \
  -d "client_id=admin-cli" -d "username=admin" -d "password=admin" -d "grant_type=password" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['access_token'])")
OP_UUID=$(curl -sk "https://keycloak.crownlabs.local/admin/realms/crownlabs/clients?clientId=operator-local" \
  -H "Authorization: Bearer $TOKEN" | python3 -c "import json,sys; print(json.load(sys.stdin)[0]['id'])")
curl -sk "https://keycloak.crownlabs.local/admin/realms/crownlabs/clients/$OP_UUID/roles" -H "Authorization: Bearer $TOKEN" \
  | python3 -c "import json,sys; print([r['name'] for r in json.load(sys.stdin) if 'standalone' in r['name']])"
```

**The frontend's already-issued token will not reflect the new role until it refreshes.**
Log out and back in (or wait for the roughly 5-minute silent-refresh cycle) after you enroll a tenant in a new workspace.

## Restarting the operator against Keycloak

The operator's Keycloak client (`gocloak`, a plain Go `http.Client` under the hood) also needs to trust the self-signed certificate presented at the Gateway.
Go's `crypto/x509` respects the `SSL_CERT_FILE` environment variable on Linux, for loading extra trusted roots.
Point it at the same `~/certs/crownlabs-ca.crt` file you extracted in step 3, then use the `run-operator-local` Makefile target (see [`../operators/README.md`](../operators/README.md), step 2):

```bash
export SSL_CERT_FILE=~/certs/crownlabs-ca.crt
cd operators
make run-operator-local KEYCLOAK_CLIENT_SECRET=operator-local-dev-secret
```

On WSL2, also override `KEYCLOAK_URL` to the bridge port: `make run-operator-local KEYCLOAK_URL=https://keycloak.crownlabs.local:8443 KEYCLOAK_CLIENT_SECRET=operator-local-dev-secret`.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| k3s fails to start: `jwt[0].issuer.url: ... URL scheme must be https` | k3s 1.34+ hard-rejects non-HTTPS OIDC issuers | Confirm the Gateway's HTTPS listener is up ([`../envoy/README.md`](../envoy/README.md)). TLS lives there now, not on Keycloak itself |
| `403 Forbidden` from the API server, even after `--oidc-*` flags are set | Base CrownLabs `ClusterRole`s not installed (never applied outside Helm) | Render and apply them with `helm template` — [`../operators/README.md`](../operators/README.md) step 1 (rationale in step 2 of this doc) |
| Token rejected, audience mismatch | The `k8s` client's own `clientId` is missing from its token's `aud` (Keycloak does not add it by default) | Add the audience mapper (step 1.2) |
| Token rejected, issuer mismatch (WSL2 only) | Keycloak (`KC_HOSTNAME_STRICT=false`) reports its issuer **without** a port, no matter which port the request came in on. Envoy Gateway does not forward `X-Forwarded-Port`, so this does not byte-for-byte match an `oidc-issuer-url` ending in `:8443` | Pin `KC_HOSTNAME=https://keycloak.crownlabs.local:8443` explicitly, see [`README.md`](README.md) step 1 |
| API server cannot validate the OIDC discovery document | Self-signed certificate not trusted by the API server's HTTP client | Add `--oidc-ca-file`, pointing at the Gateway's wildcard CA (step 3) |
| Frontend stuck in an infinite login redirect loop, no visible error | Browser does not trust the self-signed certificate. Background OIDC calls fail silently, and the app treats this as a stale session, forcing a re-login | Visit the Keycloak URL directly once, and accept the browser warning (step 5) |
| qlkube: `connect ECONNREFUSED` on the OpenAPI discovery call | Direct-API-server mode has no bootstrap credential (unlike `kubectl proxy`, which never needed one) | Set `LOCAL_K8S_BOOTSTRAP_TOKEN` (step 4) |
| qlkube: `SELF_SIGNED_CERT_IN_CHAIN` talking to `localhost:6443` | Node's default TLS verification rejects the API server's certificate | Set `NODE_TLS_REJECT_UNAUTHORIZED=0` (already in the `dev-local-cluster` script) |
| Operator cannot reach Keycloak (`x509: certificate signed by unknown authority`) | Go's HTTP client does not trust the self-signed certificate either | `export SSL_CERT_FILE=<path-to-certificate>` before starting the operator |
| `403 Forbidden` reading `templates` or `instances` in a specific `workspace-*` namespace | The tenant was never actually enrolled in a matching `Workspace` CR (ad-hoc namespace, no membership) | Create the `Workspace` and enroll the tenant (step 6); log out and back in afterwards, for a fresh token |
| `403 Forbidden` on a namespace or role that worked before, after running `make run-operator-local` or `run-instance-local` again | `operators/samples/` got re-applied as a target prerequisite. `kubectl apply`'s 3-way merge silently reverted a `kubectl patch`-only change (for example, a `Workspace` enrollment added per step 6) that was never reflected in the checked-in sample file | Re-apply the `kubectl patch`; log out and back in, for a fresh token. See the "Gotcha" note in `../operators/README.md`, step 2 |
| `bind: address already in use` restarting the operator or qlkube | The actual compiled Go or Node binary survived independently of the `go run` or `npm run` wrapper process that appeared to exit (for example, across a k3s restart) | Find the real PID with `ss -tlnp \| grep <port>` (the process name is the compiled binary, for example `imagelist`, not `go run`), then `kill -9` it directly |
| A `kubectl port-forward`, operator, or qlkube process dies right after `systemctl restart k3s` | Expected: restarting k3s restarts the whole control plane, dropping any open connection to it | Restart the affected process once `kubectl get nodes` reports `Ready` again |
| Browser: `ERR_CONNECTION_REFUSED` on `https://keycloak.crownlabs.local/...` (WSL2 only) | The Gateway's `LoadBalancer` Service has no real listening socket. WSL2's `localhostForwarding` only relays real sockets | Bridge with `kubectl port-forward -n envoy-gateway-system svc/<gateway-service> 8443:443`, and append `:8443` everywhere. See [`../envoy/README.md`](../envoy/README.md) step 5 |
| Operator: `dial tcp 127.0.0.1:XXXXX: i/o timeout` reaching Keycloak (WSL2 only) | The port-forward bridge reused the *same* port number as the Gateway's own 80/443. kube-proxy's own `iptables` rules for those ports hijack local loopback traffic before it reaches the port-forward's listener | Use a bridge port different from 80/443 (`:8443`, not `:443`) |
