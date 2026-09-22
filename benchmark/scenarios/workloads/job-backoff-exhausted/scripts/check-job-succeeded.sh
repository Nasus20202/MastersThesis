#!/bin/sh
set -eu

succeeded="$(kubectl get job app -o jsonpath='{.status.succeeded}')"
printf 'succeeded=%s\n' "$succeeded"
test -n "$succeeded" && test "$succeeded" -ge 1
