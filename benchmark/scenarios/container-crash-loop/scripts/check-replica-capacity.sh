#!/bin/sh
set -eu

replicas="$(kubectl get deployment app -o jsonpath='{.spec.replicas}')"
updated="$(kubectl get deployment app -o jsonpath='{.status.updatedReplicas}')"
ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}')"
printf 'replicas=%s updated=%s ready=%s\n' "$replicas" "$updated" "$ready"
test "$replicas" = "3" && test "$updated" = "3" && test "$ready" = "3"
