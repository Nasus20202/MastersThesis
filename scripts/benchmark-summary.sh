#!/usr/bin/env bash

# Prints a Markdown summary of one benchmark run: its state and, per
# condition, the attempts, full successes and scores.
# Usage: benchmark-summary.sh RUN_DIR.

set -euo pipefail

jq -r '
  def round3: . * 1000 | round / 1000;
  "## \(.run_id)",
  "",
  "\(.state): \(.attempts_recorded)/\(.expected_attempts) attempts, \(.error_count) errors.",
  "",
  "| Condition | Attempts | Full success | Mean score | Macro average |",
  "| --- | --- | --- | --- | --- |",
  (.by_condition | to_entries[]
    | "| \(.key) | \(.value.attempt_count)/\(.value.expected_attempts) | \(.value.full_success_count) | \(.value.mean_score | round3) | \(.value.macro_average_score | if . then round3 else "n/a" end) |")
' "${1:?run directory required}/results.json"
