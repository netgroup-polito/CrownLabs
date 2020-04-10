# METALLB
MetalLB is a load-balancer implementation for bare metal Kubernetes clusters, using standard routing protocols.

## Install MetalLB
Run the following command to install MetalLB:

````
$ kubectl apply -f https://raw.githubusercontent.com/google/metallb/v0.8.3/manifests/metallb.yaml
````

After this command, MetalLB remains in pending state waiting for a ConfigMap (see next step).

## Configuration
File [configmap.yaml](configmap.yaml) contains the ConfigMap with the set of IP addresses that are used by MetalLB to expose services.
Addresses (which are visible in the proper section of [configmap.yaml](configmap.yaml)) are applied with this command:

````
$ kubectl apply -f configmap.yaml
````
In particular, the above  configuration creates two addresses pools: one private (192.168.31.[135-199]) and one public (130.192.31.[240-244]), which can be modified in other setup.

Given the presence of *two* address pools, a service of type LoadBalancer with a public IP requires the following annotation:

````
 metallb.universe.tf/address-pool: public
````
If this annotation is omitted metallb will choose a private IP.

To see which physical node is currently in charge of a given LoadBalancer IP, you can go through the following steps:

**1.** Identify the service that has currently asked for a given LoadBalancer IP:

````
$ kubectl get services -A
NAMESPACE       NAME                TYPE           CLUSTER-IP       EXTERNAL-IP   
....
ingress-nginx   ingress-nginx       LoadBalancer   10.110.183.89    130.192.31.241
````
In this case, the service is `ingress-nginx`, in a namespace `ingress-nginx`.

**2.** Get a description of that service:

````
$ k describe svc <name-of-service> -n <service-namespace>
````
where  `<name-of-service> ` and  `<service-namespace> ` are the ones obtained in the previous step.

The information you are looking for is in the `events` label of the previous output, such as in the following:
````
Events:
  Type    Reason        Age                    From             Message
  ----    ------        ----                   ----             -------
  Normal  nodeAssigned  3m58s (x211 over 23h)  metallb-speaker  announcing from node "vinod-0-3"
````
