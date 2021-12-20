# Harbor Docker registry

## Table of contents
- [What is it](#what-is-it)
- [Why do we need it](#why-do-we-need-it)
- [Docker registry Helm Chart](#harbor-registry-helm-chart)
  - [Pre-requisites](#pre-requisites)
  - [Configuration](#configuration)
- [Installing the chart](#installing-the-chart)


## What is it
From the [Docker Registry](https://docs.docker.com/registry/) official documentation: A Registry is a stateless, highly scalable server-side application that stores and lets you distribute Docker images. [Harbor](https://goharbor.io/) is an open source registry having a lot of features, such as vulnerability scanner, robot account and so on. For more information visit the official web page.

## Why do we need it?
You should use the Harbor Registry if you want to:
- tightly control where your images are being stored
- fully own your images distribution pipeline
- integrate image storage and distribution tightly into your in-house development workflow
- leverage the high-speed network that connects your servers, avoiding to consume precious Internet bandwidth to transfer images stored in the Docker Hub public service.
- [Proxy Cache](https://goharbor.io/docs/2.4.0/administration/configure-proxy-cache/) to not exceed Docker Hub’s rate limit policy.
- have a [vulnerability scanner](https://goharbor.io/docs/2.4.0/administration/vulnerability-scanning/)
- manage your [Chart](https://goharbor.io/docs/edge/working-with-projects/working-with-images/managing-helm-charts/)
- [Create System Robot Accounts](https://goharbor.io/docs/2.4.0/administration/robot-accounts/)

Finally, consider that, in this Kubernetes setup, users instantiate mainly VMs, whose image may be rather large. Allowing users to download the VM image locally, instead of from a remote server, would greatly impact on their quality of experience in term of time required to start their service.

## Harbor Registry Helm Chart
We used the [Harbor Registry Helm Chart](https://github.com/goharbor/harbor-helm), that is a Kubernetes chart to deploy a private Harbor Registry where we have appropriately modified the values in [values.yaml](https://github.com/goharbor/harbor-helm/blob/master/values.yaml) file.

### Pre-requisites

  1. [Kubernetes cluster 1.10+](https://kubernetes.io/releases/download/)
  2. [Helm 2.8.0+](https://helm.sh/docs/intro/install/)
  3. [High available ingress controller](https://github.com/netgroup-polito/CrownLabs/tree/master/infrastructure/ingress-controller) (Harbor does not manage the external endpoint)
  4. [High available PostgreSQL 9.6+](#https://github.com/netgroup-polito/CrownLabs/tree/master/infrastructure/identity-provider#postgresql-operator) (Harbor does not handle the deployment of HA of database)
  5. [High available Redis](#https://github.com/bitnami/charts/tree/master/bitnami/redis/#installing-the-chart) (Harbor does not handle the deployment of HA of Redis)
  6. PVC that can be shared across nodes or external object storage


### Configuration
These are our modification of the various file. They all refer to the values.yaml file except the configuration n. 5 and 6. These 2 configuration refer respectively to postgres postgres_cluster_manifest.yaml and redis_service_values.yaml file.

  1. Annotations concerning the configuration of the Ingress controller. You have to set the type of the ingress (in our file values.yaml the type of the ingress is ingress) and the parameters for the type you have chosen. For example you have to set the name of the hosts of core and notary by setting the ingress.hosts.core and ingress.hosts.notary parameters.

  2. The external URL to expose the Harbor Registry. To set the external URL you have to set the externalURL parameter

  3. The deployment pods of Portal, Core and Registry to have 3 replicas of each one and to set the anti-affinity of each pod, to have pod of the same services in different nodes. We have also set the resources consumptions after have done a stress test. To do this you have to set respectively the affinity and the resources parameter

  4. For the installation of Postgres service, please see the link on the pre-requisites section.

  5. For the installation of Redis service, please see the link on the pre-requisites section. In our architecture we have a [Redis-Sentinel](https://redis.io/topics/sentinel) service, instead of [Redis Cluster](https://redis.io/topics/cluster-tutorial) by setting the following parameter in the file values of redis_service_values.yaml file. In the file you have to set an IP for the ClusterIP of the sentinel where after harbor is attached in the cluster and a name for your master. If you want have the replicas of redis running in different nodes you have to set the replica.podAntiAffinityPreset parameter to soft or hard, depending what you need. Please see this [link](https://docs.bitnami.com/tutorials/assign-pod-nodes-helm-affinity-rules/) to see the differences.
  ```yaml 
  sentinel.enabled=true
  ```
  6. For the external redis service you have to modify the redis.type parameter to external and have to set the same IP and the same master name of the previous redis service respectively in redis.external.addr parameter and in redis.external.sentinelMasterSet parameter.

## Installing the Chart
Before install the chart, the repository must be add to helm with the following command:
```bash
helm repo add harbor https://helm.goharbor.io
```
To install the chart, use the following command:
```bash
$ helm install my-release harbor/harbor
```
The above command install a release called 'my-release'. If you want install a release starting from a values.yaml file, you can use the following command:
```bash
$ helm install my-release harbor/harbor -f values.yaml
```
Look [here](https://github.com/goharbor/harbor-helm) for a complete configuration guide.

