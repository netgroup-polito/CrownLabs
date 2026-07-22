# Local development with Keycloak

This guide explains how to bring up the whole CrownLabs stack locally (Keycloak, qlkube, frontend, operators) against an already-existing k3s cluster. Follow the steps in order: each component depends on the previous one.

## Prerequisites

- The base k3s cluster, Envoy Gateway, Keycloak, and the CrownLabs operator
  already set up — see [`base-k3s/README.md`](base-k3s/README.md),
  [`envoy/README.md`](envoy/README.md), [`keycloak/README.md`](keycloak/README.md)
  and [`operators/README.md`](operators/README.md), in that order.
- Node.js (version pinned in `frontend/.nvmrc` and `qlkube/.nvmrc`).

The CrownLabs frontend dashboard itself is **not** exposed through Envoy Gateway — it
only ever runs as a local `npm start` dev server on `localhost:3000`, same as qlkube.
The Gateway is for services that need a stable, browser-reachable hostname without a
dev server in front of them (Keycloak now, [Mailpit](mailpit/README.md) later).

## 1. Base k3s cluster, Envoy Gateway, Keycloak, and the operator

Covered on their own, in order:

1. [`base-k3s/README.md`](base-k3s/README.md) — installing k3s and getting `kubectl` access.
2. [`envoy/README.md`](envoy/README.md) — the single Envoy Gateway every locally-exposed
   service (Keycloak now) is reachable through, at `https://<service>.crownlabs.local`.
3. [`keycloak/README.md`](keycloak/README.md) — Keycloak (realm, clients, scopes, base
   users) and, if you want the Kubernetes apiserver to enforce real per-user RBAC
   instead of everything running as cluster-admin under `kubectl proxy`, the OIDC
   integration with k3s.
4. [`operators/README.md`](operators/README.md) — base RBAC and running the CrownLabs main operator.

Follow those first; come back here once the main operator is up and reconciling.

## 2. qlkube

qlkube can talk to the cluster in two ways:

- **`kubectl proxy`** (`IN_CLUSTER=false`, plain `npm run dev`) — the simplest option, but `kubectl proxy` always authenticates with its own kubeconfig credentials and ignores whatever bearer token the client sends, so every request runs as cluster-admin regardless of who is logged into the frontend. Fine for quick UI iteration when you don't care about RBAC.
- **Direct apiserver connection** (`USE_LOCAL_CLUSTER=true`, `npm run dev-local-cluster`) — bypasses `kubectl proxy` and talks straight to the real apiserver, forwarding the caller's token as-is so the apiserver's own OIDC authenticator enforces per-user RBAC. Requires the k3s OIDC integration from [`keycloak/apiserver-oidc-integration.md`](keycloak/apiserver-oidc-integration.md) to be set up first, plus a bootstrap token for the one-off startup schema discovery call (see that same doc, Step 4).

Either way, qlkube itself is unaffected by the move to Envoy Gateway — it keeps running
locally and keeps talking to the apiserver's own port (`localhost:6443`) or through
`kubectl proxy`, neither of which goes through the Gateway.

Option A — `kubectl proxy`:

```bash
kubectl proxy --port=8001
```

```bash
cd qlkube
npm install
IN_CLUSTER=false CROWNLABS_QLKUBE_PORT=8085 npm run dev
```

Option B — direct apiserver connection:

```bash
export LOCAL_K8S_BOOTSTRAP_TOKEN=$(kubectl create token qlkube-local -n default --duration=87600h)  # see apiserver-oidc-integration.md Step 5 for the qlkube-local ServiceAccount setup
cd qlkube
npm install
CROWNLABS_QLKUBE_PORT=8085 npm run dev-local-cluster
```

Either way, the GraphQL playground/endpoint is at `http://localhost:8085/graph` (not the root — the path is configurable via `BASE_URL`).

