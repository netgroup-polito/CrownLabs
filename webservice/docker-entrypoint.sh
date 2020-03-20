#!/bin/sh -eu
./generate_config_js.sh >/usr/share/nginx/html/config.js

sed -i".old" 's/<body>/<body><script type="text\/javascript" src="\/config.js"><\/script>/' /usr/share/nginx/html/index.html

nginx -g "daemon off;"
