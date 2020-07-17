# Guide for creating and uploading VMs in CrownLabs

## Create and upload custom VMs to the cluster

The approach currently suggested for the creation and upload of VMs in CrownLabs involves the usage of the [setup-crownlabs-vm script](setup-crownlabs-vm/setup-crownlabs-vm.sh).
In a nutshell, it takes care of automatically:
- Create new VirtualBox VMs and install the guest OS (`xubuntu`);
- Install custom software and the tools required by CrownLabs with Ansible;
- Convert the resulting virtual HDD to the correct format and upload it to a Docker Registry.

Once the VM image has been correctly uploaded to the registry, it is possible to setup a new laboratory associated to it, as explained [here](../courses).


### Quick start guide

0. Open the `setup-crownlabs-vm` directory;
1. Customize the basic information at the beginning of the [setup-crownlabs-vm.sh](setup-crownlabs-vm/setup-crownlabs-vm.sh) script (e.g. xubuntu version and docker registry configuration);
2. Execute `./setup-crownlabs-vm.sh <vm-name> create --no-guest-additions` to create a new VM;
3. Wait for the installation to complete and login into the VM
4. Execute `./setup-crownlabs-vm <vm-name> configure ansible/xubuntu-<choose>-crownlabs-playbook.yml` to configure the VM;
5. Once the setup completed, shutdown the VM
6. Execute `./setup-crownlabs-vm <vm-name> export crownlabs` to export the VM.

Please, refer to the corresponding [README file](setup-crownlabs-vm/README.md) for additional information about this script.

Some predefined playbook are available in the [ansible](setup-crownlabs-vm/ansible) folder (those named `crownlabs` take care of installing the tools required by CrownLabs).
In case you want to install different software, you can simply create a new ansible script, and optionally new ansible roles.
Alternatively, it is possible to configure a basic VM (i.e. with the `xubuntu-base-crownlabs` playbook) and proceed with the manual installation of custom software.

**Warning**: the ansible playbooks remove many programs present in a default installation of `xubuntu` and disable different standard services, in order to limit as much as possible the size of the final VM image and speedup the boot process.
In particular, they disable the automatic startup of the graphical interface, since the remote desktop is accessed through a distinct VNC session.
Hence, if the VM is rebooted during the local configuration, it may be necessary to manually start the graphical interface with `sudo systemctl isolate graphical.target`.

### Conversion of existing VMs

The configuration of an existing VM for the execution in CrownLabs can be performed using the `xubuntu-base-crownlabs` playbook (i.e. executing `./setup-crownlabs-vm <vm-name> configure ansible/xubuntu-base-crownlabs-playbook.yml`). Then, it is possible to proceed exporting the VM as in the standard procedure.

**Warning**: depending on your configuration (e.g. virtualization platform and guest OS), the script may require some additional amount of tuning to work. Additionally, verify in advance that the packages automatically removed are of no use to you.

**Warning**: as of today, the only desktop environment officially supported by the playbooks and suggested for usage in CrownLabs is XFCE. The adoption of different desktop environments will probably require the customization of the ansible playbooks and it is not guaranteed to achieve acceptable results in terms of performance.


### Conversion of existing VMs with the legacy scripts

The configuration and conversion of existing VMs for usage in CrownLabs can also be performed by means of some legacy scripts:

- [prepare-vm.sh](legacy-scripts/prepare-vm.sh), to install and configure the tools required by CrownLabs (to be executed from within the VM);
- [convert-vm.sh](legacy-scripts/convert-vm.sh), to convert a `vdi` disk to the `qcow2` format and output the Docker file to prepare the final image.

Once, the `convert-vm.sh` script terminates, it is necessary to build the docker image from the resulting `Dockerfile`, and push it to the selected Docker Registry:
```
$ docker login <registry_url>
$ docker build -t <registry_url>/<image_name>:latest docker-output/
$ docker push <registry_url>/<image_name>:latest
```

**Warning**: this approach is not suggested, since these scripts perform less optimizations and require more manual intervention compared to `setup-crownlabs-vm`. Additionally, they are not guaranteed to be maintained up-to-date.


### CrownLabs dependencies

The tools that are required by CrownLabs and automatically installed by the different scripts are:
- **TigerVNC server**, which allows to connect to the VM desktop from a remote machine;
- **NoVNC with websockify server**, which allows the above connection to be established through HTTP/HTTPS;
- **Prometheus node exporter**, which exports some run-time information of the VM (e.g., CPU/memory consumption) to the Prometheus monitoring system, running on the Kubernetes cluster
- **cloud-init**, which enables to customize some running parameters of the VM at boot time.


## Cleaning up private Docker registry

Since the private registry can become crowded with new images, you can control its resources with two scripts:
- list of available images: [list-img-from-registry.sh](docker-scripts/list-img-from-registry.sh)
- delete an existing image: [del-img-from-registry.sh](docker-scripts/del-img-from-registry.sh)

Remember either to login in the registry before running the scripts, or to customize `user/password` in the above scripts.
**NOTE**: in case the *delete* action returns an error `405 Method Not Allowed`, modify the config of your Docker registry to enable the `DELETE` action, which is usually disabled by default, and restart the service.
