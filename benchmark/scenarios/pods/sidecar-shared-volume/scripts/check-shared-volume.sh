#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    count="$(kubectl get pod app -o jsonpath='{.spec.containers[*].name}' 2>/dev/null | wc -w | tr -d ' ' || true)"
    if [ "${count:-0}" = "2" ]; then
        empties="$(kubectl get pod app -o jsonpath='{range .spec.volumes[?(@.emptyDir)]}{.name}{"\n"}{end}' 2>/dev/null || true)"
        shared=false
        for vol in $empties; do
            m0="$(kubectl get pod app -o jsonpath="{.spec.containers[0].volumeMounts[?(@.name=='$vol')].name}" 2>/dev/null || true)"
            m1="$(kubectl get pod app -o jsonpath="{.spec.containers[1].volumeMounts[?(@.name=='$vol')].name}" 2>/dev/null || true)"
            if [ "$m0" = "$vol" ] && [ "$m1" = "$vol" ]; then
                shared=true
            fi
        done
        if [ "$shared" = true ]; then
            echo "an emptyDir volume is mounted by both containers"
            exit 0
        fi
    fi
    sleep 2
done

echo "timed out waiting for an emptyDir volume shared by both containers" >&2
exit 1
