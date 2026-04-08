# Monitoring

Cluster Monitoring is the process of assessing the performance of cluster entities either as individual nodes or as a collection. Cluster Monitoring should be able to provide information about the communication and interoperability between various nodes of the cluster.

Monitoring collects [Kubernetes manifests](https://kubernetes.io/docs/home/) and [Helm Charts](https://helm.sh/) to easy operate end-to-end Kubernetes cluster monitoring with [Grafana dashboard](https://grafana.com/), [Prometheus](https://prometheus.io/) using the Prometheus Operator, [Thanos](https://thanos.io/) and [Loki](https://grafana.com/oss/loki/) with [Grafana Alloy](https://grafana.com/docs/alloy/latest/).

Monitoring is divided in two sub-directories:
- [kube-prometheus-stack](./kube-prometheus-stack) is dedicated to metrics collection. Metrics are a measurement for the system, they indicate the use of the application and the ability of the platform.

- [Grafana Alloy](https://grafana.com/docs/alloy/latest/) is dedicated to logs collection. Logs are configured to trace all events that can be used to understand the activity of the system and to diagnose problems.

Note that in the past CrownLabs relied on [Promtail](https://grafana.com/docs/loki/v3.6.x/send-data/promtail/) for log collection, before switching to [Grafana Alloy](https://grafana.com/docs/alloy/latest/).
For a long time, Promtail was the go-to tool for collecting and sending logs to Grafana Loki.
As Promtail has officially entered its Long-Term Support (LTS) phase, the future of log collection now lies with Grafana Alloy, a single, unified agent built for logs, metrics, and traces.
