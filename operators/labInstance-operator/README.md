# Laboratory Instance Operator (LabOperator)

Based on [Kubebuilder 2.3](https://github.com/kubernetes-sigs/kubebuilder.git), the operator implements the backend logic of Crownlabs

## CRDs

### Laboratory Templates (LabTemplate)

### Laboratory Instances (LabInstance)

## Build from source

LabOperator requires Golang 1.13 and make. To build the operator:

```bash
cd operators/labInstance-operator
go build
```

## Deployment

## Pre-requisites

The only LabOperator requirements is to have Kubevirt 0.27 should be deployed on the target cluster.
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

### Development

To locally start the operator, you have to deploy the Custom Resource Definitions (CRDs) on the target cluster:

```bash
make install
make install-lab-template
```

Then, launch your operator:

```bash
make run
```
