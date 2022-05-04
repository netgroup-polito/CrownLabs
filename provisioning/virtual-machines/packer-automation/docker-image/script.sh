#!/bin/bash
service docker start
rm -f Dockerfile
echo 'FROM scratch' >>Dockerfile
echo 'ADD output.qcow2 /disk/' >>Dockerfile
docker login -u "$USERNAME" -p "$PASSWORD" harbor.crownlabs.polito.it
docker build -t "$IMAGE_NAME" .
docker push "$IMAGE_NAME"
