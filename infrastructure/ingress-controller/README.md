# Ingress Controller - nginx

We use [*ingress-ngnix*](https://kubernetes.github.io/ingress-nginx/) as ingress controller, to dispatch incoming requests to the requested service running in the cluster.
In the following, two different ingress controllers are created and coupled with two LoadBalancer services.
One operates as a *global frontend* of the cluster reachable through an external (public) IP address, while the other can be adopted to expose resources on the *internal network* (i.e. with a private IP address).

## Installation and configuration

We use Helm to install and configure the two instances of ingress-nginx.
First, it is necessary to add the `ingress-nginx` repository:

```sh
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
```

Then, it is possible to proceed with the actual installation:

```sh
helm upgrade ingress-nginx-external ingress-nginx/ingress-nginx --namespace ingress-nginx-external \
    --install --create-namespace --values configurations/ingress-nginx-external.yaml
helm upgrade ingress-nginx-internal ingress-nginx/ingress-nginx --namespace ingress-nginx-internal \
    --install --create-namespace --values configurations/ingress-nginx-internal.yaml
```

Finally, it is necessary to label the namespace of the *external* ingress controller to allow the requests
reaching the user instances, according to the respective network policies:

```sh
kubectl label namespace ingress-nginx-external crownlabs.polito.it/allow-instance-access=true
```

## Selecting the ingress controller

By default, every `Ingress` resource is attached to the IngressClass `nginx-external`, which is exposed on the external network (with public IP).
To select the Ingress Controller exposed on the internal network, it is necessary to specify the `nginx-internal` class inside the ingress definition:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-name
  namespace: ingress-namespace
spec:
  ingressClassName: nginx-internal
  rules:
    ...
```

## Exposing kube-apiserver

Not only third party application can be connected to the `ingress controller`, but also some Kubernetes object as well. This is the case of the kube-apiserver.

In order to do so the fist thing to do is to create a service and connect it to the apiserver using the command:

```sh
$ kubectl apply -f api-server/svc-apiserver.yaml
```

Once it's done the last thing to do is to apply the ingress rules for our new service:

```sh
$ kubectl apply -f api-server/ingress-apiserver.yaml
```

## Custom error pages

The `nginx ingress controller` provides the possibility to configure a default backend.
Its main functionality is to serve custom error pages whenever an error occurs.

In order to do so, in is necessary to create and upload a Docker image containing the server responsible for this task.
Please refer to [custom-error-pages](custom-error-pages/README.md) for more information.

The custom default backend is enabled by default through the Helm values file of the *external* ingress controller, while it is not configured for the internal one.
It is possible to opt-out (or select a different subset of error codes) on a per-ingress basis adding the appropriate annotation:

```yaml
annotations:
    nginx.ingress.kubernetes.io/custom-http-errors: "418"
```
