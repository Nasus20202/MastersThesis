#!/bin/sh
set -eu

if kubectl get service app >/dev/null 2>&1; then
    echo "service app unexpectedly exists before the task" >&2
    exit 1
fi
printf '%s\n' 'no service app yet'
