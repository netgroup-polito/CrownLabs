FROM python:3.8

ENV SHELL=/bin/bash

# Define user and user id default arguments
ARG USER=crownlabs
ARG UID=1010

# Create new user and set a set a valid shell for them
RUN useradd -ms /bin/bash -u ${UID} $USER

# Setup .bashrc and .profile
RUN cp /root/.bashrc /home/${USER}/.bashrc && \
    cp /root/.profile /home/${USER}/.profile && \
    echo 'export PS1="\[\e]0;\u@\h: \w\a\]\[\033[0;00m\][\A]\[\033[00;00m\]\[\033[01;34m\]\u👑\[\033[00m\]:\[\033[01;34m\]\w\[\e[91m\]\[\033[00m\]$ "' >> /home/${USER}/.bashrc

# Install jupyterlab and common used packages
RUN pip install numpy pandas jupyterlab matplotlib

COPY start.sh start.sh
RUN chmod 755 start.sh

# Remove apt and useless/dangerous packages
RUN apt-get clean && \
    apt-get remove --autoremove --purge -y apt wget curl --allow-remove-essential

USER $USER
WORKDIR /home/${USER}
ENTRYPOINT [ "/start.sh" ]