"""Training text from a trajectory, and the check that it equals serving text.

Training renders a whole conversation once with the student's Hugging Face chat
template. That is only valid if, for every assistant turn, the prompt
llama.cpp built when serving that turn is a prefix of the full render: then the
model is trained on exactly the context it saw. ``compare`` checks this
against a running llama.cpp server (``/apply-template``).
"""

from __future__ import annotations

import argparse
import json
import sys
import urllib.request
from dataclasses import asdict, dataclass
from pathlib import Path

from transformers import AutoTokenizer

from finetune import trajectories
from finetune.students import STUDENTS, Student


def tokenizer(student: Student):
    return AutoTokenizer.from_pretrained(student.checkpoint, revision=student.revision)


def bos_text(student: Student) -> str:
    """BOS as the chat template writes it; llama.cpp adds it as a token instead."""
    return "<bos>" if student.family == "gemma" else ""


def render(tok, traj: trajectories.Trajectory, upto: int | None = None) -> str:
    """Full conversation, or the generation prompt before message ``upto``.

    Reasoning is on, as llama.cpp serves both families.
    """
    messages = traj.messages if upto is None else traj.messages[:upto]
    return tok.apply_chat_template(
        messages,
        tools=traj.tools,
        tokenize=False,
        add_generation_prompt=upto is not None,
        enable_thinking=True,
    )


def served_prompt(base_url: str, model: str, traj: trajectories.Trajectory, upto: int) -> str:
    body = json.dumps(
        {"model": model, "messages": traj.messages[:upto], "tools": traj.tools}
    ).encode()
    request = urllib.request.Request(
        f"{base_url}/apply-template", body, {"Content-Type": "application/json"}
    )
    with urllib.request.urlopen(request, timeout=600) as response:
        return json.load(response)["prompt"]


@dataclass
class TurnCheck:
    turn: int
    served_equals_hf_prompt: bool
    served_is_prefix_of_full: bool
    first_difference: str


def _difference(a: str, b: str) -> str:
    i = next((i for i, (x, y) in enumerate(zip(a, b, strict=False)) if x != y), min(len(a), len(b)))
    return json.dumps(
        {"at": i, "served": a[max(0, i - 80) : i + 80], "hf": b[max(0, i - 80) : i + 80]}
    )


def compare(base_url: str, student: Student, traj: trajectories.Trajectory) -> list[TurnCheck]:
    tok = tokenizer(student)
    # The one expected difference, removed before comparing.
    bos = bos_text(student)
    full = render(tok, traj).removeprefix(bos)
    checks = []
    for i, message in enumerate(traj.messages):
        if message["role"] != "assistant":
            continue
        served = served_prompt(base_url, student.served_model, traj, i)
        hf = render(tok, traj, upto=i).removeprefix(bos)
        checks.append(
            TurnCheck(
                turn=i,
                served_equals_hf_prompt=served == hf,
                served_is_prefix_of_full=full.startswith(served),
                first_difference="" if served == hf else _difference(served, hf),
            )
        )
    return checks


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--student", choices=sorted(STUDENTS), required=True)
    parser.add_argument("--base-url", default="http://127.0.0.1:8090")
    parser.add_argument("attempts", nargs="+", type=Path)
    args = parser.parse_args()
    student = STUDENTS[args.student]
    for path in args.attempts:
        traj = trajectories.load(path)
        checks = compare(args.base_url, student, traj)
        json.dump(
            {
                "student": args.student,
                "attempt": str(path),
                "turns": len(checks),
                "served_equals_hf_prompt": sum(c.served_equals_hf_prompt for c in checks),
                "served_is_prefix_of_full": sum(c.served_is_prefix_of_full for c in checks),
                "checks": [asdict(c) for c in checks],
            },
            sys.stdout,
        )
        sys.stdout.write("\n")


if __name__ == "__main__":
    main()
