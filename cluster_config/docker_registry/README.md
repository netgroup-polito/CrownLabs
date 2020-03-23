# Docker registry

## Table of contents
- [What is it](#docker-registry-first)
- [Why do we need it](#table-of-contents)
- [Docker registry Helm Chart](#docker-registry)
  - [Configuration]()
- [Installing the chart]()
- [Warning: max HTTP body size in ingress controller]()
  
 
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
These are our modification of the values file.

  1. Annotations with respect to the authentication in the Ingress controller.
```
annotations:
   nginx.ingress.kubernetes.io/auth-realm: Authentication Required - ok
   nginx.ingress.kubernetes.io/auth-secret: basic-auth
   nginx.ingress.kubernetes.io/auth-type: basic
```

  2. Annotations with respect to the configuration of the Ingress controller.
```
ingress:
  enabled: true
  path: /
  # Used to create an Ingress record.
  hosts:
    - <>
  annotations: {}
    # kubernetes.io/ingress.class: nginx
    # kubernetes.io/tls-acme: "true"
  labels: {}
  tls:
    # Secrets must be manually created in the namespace.
     #- secretName: 
       - hosts:
         - <>
```

  3. Configuration of the persistent disk. Please dimension this disk appropriately, as VM images may be rather large and hence you may fill up this space in a very short time.
```
persistence:
  accessMode: 'ReadWriteOnce'
  enabled: true
  size: 100Gi
  storageClass: 'rook-ceph-block'
```

## Installing the Chart
To install the chart, use the following command:
```console
$ helm install stable/docker-registry -f configfile.yaml -n docker-registry --generate-name
```
Look look [here](https://github.com/helm/charts/tree/master/stable/docker-registry#configuration) for a complete configuration guide.


## Warning: max HTTP body size in ingress controller
By default, the nginx ingress controller has a limit on the maximum size of the http body. If the image to be uploaded is larger, an error is received (403). To avoid this problem, modify the configuration of your  ingress controller and add a new annotation:
```
nginx.ingress.kubernetes.io/proxy-body-size: <>
```






