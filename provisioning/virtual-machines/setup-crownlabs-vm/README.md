# Setup CrownLabs VM

This folder contains different scripts to automatize as much as possible the
creation and configuration of the VMs used for the *CrownLabs*.

All the functionalities can be triggered executing the `setup-crownlabs-vms.sh`
script, which can be customized to change some installation parameters (e.g.
Ubuntu distribution and version, VM name, credentials, locale information).

## Requirements:
- A Linux host (or Ubuntu for Windows)
- Virtualbox
- ansible (>= 2.8)
- curl
- ssh
- sshpass

## Step-by-step guide

0. Customize the basic information at the beginning of the [setup-crownlabs-vm.sh](setup-crownlabs-vm.sh)
   script (e.g. username and password and docker registry configuration).
1. Execute `./setup-crownlabs-vm.sh <vm-name> create (--no-guest-additions)` to automatically create
   a new VM, install the Ubuntu OS and (optionally) the Virtual Box Guest Additions.
2. Once the installation terminates and the OS completes the reboot, issue
   `./setup-crownlabs-vm.sh <vm-name> configure <ansible-playbook.yml>`. The script
   connects to the VM via SSH and runs the specified ansible playbook to
   perform a series of configuration tasks: for instance,
   [xubuntu-netlab-playbook](ansible/xubuntu-netlab-playbook.yml) installs
   `Wireshark`, `Docker`, `polycube` and `GNS3`.

   Some predefined playbook are already provided in the [ansible](ansible) folder.
   Specifically, those named `*-crownlabs-playbook.yaml` are an extended version to
   additionally install the software required to enable the access from CrownLabs.
   In case you want to *add/remove* new tasks, you can simply create a new
   ansible script, and optionally new ansible roles, as in
   [xubuntu-netlab-playbook.yml](ansible/xubuntu-netlab-playbook.yml). In particular,
   the first task (`xubuntu-pre`) is rather generic and aims at cleaning up the system
   and removing unnecessary packages, shrinking the size of the VM disk to a
   smaller size. However, in case you want to export the VM in *.ova* format, disk
   shrinking can take place only after the manual executed the corresponding
   post-install task listed in [README](ansible/xubuntu-post/files/README).
3. Log-in the VM and complete the remaining manual tasks as explained in the
   [README](ansible/xubuntu-post/files/README) file (copied to the Desktop).
4. Shutdown the VM and issue `./setup-crownlabs-vm.sh <vm-name> export [ova|crownlabs]`
   to export the VM to an `.ova` file or make it available for the *CrownLabs*.
5. Optionally, delete the VM with `./setup-crownlabs-vm.sh <vm-name> delete`.
