# CrownLabs Ansible Playbooks

This folder contains the playbooks necessary to configure the cluster to host the "Crown Labs".

### Tasks

1. Install a clean copy of the OS on the physical machines in the cluster.
2. Install `ansible-playbook` on your local machine.
3. Copy your SSH public key to the hosts: `ssh-copy-id <user>@<node-ip>`.
4. Configure the hosts with: `ansible-playbook --inventory k8s-cluster-ladispe-inventory.yml k8s-cluster-ladispe-hosts-setup.yml`.
   **Warning**: The first time this playbook is executed, it is necessary to manually provide the *become* password:
   `ansible-playbook --ask-become-pass --inventory ...`
5. Create the VM to host the Kubernetes master and install the OS.
6. Configure the VM with: `ansible-playbook --inventory k8s-cluster-ladispe-inventory.yml k8s-cluster-ladispe-master-vm-setup.yml`.
   **Warning**: The first time this playbook is executed, it is necessary to manually provide the *become* password:
   `ansible-playbook --ask-become-pass --inventory ...`
7. Your system files have been updated now and changes made by Ansible are persistent. It is possible to proceed with the next steps of the installation.