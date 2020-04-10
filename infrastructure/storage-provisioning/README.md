# Storage Provisioning - Rook

Rook is a cloud-native storage orchestrator for Kubernetes.
In this scenario we used Rook with Ceph storage provider.

## Install Rook-Ceph

To install Rook-Ceph apply the following commands.
Ceph uses a directory under /var/lib/Rook that is a mount point of a free partition.

```bash
$ kubectl create -f manifests/common.yaml
$ kubectl create -f manifests/operator.yaml
# edit cluster.yaml with your preferences before deploy it
$ kubectl create -f manifests/cluster.yaml
$ kubectl create -f manifests/toolbox.yaml
```

## Test

To check the status of ceph you can run the following command to open toolbox's shell.

```bash
$ kubectl -n rook-ceph exec -it $(kubectl -n rook-ceph get pod -l "app=rook-ceph-tools" -o jsonpath='{.items[0].metadata.name}') bash
```

After that in toolbox's shell you can run

```
root$ ceph status
```

To test Rook follow those commands.

```bash
$ kubectl create -f examples/storageclass.yaml
$ kubectl create -f examples/mysql.yaml
$ kubectl create -f examples/wordpress.yaml
```

Both of these apps creates a block volume and mount it to their respective pod. You can see the Kubernetes volume claims by running the following:

```
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
