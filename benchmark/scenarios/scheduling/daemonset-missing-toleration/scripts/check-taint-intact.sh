#!/bin/sh
set -eu

node="$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')"
value="$(kubectl get node "$node" -o jsonpath='{.spec.taints[?(@.key=="benchmark")].value}')"
effect="$(kubectl get node "$node" -o jsonpath='{.spec.taints[?(@.key=="benchmark")].effect}')"
printf 'taint_value=%s taint_effect=%s\n' "$value" "$effect"
test "$value" = "daemonset"
test "$effect" = "NoSchedule"

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    desired="$(kubectl get daemonset app -o jsonpath='{.status.desiredNumberScheduled}' 2>/dev/null || true)"
    if [ -n "$desired" ] && [ "$desired" -ge 1 ]; then
        printf 'desired=%s\n' "$desired"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the DaemonSet to cover the still-tainted node" >&2
exit 1
