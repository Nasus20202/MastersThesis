#!/bin/sh
set -eu

secrets="$(kubectl get deployment app -o jsonpath='{range .spec.template.spec.imagePullSecrets[*]}{.name}{"\n"}{end}' 2>/dev/null || true)"
if [ -z "$secrets" ]; then
    echo 'deployment app does not reference an imagePullSecret' >&2
    exit 1
fi

for name in $secrets; do
    secret_type="$(kubectl get secret "$name" -o jsonpath='{.type}' 2>/dev/null || true)"
    if [ "$secret_type" != "kubernetes.io/dockerconfigjson" ]; then
        continue
    fi
    config="$(kubectl get secret "$name" -o jsonpath='{.data.\.dockerconfigjson}' 2>/dev/null | base64 -d 2>/dev/null || true)"
    if printf '%s' "$config" | grep -q 'registry:5000'; then
        printf 'imagePullSecret %s authenticates registry:5000\n' "$name"
        exit 0
    fi
done

echo 'no referenced imagePullSecret authenticates registry:5000' >&2
exit 1
