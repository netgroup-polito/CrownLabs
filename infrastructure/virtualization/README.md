# Virtualization

[KubeVirt](https://kubevirt.io/) is a solution to spawn and orchestrate traditional virtual machines on top of a Kubernetes cluster.

## How to install

The following presents a summary of the basic operations required to deploy KubeVirt in a Kubernetes cluster. Please, refer to the [official instructions](https://kubevirt.io/user-guide/#/installation/installation) for more details about customizing the configuration.

```bash
# Pick an upstream version of KubeVirt to install
$ export KUBEVIRT_VERSION=v0.35.0
# Deploy the KubeVirt operator
$ kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-operator.yaml
# Create the KubeVirt CR (instance deployment request) which triggers the actual installation
$ kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml
# wait until all KubeVirt components are up
$ kubectl -n kubevirt wait kv kubevirt --for condition=Available
```
