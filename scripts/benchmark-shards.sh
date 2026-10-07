#!/usr/bin/env bash

# Splits the scenario files selected by SCENARIO and TAG (make benchmark-list)
# into at most SHARDS shards and prints them as a JSON array of space-separated
# paths, one benchmark job each. Scenarios are dealt round-robin, so shard sizes
# differ by at most one and no shard holds a whole category.
# Usage: benchmark-shards.sh SHARDS.

set -euo pipefail

count="${1:?shard count required}"
repository_root=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)

list=$(make -s -C "$repository_root" benchmark-list)
mapfile -t scenarios <<<"$list"

shards=()
for i in "${!scenarios[@]}"; do
  shard=$((i % count))
  shards[shard]+="${shards[shard]:+ }${scenarios[i]}"
done

jq -nc '$ARGS.positional' --args "${shards[@]}"
