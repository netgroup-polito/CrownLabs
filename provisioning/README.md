# Provisioning courses, users and virtual machines

This folder contains a set of resources supporting the provisioning of CrownLabs environments, both based on **virtual machines** and **containers**.
For more detailed information, please refer to the respective *README* files.

## Virtual Machines
The [virtual-machines](virtual-machines/) folder contains the resources made available to automate the creation, configuration and export of virtual machines to be used in *CrownLabs*. Additionally, it provides a set of scripts to configure and export already existing VMs, as well as to interact with the docker registry.
These scripts do not represent yet an optimal solution. Indeed, the operations still require some manual intervention from the administrators (e.g. admin credentials are required to interact with the *docker registry* to upload the resulting images).

## Containers

The [containers](containers/) folder contains the resources made available to create an application, running in a container, which is compatible with *CrownLabs*, e.g., that exports a remotely-accessible GUI via web browser. This folder will actually contain a set of containers running together (with the Kubernetes *sidecar* approach) in order to make the solution easier to maintain, e.g., splitting the creation of the virtual desktop environment from the actual application container.

## Standalone applications

The [standalone](standalone/) folder contains the resources to build standalone applications. Standalone applications are autonomous web services packaged in a container (e.g., Visual Studio Code, JupyterLab, ...) that are exposed over HTTP. Creating standalone applications is really simple and does not require knowledge about Kubernetes and how CrownLabs works. It is only necessary to follow the appropriate guidelines when creating the containers.  
