#!/bin/bash

set -e

cd /packer_builder

if [[ ! -z "$GIT_ANSIBLE_URL" ]]; then
  git clone --no-checkout "$GIT_ANSIBLE_URL" --filter=tree:0 --depth=1 --branch="$GIT_ANSIBLE_BRANCH" repo
  pushd repo
  git sparse-checkout init --cone
  git sparse-checkout set "$GIT_ANSIBLE_PATH"
  git checkout "$GIT_ANSIBLE_BRANCH"
  popd
  mv repo/"$GIT_ANSIBLE_PATH" ansible/

  /usr/bin/packer build -force "builder.pkr.hcl"
fi

if [[ ! -z "$TARGET_REGISTRY" ]]; then
  LABEL="$TARGET_REGISTRY/$REGISTRY_PREFIX/$IMAGE_NAME:$IMAGE_TAG"

  buildah login -u "$REGISTRY_USER" -p "$REGISTRY_PASS" $TARGET_REGISTRY
  buildah build -t "$LABEL" --build-arg IMAGE_NAME=$IMAGE_NAME .
  buildah push "$LABEL"
fi
