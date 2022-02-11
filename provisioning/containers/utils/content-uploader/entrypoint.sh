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
cd "$SOURCE_PATH"
zip "/tmp/$FILENAME.zip" -r .

echo "Uploading archive..."
HTTP_CODE=$(curl -v --request --fail POST --form "binfile=@\"/tmp/$FILENAME.zip\"" --form "filename=$FILENAME.zip" "$DESTINATION_URL" --write-out "%{http_code}")

# return non zero if HTTP_CODE is less than 200 or greater equal than 300
if [ "$HTTP_CODE" -lt "200" ] || [ "$HTTP_CODE" -ge "300" ]; then
  echo "Upload failed with HTTP code $HTTP_CODE"
  exit 1
fi
