#!/usr/bin/env python3
"""Benchmark results of the fine-tuning study, as Markdown tables.

Reads the raw attempts under raw/benchmark/<run-id>/<scenario>/prompt/NNN.json.

1. Validation runs (8 scenarios, 3 attempts): per model (adapter and base), per
   scenario, and pooled over the four models; then the same without the
   attempts that ended on an HTTP 5xx answer of the inference endpoint.
2. First sweeps (46 scenarios, 1 attempt): per model, split into the 38
   trained-on and the 8 validation scenarios, against the base where a base
   sweep exists (Qwen3.5 9B teacher only).
3. With --teacher deepseek-flash-data: the validation tables of the DeepSeek
   V4.1 Flash teacher's adapters, and their difference to the Qwen3.5 9B
   teacher's adapters on the same scenarios.

The paired bootstrap resamples the scenarios with replacement, as
benchmark/cmd/analyze does (10,000 draws, seed 1); the pooled difference is the
mean over the four models of each model's macro difference.

    python3 results.py [--teacher qwen35-9b-data|deepseek-flash-data] [--seed 1] [--resamples 10000]
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

# Validation runs: model -> (adapter runs, base runs) of the Qwen3.5 9B teacher.
QWEN_RUNS = {
    "Gemma 4 E2B": (["run-2026-11-26-11-23-19-500Z"], ["run-2026-11-26-10-55-41-090Z"]),
    "Gemma 4 E4B": (["run-2026-11-26-10-21-49-018Z"], ["run-2026-11-26-10-38-02-231Z"]),
    "Qwen3.5 4B": (["run-2026-11-26-13-54-10-398Z"], ["run-2026-11-26-12-44-57-945Z"]),
    "Qwen3.5 9B": (
        ["run-2026-11-26-15-50-36-767Z"],
        # the second run repeats antiaffinity-unschedulable, whose shard timed out waiting for the model
        ["run-2026-11-26-14-54-48-134Z", "run-2026-11-26-16-39-32-870Z"],
    ),
}

# The DeepSeek V4.1 Flash teacher's adapters, same scenarios and bases.
DEEPSEEK_ADAPTER_RUNS = {
    "Gemma 4 E2B": ["run-2026-11-27-02-51-29-827Z"],
    "Gemma 4 E4B": ["run-2026-11-27-00-57-17-891Z"],
    "Qwen3.5 4B": ["run-2026-11-27-02-15-34-812Z"],
    "Qwen3.5 9B": ["run-2026-11-27-04-00-17-183Z"],
}
DEEPSEEK_RUNS = {
    name: (DEEPSEEK_ADAPTER_RUNS[name], base) for name, (_, base) in QWEN_RUNS.items()
}
VALIDATION_RUNS = QWEN_RUNS


def attempts(runs, drop_server_errors=False, scenarios=VALIDATION):
    """scenario -> list of attempt dicts. Attempts that ended on an HTTP 5xx
    answer of the inference endpoint (not the model's doing) can be dropped."""
    out = {s: [] for s in scenarios}
    for run in runs:
        for scenario in scenarios:
            for path in sorted((RAW / run / scenario / "prompt").glob("0*.json")):
                attempt = json.loads(path.read_text())
                grading = attempt["grading"]
                agent = attempt["agent"]
                if drop_server_errors and agent.get("error", "").startswith(
                    "llama server returned 5"
                ):
                    continue
                timings = [
                    r["response"].get("timings", {}) for r in agent.get("responses", [])
                ]
                out[scenario].append(
                    {
                        "score": grading["score"],
                        "full": grading["full_success"],
                        "turns": agent.get("turns", 0),
                        "termination": agent.get("termination", ""),
                        "peak": agent.get("token_usage", {}).get(
                            "peak_context_tokens", 0
                        ),
                        "completion": agent.get("token_usage", {}).get(
                            "completion_tokens", 0
                        ),
                        "drafted": sum(t.get("draft_n", 0) for t in timings),
                        "accepted": sum(t.get("draft_n_accepted", 0) for t in timings),
                    }
                )
    return out


def scenario_means(data):
    return {
        s: statistics.mean(a["score"] for a in rows) for s, rows in data.items() if rows
    }


def scenario_differences(adapter, base):
    """Per scenario, the adapter's mean score minus the base's."""
    a, b = scenario_means(adapter), scenario_means(base)
    return [a[s] - b[s] for s in a if s in b]


def summary(data):
    rows = [a for s in data for a in data[s]]
    means = scenario_means(data)
    return {
        "attempts": len(rows),
        "macro": statistics.mean(means.values()),
        "full": sum(a["full"] for a in rows),
        "turns": statistics.mean(a["turns"] for a in rows),
        "peak": statistics.mean(a["peak"] for a in rows),
        "limit": sum(a["termination"] == "turn_limit" for a in rows),
        "timeout": sum(a["termination"] == "timeout" for a in rows),
        # generated tokens per agent turn
        "per_turn": sum(a["completion"] for a in rows)
        / max(1, sum(a["turns"] for a in rows)),
        "stopped": sum(a["termination"] == "completed" and not a["full"] for a in rows),
        "overflow": sum(
            a["termination"] in ("token_limit", "inference_error") for a in rows
        ),
        # accepted over drafted tokens of the multi-token-prediction head, over all responses
        "draft": sum(a["accepted"] for a in rows)
        / max(1, sum(a["drafted"] for a in rows)),
    }


def ci(values):
    values = sorted(values)
    return values[int(0.025 * len(values))], values[int(0.975 * len(values)) - 1]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--teacher",
        choices=("qwen35-9b-data", "deepseek-flash-data"),
        default="qwen35-9b-data",
    )
    parser.add_argument("--seed", type=int, default=1)
    parser.add_argument("--resamples", type=int, default=10000)
    args = parser.parse_args()
    rng = random.Random(args.seed)

    if args.teacher == "deepseek-flash-data":
        report(rng, args.resamples, False, DEEPSEEK_RUNS)
        print()
        teacher_report(random.Random(args.seed), args.resamples)
        return

    for drop in (False, True):
        report(rng, args.resamples, drop)
        if not drop:
            print()
            print(
                "Sensitivity: the same tables without the attempts that ended on an HTTP 5xx answer of the inference endpoint:"
            )
            print()
    print()
    sweep_report(random.Random(args.seed), args.resamples)


