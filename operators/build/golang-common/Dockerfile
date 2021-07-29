FROM golang:1.16 as builder
WORKDIR /tmp/builder

COPY go.mod ./go.mod
COPY go.sum ./go.sum
RUN  go mod download

ARG COMPONENT
RUN test -n "$COMPONENT" || ( echo "The COMPONENT argument is unset. Aborting" && false )

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$(go env GOARCH) go build -ldflags="-s -w" ./cmd/$COMPONENT


FROM alpine:3.14

RUN apk update && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ARG COMPONENT
COPY --from=builder /tmp/builder/$COMPONENT /usr/bin/$COMPONENT
RUN ln -s /usr/bin/$COMPONENT /usr/bin/crownlabs-component

# KubeVirt registers two versions of its API, which is not supported by
# the kubernetes client. This environment variable is required to force
# KubeVirt register only a specific version of the API
ENV KUBEVIRT_CLIENT_GO_SCHEME_REGISTRATION_VERSION v1alpha3

ENTRYPOINT [ "/usr/bin/crownlabs-component" ]
