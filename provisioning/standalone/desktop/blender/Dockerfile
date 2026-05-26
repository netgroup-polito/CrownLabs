FROM crownlabs/standalone-vnc-base:v0.1.0

ARG BLENDER_VERSION_MAJOR="4.5"
ARG BLENDER_VERSION_MINOR="4"
ENV BLENDER_VER="${BLENDER_VERSION_MAJOR}.${BLENDER_VERSION_MINOR}-linux-x64"
ENV BLENDER_PATH="/bin/${BLENDER_VERSION_MAJOR}"
ENV PATH="$PATH:${BLENDER_PATH}/python/bin/"
ENV TZ="Europe/Rome"

USER root

RUN --mount=type=bind,source=./build.sh,target=/build.sh,readonly bash /build.sh

COPY autolimits.py ${BLENDER_PATH}/scripts/startup/crownlabs_autolimits.py
COPY blender.conf /etc/supervisor/conf.d/

USER ${USER}
