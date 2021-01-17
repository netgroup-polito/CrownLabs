# Provisioning courses, users and virtual machines

This folder contains a set of resources supporting the provisioning of CrownLabs environments, both based on **virtual machines** and **containers**.
For more detailed information, please refer to the respective *README* files.

## Virtual Machines
The [virtual-machines](virtual-machines/) folder contains the resources made available to automate the creation, configuration and export of virtual machines to be used in *CrownLabs*. Additionally, it provides a set of scripts to configure and export already existing VMs, as well as to interact with the docker registry.
These scripts do not represent yet an optimal solution. Indeed, the operations still require some manual intervention from the administrators (e.g. admin credentials are required to interact with the *docker registry* to upload the resulting images).
