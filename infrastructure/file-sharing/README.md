# File Sharing - Nextcloud

Nextcloud is a suite of client-server software for creating and using file hosting services. Nextcloud is free and open-source, which means that anyone is allowed to install and operate it on their own private server devices.

More info at [Nextcloud's website](https://nextcloud.com)
## Pre-requisites
Here we assume that the following operators are installed and configured in the K8s cluster:
* [ROOK](https://rook.io/)
* [NGINX Ingress Controller](https://github.com/kubernetes/ingress-nginx)
* [cert-manager](https://cert-manager.io/)
* A namespace in K8S cluster called **nextcloud**

## Redis
We can significantly improve our Nextcloud server performance with memory caching, where frequently-requested objects are stored in memory for faster retrieval.
Having multiple Nextcloud server instances a memory caching is indispensable in order to prevent conflicts when same file is requested by different users at the same time.

**Redis** is an excellent modern memcache to use for distributed caching, and as a key-value store for Transactional File Locking because it guarantees that cached objects are available for as long as they are needed.

To install Redis we can use the [Helm Chart from Bitnami](https://github.com/bitnami/charts/tree/master/bitnami/redis).

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm upgrade --install --values redis-values.yaml redis-nextcloud bitnami/redis
```

## Postgres cluster

Proceed installing the Postgres cluster, by applying the [nextcloud-postgres-cluster-manifest.yaml](manifests/nextcloud-postgres-cluster-manifest.yaml), which will be consumed by Nextcloud:
```bash
# create a Postgres cluster
kubectl create -f nextcloud-postgres-cluster-manifest.yaml
```
For more information about the Postgres Operator please refer to this [README.md](../identity-provider/README.md)

## CEPHfs for the PVC
The Ceph File System, or CephFS, is a POSIX-compliant file system built on top of Ceph’s distributed object store, RADOS. CephFS endeavors to provide a state-of-the-art, multi-use, highly available,
and performant file store for a variety of applications, including traditional use-cases like shared home directories, HPC scratch space, and distributed workflow shared storage.
Now we will create a Persistent Volume Claim which will be attached to the Nextcloud Deployment. Applying the [nextcloud-pvc.yaml](manifests/nextcloud-pvc.yaml) we will have a PVC of 700 Gi in size provisioned
by the **csi-cephfs** storage class.
```bash
kubectl create -n nextcloud -f nextcloud-pvc.yaml
```
## Nextcloud

### Install Procedure
Now we can proceed by installing Nextcloud. We will apply the following manifests:
* [nextcloud-ingress.yaml](manifests/nextcloud-ingress.yaml), to expose Nextcloud to the internet;
* [nextcloud-php-configmap.yaml](manifests/nextcloud-php-configmap.yaml), add here the configuration options for php if you have particular needs;
* [nextcloud-service.yaml](manifests/nextcloud-service.yaml), clusterIP service for the deployment;
* [nextcloud-admin-credentials-secret.yaml](manifests/nextcloud-admin-credentials-secret.yaml), this credentials will be used during the creation of the admin user;
* [nextcloud-deployment.yaml](manifests/nextcloud-deployment.yaml), the deployment of Nextcloud.

For more information on the docker image please check the following section on [dockerhub](https://hub.docker.com/_/nextcloud/).

```bash
kubectl -n nextcloud -f nextcloud-ingress.yaml
kubectl -n nextcloud -f nextcloud-php-configmap.yaml
kubectl -n nextcloud -f nextcloud-service.yaml
kubectl -n nextcloud -f nextcloud-admin-credentials-secret.yaml
kubectl -n nextcloud -f nextcloud-deployment.yaml
```
### Configuration
Nextcloud cloud configuration is really vast argument, please consult the [official documentation](https://docs.nextcloud.com/server/19/admin_manual/configuration_server/index.html).
Please check also the above mentioned manifests to configure the deployment as you want. After the installation for sure you will need
to modify the nextcloud config file named **config.php**. Usually it is found in ```/var/www/html/config/config.php```.
In [nextcloud-deployment.yaml](manifests/nextcloud-deployment.yaml) the replica size is set to ```1``` because we need only one instance during installation and configuration of Nextcloud.
After the initial set up you can change the size and apply again the manifest or use the following command:
```
kubectl scale --replicas=3 deployment nextcloud
```
### OIDC Login
The authentication through Keycloak is made possible thanks to a third party application called **nextcloud-social-login** that is found in the Nextcloud's App website.
Here is the [Github repo](https://github.com/zorn-v/nextcloud-social-login) of the application.

### Clients
Clients for different platforms are available on the [official website](https://nextcloud.com).
