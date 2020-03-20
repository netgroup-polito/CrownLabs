#!/bin/sh -eu
./generate_config_js.sh >/usr/share/nginx/html/config.js
nginx -g "daemon off;"
