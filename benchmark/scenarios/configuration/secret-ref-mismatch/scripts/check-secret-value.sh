#!/bin/sh
set -eu

expected="${1:-s3cr3t}"
timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    name="$(kubectl get deployment app -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="APP_PASSWORD")].valueFrom.secretKeyRef.name}' 2>/dev/null || true)"
    key="$(kubectl get deployment app -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="APP_PASSWORD")].valueFrom.secretKeyRef.key}' 2>/dev/null || true)"
    if [ -n "$name" ] && [ -n "$key" ] && kubectl get secret "$name" >/dev/null 2>&1; then
        entries="$(kubectl get secret "$name" -o go-template='{{range $k,$v := .data}}{{$k}} {{$v}}{{"\n"}}{{end}}' 2>/dev/null || true)"
        value_b64="$(printf '%s\n' "$entries" | grep "^$key " | head -n 1 | cut -d' ' -f2- || true)"
        if [ -n "$value_b64" ]; then
            value="$(printf '%s' "$value_b64" | base64 -d 2>/dev/null || true)"
            if [ "$value" = "$expected" ]; then
                printf 'secret %s key %s holds the expected value\n' "$name" "$key"
                exit 0
            fi
        fi
    fi
    sleep 2
done

echo "timed out waiting for the referenced Secret to hold the expected value" >&2
exit 1
