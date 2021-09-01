# Original Dockerfile from https://github.com/nytimes/rd-blender-docker/blob/master/dist/2.92-cpu-ubuntu18.04/Dockerfile
# Includes edits to update to the latest LTS version

# ! With 18.04 there are issues with blender > 2.83.8
FROM ubuntu:20.04

# Enviorment variables
ENV DEBIAN_FRONTEND noninteractive
ENV LC_ALL C.UTF-8
ENV LANG C.UTF-8

# Install dependencies
RUN apt-get update && apt-get install -y \ 
    wget \ 
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
  && apt-get clean -y \
  && rm -rf /var/lib/apt/lists/*.*

ARG BLENDER_VERSION_MAJOR="2.92"
ARG BLENDER_VERSION_MINOR="0"
ENV BLENDER_VER="${BLENDER_VERSION_MAJOR}.${BLENDER_VERSION_MINOR}-linux64"
ENV BLENDER_PATH="/bin/${BLENDER_VERSION_MAJOR}"
ENV PATH "$PATH:${BLENDER_PATH}/python/bin/"
ENV BLENDERPIP="${BLENDER_PATH}/python/bin/pip3"
ENV BLENDERPY="${BLENDER_PATH}/python/bin/python3.7m"
ENV TZ="Europe/Rome"

# Download and install Blender
RUN wget -q https://ftp.nluug.nl/pub/graphics/blender/release/Blender${BLENDER_VERSION_MAJOR}/blender-${BLENDER_VER}.tar.xz \
  && tar -xf blender-${BLENDER_VER}.tar.xz --strip-components=1 -C /bin \ 
  && rm -rf blender-${BLENDER_VER}.tar.xz \ 
  && rm -rf ${BLENDER_PATH}/python/lib/python3.7/site-packages/numpy

# Download the Python source since it is not bundled with Blender
# Also upgrades the the bulndled outdated version of numpy with a modern one
# Must first ensurepip to install Blender pip3 and then new numpy
RUN wget -q https://www.python.org/ftp/python/3.7.10/Python-3.7.10.tgz \ 
  && tar -xzf Python-3.7.10.tgz \ 
  && cp -r Python-3.7.10/Include/* ${BLENDER_PATH}/python/include/python3.7m/ \ 
  && rm -rf Python-3.7.10.tgz \ 
  && rm -rf Python-3.7.10 \
  && ${BLENDERPY} -m ensurepip \
  && ${BLENDERPIP} install --upgrade pip \
  && ${BLENDERPIP} install numpy

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
