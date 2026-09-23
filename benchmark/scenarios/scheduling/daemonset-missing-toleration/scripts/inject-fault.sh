#!/bin/sh
set -eu

node="$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')"
kubectl taint node "$node" benchmark=daemonset:NoSchedule --overwrite

timeout_seconds=180
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    desired="$(kubectl get daemonset app -o jsonpath='{.status.desiredNumberScheduled}' 2>/dev/null || true)"
    if [ "$desired" = "0" ]; then
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the DaemonSet to stop targeting the tainted node" >&2
exit 1
