#!/bin/bash

# "svn" (instead of git) is used to download only the "ansible" folder
if [ "$GIT_ANSIBLE_BRANCH" = master ]; then BRANCH=trunk; else BRANCH="branches/$GIT_ANSIBLE_BRANCH"; fi
svn export --force "$GIT_ANSIBLE_URL/$BRANCH/provisioning/virtual-machines/ansible"
packer build \
-force \
-var "ISO_URL=$ISO_URL" \
-var "ISO_CHECKSUM=$ISO_CHECKSUM" \
-var "ANSIBLE_PLAYBOOK=$ANSIBLE_PLAYBOOK" \
-var "INSTALL_DESKTOP_ENVIRONMENT=$INSTALL_DESKTOP_ENVIRONMENT" \
-var "MEMORY=$MEMORY" \
-var "DISK_SIZE=$DISK_SIZE" \
"builder.json"
