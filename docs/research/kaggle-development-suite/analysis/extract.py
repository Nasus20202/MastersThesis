"""Per-attempt timing and outcome rows from benchmark run directories.

Usage: python3 -I extract.py LABEL=RUN_DIR [LABEL=RUN_DIR ...] > attempts.csv

Wall time of each attempt is split into server queueing (response time not
covered by llama.cpp prompt+predict timings), prefill, decode and the rest
(tool calls and client overhead).
"""

import csv
import glob
import json
import os
import sys

FIELDS = [
    "label", "model", "condition", "scenario", "score", "full", "termination",
    "duration_s", "turns", "completion_tokens", "peak_context",
    "response_s", "queue_s", "prefill_s", "decode_s", "other_s", "decode_tps",
]


def rows(label, run_dir):
    for path in sorted(glob.glob(os.path.join(run_dir, "*", "*", "*.json"))):
        with open(path) as f:
            data = json.load(f)
        agent = data.get("agent") or {}
        grading = data.get("grading") or {}
        responses = agent.get("responses") or []
        response_s = prefill_s = decode_s = 0.0
        predicted = 0
        for r in responses:
            response_s += r.get("duration_seconds") or 0
            t = (r.get("response") or {}).get("timings") or {}
            prefill_s += (t.get("prompt_ms") or 0) / 1000
            decode_s += (t.get("predicted_ms") or 0) / 1000
            predicted += t.get("predicted_n") or 0
        duration = agent.get("duration_seconds") or 0
        usage = agent.get("token_usage") or {}
        turns = agent.get("turns")
        yield {
            "label": label,
            "model": (agent.get("inference") or {}).get("model", ""),
            "condition": data.get("condition") or agent.get("condition", ""),
            "scenario": data.get("scenario_id", ""),
            "score": grading.get("score", ""),
            "full": grading.get("full_success", ""),
            "termination": agent.get("termination") or ("error" if not agent else ""),
            "duration_s": round(duration, 1),
            "turns": len(turns) if isinstance(turns, list) else turns,
            "completion_tokens": usage.get("completion_tokens", ""),
            "peak_context": usage.get("peak_context_tokens", ""),
            "response_s": round(response_s, 1),
            "queue_s": round(max(0.0, response_s - prefill_s - decode_s), 1),
            "prefill_s": round(prefill_s, 1),
            "decode_s": round(decode_s, 1),
            "other_s": round(max(0.0, duration - response_s), 1),
            "decode_tps": round(predicted / decode_s, 2) if decode_s else "",
        }


writer = csv.DictWriter(sys.stdout, FIELDS)
writer.writeheader()
for arg in sys.argv[1:]:
    label, run_dir = arg.split("=", 1)
    for row in rows(label, run_dir):
        writer.writerow(row)
