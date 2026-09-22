#!/bin/sh
set -eu

replicas="$(kubectl get deployment app -o jsonpath='{.spec.replicas}')"
updated="$(kubectl get deployment app -o jsonpath='{.status.updatedReplicas}')"
available="$(kubectl get deployment app -o jsonpath='{.status.availableReplicas}')"
printf 'replicas=%s updated=%s available=%s\n' "$replicas" "$updated" "$available"
test "$replicas" = "3" && test "$updated" = "3" && test "$available" = "3"
