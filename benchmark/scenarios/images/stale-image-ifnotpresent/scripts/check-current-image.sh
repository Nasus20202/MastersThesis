#!/bin/sh
set -eu

digest_file="${KUBECONFIG}.published-digest"
if [ ! -s "$digest_file" ]; then
    echo 'the published image digest was not recorded' >&2
    exit 1
fi
published="$(cat "$digest_file")"

deadline=$(( $(date +%s) + 120 ))
while [ "$(date +%s)" -lt "$deadline" ]; do
    running="$(kubectl get pods -l app=app --field-selector=status.phase=Running \
        -o jsonpath='{range .items[*].status.containerStatuses[*]}{.imageID}{"\n"}{end}' 2>/dev/null |
        sed -n 's/.*@\(sha256:.*\)$/\1/p')"
    count="$(printf '%s\n' "$running" | grep -c . || true)"
    unique="$(printf '%s\n' "$running" | sort -u)"
    if [ "$count" -ge 3 ] && [ "$unique" = "$published" ]; then
        printf 'all %s replicas run the published image %s\n' "$count" "$published"
        exit 0
    fi
    sleep 3
done

echo 'timed out waiting for every replica to run the published image' >&2
exit 1
