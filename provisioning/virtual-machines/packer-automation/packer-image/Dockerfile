FROM ubuntu:20.04
COPY packer_builder /packer_builder

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
    ansible && \
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

COPY script.sh /usr/src/script.sh

ENTRYPOINT [ "/usr/src/script.sh" ]