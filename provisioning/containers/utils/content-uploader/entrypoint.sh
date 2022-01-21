#!/bin/bash -e

if [ -z "$SOURCE_PATH"  ]; then
  echo "Missing source env variable SOURCE_PATH"
  exit 1
fi

if [ -z "$DESTINATION_URL"  ]; then
  echo "Missing destination env variable DESTINATION_URL"
  exit 1
fi

if [ -z "$FILENAME"  ]; then
  echo "Missing destination env variable FILENAME"
  exit 1
fi

echo "Compressing archive..."
zip "/tmp/$FILENAME.zip" -r "$SOURCE_PATH"


echo "Uploading archive..."
curl -v --fail --request POST --form "binfile=@\"/tmp/$FILENAME.zip\"" --form "filename=$FILENAME.zip" "$DESTINATION_URL"
