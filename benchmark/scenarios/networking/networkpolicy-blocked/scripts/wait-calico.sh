#!/bin/sh
set -eu

kubectl rollout status deployment/tigera-operator -n tigera-operator --timeout=300s >/dev/null

deadline=$(( $(date +%s) + 300 ))
while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl get daemonset calico-node -n calico-system >/dev/null 2>&1; then
        if kubectl rollout status daemonset/calico-node -n calico-system --timeout=30s >/dev/null 2>&1; then
            kubectl rollout status deployment/calico-kube-controllers -n calico-system --timeout=60s >/dev/null 2>&1 || true
            exit 0
        fi
    fi
    sleep 5
done

echo "timed out waiting for Calico to become ready" >&2
exit 1
