# VM Guide

## Upload custom VMs to the cluster

### Dependencies

Before uploading your VM, you must run the [prepare-vm.sh](scripts/prepare-vm.sh) script from inside the VM.
The script will install and configure:
- **TigerVNC server**, which allows to connect to the VM desktop from a remote machine;
- **NoVNC with websockify server**, which allows the above connection to be established through HTTP/HTTPS;
- **Prometheus node exporter**, which exports some run-time information of the VM (e.g., CPU/memory consumption) to the Prometheus monitoring system, running on the Kubernetes cluster
- **cloud-init**, which enables to customize some running parameters of the VM at boot time.

To verify that the setup works, reboot the machine after running the `prepare-vm.sh` script.
From inside the machine, start a browser and connect to page `http://localhost:6080`, using password `ccroot`.

### Conversion to raw format

Once you made sure that the VM has been properly configured and runs smothly, shutdown again the VM and convert it to the `qcow2` format, which is used by the Kubernetes virtualization module (Kube-virt).
This can be done with the [convert-vm.sh](scripts/convert-vm.sh) script, typing the following command:

```sh
$ convert-vm.sh <your-vm>.vdi
```

**NOTES**:
- Virtualbox uses, by default, disks in VDI format, which is the format supported by this script. Other tools are available to convert your images into VDI, or directly into the QCOW2 `raw` format, which is used in the next steps of our processing.
- the above command assumes that the VM runs on a Linux host. If not, please transfer your image to a Linux machine and run the `convert-vm.sh` script from there.

The script generates a folder called `docker-output` in the directory of the `vdi` image, which contains (1) the converted image in `qcow2` format and (2) a `Dockerfile`.

### Create Docker and upload on private registry

For this step, you have to login in CrownLabs's Docker registry using the proper credentials that you created you set up the service:

```sh
$ docker login registry.crown-labs.ipv6.polito.it
```
**Note**: in case this command fails, try using `sudo` before the command itself. Depending on how Docker has been installed on your machine, you may have to use `sudo` before all the commands involving Docker.

Now you can build the Docker image with the following command:

```sh
$ docker build -t registry.crown-labs.ipv6.polito.it/<image_name>:latest docker-output/
```
where:
- `<image_name>` is a [tag](https://docs.docker.com/engine/reference/commandline/tag/), used by Docker, which can be used to identify better an image. Example values can be `fedora/httpd`, or `alice/networklabs`, and more.
- `registry.crown-labs.ipv6.polito.it` is the name of the private registry your are using, which is compulsory in Docker in order to tell the daemon that the private registry (instead of DockerHub) has to be used.

Note also that you have to run this command from the directory that contains `docker-output`.

You can check that your image is stored locally, on your host machine, with this command:

```sh
$ docker image list
```

Now simply login to the docker registry (with `sudo docker login <registry_url>`) and push the image (with `sudo docker push`).

```sh
$ docker push registry.crown-labs.ipv6.polito.it/<image_name>:latest
```

## Cleaning up private registry

Since the private registry can become crowded with new images, you can control its resources with two scripts:
- list of available images: [list-img-from-registry.sh](scripts/list-img-from-registry.sh)
- delete an existing image: [del-img-from-registry.sh](scripts/del-img-from-registry.sh)

Remember either to login in the registry before running the scripts, or to customize `user/password` in the above scripts.


## Run on the cluster

To run the VM on the cluster you simply have to deploy two resources:

- a `Secret` containing the cloud-init configuration of the VM ([template](templates/cloudinit.yaml))
- a `VirtualMachineInstance` that uses as image the one pushed on Docker ([template](templates/vm.yaml))

**Warning**: the name of the secret referenced by the VM manifest must match the name of the secret.
