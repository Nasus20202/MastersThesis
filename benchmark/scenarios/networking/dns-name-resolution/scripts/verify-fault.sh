#!/bin/sh
set -eu

timeout_seconds=120
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    value="$(kubectl get deployment app -o jsonpath='{.spec.template.spec.containers[?(@.name=="client")].env[?(@.name=="PEER_HOST")].value}' 2>/dev/null || true)"
    available="$(kubectl get deployment app -o jsonpath='{.status.availableReplicas}' 2>/dev/null || true)"
    pod="$(kubectl get pods -l app=app -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
    status=""
    if [ -n "$pod" ]; then
        status="$(kubectl exec "$pod" -- cat /shared/status 2>/dev/null || true)"
    fi
    if [ "$value" = "peer.prod" ] && [ "$status" != "ok" ] && [ "${available:-0}" -lt 2 ]; then
        printf 'peer host set to %s and the client reports %s\n' "$value" "${status:-no-status}"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the application to fail resolving its peer Service" >&2
exit 1
