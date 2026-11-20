"""An adapted model next to its base in the llama.cpp router (#24 steps 5 and 6).

The router preset gets one extra section per adapted model: the same 4-bit
GGUF file plus ``lora``, under its own model name, so the adapter is always on
for that name and never switched per request (requests with different adapters
are not batched together). ``compare`` sends the same fixed prompts to the
base and the adapted name, greedy, and records llama.cpp's own timings:
generation speed and multi-token-prediction draft acceptance.
"""

from __future__ import annotations

import argparse
import json
import urllib.request
from pathlib import Path

from finetune import trajectories

MAX_TOKENS = 256


def preset_section(name: str, gguf: str, adapter: str, sampling: dict[str, str]) -> str:
    lines = [f"[{name}]", f"model = {gguf}", f"lora = {adapter}"]
    lines += [f"{key} = {value}" for key, value in sampling.items()]
    return "\n".join(lines) + "\n"


def prompts(paths: list[Path]) -> list[dict]:
    """The first agent turn of each attempt: system prompt, task and tools."""
    out = []
    for path in paths:
        traj = trajectories.load(path)
        out.append({"source": str(path), "messages": traj.messages[:2], "tools": traj.tools})
    return out


def complete(base_url: str, model: str, prompt: dict) -> dict:
    body = json.dumps(
        {
            "model": model,
            "messages": prompt["messages"],
            "tools": prompt["tools"],
            "temperature": 0,
            "max_tokens": MAX_TOKENS,
        }
    ).encode()
    request = urllib.request.Request(
        f"{base_url}/v1/chat/completions", body, {"Content-Type": "application/json"}
    )
    with urllib.request.urlopen(request, timeout=900) as response:
        data = json.load(response)
    timings = data.get("timings", {})
    message = data["choices"][0]["message"]
    return {
        "text": (message.get("reasoning_content") or "") + (message.get("content") or ""),
        "tool_calls": message.get("tool_calls"),
        "predicted_n": timings.get("predicted_n"),
        "predicted_per_second": timings.get("predicted_per_second"),
        "draft_n": timings.get("draft_n"),
        "draft_n_accepted": timings.get("draft_n_accepted"),
    }


def summary(runs: list[dict]) -> dict:
    drafted = sum(r["draft_n"] or 0 for r in runs)
    accepted = sum(r["draft_n_accepted"] or 0 for r in runs)
    speeds = [r["predicted_per_second"] for r in runs if r["predicted_per_second"]]
    return {
        "prompts": len(runs),
        "draft_acceptance": accepted / drafted if drafted else None,
        "mean_tokens_per_second": sum(speeds) / len(speeds) if speeds else None,
    }


def compare(base_url: str, base: str, adapted: str, items: list[dict]) -> dict:
    result = {"base": base, "adapted": adapted, "runs": []}
    for item in items:
        a = complete(base_url, base, item)
        b = complete(base_url, adapted, item)
        result["runs"].append(
            {"source": item["source"], "base": a, "adapted": b, "same_text": a["text"] == b["text"]}
        )
    result["base_summary"] = summary([r["base"] for r in result["runs"]])
    result["adapted_summary"] = summary([r["adapted"] for r in result["runs"]])
    result["same_text"] = sum(r["same_text"] for r in result["runs"])
    return result


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default="http://127.0.0.1:8091")
    parser.add_argument("--base", required=True, help="router name of the base model")
    parser.add_argument("--adapted", required=True, help="router name of the adapted model")
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("attempts", nargs="+", type=Path)
    args = parser.parse_args()
    result = compare(args.base_url, args.base, args.adapted, prompts(args.attempts))
    args.out.write_text(json.dumps(result, indent=2))
    print(json.dumps({k: v for k, v in result.items() if k != "runs"}, indent=2))


if __name__ == "__main__":
    main()
