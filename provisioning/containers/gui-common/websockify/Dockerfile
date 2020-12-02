FROM alpine:3.12.1

# Build Websockify (C version) & cleanup
RUN apk add --no-cache --virtual .build-deps git openssl-dev musl-dev gcc make && \
    git clone https://github.com/novnc/websockify-other.git /wsckfy && \
    cd /wsckfy/c && \
    make && \
    mv /wsckfy/c/websockify /bin/websockify && \
    rm -rf /wsckfy && \
    apk del .build-deps && \
    cd /

ARG UID=1010
ARG USER=crownlabs

RUN adduser -D -H -u ${UID} -s /sbin/nologin ${USER}

# Configure default VNC endpoint
ENV VNC_SERVER=localhost \
    VNC_PORT=5900 \
    WS_PORT=8888 \
    USER=${USER}

EXPOSE ${WS_PORT}

USER ${USER}

# Start websockify
CMD websockify $WS_PORT $VNC_SERVER:$VNC_PORT
