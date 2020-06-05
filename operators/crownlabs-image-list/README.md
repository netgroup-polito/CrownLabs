# CrownLabs Image List

This folder contains the material necessary to gather the list of available images from a Docker Registry and expose it as an ImageList custom resource, to be consumed from the front-end web-page.

## Main components
The core components necessary for building and executing the application are:

* [image-lists-crd.yaml](image-lists-crd.yaml): the definition of the ImageList CRD.
* [update-crownlabs-image-list.yaml](update-crownlabs-image-list.yaml): the definition of the Kubernetes resources necessary to run the application.
* [update-crownlabs-image-list.py](update-crownlabs-image-list.py): the source code of the python application responsible for retrieving the list of images and storing it as an ImageList object.

## Usage

```
usage: update-crownlabs-image-list.py [-h]
    --advertised-registry-name ADVERTISED_REGISTRY_NAME
    --image-list-name IMAGE_LIST_NAME
    --registry-url REGISTRY_URL
    [--registry-username REGISTRY_USERNAME]
    [--registry-password REGISTRY_PASSWORD]
    --update-interval UPDATE_INTERVAL

Periodically requests the list of images from a Docker registry and stores it as a Kubernetes CR

Arguments:
  -h, --help            show this help message and exit
  --advertised-registry-name ADVERTISED_REGISTRY_NAME
                        the host name of the Docker registry where the images can be retrieved
  --image-list-name IMAGE_LIST_NAME
                        the name assigned to the resulting ImageList object
  --registry-url REGISTRY_URL
                        the URL used to contact the Docker registry
  --registry-username REGISTRY_USERNAME
                        the username used to access the Docker registry
  --registry-password REGISTRY_PASSWORD
                        the password used to access the Docker registry
  --update-interval UPDATE_INTERVAL
                        the interval (in seconds) between one update and the following
```

## How to build
The creation of the Docker image is automatized through the [Makefile](Makefile). Please, customize the repository configuration according to your requirements.

```bash
$ make
$ make push
```

## How to deploy

```bash
$ kubectl create -f image-lists-crd.yaml
$ kubectl create -f update-crownlabs-image-list.yaml
```