# Loki - Promtail for logs

- [Loki](#loki)
  -  [Prerequisites](#prerequisites)
  -  [Installation](#installation)
- [Promtail](#promtail)
  -  [Prerequisites](#prerequisites-1)
  -  [Installation](#installation-1)
- [Bibliography](#bibliography)

# Loki

Grafana Loki is a set of elements that can be composed into a fully featured logging stack.
Loki is built around the idea of indexing logs metadata only by means of labels (just like Prometheus labels).
The actual logs data is packed into chunks and stored in object stores like S3. 
A small index and highly compressed chunks simplify the operations and significantly lower the cost of Loki.

In our cluster we rely on Rook Ceph as a distributed storage provider. For a more detailed explanation of the object storage creation please refer to the [official documentation](https://rook.io/docs/rook/v1.7/ceph-object.html).

## Prerequisites
Helm must be installed to deploy the chart. Please follow the [Helm installing guide](https://helm.sh/docs/intro/install/) before getting started.

After Helm is ready, add the Grafana repo as follow:

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```
To see the Grafana charts, you can run:

```bash
helm search repo grafana
```

## Installation

To install the chart with the release name `loki` and apply the configuration specified by the `loki-values.yaml` file, it is possible to proceed as follows:

```bash
helm upgrade --install --namespace monitoring --create-namespace loki -f loki-values.yaml grafana/loki
```

The [loki-values.yaml](./loki-values.yaml) defines the Loki configuration. You must configure the values file according to your specifications; here, we modified it to suit the CrownLabs needs. The main changes are below, the original values file is located [here](https://github.com/grafana/helm-charts/blob/main/charts/loki/values.yaml).

```yaml
  ingester:
    ...
    wal:
      dir: /data/loki/wal
    lifecycler:
      ring:
        kvstore:
          store: memberlist
        replication_factor: 3
    ...
  memberlist:
    abort_if_cluster_join_fails: false
    join_members:
    - loki-headless
```
The **store** is the backend storage to use for the ring, which is a space used to share logs across multiple ingesters (components that receive data, such as Loki replicas). The **store** can be **inmemory**, **memberlist**, **cunsul** and **etcd**.

With the **memberlist** configuration, the information is shared between the replicas, so that it spreads faster, reducing collisions and allowing to detect replica failures.

```yaml
  schema_config:
    configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: s3 #filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h
    ...
  storage_config:
    boltdb_shipper:
      active_index_directory: /data/loki/boltdb-shipper-active
      cache_location: /data/loki/boltdb-shipper-cache
      cache_ttl: 24h         # Can be increased for faster performance over longer query periods, uses more disk space
      shared_store: s3 #filesystem
    aws:
      s3: s3://<ACCESS_KEY>:<SECRET_ACCESS_KEY>@BUCKET_HOST/BUCKET_NAME
      s3forcepathstyle: true
    filesystem:
      directory: /data/loki/chunks
    ...
  compactor:
    working_directory: /data/loki/boltdb-shipper-compactor
    shared_store: s3 #filesystem
```
You can choose where you want to store the data, either in the local filesystem (PVC) or in a S3 object store. For the latter, you must configure the S3 bucket parameters as indicated. Some examples of Loki configurations are also available [here](https://grafana.com/docs/loki/latest/configuration/examples/#loki-configuration-examples).

*NOTE: some buckets, for example AWS, want the region before the BUCKET_HOST. So the full S3 path is s3://<ACCESS_KEY>:<SECRET_ACCESS_KEY>@REGION_BUKET.BUCKET_HOST/BUCKET_NAME*.

# Promtail

Promtail is an agent which ships the contents of local logs to a private Grafana Loki instance. It is usually deployed to every machine that has applications needed to be monitored.
It primarily:
 - Discovers targets
 - Attaches labels to log streams
 - Pushes them to the Loki instance

## Prerequisites

Helm must be installed to use the charts value. Please follow the [Helm installing guide](https://helm.sh/docs/intro/install/) before getting starting.

After Helm is ready, add the Grafana repo as follow:

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```
To see the Grafana charts, you can run:

```bash
helm search repo grafana
```

## Installation

To install the chart with the release name `promtail` and apply the configuration specified by the `promtail-values.yaml` file, it is possible to proceed as follows:

```bash
helm upgrade --install --namespace monitoring --create-namespace promtail -f promtail-values.yaml grafana/promtail
```

The [promtail-values.yaml](./promtail-values.yaml) describes the Promtail configuration. Below we list the most relevant modifications compared to the original configuration that can be found [here](https://github.com/grafana/helm-charts/blob/main/charts/promtail/values.yaml).

```yaml
  snippets:
    pipelineStages:
      #- cri: {}
      - docker: {}
```

The **pipelineStages** is the way of extracting data by parsing the log line. You can use the [Docker standard](https://grafana.com/docs/loki/latest/clients/promtail/stages/docker/), the [CRI standard](https://grafana.com/docs/loki/latest/clients/promtail/stages/cri/), a [regular expression](https://grafana.com/docs/loki/latest/clients/promtail/stages/regex/) or parse the log line as [JSON](https://grafana.com/docs/loki/latest/clients/promtail/stages/json/). You can only use one of these standard. 

*NOTE: for JSON, you must configure **all** the expressions of your logs manually ([Using extraced data example](https://grafana.com/docs/loki/latest/clients/promtail/stages/json/#using-extracted-data)), because it does not recognize the key-value fields of the JSON, which means non-string elements like numbers, booleans or timestamps will not be assigned to those types.*

### Bibliography

1. [Loki Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/loki)
2. [Loki documentations](https://grafana.com/docs/loki/latest/configuration/)
3. [Memberlist](https://github.com/hashicorp/memberlist)
4. [Promtail stages](https://grafana.com/docs/loki/latest/clients/promtail/stages/)
5. [Promtail documentations](https://grafana.com/docs/loki/latest/clients/promtail/)
