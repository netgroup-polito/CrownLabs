# Ingress Controller - nginx

We use `ngnix` as ingress controller, to dispatch incoming requests to the requested service running in the cluster.
In the following, two different ingress controllers are created and coupled with two LoadBalancer services. One operates as a *global frontend* of the cluster reachable through an external (public) IP address, while the other can be adopted to expose resources on the *internal network* (i.e. with a private IP address).


## Setup informations
In order to deploy the two ingress controllers, we have to apply the following files, containing the `ngnix ingress controller` deployments and all the configurations required:

```sh
$ kubectl apply -f manifests/ingress-controller-external.yaml
$ kubectl apply -f manifests/ingress-controller-internal.yaml
```

Now we can create the LoadBalancer services for the ingress controllers, i.e., the external IP addresses that will be used to reach this services:

```sh
$ kubectl apply -f manifests/svc-ingress-nginx-external.yaml
$ kubectl apply -f manifests/svc-ingress-nginx-internal.yaml
```

Once the LB services are created, we can check which IP address have been assigned to them:

```sh
$ kubectl get svc -n ingress-nginx -o wide
$ kubectl get svc -n ingress-nginx-internal -o wide
```

The output should be similar to the following one, where in this case the external IP is `130.192.31.240`:

```sh
NAME            TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)                                     AGE
ingress-nginx   LoadBalancer   10.104.98.160   130.192.31.240   80:31718/TCP,443:30654/TCP,4443:30423/TCP   60m
```

## Selecting the ingress controller
By default, every `Ingress` resource is attached to the Ingress Controller exposed on the external network (with public IP). To select the Ingress Controller exposed on the internal network, it is necessary to add the ad-hoc annotation and specify the `nginx-internal` class:
```yaml
  annotations:
    kubernetes.io/ingress.class: "nginx-internal"
```

## Exposing ingress controller metrics
The `ingress controller` exposes itself some metrics that can be collected using Prometheus and monitored using Grafana. In order to collect them the first thing to do in to create another service (this time a clusterIP service) and to connect it to the `ingress controller` via the command:

```sh
$ kubectl apply -f manifests/svc-ingress-metrics-external.yaml
$ kubectl apply -f manifests/svc-ingress-metrics-internal.yaml
```

This creates the clusterIP service that exposes the port 10254 of the ingress controller for Prometheus.

Once the service is created the last thing to do is to create a ServiceMonitor object in order to make Prometheus aware that our ingress controller is willing to expose his metrics. This can be easily done using the command:

```sh
$ kubectl apply -f manifests/servicemonitor-ingress-external.yaml
$ kubectl apply -f manifests/servicemonitor-ingress-internal.yaml
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
The `nginx ingress controller` provides the possibility to configure a default backend. Its main functionality is to serve custom error pages whenever an error occurs.

In order to do so, in is necessary to create and upload a Docker image containing the server responsible for this task as well as the necessary resources. Please refer to [custom-error-pages](custom-error-pages/README.md) for more information.

Once the image is available, it is possible to create the necessary resources (i.e. deployments, services, ...) through the ad-hoc manifest:
```bash
$ kubectl apply -f manifests/custom-error-pages.yaml
```

Finally, the custom backend can be globally enabled by adding the corresponding flag to the container arguments of the `nginx` deployment:
```
- --default-backend-service=ingress-nginx/nginx-custom-error-pages
```
and configuring the errors to be managed through the `nginx-configuration` `configmap`:
```
custom-http-errors: 400,401,402,403,404,405,406,407,408,409,410,411,412,413,414,415,416,417,418,421,422,423,424,425,426,428,429,431,451,500,501,502,503,504,505,506,507,508,510,511
```

## Securing the ingress controller with ModSecurity

[ModSecurity](https://modsecurity.org/) is an open source, cross-platform web application firewall (WAF) module. Known as the "Swiss Army Knife" of WAFs, it enables web application defenders to gain visibility into HTTP(S) traffic and provides a power rules language and API to implement advanced protections.

### Global configuration
ModSecurity is enabled globally through the [nginx-configuration](manifests/ingress-controller-external.yaml) ConfigMap. By default, it operates in Audit only mode; hence, the violations detected are logged but not blocked. Finally, the ConfigMap introduces also some whitelisting rules to prevent legitimate requests from being blocked.

### Per-Ingress configuration
ModSecurity can be configured in enforcement mode by setting an ad-hoc annotation to the Ingress resource:
```yaml
annotations:
    nginx.ingress.kubernetes.io/configuration-snippet: |
        modsecurity_rules '
            SecRuleEngine On
        ';
```

### Additional references
1. [Enabling ModSecurity in the Kubernetes Ingress-NGINX Controller](https://awkwardferny.medium.com/enabling-modsecurity-in-the-kubernetes-ingress-nginx-controller-111f9c877998)
2. [Creating an OpenWAF solution with Nginx, ElasticSearch and ModSecurity](https://karlstoney.com/2018/02/23/nginx-ingress-modsecurity-and-secchatops/)
3. [ModSecurity does not block request, only logs, while SecRuleEngine is set to On](https://github.com/kubernetes/ingress-nginx/issues/4385)
