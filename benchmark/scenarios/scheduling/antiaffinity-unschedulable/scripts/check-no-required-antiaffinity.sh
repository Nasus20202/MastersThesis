#!/bin/sh
set -eu

terms="$(kubectl get deployment app -o jsonpath='{range .spec.template.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[*]}{.topologyKey}{"|"}{.labelSelector.matchLabels.app}{"\n"}{end}')"
printf 'required_terms=%s\n' "$(printf '%s' "$terms" | tr '\n' ',')"

if printf '%s\n' "$terms" | grep -Fxq 'kubernetes.io/hostname|app'; then
    echo "an unsatisfiable required pod anti-affinity rule remains" >&2
    exit 1
fi

exit 0
