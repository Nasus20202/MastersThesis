"""Check that training text equals serving text.

Training renders a whole conversation once with the checkpoint's chat
template. That is only valid if, for every assistant turn, the prompt
llama.cpp built when serving that turn is a prefix of the full render: then
the model is trained on exactly the context it saw. ``compare`` checks this
against a running llama.cpp server (``/apply-template``).
"""

from __future__ import annotations

import argparse
import json
import logging
import urllib.request
from pathlib import Path

from finetune import logs
from finetune.config import STUDENTS, Student
from finetune.data import render, trajectories
from finetune.data.models import Trajectory
from finetune.serving.models import AttemptCheck, Difference, TurnCheck

log = logging.getLogger(__name__)


def served_prompt(base_url: str, model: str, traj: Trajectory, upto: int) -> str:
    body = json.dumps(
        {"model": model, "messages": traj.messages[:upto], "tools": traj.tools}
    ).encode()
    request = urllib.request.Request(
        f"{base_url}/apply-template", body, {"Content-Type": "application/json"}
    )
    with urllib.request.urlopen(request, timeout=600) as response:
        return json.load(response)["prompt"]


def _difference(a: str, b: str) -> Difference:
    i = next((i for i, (x, y) in enumerate(zip(a, b, strict=False)) if x != y), min(len(a), len(b)))
    return Difference(i, a[max(0, i - 80) : i + 80], b[max(0, i - 80) : i + 80])


def compare(base_url: str, student: Student, traj: Trajectory) -> list[TurnCheck]:
    tok = render.tokenizer(student)
    # The one expected difference, removed before comparing.
    bos = render.bos_text(student)
    full = render.render(tok, traj).removeprefix(bos)
    checks = []
    for i, message in enumerate(traj.messages):
        if message["role"] != "assistant":
            continue
        served = served_prompt(base_url, student.served_model, traj, i)
        hf = render.render(tok, traj, upto=i).removeprefix(bos)
        checks.append(
            TurnCheck(
                turn=i,
                served_equals_hf_prompt=served == hf,
                served_is_prefix_of_full=full.startswith(served),
                first_difference=None if served == hf else _difference(served, hf),
            )
        )
    return checks


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--student", choices=sorted(STUDENTS), required=True)
    parser.add_argument("--base-url", default="http://127.0.0.1:8090")
    parser.add_argument("--out", type=Path, required=True, help="JSONL, one line per attempt")
    parser.add_argument("attempts", nargs="+", type=Path)
    args = parser.parse_args()
    logs.setup()
    student = STUDENTS[args.student]
    with args.out.open("w") as out:
        for path in args.attempts:
            traj = trajectories.load(path)
            check = AttemptCheck(args.student, str(path), compare(args.base_url, student, traj))
            out.write(check.to_json() + "\n")
            log.info(
                "%s: %d turns, %d equal to the HF render, %d prefixes of the full render",
                path,
                len(check.checks),
                check.served_equals_hf_prompt,
                check.served_is_prefix_of_full,
            )


if __name__ == "__main__":
    main()
