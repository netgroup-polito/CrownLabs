# Build the manager binary
FROM golang:1.15 as builder
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
COPY ./ /go/src/github.com/netgroup-polito/CrownLabs/operators/
WORKDIR /go/src/github.com/netgroup-polito/CrownLabs/operators/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o controller ./cmd/instance-operator/main.go
RUN cp controller /usr/bin/controller

FROM busybox
COPY --from=builder /usr/bin/controller /usr/bin/controller
USER 20000:20000

# KubeVirt registers two versions of its API, which is not supported by
# the kubernetes client. This environment variable is required to force
# KubeVirt register only a specific version of the API
ENV KUBEVIRT_CLIENT_GO_SCHEME_REGISTRATION_VERSION v1alpha3

ENTRYPOINT [ "/usr/bin/controller" ]
