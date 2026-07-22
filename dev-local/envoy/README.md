# Envoy Gateway — single ingress point for dev-local

A single [Envoy Gateway](https://gateway.envoyproxy.io/) (Kubernetes Gateway API)
fronting every locally-exposed dev-local service under one local domain
(`*.crownlabs.local`), replacing the old per-service `NodePort` pattern (used by
Keycloak until now). Do this **before** [`../keycloak/README.md`](../keycloak/README.md)
— Keycloak's manifest now includes an `HTTPRoute` that depends on the `Gateway`
created here.

TLS terminates here, at the Gateway edge, with one self-signed wildcard
certificate. Every backend behind it (Keycloak, [Mailpit](../mailpit/README.md))
is reached over plain HTTP internally — nothing else needs its own
certificate or HTTPS listener.

## 0. Prerequisites

- The base k3s cluster — see [`../base-k3s/README.md`](../base-k3s/README.md).
- cert-manager, for the wildcard certificate below:
  ```bash
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.20.3/cert-manager.yaml
  kubectl wait --for=condition=Available deployment --all -n cert-manager --timeout=120s
  ```

## 1. Disable k3s's bundled Traefik

k3s ships Traefik as its default ingress controller, with its own `LoadBalancer`
`Service` bound to the node's 80/443 via k3s's `ServiceLB` (`svclb-traefik`). Envoy
Gateway's own `Service` needs those same ports, so Traefik has to go first — this repo
uses Envoy Gateway instead of Traefik's own (more limited) Gateway API support.

Add a dedicated drop-in file under `/etc/rancher/k3s/config.yaml.d/` (requires
root) — k3s merges every `*.yaml` there into its config automatically, in
alphabetical order, so this doesn't touch whatever
[`../base-k3s/README.md`](../base-k3s/README.md) or
[`../keycloak/apiserver-oidc-integration.md`](../keycloak/apiserver-oidc-integration.md)
already put there, and vice versa — no shared file to merge into by hand:

```bash
sudo mkdir -p /etc/rancher/k3s/config.yaml.d
sudo tee /etc/rancher/k3s/config.yaml.d/20-disable-traefik.yaml > /dev/null <<'EOF'
disable:
  - traefik
EOF
```

```bash
sudo systemctl restart k3s
journalctl -u k3s -f   # confirm no errors, then Ctrl+C
kubectl get nodes      # expected: Ready
```

**This restarts the whole control plane** — same caveat as the apiserver OIDC step:
briefly unreachable, and any open `kubectl proxy`/`port-forward`/operator connection
needs restarting afterwards.

Confirm Traefik is gone:

```bash
kubectl get pods -n kube-system | grep -i traefik
# expected: no traefik/svclb-traefik pods (only the completed helm-install-traefik*
# jobs may still be listed — harmless, they're one-shot Jobs, not live components)
```

If disabling `traefik` on an already-running cluster doesn't clean up the leftover
`traefik`/`svclb-traefik` Deployment/DaemonSet on your k3s version, remove them
directly (this does **not** touch the Gateway API CRDs — those come from the separate
`traefik-crd` HelmChart, which `--disable=traefik` does not affect, and which Envoy
Gateway's own CRDs coexist with/upgrade in step 2 below):

```bash
kubectl delete deployment traefik -n kube-system
kubectl delete svc traefik -n kube-system
```

## 2. Install the Gateway API CRDs + Envoy Gateway controller

Pin an explicit release (not `latest`) — the `latest` *tag* for the install manifest and
the `latest` *image* tag can drift out of sync with each other (tested: this produced a
controller crash — `no matches for kind "ListenerSet" in version
"gateway.networking.k8s.io/v1"` — because the CRD bundle was older than what the
controller image expected):

```bash
kubectl apply --server-side --force-conflicts -f https://github.com/envoyproxy/gateway/releases/download/v1.8.2/install.yaml
kubectl wait --timeout=120s -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

`--force-conflicts` is required here: k3s's bundled `traefik-crd` chart already installs
the core Gateway API CRDs (`GatewayClass`/`Gateway`/`HTTPRoute`/...) at an older schema
version, field-managed by `helm`; this takes over that field ownership so Envoy Gateway
can upgrade them. Safe to do even before Traefik itself is disabled — Helm never deletes
CRDs on chart uninstall, so this doesn't put those CRDs at risk once `traefik-crd`'s
release (if ever) goes away.

> If you hit `The CustomResourceDefinition "backendtlspolicies.gateway.networking.k8s.io"
> is invalid: status.storedVersions[0]: ... must remain in spec.versions` — confirm no
> `BackendTLSPolicy` objects exist yet (`kubectl get backendtlspolicies -A`, expected:
> none on a fresh install) and delete just that CRD before re-running the command above:
> `kubectl delete crd backendtlspolicies.gateway.networking.k8s.io`.

## 3. Apply the shared Gateway

```bash
kubectl apply -f dev-local/envoy/manifests
kubectl wait --for=condition=Ready certificate/crownlabs-tls -n default --timeout=30s
```

This applies:
- `ClusterIssuer selfsigned` + `Certificate crownlabs-tls` — the wildcard cert
  (`*.crownlabs.local`), cert-manager-issued, in the `crownlabs-tls` Secret.
- `GatewayClass envoy-gateway` and `Gateway crownlabs` (namespace `default`) — one HTTPS
  listener (443, terminates TLS with the cert above) and one HTTP listener (80, redirect
  only).
- `HTTPRoute http-to-https-redirect` — any plain-HTTP request to `*.crownlabs.local`
  gets a 301 to HTTPS.

Envoy Gateway auto-creates the actual `LoadBalancer` `Service` backing this `Gateway`,
under `envoy-gateway-system`, with an auto-generated name. Find it:

```bash
kubectl get svc -n envoy-gateway-system
# look for the LoadBalancer-type Service, e.g. envoy-default-crownlabs-<hash>
```

Once Traefik is disabled (step 1) and its DaemonSet has released 80/443, k3s's
`ServiceLB` assigns those same ports to this Service — confirm:

```bash
kubectl get gateway crownlabs -n default
# expected: PROGRAMMED = True (it's False / "AddressNotAssigned" until Traefik is
# actually gone and 80/443 are free)
```

## 4. `/etc/hosts`

One line per hostname routed through the Gateway, pointing at `127.0.0.1` (this is a
single-node cluster, so the node *is* `127.0.0.1`):

```bash
echo "127.0.0.1 keycloak.crownlabs.local" | sudo tee -a /etc/hosts
```

```bash
echo "127.0.0.1 mail.crownlabs.local" | sudo tee -a /etc/hosts
```

> **On WSL2, this alone is not enough for the browser.** The command above only
> edits the *Linux/WSL2-side* `/etc/hosts` — it's what lets processes running
> inside WSL2 itself (the apiserver, `curl`, the operator) resolve the hostname.
> The browser runs on **Windows**, which resolves hostnames using its own,
> separate hosts file — tested and confirmed: without this, the browser fails
> with `DNS_PROBE_FINISHED_NXDOMAIN`, not a certificate warning, since it never
> even gets far enough to open a connection. Add the same line there too, from
> an **elevated** PowerShell (plain Notepad-as-administrator can silently fail
> here — Windows' File/Registry Virtualization redirects the save to a
> per-user shadow copy with no error if the elevation didn't actually take,
> so the edit looks successful but the real file is untouched):
> ```powershell
> Add-Content -Path C:\Windows\System32\drivers\etc\hosts -Value "127.0.0.1 keycloak.crownlabs.local"
> ```
> Verify it actually landed (`Get-Content C:\Windows\System32\drivers\etc\hosts
> | Select-String crownlabs`), then `ipconfig /flushdns` and `ping
> keycloak.crownlabs.local` (expect a reply from `127.0.0.1`) before retrying
> the browser.

## 5. On WSL2: bridge the Gateway

Same underlying limitation already documented for Keycloak's old NodePort (tested and
confirmed on both `networkingMode`s): a `LoadBalancer` Service backed by k3s's
`ServiceLB` has no *real* listening socket on 80/443 — it's iptables DNAT — so Windows'
`localhostForwarding` has nothing to relay, and the browser gets
`ERR_CONNECTION_REFUSED`. The fix is the same pattern as before, but now there's only
**one** bridge to keep open, covering every hostname behind the Gateway instead of one
per service:

```bash
kubectl port-forward -n envoy-gateway-system svc/<gateway-service-name> 8443:443
```

(substitute the Service name found in step 3). From here on, **every hostname in this
guide needs `:8443` appended on WSL2** (`https://keycloak.crownlabs.local:8443`, etc.) —
same reasoning as the old `:18543` bridge: reusing port 443 itself gets hijacked by
kube-proxy's own iptables rules for that port and times out.

## Final checks

```bash
curl -sk -o /dev/null -w "%{http_code}\n" https://keycloak.crownlabs.local/realms/crownlabs
# WSL2: https://keycloak.crownlabs.local:8443/realms/crownlabs
# expected: 200, once ../keycloak/README.md's manifests are applied
```

Expected end state:
- ✅ Traefik disabled, Gateway API CRDs upgraded, Envoy Gateway controller running.
- ✅ `Gateway crownlabs` `PROGRAMMED = True`.
- ✅ `/etc/hosts` has an entry for every hostname you use.
- ✅ (WSL2 only) one `kubectl port-forward` bridge open.

## Next

- [`../keycloak/README.md`](../keycloak/README.md) — Keycloak, now reachable at
  `https://keycloak.crownlabs.local` via the `HTTPRoute` in its own manifest.
- [`../mailpit/README.md`](../mailpit/README.md) — fake SMTP + web UI, reachable at
  `https://mail.crownlabs.local` the same way.
