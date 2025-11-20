# syntax = edrevo/dockerfile-plus

INCLUDE+ ./base/Dockerfile

ENV SUDO_FORCE_REMOVE yes
COPY ./latex/settings.json ${VSCODE_SRV_DIR}/data/User/settings.json
COPY --chown=${USER}:${USER} ./latex/project /example_project/project
COPY --chown=${USER}:${USER} ./latex/project/.vscode /example_project/project/.vscode

# Install required packages and remove apt and useless/dangerous packages
# Installation steps are taken from https://www.tug.org/texlive/quickinstall.html
RUN apt-get update && \
    apt-get install -y curl perl python3 python-is-python3 && \
    cd /tmp && \
    curl -L https://mirror.ctan.org/systems/texlive/tlnet/install-tl-unx.tar.gz | tar -zx && \
    cd install-tl-2* && \
    perl ./install-tl --scheme=small --no-doc-install --no-src-install --no-interaction && \
    apt-get clean && \
    apt-get remove --autoremove --purge -y apt sudo --allow-remove-essential && \
    rm -rf /tmp/install-tl-2*

# Install extensions and setup permissions
RUN code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions --install-extension James-Yu.latex-workshop && \
    chown -R ${USER}:${USER} ${VSCODE_SRV_DIR} && \
    chown -R ${USER}:${USER} /example_project && \
    chown -R ${USER}:${USER} /usr/local/texlive

# Finalizing TeX Live installation
USER ${USER}
ENV HOME=${VSCODE_SRV_DIR}
RUN YEAR=$(ls /usr/local/texlive | head -n 1) && \
    ARCH=$(ls /usr/local/texlive/$YEAR/bin | head -n 1) && \
    TLPATH=/usr/local/texlive/$YEAR/bin/$ARCH && \
    echo "export PATH=$TLPATH:\$PATH" >> $HOME/.bashrc && \
    $TLPATH/tlmgr install texliveonfly


ENTRYPOINT [ "/start.sh" ]
CMD [ "--load-example", "/vscode/workspace/project" ]
