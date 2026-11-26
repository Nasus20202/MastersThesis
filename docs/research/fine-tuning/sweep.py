#!/usr/bin/env python3
"""First-sweep results of the fine-tuning study: all 46 development scenarios,
one attempt each, split into the 38 trained-on and the 8 validation scenarios.

Uses the helpers of pooled.py. Prints per model and scenario group the
adapter's (and, where a base sweep exists, the base's) summary, the paired
bootstrap of adapter minus base over scenarios (10,000 draws, seed 1), and
the per-scenario scores.

    python3 sweep.py [--seed 1] [--resamples 10000]
"""

import argparse
import random
import statistics

from pooled import RAW, VALIDATION, attempts, ci, scenario_differences, scenario_means, summary

# model -> (adapter run, base run or None).
MODELS = {
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


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--seed", type=int, default=1)
    parser.add_argument("--resamples", type=int, default=10000)
    args = parser.parse_args()
    rng = random.Random(args.seed)

    every = scenarios_of(next(iter(MODELS.values()))[0])
    trained = [s for s in every if s not in VALIDATION]
    groups = {"all 46": every, "38 trained-on": trained, "8 validation": VALIDATION}
    data = {
        name: {
            group: (attempts([a], scenarios=scen), attempts([b], scenarios=scen) if b else None)
            for group, scen in groups.items()
        }
        for name, (a, b) in MODELS.items()
    }

    print("| Model | Scenarios | Arm | Macro score | Full success | Mean turns | Turn limit | Stopped without full success | Context or server error | Mean peak context | Draft acceptance |")
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
    print("| Model | Scenarios | Adapter minus base | 95% CI | Higher / equal / lower |")
    print("| --- | --- | ---: | --- | --- |")
    for name, by_group in data.items():
        for group, (adapter, base) in by_group.items():
            if base is None:
                continue
            pool = scenario_differences(adapter, base)
            draws = [statistics.mean(rng.choice(pool) for _ in pool) for _ in range(args.resamples)]
            lo, hi = ci(draws)
            counts = f"{sum(d > 0 for d in pool)} / {sum(d == 0 for d in pool)} / {sum(d < 0 for d in pool)}"
            print(f"| {name} | {group} | {statistics.mean(pool):+.3f} | [{lo:+.3f}, {hi:+.3f}] | {counts} |")

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
            cells.append(f"{scenario_means(base)[s]:.2f} → {a:.2f}" if base else f"{a:.2f}")
        mark = " (validation)" if s in VALIDATION else " (not in Qwen-teacher data)" if s == UNTRAINED_BY_QWEN_TEACHER else ""
        print(f"| {s}{mark} | " + " | ".join(cells) + " |")


if __name__ == "__main__":
    main()
