#!/bin/sh
set -eu

node="$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')"
effect="$(kubectl get node "$node" -o jsonpath='{.spec.taints[?(@.key=="dedicated")].effect}')"
printf 'taint_effect=%s\n' "$effect"
test "$effect" = "NoSchedule"

if kubectl get deployment app >/dev/null 2>&1; then
    echo "deployment app unexpectedly exists before the task" >&2
    exit 1
fi
