#!/bin/bash

set -e

# Prepare repos
apt-get update
apt-get install -y curl gnupg2 supervisor nginx fluxbox gettext-base apt-transport-https ca-certificates --no-install-recommends

curl -sSL "https://packagecloud.io/dcommander/turbovnc/gpgkey"      | gpg --dearmor > /etc/apt/trusted.gpg.d/TurboVNC.gpg
curl -sSL "https://packagecloud.io/dcommander/libjpeg-turbo/gpgkey" | gpg --dearmor > /etc/apt/trusted.gpg.d/libjpeg-turbo.gpg
curl -sSL "https://packagecloud.io/dcommander/virtualgl/gpgkey"     | gpg --dearmor > /etc/apt/trusted.gpg.d/VirtualGL.gpg

curl -sSL -o "/etc/apt/sources.list.d/TurboVNC.list"      "https://raw.githubusercontent.com/TurboVNC/repo/main/TurboVNC.list"
curl -sSL -o "/etc/apt/sources.list.d/libjpeg-turbo.list" "https://raw.githubusercontent.com/libjpeg-turbo/repo/main/libjpeg-turbo.list"
curl -sSL -o "/etc/apt/sources.list.d/VirtualGL.list"     "https://raw.githubusercontent.com/VirtualGL/repo/main/VirtualGL.list"

# Install turboVNC & co
apt-get update
apt-get install -y turbovnc virtualgl libjpeg-turbo-official feh --no-install-recommends

# Create user
useradd -ms /bin/bash -u ${UID} ${USER}

# Prepare directories and permissions
mkdir -p /usr/share/novnc /var/log/supervisor /etc/filebrowser /home/${USER}/.vnc
touch /etc/nginx/nginx.conf
chown -R ${USER}:${USER} /home/${USER} /var/log/{supervisor,nginx} /var/run /run /etc/filebrowser /var/lib/nginx /etc/nginx/nginx.conf

# Download noVNC and FileBrowser
curl -sSL https://github.com/novnc/noVNC/archive/refs/tags/v1.6.0.tar.gz | tar xz -C /usr/share/novnc --strip-components=1
curl -sSL https://github.com/filebrowser/filebrowser/releases/download/v2.45.0/linux-amd64-filebrowser.tar.gz | tar xz -C /usr/bin filebrowser

# Clean up
apt-get remove -y curl gnupg2
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/* /tmp/*
