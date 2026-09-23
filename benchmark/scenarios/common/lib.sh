#!/bin/sh
# Shared helpers for scenario scripts. Source this file from a script with:
#   . "$(dirname "$0")/../../../common/lib.sh"

# Prints the name of a Running, non-terminating Pod matching the label
# selector passed as the first argument, or nothing when no such Pod exists.
# When the selector matches a Deployment, only Pods from its current
# ReplicaSet (the one with the highest deployment.kubernetes.io/revision) are
# considered, so a rollout in progress cannot return a Pod from the previous
# generation. Terminating Pods are always ignored.
running_pod() {
    _selector=$1
    _hash="$(kubectl get replicasets -l "$_selector" \
        -o go-template='{{range .items}}{{index .metadata.annotations "deployment.kubernetes.io/revision"}} {{.metadata.labels.pod-template-hash}}{{"\n"}}{{end}}' \
        2>/dev/null | sort -n | tail -n 1 | cut -d' ' -f2)"
    if [ -n "$_hash" ]; then
        _selector="$_selector,pod-template-hash=$_hash"
    fi
    kubectl get pods -l "$_selector" \
        -o go-template='{{range .items}}{{if and (eq .status.phase "Running") (not .metadata.deletionTimestamp)}}{{.metadata.name}}{{"\n"}}{{end}}{{end}}' \
        2>/dev/null | head -n 1
}

# Prints the names of the non-terminating Pods matching the label selector,
# one per line.
pods_by_label() {
    kubectl get pods -l "$1" \
        -o go-template='{{range .items}}{{if not .metadata.deletionTimestamp}}{{.metadata.name}}{{"\n"}}{{end}}{{end}}' \
        2>/dev/null
}

# Prints the container restart counts of the matching Pods, one Pod per line.
restart_counts() {
    for _pod in $(pods_by_label "$1"); do
        kubectl get pod "$_pod" -o jsonpath='{.status.containerStatuses[*].restartCount}' 2>/dev/null || true
        printf '\n'
    done
}

# Prints the container waiting and last-terminated reasons of the matching
# Pods, one line per container.
container_markers() {
    for _pod in $(pods_by_label "$1"); do
        kubectl get pod "$_pod" -o jsonpath='{range .status.containerStatuses[*]}{.state.waiting.reason} {.lastState.terminated.reason}{"\n"}{end}' 2>/dev/null || true
    done
}

# Succeeds when workload KIND/NAME in namespace (default "default") has
# spec.replicas and status.readyReplicas equal to EXPECTED. Prints the
# observed counts on success.
workload_ready() {
    _kind=$1
    _name=$2
    _expected=$3
    _namespace=${4:-default}
    _replicas="$(kubectl get "$_kind" "$_name" -n "$_namespace" -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
    _ready="$(kubectl get "$_kind" "$_name" -n "$_namespace" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ "$_replicas" = "$_expected" ] && [ "$_ready" = "$_expected" ]; then
        printf 'replicas=%s ready=%s\n' "$_replicas" "$_ready"
        return 0
    fi
    return 1
}

# Succeeds when workload KIND/NAME in namespace (default "default") has
# spec.replicas, status.updatedReplicas and status.readyReplicas equal to
# EXPECTED. Prints the observed counts on success.
workload_rollout_ready() {
    _kind=$1
    _name=$2
    _expected=$3
    _namespace=${4:-default}
    _replicas="$(kubectl get "$_kind" "$_name" -n "$_namespace" -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
    _updated="$(kubectl get "$_kind" "$_name" -n "$_namespace" -o jsonpath='{.status.updatedReplicas}' 2>/dev/null || true)"
    _ready="$(kubectl get "$_kind" "$_name" -n "$_namespace" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ "$_replicas" = "$_expected" ] && [ "$_updated" = "$_expected" ] && [ "$_ready" = "$_expected" ]; then
        printf 'replicas=%s updated=%s ready=%s\n' "$_replicas" "$_updated" "$_ready"
        return 0
    fi
    return 1
}
