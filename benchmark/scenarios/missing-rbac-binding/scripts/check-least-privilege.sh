#!/bin/sh
set -eu

service_account='system:serviceaccount:default:config-reader'
can_get_named="$(kubectl auth can-i get configmap/app-config --namespace=default --as="$service_account" || true)"
can_list_configmaps="$(kubectl auth can-i list configmaps --namespace=default --as="$service_account" || true)"
can_get_secrets="$(kubectl auth can-i get secrets --namespace=default --as="$service_account" || true)"

printf 'get_app_config=%s list_configmaps=%s get_secrets=%s\n' \
    "$can_get_named" "$can_list_configmaps" "$can_get_secrets"

test "$can_get_named" = "yes"
test "$can_list_configmaps" = "no"
test "$can_get_secrets" = "no"
