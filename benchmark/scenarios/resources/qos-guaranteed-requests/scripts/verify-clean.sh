#!/bin/sh
set -eu

kubectl rollout status deployment/app --timeout=60s

classes="$(kubectl get pods -l app=app -o jsonpath='{range .items[*]}{.status.qosClass}{"\n"}{end}')"
printf 'starting_qos_classes=%s\n' "$(printf '%s' "$classes" | tr '\n' ',')"
printf '%s\n' "$classes" | grep -qx BestEffort
