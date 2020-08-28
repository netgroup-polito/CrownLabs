# Setup CrownLabs VMs

This folder contains different scripts that automate the creation and configuration of the VMs used in *CrownLabs*.

All the functionalities can be triggered executing the `setup-crownlabs-vms.sh` script, which can be customized by changing some installation parameters (e.g. Ubuntu distribution and version, VM name, credentials, local information).

## Requirements:
- **Linux Desktop** host: this represents the machine on which this script has to be executed. Since the VM creation step may require some manual configuration inside the VM, we require to start the machine with the graphical interface, hence the need for a Desktop Linux (not server). Currently Ubuntu for Windows does not appear to be working.
- **Virtualbox**: a lightweight virtualizer, used by this script to create and customize the basic environments of the guest operating system.
- **Ansible** (>= 2.8): used to install and customize any additional required packages (e.g., applications required in the VM).
- **curl**: used to download the guest operating system installer (e.g., the ISO image of the guest operating system).
- **ssh**: required by ansible to interact with the VMs and configure the additional software required in the guest environment (e.g., additional applications).
- **sshpass**: required by ansible to (silently) authenticate with the VM when establishing the SSH connection (to configure additional software packages).
- **ssh-keygen**: used to remove any existing SSH associations and prevent the "Host key verification failed" message.
- **virt-sparsify**: used to export the resulting VM images into the format required by CrownLabs, compacting the resulting disk image in order to save space (and decrease the boot time).
- **docker**: used to export the resulting VM images into the format required by CrownLabs

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
The [ansible](setup-crownlabs-vm/ansible) folder includes some playbooks for courses active at Politecnico di Torino; as an example, the [xubuntu-netlab-playbook](ansible/xubuntu-netlab-playbook.yml) targets a computer network course and installs `Wireshark`, `Docker`, `polycube` and `GNS3`.

In case you want to install different software, you can simply create a new ansible playbook and start the `setup-crownlabs-vms.sh` script with the optional Ansible playbook that has to be executed for your VM.
Given the modularity of Ansible, you can define new playbooks that refer to already existing playbooks, hence leveraging the install procedure defined in other playbooks to (partially) configure your VM as well.

Alternatively, you can create a basic VM (i.e. with the `xubuntu-base-crownlabs` playbook) and proceed with the manual installation of custom software.

Playbooks named `*-crownlabs-playbook.yaml` available in the [ansible](setup-crownlabs-vm/ansible) folder are used to install all the tools required by CrownLabs to work (e.g., the software required to enable the remote access to the VM through an HTTPS session), therefore should be included in any other playbook.

In case you want to *add/remove* new tasks, you can simply create a new ansible script, and optionally new ansible roles, as in [xubuntu-netlab-playbook.yml](ansible/xubuntu-netlab-playbook.yml).
In particular, the first task (`xubuntu-pre`) is rather generic and aims at cleaning up the system and removing unnecessary packages, shrinking the size of the VM disk to a smaller size.
However, in case you want to export the VM in *.ova* format, disk shrinking can take place only after the manual executed the corresponding post-install task listed in [README](ansible/xubuntu-post/files/README).

**Warning**: Crownlabs ansible playbooks remove many programs present in a default installation of `xubuntu` and disable different standard services, in order to reduce the size of the final VM image and speedup the boot process.
In particular, they disable the automatic startup of the graphical interface, since the remote desktop is accessed through a dedicated VNC session.
Hence, if the VM is rebooted during the local configuration, it may be necessary to manually start the graphical interface with `sudo systemctl isolate graphical.target`.


### CrownLabs run-time dependencies

The tools that are required by CrownLabs to work and are automatically installed by the Crownlabs ansible playbooks are:
- **TigerVNC server**: it allows to connect to the VM desktop from a remote machine;
- **NoVNC with websockify server**: it allows the above connection to be established through HTTP/HTTPS;
- **Prometheus node exporter**: it exports some run-time information of the VM (e.g., CPU/memory consumption) to the Prometheus monitoring system, running on the Kubernetes cluster;
- **cloud-init**: it enables to customize some running parameters of the VM at boot time.


## Creating a new VM: step-by-step guide

1. Customize the basic information at the beginning of the [setup-crownlabs-vm.sh](setup-crownlabs-vm.sh) script (e.g. username and password and docker registry configuration).

2. Execute `./setup-crownlabs-vm.sh <vm-name> create (--no-guest-additions)` to automatically create a new VM, install the Ubuntu OS and (optionally) the Virtual Box Guest Additions.

   **NOTE-1**: the *guest additions* are useful if you want to use the created VM on Virtuabox, on your desktop computer. Vice versa, these are useless when the VM runs in Crownlabs. Therefore, we suggest to install the guest additions when you want to test your VM on your computer, while omitting this software package when creating the final version of your VM and uploading in CrownLabs.

   **NOTE-2**: the current VM creation process requires the VM to boot and perform an _unattended OS install_ (i.e., the user is not required to interact with the new VM). However, this requires the host machine to have a GUI running, as this process is done in a Virtualbox window that is automatically started by the `./setup-crownlabs-vm.sh` script.

3. Once the installation terminates and the OS completes the reboot, issue `./setup-crownlabs-vm.sh <vm-name> configure <ansible-playbook.yml>`, where the _playbook_ contains the instructions to configure and install the additional software packages you need to run in the VM.
The script connects to the VM via SSH and runs the specified ansible playbook.
For more detailed instructions about creating and customizing Ansible playbook, look at the [dedicated subsection](#ansible).
  
4. Log-in the VM and complete the remaining manual tasks as explained in the [README](ansible/xubuntu-post/files/README) file (copied to the Desktop).

5. Shutdown the VM and issue `./setup-crownlabs-vm.sh <vm-name> export [ova|crownlabs]` to export the VM to an `.ova` file or make it available for the *CrownLabs*.

6. Optionally, delete the VM with `./setup-crownlabs-vm.sh <vm-name> delete`.
