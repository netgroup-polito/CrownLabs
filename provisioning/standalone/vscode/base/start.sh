#!/usr/bin/env bash
shopt -s dotglob

# Check if in the passed arguments is specified to disable workspace through the option --disable-marketplace
for ARGUMENT in "$@"; do
    shift
    if [ "$ARGUMENT" == "--disable-marketplace" ] ; then
        echo "Disabling marketplace"
        export EXTENSIONS_GALLERY='{"serviceUrl": ""}'
        continue
    elif [ "$ARGUMENT" == "--init-example-project" ] && [ ! -f "/config/workspace/.vscode/.startup" ] ; then
        echo "Copying example project"
        rm -rf /config/workspace/*
        mkdir -p /config/workspace/.vscode/
        cp -R /example_project/* /config/workspace
        echo "[Persistent Only Feature]" > /config/workspace/.vscode/.startup
        echo "If your CrownLabs instance is persistent, delete this file if you want to reset the workspace on next startup." >> /config/workspace/.vscode/.startup
        continue
    fi
    set -- "$@" "$ARGUMENT"
done

echo "Starting vscode-server"

exec \
code-server \
--disable-update-check \
--auth none \
--bind-addr 0.0.0.0:"${CROWNLABS_LISTEN_PORT}" \
--user-data-dir /config/data \
--extensions-dir /config/extensions \
--disable-telemetry \
--new-window \
"$@"
