#!/bin/sh

# This script lists all the images stored in a private repository

# Print on screen all the executed commands
set -x

# Registry, e.g., 'registry.crownlabs.polito.it'
registry='<registry host>'

# Username and password; not needed if you are already logged in
auth='-u <username>:<password>'

# Print on screen all the executed commands
set -x

curl "$auth" -ssL -k "https://${registry}/v2/_catalog"
