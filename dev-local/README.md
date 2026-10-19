# Local deployment support files

When developing backend features, it is usually very useful to have a local instance of the CrownLabs infrastructure. This allows the developer to give himself rights he does not have on the real system, and to try modifications in a risk-free environment.

CrownLabs is a complex system, and installing all the required tools is not a straight-forward task. Previously, the installing instructions were scattered across ad-hoc locations or lost in someone's shell history. The files in this folder aim instead to collect all the information in a single place.

## Systems

The CrownLabs infrastructure is composed of the following systems, each one described in a subfolder of this directory.

Please, set them up in the following order, since each one is a prerequisite for the next.

- [`base-k3s/`](base-k3s/README.md): CrownLabs needs Kubernetes to run.
  K3s is a lightweight Kubernetes distribution, that can be easily run on your laptop. The guide includes instructions to both install it from scratch, and to merge multiple kubeconfig files, in case you already have another cluster.
- [`envoy/`](envoy/README.md): this is the new ingress controller for Kubernetes, allowing to correctly expose a service on a URL like `https://<service>.crownlabs.local`.
- [`keycloak/`](keycloak/README.md): the identity and access management solution, providing single-sign-on and authentication for the various services. It is exposed through an `HTTPRoute` on the Gateway from `envoy/`.
- [`operators/`](operators/README.md): the components that constitute the server-side of the CrownLabs business logic.
- [`mailpit/`](mailpit/README.md): a fake SMTP server with a web UI, for local email testing. Like Keycloak, it is exposed through an `HTTPRoute` on the Gateway from `envoy/`.
  While all the other systems are almost always required, this one may be skipped, if the system enhancements you have to do are not related to sending emails.

For each system, the dedicated subfolder contains:

- a `README.md` file, listing the prerequisites, and explaining how to install the system (including what, where and when to deploy).
- a `manifests` folder, containing all the Kubernetes YAML files that should be applied using `kubectl apply`.

    You can apply the manifests in this folder as-is, with `kubectl apply -f`.
    If a resource is already templated in the main `deploy/` Helm chart, render it from that chart instead (`helm template -s ...`).
    Do not keep a hand-maintained, de-templated copy here: it can drift from the original.

## Running the full stack

See [`local-development.md`](local-development.md) for the end-to-end guide.
It shows how to bring up Keycloak, qlkube, the frontend, and the operators locally, once the cluster itself exists.
