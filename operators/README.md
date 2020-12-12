# Crownlabs Operators

## APIs/CRDs

The Laboratory Operator (LabOperator) implements the backend logic necessary to spawn new laboratories starting from a predefined template. LabOperator relies on two Kubernetes Custom Resource
Definitions (CRDs) which implement the basic APIs:
* **Laboratory Template (LabTemplate)** defines the size of the execution environment (e.g.; Virtual Machine), its base image and a description. This object is created by professors and read by students, while creating new instances.
* **Laboratory Instance (LabInstance)** defines an instance of a certain template. The manipulation of those objects triggers the reconciliation logic in LabOperator, which creates/destroy associated resources (e.g.; Virtual Machines).



Both LabTemplates and LabInstances are **namespaced**.

#### Add CRDs to the cluster

Before the deploying the operator, we have to add the LabInstance and LabTemplate CRDs. This can be done via the Makefile:

```bash
make install
```

## Laboratory Instance Operator (LabOperator)

Based on [Kubebuilder 2.3](https://github.com/kubernetes-sigs/kubebuilder.git), the operator implements the laboratory creation logic of Crownlabs.

Upon the creation of a *LabInstance*, the operator triggers the creation of the following components:
* Kubevirt VirtualMachine Instance and the logic to access the noVNC instance inside the VM (Service, Ingress)
* An instance of [Oauth2 Proxy](https://github.com/oauth2-proxy/oauth2-proxy) (Deployment, Service, Ingress) to regulate access to the VM.

All those resources are binded to the LabInstance life-cycle via the [OwnerRef property](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/)

### Installation

#### Pre-requirements

The only LabOperator requirement is to have Kubevirt deployed.
This can be done with the following commands, as reported by the official website:

```bash
# On other OS you might need to define it like
export KUBEVIRT_VERSION="v0.34.0"

# Deploy the KubeVirt operator
kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-operator.yaml
# Only if HW Virtualization is not available
kubectl create configmap kubevirt-config -n kubevirt --from-literal debug.useEmulation=true
# Deploy Kubevirt
kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml
```

#### Deployment
To deploy the LabOperator in your cluster, you have to do the following steps.

First, set the desired values in `operators/deploy/laboratory-operator/k8s-manifest-example.env` .

Then export the environment variables and generate the manifest from the template using:

```
cd operators/deploy/laboratory-operator
export $(xargs < k8s-manifest-example.env)
envsubst < k8s-manifest.yaml.tmpl > k8s-manifest.yaml
```

After the manifest have been correctly generated. You can deploy the labOperator using:

```
kubectl apply -f k8s-manifest.yaml
```

### Build from source

LabOperator requires Golang 1.13 and make. To build the operator:

```bash
go build ./cmd/laboratory-operator/main.go
```

#### Testing

After having installed Kubevirt in your testing cluster, you have to deploy the Custom Resource Definitions (CRDs) on the target cluster:

```bash
make install
```

N.B. So far, the readiness check for VirtualMachines is performed by assuming that the operator is running on the same cluster of the Virtual Machines. This prevents the possibility to have *ready* VMs when testing the operator outside the cluster.

## SSH bastion

The SSH bastion is composed of a two basic blocks:
1. `bastion-operator`: an operator based on on [Kubebuilder 2.3](https://github.com/kubernetes-sigs/kubebuilder.git)
2. `ssh-bastion`: a lightweight alpine based container running [sshd](https://linux.die.net/man/8/sshd)

### Installation

#### Pre-requirements

The only pre-requirement needed in order to deploy the SSH bastion is `ssh-keygen` and it is needed only in case you don't already have the host keys that sshd will use.
You can check if you already have `ssh-keygen` install running:
```bash
ssh-keygen --help
```
To install it (i.e. on Ubuntu) run:
```bash
apt install openssh-client
```

#### Deployment

To deploy the SSH bastion in your cluster, you have to do the following steps.

First, generate the host keys needed to run sshd using:
```bash
# Generate the keys in this folder (they will be ignored by git) or in a folder outside the project
ssh-keygen -f ssh_host_key_ecdsa -N "" -t ecdsa
ssh-keygen -f ssh_host_key_ed25519 -N "" -t ed25519
ssh-keygen -f ssh_host_key_rsa -N "" -t rsa
```

Now create the secret holding the keys. If the bastion is going to run on a namespace different than default add the `--namespace=<namespace>` option.
```bash
kubectl create secret generic ssh-bastion-host-keys \
  --from-file=./ssh_host_key_ecdsa \
  --from-file=./ssh_host_key_ed25519 \
  --from-file=./ssh_host_key_rsa
```

Then set the desired values in `operators/deploy/bastion-operator/k8s-manifest-example.env` .

Export the environment variables and generate the manifest from the template using:

```bash
cd operators/deploy/bastion-operator
export $(xargs < k8s-manifest-example.env)
envsubst < k8s-manifest.yaml.tmpl > k8s-manifest.yaml
```

After the manifest have been correctly generated you can install the cluster role and deploy the SSH bastion using:

```bash
kubectl apply -f k8s-cluster-role.yaml
kubectl apply -f k8s-manifest.yaml
```

## CrownLabs Image List

The CrownLabs Image List script allows to to gather the list of available images from a Docker Registry and expose it as an ImageList custom resource, to be consumed from the CrownLabs dashboard.

### Usage

```
usage: update-crownlabs-image-list.py [-h]
    --advertised-registry-name ADVERTISED_REGISTRY_NAME
    --image-list-name IMAGE_LIST_NAME
    --registry-url REGISTRY_URL
    [--registry-username REGISTRY_USERNAME]
    [--registry-password REGISTRY_PASSWORD]
    --update-interval UPDATE_INTERVAL

Periodically requests the list of images from a Docker registry and stores it as a Kubernetes CR

Arguments:
  -h, --help            show this help message and exit
  --advertised-registry-name ADVERTISED_REGISTRY_NAME
                        the host name of the Docker registry where the images can be retrieved
  --image-list-name IMAGE_LIST_NAME
                        the name assigned to the resulting ImageList object
  --registry-url REGISTRY_URL
                        the URL used to contact the Docker registry
  --registry-username REGISTRY_USERNAME
                        the username used to access the Docker registry
  --registry-password REGISTRY_PASSWORD
                        the password used to access the Docker registry
  --update-interval UPDATE_INTERVAL
                        the interval (in seconds) between one update and the following
```

### Deployment

A sample configuration required for a deployment in a Kubernetes cluster is available in the [deploy folder](deploy/crownlabs-image-list).