Verify:

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8085/graph/schema
# expected: 200
```

### qlkube's local configuration files

`qlkube/src/wrappers.js` and `qlkube/src/subscriptions.js` are **development** files: in production they're overwritten by the configuration Helm generates. If you get GraphQL errors like `Cannot query field "workspaceWrapperTenantV1alpha2"` or `Cannot query field "itPolitoCrownlabsV1alpha2TenantUpdate" on type "Subscription"` during local development, it means these files don't include every wrapper/subscription the frontend needs. Check the `*Wrapper*` fields referenced in `frontend/src/graphql-components/**` and make sure every referenced type (e.g. `TenantCrownlabsPolitoItTenantRef`, `WorkspacesListItem`) has a matching entry in `wrappers.js`, and every resource needed as a subscription (e.g. `tenants`) is listed in `subscriptions.js`.

## 3. Frontend

```bash
cd frontend
cp .env.example .env
```

Edit `frontend/.env` (never committed, it's in `.gitignore`) with the local endpoints:

```bash
VITE_APP_CROWNLABS_OIDC_AUTHORITY=https://keycloak.crownlabs.local/realms/crownlabs
VITE_APP_CROWNLABS_OIDC_CLIENT_ID=k8s
VITE_APP_CROWNLABS_GRAPHQL_URL=http://localhost:8085/graph
```

On WSL2, use `https://keycloak.crownlabs.local:8443/realms/crownlabs` for
`VITE_APP_CROWNLABS_OIDC_AUTHORITY` instead — the Gateway bridge port from
[`envoy/README.md`](envoy/README.md) step 5. This must exactly match whatever URL the
browser actually reaches Keycloak through — see
[`keycloak/apiserver-oidc-integration.md`](keycloak/apiserver-oidc-integration.md) Step 5.

The other variables (`VITE_APP_MYDRIVE_*`, `VITE_APP_CROWNLABS_IMAGELIST_*`, etc.) can stay as they are in `.env.example`.

```bash
npm install
npm start
```

The frontend is on `http://localhost:3000`.

> **Don't edit `frontend/.env.example`** with local values: it's the shared template committed to the repo (used in production/by other developers). Local customizations only go in `.env`.

### If you change the GraphQL schema (wrappers.js, subscriptions.js, or frontend queries)

Regenerate the TypeScript types from qlkube's live schema:

```bash
cd frontend
GRAPHQL_URL=http://localhost:8085/graph npm run generate
```

## 4. instance-operator

The CrownLabs **main** operator (`Tenant`/`Workspace`/Keycloak roles) is
already up at this point — it's a prerequisite covered in
[`operators/README.md`](operators/README.md) (section 1 above). The
separate **instance-operator** (manages `Instance` resources: deployment,
VM, PVC, exposure) is a `local-development.md`-only step:

```bash
cd operators
export KUBECONFIG=~/.kube/config
make run-instance-local
```

This target also applies the CRDs and the samples (`operators/samples/`) before starting the operator.

## 5. If the machine is low on RAM

`keycloak.yaml` already ships with a single replica by default (high availability isn't needed for solo local development). If you scaled it up to test Keycloak's HA/Infinispan clustering and want to scale back down to free up RAM — a local cluster with Keycloak HA + Postgres + several test instances can easily saturate a k3s single-node's memory (e.g. ~8 GiB under WSL2):

```bash
kubectl scale statefulset keycloak --replicas=1
```

The pod with the lowest index (`keycloak-0`) is always the last one to go when scaling a StatefulSet down, so `https://keycloak.crownlabs.local` (routed through the Gateway to the `keycloak` Service, not tied to any specific pod) keeps working unchanged before, during and after the scale-down.

## 6. Login

Open `http://localhost:3000` and log in with the test user's credentials (see [`keycloak/README.md`](keycloak/README.md) for the default `john.doe` / `johndoe123`).

## Port summary

| Service | Local port | Command |
|---|---|---|
| Envoy Gateway (everything behind `*.crownlabs.local`, incl. Keycloak) | 80/443 (`:8443` on WSL2, bridged) | already exposed (`ServiceLB`, see [`envoy/README.md`](envoy/README.md)); WSL2: `kubectl port-forward -n envoy-gateway-system svc/<gateway-service> 8443:443` |
| kubectl proxy (apiserver, option A) | 8001 | `kubectl proxy --port=8001` |
| qlkube (GraphQL) | 8085, path `/graph` | `CROWNLABS_QLKUBE_PORT=8085 npm run dev` (or `dev-local-cluster`) |
| Frontend | 3000 | `npm start` |
| instance-operator | 8080/8081 (default) | `make run-instance-local` |
| main operator | 18080/18081/18082 | `make run-operator-local` |

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `invalid_scope: Invalid scopes: openid profile email api` at login | Client scope `api` missing from the realm | Should already be present if you installed from `keycloak/manifests/`; otherwise create it (Admin Console → Client Scopes) and add it as a default scope on the `k8s` client |
| `Cannot query field "workspaceWrapperTenantV1alpha2"` / `"tenantV1alpha2Wrapper"` | Missing wrapper in `qlkube/src/wrappers.js` | See section 2, "qlkube's local configuration files" |
| `Cannot query field "itPolitoCrownlabsV1alpha2TenantUpdate" on type "Subscription"` | Missing `tenants` resource in `qlkube/src/subscriptions.js` | See section 2 |
| Login OK but empty dashboard, tenant always "not ready" | Keycloak user not verified (`emailVerified: false`) and/or main operator not running | Verify the user in Keycloak, run the main operator (`operators/README.md`) |
| `bind: address already in use` restarting a port-forward/operator | A previous process (often the compiled Go binary, not the wrapper script) is still alive | `ss -tlnp \| grep <port>` to find the real PID, then `kill -9 <pid>` |
| Refresh token rejected (`400` on `/protocol/openid-connect/token`, Keycloak log `"Token is not active"`) | With more than 1 Keycloak replica, the Service can route a request to a different pod than the one that issued the token (Infinispan session cache isn't perfectly synced) | Scale down to 1 replica (section 5, and the default) |
| `PersistentVolumeClaim ... spec.resources[storage]: Invalid value: "0"` | Template with `persistent: true` but no `disk` set | Set a `disk` value in the template |
| `Service "..." is invalid: spec.ports: Required value` | `Container`/`Standalone` environment with `guiEnabled: false` (bug fixed in `operators/pkg/forge/services.go`) | Confirm the fix is present in the branch you're using |
| Browser: `ERR_CONNECTION_REFUSED` on `https://keycloak.crownlabs.local/...` (WSL2 only) | The Gateway's `LoadBalancer` Service has no real listening socket; WSL2's `localhostForwarding` only relays real sockets — tested and confirmed true in both `networkingMode` settings (`NAT` and `mirrored`), switching to mirrored does **not** fix it | `kubectl port-forward -n envoy-gateway-system svc/<gateway-service> 8443:443` (a *different* port than 80/443 themselves — reusing them gets hijacked by kube-proxy's own iptables rules and times out) — see `envoy/README.md` step 5 |
