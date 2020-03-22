# docker-registry


## Table of contents

- [Docker registry](#docker-registry-first)
  - [Table of contents](#table-of-contents)
  - [Helm](#helm)
  - [Docker registry](#docker-registry)
  
  
  
## Docker Registry

### What is it
The Registry is a stateless, highly scalable server side application that stores and lets you distribute Docker images. The Registry is open-source, under the permissive Apache license.

## Why we need it?
You should use the Registry if you want to:
* tightly control where your images are being stored
* fully own your images distribution pipeline
* integrate image storage and distribution tightly into your in-house development workflow

## Helm
  
Helm is a tool for managing Charts. Charts are packages of pre-configured Kubernetes resources.

Use Helm to:

- Find and use [popular software packaged as Helm Charts](https://hub.helm.sh) to run in Kubernetes
- Share your own applications as Helm Charts
- Create reproducible builds of your Kubernetes applications
- Intelligently manage your Kubernetes manifest files
- Manage releases of Helm packages

### How to install Helm
There are different ways to install it, you can check it [here](https://helm.sh/docs/intro/install/) or you can directly do:
```
$ curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
$ chmod 700 get_helm.sh
$ ./get_helm.sh
```

## Docker Registry Helm Chart
It is a Kubernetes chart to deploy a private Docker Registry

### Prerequisites Details

* PV support on underlying infrastructure (if persistence is required)

### Chart Details

This chart will do the following:

* Implement a Docker registry deployment

### Installing the Chart

To install the chart, use the following:

```console
$ helm install stable/docker-registry
```
### Configuration
These are our modification of the [values.yaml](https://github.com/helm/charts/blob/master/stable/docker-registry/values.yaml) file:
* annotations
* ingress
* persistence
```
annotations:
   nginx.ingress.kubernetes.io/auth-realm: Authentication Required - ok
   nginx.ingress.kubernetes.io/auth-secret: basic-auth
   nginx.ingress.kubernetes.io/auth-type: basic
```
```
ingress:
  enabled: true
  path: /
  # Used to create an Ingress record.
  hosts:
    - registry.crown-labs.ipv6.polito.it
  annotations: {}
    # kubernetes.io/ingress.class: nginx
    # kubernetes.io/tls-acme: "true"
  labels: {}
  tls:
    # Secrets must be manually created in the namespace.
     #- secretName: 
       - hosts:
         - registry.crown-labs.ipv6.polito.it
```

```
persistence:
  accessMode: 'ReadWriteOnce'
  enabled: true
  size: 100Gi
  storageClass: 'rook-ceph-block'
```

docker pull
docker push
docker 

L'ingress del Docker Registry contiene l'annotation: nginx.ingress.kubernetes.io/proxy-body-size: 4g 
questa è la dimensione del body che possiamo pushare sulla docker registry

### *warning*
The ingress has protection on the maximum size of the http body. If the image to be uploaded is larger, an error is received (403). To get around this problem just go to the ingress and add 0 to:
`` `
nginx.ingress.kubernetes.io/proxy-body-size: "0"
`` `


For a complete configuration looks [here](https://github.com/helm/charts/tree/master/stable/docker-registry#configuration)




  

