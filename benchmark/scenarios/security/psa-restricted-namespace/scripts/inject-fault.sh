#!/bin/sh
set -eu

kubectl patch deployment app -n restricted-app --type=json \
    -p='[{"op":"remove","path":"/spec/template/spec/securityContext"},{"op":"remove","path":"/spec/template/spec/containers/0/securityContext"}]'
