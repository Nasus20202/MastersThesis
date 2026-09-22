#!/bin/sh
set -eu

kubectl rollout status deployment/app --timeout=120s >/dev/null

deadline=$(( $(date +%s) + 90 ))
proxy_path='/api/v1/namespaces/default/services/http:app:80/proxy/'
while [ "$(date +%s)" -lt "$deadline" ]; do
    body="$(kubectl get --raw "$proxy_path" 2>/dev/null || true)"
    if printf '%s' "$body" | grep -q 'Welcome to nginx!'; then
        printf '%s\n' 'application served the expected response'
        exit 0
    fi
    sleep 2
done

echo 'timed out waiting for the application response' >&2
exit 1
