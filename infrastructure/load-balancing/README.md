# Load Balancing - MetalLB

[MetalLB](https://metallb.universe.tf) is a load-balancer implementation for bare metal Kubernetes clusters, using standard routing protocols.

## Install and Configure MetalLB

MetalLB can be easily installed and configured with Helm:

```bash
helm repo add metallb https://metallb.github.io/metallb --namespace metallb-system \
    --install --create-namespace --values metallb-values.yaml
```

Among the different configurations, the [values file](./metallb-values.yaml) specifies the set of address pools managed by MetalLB, along with the announce mode (i.e., Layer2 or BGP).
Currently, we configured MetalLB to announce two pools, one with private addresses and one with public addresses, in both cases leveraging the BGP mode.

## Configure LoadBalancer Services

With the given configuration, a service of type LoadBalancer is assigned by default an IP from the private pool.
A public IP, on the other hand, can be requested adding an appropriate annotation:

```yaml
annotations:
    metallb.universe.tf/address-pool: public
```

**Note:** when leveraging the BGP mode, it is appropriate to configure the service `ExternalTrafficPolicy` to `Local`, to ensure traffic is load balanced only across those nodes that are currently hosting the service.
Hence, preventing “horizontal” traffic flow between nodes and avoiding source IP modifications.
Please refer to the [official documentation](https://metallb.universe.tf/usage/#bgp) for additional information.

## Debugging

To see which physical nodes are currently announcing the IP of a LoadBalancer service, you can leverage:

```bash
kubectl describe svc <name-of-service> -n <service-namespace>
```

The *Events* sections presents the information of interest:

```txt
Events:
  Type    Reason                 Age                   From                Message
  ----    ------                 ----                  ----                -------
  Normal  nodeAssigned           10m (x2 over 11m)     metallb-speaker     announcing from node "worker-1"
  Normal  nodeAssigned           10m (x2 over 11m)     metallb-speaker     announcing from node "worker-4"
  Normal  nodeAssigned           10m (x2 over 10m)     metallb-speaker     announcing from node "worker-6"

```
