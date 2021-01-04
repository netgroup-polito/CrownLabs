#!/bin/sh
set -e

# Finalize config
envsubst "\$WEBSOCKIFY_HOST \$WEBSOCKIFY_PORT \$HTTP_PORT" \
    < /etc/nginx/nginx.conf.template \
    > /etc/nginx/nginx.conf

# Hide noVNC control bar (mostly used for clipboard) if requested
if [ "$HIDE_NOVNC_BAR" = true ]; then
  echo "#noVNC_control_bar_anchor {display:none !important;}" \
    >> "$HTML_DATA"/app/styles/base.css
fi

# Start nginx (not in daemon mode for Docker use)
nginx -g "daemon off;" "$@"
