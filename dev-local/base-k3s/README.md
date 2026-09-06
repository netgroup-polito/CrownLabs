# Base k3s cluster

CrownLabs needs a Kubernetes cluster, to host:

- the required tooling (envoy, keycloak, ...) installed in the other steps of these guides
- any container that you start from inside the frontend

In the following steps, you can find the commands to set it up.

## 1. Install k3s

```bash
curl -sfL https://get.k3s.io | sh -
```

k3s is a `systemctl` service, therefore you have the usual commands:

- `sudo systemctl start k3s`: power the cluster on
- `sudo systemctl stop k3s`: power the cluster off
- `sudo systemctl enable k3s`: enable the automatic start on system boot
- `sudo systemctl disable k3s`: disable the automatic start on system boot

Leaving the cluster on will not host anything running by itself.
However, you will probably install (and have always running):

- the envoy ingress controller
- the keycloak process
- any container you start from the

## 2. kubeconfig access

[kubeconfig]: ## "The kubeconfig is a file that contains all the authorization details to connect to a Kubernetes cluster with the kubectl command. While the guides don't use the command directly, it's still used inside the commands we use. For this reason, we need to have a correct kubeconfig pointing to the k3s cluster we just created. kubectl will read the config from ~/.kube/config."

k3s writes its [kubeconfig][kubeconfig] to `/etc/rancher/k3s/k3s.yaml`, owned by root and readable only by root by default.

The intuitive setup would be to copy this file to `~/.kube/config`.
However, this approach does not work, since k3s periodically updates its certificates.
Whenever that happens, your copied kubeconfig would not work anymore, since the certificates do not match.

Instead, the best option consists in instructing kubectl to use the original file, that is always updated.
This however requires to make it accessible enough.

### 2.1. Making the kubeconfig accessible

Inside the OS, we create a new group called `k3s-admins`, to which we register:

```bash
sudo groupadd --system k3s-admins 2>/dev/null || true
sudo usermod -aG k3s-admins "$USER"
```

Then, we need to configure k3s so that the kubeconfig is readable by this group.
This can be done by means of a yaml file in the config folder for k3s.

We cannot modify directly the `config.yaml` of k3s, but we can specify our changes in a `.yaml` file inside the folder `config.yaml.d`.
When starting, k3s will take care of integrating all the files of this folder in alphabetical order into its `config.yaml`.
We will exploit this functionality in the future guides, when we will apply more modifications to the k3s config.

The following command creates the `config.yaml.d` folder if not present.
Then, it creates a config modification that tells k3s to create the kubeconfig with group `k3s-admins` and permission 640 (owner RW, group R, other nothing).

```bash
sudo mkdir -p /etc/rancher/k3s/config.yaml.d
sudo tee /etc/rancher/k3s/config.yaml.d/10-kubeconfig-access.yaml > /dev/null <<'EOF'
write-kubeconfig-mode: "0640"
write-kubeconfig-group: k3s-admins
EOF
```

The cluster needs to be restarted for the changes to take place:

```bash
sudo systemctl restart k3s
```

Finally, we need the group change to take effect in the shell.
This can be done by either:

- running the command `newgrp k3s-admins` to update the single shell; or
- logging out and in from the system, to update all shells.

### 2.2. Instructing kubectl to use the file (single kubeconfig)

Note: if you already have a kubeconfig on your system, please skip to the next section.
This step assumes that either the local k3s is the only cluster you are accessing, or you don't care about the previous kubeconfig you had.

kubectl uses `~/.kube/config` as config location, unless a `KUBECONFIG` environment variable is set.
Therefore, we just need to set it to point to the correct file (both in the shell and in `~/.bashrc`, so that it is set in every shell):

```bash
echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' >> ~/.bashrc
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

### 2.2. Instructing kubectl to use the file (multiple kubeconfigs)

In this case, we can have the `KUBECONFIG` env variable point to multiple configuration files:

```bash
echo 'export KUBECONFIG=~/.kube/config:/etc/rancher/k3s/k3s.yaml' >> ~/.bashrc
export KUBECONFIG=~/.kube/config:/etc/rancher/k3s/k3s.yaml
```

This way, kubectl will be able to load multiple kubeconfigs at the same time.
Kubeconfigs will be read in the order they appear, in this case the default one will be before the k3s.
Each cluster, context and user in a new file will be considered _if and only if_ it has a name different from any previous one.
We have to ensure that.

As we said, k3s will regenerate the file at "random" intervals, so we cannot rely on changing its names.
In particular, it will always use `default` for both the cluster, the context and the user.
Therefore, we have to change the names in the pre-existing kubeconfig(s) to something different than `default`.
Remember to change both the cluster, the context and the user names, and to update the content of the context accordingly

<!-- prettier-ignore-start -->

> Here is an example of how to update the kubeconfig:
>
> ```yaml
> apiVersion: v1
> kind: Config
> clusters:
> - cluster: ...
>   name: default # --> nameCluster
> contexts:
> - context:
>     cluster: default # --> nameCluster
>     user: default # --> nameUser
>     name: default # --> nameContext
> current-context: default # --> nameContext
> users:
> - name: default # --> nameUser
>   user: ...
> ```
>
> The three names do not necessarily need to be different.

<!-- prettier-ignore-end -->

Once you have the two (or more) kubeconfigs, you can check that kubectl can see both:

```bash
kubectl config get-contexts
```

You can switch from one to the other using:

_NOTE: after further testing, the following commands do not work.
Initially, they were without the sudo, but they were raising an error.
The error disappeared with the sudo, but it was later discovered that the context change would not take place for the user, if the command is run with sudo.
The correct command will be published in a future commit, once found.
In the meantime, please use a single kubeconfig (that part of the guide is confirmed to work)._

```bash
sudo kubectl config use-context default       # switch to k3s
sudo kubectl config use-context <other-name>  # switch to another cluster
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
