# Docker registry

## Table of contents
- [What is it](#what-is-it)
- [Why do we need it](#why-do-we-need-it)
- [Docker registry Helm Chart](#docker-registry-helm-chart)
  - [Configuration](#configuration)
- [Installing the chart](#installing-the-chart)
- [Warning: max HTTP body size in ingress controller](#warning-max-http-body-size-in-ingress-controller)


## What is it
From the [Docker Registry](https://docs.docker.com/registry/) official documentation: the Registry is a stateless, highly scalable server-side application that stores and lets you distribute Docker images.

## Why do we need it?
You should use the Registry if you want to:
- tightly control where your images are being stored
- fully own your images distribution pipeline
- integrate image storage and distribution tightly into your in-house development workflow
- leverage the high-speed network that connects your servers, avoiding to consume precious Internet bandwidth to transfer images stored in the Docker Hub public service.

Finally, consider that, in this Kubernetes setup, users instantiate mainly VMs, whose image may be rather large. Allowing users to download the VM image locally, instead of from a remote server, would greatly impact on their quality of experience in term of time required to start their service.

## Docker Registry Helm Chart
We used the [Docker Registry Helm Chart](https://github.com/helm/charts/tree/master/stable/docker-registry), that is a Kubernetes chart to deploy a private Docker Registry where we have appropriately modified the values in [values.yaml](https://github.com/helm/charts/blob/master/stable/docker-registry/values.yaml) file.

### Configuration
These are our modification of the various file.

  1. Annotations concerning the authentication in the Ingress controller.
```yaml
annotations:
   nginx.ingress.kubernetes.io/auth-realm: Authentication Required - ok
   nginx.ingress.kubernetes.io/auth-secret: basic-auth
   nginx.ingress.kubernetes.io/auth-type: basic
```

  2. Annotations concerning the configuration of the Ingress controller.
```yaml
ingress:
  enabled: true
  path: /
  hosts:
    - <registry.domain.name>
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-production
  labels: {}
  tls:
    - hosts:
      - <registry.domain.name>
```

  3. Configuration of the persistent disk. Please dimension this disk appropriately, as VM images may be rather large and hence you may fill up this space in a very short time.
```yaml
persistence:
  accessMode: 'ReadWriteOnce'
  enabled: true
  size: 500Gi
  storageClass: 'rook-ceph-block'
```

## Installing the Chart
To install the chart, use the following command:
```bash
$ helm install stable/docker-registry -f docker-registry-configuration.yaml -n docker-registry --generate-name
```
Look [here](https://github.com/helm/charts/tree/master/stable/docker-registry#configuration) for a complete configuration guide.


## Warning: max HTTP body size in ingress controller
By default, the nginx ingress controller has a limit on the maximum size of the HTTP body. If the image to be uploaded is larger, an error is received (403). To avoid this problem, modify the configuration of your ingress controller and add a new annotation:
```yaml
nginx.ingress.kubernetes.io/proxy-body-size: "0"
```
