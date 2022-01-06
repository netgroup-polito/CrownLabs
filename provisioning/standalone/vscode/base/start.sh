#!/usr/bin/env bash
shopt -s dotglob

if [ ! -f "/config/workspace/.vscode/.startup" ]; then
    mkdir -p /config/workspace/.vscode/
    rm -r /config/workspace
    cp -R /example_project/* /config/workspace
    echo "[Persistent Only Feature]" > /config/workspace/.vscode/.startup
    echo "If your CrownLabs instance is persistent, delete this file if you want to reset the workspace on next startup." >> /config/workspace/.vscode/.startup
fi

for ARGUMENT in "$@"; do
    if [ "$ARGUMENT" == "--disable-marketplace" ] ; then
        export EXTENSIONS_GALLERY='{"serviceUrl": ""}'
    fi
done

if [ "${CODETOGETHER_ENABLED}" == "true" ]; then
    CODETOGETHER_ENABLED_ARG="--enable-proposed-api=genuitecllc.codetogether"
else
    CODETOGETHER_ENABLED_ARG=""
fi

exec \
code-server \
--disable-update-check \
--auth none \
"${CODETOGETHER_ENABLED_ARG}" \
--bind-addr 0.0.0.0:"${CROWNLABS_LISTEN_PORT}" \
--user-data-dir /config/data \
--extensions-dir /config/extensions \
--disable-telemetry \
--new-window \
/config/workspace
