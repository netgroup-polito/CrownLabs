# kube-prometheus-stack

  - [Alertmanager](#alertmanager)
  - [Grafana](#grafana)
  - [Prometheus](#prometheus)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
    - [Persistent storage](#persistent-storage)
    - [Alert Notification](#alert-notification)
    - [OAuth2 Authentication](#oauth2-authentication)
    - [Monitor the Bind DNS Server](#monitor-the-bind-dns-server)
  - [Other information](#other-information)

## Alertmanager
From Promtheus [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) official documentation.
The Alertmanager handles alerts sent by client applications such as the Prometheus server. It takes care of deduplicating, grouping, and routing them to the correct receiver integration such as email, PagerDuty, or OpsGenie. It also takes care of silencing and inhibition of alerts

## Grafana
From [Grafana](https://grafana.com/docs/grafana/latest/getting-started/#what-is-grafana) official documentation.
Grafana is an open source visualization and analytics software. It allows you to query, visualize, alert on, and explore your metrics no matter where they are stored. We can also say that Grafana is the tool for beautiful monitoring and metric analytics & dashboards for Graphite, InfluxDB & Prometheus & More.

## Prometheus
From [Prometheus](https://prometheus.io/docs/introduction/overview/) official documentation.
Prometheus is an open-source systems monitoring and alerting toolkit.
Prometheus's main features are:
- a multi-dimensional data model with time series data identified by metric name and key/value pairs
- PromQL, a flexible query language to leverage this dimensionality
- no reliance on distributed storage; single server nodes are autonomous
- time series collection happens via a pull model over HTTP
- pushing time series is supported via an intermediary gateway
- targets are discovered via service discovery or static configuration
- multiple modes of graphing and dashboarding support

## Prerequisites
Helm must be installed to deploy the chart. Please follow the [Helm installing guide](https://helm.sh/docs/intro/install/) before getting started.

After Helm is ready, add the Prometheus Community repo as follow:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```
To see the Prometheus Community charts, you can run:

```bash
helm search repo prometheus-community
```

## Installation
To install the chart with the release name `kube-prometheus-stack` and apply the configuration specified by the `kube-prometheus-stack-values.yaml` file, it is possible to proceed as follows:

```bash
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
     --namespace=monitoring --values=kube-prometheus-stack-values.yaml
```

*NOTE: the release name must be `kube-prometheus-stack` because [name overrides cause webhook calls to fail](https://github.com/prometheus-community/helm-charts/issues/257).*

### Persistent storage

Running cluster monitoring with persistent storage means that your metrics are stored to a Persistent Volume and can survive a pod being restarted or recreated. This is ideal if you require your metrics or alerting data to be guarded from data loss. For production environments, it is highly recommended to configure persistent storage.

We need to modify the relative fields for Grafana and Prometheus, to have persistent storage:

- [Grafana persistence](./kube-prometheus-stack-values.yaml#L872)
  ```yaml
  persistence:
    type: pvc
    enabled: true
    storageClassName: rook-cephfs-primary
    accessModes:
      - ReadWriteMany
    size: 10Gi
  ```

- [Prometheus retention](./kube-prometheus-stack-values.yaml#L2417).

  ```yaml
  ## How long to retain metrics
  ##
  retention: 15d
  ## Maximum size of metrics
  ##
  retentionSize: 90GB
  ```

- [Prometheus storage persistence](./kube-prometheus-stack-values.yaml#L2517)

  ```yaml
  storageSpec:
  ## Using PersistentVolumeClaim
  ##
    volumeClaimTemplate:
      spec:
        storageClassName: rook-ceph-block
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 100Gi
  ```

Before applying your cluster configuration, you have to enter the correct value for `podAntiAffinity` for [Prometheus](./kube-prometheus-stack-values.yaml#L2470), it allows you to specify rules about how pods should be placed relative to other pods and it can prevent the scheduler from locating a new pod on the same node.

```yaml
## Pod anti-affinity can prevent the scheduler from placing Prometheus replicas on the same node.
## The default value "soft" means that the scheduler should *prefer* to not schedule two replica pods onto the same node but no guarantee is provided.
## The value "hard" means that the scheduler is *required* to not schedule two replica pods onto the same node.
## The value "" will disable pod anti-affinity so that no anti-affinity rules will be configured.
podAntiAffinity: "hard"
```


### Alert Notification
When an alert is sent to AlertManager, those are also sent to Slack.
This template differentiates alerts based on severity and sends them on the correct Slack channel.

- [alertmanager config](./kube-prometheus-stack-values.yaml#L152)
  ```yaml
      - matchers:
        - alertname = Watchdog
        receiver: 'null'
      - matchers:
        - severity = critical
        receiver: 'critical'
      - matchers:
        - severity = warning
        receiver: 'warning'
  receivers:
    - name: 'critical'
      slack_configs:
      - channel: 'alerts-critical'
        send_resolved: true
        api_url: <YOUR HOOK URL>
    - name: 'warning'
      slack_configs:
      - channel: 'alerts-warning'
        send_resolved: true
        api_url: <YOUR HOOK URL>
    - name: 'info'
      slack_configs:
      - channel: 'alerts-info'
        send_resolved: true
        api_url: <YOUR HOOK URL>
  ```

### OAuth2 Authentication
In the following, we will setup Alertmanager, Grafana and Prometheus to use Keycloak as identity provider for the authentication. Grafana can natively use Keycloak as identity provider, while the authentication for the other two services is managed through the ingress controller and [oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy).

*Note: this guide assumes Keycloak to be already deployed and available. Please refer to the [keycloak deployment guide](../../identity-provider/README.md) for more information.*

#### Keycloak configuration

1. Create a new client for `monitoring`;
2. Configure the Client Protocol to be `openid-connect`;
3. Set Access Type to `confidential`;
4. Configure the Valid Redirect URIs to the alertmanager, grafana and prometheus URLs (e.g. https://alertmanager.example.com/*, https://grafana.example.com/*, https://prometheus.example.com/*); <!-- markdown-link-check-disable-line -->
5. From the Credentials tab, copy the Client Secret that has been generated;
6. From the Mappers tab, add a new `Group Membership` mapper with Token Claim Name equal to `groups`.

#### Alertmanager and Prometheus configuration

Edit the ingress in the [kube-prometheus-stack file values](./kube-prometheus-stack-values.yaml) and adapt it to your configuration. In particular, it is necessary to instert the `annotations`, specify the `hosts`, and configure the `tls`. This needs to be done for [alertmanager ingress](./kube-prometheus-stack-values.yaml#L268) and [prometheus ingress](./kube-prometheus-stack-values.yaml#L2067)

#### Grafana configuration

Edit the [grafana.ini config](./kube-prometheus-stack-values.yaml#L783) and adapt it to your configuration (each line corresponds to an environment variable). In particular, it is necessary to adapt the different URIs and specify the Client ID and Secrets generated in Keycloak. The meaning of the different fields is specified by the embedded comments. More information can be found in the [official documentation](https://grafana.com/docs/grafana/latest/auth/generic-oauth/).

#### Limit access to a subset of Keycloak users

The access to Alertmanager and Prometheus is limited to users belonging to the `monitoring` group. Hence, it is possible to grant access to the users by simply adding them to the `monitoring` group.

**Warning:** At the time of writing, in Grafana it seems not possible to restrict login by role/group when using the `generic_oauth` feature. Hence, all valid users of the Keycloak realm will be able to access Grafana.

#### Additional references

1. [ingress-nginx - External OAUTH authentication](https://github.com/kubernetes/ingress-nginx/tree/master/docs/examples/auth/oauth-external-auth)
2. [oauth2-proxy - Configuration](https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview)

### Monitor the Bind DNS Server

[bind_exporter](https://github.com/prometheus-community/bind_exporter) is a Prometheus exporter from Bind.
<!-- markdown-link-check-disable-next-line -->
The following guide is an adapted version of this [blog post](https://computingforgeeks.com/how-to-monitor-bind-dns-server-with-prometheus-and-grafana/).

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
For more information, look at the Github page of [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) and of [kube-prometheus](https://github.com/coreos/kube-prometheus).
