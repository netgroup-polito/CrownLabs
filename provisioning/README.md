# Provisioning

This folder contains a set of scripts to automate as much as possible the provisioning of CrownLabs **courses** and **virtual machines**.
For more detailed information, please refer to the respective *README* files.

These scripts do not represent yet an optimal solution. Indeed, the operations cannot be performed directly from the website GUI. Additionally, some intervention from the administrators is still required (i.e. admin credentials are required to interact with *kubernetes*, the *identity provider* and the *docker registry*). However, the scripts are made available until those features will be integrated in the GUI for a fully automated workflow.

## Courses

The [courses](courses/) folder contains the resources made available to create *CrownLabs* **courses**, including the different **laboratories** and **students and professors accounts**.

## Virtual Machines
The [virtual-machines](virtual-machines/) folder contains the resources made available to automate the creation, configuration and export of virtual machines to be used in *CrownLabs*. Additionally, it provides a set of scripts to configure and export already existing VMs, as well as to interact with the docker registry.