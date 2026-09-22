#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    classes="$(kubectl get pods -l app=app -o jsonpath='{range .items[*]}{.status.qosClass}{"\n"}{end}' 2>/dev/null || true)"
    total="$(printf '%s\n' "$classes" | grep -c . || true)"
    guaranteed="$(printf '%s\n' "$classes" | grep -cx Guaranteed || true)"
    if [ "$total" -ge 1 ] && [ "$total" = "$guaranteed" ]; then
        printf 'qos_classes=%s\n' "$(printf '%s' "$classes" | tr '\n' ',')"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for Guaranteed QoS" >&2
exit 1
