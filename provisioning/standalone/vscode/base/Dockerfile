FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive
ENV CODESERVER_VERSION=4.98.2
ENV SERVICE_URL=https://open-vsx.org/vscode/gallery
ENV ITEM_URL=https://open-vsx.org/vscode/item

# Install code-server and required packages
RUN apt-get update &&\
    apt-get install -y curl git &&\ 
    curl -fsSL https://code-server.dev/install.sh | sh -s -- --version=${CODESERVER_VERSION} &&\
    apt-get purge -y curl &&\
    apt-get clean

# Define user and user id default arguments
ARG USER=crownlabs
ARG UID=1010
ARG VSCODE_SRV_DIR=/vscode
ENV VSCODE_SRV_DIR=${VSCODE_SRV_DIR}

# Create new user, setup home folder, .bashrc, .profile and .bash_aliases
RUN useradd -ms /bin/bash -u ${UID} $USER && \
    usermod -d ${VSCODE_SRV_DIR} $USER && \
    mkdir -p ${VSCODE_SRV_DIR}/extensions && \
    mkdir -p ${VSCODE_SRV_DIR}/data && \
    mkdir -p ${VSCODE_SRV_DIR}/workspace && \
    cp /root/.bashrc ${VSCODE_SRV_DIR}/.bashrc && \
    cp /root/.profile ${VSCODE_SRV_DIR}/.profile && \
    echo 'alias code=code-server' >> ${VSCODE_SRV_DIR}/.bashrc && \
    echo 'export PS1="\[\e]0;\u@\h: \w\a\]\[\033[0;00m\][\A]\[\033[00;00m\]\[\033[01;34m\]\u👑\[\033[00m\]:\[\033[01;34m\]\w\[\e[91m\]\[\033[00m\]$ "' >> ${VSCODE_SRV_DIR}/.bashrc

COPY ./base/start.sh start.sh
RUN chmod 755 start.sh
