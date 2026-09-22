#!/bin/sh
set -eu

kubectl patch httproute app-route --type=json \
    -p='[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway-missing"}]'
