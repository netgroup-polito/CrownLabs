# Thanos

Thanos leverages the Prometheus 2.0 storage format to cost-efficiently store historical metric data in any object storage while retaining fast query latencies. Additionally, it provides a global query view across all Prometheus installations and can merge data from Prometheus HA pairs on the fly.

Concretely the aims of the project are:
1. Global query view of metrics.
2. Unlimited retention of metrics.
3. High availability of components, including Prometheus.

## Installation
In order to provide all of these nice features Thanos needs some object store. In our cluster we rely on Rook Ceph as a distributed storage provider. For a more detailed explanation of the object storage creation the official guide is here https://rook.io/docs/rook/v1.3/ceph-object.html, in this guide we present only the main steps.

First of all the CephObjectStore resource need to be created 

```sh
$ kubectl apply -f ./manifests/objectStorage.yaml
```

Now that the object store is configured, next we need to create a bucket where a client can read and write objects. A bucket can be created by defining a storage class

```sh
$ kubectl apply -f ./manifests/storageClass.yaml
```

Based on this storage class, an object client can now request a bucket by creating an Object Bucket Claim


```sh
$ kubectl apply -f ./manifests/objectBucketClaim.yaml
```

Now that the object store is configured and a bucket created, it's possible to consume the object storage from an S3 client. In order to access the object storage these three environment variable need to be set.

```sh
export AWS_HOST=$(kubectl -n default get cm ceph-thanos-bucket -o yaml | grep " BUCKET_HOST" | awk '{print $2}')
export AWS_ACCESS_KEY_ID=$(kubectl -n default get secret ceph-thanos-bucket -o yaml | grep " AWS_ACCESS_KEY_ID" | awk '{print $2}' | base64 --decode)
export AWS_SECRET_ACCESS_KEY=$(kubectl -n default get secret ceph-thanos-bucket -o yaml | grep " AWS_SECRET_ACCESS_KEY" | awk '{print $2}' | base64 --decode) 
```

Once we have exported these values we need to create a configuration file for kubernetes in order make Thanos aware of how to connect to the object storage (replace with the corresponding values)


```yaml
type: s3
config:
  bucket: ceph-thanos-bkt
  endpoint: <<BUCKET_HOST>>
  access_key: <<AWS_ACCESS_KEY_ID>>
  secret_key: <<AWS_SECRET_ACCESS_KEY>>
```

Save it and then create a secret with it using the command

```sh
$ kubectl -n monitoring create secret generic thanos-objstore-config --from-file=thanos.yaml=./<<your-file-name>>.yaml
```

And finnaly the last thing to do is to add the Thanos sidecar to Prometheus with the secret we just created. To do so add the following lines to the Prometheus CRD

```yaml
...
spec:
  ...
  thanos:
    baseImage: quay.io/thanos/thanos
    version: v0.8.1
    objectStorageConfig:
      key: thanos.yaml
      name: thanos-objstore-config
...
```

Thanos provide a dashboard to query the metrics. To access it Thanos needs to be connected to the ingress controller and this can be done creating a service around the container

```sh
$ kubectl apply -f ./manifests/thanos-service.yaml
```

And then link the ingress resource to it.

```sh
kubectl apply -f ./manifests/ingress.yaml
kubectl apply -f ./manifests/ingress-oauth2.yaml
```

## Extensions
Thanos comes with a lot of extensions which enable other features. In our case we decided to add the following components:
1. thanos compact, applies the compaction procedure of the Prometheus 2.0 storage engine to block data stored in object storage. It is also responsible for downsampling of data. It can be created with [thanos-compact-statefulSet.yaml](./manifests/thanos-compact-statefulSet.yaml)
2. thanos querier, implements the Prometheus HTTP v1 API to query data in a Thanos cluster via PromQL. In short, it gathers the data needed to evaluate the query from underlying StoreAPIs, evaluates the query and returns the result. It can be created with [thanos-query-deployment.yaml](manifests/thanos-query-deployment.yaml), [thanos-query-service.yaml](manifests/thanos-query-service.yaml), [thanos-query-serviceMonitor.yaml](manifests/thanos-query-serviceMonitor.yaml)
3. thanos store, implements the Store API on top of historical data in an object storage bucket. It can be created with [thanos-store-service.yaml](manifests/thanos-store-service.yaml), [thanos-store-serviceMonitor.yaml](manifests/thanos-store-serviceMonitor.yaml), [thanos-store-statefulSet.yaml](manifests/thanos-store-statefulSet.yaml)
