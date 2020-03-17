# kube-prometheus
Kube-prometheus collects Kubernetes manifests, Grafana dashboards, and Prometheus rules to provide easy to operate end-to-end Kubernetes cluster monitoring with Prometheus using the Prometheus Operator.


## Table of contents

- [kube-prometheus](#kube-prometheus)
  - [Table of contents](#table-of-contents)
  - [Introduction](#introduction)
    - [Monitoring](#monitoring)
    - [Prometheus](#prometheus)
    - [Grafana](#grafana)
    - [Alertmanager](#alertmanager)
  - [Install](#install)
    - [Manifests](#manifests)
    - [Quickstart](#quickstart)
    - [Persistent storage](#persistent-storage)
    - [Grafana OAuth2 Authentication](#grafana-oauth2-authentication)
  - [Other information](#other-information)

## Introduction

### Monitoring
Cluster Monitoring is the process of assessing the performance of cluster entities either as individual nodes or as a collection. Cluster Monitoring should be able to provide information about the communication and interoperability between various nodes of the cluster.

### Prometheus
Prometheus is an open-source systems monitoring and alerting toolkit.
Prometheus's main features are:
- a multi-dimensional data model with time series data identified by metric name and key/value pairs
- PromQL, a flexible query language to leverage this dimensionality
- no reliance on distributed storage; single server nodes are autonomous
- time series collection happens via a pull model over HTTP
- pushing time series is supported via an intermediary gateway
- targets are discovered via service discovery or static configuration
- multiple modes of graphing and dashboarding support

### Grafana
Grafana is an open source visualization and analytics software. It allows you to query, visualize, alert on, and explore your metrics no matter where they are stored. We can also say that Grafana is the tool for beautiful monitoring and metric analytics & dashboards for Graphite, InfluxDB & Prometheus & More.

### Alertmanager
The Alertmanager handles alerts sent by client applications such as the Prometheus server. It takes care of deduplicating, grouping, and routing them to the correct receiver integration such as email, PagerDuty, or OpsGenie. It also takes care of silencing and inhibition of alerts


## Install

## Manifests
These manifests contain the most important elements required to monitor the cluster:
- The namespace
- The Prometheus Operator
- Highly available Prometheus
- Highly available Alertmanager
- Prometheus node-exporter
- Prometheus Adapter for Kubernetes Metrics APIs
- kube-state-metrics
- Grafana

## Quickstart
1. Create the monitoring stack using the config in the `manifests` directory:

```shell
# Create the namespace and CRDs, and then wait for them to be available before creating the remaining resources
kubectl create -f manifests/setup
until kubectl get servicemonitors --all-namespaces ; do date; sleep 1; echo ""; done
kubectl create -f manifests/
```

2. Now, teardown the stack:
```shell
kubectl delete --ignore-not-found=true -f manifests/ -f manifests/setup
```


## Persistent storage

### Why?
Running cluster monitoring with persistent storage means that your metrics are stored to a Persistent Volume and can survive a pod being restarted or recreated. This is ideal if you require your metrics or alerting data to be guarded from data loss. For production environments, it is highly recommended to configure persistent storage.

### How?
We need to modify two manifests (for Grafana and Prometheus) to have persistent storage.

1. [Grafana](https://github.com/netgroup-polito/CrownLabs/blob/kube-prometheus/cluster_config/kube-prometheus/manifests/grafana-deployment.yaml) manifest.
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
```
 volumes:

      - name: grafana-storage
        persistentVolumeClaim:
          claimName: pv-claim-grafana
```
2. [Prometheus](https://github.com/netgroup-polito/CrownLabs/blob/kube-prometheus/cluster_config/kube-prometheus/manifests/prometheus-prometheus.yaml) manifest.
```
retention: 15d
  resources:
    requests:
      memory: 2Gi
```
3. Before applying your cluster configuration, you have to enter the correct value for the
```
externalUrl: <>
```
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

## Grafana OAuth2 Authentication
It is possible to configure many different oauth2 authentication services with Grafana using the `generic_oauth` feature. In the following, we will setup Grafana to use Keycloak as identity provider. *Note*: this guide assumes Keycloak to be already deployed and available. Please refer to the [keycloak deployment guide](../Keycloak/README.md) for more information.

### Keycloak configuration

1. Create a new client for Grafana;
2. Configure the Client Protocol to be `openid-connect`;
3. Set Access Type to `confidential`;
4. Configure the Root URL and the Base URL to the grafana URL (e.g. https://grafana.example.com/);
5. Configure the Valid Redirect URIs to the grafana URL (e.g. https://grafana.example.com/*);
6. From the Credentials tab, copy the Client Secret that has been generated.

### Grafana configuration

1. Edit the [grafana-configuration configmap](manifests/grafana-configuration.yaml) and adapt it to your configuration (each line corresponds to an environment variable). In particular, it is necessary to adapt the different URIs and specify the Client ID and Secrets generated in Keycloak. The meaning of the different fields is specified by the embedded comments. More information can be found in the [official documentation](https://grafana.com/docs/grafana/latest/auth/generic-oauth/).
2. Restart the Grafana deployment:
   ```sh
   $ kubectl rollout restart deploy/grafana -n monitoring
   ```

### Limit access to a subset of Keycloak users
**Warning:** At the time of writing, it seems not possible to restrict login by role/group when using the `generic_oauth` feature. Hence, all valid users of the Keycloak realm would be able to access Grafana. As a temporary workaround, the [grafana-configuration configmap](manifests/grafana-configuration.yaml) disables the access to users not already present within the Grafana database. As soon as this inherent [PR](https://github.com/grafana/grafana/pull/22383) is merged and the new version of Grafana released, it should be possible to limit the access to only a subset of Keycloak users directly from grafana, without needing to create duplicated accounts.


## Other information
For more information, look at the Github page of [kube-prometheus](https://github.com/coreos/kube-prometheus).
