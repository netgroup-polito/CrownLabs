# About Kyverno

[Kyverno](https://kyverno.io/) is a policy engine designed specifically for Kubernetes. With Kyverno, policies are managed as Kubernetes resources and no new language is required to write policies. This allows using familiar tools such as kubectl, git, and kustomize to manage policies.

## How Kyverno works

Kyverno runs as a dynamic admission controller in a Kubernetes cluster. Kyverno receives validating and mutating admission webhook HTTP callbacks from the kube-apiserver and applies matching policies to return results that enforce admission policies or reject requests.

### Installing Kyverno

Before installing the chart, the Kyverno repository must be added to helm with the following command:

```bash
helm repo add kyverno https://kyverno.github.io/kyverno/
helm repo update
```

Then, it is possible to install Kyverno through:

```bash
helm install kyverno kyverno/kyverno --namespace kyverno --create-namespace --values kyverno-values.yaml
```
