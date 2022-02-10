#!/bin/bash -e

if [ -z "$NONVC_VER" ]; then
    NONVC_VER=v1.3.0
fi

if [ -z "$HTML_DATA" ]; then
    echo "HTML_DATA is not set"
    exit 1
fi

# ensure directory exists
mkdir -p "$HTML_DATA"

echo "Downloading and extracting noVNC $NONVC_VER"
wget -qO- https://github.com/novnc/noVNC/archive/$NONVC_VER.tar.gz | tar xz -C "$HTML_DATA" --strip 1

echo "Injecting noVNC customizations"
for filename in $(cd novnc-overrides && find . -type f); do
    echo "Customizing $filename"
    DIRNAME=$(dirname "$filename")
    mkdir -p "$HTML_DATA"/"$DIRNAME"
    if [[ $(basename "$filename") == *.js || $(basename "$filename") == *.css ]]; then
        printf "\n/* CrownLabs overrides */\n" >> "$HTML_DATA"/"$filename"
    fi
    cat novnc-overrides/"$filename" >> "$HTML_DATA"/"$filename"
done

# cleanup
echo "Cleaning up"
rm -rf "$HTML_DATA"/docs
rm -rf "$HTML_DATA"/snap
rm -rf "$HTML_DATA"/tests
rm -rf "$HTML_DATA"/utils
rm -rf "$HTML_DATA"/po
rm -rf "$HTML_DATA"/.github

echo "Done"
