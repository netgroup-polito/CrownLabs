## CNI Setup

In order to correctly setup the cluster, the [calico.yaml](calico.yaml) configuration file of the CNI have been slightly modified, in particular the pod network CIDR as shown in the following snippet: 

```sh
...
- name: CALICO_IPV4POOL_CIDR
              value: "172.16.0.0/16"
...
```

Now, apply the [calico.yaml](calico.yaml) file:

```sh
kubectl apply -f calico.yaml
```

This will setup Calico with the following networking configuration of the cluster:
 - IP addresses of pods: 172.16.0.0/16
 - IP addresses of services: 10.96.0.0/12
 
IP addresses of the worker nodes are outside CALICO configuration.
