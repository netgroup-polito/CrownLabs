# Ingress controller

We use `ngnix` as ingress controller, to dispatch incoming requests to the requested service running in the cluster.
The ingress controller is coupled with a LoadBalancer service, in order to have this *global frontend* of the cluster reachable through an external (public) IP address.


## Setup informations
In order to deploy the ingress controller, we have to apply the following mandatory file, containing the `ngnix ingress controller` deployment and all the configuration for the service itself:

```sh
kubectl apply -f mandatory.yaml
```

Now we can create the LoadBalancer service for the ingress controller, i.e., the external IP address that will be used to reach this service:

```sh
kubectl apply -f svc-ingress-nginx.yaml
```

Once the LB service is created, we must check which IP address have been assigned to it: 

```sh
kubectl get svc -n ingress-nginx -o wide
```

The output will be similar to the following one, where in this case the external IP is `192.168.31.136`:

```sh
NAME            TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)                                     AGE
ingress-nginx   LoadBalancer   10.104.98.160   192.168.31.136   80:31718/TCP,443:30654/TCP,4443:30423/TCP   60m
```

Once everything is checked we can apply the ingress rules:

```sh
kubectl apply -f ingress-monitoring.yaml
```

The file `ingress-monitoring.yaml` contains all the ingress rules for Prometheus, Grafana and Alert Manager. Once it's applied the three services are connected to the ingress controller, associated to a public name (exposed to the internet) and reachable from aoutside the cluster.

## Exposing ingress controller metrics
The `ingress controller` exposes itself some very interesting metrics that can be collected using Prometheus and monitored using Grafana. In order to collect them the first thing to do in to create another service (this time a clusterIP service) and to connect it to the `ingress controller` via the command:

```sh
kubectl apply -f svc-ingress-metrics.yaml
```

This creates the clusterIP service that exposes the port 10254 of the ingress controller for Prometheus.

Once the service is created the last thing to do is to create a ServiceMonitor object in order to make Prometheus aware that our ingress controller is willing to expose his metrics. This can be easily done using the command:

```sh
kubectl apply -f servicemonitor-ingress.yaml
```

## Exposing kube-apiserver
Not only third party application can be connected to the `ingress controller`, but also some Kubernetes object as well. This is the case of the kube-apiserver. 

In order to do so the fist thing to do is to create a service and connect it to the apiserver using the command:

```sh
kubectl apply -f svc-apiserver.yaml
```

Once it's done the last thing to do is to apply the ingress rules for our new service:

```sh
kubectl apply -f ingress-apiserver.yaml
```