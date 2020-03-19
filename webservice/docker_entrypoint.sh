#!/bin/bash

echo "building the service..."
npm run build

mv ./webservice/website/dist /usr/share/nginx/html

echo "starting nginx..."
nginx -g 'daemon off;'
