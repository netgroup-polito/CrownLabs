# Kyverno policies for sandbox namespaces in CrownLabs

This folder contains a set of policies restricting the operations that can be performed in sandbox namespaces. They address three main security requirements, grouped by the “least privilege principle”:

- Pod Security Standards
- Avoid creation of services like load balancers (i.e., avoid creation of routable IPs) and node ports.
- Force specific names for ingress hostname.

The above mentioned policies were mostly taken from [Kyverno Best Practices](https://kyverno.io/policies/?policytypes=Best%2520Practices).

## Policies

A Kyverno policy is a collection of rules. Each rule consists of a match declaration, an optional exclude declaration, and one of a validate, mutate, generate, or verifyImages declaration. Each rule can contain only a single validate, mutate, generate, or verifyImages child declaration.
This initial set of policies is of the resources validation type. When a new resource is created by a user or process into the sandbox namespace, the properties of that resource are checked by Kyverno against the validation rules. If those properties are validated, meaning there is agreement, the resource is allowed to be created. If those properties are different, the creation is blocked.

### Pod Security Standards

Pods are configured to follow security best practices:

- `privileged` is set to `false`
- `spec.hostNetwork`, `spec.hostIPC`, and `spec.hostPID` must be unset
- `spec.volumes[*].hostPath` must be unset
- `spec.containers[*].ports[*].hostPort`, `spec.initContainers[*].ports[*].hostPort`, and `spec.ephemeralContainers[*].ports[*].hostPort` must be unset

### Avoid creation of load balancer and nodePort services

- **Disallow Service Type LoadBalancer**: This policy restricts use of the Service type LoadBalancer.
- **Disallow NodePort**: This policy validates that any new Services do not use the `NodePort` type.

### Force specific name for ingress hostname

- **Disallow empty Ingress host**: This policy ensures that there is a hostname for each rule defined.
- **Restrict Ingress host**: This policy ensures thatthe hostname has the required format.
