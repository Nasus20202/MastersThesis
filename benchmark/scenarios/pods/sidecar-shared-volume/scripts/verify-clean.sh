#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    ready="$(kubectl get pod app -o jsonpath='{range .status.conditions[?(@.type=="Ready")]}{.status}{end}' 2>/dev/null || true)"
    if [ "$ready" = "True" ] && kubectl exec app -c web -- sh -c 'test ! -s /srv/index.html' 2>/dev/null; then
        echo "the prepared Pod is Ready and serves an empty directory"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the prepared Pod serving an empty directory" >&2
exit 1
