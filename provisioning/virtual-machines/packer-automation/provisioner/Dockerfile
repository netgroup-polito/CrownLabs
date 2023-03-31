FROM ubuntu:22.04

RUN apt-get update && DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get install -y \
    apt-utils \
    software-properties-common \
    gnupg2 \
    qemu-kvm \
    libvirt-daemon-system \
    libvirt-clients \
    bridge-utils \
    libguestfs-tools \
    linux-image-generic \
    xorriso \
    curl \
    subversion \
    lsb-release \
    ansible \
    sshpass \
    buildah \
    && \
    curl -fsSL https://apt.releases.hashicorp.com/gpg | apt-key add - && \
    apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main" && \
    apt-get update && apt-get install packer

WORKDIR /packer_builder

ENV PACKER_LOG=1
ENV ISO_URL=
ENV ISO_CHECKSUM=

ENV INSTALL_DESKTOP_ENVIRONMENT=false
ENV GIT_ANSIBLE_URL=https://github.com/netgroup-polito/CrownLabs.git
ENV GIT_ANSIBLE_BRANCH=master
ENV ANSIBLE_PLAYBOOK=
ENV MEMORY=2048
ENV DISK_SIZE=10G

ENV TARGET_REGISTRY=harbor.crownlabs.polito.it
ENV REGISTRY_PREFIX=crownlabs-containerdisks
ENV REGISTRY_USER=packer
ENV REGISTRY_PASS=password
ENV IMAGE_NAME=""
ENV IMAGE_TAG=""

COPY packer_builder/ /packer_builder/

ENTRYPOINT [ "/bin/bash", "/packer_builder/startup.sh" ]
