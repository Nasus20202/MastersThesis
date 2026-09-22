#!/bin/sh
set -eu

kubectl rollout status deployment/app --timeout=120s >/dev/null
ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
if [ "${ready:-0}" -lt 3 ]; then
    echo "expected 3 ready replicas, got ${ready:-0}" >&2
    exit 1
fi
printf 'deployment app has %s ready replicas\n' "$ready"
