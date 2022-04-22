# syntax = edrevo/dockerfile-plus

INCLUDE+ ./base/Dockerfile

ENV SUDO_FORCE_REMOVE yes
ENV GOPATH=/config/go

COPY ./golang/project/main.go /example_project/main.go
COPY ./golang/project/go.mod /example_project/go.mod
COPY ./golang/settings.json /config/data/User/settings.json

# Install required packages
RUN apt-get update && \
    apt-get install -y gcc wget

# Download and install golang and required tools
ADD https://go.dev/dl/go1.18.linux-amd64.tar.gz /usr/local/go.tar.gz
ADD https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh install.sh
RUN tar -xzf /usr/local/go.tar.gz -C /usr/local && \
    echo 'export PATH=$PATH:/usr/local/go/bin'>>/root/.bashrc && \
    echo 'export PATH=$PATH:/config/go/bin'>>/root/.bashrc && \
    echo 'export GOPATH=/config/go'>>/root/.bashrc && \
    rm /usr/local/go.tar.gz && \
    /usr/local/go/bin/go install golang.org/x/tools/gopls@latest && \
    /usr/local/go/bin/go install github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest && \ 
    /usr/local/go/bin/go install github.com/ramya-rao-a/go-outline@latest && \
    /usr/local/go/bin/go install github.com/go-delve/delve/cmd/dlv@latest && \
    /usr/local/go/bin/go install honnef.co/go/tools/cmd/staticcheck@latest && \
    /usr/local/go/bin/go env GOPATH && \
    cp /config/go/bin/dlv /config/go/bin/dlv-dap && \
    /bin/bash ./install.sh -b $(/usr/local/go/bin/go env GOPATH)/bin v1.43.0 && \
    rm install.sh

# Remove apt and useless/dangerous packages
RUN apt-get clean && \
    apt-get --autoremove purge -y apt wget sudo --allow-remove-essential

# Install extensions
RUN code-server --extensions-dir /config/extensions --install-extension golang.Go && \
    code-server --extensions-dir /config/extensions --install-extension formulahendry.code-runner

# Setup permissions
RUN chown -R ${USER}:${USER} /config && \
    chown -R ${USER}:${USER} /example_project

USER ${USER}
ENTRYPOINT [ "/start.sh" ]
