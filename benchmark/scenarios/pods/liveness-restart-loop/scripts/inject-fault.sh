#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","livenessProbe":{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":1,"periodSeconds":1,"failureThreshold":1,"timeoutSeconds":1}}]}}}}'
