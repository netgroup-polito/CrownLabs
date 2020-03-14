# VM Guide

## Upload custom VMs to the cluster

### Dependencies

Before uploading your vm, you must run [this script](scripts/prepare_vm.sh) inside the vm. The script will install and configure:

- tigervnc server
- novnc with websockify server
- prometheus node exporter
- cloud-init

To verify that the setup works, try to reboot the machine and from a browser visit `<IP_of_the_vm>:6080` (by now the password to connect to vnc is `ccroot`).

### Conversion and upload

At this point shut down the vm and convert it using [this script](scripts/convert_vm.sh). The usage is the following:

```sh
$ convert_vm.sh <your_vm>.vdi
```

The script generates a folder called `docker_output` in the directory of the `vdi` image containing the converted image in `qcow2` format and a `Dockerfile`. Build the image with:

```sh
$ docker build -t user/image:latest docker_output/
```

Now simply login to the docker registry (with `docker login <registry>`) and push the image (with `docker push`).

## Run on the cluster

To run the vm on the cluster you simply have to deploy two resources:

- A `Secret` containing the cloud-init configuration of the vm ([template](templates/cloudinit.yaml))
- A `VirtualMachineInstance` that uses as image the one pushed on docker ([template](templates/vm.yaml))

**Warning**: the name of the secret referenced by the vm manifest must match the name of the secret.
