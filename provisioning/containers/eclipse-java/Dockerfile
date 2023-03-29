FROM ubuntu:22.04

# Prevent apt from prompting the user for input during install/configuration
ENV DEBIAN_FRONTEND=noninteractive

# Install required libraries to render PyCharm
RUN apt-get update && \
  apt-get install -y --no-install-recommends libxrender1 libxtst6 libxi6 libfontconfig1 wget  libswt-gtk-4-java ca-certificates libwebkit2gtk-4.0-37 xdg-utils && \
  apt-get clean -y && \
  rm -rf /var/lib/apt/lists/*.* && \
  update-mime-database /usr/share/mime/

ARG USER=crownlabs
ARG UID=1010

ENV DISPLAY=:0 \
  USER=${USER} \
  HOME=/home/$USER \
  TZ=Europe/Rome

RUN useradd -ms /bin/bash -u ${UID} $USER && chown -R $USER:$USER $HOME

ARG ECLIPSE_URL='https://www.eclipse.org/downloads/download.php?file=/technology/epp/downloads/release/2023-03/R/eclipse-java-2023-03-R-linux-gtk-x86_64.tar.gz&r=1'

RUN wget ${ECLIPSE_URL} -qO- /tmp/eclipse.tar.gz | tar xz -C /usr/lib && \
    ln -s /usr/lib/eclipse/eclipse /usr/bin/eclipse

USER $USER

ENTRYPOINT ["/usr/bin/eclipse"]
