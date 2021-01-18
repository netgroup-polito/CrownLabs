# Multi-stage build for Websockify (C version), cleanup & add non-root user

# Builder container
FROM alpine:3.12.1 AS builder

# Install build tools
RUN apk add git openssl-dev musl-dev gcc make

# Clone repo & compile
RUN git clone https://github.com/novnc/websockify-other.git /wsckfy
RUN cd /wsckfy/c && make



# Final container
FROM alpine:3.12.1

ARG UID=1010
ARG USER=crownlabs

RUN adduser -D -H -u ${UID} -s /sbin/nologin ${USER}

# Copy compiled binary from builder container
COPY --from=builder /wsckfy/c/websockify /bin/websockify

# Configure default VNC endpoint
ENV VNC_SERVER=localhost \
    VNC_PORT=5900 \
    WS_PORT=8888 \
    USER=${USER}

EXPOSE ${WS_PORT}

# Enforce non-root user
USER ${USER}

# Start websockify
CMD websockify $WS_PORT $VNC_SERVER:$VNC_PORT
