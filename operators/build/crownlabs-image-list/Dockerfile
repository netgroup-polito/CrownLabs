# Build stage
FROM golang:1.24.2 as builder

WORKDIR /workspace

# Copy go mod and sum files
COPY operators/go.mod operators/go.sum ./

# Copy the application source code
COPY operators/cmd/crownlabs-image-list/ ./cmd/crownlabs-image-list/
COPY operators/pkg/ ./pkg/

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o crownlabs-image-list \
    ./cmd/crownlabs-image-list

# Runtime stage
FROM alpine:3.20

WORKDIR /usr/share/app

# Copy the compiled binary from builder
COPY --from=builder /workspace/crownlabs-image-list ./

ENTRYPOINT [ "./crownlabs-image-list" ]
