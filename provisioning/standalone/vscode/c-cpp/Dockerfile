# syntax = edrevo/dockerfile-plus

# This stage can be used to compile locally the extension, and it can be useful with those releases which don't include a .vsix file
#FROM node:16 as builder

#ENV CPPTOOLS_VERSION=1.8.2
#RUN git clone https://github.com/microsoft/vscode-cpptools.git
#WORKDIR /vscode-cpptools/Extension
#RUN ls
#RUN git checkout tags/${CPPTOOLS_VERSION}
#RUN npm install -g vsce
#RUN touch LICENSE.txt
#RUN vsce package --yarn

INCLUDE+ ./base/Dockerfile

ENV SUDO_FORCE_REMOVE yes
ENV CPPTOOLS_VERSION=1.7.1

COPY ./c-cpp/project/main.c /example_project/main.c
COPY ./c-cpp/project/main.cpp /example_project/main.cpp
COPY ./c-cpp/project/vscode /example_project/.vscode
COPY ./c-cpp/settings.json /config/data/User/settings.json

RUN apt-get update && apt-get install -y build-essential cmake gdb && \
    apt-get clean && \
    apt-get remove --autoremove --purge -y sudo apt --allow-remove-essential

#This line can be used to retrieve the locally compiled package from the building stage
#COPY --from=builder /vscode-cpptools/Extension/cpptools-${CPPTOOLS_VERSION}-main.vsix ./cpptools-linux.vsix
ADD "https://github.com/microsoft/vscode-cpptools/releases/download/${CPPTOOLS_VERSION}/cpptools-linux.vsix" "./cpptools-linux.vsix"

RUN chown -R crownlabs:crownlabs /config && \
    chown -R crownlabs:crownlabs /example_project && \
    chown crownlabs:crownlabs cpptools-linux.vsix && \
    chmod 755 cpptools-linux.vsix

USER crownlabs
RUN code-server --extensions-dir /config/extensions --install-extension cpptools-linux.vsix && \
    code-server --extensions-dir /config/extensions --install-extension formulahendry.code-runner
ENTRYPOINT [ "/start.sh" ]
