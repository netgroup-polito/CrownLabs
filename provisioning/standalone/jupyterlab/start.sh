#!/usr/bin/env bash

exec jupyter-lab \
        --ServerApp.base_url="${CROWNLABS_BASE_PATH}" \
        --ServerApp.token="" \
        --ServerApp.allow_remote_access=True \
        --ServerApp.allow_origin='*' \
        --ServerApp.default_url="${CROWNLABS_BASE_PATH}" \
        --ServerApp.ip="0.0.0.0" \
        --ServerApp.local_hostnames="${CROWNLABS_DOMAIN}"] \
        --port "${CROWNLABS_LISTEN_PORT}"

