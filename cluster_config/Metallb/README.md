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

To see which physical node is currently in charge of the LoadBalancer IP you can run the following command, which gets a description of the service:
````
$ k describe svc <name-of-service> -n <service-namespace>
````
and check events label. This is an example:
````
Events:
  Type    Reason        Age                    From             Message
  ----    ------        ----                   ----             -------
  Normal  nodeAssigned  3m58s (x211 over 23h)  metallb-speaker  announcing from node "vinod-0-3"
````
