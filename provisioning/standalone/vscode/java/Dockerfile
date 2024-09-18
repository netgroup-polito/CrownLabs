# syntax = edrevo/dockerfile-plus

INCLUDE+ ./base/Dockerfile


# Setup permissions

# Install required packages
RUN apt-get update && apt-get install -y openjdk-17-jdk openjdk-17-source maven --no-install-recommends

# Install VS code extension
RUN code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions --install-extension vscjava.vscode-java-pack

# Remove apt and useless/dangerous packages
RUN apt-get clean && \
    apt-get remove --autoremove --purge -y apt sudo --allow-remove-essential

COPY ./java/settings.json ${VSCODE_SRV_DIR}/data/User/settings.json

RUN chown -R ${USER}:${USER} ${VSCODE_SRV_DIR}

USER ${USER}

WORKDIR ${VSCODE_SRV_DIR}

# Forces Maven to preload dependencies
RUN --mount=type=bind,source=./java/triggerproject/,target=./triggerproject cd ${VSCODE_SRV_DIR}/triggerproject && mvn test && rm -rf /tmp/target

RUN git config --global credential.helper 'cache --timeout=10800'&&\
    git config --global user.email student@crownlabs.polito.it &&\
    git config --global user.name crownlabs
    
ENTRYPOINT [ "/start.sh" ]
