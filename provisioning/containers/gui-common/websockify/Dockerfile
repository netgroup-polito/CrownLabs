
FROM golang:1.17 as builder
WORKDIR /tmp/builder

COPY go.mod ./go.mod
COPY go.sum ./go.sum
RUN  go mod download

COPY . ./

ENV HTML_DATA=novnc
RUN bash ./prepare-novnc.sh

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$(go env GOARCH) go build -ldflags="-s -w" .

FROM alpine:3.14

RUN apk update && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ARG UID=1010
ARG USER=crownlabs

RUN adduser -D -H -u ${UID} -s /sbin/nologin ${USER}

# Copy compiled binary from builder container
COPY --from=builder /tmp/builder/websockify /usr/bin/websockify

# Configure default VNC endpoint
ENV USER=${USER}

EXPOSE ${WS_PORT}

# Enforce non-root user
USER ${USER}

# Start websockify
ENTRYPOINT ["/usr/bin/websockify"]
