FROM alpine:3.13.4

# Install the qemu-img useful to convert the image
RUN apk add --update --no-cache qemu-img

# Copy the entrypoint script
COPY exporter.sh /

# Run the entrypoint which converts the image and creates the dockerfile
CMD ["/exporter.sh"]
