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

### Required namespaces for MyDrive and instance snapshots

Before deploying CrownLabs, manually create the two shared namespaces used by MyDrive and the public instance snapshot catalog. Neither the Helm charts nor the operators create these namespaces. The `--create-namespace` Helm option in the installation command below only creates the release namespace (`crownlabs-production`).

| Purpose | Default namespace | Umbrella chart value |
| --- | --- | --- |
| MyDrive: stores tenants' personal-drive PVCs | `mydrive-pvcs` | `operator.configurations.mydrivePVCsNamespace` |
| Instance snapshots: stores public snapshot resources and their DataVolumes/PVCs | `crownlabs-public-snapshots` | `operator.configurations.snapshotPublicNamespace` |

Create the namespaces if they do not already exist, and label the public snapshot namespace so it is covered by the validating webhooks:

```bash
kubectl create namespace mydrive-pvcs
kubectl create namespace crownlabs-public-snapshots
kubectl label namespace crownlabs-public-snapshots \
  crownlabs.polito.it/operator-selector=production --overwrite
```

The commands above use the default names and `operator.configurations.targetLabel`. If you customize them, use the same names and label in the commands and in your deployment configuration:

```yaml
operator:
  configurations:
    mydrivePVCsNamespace: mydrive-pvcs
    snapshotPublicNamespace: crownlabs-public-snapshots
    targetLabel: crownlabs.polito.it/operator-selector=production
```

These names already have defaults in the [operator chart](../../operators/deploy/operator/values.yaml); override them in your deployment values when using different namespaces. No source-code changes are required. For the standalone operator chart, omit the `operator` prefix; when running the binary directly, use `--mydrive-pvcs-namespace` and `--snapshot-public-namespace`. Configuring a name does not create the corresponding namespace.

### Instance snapshots

The main operator runs the snapshot controller when `operator.configurations.features.instanceSnapshot: true`. Snapshot and LocalVM admission checks also require `operator.configurations.features.webhooks: true` and `operator.webhook.enableValidating: true`; all three are enabled in the umbrella chart. Install the updated [InstanceSnapshot CRD](../../operators/deploy/crds/crownlabs.polito.it_instancesnapshots.yaml) before upgrading, and ensure CDI and the storage backend can clone the source PVCs.

The public snapshot namespace defaults to `crownlabs-public-snapshots`, explicitly configured in the umbrella chart's values. Override `operator.configurations.snapshotPublicNamespace` in your deployment values to change it. The operator's admission checks and the public snapshot publisher RoleBinding use this same value.

Provision the public catalog as described in [Required namespaces for MyDrive and instance snapshots](#required-namespaces-for-mydrive-and-instance-snapshots). Any other snapshot destination namespace must also exist and carry the label configured by `operator.configurations.targetLabel` so the validating webhooks cover it.

Publishing is controlled by `operator.configurations.snapshotPublisherGroup` (default `kubernetes:image-publisher`) and `operator.webhook.deployment.snapshotWebhookBypassGroups` (default `system:masters,kubernetes:admin`). The publisher group's RoleBinding grants snapshot creation only in the public namespace. Bypass groups skip creation authorization checks but still need Kubernetes RBAC permissions.

Clients must identify the creator through the immutable `crownlabs.polito.it/tenant` label instead of the former `spec.tenantRef` field. See the [InstanceSnapshot controller documentation](../../operators/README.md#crownlabs-instancesnapshot-controller) for resource examples, lifecycle, LocalVM disk-size checks and troubleshooting.

### Gateway API & Routing Configuration

CrownLabs uses **Envoy Gateway** implementing the Kubernetes Gateway API (`gateway.networking.k8s.io/v1`) for L7 traffic routing and authentication.

For detailed information regarding the Gateway Configuration Model, Control Flags, and Authentication mechanisms, please refer to the dedicated [Authentication and Centralized Gateway](docs/authentication-and-gateway.md) document.
---


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
