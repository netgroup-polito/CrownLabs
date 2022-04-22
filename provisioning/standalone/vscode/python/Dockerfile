# syntax = edrevo/dockerfile-plus

INCLUDE+ ./base/Dockerfile

ENV SUDO_FORCE_REMOVE yes

COPY ./python/project/main.py /example_project/main.py
COPY ./python/project/vscode /example_project/.vscode
COPY ./python/settings.json /config/data/User/settings.json

# Install required packages and remove apt and useless/dangerous packages
RUN apt-get update && \
    apt-get install -y python3 python3-pip python-is-python3 && \
    pip3 install autopep8 pylint && \
    apt-get clean && \
    apt-get remove --autoremove --purge -y apt sudo --allow-remove-essential

# Install extension
RUN code-server --extensions-dir /config/extensions --install-extension ms-python.python

# Setup permissions
RUN chown -R ${USER}:${USER} /config && \
    chown -R ${USER}:${USER} /example_project

USER ${USER}
ENTRYPOINT [ "/start.sh" ]