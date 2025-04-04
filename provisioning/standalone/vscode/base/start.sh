#!/usr/bin/env bash
shopt -s dotglob

TARGETDIR="$VSCODE_SRV_DIR/workspace"

# Check if in the passed arguments is specified to disable workspace through the option --disable-marketplace
for ARGUMENT in "$@"; do
    if [ "$ARGUMENT" == "--disable-marketplace" ] ; then
        export EXTENSIONS_GALLERY='{"serviceUrl": ""}'
    fi
    if [ "$ARGUMENT" == "--load-example" ] ; then
        if [ ! -f "$VSCODE_SRV_DIR/workspace/.vscode/.startup" ]; then
            rm -rf "$VSCODE_SRV_DIR"/workspace/*
            mkdir -p "$VSCODE_SRV_DIR/workspace/.vscode/"
            cp -R /example_project/* "$VSCODE_SRV_DIR/workspace"
            echo "[Persistent Only Feature]" > "$VSCODE_SRV_DIR/workspace/.vscode/.startup"
            echo "If your CrownLabs instance is persistent, delete this file if you want to reset the workspace on next startup." >> "$VSCODE_SRV_DIR/workspace/.vscode/.startup"
        fi
    fi
    if [ -d "$ARGUMENT" ] ; then
        TARGETDIR="$ARGUMENT"
    fi
done

exec \
code-server \
--disable-update-check \
--auth none \
--bind-addr 0.0.0.0:"${CROWNLABS_LISTEN_PORT}" \
--user-data-dir "$VSCODE_SRV_DIR/data" \
--extensions-dir "$VSCODE_SRV_DIR/extensions" \
--disable-telemetry \
"$TARGETDIR"
