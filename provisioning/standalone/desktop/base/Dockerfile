FROM ubuntu:24.04

ARG USER=crownlabs
ARG UID=1010

### ENV DEFAULTS
ENV DISPLAY=:0 \
    USER=${USER} \
    DEBIAN_FRONTEND=noninteractive

### Prepare usermode
ENV HOME=/home/$USER \
  TZ=Europe/Rome \
  TVNC_WM=fluxbox-session

RUN --mount=type=bind,source=./build.sh,target=/build.sh,readonly bash /build.sh
    
# Copy VNC and supervisor configurations
COPY root/ /
COPY --chown=${USER}:${USER} fluxbox $HOME/.fluxbox

ENV WORKSPACE_DIR=/workspace

USER ${USER}

WORKDIR $WORKSPACE_DIR

ENTRYPOINT ["/bin/bash", "/entry.sh"]
CMD ["/usr/bin/supervisord", "-n"]
