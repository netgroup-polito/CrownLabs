FROM ubuntu:21.10

ENV DEBIAN_FRONTEND=noninteractive
ENV CODESERVER_VERSION=4.1.0
ARG CODETOGETHER_ENABLED_ARG=false
ENV CODETOGETHER_ENABLED=$CODETOGETHER_ENABLED_ARG
ENV SERVICE_URL=https://open-vsx.org/vscode/gallery
ENV ITEM_URL=https://open-vsx.org/vscode/item

RUN apt-get update &&\
    apt-get install -y curl git &&\ 
    curl -fsSL https://code-server.dev/install.sh | sh -s -- --version=${CODESERVER_VERSION} &&\
    apt-get purge -y curl &&\
    apt-get clean

# Define user and user id default arguments
ARG USER=crownlabs
ARG UID=1010

# Create new user and set a set a valid shell for them
RUN useradd -ms /bin/bash -u ${UID} $USER

RUN usermod -d /config $USER && \
    mkdir /app && \
    mkdir /config && \
    mkdir /defaults

RUN chown $USER:$USER /app && \
    chown $USER:$USER /config && \
    chown $USER:$USER /defaults

COPY ./base/config.sh config.sh
RUN /bin/bash ./config.sh && \
    rm ./config.sh

RUN if [ "${CODETOGETHER_ENABLED}" = "true" ]; then code-server --extensions-dir /config/extensions --install-extension genuitecllc.codetogether; fi

COPY ./base/start.sh start.sh
RUN chmod 777 start.sh
