 
## CNI CONFIGURATION NOTES

In order to correctly setup the cluster the `calico.yaml` confiuration file of the CNI have been slightly modified, in particular the pod network CIDR have been changed, as it's shown in the following snippet: 

```sh
...
- name: CALICO_IPV4POOL_CIDR
              value: "172.16.0.0/16"
...
```

Once the `calico.yaml` file is applied  

```sh
kubectl apply -f calico.yaml
```

we end up with the following configuration of the cluster:
 - IP addresses of the workers: 192.168.31.[101-104] (outside CALICO configuration)
 - IP addresses of the pods: 172.16.0.0/16
 - IP addresses of the services: 10.96.0.0/12