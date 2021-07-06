# Virtualization

[KubeVirt](https://kubevirt.io/) is a solution to spawn and orchestrate traditional virtual machines on top of a Kubernetes cluster.

## How to install

The following presents a summary of the basic operations required to deploy KubeVirt in a Kubernetes cluster. Please, refer to the [official instructions](https://kubevirt.io/user-guide/#/installation/installation) for more details about customizing the configuration.

```bash
# Pick an upstream version of KubeVirt to install
$ export KUBEVIRT_VERSION=v0.42.1
# Deploy the KubeVirt operator
$ kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-operator.yaml
# Create the KubeVirt CR (instance deployment request) which triggers the actual installation
$ kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml
# wait until all KubeVirt components are up
$ kubectl -n kubevirt wait kv kubevirt --for condition=Available
```

# Containerized data importer
[Containerized-Data-Importer (CDI)](https://github.com/kubevirt/containerized-data-importer) is an operator which automates the creation and population of PVCs with VM images, to be attached to KubeVirt VMs. In particular, it relies on the `Datavolume` resource, which essentially describes the source the image is imported from. Please, refer to the [official documentation](https://github.com/kubevirt/containerized-data-importer/blob/master/doc/image-from-registry.md) for more information about this process.

## How to install
In the following you can find the commands you need to execute in order to deploy CDI in a Kubernetes cluster:
```bash
#Export last cdi version
export VERSION=$(curl -s https://github.com/kubevirt/containerized-data-importer/releases/latest | grep -o "v[0-9]\.[0-9]*\.[0-9]*")
#Deploy the cdi operator
kubectl create -f https://github.com/kubevirt/containerized-data-importer/releases/download/$VERSION/cdi-operator.yaml
#Deploy cdi CR which triggers the actual installation
kubectl create -f https://github.com/kubevirt/containerized-data-importer/releases/download/$VERSION/cdi-cr.yaml
```
