# Base k3s cluster

The single-node k3s cluster everything else in `dev-local/` (Keycloak,
operators) runs on top of. Do this first.

## 1. Install k3s

```bash
curl -sfL https://get.k3s.io | sh -
```

## 2. kubeconfig access

k3s writes its kubeconfig to `/etc/rancher/k3s/k3s.yaml`, readable only by
root. How you get access to it depends on whether this is the only cluster
you work with on this machine.

### If you don't already have a kubeconfig (or don't care about other clusters)

```bash
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config
chmod 600 ~/.kube/config
export KUBECONFIG=~/.kube/config
```

### If you already have a kubeconfig with other clusters/contexts

Don't overwrite `~/.kube/config` — merge k3s's kubeconfig into it instead.
k3s always names its cluster, user and context `default`, which will very
likely collide with an entry already in your kubeconfig (e.g. Docker
Desktop, kind, minikube also default to `default`) — a plain merge would
silently clobber whichever one is listed second. Rename them to something
unique first, *then* merge:

```bash
sudo cp /etc/rancher/k3s/k3s.yaml /tmp/k3s.yaml
sudo chown $(id -u):$(id -g) /tmp/k3s.yaml

# k3s.yaml has exactly one cluster/user/context, all literally named "default" —
# rename all of them to something unique before merging, anchored to their
# specific YAML keys so nothing inside the embedded certificate data is touched.
sed -i \
  -e 's/^  name: default$/  name: k3s-local/' \
  -e 's/^- name: default$/- name: k3s-local/' \
  -e 's/^current-context: default$/current-context: k3s-local/' \
  -e 's/^    cluster: default$/    cluster: k3s-local/' \
  -e 's/^    user: default$/    user: k3s-local/' \
  /tmp/k3s.yaml

KUBECONFIG=~/.kube/config:/tmp/k3s.yaml kubectl config view --flatten > /tmp/merged-config
mv /tmp/merged-config ~/.kube/config
chmod 600 ~/.kube/config
rm /tmp/k3s.yaml

kubectl config use-context k3s-local   # switch to it now, or later with the same command
export KUBECONFIG=~/.kube/config
```

Either way, add `export KUBECONFIG=~/.kube/config` to your `.bashrc`/`.zshrc`
— every `kubectl` command in the rest of `dev-local/` assumes it's set.

## Verify

```bash
kubectl get nodes
# expected: your node, Ready
kubectl config current-context
# expected: k3s-local (if you merged) or default (if you copied directly)
```

## Next

- [`../keycloak/README.md`](../keycloak/README.md) — Keycloak, HTTPS, and the k3s apiserver OIDC integration.
- [`../operators/README.md`](../operators/README.md) — base RBAC and running the CrownLabs operator (needs Keycloak first).
