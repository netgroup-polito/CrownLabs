# syntax = edrevo/dockerfile-plus

INCLUDE+ ./base/Dockerfile

ENV SUDO_FORCE_REMOVE yes
ENV RSPATH=${VSCODE_SRV_DIR}/.cargo/bin

COPY ./rust/settings.json ${VSCODE_SRV_DIR}/data/User/settings.json
COPY --chown=${USER}:${USER} ./rust/project /example_project/project
COPY ./rust/config.toml ${RSPATH}/../config.toml

# Install required packages and remove apt and useless/dangerous packages
RUN apt-get update && \
    apt-get install -y curl gcc bash-completion && \
    curl -L -o /tmp/codelldb-linux-x64.vsix https://github.com/vadimcn/codelldb/releases/download/v1.11.4/codelldb-linux-x64.vsix && \
    apt-get clean && \
    apt-get remove --autoremove --purge -y apt sudo --allow-remove-essential

# Install extensions and setup permissions
RUN code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions --install-extension formulahendry.code-runner && \
    code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions --install-extension rust-lang.rust-analyzer && \
    code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions --install-extension /tmp/codelldb-linux-x64.vsix && \
    rm -f /tmp/codelldb-linux-x64.vsix && \
    chown -R ${USER}:${USER} ${VSCODE_SRV_DIR} && \
    chown -R ${USER}:${USER} /example_project

# Install Rust
USER ${USER}
ENV HOME=${VSCODE_SRV_DIR}
RUN curl -sSf https://sh.rustup.rs | /bin/bash -s -- -y -v && \
    ${RSPATH}/rustup component remove rust-docs && \
    ${RSPATH}/rustup completions bash >> $HOME/.bashrc && \
    ${RSPATH}/rustup completions bash cargo >> $HOME/.bashrc

ENTRYPOINT [ "/start.sh" ]
CMD [ "--load-example", "/vscode/workspace/project" ]
