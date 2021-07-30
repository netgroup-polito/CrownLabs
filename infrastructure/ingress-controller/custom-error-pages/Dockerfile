FROM golang:1.16 as builder

COPY ./ /tmp/custom-error-pages
WORKDIR /tmp/custom-error-pages
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o custom-error-pages ./server/*.go


FROM alpine:3.14

COPY ./static/templates /templates/
COPY --from=builder /tmp/custom-error-pages/custom-error-pages /

ENTRYPOINT [ "/custom-error-pages" ]
