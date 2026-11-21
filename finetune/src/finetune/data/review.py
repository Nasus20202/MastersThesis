"""Review sheets: one markdown file per trajectory the researcher reviews, with
the scenario, its write commands and the final report (#25).

The researcher sets ``verdict`` (accepted | rejected) and ``reason`` in the
front matter. A rejected attempt is dropped when the dataset is built again and
the next candidate of its scenario takes its place, to be reviewed in turn.
"""

from __future__ import annotations

import json
import re
from dataclasses import dataclass
from pathlib import Path

import yaml

from finetune.config import ROOT
from finetune.data import select
from finetune.data.models import Entry

SCENARIOS = ROOT / "benchmark" / "scenarios"
PENDING, ACCEPTED, REJECTED = "pending", "accepted", "rejected"
FRONT = re.compile(r"\A---\n(.*?)\n---\n", re.DOTALL)


@dataclass(frozen=True)
class Verdict:
    attempt: str  # raw attempt record, relative to the repository
    verdict: str
    reason: str


def sheet_name(entry: Entry) -> str:
    return f"{entry.scenario}-{entry.run}-{entry.attempt:03d}.md"


def sources(scenario: str) -> str:
    """The documentation sections the scenario's repair is documented in."""
    path = next(SCENARIOS.glob(f"*/{scenario}/scenario.yaml"))
    listed = yaml.safe_load(path.read_text()).get("sources") or []
    return "\n".join(f"- {s['path']}: {', '.join(s.get('sections') or [])}" for s in listed)


def render(entry: Entry) -> str:
    data = json.loads((ROOT / entry.path).read_text())
    messages = data["agent"]["messages"]
    task = next(m["content"] for m in messages if m["role"] == "user")
    commands = "\n".join(f"{i}. `{w}`" for i, w in enumerate(select.write_commands(messages), 1))
    final = messages[-1]["content"]
    front = yaml.safe_dump(
        {"attempt": entry.path, "verdict": PENDING, "reason": ""}, sort_keys=False
    )
    return (
        f"---\n{front}---\n"
        f"# {entry.scenario} · {entry.teacher} · {entry.run} · attempt {entry.attempt}\n\n"
        f"## Task\n\n{task}\n\n"
        f"Documented in:\n{sources(entry.scenario) or '(no sources)'}\n\n"
        f"## Write commands\n\n{commands or '(none)'}\n\n"
        f"## Final report\n\n{final}\n"
    )


def read(sheet: Path) -> Verdict:
    front = yaml.safe_load(FRONT.match(sheet.read_text()).group(1))
    return Verdict(front["attempt"], front["verdict"], front.get("reason") or "")


def verdicts(directory: Path) -> dict[str, Verdict]:
    return {v.attempt: v for v in map(read, sorted(directory.glob("*.md")))}


def write_missing(directory: Path, entries: list[Entry]) -> None:
    """A sheet for every entry that needs review and has none yet; a pending
    sheet of an attempt that is no longer kept is removed."""
    directory.mkdir(parents=True, exist_ok=True)
    kept = {e.path for e in entries}
    for sheet in directory.glob("*.md"):
        found = read(sheet)
        if found.verdict == PENDING and found.attempt not in kept:
            sheet.unlink()
    for entry in entries:
        sheet = directory / sheet_name(entry)
        if entry.review == PENDING and not sheet.exists():
            sheet.write_text(render(entry))
