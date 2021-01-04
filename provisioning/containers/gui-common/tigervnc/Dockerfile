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
    apt-get clean -y && \
    rm -rf /var/lib/apt/lists/*.*

EXPOSE 5900

### Prepare usermode
RUN useradd -ms /bin/bash -u ${UID} ${USER}
ENV HOME=/home/$USER

# Copy fluxbox config (ensuring auth) & vnc session tweaks
COPY --chown=${USER}:${USER} fluxbox $HOME/.fluxbox
COPY Xvnc-session /etc/X11/Xvnc-session

# Enforce non-root user
USER ${USER}

# - Start X server
# - Authorize any connection from current host (pod)
# - Start window manager (fluxbox)
CMD vncserver $DISPLAY -SecurityTypes None -noxstartup &&\
    xhost +local: &&\
    fluxbox
