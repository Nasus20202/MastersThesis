#!/usr/bin/env bash
set -uo pipefail

IMAGE="ghcr.io/ggml-org/llama.cpp:server-vulkan-b10964@sha256:43e0e25ca654d839ebda39fd6c2f200b36e9efb3e597ba90d0aaeff1be95ca53"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
MODELS="$ROOT/benchmark/models"
PRESET="$ROOT/benchmark/models-preset.ini"
PORT=18099
OUT="${SWEEP_OUT:-/tmp/opencode/mtp-sweep10}"
RAW="$OUT/raw"
mkdir -p "$RAW"

grep -v '^model-draft' "$PRESET" > "$OUT/models-preset-nodraft.ini"

prompts=(
"Explain in detail how a bicycle works, step by step."
"Write a Python function that merges two sorted lists, and explain its time and space complexity."
"Summarize the main causes of the First World War in three paragraphs."
"A train leaves Station A at 14:00 travelling at 80 km/h. Another leaves Station B, 300 km away, at 14:30 travelling at 100 km/h towards A. When and where do they meet? Show your reasoning."
"List the main differences between TCP and UDP, with one example use case for each."
"Describe how photosynthesis works in plants, from light absorption to glucose production."
"Write a short story about a lighthouse keeper who discovers something unusual."
"Explain recursion, and give a worked example in JavaScript."
"Compare relational databases and document databases, discussing the trade-offs."
"Describe the water cycle, including evaporation, condensation, precipitation and collection."
)

vram_used() { rocm-smi --showmeminfo vram 2>/dev/null | awk -F: '/Used/{gsub(/ /,"",$3); print $3; exit}'; }

[[ -f "$OUT/runs.csv" ]] || echo "profile,model,spec,nmax,prompt_id,tg_tps,accept,meanlen,gen_tokens" > "$OUT/runs.csv"
[[ -f "$OUT/vram.csv" ]] || echo "profile,spec,nmax,vram_bytes" > "$OUT/vram.csv"

run_setting() {
  local profile=$1 name=$2 spectype=$3 nmax=$4 preset=$5
  docker rm -f mtp-sweep >/dev/null 2>&1
  local args=(--models-dir /models --models-max 1 --models-autoload
    --models-preset /models-preset.ini
    --spec-type "$spectype" --host 0.0.0.0 --port 8080
    --kv-unified-per-slot 32768 --parallel 2 --flash-attn auto
    --cache-type-k f16 --cache-type-v f16 --n-gpu-layers 999
    --reasoning on --reasoning-budget 4096)
  [[ -n "$nmax" ]] && args+=(--spec-draft-n-max "$nmax")

  docker run -d --rm --name mtp-sweep --device /dev/dri/renderD128 \
    -v "$MODELS:/models:ro" -v "$preset:/models-preset.ini:ro" -p "$PORT:8080" \
    "$IMAGE" "${args[@]}" >/dev/null

  local ready=no
  for _ in $(seq 1 60); do
    curl -sf -m 2 "http://127.0.0.1:$PORT/health" >/dev/null 2>&1 && { ready=yes; break; }
    sleep 1
  done
  if [[ "$ready" != yes ]]; then
    echo "[$profile $spectype ${nmax:-off}] ERROR: server not ready" >&2
    docker rm -f mtp-sweep >/dev/null 2>&1
    return
  fi

  for i in "${!prompts[@]}"; do
    local tag="${profile}-${spectype}-${nmax:-off}-$(printf '%02d' "$i")"
    curl -s -m 900 "http://127.0.0.1:$PORT/v1/chat/completions" \
      -H 'Content-Type: application/json' \
      -d "$(python3 -c 'import json,sys; print(json.dumps({"model":sys.argv[1],"messages":[{"role":"user","content":sys.argv[2]}],"max_tokens":256,"temperature":0}))' \
          "$name" "${prompts[$i]}")" > "$RAW/$tag.json"
    sleep 1

    local logs tg accept meanlen gen
    logs=$(docker logs mtp-sweep 2>&1)
    tg=$(printf '%s' "$logs" | grep 'eval time =' | tail -1 | grep -oE '[0-9.]+ tokens per second' | grep -oE '^[0-9.]+')
    gen=$(printf '%s' "$logs" | grep 'eval time =' | tail -1 | grep -oE '/ *[0-9]+ tokens' | grep -oE '[0-9]+' | tail -1)
    accept=$(printf '%s' "$logs" | grep 'draft acceptance' | tail -1 | sed -E 's/.*draft acceptance = ([0-9.]+).*/\1/')
    meanlen=$(printf '%s' "$logs" | grep 'draft acceptance' | tail -1 | sed -E 's/.*mean len = *([0-9.]+).*/\1/')
    [[ "$accept" == *draft* ]] && accept=""
    [[ "$meanlen" == *mean* ]] && meanlen=""
    echo "$profile,$name,$spectype,${nmax:-off},$i,${tg:-},${accept:-},${meanlen:-},${gen:-}" >> "$OUT/runs.csv"
    printf '  %-8s %-6s run %02d: %6s t/s acc=%-7s len=%-5s gen=%s\n' \
      "$profile" "${nmax:-off}" "$i" "${tg:-?}" "${accept:--}" "${meanlen:--}" "${gen:-?}"
  done

  echo "$profile,$spectype,${nmax:-off},$(vram_used)" >> "$OUT/vram.csv"
  docker rm -f mtp-sweep >/dev/null 2>&1
}

for profile in "$@"; do
  case "$profile" in
    e2b) name=gemma-4-E2B-it-qat-UD-Q4_K_XL ;;
    e4b) name=gemma-4-E4B-it-qat-UD-Q4_K_XL ;;
    q4b) name=Qwen3.5-4B-Q4_K_M ;;
    q9b) name=Qwen3.5-9B-Q4_K_M ;;
    *) echo "unknown profile $profile"; continue ;;
  esac
  echo "### $profile ($name)"
  run_setting "$profile" "$name" none "" "$OUT/models-preset-nodraft.ini"
  for n in 1 2 3 4; do
    run_setting "$profile" "$name" draft-mtp "$n" "$PRESET"
  done
done
echo "done -> $OUT/runs.csv"
