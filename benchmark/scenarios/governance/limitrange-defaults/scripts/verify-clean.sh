#!/bin/sh
set -eu

kubectl get limitrange app-limits >/dev/null
if kubectl get deployment app >/dev/null 2>&1; then
    echo "deployment app unexpectedly exists before the task" >&2
    exit 1
fi
printf '%s\n' 'limitrange present and no app deployment yet'
