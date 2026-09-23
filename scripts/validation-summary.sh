#!/usr/bin/env bash

# Prints a Markdown summary of benchmark validation results for the GitHub
# Step Summary. Usage: validation-summary.sh RESULTS_DIR CATEGORY.

set -euo pipefail

results_dir="${1:?results directory required}"
category="${2:?category required}"

mapfile -t files < <(find "$results_dir" -type f -name '*.json' 2>/dev/null)
if [ "${#files[@]}" -eq 0 ]; then
  printf '## Benchmark validation: %s\n\nNo validation result files were produced.\n' "$category"
  exit 0
fi

jq -s -r --arg category "$category" '
  map(objects | select(has("passed"))) as $cases
  | ([$cases[] | select(.passed)] | length) as $passed
  | ($cases | length) as $total
  | "## Benchmark validation: \($category)",
    "",
    "**\($passed)/\($total) cases passed** (\($total - $passed) failed).",
    (if $total - $passed > 0 then
       "",
       "| Case | Score | Expected | Error |",
       "| --- | --- | --- | --- |",
       ($cases[] | select(.passed | not)
         | "| `\(.scenario_id)/\(.case_id)` | \(.grading.score // "n/a") | \(.expected_score // "n/a") | \(.error // "" | gsub("\n"; " ")) |")
     else empty end)
' "${files[@]}"
