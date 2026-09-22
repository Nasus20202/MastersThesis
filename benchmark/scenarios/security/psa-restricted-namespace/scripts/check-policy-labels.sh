#!/bin/sh
set -eu

enforce="$(kubectl get namespace restricted-app -o go-template='{{ index .metadata.labels "pod-security.kubernetes.io/enforce" }}')"
warn="$(kubectl get namespace restricted-app -o go-template='{{ index .metadata.labels "pod-security.kubernetes.io/warn" }}')"
audit="$(kubectl get namespace restricted-app -o go-template='{{ index .metadata.labels "pod-security.kubernetes.io/audit" }}')"

printf 'enforce=%s warn=%s audit=%s\n' "$enforce" "$warn" "$audit"
test "$enforce" = "restricted"
test "$warn" = "restricted"
test "$audit" = "restricted"
