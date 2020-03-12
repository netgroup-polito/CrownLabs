 
## CNI CONFIGURATION NOTES

In order to correctly setup the CNI the yaml file have been modified to use the address range 172.16.0.0/16 as pod network cidr

```sh
...
- name: CALICO_IPV4POOL_CIDR
              value: "172.16.0.0/16"
...
```

Instead the service address range is 10.96.0.0/16
