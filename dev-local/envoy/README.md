# Envoy Gateway — single ingress point for dev-local

This guide sets up a single [Envoy Gateway](https://gateway.envoyproxy.io/) (Kubernetes Gateway API) in front of every locally-exposed dev-local service.
Every service becomes reachable under one local domain, `*.crownlabs.local`.
This replaces the old `NodePort` pattern, which needed a separate port for each service.

Set up the Gateway before you set up Keycloak.
See [`../keycloak/README.md`](../keycloak/README.md) for the Keycloak setup.
Keycloak's manifest defines an `HTTPRoute` object.
That `HTTPRoute` needs the `Gateway` you create in this guide, so the Gateway must exist first.

TLS terminates at the Gateway, using one self-signed wildcard certificate.
Every backend behind the Gateway, such as Keycloak and [Mailpit](../mailpit/README.md), is reached over plain HTTP internally.
No other component needs its own certificate or HTTPS listener.

## 0. Prerequisites

To run the Envoy Gateway, you need the following components:

- The base k3s cluster. See [`../base-k3s/README.md`](../base-k3s/README.md).
- cert-manager, which issues the wildcard certificate used below:
  ```bash
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.20.3/cert-manager.yaml
  kubectl wait --for=condition=Available deployment --all -n cert-manager --timeout=120s
  ```

## 1. Disable k3s's bundled Traefik

Traefik is k3s's default ingress controller.
It runs its own `LoadBalancer` `Service`, bound to the node's ports 80 and 443 through k3s's `ServiceLB` (the `svclb-traefik` pod).
The Envoy Gateway's own `Service` needs the same ports, so Traefik has to be removed.
This project uses Envoy Gateway instead of Traefik's built-in Gateway API support, because Envoy Gateway supports more features.

To disable Traefik, add a dedicated drop-in file under `/etc/rancher/k3s/config.yaml.d/` (this requires root access).
k3s automatically merges every `*.yaml` file in that folder into its configuration, in alphabetical order.
This means your new file will not touch anything that [`../base-k3s/README.md`](../base-k3s/README.md) or [`../keycloak/apiserver-oidc-integration.md`](../keycloak/apiserver-oidc-integration.md) already added there, and vice versa.
There is no single shared file to edit by hand.

```bash
sudo mkdir -p /etc/rancher/k3s/config.yaml.d
sudo tee /etc/rancher/k3s/config.yaml.d/20-disable-traefik.yaml > /dev/null <<'EOF'
disable:
  - traefik
EOF
```

```bash
sudo systemctl restart k3s
journalctl -u k3s -f   # Confirm there are no errors, then press Ctrl+C.
kubectl get nodes      # Expected: the node is Ready.
```

Restarting k3s restarts the whole control plane.
This is the same caveat as the API server OIDC step: the cluster is briefly unreachable, and you need to restart any open `kubectl proxy`, `port-forward`, or operator connection afterwards.

Confirm Traefik is gone:

```bash
kubectl get pods -n kube-system | grep -i traefik
# Expected: no traefik or svclb-traefik pods.
# The completed helm-install-traefik* jobs may still be listed: this is harmless, since they are one-shot Jobs, not live components.
```

If disabling `traefik` on an already-running cluster does not clean up the leftover `traefik` or `svclb-traefik` Deployment/DaemonSet on your k3s version, remove them directly:

```bash
kubectl delete deployment traefik -n kube-system
kubectl delete svc traefik -n kube-system
```

This does **not** touch the Gateway API CRDs.
Those CRDs come from a separate `traefik-crd` Helm chart, which `--disable=traefik` does not affect.
Envoy Gateway's own CRDs coexist with, and upgrade, that chart's CRDs in step 2 below.

## 2. Install the Gateway API CRDs and the Envoy Gateway controller

Always pin an explicit release, and never use `latest`.
The `latest` *tag* for the install manifest and the `latest` *image* tag can drift out of sync with each other.
We tested this: it produced a controller crash, with the error `no matches for kind "ListenerSet" in version "gateway.networking.k8s.io/v1"`, because the CRD bundle was older than what the controller image expected.

```bash
kubectl apply --server-side --force-conflicts -f https://github.com/envoyproxy/gateway/releases/download/v1.8.2/install.yaml
kubectl wait --timeout=120s -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

The `--force-conflicts` flag is required here.
k3s's bundled `traefik-crd` chart already installs the core Gateway API CRDs (`GatewayClass`, `Gateway`, `HTTPRoute`, and others) at an older schema version, managed by `helm`.
This flag lets Envoy Gateway take over that field ownership, so it can upgrade the CRDs.
It is safe to run this command even before you disable Traefik itself, because Helm never deletes CRDs when you uninstall a chart.
This means the CRDs stay safe even after the `traefik-crd` release eventually goes away.

> If you hit the error `The CustomResourceDefinition "backendtlspolicies.gateway.networking.k8s.io" is invalid: status.storedVersions[0]: ... must remain in spec.versions`, please do the following:
> 1) Confirm no `BackendTLSPolicy` objects exist yet (`kubectl get backendtlspolicies -A`, expected: none on a fresh install);
> 2) Delete just that CRD before re-running the command above:
> `kubectl delete crd backendtlspolicies.gateway.networking.k8s.io`.

## 3. Apply the shared Gateway

```bash
kubectl apply -f dev-local/envoy/manifests
kubectl wait --for=condition=Ready certificate/crownlabs-tls -n default --timeout=30s
```

This applies:
- `ClusterIssuer selfsigned` and `Certificate crownlabs-tls`: the wildcard certificate (`*.crownlabs.local`), issued by cert-manager and stored in the `crownlabs-tls` Secret.
- `GatewayClass envoy-gateway` and `Gateway crownlabs` (in the `default` namespace): one HTTPS listener on port 443, which terminates TLS with the certificate above, and one HTTP listener on port 80, used only for redirects.
- `HTTPRoute http-to-https-redirect`: any plain HTTP request to `*.crownlabs.local` gets a 301 redirect to HTTPS.

Envoy Gateway automatically creates the actual `LoadBalancer` `Service` that backs this `Gateway`, under the `envoy-gateway-system` namespace, with an auto-generated name.
To find it:

```bash
kubectl get svc -n envoy-gateway-system
# Look for the LoadBalancer-type Service, for example envoy-default-crownlabs-<hash>.
```

Once Traefik is disabled (step 1) and its DaemonSet has released ports 80 and 443, k3s's `ServiceLB` assigns those same ports to this Service.
You can verify with the following command:

```bash
kubectl get gateway crownlabs -n default
# Expected: PROGRAMMED = True.
# It stays False, with reason "AddressNotAssigned", until Traefik is actually gone and ports 80/443 are free.
```

## 4. `/etc/hosts`

Add one line per hostname routed through the Gateway, pointing at `127.0.0.1`.
This is a single-node cluster, so the node itself *is* `127.0.0.1`.

```bash
echo "127.0.0.1 keycloak.crownlabs.local" | sudo tee -a /etc/hosts
```

```bash
echo "127.0.0.1 mail.crownlabs.local" | sudo tee -a /etc/hosts
```

Please note that the changes to the local `/etc/hosts` file only take effect when you connect to the cluster from the same host.
If a second machine also connects to the k3s cluster, you need to update the `/etc/hosts` file on that machine too, replacing `127.0.0.1` with the actual IP address of the k3s host.

> **On WSL2, editing `/etc/hosts` alone is not enough for the browser.**
> The command above only edits the *Linux/WSL2-side* `/etc/hosts` file.
> This lets processes running inside WSL2 itself (the API server, `curl`, the operator) resolve the hostname.
> The browser, however, runs on **Windows**, and Windows resolves hostnames using its own, separate hosts file.
> We tested this: without updating the Windows hosts file, the browser fails with `DNS_PROBE_FINISHED_NXDOMAIN`, not a certificate warning, because the connection never even opens.
> Add the same line to the Windows hosts file too, from an **elevated** PowerShell window.
> Plain Notepad-as-administrator can silently fail here: if the elevation did not actually take effect, Windows' File/Registry Virtualization redirects the save to a per-user shadow copy, with no error message. The edit then looks successful, but the real file stays untouched.
> ```powershell
> Add-Content -Path C:\Windows\System32\drivers\etc\hosts -Value "127.0.0.1 keycloak.crownlabs.local"
> ```
> Verify the change actually landed:
> ```powershell
> Get-Content C:\Windows\System32\drivers\etc\hosts | Select-String crownlabs
> ```
> Then run `ipconfig /flushdns` and `ping keycloak.crownlabs.local` (expect a reply from `127.0.0.1`) before you retry the browser.

## 5. On WSL2: bridge the Gateway

On WSL2, the browser on Windows cannot reach the Gateway's ports directly, for the same reason described earlier for Keycloak's old NodePort setup.
A `LoadBalancer` Service backed by k3s's `ServiceLB` does not open a real listening socket on ports 80 and 443.
Instead, it uses `iptables` rules to redirect traffic (DNAT).
Windows' `localhostForwarding` feature only relays real listening sockets, so it has nothing to forward here, and the browser fails with `ERR_CONNECTION_REFUSED`.

The fix follows the same pattern used before, but now you only need **one** bridge, covering every hostname behind the Gateway instead of one bridge per service:

```bash
kubectl port-forward -n envoy-gateway-system svc/<gateway-service-name> 8443:443
```

Replace `<gateway-service-name>` with the Service name you found in step 3.
From here on, **append `:8443` to every hostname in this guide when you use WSL2** (for example, `https://keycloak.crownlabs.local:8443`).
This is the same reasoning as the old `:18543` bridge: if you try to reuse port 443 itself, kube-proxy's own `iptables` rules for that port hijack the connection, and it times out.

## Final checks

```bash
curl -sk -o /dev/null -w "%{http_code}\n" https://keycloak.crownlabs.local/realms/crownlabs
# On WSL2: https://keycloak.crownlabs.local:8443/realms/crownlabs
# Expected: 200, once you apply the manifests from ../keycloak/README.md.
```

Expected end state:
- ✅ Traefik is disabled, the Gateway API CRDs are upgraded, and the Envoy Gateway controller is running.
- ✅ `Gateway crownlabs` shows `PROGRAMMED = True`.
- ✅ `/etc/hosts` has an entry for every hostname you use.
- ✅ (WSL2 only) one `kubectl port-forward` bridge is open.

## Next

- [`../keycloak/README.md`](../keycloak/README.md): set up Keycloak, now reachable at `https://keycloak.crownlabs.local` through the `HTTPRoute` in its own manifest.
- [`../mailpit/README.md`](../mailpit/README.md): set up the fake SMTP server and web UI, reachable at `https://mail.crownlabs.local` the same way.
