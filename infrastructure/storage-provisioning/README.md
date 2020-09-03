# Storage Provisioning - Rook-Ceph

[Rook](https://rook.io/) is a cloud-native storage orchestrator for Kubernetes.
Among the different alternatives supported by Rook, we adopted [Ceph](https://ceph.io) as the selected storage provider.

## Install Rook-Ceph

### Deploy the Rook Operator
In order to set-up Rook-Ceph, it is first necessary to deploy the Rook Operator, together with the set of CRDs and permissions required for its operations.
Adopting the out-of-the-box configurations, it is possible to leverage the manifests provided within the Rook repository for its deployment:

```bash
$ export ROOK_VERSION=1.4
$ kubectl apply -f https://raw.githubusercontent.com/rook/rook/release-${ROOK_VERSION}/cluster/examples/kubernetes/ceph/common.yaml
$ kubectl apply -f https://raw.githubusercontent.com/rook/rook/release-${ROOK_VERSION}/cluster/examples/kubernetes/ceph/operator.yaml
```

### Create the Ceph Clusters
Once the Rook Operator is ready, it is possible to trigger the creation of the desired Ceph Clusters through the definition of the corresponding `CephCluster` CRs.
Two different clusters are defined in the following, one leveraging faster SSD storage (yet, with lower available capacity) and the other backed by traditional HDDs.
While representing a working example, these manifests need to be customized depending on the specific characteristics of the cluster where they are applied (e.g. to define which drives belong to each cluster).
Additionally, it may be necessary to create the `namespace` where the secondary cluster is going to be defined.

```bash
$ kubectl create -f ceph-clusters/ceph-cluster-primary.yaml
$ kubectl create -f ceph-clusters/ceph-cluster-secondary.yaml
```

### Deploy the Rook Toolbox

The Rook toolbox is a container with common tools used for rook debugging and testing. Specifically, it allows to interact with the `ceph` cluster to check its status and trigger maintenance operations.
In order to deploy the toolbox, please refer to the illustrative `deployment` definition available in the [official documentation](https://rook.io/docs/rook/v1.4/ceph-toolbox.html) (a different instance of the toolbox needs to be created for each Ceph cluster).

Once the toolbox is correctly deployed, it is possible to enter a shell with:

```bash
$ kubectl -n rook-ceph exec -it $(kubectl -n rook-ceph get pod -l "app=rook-ceph-tools" -o jsonpath='{.items[0].metadata.name}') -- /bin/bash
```

Once in the toolbox's shell, it is possible to run, e.g., `ceph status` to verify the status of the cluster.

## Upgrade Rook-Ceph

### Upgrade Rook
To upgrade Rook, it is necessary to edit the image version of the operator deployment. In turn, it will proceed to upgrade all the other components.
Patch release upgrades (e.g. from v1.4.1 to v1.4.2) are as easy as issuing:

```bash
$ kubectl -n rook-ceph set image deploy/rook-ceph-operator rook-ceph-operator=rook/ceph:v1.4.2
```

The upgrade between actual versions (e.g. from v1.3.10 to v1.4.2), on the other hand, typically involves additional preparation steps to update the CRD definitions and the RBAC settings.
To this end, it is suggested to carefully follow the specific instructions available on the [rook.io](https://rook.io/docs/rook/v1.4/ceph-upgrade.html) website.

### Upgrade Ceph
To upgrade Ceph, it is necessary to edit the image version specified within the `CephCluster` CR.
With reference to the clusters previously created, this operation can be completed with:

```bash
$ export CEPH_IMAGE='ceph/ceph:v15.2.4'
$ kubectl -n rook-ceph patch CephCluster rook-ceph --type=merge -p "{\"spec\": {\"cephVersion\": {\"image\": \"$CEPH_IMAGE\"}}}"
$ kubectl -n iscsi-rook-ceph patch CephCluster iscsi-rook-ceph --type=merge -p "{\"spec\": {\"cephVersion\": {\"image\": \"$CEPH_IMAGE\"}}}"
```

## Test the PVC provisioning
To test Rook using an illustrative example, follow those commands, which will create a `StorageClass` and some `PersistentVolumeClaims` mounted by the corresponding applications.

```bash
$ kubectl create -f examples/storageclass.yaml
$ kubectl create -f examples/mysql.yaml
$ kubectl create -f examples/wordpress.yaml
```

Both of these apps creates a block volume and mount it to their respective pod. You can see the Kubernetes volume claims by running the following:

```bash
$ kubectl get pvc
NAME             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
mysql-pv-claim   Bound    pvc-2a53d32d-0f38-4d5a-816f-de09d07768f6   20Gi       RWO            rook-ceph-block   134m
wp-pv-claim      Bound    pvc-8d5ec321-eca5-47a1-817a-bb0d04d7064e   20Gi       RWO            rook-ceph-block   134m
```

After that you can delete test with commands
```bash
$ kubectl delete -f examples/wordpress.yaml
$ kubectl delete -f examples/mysql.yaml
```
