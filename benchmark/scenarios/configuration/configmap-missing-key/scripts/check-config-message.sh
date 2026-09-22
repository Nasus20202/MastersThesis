#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    pods="$(kubectl get pods -l app=app -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null || true)"
    total="$(printf '%s\n' "$pods" | grep -c . || true)"
    healthy=true
    if [ "$total" -lt 3 ]; then
        healthy=false
    fi
    for pod in $pods; do
        value="$(kubectl exec "$pod" -- printenv CONFIG_MESSAGE 2>/dev/null || true)"
        if [ "$value" != "healthy" ]; then
            healthy=false
        fi
    done
    if [ "$healthy" = true ]; then
        printf 'all %s application Pods report CONFIG_MESSAGE=healthy\n' "$total"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for every application Pod to report the ConfigMap value" >&2
exit 1
