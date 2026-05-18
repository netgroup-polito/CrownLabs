FROM ubuntu:24.04

ARG USER=crownlabs
ARG UID=1010

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
  apt-get install -y --no-install-recommends unar curl zip file && \
  apt-get clean && \
  rm -rf /var/lib/apt /var/{cache,log} && \
  useradd -m -u "${UID}" -s /bin/sh "${USER}"

COPY --chmod=755 scripts/* /usr/bin/

USER ${USER}
