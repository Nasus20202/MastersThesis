#!/bin/sh
set -eu

if kubectl get ingress app >/dev/null 2>&1; then
    echo "ingress app unexpectedly exists before the task" >&2
    exit 1
fi

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))
proxy_path='/api/v1/namespaces/default/services/http:app:80/proxy/'

while [ "$(date +%s)" -lt "$deadline" ]; do
    body="$(kubectl get --raw "$proxy_path" 2>/dev/null || true)"
    if printf '%s' "$body" | grep -q 'Welcome to nginx!'; then
        printf 'service app is healthy and no Ingress exists yet\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the Service app to respond" >&2
exit 1
