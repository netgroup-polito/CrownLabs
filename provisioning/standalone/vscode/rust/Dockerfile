# syntax = edrevo/dockerfile-plus

INCLUDE+ ./base/Dockerfile

ENV SUDO_FORCE_REMOVE yes
ENV RSPATH=${VSCODE_SRV_DIR}/.cargo/bin

COPY ./rust/settings.json ${VSCODE_SRV_DIR}/data/User/settings.json

# Install required packages and remove apt and useless/dangerous packages
RUN apt-get update && \
    apt-get install -y gcc curl && \
    apt-get clean && \
    apt-get remove --autoremove --purge -y apt sudo --allow-remove-essential

# Install extensions
RUN code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions --install-extension formulahendry.code-runner && \
    code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions --install-extension rust-lang.rust-analyzer && \
    code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions --install-extension vadimcn.vscode-lldb

# Setup permissions
RUN chown -R ${USER}:${USER} ${VSCODE_SRV_DIR}

# Install Rust and create example project
USER ${USER}
ENV HOME=${VSCODE_SRV_DIR}
RUN curl -sSf https://sh.rustup.rs | /bin/bash -s -- -y -v && \
    cd $HOME/workspace && \
    ${RSPATH}/cargo new example_project

ENTRYPOINT [ "/start.sh" ]
