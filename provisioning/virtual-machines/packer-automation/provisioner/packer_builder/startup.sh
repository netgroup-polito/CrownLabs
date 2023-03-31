#!/bin/bash

set -e

cd /packer_builder

if [[ ! -z "$GIT_ANSIBLE_URL" ]]; then
  BRANCH="branches/$GIT_ANSIBLE_BRANCH"
  if [ "$GIT_ANSIBLE_BRANCH" = master ]; then
    BRANCH=trunk
  fi

  svn export --force "$GIT_ANSIBLE_URL/$BRANCH/provisioning/virtual-machines/ansible"

  /usr/bin/packer build -force "builder.pkr.hcl"
fi

if [[ ! -z "$TARGET_REGISTRY" ]]; then
  LABEL="$TARGET_REGISTRY/$REGISTRY_PREFIX/$IMAGE_NAME:$IMAGE_TAG"

  buildah login -u "$REGISTRY_USER" -p "$REGISTRY_PASS" $TARGET_REGISTRY
  buildah build -t "$LABEL" --build-arg IMAGE_NAME=$IMAGE_NAME .
  buildah push "$LABEL"
fi
