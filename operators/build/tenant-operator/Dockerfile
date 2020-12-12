FROM alpine as certs
RUN apk update && apk add ca-certificates

# Build the manager binary
FROM golang:1.15 as builder
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
COPY ./ /go/src/github.com/netgroup-polito/CrownLabs/operators/
WORKDIR /go/src/github.com/netgroup-polito/CrownLabs/operators/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o controller ./cmd/tenant-operator/main.go
RUN cp controller /usr/bin/controller

FROM busybox
COPY --from=certs /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/bin/controller /usr/bin/controller
USER 20000:20000
ENTRYPOINT [ "/usr/bin/controller" ]
