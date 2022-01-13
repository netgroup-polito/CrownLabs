# Thanos

Thanos leverages the Prometheus 2.0 storage format to cost-efficiently store historical metric data in any object storage while retaining fast query latencies. Additionally, it provides a global query view across all Prometheus installations and can merge data from Prometheus HA pairs on the fly.

Concretely the aims of the project are:
1. Global query view of metrics.
2. Unlimited retention of metrics.
3. High availability of components, including Prometheus.

## Prerequisites
Helm must be installed to deploy the chart. Please follow the [Helm installing guide](https://helm.sh/docs/intro/install/) before getting started.

After Helm is ready, add the Bitnami repo as follow:

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```
To see the Bitnami charts, you can run:

```bash
helm search repo bitnami
```

## Installation

To install the chart with the release name `thanos` and apply the configuration specified by the `thanos-values.yaml` file, it is possible to proceed as follows:

```bash
helm upgrade --install --namespace monitoring thanos bitnami/thanos --values=thanos-values.yaml
```
In order to provide all of these nice features Thanos needs some object store. In our cluster we rely on Rook Ceph as a distributed storage provider. For a more detailed explanation of the object storage creation the official guide is [here](https://rook.io/docs/rook/v1.7/ceph-object.html), in this guide we present only the main steps.

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

And finally the last thing to do is to add the Thanos sidecar to Prometheus with the secret we just created. To do so modify the following lines to the [kube-prometheus-stack-values.yaml](../kube-prometheus-stack-values.yaml#L2647).



```yaml
prometheus:
  ...
  prometheusSpec:
    ## Thanos configuration allows configuring various aspects of a Prometheus server in a Thanos environment.
    ## This section is experimental, it may change significantly without deprecation notice in any release.
    ## This is experimental and may change significantly without backward compatibility in any release.
    ## ref: https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#thanosspec
    ##
    thanos: 
      objectStorageConfig:
        key: thanos.yaml
        name: thanos-objstore-config
      image: quay.io/thanos/thanos:v0.24.0
      version: v0.24.0
  ...    
```
 The [thanos-values.yaml](./thanos-values.yaml) defines the Thanos configuration. You must configure the values file according to your specifications; here, we modified it to suit the CrownLabs needs. The main changes are below, the original values file is located [here](https://github.com/bitnami/charts/blob/master/bitnami/thanos/values.yaml).

We have enabled the following components:
- [query](./thanos-values.yaml#L96) implements the Prometheus HTTP v1 API to query data in a Thanos cluster via PromQL. In short, it gathers the data needed to evaluate the query from underlying StoreAPIs, evaluates the query and returns the result.

- [queryFrontend](./thanos-values.yaml#L668) implements a service that can be put in front of Thanos Queriers to improve the read path. Query Frontend is fully stateless and horizontally scalable.
  ```yaml
    queryFrontend:
      enabled: true
      ...
      ingress:
        enabled: true
        ## @param queryFrontend.ingress.hostname Default host for the ingress resource
        ##
        hostname: thanos.crownlabs.polito.it
        ...
        ## @param queryFrontend.ingress.annotations Additional annotations for the Ingress resource. To enable certificate autogeneration, place here your cert-manager annotations.
        ## For a full list of possible ingress annotations, please see
        ## ref: https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/nginx-configuration/annotations.md
        ## Use this parameter to set the required annotations for cert-manager, see
        ## ref: https://cert-manager.io/docs/usage/ingress/#supported-annotations
        ##
        ## e.g:
        ## annotations:
        ##   kubernetes.io/ingress.class: nginx
        ##   cert-manager.io/cluster-issuer: cluster-issuer-name
        ##
        annotations:
          cert-manager.io/cluster-issuer: letsencrypt-production
          nginx.ingress.kubernetes.io/auth-signin: https://$host/oauth2/start?rd=$escaped_request_uri
          nginx.ingress.kubernetes.io/auth-url: https://$host/oauth2/auth
        ...  
      ...
  ```
- [compactor](./thanos-values.yaml#L1455) applies the compaction procedure of the Prometheus storage engine to block data stored in object storage. It is also responsible for downsampling of data.
- [storegateway](./thanos-values.yaml#L1841) implements the Store API on top of historical data in an object storage bucket.
- [metrics](./thanos-values.yaml#L3584) enable the export of Prometheus metrics.

Thanos provide a dashboard to query the metrics. To access it Thanos needs to be connected to the ingress controller and this can be done creating a service around the container. Enable the [thanosService](../kube-prometheus-stack-values.yaml#L1863) in the `kube-prometheus-stack-values.yaml` as follow:

```yaml
thanosService:
  enabled: true
  annotations: {}
  labels: {}

  ## Service type
  ##
  type: ClusterIP

  ## gRPC port config
  portName: grpc
  port: 10901
  targetPort: "grpc"
  ## HTTP port config (for metrics)
  httpPortName: http
  httpPort: 10902
  targetHttpPort: "http"
  ## ClusterIP to assign
  # Default is to make this a headless service ("None")
  clusterIP: "None"
  ## Port to expose on each node, if service type is NodePort
  ##
  nodePort: 30901
  httpNodePort: 30902
```
