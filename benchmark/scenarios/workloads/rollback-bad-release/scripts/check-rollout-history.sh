#!/bin/sh
set -eu

image="$(kubectl get deployment app -o jsonpath='{.spec.template.spec.containers[0].image}')"
revisions="$(kubectl rollout history deployment/app 2>/dev/null | grep -cE '^[0-9]+' || true)"
printf 'current_image=%s revisions=%s\n' "$image" "$revisions"
test "$revisions" -ge 2
test "$image" = "nginx:1.31.5"
