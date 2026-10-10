#!/bin/sh

kubectl port-forward -n envoy-gateway-system svc/$GW_SERVICE_NAME 8443:443
