#!/bin/sh -eu
if [ -z "${OIDC_PROVIDER_URL:-}" ]; then
    OIDC_PROVIDER_URL=undefined
else
    OIDC_PROVIDER_URL=$(jq -n --arg oidc_provider_url '$OIDC_PROVIDER_URL' '$oidc_provider_url')
fi
if [ -z "${OIDC_CLIENT_ID:-}" ]; then
    OIDC_CLIENT_ID=undefined
else
    OIDC_CLIENT_ID=$(jq -n --arg oidc_client_id '$OIDC_CLIENT_ID' '$oidc_client_id')
if [ -z "${APISERVER_URL:-}" ]; then
    APISERVER_URL=undefined
else
    APISERVER_URL=$(jq -n --arg apiserver_url '$APISERVER_URL' '$apiserver_url')
if [ -z "${OIDC_REDIRECT_URI:-}" ]; then
    OIDC_REDIRECT_URI=undefined
else
    OIDC_REDIRECT_URI=$(jq -n --arg oidc_redirect_uri '$OIDC_REDIRECT_URI' '$oidc_redirect_uri')
fi
 
cat <<EOF
window.OIDC_PROVIDER_URL=$OIDC_PROVIDER_URL;
window.OIDC_CLIENT_ID=$OIDC_CLIENT_ID;
window.APISERVER_URL=$APISERVER_URL;
window.OIDC_REDIRECT_URI=$OIDC_REDIRECT_URI;
EOF
