# Kubernetes Networking - CNI Setup

As for it concerns Kubernetes networking, we selected [Project Calico](https://www.projectcalico.org/), since it is one of the most popular CNI plugins.
In short, it limits the overhead by requiring no overlay and supports advanced features such as the definition of network policies to isolate the traffic between different containers.

## Calico Installation
In order to install Calico, you can perform the following operations, which will download the default configuration from the official webpage and apply it customizing the pod network CIDR according to the selected cluster setup:

```bash
$ export CALICO_VERSION=v3.16
$ curl https://docs.projectcalico.org/${CALICO_VERSION}/manifests/calico.yaml -o calico.yaml
$ kubectl apply -k .
```

## Selected cluster networking configuration
- IP addresses of pods: 172.16.0.0/16
- IP addresses of services: 10.96.0.0/12
