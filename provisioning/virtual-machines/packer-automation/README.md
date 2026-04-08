# Guide for creating and uploading Packer VMs in CrownLabs

## Intro to Packer 
[Packer](https://developer.hashicorp.com/packer) is a tool by Hashicorp for creating and configuring machine images in a very simple way.
The `packer build` command takes a template and creates the artifact.
It is based on three principal blocks:
- **Builders**: they are blocks resposible for creating machines and generating images from them. Packer makes available different builders, for example for VirtualBox, VMware, etc. In our case we use the QEMU Builder to create the virtual machine images.
- **Provisioners**: they are responsible for installing and configuring the machine images after booting. For instance, thanks to provisioners we can install packages and download application code. We use this block to run ansible playbooks and to clean cloud-init.
- **Post-Processors**: they run after the image is built. With this block we can make some elaboration on the artifact created. We use this block for reducing the size of the image created through `virt-sparsify`.

## Docker Images
Two images are involved in this mechanism:
- **[packer-image](packer-image/Dockerfile)**: as we can see in the Dockerfile, this image contains all it is necessary for creating a cloud image with Packer. Before Packer is launched, the ansible playbooks are downloaded from the git repository indicated in the environment variables. The environment variables are:
  - `PACKER_LOG`, it can be `0` or `1` and it indicates the log level of packer (default: 1);
  - `ISO_URL`, it is the URL from which the ISO image is downloaded, Ubuntu ISO image can be found [here](https://cloud-images.ubuntu.com/);
  - `ISO_CHECKSUM`, it is the checksum of the ISO image;
  - `INSTALL_DESKTOP_ENVIRONMENT`, it can be `true` or `false` and it indicates if the image that we want to create should have the desktop environment (default: false);
  - `GIT_ANSIBLE_URL`, it indicates from which git repository we have to download the ansible playbooks (default: https://github.com/netgroup-polito/CrownLabs.git);
  - `GIT_ANSIBLE_BRANCH`, it indicates from which branch of the git repository we have to download the ansible playbooks (default: master);
  - `ANSIBLE_PLAYBOOK`, it indicates which ansible playbooks we want to run.
- **[docker-image](docker-image/Dockerfile)**: as we can see in the Dockerfile, this image contains all it is necessary for login in the harbor registry, with the credential provided in the environment variables, and for creating and pushing the image in this registry. The environment variables are:
  - `IMAGE_NAME`, it indicates the name of the image that will be pushed in Harbor;
  - `USERNAME`, it indicates the username used for Harbor;
  - `PASSWORD`, it indicates the password used for Harbor.

## Creating and uploading a custom image
To create and upload a custom Crownlabs image, with Packer, it is possible to leverage the job template provided [here](deploy/job.yaml), appropriately configuring the environment variable values.
It essentially combines the Docker images mentioned in the previous section. In this case the packer-image is used as an `initContainer` and the docker-image is used as a standard container. In this way we can be sure that the uploading of the image in the Harbor registry happens after the creation of the image itself.
The job executes completely unattended. Once it terminates successfully, the resulting image will be available in the configured registry.
