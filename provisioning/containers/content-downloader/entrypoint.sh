#!/bin/bash -e

if [ -z "$SOURCE_ARCHIVE"  ]; then
  echo "Missing source env variable SOURCE_ARCHIVE"
  exit 1
fi

if [ -z "$DESTINATION_PATH"  ]; then
  echo "Missing destination env variable DESTINATION_PATH"
  exit 1
fi

echo "Downloading archive..."
curl --output /tmp/archive "$SOURCE_ARCHIVE"

echo "Extracting archive..."
# force-skip: skip existing file
# no-directory: extract directly in the specified directory
unar -force-skip -no-directory -output-directory "$DESTINATION_PATH" /tmp/archive
