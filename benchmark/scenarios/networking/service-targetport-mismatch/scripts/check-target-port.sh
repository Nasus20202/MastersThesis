#!/bin/sh
set -eu

target_port="$(kubectl get service app -o jsonpath='{.spec.ports[0].targetPort}')"
container_port="$(kubectl get deployment app -o jsonpath='{.spec.template.spec.containers[0].ports[0].containerPort}')"
port_name="$(kubectl get deployment app -o jsonpath='{.spec.template.spec.containers[0].ports[0].name}')"
printf 'target_port=%s container_port=%s port_name=%s\n' "$target_port" "$container_port" "$port_name"
if [ "$target_port" = "$container_port" ]; then
    exit 0
fi
if [ -n "$port_name" ] && [ "$target_port" = "$port_name" ]; then
    exit 0
fi
echo "service targetPort $target_port does not route to containerPort $container_port" >&2
exit 1
