FROM ubuntu:20.04

ARG USER=crownlabs
ARG UID=1010

### ENV DEFAULTS
ENV DISPLAY=:0 \
    USER=${USER} \
    DEBIAN_FRONTEND=noninteractive

### Install required softwares & cleanup
RUN apt-get update && \
    apt-get install -y \
        fluxbox \
        tigervnc-standalone-server && \
    apt-get --purge autoremove -y && \
    apt-get autoclean -y

EXPOSE 5900

### Prepare usermode
RUN useradd -ms /bin/bash -u ${UID} ${USER}
ENV HOME=/home/$USER

# Copy fluxbox config & vnc session tweaks
COPY --chown=${USER}:${USER} fluxbox $HOME/.fluxbox
COPY Xvnc-session /etc/X11/Xvnc-session

USER ${USER}

CMD vncserver $DISPLAY -SecurityTypes None -noxstartup &&\
    xhost +local: &&\
    fluxbox
