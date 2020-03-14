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
This configuration creates two addresses pools: one private and one public.
 - private: 192.168.31.[135-199]
 - public: 130.192.31.[240-244]

In order to create a service of type LoadBalancer with a public ip you must add the following annotation:
````
 metallb.universe.tf/address-pool: public
````
If this annotation is omitted metallb will choose a private ip.

To check which node has been assigned to LoadBalancer ip you can run the following command to get a description of a service:
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