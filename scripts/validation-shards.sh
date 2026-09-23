#!/usr/bin/env bash

# Prints the benchmark validation shard categories as a JSON string array.
# A category is the top-level directory under benchmark/scenarios that holds
# scenarios with a validation manifest.

set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repository_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
cd "$repository_root"

find benchmark/scenarios -name 'validation.y*ml' -printf '%h\n' \
  | sed 's#^benchmark/scenarios/##' \
  | cut -d/ -f1 \
  | sort -u \
  | jq -R -s -c 'split("\n") | map(select(length > 0))'
