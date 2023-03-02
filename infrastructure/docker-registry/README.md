# Harbor Docker registry

## Table of contents
- [What is it](#what-is-it)
- [Why do we need it](#why-do-we-need-it)
- [Docker registry Helm Chart](#harbor-registry-helm-chart)
  - [Pre-requisites](#pre-requisites)
  - [Redis Configuration](#redis-configuration)
  - [Postgres Configuration](#postgres-configuration)
  - [Harbor Configuration](#harbor-configuration)
- [Installing the chart](#installing-the-chart)


## What is it
From the [Docker Registry](https://docs.docker.com/registry/) official documentation: A Registry is a stateless, highly scalable server-side application that stores and lets you distribute Docker images.
[Harbor](https://goharbor.io/) is an open source registry having a lot of features, such as an advanced UI, a vulnerability scanner, robot accounts and so on. For more information visit the official web page.

## Why do we need it?
You should use the Harbor Registry if you want to:
- tightly control where your images are being stored.
- fully own your images distribution pipeline.
- integrate image storage and distribution tightly into your in-house development workflow.
- leverage the high-speed network that connects your servers, avoiding to consume precious Internet bandwidth to transfer images stored in the Docker Hub public service.
- leverage [Proxy Cache](https://goharbor.io/docs/2.7.0/administration/configure-proxy-cache/) functionalities, to not exceed Docker Hub’s rate limiting policy.
- have a [vulnerability scanner](https://goharbor.io/docs/2.7.0/administration/vulnerability-scanning/) to detect possible image vulnerabilities.
- manage your [Helm Charts](https://goharbor.io/docs/edge/working-with-projects/working-with-images/managing-helm-charts/).

Finally, consider that, in this Kubernetes setup, users instantiate mainly VMs, whose image may be rather large. Allowing users to download the VM image locally, instead of from a remote server, would greatly impact on their quality of experience in term of time required to start their service.

## Harbor Registry Helm Chart
To install Harbor, it is possible to leverage the [official Helm Chart](https://github.com/helm/charts/tree/master/stable/docker-registry), appropriately configuring the `values.yaml` file (additional details follow in the next sections).

### Pre-requisites

  1. Kubernetes cluster 1.10+
  2. [Helm 3](https://helm.sh/docs/intro/install/)
  3. High available ingress controller (Harbor does not manage the external endpoint)
  4. [High available PostgreSQL 9.6+](#postgres-configuration) (Harbor does not handle the HA deployment of the database)
  5. [High available Redis](#redis-configuration) (Harbor does not handle the HA deployment of Redis)
  6. PVC that can be shared across nodes (i.e., with `ReadWriteMany` access mode) or external object storage

### Redis Configuration
In our architecture we have a [Redis-Sentinel](https://redis.io/docs/manual/sentinel/) service, instead of [Redis Cluster](https://redis.io/docs/manual/scaling/), because with this architecture Sentinel manages automatically the failover of the master.
To enable the `Redis-Sentinel ` architecture it is necessary to configure the following parameter in the redis file values (`redis-service-values.yaml`):
```yaml
  sentinel.enabled=true
```

To deploy Redis service, it is possible to proceed as follows:
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm upgrade redis bitnami/redis --namespace harbor \
    --install --create-namespace --values redis-service-values.yaml
```

### Postgres Configuration
A pre-requisite to deploy PostregreSQL cluster is the [PostgreSQL-Operator](https://github.com/netgroup-polito/CrownLabs/tree/master/infrastructure/identity-provider#postgresql-operator), because it delivers an easy way to run highly-available PostgreSQL clusters on Kubernetes.
Once you have a PostgreSQL-Operator running, you can create your PostgreSQL cluster with the following command:
```bash
kubectl apply -f postgres-cluster-manifest.yaml
```
This command creates the database and applies the configuration specified by the `postgres-cluster-manifest.yaml`.

### Harbor Configuration
The following outlines the most relevant modifications applied to the Harbor values file (`harbor-values.yaml`):

  1. Configuration of how the harbor registry is exposed (i.e., by means of an ingress), the external URL and accessory parameters (e.g., the annotations concerning certificate generation).

  2. Configuration of the parameters to access the Postgres database created previously.

  3. Configuration of the parameters to access the Redis service deployed previously.

## Installing the Chart
Before installing the chart, the Harbor repository must be added to helm with the following command:
```bash
helm repo add harbor https://helm.goharbor.io
helm repo update
```
To install the chart with the release name `harbor` and apply the configuration specified by the `harbor-values.yaml` file, it is possible to proceed as follows:
```bash
helm upgrade harbor harbor/harbor --namespace harbor \
    --install --create-namespace --values harbor-values.yaml
```
Warning: credentials and secret parameters have been redacted from the values file stored in this repository.
Look [here](https://github.com/goharbor/harbor-helm) for a complete configuration guide.