def report(rng, resamples, drop, runs=None):
    runs = runs or VALIDATION_RUNS
    data = {
        name: (attempts(a, drop), attempts(b, drop)) for name, (a, b) in runs.items()
    }

    print(
        "| Model | Arm | Attempts | Macro score | Full success | Mean turns | Turn limit | Stopped without full success | Context or server error | Timed out | Mean peak context | Tokens per turn | Draft acceptance |"
    )
    print(
        "| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |"
    )
    for name, (adapter, base) in data.items():
        for arm, d in (("adapter", adapter), ("base", base)):
            s = summary(d)
            print(
                f"| {name} | {arm} | {s['attempts']} | {s['macro']:.3f} | {s['full']}/{s['attempts']} | "
                f"{s['turns']:.1f} | {s['limit']} | {s['stopped']} | {s['overflow']} | {s['timeout']} | {s['peak']:,.0f} | "
                f"{s['per_turn']:.0f} | {s['draft']:.3f} |"
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
        print(
            f"| {name} | {statistics.mean(scenario_differences(adapter, base)):+.3f} | [{lo:+.3f}, {hi:+.3f}] |"
        )
    lo, hi = ci(pooled)
    point = statistics.mean(
        statistics.mean(scenario_differences(a, b)) for a, b in data.values()
    )
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


def teacher_report(rng, resamples):
    """DeepSeek-teacher adapter minus Qwen-teacher adapter, per model and pooled."""
    pairs = {
        name: (attempts(DEEPSEEK_ADAPTER_RUNS[name]), attempts(QWEN_RUNS[name][0]))
        for name in QWEN_RUNS
    }
    print("| Model | DeepSeek minus Qwen3.5 9B teacher | 95% CI |")
    print("| --- | ---: | --- |")
    diffs = {name: [] for name in pairs}
    pooled = []
    for _ in range(resamples):
        round_diffs = []
        for name, (deepseek, qwen) in pairs.items():
            pool = scenario_differences(deepseek, qwen)
            d = statistics.mean(rng.choice(pool) for _ in pool)
            diffs[name].append(d)
            round_diffs.append(d)
        pooled.append(statistics.mean(round_diffs))
    for name, (deepseek, qwen) in pairs.items():
        lo, hi = ci(diffs[name])
        print(
            f"| {name} | {statistics.mean(scenario_differences(deepseek, qwen)):+.3f} | [{lo:+.3f}, {hi:+.3f}] |"
        )
    lo, hi = ci(pooled)
    point = statistics.mean(
        statistics.mean(scenario_differences(a, b)) for a, b in pairs.values()
    )
    print(f"| All four models, pooled | {point:+.3f} | [{lo:+.3f}, {hi:+.3f}] |")


# First sweeps: model -> (adapter run, base run or None).
SWEEP_RUNS = {
    "Gemma 4 E2B": ("run-2026-11-26-09-10-02-271Z", None),
    "Gemma 4 E4B": ("run-2026-11-26-08-04-10-096Z", "run-2026-11-26-08-39-29-198Z"),
    "Qwen3.5 4B": ("run-2026-11-26-17-55-55-486Z", None),
    "Qwen3.5 9B": ("run-2026-11-26-18-54-30-234Z", None),
}

# The Qwen3.5 9B teacher has no trajectory on readonly-rootfs, so its adapters
# trained on 37 of the 38 (training-dataset study).
UNTRAINED_BY_QWEN_TEACHER = "readonly-rootfs"


def scenarios_of(run):
    return sorted(p.parent.parent.name for p in (RAW / run).glob("*/prompt/0*.json"))


def sweep_report(rng, resamples):

    every = scenarios_of(next(iter(SWEEP_RUNS.values()))[0])
    trained = [s for s in every if s not in VALIDATION]
    groups = {"all 46": every, "38 trained-on": trained, "8 validation": VALIDATION}
    data = {
        name: {
            group: (
                attempts([a], scenarios=scen),
                attempts([b], scenarios=scen) if b else None,
            )
            for group, scen in groups.items()
        }
        for name, (a, b) in SWEEP_RUNS.items()
    }

    print(
        "| Model | Scenarios | Arm | Macro score | Full success | Mean turns | Turn limit | Stopped without full success | Context or server error | Mean peak context | Draft acceptance |"
    )
    print("| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
    for name, by_group in data.items():
        for group, (adapter, base) in by_group.items():
            for arm, d in (("adapter", adapter), ("base", base)):
                if d is None:
                    continue
                s = summary(d)
                print(
                    f"| {name} | {group} | {arm} | {s['macro']:.3f} | {s['full']}/{s['attempts']} | "
                    f"{s['turns']:.1f} | {s['limit']} | {s['stopped']} | {s['overflow']} | {s['peak']:,.0f} | {s['draft']:.3f} |"
                )

    print()
    print(
        "| Model | Scenarios | Adapter minus base | 95% CI | Higher / equal / lower |"
    )
    print("| --- | --- | ---: | --- | --- |")
    for name, by_group in data.items():
        for group, (adapter, base) in by_group.items():
            if base is None:
                continue
            pool = scenario_differences(adapter, base)
            draws = [
                statistics.mean(rng.choice(pool) for _ in pool)
                for _ in range(resamples)
            ]
            lo, hi = ci(draws)
            counts = f"{sum(d > 0 for d in pool)} / {sum(d == 0 for d in pool)} / {sum(d < 0 for d in pool)}"
            print(
                f"| {name} | {group} | {statistics.mean(pool):+.3f} | [{lo:+.3f}, {hi:+.3f}] | {counts} |"
            )

    print()
    print("Score per scenario (one attempt; base → adapter where a base sweep exists):")
    print()
    print("| Scenario | " + " | ".join(data) + " |")
    print("| --- | " + " | ".join("---:" for _ in data) + " |")
    for s in trained + VALIDATION:
        cells = []
        for by_group in data.values():
            adapter, base = by_group["all 46"]
            a = scenario_means(adapter)[s]
            cells.append(
                f"{scenario_means(base)[s]:.2f} → {a:.2f}" if base else f"{a:.2f}"
            )
        mark = (
            " (validation)"
            if s in VALIDATION
            else " (not in Qwen-teacher data)"
            if s == UNTRAINED_BY_QWEN_TEACHER
            else ""
        )
        print(f"| {s}{mark} | " + " | ".join(cells) + " |")


if __name__ == "__main__":
    main()
