#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl get ingress app >/dev/null 2>&1; then
        class="$(kubectl get ingress app -o jsonpath='{.spec.ingressClassName}' 2>/dev/null || true)"
        backends="$(kubectl get ingress app -o jsonpath='{range .spec.rules[*].http.paths[*]}{.backend.service.name}{"\n"}{end}' 2>/dev/null || true)"
        if { [ -z "$class" ] || [ "$class" = "traefik" ]; } && printf '%s\n' "$backends" | grep -qx 'app'; then
            printf 'ingress admitted with class %s and backend Service app\n' "${class:-<default>}"
            exit 0
        fi
    fi
    sleep 2
done

echo "timed out waiting for an admitted Ingress routing to Service app on port 80" >&2
exit 1
