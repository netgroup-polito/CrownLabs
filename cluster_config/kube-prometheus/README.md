# kube-prometheus

## Manifests

The manifests contain the definition of namespace (monitoring), the various CRD (Custom Resource Definitions), Prometheus rules, dashboards and so on.

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


### Persistent storage
We have modified two manifests (for grafana and prometheus) to have persistent storage.

* grafana-deployment.yaml
```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pv-claim-grafana
  labels:
    app: grafana
spec:
  storageClassName: rook-ceph-block
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
```
* grafana-deployment.yaml
```
 volumes:
      
      - name: grafana-storage
        persistentVolumeClaim:
          claimName: pv-claim-grafana
```
* prometheus-prometheus.yaml
```
retention: 15d
  resources:
    requests:
      memory: 2Gi
```

* prometheus-prometheus.yaml: before applying your cluster configuration you have to enter the correct value
```
externalUrl: <>
```
* prometheus-prometheus.yaml
```
storage:
    volumeClaimTemplate:
      metadata:
        annotations:
          name: prometheus-storage
      spec:
        resources:
          requests:
            storage: 50Gi
        accessModes:
          - ReadWriteOnce
        storageClassName: rook-ceph-block
  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 99
        podAffinityTerm:
          labelSelector:
            matchExpressions:
            - key: app
              operator: In
              values:
              - prometheus
          topologyKey: kubernetes.io/hostname
```



### Other information
Github of [kube-prometheus](https://github.com/coreos/kube-prometheus).
