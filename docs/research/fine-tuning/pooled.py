#!/usr/bin/env python3
"""Validation results of the fine-tuning study, per model and pooled.

Reads the raw attempts under raw/benchmark/<run-id>/<scenario>/prompt/NNN.json
and prints Markdown tables: per model (adapter and base), per scenario, and
pooled over the four models. The paired bootstrap resamples the scenarios with
replacement, as benchmark/cmd/analyze does (10,000 draws, seed 1); the pooled
difference is the mean over the four models of each model's macro difference.

    python3 pooled.py [--seed 1] [--resamples 10000]
"""

import argparse
import json
import random
import statistics
from pathlib import Path

RAW = Path(__file__).parent / "raw" / "benchmark"

VALIDATION = [
    "antiaffinity-unschedulable",
    "configmap-missing-key",
    "container-crash-loop",
    "dns-name-resolution",
    "hostpath-fsgroup",
    "missing-rbac-binding",
    "stale-image-ifnotpresent",
    "unschedulable-cpu-request",
]

# model -> (adapter runs, base runs).
MODELS = {
    "Gemma 4 E2B": (["run-2026-11-26-11-23-19-500Z"], ["run-2026-11-26-10-55-41-090Z"]),
    "Gemma 4 E4B": (["run-2026-11-26-10-21-49-018Z"], ["run-2026-11-26-10-38-02-231Z"]),
    "Qwen3.5 4B": (["run-2026-11-26-13-54-10-398Z"], ["run-2026-11-26-12-44-57-945Z"]),
    "Qwen3.5 9B": (
        ["run-2026-11-26-15-50-36-767Z"],
        # the second run repeats antiaffinity-unschedulable, whose shard timed out waiting for the model
        ["run-2026-11-26-14-54-48-134Z", "run-2026-11-26-16-39-32-870Z"],
    ),
}


def attempts(runs, drop_server_errors=False):
    """scenario -> list of attempt dicts. Attempts that ended on an HTTP 5xx
    answer of the inference endpoint (not the model's doing) can be dropped."""
    out = {s: [] for s in VALIDATION}
    for run in runs:
        for scenario in VALIDATION:
            for path in sorted((RAW / run / scenario / "prompt").glob("0*.json")):
                attempt = json.loads(path.read_text())
                grading = attempt["grading"]
                agent = attempt["agent"]
                if drop_server_errors and agent.get("error", "").startswith("llama server returned 5"):
                    continue
                out[scenario].append(
                    {
                        "score": grading["score"],
                        "full": grading["full_success"],
                        "turns": agent.get("turns", 0),
                        "termination": agent.get("termination", ""),
                        "peak": agent.get("token_usage", {}).get("peak_context_tokens", 0),
                    }
                )
    return out


def scenario_means(data):
    return {s: statistics.mean(a["score"] for a in data[s]) for s in VALIDATION if data[s]}


def scenario_differences(adapter, base):
    """Per scenario, the adapter's mean score minus the base's."""
    a, b = scenario_means(adapter), scenario_means(base)
    return [a[s] - b[s] for s in VALIDATION if s in a and s in b]


def summary(data):
    rows = [a for s in VALIDATION for a in data[s]]
    means = scenario_means(data)
    return {
        "attempts": len(rows),
        "macro": statistics.mean(means.values()),
        "full": sum(a["full"] for a in rows),
        "turns": statistics.mean(a["turns"] for a in rows),
        "peak": statistics.mean(a["peak"] for a in rows),
        "limit": sum(a["termination"] == "turn_limit" for a in rows),
        "stopped": sum(a["termination"] == "completed" and not a["full"] for a in rows),
        "overflow": sum(a["termination"] in ("token_limit", "inference_error") for a in rows),
    }


def ci(values):
    values = sorted(values)
    return values[int(0.025 * len(values))], values[int(0.975 * len(values)) - 1]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--seed", type=int, default=1)
    parser.add_argument("--resamples", type=int, default=10000)
    args = parser.parse_args()
    rng = random.Random(args.seed)

    for drop in (False, True):
        report(rng, args.resamples, drop)
        if not drop:
            print()
            print("Sensitivity: the same tables without the attempts that ended on an HTTP 5xx answer of the inference endpoint:")
            print()


def report(rng, resamples, drop):
    data = {name: (attempts(a, drop), attempts(b, drop)) for name, (a, b) in MODELS.items()}

    print("| Model | Arm | Attempts | Macro score | Full success | Mean turns | Turn limit | Stopped without full success | Context or server error | Mean peak context |")
    print("| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
    for name, (adapter, base) in data.items():
        for arm, d in (("adapter", adapter), ("base", base)):
            s = summary(d)
            print(
                f"| {name} | {arm} | {s['attempts']} | {s['macro']:.3f} | {s['full']}/{s['attempts']} | "
                f"{s['turns']:.1f} | {s['limit']} | {s['stopped']} | {s['overflow']} | {s['peak']:,.0f} |"
            )

    print()
    print("| Model | Adapter minus base | 95% CI |")
    print("| --- | ---: | --- |")
    diffs = {name: [] for name in data}
    pooled = []
    for _ in range(resamples):
        round_diffs = []
        for name, (adapter, base) in data.items():
            pool = scenario_differences(adapter, base)
            d = statistics.mean(rng.choice(pool) for _ in pool)
            diffs[name].append(d)
            round_diffs.append(d)
        pooled.append(statistics.mean(round_diffs))
    for name, (adapter, base) in data.items():
        lo, hi = ci(diffs[name])
        print(f"| {name} | {statistics.mean(scenario_differences(adapter, base)):+.3f} | [{lo:+.3f}, {hi:+.3f}] |")
    lo, hi = ci(pooled)
    point = statistics.mean(statistics.mean(scenario_differences(a, b)) for a, b in data.values())
    print(f"| All four models, pooled | {point:+.3f} | [{lo:+.3f}, {hi:+.3f}] |")

    print()
    print("Mean score per scenario, base → adapter (3 attempts each):")
    print()
    print("| Scenario | " + " | ".join(data) + " |")
    print("| --- | " + " | ".join("---:" for _ in data) + " |")
    for s in VALIDATION:
        cells = []
        for adapter, base in data.values():
            a, b = scenario_means(adapter)[s], scenario_means(base)[s]
            cells.append(f"{b:.2f} → {a:.2f}")
        print(f"| {s} | " + " | ".join(cells) + " |")


if __name__ == "__main__":
    main()
