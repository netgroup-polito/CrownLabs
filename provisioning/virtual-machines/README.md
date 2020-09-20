# Guide for creating and uploading VMs in CrownLabs

## Create and upload custom VMs to the cluster

The suggested approach to create and upload new VMs in CrownLabs involves the usage of the `setup-crownlabs-vms.sh` script, which takes care of:
- **Creating** a new VirtualBox VMs and installing the guest OS (`xubuntu`, which represents a good compromise between the necessity to have a friendly, GUI-based guest OS and the resources consumed by the graphical interface);
- **Installing** additional software (e.g., application software packages) and the background tools required by CrownLabs with Ansible (e.g., VNC server);
- **Converting** the resulting virtual HDD to the correct format and upload it to the proper Docker Registry.

Jump to the [Create and configure CrownLabs VMs guide](README-VM-create.md) for creating and uploading VMs in CrownLabs; the above guide contains also detailed instructions for running and customizing the `setup-crownlabs-vms.sh` script.

Once the VM image has been correctly uploaded to the registry, a new laboratory can be configured in the CrownLab live environment, as explained [in the Course/Lab setup](../courses) documentation section.


## Create your own VM manually
We strongly suggest **not to create your own VM manually**, relying instead on the [setup-crownlabs-vm.sh](setup-crownlabs-vm.sh) script and the [Create and configure CrownLabs VMs guide](README-VM-create.md).

In case you are a very experienced user and you want to proceed manually, we suggest to (1) read how the `setup-crownlabs-vms.sh` script works (look at the [VM setup documentation](README-VM-create.md)), then (2) continue with this guide.


### Conversion of existing VMs

The configuration of an existing VM for the execution in CrownLabs can be performed using the `ubuntu-base` playbook (i.e. executing `./setup-crownlabs-vm <vm-name> configure ansible/playbooks/ubuntu-base.yml`). Then, you can export the VM as documented in the `setup-crownlabs-vms.sh` script.

**Warning**: depending on your configuration (e.g. virtualization platform and guest OS), the script may require some additional amount of tuning to work. Additionally, verify in advance that the packages automatically removed are of no use to you.

**Warning**: currently CrownLabs supports only VMs running the XFCE desktop environment. The adoption of a different desktop environments will probably require the customization of the ansible playbooks and it is not guaranteed to achieve acceptable results in terms of performance. In fact, given the amount of resources consumed in a cloud environment by the GUI subsystem of the VM, we strongly suggest not to use a full fledged graphical environment (e.g., Ubuntu), but to privilege more resource-saving ones, such as XFCE (e.g., Xubuntu).


### Conversion of existing VMs with the legacy scripts

The configuration and conversion of existing VMs for usage in CrownLabs can also be performed by means of some legacy scripts:

- [prepare-vm.sh](legacy-scripts/prepare-vm.sh), to install and configure the tools required by CrownLabs (to be executed from within the VM);
- [convert-vm.sh](legacy-scripts/convert-vm.sh), to convert a `vdi` disk to the `qcow2` format and output the Docker file to prepare the final image.

Once, the `convert-vm.sh` script terminates, you have to build the docker image from the resulting `Dockerfile`, and push it to the selected Docker Registry:
```
$ docker login <registry_url>
$ docker build -t <registry_url>/<image_name>:latest docker-output/
$ docker push <registry_url>/<image_name>:latest
```

**Warning**: this approach is not suggested, since these scripts perform less optimizations and require more manual intervention compared to `setup-crownlabs-vm`. Additionally, they are not guaranteed to be maintained up-to-date.


## Cleaning up private Docker registry

Since the private registry can become crowded with new images, you can control its resources with two scripts:
- list of available images: [list-img-from-registry.sh](docker-scripts/list-img-from-registry.sh)
- delete an existing image: [del-img-from-registry.sh](docker-scripts/del-img-from-registry.sh)

Remember either to login in the registry before running the scripts, or to customize `user/password` in the above scripts.
**NOTE**: in case the *delete* action returns an error `405 Method Not Allowed`, modify the config of your Docker registry to enable the `DELETE` action, which is usually disabled by default, and restart the service.
