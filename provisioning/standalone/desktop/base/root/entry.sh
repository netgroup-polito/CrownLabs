#!/bin/bash
shopt -s extglob;

BASE_PATH=${CROWNLABS_BASE_PATH##+(/)}
BASE_PATH=${BASE_PATH%%+(/)}

[[ -n "$BASE_PATH" ]] && BASE_PATH="$BASE_PATH/"

export LISTEN_PORT=${CROWNLABS_LISTEN_PORT:-8080}
export BASE_PATH=${BASE_PATH}

envsubst '${LISTEN_PORT} ${BASE_PATH}' < /etc/nginx/nginx.conf.tpl > /etc/nginx/nginx.conf

rm -rf /tmp/.X*-*

exec "$@"
