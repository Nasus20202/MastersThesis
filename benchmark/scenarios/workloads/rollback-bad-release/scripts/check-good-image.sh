#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    states="$(kubectl get pods -l app=app -o jsonpath='{range .items[*]}{.status.phase}{" "}{.spec.containers[0].image}{"\n"}{end}' 2>/dev/null || true)"
    total="$(printf '%s\n' "$states" | grep -c . || true)"
    good="$(printf '%s\n' "$states" | grep -c '^Running nginx:1.31.5$' || true)"
    if [ "$total" = "3" ] && [ "$good" = "3" ]; then
        printf 'running_good_image=%s\n' "$good"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for three Pods to run nginx:1.31.5" >&2
exit 1
