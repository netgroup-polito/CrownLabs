# Monitoring - kube-prometheus
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
    - [Alert Notification](#alert-notification)
    - [Monitor the Bind DNS Server](#monitor-the-bind-dns-server)
    - [Monitor all namespaces](#monitor-all-namespaces)
    - [Blackbox monitoring](#blackbox-monitoring)
    - [OAuth2 Authentication](#oauth2-authentication)
  - [Other information](#other-information)

## Introduction

### Monitoring
Cluster Monitoring is the process of assessing the performance of cluster entities either as individual nodes or as a collection. Cluster Monitoring should be able to provide information about the communication and interoperability between various nodes of the cluster.

### Prometheus
From [Promtheus](https://prometheus.io/docs/introduction/overview/) official documentation.
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
From [Grafana](https://grafana.com/docs/grafana/latest/getting-started/#what-is-grafana) official documentation.
Grafana is an open source visualization and analytics software. It allows you to query, visualize, alert on, and explore your metrics no matter where they are stored. We can also say that Grafana is the tool for beautiful monitoring and metric analytics & dashboards for Graphite, InfluxDB & Prometheus & More.

### Alertmanager
From Promtheus [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) official documentation.
The Alertmanager handles alerts sent by client applications such as the Prometheus server. It takes care of deduplicating, grouping, and routing them to the correct receiver integration such as email, PagerDuty, or OpsGenie. It also takes care of silencing and inhibition of alerts


## Install

### Manifests
These manifests contain the most important elements required to monitor the cluster:
- The namespace
- The Prometheus Operator
- Highly available Prometheus
- Highly available Alertmanager
- Prometheus node-exporter
- Prometheus Adapter for Kubernetes Metrics APIs
- kube-state-metrics
- Grafana

### Quickstart
1. Create the monitoring stack using the config in the `manifests` directory:

```bash
# Create the namespace and CRDs, and then wait for them to be available before creating the remaining resources
$ kubectl create -f manifests/setup
$ until kubectl get servicemonitors --all-namespaces ; do date; sleep 1; echo ""; done
$ kubectl create -f manifests/
```

2. If you want to teardown the stack:
```bash
$ kubectl delete --ignore-not-found=true -f manifests/ -f manifests/setup
```


### Persistent storage

#### Why?
Running cluster monitoring with persistent storage means that your metrics are stored to a Persistent Volume and can survive a pod being restarted or recreated. This is ideal if you require your metrics or alerting data to be guarded from data loss. For production environments, it is highly recommended to configure persistent storage.

#### How?
We need to modify two manifests (for Grafana and Prometheus) to have persistent storage.

1. [Grafana](manifests/grafana-deployment.yaml) manifest.
```yaml
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
```yaml
 volumes:

      - name: grafana-storage
        persistentVolumeClaim:
          claimName: pv-claim-grafana
```
2. [Prometheus](manifests/prometheus-prometheus.yaml) manifest.
```yaml
retention: 15d
  resources:
    requests:
      memory: 2Gi
```
3. Before applying your cluster configuration, you have to enter the correct value for the
```yaml
externalUrl: <>
```
```yaml
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
 ```
 ```yaml
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

### Monitor all namespaces
We need to give permission to the pod prometheus to be able to get on endpoints that it does not know. To know them he has to talk with his APIserver and to do so he need an identity i.e. a ServiceAccount, finally the permissions are needed. To do this we use the concepts of ClusterRole and ClusterRoleBinding.
So
```bash
$ kubectl apply -f prometheus-scraper-cluster-role.yaml
```
Now Prometheus can scrape all the namespaces in the cluster.


### Blackbox monitoring
Testing externally visible behavior as a user would see it.

It allows us to tell if someone is seeing us or not and it is useful if something unexpected happens. We do the scraping from the outside in order to see that there are no access problems, for some reason, perhaps due to some crash.

#### Blackbox exporter
From [Blackbox exporter](https://github.com/prometheus/blackbox_exporter) official Github.
The blackbox exporter allows blackbox probing of endpoints over HTTP, HTTPS, DNS, TCP and ICMP.
TO deploy it, we use the [blackbox exporter chart](https://github.com/helm/charts/tree/master/stable/prometheus-blackbox-exporter) by appropriately changing the values of [values](https://github.com/helm/charts/blob/master/stable/prometheus-blackbox-exporter/values.yaml) file.

### Alert Notification
When an alert is sent to AlertManager, those are also sent to Slack.
This template differentiates alerts based on severity and sends them on the correct Slack channel.

#### How to install the template
To install this template, you have to follow the steps below:

1) Configure the fields `api_url` in [alertmanager-slack.yaml](alertmanager-templates/alertmanager-slack.yaml) with your own Slack hook(s).

2) Then encode the above template in base64 (in our case, `<file-template>` is `alertmanager-templates/alertmanager-slack.yaml`):
```bash
$ cat <file-template> | base64 -w0
```

3) Now, you have to edit the secrets of your alertmanager deployment and add the above output (i.e., the entire content of `alertmanager-slack.yaml`, encoded in based 64) as a *secret* in correspondence of field `alertmanager-slack.yaml`. The above command will open an editor that will allow to complete this action:
```bash
$ kubectl edit secrets -n <alertmanager namespace> <alertmanager secret name> -o yaml
```

### OAuth2 Authentication
In the following, we will setup Alertmanager, Grafana and Prometheus to use Keycloak as identity provider for the authentication. Grafana can natively use Keycloak as identity provider, while the authentication for the other two services is managed through the ingress controller and [oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy). *Note*: this guide assumes Keycloak to be already deployed and available. Please refer to the [keycloak deployment guide](../identity-provider/README.md) for more information.

#### Keycloak configuration

1. Create a new client for `monitoring`;
2. Configure the Client Protocol to be `openid-connect`;
3. Set Access Type to `confidential`;
4. Configure the Valid Redirect URIs to the alertmanager, grafana and prometheus URLs (e.g. https://alertmanager.example.com/*, https://grafana.example.com/*, https://prometheus.example.com/*); <!-- markdown-link-check-disable-line -->
5. From the Credentials tab, copy the Client Secret that has been generated;
6. From the Mappers tab, add a new `Group Membership` mapper with Token Claim Name equal to `groups`.

#### Alertmanager and Prometheus configuration

1. Generate a new Cookie Secret:
    ```sh
    python -c 'import os,base64; print(base64.urlsafe_b64encode(os.urandom(16)).decode())'
    ```
2. Edit the [monitoring-oauth2-proxy deployment](manifests/monitoring-oauth2-proxy-deployment.yaml) and adapt it to your configuration. In particular, it is necessary to adapt the different URIs, specify the Cookie Secret previously created and the Client ID and Secret generated in Keycloak. The meaning of the different fields is specified by the embedded comments.
3. Configure the `Ingress` objects to perform OAuth2 authentication. See [alertmanager-ingress.yaml](manifests/alertmanager-ingress.yaml), [alertmanager-ingress-oauth2.yaml](manifests/alertmanager-ingress-oauth2.yaml), [prometheus-ingress.yaml](manifests/prometheus-ingress.yaml), [prometheus-ingress-oauth2.yaml](manifests/prometheus-ingress-oauth2.yaml) for the complete configuration.

#### Grafana configuration

1. Edit the [grafana-configuration configmap](manifests/grafana-configuration.yaml) and adapt it to your configuration (each line corresponds to an environment variable). In particular, it is necessary to adapt the different URIs and specify the Client ID and Secrets generated in Keycloak. The meaning of the different fields is specified by the embedded comments. More information can be found in the [official documentation](https://grafana.com/docs/grafana/latest/auth/generic-oauth/).
2. Apply the `ConfigMap`:
   ```sh
   $ kubectl create -f manifests/grafana-configuration.yaml
   ```
3. Restart the Grafana deployment:
   ```sh
   $ kubectl rollout restart deploy/grafana -n monitoring
   ```

#### Limit access to a subset of Keycloak users

The access to Alertmanager and Prometheus is limited to users belonging to the `monitoring` group. Hence, it is possible to grant access to the users by simply adding them to the `monitoring` group.

**Warning:** At the time of writing, in Grafana it seems not possible to restrict login by role/group when using the `generic_oauth` feature. Hence, all valid users of the Keycloak realm would be able to access Grafana. As a temporary workaround, the [grafana-configuration configmap](manifests/grafana-configuration.yaml) disables the access to users not already present within the Grafana database. As soon as this inherent [PR](https://github.com/grafana/grafana/pull/22383) is merged and the new version of Grafana released, it should be possible to limit the access to only a subset of Keycloak users directly from grafana, without needing to create duplicated accounts.

#### Additional references

1. [ingress-nginx - External OAUTH authentication](https://github.com/kubernetes/ingress-nginx/tree/master/docs/examples/auth/oauth-external-auth)
2. [oauth2-proxy - Configuration](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview)

### Monitor the Bind DNS Server

[bind_exporter](https://github.com/prometheus-community/bind_exporter) is a Prometheus exporter from Bind. The following guide is an adapted version of this [blog post](https://computingforgeeks.com/how-to-monitor-bind-dns-server-with-prometheus-and-grafana/).

#### Bind configuration
1. Download the latest release of the `bind_exporter` binary:
    ```bash
    $ wget -qO - https://api.github.com/repos/prometheus-community/bind_exporter/releases/latest | grep browser_download_url | grep linux-amd64 |  cut -d '"' -f 4 | wget -qi -
    ```
2. Extract the binary and move it to the /url/local/bin folder:
    ```bash
    $ sudo tar xvf bind_exporter-*.tar.gz --directory /usr/local/bin --wildcards --strip-components 1 '*/bind_exporter'
    ```
3. Edit `/etc/bind/named.conf.options`, and open a statistics channel:
    ```
    statistics-channels {
        inet 127.0.0.1 port 8053 allow { 127.0.0.1; };
    };
    ```
4. Reload the `bind9` configuration:
    ```bash
    $ sudo rndc reload
    ```
5. Add the `prometheus` system user account
    ```bash
    $ sudo groupadd --system prometheus
    $ sudo useradd -s /sbin/nologin --system -g prometheus prometheus
    ```
6. Create a `systemd` unit file for `bind_exporter`:
    ```bash
    $ sudo tee /etc/systemd/system/bind_exporter.service<<'EOF'
    [Unit]
    Description=Prometheus
    Documentation=https://github.com/digitalocean/bind_exporter
    Wants=network-online.target
    After=network-online.target

    [Service]
    Type=simple
    User=prometheus
    Group=prometheus
    ExecReload=/bin/kill -HUP $MAINPID
    ExecStart=/usr/local/bin/bind_exporter \
      --bind.pid-file=/var/run/named/named.pid \
      --bind.timeout=20s \
      --web.listen-address=0.0.0.0:9153 \
      --web.telemetry-path=/metrics \
      --bind.stats-url=http://localhost:8053/ \
      --bind.stats-groups=server,view,tasks

    SyslogIdentifier=prometheus
    Restart=always

    [Install]
    WantedBy=multi-user.target
    EOF
    ```
7. Reload `systemd` and start the `bind_exporter` service:
    ```bash
    $ sudo systemctl daemon-reload
    $ sudo systemctl enable bind_exporter.service
    $ sudo systemctl restart bind_exporter.service
    ```
8. Configure `iptables` to limit the access to `bind_exporter` to specific IP addresses (optional):
    ```bash
    $ sudo iptables -A INPUT -p tcp -s 130.192.0.0/16 --dport 9153 -j ACCEPT
    $ sudo iptables -A INPUT -p tcp -s 192.168.0.0/16 --dport 9153 -j ACCEPT
    $ sudo iptables -A INPUT -p tcp --dport 9153 -j DROP
    ```

#### Prometheus and Grafana configuration
1. Create the `Endpoint`, `Service` and `ServiceMonitor` resources to scrape the metrics exported by `bind_exporter`:
    ```bash
    $ kubectl create -f manifests/prometheus-bind-exporter.yaml
    ```
2. Open the Grafana web page and import the dashboard with ID 1666.

### Other information
For more information, look at the Github page of [kube-prometheus](https://github.com/coreos/kube-prometheus).
