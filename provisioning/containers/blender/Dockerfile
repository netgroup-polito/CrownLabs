# Original Dockerfile from https://github.com/nytimes/rd-blender-docker/blob/master/dist/2.92-cpu-ubuntu18.04/Dockerfile
# Includes edits to update to the latest LTS version

FROM ubuntu:22.04

# Enviorment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

# Install dependencies
RUN apt-get update && apt-get install -y \
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
    libgl1-mesa-glx \
    xz-utils \
    tzdata \
    libxkbcommon-tools \
  && apt-get clean -y \
  && rm -rf /var/lib/apt/lists/*.*

ARG BLENDER_VERSION_MAJOR="3.3"
ARG BLENDER_VERSION_MINOR="6"
ENV BLENDER_VER="${BLENDER_VERSION_MAJOR}.${BLENDER_VERSION_MINOR}-linux-x64"
ENV BLENDER_PATH="/bin/${BLENDER_VERSION_MAJOR}"
ENV PATH "$PATH:${BLENDER_PATH}/python/bin/"
ENV TZ="Europe/Rome"

# Download and install Blender
RUN curl https://ftp.nluug.nl/pub/graphics/blender/release/Blender${BLENDER_VERSION_MAJOR}/blender-${BLENDER_VER}.tar.xz \
  | tar -Jxv --strip-components=1 -C /bin

# Define user and user id default arguments
ARG USER=crownlabs
ARG UID=1010

# Define basic default enviroment variables
ENV DISPLAY=:0 \
  USER=${USER} \
  HOME=/mydrive

# Create new user and set a set a valid shell for them
RUN mkdir -p $HOME && useradd -ms /bin/bash -u ${UID} $USER

# Set permissions on user home
RUN chown -R $USER:$USER $HOME

# Copy the startup script for resources limiting
COPY autolimits.py ${BLENDER_PATH}/scripts/startup/crownlabs_autolimits.py

# Set the user to use
USER $USER

WORKDIR $HOME

CMD blender
