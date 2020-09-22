# Create and configure CrownLabs VMs

This folder contains different scripts that automate the creation and configuration of the VMs used in *CrownLabs*.

All the functionalities can be triggered executing the `setup-crownlabs-vms.sh` script, which can be customized by changing some installation parameters (e.g. Ubuntu distribution and version, VM name, credentials, local information).

## Requirements:
- **Linux Desktop host**: since the configuration of the VM may require some manual step to be completed *inside* the VM, the VM has to be started with the graphical interface, hence the host that runs the `setup-crownlabs-vms.sh` script needs to be a Desktop Linux (not server). Some steps of the scripts may be working also on Ubuntu for Windows, but we provide no guarantees in this respect.
- **Virtualbox**: a lightweight virtualizer, used by this script to create and customize the basic environments of the guest operating system.
- **curl**: used to download the guest operating system installer (e.g., the ISO image of the guest operating system).
- **Ansible** (>= 2.8): used to install and customize any additional required packages (e.g., applications required in the VM).
- **ssh**: required by ansible to interact with the VMs and configure the additional software required in the guest environment (e.g., additional applications).
- **sshpass**: required by ansible to (silently) authenticate with the VM when establishing the SSH connection.
- **virt-sparsify**: used to export the resulting VM images into the format required by CrownLabs, compacting the resulting disk image in order to save space (and decrease the boot time).
- **docker**: used to push the resulting VM image into the docker registry used by CrownLabs to keep all the VMs available.

On **Ubuntu 20.04**, all the above dependencies can be installed with the following command:
```
# Install required packages
sudo apt install virtualbox ansible curl ssh sshpass docker.io libguestfs-tools

# Add yourself to the 'docker' group
sudo sudo adduser <your_user> docker
newgrp docker
```

<a name="ansible"></a>
## Installing software packages with Ansible: overview

The `setup-crownlabs-vms.sh` script makes use of Ansible Playbooks to install additional software packages in the VM (e.g., applications that you need to run in the guest OS).
For instance, each course may have its own playbook that installs the software packages required by the students of that course.
The [ansible/playbooks](ansible/playbooks) folder includes some playbooks for courses active at Politecnico di Torino; as an example, the [ubuntu-netlab playbook](ansible/playbooks/ubuntu-netlab.yml) has been defined for a computer network course and installs `Wireshark`, `Docker`, `polycube` and `GNS3`.

In case you want to install different software, i.e., *add/remove* new tasks, you can simply create a new ansible script, and optionally new ansible roles, as in [ubuntu-netlab.yml](ansible/playbooks/ubuntu-netlab.yml).
Then, you have to start the `setup-crownlabs-vms.sh` script with the Ansible playbook that contains the instructions to build and customize your VM.
Given the modularity of Ansible, you can define new playbooks that refer to already existing Ansible tasks, hence leveraging the install procedure defined in other playbooks to (partially) configure your VM as well.

Among the existing tasks, the first one (`ubuntu-pre`) is rather generic and aims at cleaning up the system and removing unnecessary packages, shrinking the size of the VM disk to a smaller size.
However, in case you want to export the VM in *.ova* format, disk shrinking can take place only by manually executing the post-install tasks listed in [ansible/roles/ubuntu-post/files/README](ansible/roles/ubuntu-post/files/README).

All the playbooks available in the [ansible/playbooks](ansible/playbooks) folder can be used both to configure a vanilla VM (i.e. to be used directly in VirtualBox and exported in the *.ova* format) or a CrownLabs VM.
In the latter case, they install also all the tools required by CrownLabs to work (e.g., the software required to enable the remote access to the VM through an HTTPS session) through the `crownlabs` role, which should be included in any other playbook. The selection between the two operating modes can be performed through a flag of the `setup-crownlabs-vms.sh` script (more details in the guide below).

Alternatively, you can create a basic VM (i.e. with the `ubuntu-base` playbooks) and proceed with the manual installation of custom software.

**Warning**: Crownlabs ansible playbooks remove many programs present in a default installation of `xubuntu` and disable different standard services in order to reduce the size of the final VM image and speedup the boot process.
In particular, they disable the automatic startup of the graphical interface, since the remote desktop is accessed through a dedicated VNC session.
Hence, if the VM is rebooted during the local configuration, it may be necessary to manually start the graphical interface by logging in the machine and typing `sudo systemctl isolate graphical.target`.


