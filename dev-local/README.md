# Local deployment support files

General place for support files/manifests needed to run pieces of CrownLabs'
infrastructure locally, so they don't get scattered across ad-hoc locations
or left undocumented in someone's shell history.

## Structure

```
dev-local/<system>/manifests/<yaml-files>
dev-local/<system>/README.md
```

Each `<system>` is a self-contained piece of local infrastructure (e.g.
`keycloak`). Its `README.md` documents prerequisites and how/where/what to
deploy; its `manifests/` folder holds the plain Kubernetes YAML to
`kubectl apply`.

Manifests in here are meant to be applied as-is with `kubectl apply -f`. If
a resource is already templated in the main `deploy/` Helm chart, prefer
rendering it on demand from that chart (`helm template -s ...`) instead of
keeping a hand-maintained, de-templated copy here that can drift from the
original — see `operators/README.md` step 1 for an example.

## Systems

Set up in this order — each one is a prerequisite of the next:

- [`base-k3s/`](base-k3s/README.md) — installing k3s itself and getting
  `kubectl` access (including how to merge its kubeconfig into one that
  already has other clusters).
- [`envoy/`](envoy/README.md) — the single Envoy Gateway (Kubernetes Gateway
  API) every locally-exposed service is reachable through, at
  `https://<service>.crownlabs.local`. Do this before `keycloak/`.
- [`keycloak/`](keycloak/README.md) — local Keycloak (realm import) and the
  Kubernetes apiserver OIDC integration, exposed via an `HTTPRoute` through
  `envoy/`.
- [`operators/`](operators/README.md) — base RBAC and running the CrownLabs
  operator against the local realm.
- [`mailpit/`](mailpit/README.md) — fake SMTP + web UI for local email testing,
  exposed via an `HTTPRoute` through `envoy/` like Keycloak.

## Running the full stack

- [`local-development.md`](local-development.md) — end-to-end guide to bring
  up Keycloak, qlkube, the frontend and the operators locally, once the
  cluster itself exists.
