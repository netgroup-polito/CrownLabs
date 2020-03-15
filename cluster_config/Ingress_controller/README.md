# Ingress controller

We use `ngnix` as ingress controller, to dispatch incoming requests to the requested service running in the cluster.
The ingress controller is coupled with a LoadBalancer service, in order to have this *global frontend* of the cluster reachable through an external (public) IP address.


## SETUP INFORMATION
In order to deploy the ingress controller, we have to apply the following mandatory file, containing the `ngnix ingress controller` deployment and all the configuration for the service itself:

```sh
kubectl apply -f mandatory.yaml
```

Now we can create the LoadBalancer service for the ingress controller, i.e., the external IP address that will be used to reach this service:

```sh
kubectl apply -f ingress_nginx.yaml
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
kubectl apply -f ingress.yaml
```

The file `ingress.yaml` contains all the ingress rules for Prometheus, Grafana and Alert Manager.