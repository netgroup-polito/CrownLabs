# docker-registry
## Table of contents
- [Docker registry](#docker-registry-first)
  - [Table of contents](#table-of-contents)
  - [Helm](#helm)
  - [Docker registry](#docker-registry)
  
  
## Docker Registry


### What is it
From the [Docker Registry](https://docs.docker.com/registry/) official documentation.

The Registry is a stateless, highly scalable server side application that stores and lets you distribute Docker images. The Registry is open-source, under the permissive Apache license.


## Why we need it?
You should use the Registry if you want to:
* tightly control where your images are being stored
* fully own your images distribution pipeline
* integrate image storage and distribution tightly into your in-house development workflow


### Docker Registry Helm Chart
We used the [Docker Registry Helm Chart](https://github.com/helm/charts/tree/master/stable/docker-registry), that is a Kubernetes chart to deploy a private Docker Registry where we have appropriately modified the values in [values](https://github.com/helm/charts/blob/master/stable/docker-registry/values.yaml) file.
### Configuration
These are our modification of the values file.
*
```
annotations:
   nginx.ingress.kubernetes.io/auth-realm: Authentication Required - ok
   nginx.ingress.kubernetes.io/auth-secret: basic-auth
   nginx.ingress.kubernetes.io/auth-type: basic
```
*
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

### Installing the Chart
To install the chart, use the following:
```console
$ helm install stable/docker-registry -f configfile.yaml -n docker-registry --generate-name
```
For a complete configuration looks [here](https://github.com/helm/charts/tree/master/stable/docker-registry#configuration)




### *warning*
The ingress has protection on the maximum size of the http body. If the image to be uploaded is larger, an error is received (403). To get around this problem just go to the ingress and add 0 to:
`` `
nginx.ingress.kubernetes.io/proxy-body-size: "0"
`` `






