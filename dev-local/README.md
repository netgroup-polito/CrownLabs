# Local deployment support files

This folder holds the support files and manifests you need to run pieces of the CrownLabs infrastructure locally.
The goal is to keep everything in one place, instead of scattered across ad-hoc locations or lost in someone's shell history.

## Structure

```
dev-local/<system>/manifests/<yaml-files>
dev-local/<system>/README.md
```

Each `<system>` is a self-contained piece of local infrastructure, for example `keycloak`.
Its `README.md` file lists the prerequisites and explains how, where, and what to deploy.
Its `manifests/` folder holds the plain Kubernetes YAML files you apply with `kubectl apply`.

You can apply the manifests in this folder as-is, with `kubectl apply -f`.
If a resource is already templated in the main `deploy/` Helm chart, render it from that chart instead (`helm template -s ...`).
Do not keep a hand-maintained, de-templated copy here: it can drift from the original.
See step 1 in `operators/README.md` for an example.

## Systems

Set these up in order. Each one is a prerequisite for the next.

- [`base-k3s/`](base-k3s/README.md): install k3s and get `kubectl` access. This also explains how to merge k3s's kubeconfig into one that already has other clusters.
- [`envoy/`](envoy/README.md): the single Envoy Gateway (Kubernetes Gateway API) that every locally-exposed service goes through, at `https://<service>.crownlabs.local`. Do this before `keycloak/`.
- [`keycloak/`](keycloak/README.md): the local Keycloak setup (realm import) and the Kubernetes API server OIDC integration. Keycloak is exposed through an `HTTPRoute` on the Gateway from `envoy/`.
- [`operators/`](operators/README.md): the base RBAC setup and how to run the CrownLabs operator against the local realm.
- [`mailpit/`](mailpit/README.md): a fake SMTP server with a web UI, for local email testing. Like Keycloak, it is exposed through an `HTTPRoute` on the Gateway from `envoy/`.

## Running the full stack

See [`local-development.md`](local-development.md) for the end-to-end guide.
It shows how to bring up Keycloak, qlkube, the frontend, and the operators locally, once the cluster itself exists.
