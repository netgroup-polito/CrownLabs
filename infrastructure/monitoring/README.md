# Monitoring

Cluster Monitoring is the process of assessing the performance of cluster entities either as individual nodes or as a collection. Cluster Monitoring should be able to provide information about the communication and interoperability between various nodes of the cluster.

Monitoring collects [Kubernetes manifests](https://kubernetes.io/docs/home/) and [Helm Charts](https://helm.sh/) to easy operate end-to-end Kubernetes cluster monitoring with [Grafana dashboard](https://grafana.com/), [Prometheus](https://prometheus.io/) using the Prometheus Operator, [Thanos](https://thanos.io/) and [Loki](https://grafana.com/oss/loki/) with [Promtail agent](https://grafana.com/docs/loki/latest/clients/promtail/).

Monitoring is divided in two sub-directories:
- [kube-prometheus-stack](./kube-prometheus-stack) is dedicated to metrics collection. Metrics are a measurement for the system, they indicate the use of the application and the ability of the platform.
  
- [Loki Promtail](./loki-promtail) is dedicated to logs collection. Logs are configured to trace all events that can be used to understand the activity of the system and to diagnose problems.