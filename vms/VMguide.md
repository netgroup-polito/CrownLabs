# VM Guide

## Upload custom VMs to the cluster

### Dependencies

Before uploading your VM, you must run the [prepare_vm.sh](scripts/prepare_vm.sh) script from inside the VM.
The script will install and configure:

- TigerVNC server, which allows to connect to the VM desktop from a remote machine;
- NoVNC with websockify server, which allows the above connection to be established through HTTP/HTTPS;
- Prometheus node exporter, which exports some run-time information of the VM (e.g., CPU/memory consumption) to the Prometheus monitoring system, running on the Kubernetes cluster
- cloud-init, which enables to customize some running parameters of the VM at boot time.

To verify that the setup works, reboot the machine after running the `prepare_vm.sh` script.
From inside the machine, start a browser and connect to page `http://localhost:6080`, using password `ccroot`.

### Conversion 

Once you made sure that the VM has been properly configured and runs smothly, shutdown again the VM and convert it to the `qcow2` format, which is used by the Kubernetes virtualization module (Kube-virt).
This can be done with the [convert_vm.sh](scripts/convert_vm.sh) script, typing the following command:

```sh
$ convert_vm.sh <your_vm>.vdi
```

**NOTE**: the above command assumes that the VM runs on a Linux host. If not, please transfer your image to a Linux machine and run the `convert_vm.sh` script from there.

The script generates a folder called `docker_output` in the directory of the `vdi` image, which contains (1) the converted image in `qcow2` format and (2) a `Dockerfile`.

### Upload on Crown Team registry

To upload the image on the Crown Team's docker registry, first you have to login into it:

```sh
$ docker login registry.crown-labs.ipv6.polito.it
Username: netgroup
Password:
```

Now you can build the Docker image with the following command:

```sh
$ docker build -t registry.crown-labs.ipv6.polito.it/netgroup/<image_name>:latest docker_output/
```

**Remember** to change `<image_name>` with the name you want to give to the image; **also** note that you have to run this command from the directory that contains `docker_output`.

Finally push the image with the following command:

```sh
$ docker push registry.crown-labs.ipv6.polito.it/netgroup/<image_name>:latest
```

## Run on the cluster

To run the VM on the cluster you simply have to deploy two resources:
- a `Secret` containing the cloud-init configuration of the VM ([template](templates/cloudinit.yaml))
- a `VirtualMachineInstance` that uses as image the one pushed on Docker ([template](templates/vm.yaml))

**Warning**: the name of the secret referenced by the VM manifest must match the name of the secret.
