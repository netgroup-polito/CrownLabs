# Backup system - Velero

[Velero](https://velero.io/) is a backup & migration cloud native software designed to perform disaster recovery and migrate resources across cluster.

The main purpose of Velero in CrownLabs is to _mitigate_ the following eventualities:
- failed/buggy updates
- unintentional resources deletion
- catastrophical physical cluster destruction/loss of all control plane nodes (*)

(*) this would require backups to be stored in a physically different/distant location relatively to the cluster.

Notes:
- storing backups on the cluster itself could help to mitigate unintentional resource deletions but it is not a great strategy, since the deletion of volumes or related resources could lead to the deletion of backups themselves.

## How has been deployed

The backup solution chosen for CrownLabs is based on [Velero](https://velero.io/). 

Velero can store backups of Kubernetes resources and possibly associated volumes on several cloud native/proprietary storage solutions (such as AWS, GCP, Azure, VMWare, OpenStack, etc.).

In case of full bare-metal environments, it is possible to use [Minio](https://min.io/), an open source, Kubernetes-native object storage solution which is compliant with Amazon's S3 API. This provides a valid storage target for Velero, provided that the mounted data path is not being stored on the cluster itself.

The provided sample configuration includes the deployment of a Minio instance that uses an iSCSI target (hosted outside of the cluster) for storing data.

Velero offers not only the possibility to backup resources but also the content of the volumes present on the cluster. This can be done rather automatically by mean of Volume Snapshotters but recent versions of Velero include an integration with [Restic](https://restic.net/) (that can be deployed automatically by Velero itself) which enables volumes snapshots without requiring snapshotters and regardless of the storage class of the volumes to backup.

### Deployment

```bash
helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts
helm install velero vmware-tanzu/velero --namespace velero --create-namespace -f values.yaml
```

See the [Velero Helm chart reference](https://github.com/vmware-tanzu/helm-charts/blob/main/charts/velero/README.md) for more information about values.

#### Prometheus alert

A simple _PrometheusRule_ is provided to ease monitoring of backups. It would trigger alerts in case of total and/or partial backup failures.
To deploy the rule run the following:

```bash
kubectl apply --namespace velero -f prometheus-alert.yaml
```

## Usage

It is suggested to use Velero's client in order to interact with its deployment on the cluster. [The official documentation](https://velero.io/docs/) provides the information about how to install the client. In particular [here](https://velero.io/docs/v1.6/basic-install/#install-the-cli) it is possible to get further instructions for the currently installed version.

### Prerequisites: volumes preparation

The volume snapshotting of Velero through Restic can work on a whitelist or blacklist basis. In the case of CrownLabs it has been chosen to use the whitelist approach since crucial volumes to be persisted can be identified. In order to request Velero to backup a volume it is necessary to add a special annotation to the pods to which such volumes are mounted: `backup.velero.io/backup-volumes=VOL_NAME_2_,VOL_NAME_2,...`. 

### Backup schedule

To create a backup schedule it is sufficient to run 
```bash
velero schedule create <SCHEDULE_NAME> --schedule="0 2 * * *"
```
The previous example would trigger a backup every day at 2 AM, for every resource found in every namespace. 
- It is possible to specify a time to live for created backups through the `--ttl` switch (that accepts a string like `12h34m56s`) after which the backup and the associated data would be removed from the backup location. It defaults to 30 days. 
- It is possible to specify which namespaces have to be backed up through the switch `--include-namespaces` (comma separated). In case this is not specified, any namespace will be included in the backup. 

### Single backup

It is possible to trigger single backups using `velero backup create <BACKUP_NAME>` together with the other switches and/or in combination with `--from-schedule <SCHEDULE_NAME>` to inherit options from the specified schedule.

### Restore

It is possible to make Velero restore a backup through the option `velero restore create <BACKUP_NAME>`. Options are available to perform a partial restore (e.g., specify a single namespace only).

### Monitoring

The previous commands will request Velero to perform respective operations but the command would return immediately, effectively just creating the corresponding Velero resource.
To monitor the progress of the requested operation it is necessary to run `velero backup get` or `velero restore get` to obtain information about running operations and the corresponding `describe` command (instead of `get`), available at the end of each process, to see the results. 

## Disaster recovery

In case of disaster, it should be possible to recover the whole cluster by deploying Velero (and Minio) and starting the restore of the previous backup that would become available once the Velero operator has settled.
