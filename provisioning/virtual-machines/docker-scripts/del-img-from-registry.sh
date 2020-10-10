#!/bin/sh

# This script allows to delete an image from a private repository
# This helps to clean up the space, if needed

# Print on screen all the executed commands
set -x

# Registry, e.g., 'registry.crownlabs.polito.it'
registry='<registry host>'

# Image name, e.g., 'ubuntu/1804'
name='<image name>'

# Username and password; not needed if you are already logged in
auth='-u <username>:<password>'

# Tag of the image, e.g., 'latest'
tag='<tag>'

curl "$auth" -X DELETE -sI -k "https://${registry}/v2/${name}/manifests/$(
  curl "$auth" -sI -k \
    -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
    "https://${registry}/v2/${name}/manifests/${tag}" \
    | tr -d '\r' | sed -En 's/^Docker-Content-Digest: (.*)/\1/pi'
)"
