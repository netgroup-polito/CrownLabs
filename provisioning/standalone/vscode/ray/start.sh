#!/usr/bin/env bash
shopt -s dotglob

TARGETDIR="$VSCODE_SRV_DIR/workspace"

while [ "$#" -gt 0 ]; do
    case "$1" in
        --disable-marketplace)
            export EXTENSIONS_GALLERY='{"serviceUrl": ""}'
            ;;

        # ---- LOAD EXAMPLE ----
        --load-example)
            if [ ! -f "$VSCODE_SRV_DIR/workspace/.vscode/.startup" ]; then
                rm -rf "$VSCODE_SRV_DIR"/workspace/*
                mkdir -p "$VSCODE_SRV_DIR/workspace/.vscode/"
                cp -R /example_project/* "$VSCODE_SRV_DIR/workspace"
                echo "[Persistent Only Feature]" > "$VSCODE_SRV_DIR/workspace/.vscode/.startup"
                echo "If your CrownLabs instance is persistent, delete this file if you want to reset the workspace on next startup." >> "$VSCODE_SRV_DIR/workspace/.vscode/.startup"
            fi
            ;;

        # ---- RAY_ADDRESS ----
        --ray-address)
            shift
            if [ -z "$1" ]; then
                echo "Error: --ray-address requires a value" >&2
                exit 1
            fi
            export RAY_ADDRESS="$1"
            ;;
        --ray-address=*)
            export RAY_ADDRESS="${1#--ray-address=}"
            ;;

        # ---- SHARED_STORAGE_PATH ----
        --shared-storage-path)
            shift
            if [ -z "$1" ]; then
                echo "Error: --shared-storage-path requires a value" >&2
                exit 1
            fi
            export SHARED_STORAGE_PATH="$1"
            ;;
        --shared-storage-path=*)
            export SHARED_STORAGE_PATH="${1#--shared-storage-path=}"
            ;;

        # ---- WORKSPACE_PVC_NAME ----
        --workspace-pvc-name)
            shift
            if [ -z "$1" ]; then
                echo "Error: --workspace-pvc-name requires a value" >&2
                exit 1
            fi
            export WORKSPACE_PVC_NAME="$1"
            ;;
        --workspace-pvc-name=*)
            export WORKSPACE_PVC_NAME="${1#--workspace-pvc-name=}"
            ;;

        # ---- DEFAULT / TARGETDIR ----
        *)
            if [ -d "$1" ]; then
                TARGETDIR="$1"
            fi
            ;;
    esac

    shift
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