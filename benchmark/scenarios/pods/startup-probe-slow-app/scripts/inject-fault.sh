#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","startupProbe":null,"livenessProbe":{"httpGet":{"path":"/","port":80},"initialDelaySeconds":1,"periodSeconds":1,"failureThreshold":1,"timeoutSeconds":1}}]}}}}'
