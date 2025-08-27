FROM golang:1-alpine AS builder
WORKDIR /tmp/builder

RUN apk add --no-cache build-base libpcap-dev linux-headers

COPY go.mod ./go.mod
COPY go.sum ./go.sum
RUN  go mod download

ARG COMPONENT
RUN test -n "$COMPONENT" || ( echo "The COMPONENT argument is unset. Aborting" && false )

COPY . ./
RUN CGO_ENABLED=1 GOOS=linux GOARCH=$(go env GOARCH) go build -ldflags="-s -w -linkmode=external -extldflags=-static" ./cmd/$COMPONENT


FROM alpine:3.19

RUN apk update && \
    apk add --no-cache ca-certificates libpcap tcpdump && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ARG COMPONENT
COPY --from=builder /tmp/builder/$COMPONENT /usr/bin/$COMPONENT
RUN ln -s /usr/bin/$COMPONENT /usr/bin/crownlabs-component

ENTRYPOINT [ "/usr/bin/crownlabs-component" ]
