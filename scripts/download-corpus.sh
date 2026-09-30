#!/usr/bin/env bash

# Fetches the frozen documentation corpus (D-017): the content/en/docs subtree
# of kubernetes/website at the pinned commit, as a sparse Git checkout.

set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repository_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
cd "$repository_root/benchmark"

: "${CORPUS_REPOSITORY:?CORPUS_REPOSITORY is required}"
: "${CORPUS_REVISION:?CORPUS_REVISION is required}"
: "${CORPUS_SUBTREE:?CORPUS_SUBTREE is required}"
: "${CORPUS_DIR:?CORPUS_DIR is required}"

checkout="$CORPUS_DIR/website"

printf '[download-corpus] repository: %s\n' "$CORPUS_REPOSITORY"
printf '[download-corpus] revision: %s\n' "$CORPUS_REVISION"
printf '[download-corpus] subtree: %s\n' "$CORPUS_SUBTREE"
printf '[download-corpus] destination: benchmark/%s\n' "$checkout"

if [[ ! -d "$checkout/.git" ]]; then
  mkdir -p "$CORPUS_DIR"
  git clone --filter=blob:none --no-checkout --sparse "$CORPUS_REPOSITORY" "$checkout"
fi
git -C "$checkout" sparse-checkout set "$CORPUS_SUBTREE"
git -C "$checkout" fetch --depth 1 origin "$CORPUS_REVISION"
git -C "$checkout" -c advice.detachedHead=false checkout --force "$CORPUS_REVISION"

actual_revision=$(git -C "$checkout" rev-parse HEAD)
if [[ "$actual_revision" != "$CORPUS_REVISION" ]]; then
  printf '[download-corpus] error: checked out %s, expected %s\n' "$actual_revision" "$CORPUS_REVISION" >&2
  exit 1
fi
if [[ -n "$(git -C "$checkout" status --porcelain)" ]]; then
  printf '[download-corpus] error: the corpus checkout has local changes\n' >&2
  exit 1
fi

files=$(find "$checkout/$CORPUS_SUBTREE" -type f -name '*.md' | wc -l)
printf '[download-corpus] checkout verified; %s Markdown files\n' "$files"
