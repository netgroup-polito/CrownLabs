# Base k3s cluster

The single-node k3s cluster everything else in `dev-local/` (Keycloak,
operators) runs on top of. Do this first.

## 1. Install k3s

```bash
curl -sfL https://get.k3s.io | sh -
```

## 2. kubeconfig access

k3s writes its kubeconfig to `/etc/rancher/k3s/k3s.yaml`, owned by root and
readable only by root by default.

**Don't copy it to `~/.kube/config`** — a copy goes stale the moment k3s
regenerates its certificates (which it does on its own schedule, not just on
`systemctl restart`), silently leaving you with an unusable kubeconfig until
you remember to re-copy it. Instead, tell k3s to write the *original* file
group-readable, and read it from its real location directly — it's always
fresh, because it's the actual file k3s keeps updating in place, not a
snapshot of it.

```bash
# One-time: a dedicated group whose members can read k3s's kubeconfig,
# and you as a member of it — root remains the file's owner either way.
sudo groupadd --system k3s-admins 2>/dev/null || true
sudo usermod -aG k3s-admins "$USER"

# k3s.yaml.d/*.yaml drop-ins merge automatically (alphabetical order) into
# the main config.yaml — each concern gets its own file instead of everyone
# editing the same shared one (see envoy/README.md and
# keycloak/apiserver-oidc-integration.md, which each add their own drop-in
# here later, with no risk of clobbering this one).
sudo mkdir -p /etc/rancher/k3s/config.yaml.d
sudo tee /etc/rancher/k3s/config.yaml.d/10-kubeconfig-access.yaml > /dev/null <<'EOF'
write-kubeconfig-mode: "0640"
write-kubeconfig-group: k3s-admins
EOF

sudo systemctl restart k3s
```

**Log out and back in** (or `newgrp k3s-admins` for just the current shell)
so your new group membership actually takes effect — group changes don't
apply to already-open sessions.

### If you don't already have a kubeconfig (or don't care about other clusters)

```bash
echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' >> ~/.bashrc
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

### If you already have a kubeconfig with other clusters/contexts

Point `KUBECONFIG` at **both files, live** — not a one-time flattened copy —
so k3s's own file keeps being read straight from disk, fresh, no matter how
often its certs get regenerated:

```bash
echo 'export KUBECONFIG=~/.kube/config:/etc/rancher/k3s/k3s.yaml' >> ~/.bashrc
export KUBECONFIG=~/.kube/config:/etc/rancher/k3s/k3s.yaml
```

k3s always names its cluster, user and context literally `default`, which
will very likely collide with an entry already in `~/.kube/config` (e.g.
Docker Desktop, kind, minikube also default to `default`) — when `KUBECONFIG`
lists multiple files, `kubectl` silently lets the first file's entry shadow
the second's, independently for the cluster, user and context maps. If
that's your case, rename the **other** entry's cluster, user *and* context
(all three — renaming only the context still leaves the cluster/user names
colliding, which points k3s's context at the other tool's cluster) to
anything else, however you prefer (`kubectl config rename-context`,
`set-cluster`/`set-credentials`, or just an editor) — you own
`~/.kube/config` and can edit it freely; `/etc/rancher/k3s/k3s.yaml` is
group-*readable* only, and must stay untouched to keep being read live:

```bash
kubectl config use-context default   # k3s
```

## Verify

```bash
kubectl get nodes
# expected: your node, Ready
kubectl config current-context
# expected: default
ls -l /etc/rancher/k3s/k3s.yaml
# expected: -rw-r----- root:k3s-admins (mode 0640, group k3s-admins)
```
