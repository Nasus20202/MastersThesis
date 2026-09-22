#!/bin/sh
set -eu

kubectl patch role api-reader --type=json \
    -p='[{"op":"replace","path":"/rules","value":[{"apiGroups":[""],"resources":["configmaps","secrets","pods"],"verbs":["get","list","watch"]}]}]'
