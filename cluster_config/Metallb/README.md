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
