# Base k3s cluster

This is the base k3s cluster for the local development environment. Set it up before anything else in `dev-local/`.

## 1. Install k3s

```bash
curl -sfL https://get.k3s.io | sh -
```

## 2. kubeconfig access

k3s writes its kubeconfig to `/etc/rancher/k3s/k3s.yaml`, owned by root and readable only by root by default.

**Don't copy it to `~/.kube/config`** — a copy goes stale the moment k3s regenerates its certificates (which it does on its own schedule, not just on `systemctl restart`), silently leaving you with an unusable kubeconfig until you remember to re-copy it.
Instead, tell k3s to write the *original* file group-readable, and read it from its real location directly — it's always fresh, because it's the actual file k3s keeps updating in place, not a snapshot of it.

This can be done with the following commands:

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

**Log out and back in** (or `newgrp k3s-admins` for just the current shell) so your new group membership actually takes effect — group changes don't apply to already-open sessions.

### If you don't already have a kubeconfig (or don't care about other clusters)

```bash
echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' >> ~/.bashrc
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

### If you already have a kubeconfig with other clusters/contexts

Point `KUBECONFIG` at **both files at once** — not a one-time copy.
This way k3s's own file is always read fresh from disk, however often its security certificates get regenerated:

```bash
echo 'export KUBECONFIG=~/.kube/config:/etc/rancher/k3s/k3s.yaml' >> ~/.bashrc
export KUBECONFIG=~/.kube/config:/etc/rancher/k3s/k3s.yaml
```

k3s always names its cluster, user and context `default`.
This will likely collide with an entry already in `~/.kube/config` — tools like Docker Desktop, kind, and minikube also pick the name `default`.
When `KUBECONFIG` lists multiple files, `kubectl` silently lets the first file's entry shadow the second one, separately for the cluster, user and context maps.
If that's your case, rename the **other** entry's cluster, user *and* context — all three, since renaming only the context still leaves the cluster and user names colliding, which would point k3s's context at the other tool's cluster.
You can rename them however you prefer: `kubectl config rename-context`, `set-cluster`/`set-credentials`, or just an editor.
You own `~/.kube/config`, so it's safe to edit freely.
Leave `/etc/rancher/k3s/k3s.yaml` untouched, though — it's only group-*readable*, and must stay that way to keep being read live:

```bash
kubectl config use-context default   # k3s
```

## Verify

```bash
kubectl get nodes
# expected: your node (i.e., the machine k3s is running upon), Ready
kubectl config current-context
# expected: default
ls -l /etc/rancher/k3s/k3s.yaml
# expected: -rw-r----- root:k3s-admins (mode 0640, group k3s-admins)
```
