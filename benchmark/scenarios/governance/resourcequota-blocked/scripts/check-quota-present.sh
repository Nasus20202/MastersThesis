#!/bin/sh
set -eu

kubectl get resourcequota app-quota >/dev/null
printf '%s\n' 'resourcequota app-quota present'
