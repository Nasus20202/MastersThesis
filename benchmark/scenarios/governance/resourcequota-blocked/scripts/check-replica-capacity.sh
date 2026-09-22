#!/bin/sh
set -eu

replicas="$(kubectl get deployment app -o jsonpath='{.spec.replicas}')"
ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}')"
printf 'replicas=%s ready=%s\n' "$replicas" "$ready"
test "$replicas" = "3" && test "$ready" = "3"
