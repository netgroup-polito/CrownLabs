# How to deploy CrownLabs

The CrownLabs business logic is composed of multiple containerized components, which are necessary to implement the desired services.
Specifically, it encompasses multiple Kubernetes operators, implementing the server-side logic, as well as a web-based dashboard, which exposes the different functionalities to the end users.

To simplify the deployment and the configuration of the different components, CrownLabs leverages an [Helm](https://helm.sh/) chart.
This folder contains the parent Helm chart which depends upon the different sub-charts responsible for the installation of the single components (e.g. the dashboard and each operator), available in the respective folders.
In the following, it is presented a brief description of the different steps required to deploy CrownLabs on your own cluster.

## Pre-requirements

CrownLabs relies upon a Kubernetes Cluster for the orchestration of the different components. Additionally, it depends on multiple infrastructural components, as better detailed in the [infrastructure folder](../../infrastructure). In particular, [KubeVirt](../../infrastructure/virtualization/README.md) is required in order to spawn virtual machines on top of the Kubernetes cluster.

## Deploying the Custom Resource Definitions (CRDs)

Before deploying the different CrownLabs components, and in particular the operators, it is necessary to install the CRDs they depend on.

At the moment, this operation is not automated by the Helm chart, and can be performed with the following command (from the CrownLabs root directory):

```bash
kubectl apply -f operators/deploy/crds
```

## Deploying CrownLabs

Once the CRDs have been correctly installed, it is possible to deploy CrownLabs.

First, it is necessary to configure the different parameters (e.g. number of replicas, URLs, credentials, ...), depending on the specific set-up.
In particular, this operation can be completed creating a copy of the [default configuration](values.yaml), and customizing it with the suitable values.

Then, it is possible to proceed with the deployment/upgrade of CrownLabs (all commands are relative to the CrownLabs root directory):

```bash
# Get the version to be deployed (e.g. the latest commit on master)
git fetch origin master
VERSION=$(git show-ref -s origin/master)

# Update the sub-chart dependencies
helm dependency update deploy/crownlabs

# Package the Helm chart with the desired version
helm package deploy/crownlabs --app-version=${VERSION}

# Perform the CrownLabs installation/upgrade
helm upgrade crownlabs crownlabs-*.tgz \
  --install --create-namespace \
  --namespace crownlabs-production \
  --values path/to/configuration.yaml \
  --set global.version=${VERSION}
```
