# Envoy Gateway — single ingress point for dev-local

This guide sets up a single [Envoy Gateway](https://gateway.envoyproxy.io/) (Kubernetes Gateway API) in front of every locally-exposed dev-local service.
Every service becomes reachable under one local domain, `*.crownlabs.local`.
This replaces the old `NodePort` pattern, which needed a separate port for each service.

<!-- Set up the Gateway before you set up Keycloak.
See [`../keycloak/README.md`](../keycloak/README.md) for the Keycloak setup.
Keycloak's manifest defines an `HTTPRoute` object.
That `HTTPRoute` needs the `Gateway` you create in this guide, so the Gateway must exist first. -->

TLS terminates at the Gateway, using one self-signed wildcard certificate.
Every backend behind the Gateway, such as Keycloak and [Mailpit](../mailpit/README.md), is reached over plain HTTP internally.
No other component needs its own certificate or HTTPS listener.

## 1. Disable k3s's bundled Traefik

Traefik is k3s's default ingress controller.
It runs its own `LoadBalancer` `Service`, bound to the node's ports 80 and 443 through k3s's `ServiceLB` (the `svclb-traefik` pod).
The Envoy Gateway's own `Service` needs the same ports, so Traefik has to be removed.
This project uses Envoy Gateway instead of Traefik's built-in Gateway API support, because Envoy Gateway supports more features.

To disable Traefik, you can add a new file inside the `config.yaml.d` folder, as was already done to change the kubeconfig file settings:

```bash
sudo mkdir -p /etc/rancher/k3s/config.yaml.d
sudo tee /etc/rancher/k3s/config.yaml.d/20-disable-traefik.yaml > /dev/null <<'EOF'
disable:
  - traefik
EOF
```

To have these changes apply, a restart of the k3s cluster is required:

```bash
sudo systemctl restart k3s
journalctl -u k3s -f   # Confirm there are no errors, then press Ctrl+C.
kubectl get nodes      # Expected: the node is Ready.
```

> Note: if you get an error `failed to list *v1.PartialObjectMetadata` in the journalctl, that should not be a problem, as explained [here](https://github.com/rancher/rke2/discussions/6494#discussioncomment-10291104).

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

> [Gateway API](https://kubernetes.io/docs/concepts/services-networking/gateway/) [CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (CRDs = Custom Resources) is a family of API kinds that provide dynamic infrastructure provisioning and advanced traffic routing.

In this guide, we install Envoy 1.8.2.

Please never use `latest`, as the Gateway API and Envoy images tagged as `latest` can drift out of sync.

> In a test we conducted using the `latest` tag, the controller crashed with the error `no matches for kind "ListenerSet" in version "gateway.networking.k8s.io/v1"`.
> This is because the CRD bundle was older than what the controller image expected.

```bash
kubectl apply --server-side --force-conflicts -f https://github.com/envoyproxy/gateway/releases/download/v1.8.2/install.yaml
kubectl wait --timeout=120s -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

The `--force-conflicts` flag allows the Envoy Gateway to take over the Gateway API CRDs ownership, to upgrade them.
It is necessary, since there are already some Gateway API CRDs installed by traefik.

> If you hit the error `The CustomResourceDefinition "backendtlspolicies.gateway.networking.k8s.io" is invalid: status.storedVersions[0]: ... must remain in spec.versions`, please do the following:
>
> 1. Confirm no `BackendTLSPolicy` objects exist yet (`kubectl get backendtlspolicies -A`, expected: none on a fresh install);
> 2. Delete just that CRD before re-running the command above:
>    `kubectl delete crd backendtlspolicies.gateway.networking.k8s.io`.

## 3. Install cert-manager

Since Envoy terminates the TLS connection, it requires some way to generate certificates.
For development, we will only use self-signed certificates.
We will generate them during the next steps, now we start installing the container to handle them:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.20.3/cert-manager.yaml
kubectl wait --for=condition=Available deployment --all -n cert-manager --timeout=120s
```

## 4. Deploy Envoy

First of all, we have to generate a self-signed CA root authority.
We will use it to sign the next certificates, to be able to serve the pages via HTTPS.
The first time we will open the browser on a hosted page, we will have to instruct it to trust the certificate.

Then, we need a certificate.
For ease of use during development, we will only generate a single wildcard certificate `*.crownlabs.local`, that will work for all pages.
It is be stored in the `crownlabs-tls` secret.

For the gateway itself, we need to create a `GatewayClass` and a `Gateway` listening on ports 80 and 443.
For HTTPS, it uses the `crownlabs-tls` certificate we just created.

Finally, we add a redirect to any HTTP request.
A status 301 will prompt the user to use HTTPS.

All these things are done automatically by applying the manifest in this folder:

```bash
kubectl apply -f dev-local/envoy/manifests
kubectl wait --for=condition=Ready certificate/crownlabs-tls -n default --timeout=30s
```

Envoy Gateway automatically creates the actual `LoadBalancer` `Service` that backs this `Gateway`, under the `envoy-gateway-system` namespace, with an auto-generated name.
To find it:

```bash
kubectl get svc -n envoy-gateway-system
# Look for the LoadBalancer-type Service, for example envoy-default-crownlabs-<hash>.
```

Please take note of the name of the LoadBalancer service, as it will be required later in the guide.

Once Traefik is disabled (step 1) and its DaemonSet has released ports 80 and 443, k3s's `ServiceLB` assigns those same ports to this Service.
You can verify with the following command:

```bash
kubectl get gateway crownlabs -n default
# Expected: PROGRAMMED = True.
# It stays False, with reason "AddressNotAssigned", until Traefik is actually gone and ports 80/443 are free.
```

## 5. Configure the static DNS

In your local machine, you will try accessing URLs such as `keycloak.crownlabs.local`.
Since there is no DNS to point the URL to the right IP, you will have to add them manually.
This is done by inserting a tuple `<ip> <URL>` in the file `/etc/hosts` in your machine.

In the next steps of this guide, we will setup the keycloak and mailpit services, so we can start adding the DNS resolutions right now.
The target IP address is simply the host itself, so `127.0.0.1`:

```bash
echo "127.0.0.1 keycloak.crownlabs.local" | sudo tee -a /etc/hosts
echo "127.0.0.1 mail.crownlabs.local" | sudo tee -a /etc/hosts
```

Keep in mind that now we have the routings, but there is still nothing listening listening behind.
To really try if everything is working, you will first have to setup the next services.

These rules only work on the same machine where k3s is hosted, due to the reflexive nature of the address.
If you are accessing the pages from a different host in the same network, you would have to modify these rules, using the IP of the k3s host instead of `127.0.0.1`.

## 5b. Configure the static DNS on WSL2

_If you are not using WSL2, or you are accessing the pages directly on the Linux environment, you can ignore this step._

The `/etc/hosts` edited in the previous step only takes affect in the Linux side.
They will be in place if you access the URLs directly from Linux, for example with a `curl` command.
However, if you will use your Windows browser to access the URLs, they won't work.
Instead, you would see a `DNS_PROBE_FINISHED_NXDOMAIN` error, because the DNS cannot be solved.

To solve this issue, you have to add the static DNS entries also on Windows.

This can be done with the following commands, executed on an elevated PowerShell window:

```powershell
Add-Content -Path C:\Windows\System32\drivers\etc\hosts -Value "127.0.0.1 keycloak.crownlabs.local"
Add-Content -Path C:\Windows\System32\drivers\etc\hosts -Value "127.0.0.1 mail.crownlabs.local"
```

Please note that directly editing the file `C:\Windows\System32\drivers\etc\hosts` may not work, despite using Notepad "run as Administrator": if the elevation did not actually take effect, you would be editing a per-user shallow copy, without any notification.
You would then successfully edit this other file, but the real one remains untouched.

You can verify that the file was correctly edited with the following command:

```powershell
Get-Content C:\Windows\System32\drivers\etc\hosts | Select-String crownlabs
```

Moreover, you can flush the DNS and try to ping one of the two addresses, to check that the correct address (`127.0.0.1`) is used:

```powershell
ipconfig /flushdns
ping keycloak.crownlabs.local
```

## 6. Bridging the Gateway

_Strictly speaking, this step is only required if you are working on WSL2.
However, following it in the other cases does not produce any disadvantage.
For this reason, all URLs in the following guides will be shown assuming this step is active, in order to eliminate unnecessary distinctions in the guides._

The Envoy Gateway is backed by a k3s `ServiceLB`, which does not open a real listening socket on ports 80 and 443.
Instead, it uses `iptables` rules to redirect traffic (DNAT).

While this works on the Linux (WSL2) environment, it is not translated to Windows.
That's because the `localhostForwarding` from WSL2 to Windows only forwards real listening sockets.
Here there is none, so nothing is forwarded.
If you open the URLs in a Windows browser, it cannot connect, and you get an error `ERR_CONNECTION_REFUSED`.

The solution to this is to use the `kubectl port-forward` command to expose the port 443 from the envoy gateway to a real port in the k3s host. In this case, port 8443 is chosen.
For simplicity, the full command is saved in the [`./WSL2_bridge.sh`](./WSL2_bridge.sh) script.
Before running it, you need to configure it once with the following commands, replacing `<gateway-service-name>` with the Service name found in step 4:

```bash
GATEWAY_SERVICE_NAME=<gateway-service-name>
echo "export GW_SERVICE_NAME=$GATEWAY_SERVICE_NAME" >> ~/.bashrc
export GW_SERVICE_NAME=$GATEWAY_SERVICE_NAME
```

When running the `WSL2_bridge.sh` script, a process is started to keep the bridge opened as long as the script itself is running.
With the bridge opened, the URLs can be accessed from any location, both on Linux and on Windows. The only care required is that the port `:8443` must always be specified, since it is not the standard HTTPS port.

Please, keep in mind that:

- the command explained for the bridge only needs to be executed once (unless the gateway service name changes)
- `WSL2_bridge.sh` must instead be running every time the bridge is needed

_Note: if you elected to skip this step, all the URLs are still accessible in the environment where the k3s cluster is.
The only difference is that they will be on port 443 instead of port 8443.
Please keep this in mind also for all the URLs in the next guides._

## Final checks

As previously stated, there is no way to check the correctness of the envoy setup at this point, since it has nowhere to redirect requests to.

Once there are services running, you can try connecting them, to check that both the service and envoy are correctly running.

Expected end state:

- ✅ Traefik is disabled, the Gateway API CRDs are upgraded, and the Envoy Gateway controller is running.
- ✅ `Gateway crownlabs` shows `PROGRAMMED = True`.
- ✅ `/etc/hosts` has an entry for every hostname you use.
- ✅ `WSL2_bridge.sh` bridge is setup.
