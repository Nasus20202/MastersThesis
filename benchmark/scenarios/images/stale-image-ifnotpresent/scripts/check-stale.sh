#!/bin/sh
set -eu

digest_file="${KUBECONFIG}.published-digest"
if [ ! -s "$digest_file" ]; then
    echo 'the published image digest was not recorded' >&2
    exit 1
fi
published="$(cat "$digest_file")"

deadline=$(( $(date +%s) + 60 ))
while [ "$(date +%s)" -lt "$deadline" ]; do
    running="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.imageID}{"\n"}{end}' 2>/dev/null |
        sed -n 's/.*@\(sha256:.*\)$/\1/p' |
        head -n 1)"
    if [ -n "$running" ] && [ "$published" != "$running" ]; then
        printf 'published=%s running=%s\n' "$published" "$running"
        exit 0
    fi
    sleep 2
done

echo 'timed out waiting for the running image to be stale' >&2
exit 1
