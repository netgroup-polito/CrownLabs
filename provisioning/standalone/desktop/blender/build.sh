#!/bin/bash

set -e

apt-get update && apt-get install -y \
    curl \
    libopenexr-dev \
    bzip2 \
    build-essential \
    zlib1g-dev \
    libxmu-dev \
    libxi-dev \
    libxxf86vm-dev \
    libfontconfig1 \
    libxrender1 \
    xz-utils \
    tzdata \
    libxkbcommon-tools --no-install-recommends

apt-get clean -y 
rm -rf /var/lib/apt/lists/*.*

# Download and install Blender
curl https://ftp.nluug.nl/pub/graphics/blender/release/Blender${BLENDER_VERSION_MAJOR}/blender-${BLENDER_VER}.tar.xz | tar -Jxv --strip-components=1 -C /bin