### CrownLabs run-time dependencies

The tools that are required by CrownLabs to work and are automatically installed by the Crownlabs ansible playbooks are:
- **TigerVNC server**: it allows to connect to the VM desktop from a remote machine.
- **NoVNC with websockify server**: it allows the above connection to be established through HTTP/HTTPS, hence allowing to connect from a browser, without any other desktop client applications (e.g., VNC client).
- **Prometheus node exporter**: it exports some run-time information of the VM (e.g., CPU/memory consumption) to the Prometheus monitoring system, running on the Kubernetes cluster, hence facilitating the monitoring of your system.
- **cloud-init**: it enables to customize some running parameters of the VM at boot time.


## Creating a new VM: step-by-step guide

1. Customize the basic information at the beginning of the [setup-crownlabs-vm.sh](setup-crownlabs-vm.sh) script (e.g. username and password, address of the docker registry).

2. Execute `./setup-crownlabs-vm.sh <vm-name> create [desktop|server] <ubuntu-version> (--install-guest-additions)` to automatically create a new VM, install the Ubuntu OS and (optionally) the Virtualbox Guest Additions. You need to choose between *Ubuntu Desktop* (i.e. `xubuntu`) and *Ubuntu Server*, as well as specify the Ubuntu version to be installed (e.g. 20.04). In case you select the _server_ version, by default no desktop environment will be installed.

   **NOTE-1**: *Guest additions* are useful if you want to use the created VM on Virtuabox, on your desktop computer. Vice versa, these are useless when the VM runs in Crownlabs. Therefore, we suggest to install the guest additions when you want to test your VM on your computer, while omitting this software package when creating the final version of your VM and uploading in CrownLabs.

   **NOTE-2**: the current VM creation process requires the VM to boot and perform an _unattended OS install_ (i.e., the user is not required to interact with the new VM). However, this requires the host machine to have a GUI running, as this process is done in a Virtualbox window that is automatically started by the `./setup-crownlabs-vm.sh` script.

   **NOTE-3**: in some cases, we experienced a bug in Virtualbox that prevents the GUI to start. This can be fixed by manually changing the graphical card in the Virtuabox VM settings from `VboxVGA` to `VMSVGA`.

3. Once the installation terminates and the OS completes the reboot, issue `./setup-crownlabs-vm.sh <vm-name> configure <ansible-playbook.yml> (--vbox-only)`, where the _playbook_ contains the instructions to configure and install the additional software packages you need to run in the VM.
The script connects to the VM via SSH and runs the specified ansible playbook.
For more detailed instructions about creating and customizing Ansible playbook, look at the [dedicated subsection](#ansible).

   **NOTE**: if the command is executed with no additional parameters, the scripts assumes that the VM is going to be used in CrownLabs and proceeds with the installation of the tools required for its operations. The `--vbox-only` flag tells the script to opt out this configuration and prepares vanilla VMs meant to be used directly in VirtualBox, as well as exported in the *.ova* format.

4. In some cases, some manual configuration steps are required (e.g., in case the VM is going to be prepared to be executed locally on Virtualbox). In this case, log-in the VM and complete the remaining manual tasks as explained in the [README](ansible/roles/ubuntu-post/files/README) file that is copied to the Desktop.

5. Shutdown the VM and issue `./setup-crownlabs-vm.sh <vm-name> export [ova|crownlabs]` to export the VM to an `.ova` file or make it available in your *CrownLabs* cluster. In case you export the VM to Crownlabs, the script will upload the (shrinked) VM image on the Docker registry used in your Crownlabs cluster.
This requires the administrative username/password of your Docker registry (the script will ask for this information interactively) and it may take some time to upload the image, particulatly in case your local machine does not have a fast connection to the Internet.

6. Optionally, delete the VM with `./setup-crownlabs-vm.sh <vm-name> delete`.

Finally, refer to the [Guide to create courses, labs and student/professor accounts](../courses/) in case you need to create or update the above information in your CrownLabs cluster.
