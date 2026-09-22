#!/bin/sh
set -eu

expected="${1:?expected value is required}"
timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    pods="$(kubectl get pods -l app=app -o name 2>/dev/null || true)"
    count=0
    ok=true
    for pod in $pods; do
        value="$(kubectl exec "$pod" -- printenv CONFIG_MESSAGE 2>/dev/null || true)"
        if [ "$value" != "$expected" ]; then
            ok=false
        fi
        count=$(( count + 1 ))
    done
    if [ "$ok" = true ] && [ "$count" -eq 3 ]; then
        printf 'all %s application Pods report %s\n' "$count" "$expected"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for every application Pod to report the expected value" >&2
exit 1
