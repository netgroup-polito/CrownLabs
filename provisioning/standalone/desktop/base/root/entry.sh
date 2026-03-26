#!/bin/bash
shopt -s extglob;

# ensure that the base path does not start or end with a slash, as this would break nginx configuration
BASE_PATH=${CROWNLABS_BASE_PATH##+(/)}
BASE_PATH=${BASE_PATH%%+(/)}

REDIRECT_SNIP=""
if [[ -n "$BASE_PATH" ]]; then
  REDIRECT_SNIP="rewrite ^/$BASE_PATH\$ /$BASE_PATH/ permanent;"
  BASE_PATH="$BASE_PATH/"
fi

export LISTEN_PORT=${CROWNLABS_LISTEN_PORT:-8080}
export BASE_PATH=${BASE_PATH}
export REDIRECT_SNIP=${REDIRECT_SNIP}

# shellcheck disable=SC2016
envsubst '${LISTEN_PORT} ${BASE_PATH} ${REDIRECT_SNIP}' < /etc/nginx/nginx.conf.tpl > /etc/nginx/nginx.conf

rm -rf /tmp/.X*-*

exec "$@"
