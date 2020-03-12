# kube-prometheus

## Quickstart
* Create the monitoring stack using the config in the `manifests` directory:

```shell
# Create the namespace and CRDs, and then wait for them to be availble before creating the remaining resources
kubectl create -f manifests/setup
until kubectl get servicemonitors --all-namespaces ; do date; sleep 1; echo ""; done
kubectl create -f manifests/
```

 * And to teardown the stack:
```shell
kubectl delete --ignore-not-found=true -f manifests/ -f manifests/setup
```

### Access the dashboards
We use an ingress controller to access to these services.

Prometheus: access via [https://prometheus-ladispe.ipv6.polito.it](https://prometheus-ladispe.ipv6.polito.it)

Grafana: access via [https://grafana-ladispe.ipv6.polito.it](https://grafana-ladispe.ipv6.polito.it)

Alert Manager: access via [https://alertmanager-ladispe.ipv6.polito.it](https://alertmanager-ladispe.ipv6.polito.it)



### Persistent storage
We have modified two posters (for grafana and prometheus) to have persistent storage.

### Other information
[https://github.com/coreos/kube-prometheus](https://github.com/coreos/kube-prometheus).




