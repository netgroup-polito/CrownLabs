# Laboratory Instance Operator (LabOperator)

Based on [Kubebuilder 2.3](https://github.com/kubernetes-sigs/kubebuilder.git), the operator implements the backend logic of Crownlabs

# Basic Functioning

## CRDs

The Laboratory Operator (LabOperator) implements the backend logic necessary to spawn new laboratories starting from a predefined template. LabOperator relies on two Kubernetes Custom Resource
Definitions (CRDs) which implement the basic APIs:
* **Laboratory Template (LabTemplate)** defines the size of the execution environment (e.g.; Virtual Machine), its base image and a description. This object is created by professors and read by students, while creating new instances.
* **Laboratory Instance (LabInstance)** defines an instance of a certain template. The manipulation of those objects triggers the reconciliation logic in LabOperator, which creates/destroy associated resources (e.g.; Virtual Machines).

A *LabInstance* resource triggers the creation of the following components:
* Kubevirt VirtualMachine Instance and the logic to access the noVNC instance inside the VM (Service, Ingress)
* An instance of [Oauth2 Proxy](https://github.com/oauth2-proxy/oauth2-proxy) (Deployment, Service, Ingress) to regulate access to the VM.

All those resources are binded to the LabInstance life-cycle via the [OwnerRef property](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/)

Both LabTemplates and LabInstances are **namespaced**. 

## Installation

### Pre-requirements

The only LabOperator requirement is to have Kubevirt 0.27 deployed.
This can be done with the following commands, as reported by the official website:

```bash
# On other OS you might need to define it like
export KUBEVIRT_VERSION="v0.27.0"

kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml 
# Only if HW Virtualization is not available
kubectl create configmap kubevirt-config -n kubevirt --from-literal debug.useEmulation=true
# Deploy Kubevirt
kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml
```

### Install the CRDs

Before the deploying the operator, we have to add the LabInstance and LabTemplate CRDs. This can be done via the Kubebuilder-provided Makefile:

```bash
make install
make install-lab-template
```

or directly via the commands:

```bash
kubectl kustomize config/crd | kubectl apply -f -
kubectl kustomize config/crd | kubectl delete -f -
```

### Deployment
To deploy the LabOperator in your cluster, you have to do the following steps.

First, set the desired values in `operators/labInstance-operator/k8s/operator/configmap.yaml`.

After the values have been correctly set for your environment. You can deploy the labOperator using:

```
kubectl create ns lab-operator
kubectl apply -f operators/labInstance-operator/k8s
```

## Development

### Build from source

LabOperator requires Golang 1.13 and make. To build the operator:

```bash
cd operators/labInstance-operator
go build
```

### Testing

After having installed Kubevirt in your testing cluster, you have to deploy the Custom Resource Definitions (CRDs) on the target cluster:

```bash
make install
make install-lab-template
```

Then, you can launch locally your operator:

```bash
make run
```

N.B. So far, the readiness check for VirtualMachines is performed by assuming that the operator is running on the same cluster of the Virtual Machines. This prevents the possibility to have *ready* VMs when testing the operator outside the cluster. 
