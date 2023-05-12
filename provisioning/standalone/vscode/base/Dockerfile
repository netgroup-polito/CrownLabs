FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive
ENV CODESERVER_VERSION=4.12.0
ARG CODETOGETHER_ENABLED_ARG=false
ENV CODETOGETHER_ENABLED=$CODETOGETHER_ENABLED_ARG
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

ENV HOME /home/$USER

# Create new user, setup home folder, .bashrc, .profile and .bash_aliases
RUN useradd -ms /bin/bash -u ${UID} $USER && usermod -d /config $USER && \
    mkdir -p $HOME/{extensions,data,workspace} && \
    cp /root/.bashrc /root/.profile $HOME/ && \
    echo 'alias code=code-server' >> $HOME/.bashrc && \
    echo 'export PS1="👑 \[\e]0;\u@\h: \w\a\]\[\033[0;00m\][\A]\[\033[00;00m\]\[\033[01;32m\] \u\[\033[00m\]:\[\033[01;34m\]\w\[\e[91m\]\[\033[00m\]$ "' >> $HOME/.bashrc

COPY --chmod=755 ./base/start.sh start.sh

# Install codetogether if specified
RUN if [ "${CODETOGETHER_ENABLED}" = "true" ]; then code-server --extensions-dir $HOME/extensions --install-extension genuitecllc.codetogether; fi
# add --enable-proposed-api=genuitecllc.codetogether to args at runtime to enable

ENTRYPOINT [ "/start.sh" ]
CMD ["/config/workspace"]
