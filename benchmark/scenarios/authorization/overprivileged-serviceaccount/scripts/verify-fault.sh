#!/bin/sh
set -eu

service_account='system:serviceaccount:default:api-reader'
list_configmaps="$(kubectl auth can-i list configmaps --namespace=default --as="$service_account" || true)"
get_secrets="$(kubectl auth can-i get secrets --namespace=default --as="$service_account" || true)"

printf 'list_configmaps=%s get_secrets=%s\n' "$list_configmaps" "$get_secrets"
test "$list_configmaps" = "yes"
test "$get_secrets" = "yes"
