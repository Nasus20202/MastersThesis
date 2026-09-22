#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl exec app -c web -- sh -c 'test -s /srv/index.html' 2>/dev/null; then
        body="$(kubectl exec app -c web -- wget -qO- http://127.0.0.1:8080/ 2>/dev/null || true)"
        if [ -n "$body" ] && ! printf '%s' "$body" | grep -qi '404'; then
            echo "the web server serves content from the shared volume"
            exit 0
        fi
    fi
    sleep 2
done

echo "timed out waiting for the web server to serve content from the shared volume" >&2
exit 1
