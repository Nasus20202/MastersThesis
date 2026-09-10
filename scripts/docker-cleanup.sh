#!/usr/bin/env bash

set -euo pipefail

clusters=$(kind get clusters)
while IFS= read -r cluster; do
  case "$cluster" in
    benchmark*) kind delete cluster --name "$cluster" ;;
  esac
done <<< "$clusters"

containers=$(docker ps -aq --filter name=sandbox)
while IFS= read -r container; do
  if [[ -n "$container" ]]; then
    docker rm --force "$container"
  fi
done <<< "$containers"

networks=$(docker network ls -q --filter name=sandbox-network)
while IFS= read -r network; do
  if [[ -n "$network" ]]; then
    docker network rm "$network"
  fi
done <<< "$networks"
