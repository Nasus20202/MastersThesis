#!/bin/sh
set -eu

read_only="$(kubectl get deployment app -o jsonpath='{.spec.template.spec.containers[0].securityContext.readOnlyRootFilesystem}')"
mounts="$(kubectl get deployment app -o jsonpath='{range .spec.template.spec.containers[0].volumeMounts[*]}{.mountPath}{"\n"}{end}')"
printf 'read_only=%s mounts=%s\n' "$read_only" "$(printf '%s' "$mounts" | tr '\n' ',')"
test "$read_only" = "true"
printf '%s\n' "$mounts" | grep -qx '/var/run'
printf '%s\n' "$mounts" | grep -qx '/var/cache/nginx'
