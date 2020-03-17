#!/bin/sh

if [ $# -ne 1 ]; then
    echo "Usage: convert-vm.sh <vdi_file>"
    exit 1
fi

DIR=$(dirname $1)

# Install qemu-utils needed by this script
sudo apt-get install -y qemu-utils

# Convert the image
mkdir -p $DIR/docker_output
qemu-img convert -f vdi -O qcow2 $1 docker-output/vm.qcow2

# Create dockerfile for the build
tee $DIR/docker-output/Dockerfile > /dev/null <<EOT
FROM scratch
ADD vm.qcow2 /disk/
EOT
