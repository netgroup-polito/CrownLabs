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

The [loki-values.yaml](./loki-values.yaml) defines the Loki configuration. You must configure the values file according to your specifications; here, we modified it to suit the CrownLabs needs. The main changes are below, the original values file is located [here](https://github.com/grafana/helm-charts/blob/loki-2.8.1/charts/loki/values.yaml).

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
You can choose where you want to store the data, either in the local filesystem (PVC) or in a S3 object store. For the latter, you must configure the S3 bucket parameters as indicated. Some examples of Loki configurations are also available [here](https://grafana.com/docs/loki/latest/configure/examples/).

*NOTE: some buckets, for example AWS, want the region before the BUCKET_HOST. So the full S3 path is s3://<ACCESS_KEY>:<SECRET_ACCESS_KEY>@REGION_BUKET.BUCKET_HOST/BUCKET_NAME*.

