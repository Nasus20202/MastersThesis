#!/bin/sh
set -eu

kubectl -n registry rollout status deployment/registry --timeout=60s >/dev/null
kubectl -n registry exec deploy/registry -- \
    test -d /var/lib/registry/docker/registry/v2/repositories/private/hello/_manifests/tags/1.0
printf 'registry serves private/hello:1.0\n'
