FROM alpine:3.14 AS builder

WORKDIR /build
RUN mkdir -p /BUILD
ENV HTML_DATA=/build

RUN apk add bash

COPY . .
RUN bash ./prepare-novnc.sh


FROM nginx:1.19.5-alpine

ENV HTML_DATA=/usr/share/nginx/html

# Copy config template and entrypoint
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY entrypoint.sh /entrypoint.sh
COPY --from=builder /build ${HTML_DATA}

# Set default build arguments
ARG HTTP_PORT=8080

ARG UID=1010
ARG USER=crownlabs

# Create default env variables
ENV HTTP_PORT=${HTTP_PORT}\
    HIDE_NOVNC_BAR=false\
    USER=${USER}

# Add non-root user and ensure authorizations
RUN adduser -D -H -u ${UID} -s /sbin/nologin ${USER} && \
    chown -R ${USER}:${USER} $HTML_DATA && \
    chown -R ${USER}:${USER} /var/cache/nginx && \
    chown -R ${USER}:${USER} /var/log/nginx && \
    chown -R ${USER}:${USER} /etc/nginx/nginx.conf && \
    chmod a+x /entrypoint.sh

EXPOSE ${HTTP_PORT}

# Enforce non-root user
USER ${USER}

ENTRYPOINT [ "/entrypoint.sh" ]
