#!/bin/sh
set -eu

cluster_ip="$(kubectl get service app -o jsonpath='{.spec.clusterIP}')"
selector="$(kubectl get service app -o jsonpath='{.spec.selector.app}')"
port="$(kubectl get service app -o jsonpath='{.spec.ports[0].port}')"
printf 'cluster_ip=%s selector=%s port=%s\n' "$cluster_ip" "$selector" "$port"
test "$cluster_ip" = "None" && test "$selector" = "app" && test "$port" = "80"
