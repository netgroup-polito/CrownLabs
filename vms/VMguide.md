# VM Guide

## Prepare custom VMs

### Prerequisites

Before uploading your vm, you must run [requirements.sh](scripts/requirements.sh). The script will install and configure tigervnc server, novnc, node exporter and cloud-init.

### Convert and upload on docker-hub

Then you have to convert it to the ```qcow2``` format with ```qemu```:

```sh
qemu-img convert -f vdi -O qcow2 yourimage.vdi output.qcow2
```

Now create the following Dockerfile:

```text
FROM scratch
ADD output.qcow2 /disk/
```

And then build and push the image.

```sh
docker build -t tag/name:latest .
docker push tag/name:latest
```

## Run on the cluster

To run the vm on the cluster you simply have to deploy to resources:

- A `Secret` containing the cloud-init configuration of the vm ([template](templates/cloudinit.yaml))
- A `VirtualMachineInstance` that uses as image the one pushed on docker ([template](templates/vm.yaml))

**Warning**: the name of the secret referenced by the vm manifest must match the name of the secret.
