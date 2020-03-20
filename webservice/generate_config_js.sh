#!/bin/sh -eu
if [ -z "${OIDC_PROVIDER_URL:-}" ]; then
    OIDC_PROVIDER_URL=undefined
fi

if [ -z "${OIDC_CLIENT_ID:-}" ]; then
    OIDC_CLIENT_ID=undefined
fi

if [ -z "${OIDC_CLIENT_SECRET:-}" ]; then
    OIDC_CLIENT_SECRET=undefined
fi

if [ -z "${APISERVER_URL:-}" ]; then
    APISERVER_URL=undefined
fi

if [ -z "${OIDC_REDIRECT_URI:-}" ]; then
    OIDC_REDIRECT_URI=undefined
fi

cat << EOF
window.OIDC_PROVIDER_URL="$OIDC_PROVIDER_URL";
window.OIDC_CLIENT_ID="$OIDC_CLIENT_ID";
window.APISERVER_URL="$APISERVER_URL";
window.OIDC_CLIENT_SECRET="$OIDC_CLIENT_SECRET";
window.OIDC_REDIRECT_URI="$OIDC_REDIRECT_URI";
EOF
