#!/bin/sh
set -eu

node="$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')"
effect="$(kubectl get node "$node" -o jsonpath='{.spec.taints[?(@.key=="dedicated")].effect}')"
keys="$(kubectl get deployment app -o jsonpath='{range .spec.template.spec.tolerations[*]}{.key}{"\n"}{end}')"
values="$(kubectl get deployment app -o jsonpath='{range .spec.template.spec.tolerations[*]}{.value}{"\n"}{end}')"
effects="$(kubectl get deployment app -o jsonpath='{range .spec.template.spec.tolerations[*]}{.effect}{"\n"}{end}')"
printf 'taint=%s toleration_keys=%s\n' "$effect" "$(printf '%s' "$keys" | tr '\n' ',')"
test "$effect" = "NoSchedule"
printf '%s\n' "$keys" | grep -qx dedicated
printf '%s\n' "$values" | grep -qx gpu
printf '%s\n' "$effects" | grep -qx NoSchedule
