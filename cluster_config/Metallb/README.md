# METALLB
MetalLB is a load-balancer implementation for bare metal Kubernetes clusters, using standard routing protocols.
## Install MetalLB
Run this command to install MetalLB. After this command it will stay in pending state waiting for a ConfigMap- 

````
$ kubectl apply -f https://raw.githubusercontent.com/google/metallb/v0.8.3/manifests/metallb.yaml
````

## Configuration
The configmap.yaml contains the ConfigMap with addresses that are used by MetalLB to expose services.

We assigned addresses from 192.168.31.135 to 192.168.31.199 

````
$ kubectl apply -f configmap.yaml
````
