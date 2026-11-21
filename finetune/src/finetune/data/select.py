"""Selection of training attempts for one teacher (D-054, #25).

An attempt is kept when it has a full score, ended by the agent itself, did not
overflow the context and made no malformed tool call. Two attempts are the same
trajectory when their sequences of write commands are equal after whitespace
normalisation; the one with fewer tokens stays. Up to ``cap`` attempts per
scenario are kept, fewest tokens first.
"""

from __future__ import annotations

import json
import logging
import re
from collections import defaultdict
from pathlib import Path

from finetune.config import ROOT
from finetune.data.models import Candidate

log = logging.getLogger(__name__)

# Commands that change the cluster; reads (get, describe, logs, ...) are not writes.
KUBECTL_WRITE = re.compile(
    r"\bkubectl\b[^|;&\n]*?\s(?:apply|create|patch|delete|replace|edit|scale|set|label|"
    r"annotate|taint|cordon|uncordon|drain|expose|run|autoscale|"
    r"rollout\s+(?:undo|restart))\b[^|;&\n]*"
)
FILE_WRITE_TOOLS = {"write_file", "edit_file"}


def _normalise(text: str) -> str:
    return " ".join(text.split())


def write_commands(messages: list[dict]) -> tuple[str, ...]:
    """The attempt's write commands in order: kubectl writes found in shell
    commands, and every file write or edit."""
    writes: list[str] = []
    for message in messages:
        for call in message.get("tool_calls") or []:
            arguments = json.loads(call["arguments"])
            if call["name"] == "bash":
                writes += [
                    _normalise(m.group(0)) for m in KUBECTL_WRITE.finditer(arguments["command"])
                ]
            elif call["name"] in FILE_WRITE_TOOLS:
                writes.append(_normalise(f"{call['name']} {json.dumps(arguments, sort_keys=True)}"))
    return tuple(writes)


def _malformed(agent: dict) -> bool:
    for message in agent["messages"]:
        for call in message.get("tool_calls") or []:
            try:
                json.loads(call["arguments"])
            except ValueError:
                return True
    return any(c.get("error", "").startswith("malformed") for c in agent["tool_calls"])


def candidate(path: Path) -> Candidate | None:
    """The attempt as a candidate, or None when it fails the D-054 filter."""
    data = json.loads(path.read_text())
    agent = data.get("agent") or {}
    if not data["grading"]["full_success"] or agent.get("termination") != "completed":
        return None
    if agent.get("context_overflow") or _malformed(agent):
        return None
    return Candidate(
        path=path.resolve(),
        run=next(p for p in path.parts if p.startswith("run-")),
        scenario_id=data["scenario_id"],
        attempt=data["attempt"],
        teacher=agent["inference"]["model"],
        tokens=agent["token_usage"]["peak_context_tokens"],
        writes=write_commands(agent["messages"]),
    )


def select(candidates: list[Candidate], cap: int) -> list[Candidate]:
    """Up to ``cap`` distinct trajectories per scenario, fewest tokens first."""
    by_scenario: dict[str, list[Candidate]] = defaultdict(list)
    for c in sorted(candidates, key=lambda c: (c.tokens, c.run, c.attempt)):
        by_scenario[c.scenario_id].append(c)
    kept: list[Candidate] = []
    for scenario, ranked in sorted(by_scenario.items()):
        seen: set[tuple[str, ...]] = set()
        chosen = []
        for c in ranked:
            if c.writes not in seen and len(chosen) < cap:
                seen.add(c.writes)
                chosen.append(c)
        log.info(
            "%s: %d clean, %d distinct, %d kept", scenario, len(ranked), len(seen), len(chosen)
        )
        kept += chosen
    return kept


def attempt_files(run_dirs: list[Path], agent: str = "prompt") -> list[Path]:
    return sorted(p for d in run_dirs for p in d.glob(f"*/{agent}/*.json"))


def relative(path: Path) -> str:
    return str(path.resolve().relative_to(ROOT))
