"""The truncation dataset: each kept trajectory cut after its last write.

The cut follows the last write command and the first successful read-only
observation after it; no write comes after the cut, so the cluster state and
the score stay the same. The cut conversation ends with the source
trajectory's own final report, reused unchanged, which may mention checks the
cut removed. The cut conversations are rendered for each student and committed
with the cut points.

    uv run python -m finetune.data.truncation --name deepseek-flash
"""

from __future__ import annotations

import argparse
import dataclasses
import hashlib
import json
import logging
from dataclasses import asdict, dataclass
from pathlib import Path

import yaml

from finetune import logs
from finetune.config import CONFIG, ROOT
from finetune.data import dataset, render, rows, select, trajectories
from finetune.data.models import Trajectory

log = logging.getLogger(__name__)

SUCCESS = "exit_code: 0"


@dataclass(frozen=True)
class Cut:
    source: str  # raw attempt record, relative to the repository
    scenario: str
    keep_messages: int  # messages kept from the start of the conversation
    total_messages: int
    writes: int  # write commands, equal before and after the cut


def _is_write(call: dict) -> bool:
    return bool(select.write_commands([{"tool_calls": [call]}]))


def cut_point(messages: list[dict]) -> int | None:
    """Number of messages to keep, or None when no successful read-only
    observation follows the last write."""
    turns = [i for i, m in enumerate(messages) if m["role"] == "assistant" and m.get("tool_calls")]
    written = [i for i in turns if any(_is_write(c) for c in messages[i]["tool_calls"])]
    if not written:
        return None
    for turn in (t for t in turns if t > written[-1]):
        results = {}
        end = turn + 1
        while end < len(messages) and messages[end]["role"] == "tool":
            results[messages[end]["tool_call_id"]] = messages[end]["content"]
            end += 1
        if any(
            not _is_write(c) and results.get(c["id"], "").startswith(SUCCESS)
            for c in messages[turn]["tool_calls"]
        ):
            return end
    return None


def cuts(name: str) -> list[Cut]:
    manifest = yaml.safe_load((dataset.DATASETS / name / "manifest.yaml").read_text())
    found = []
    for entry in manifest["attempts"]:
        messages = json.loads((ROOT / entry["path"]).read_text())["agent"]["messages"]
        keep = cut_point(messages)
        if keep is None:
            log.info("no cut point: %s", entry["path"])
            continue
        before, after = select.write_commands(messages), select.write_commands(messages[:keep])
        if before != after:  # a write after the cut would change the cluster state
            raise ValueError(f"cut changes the write commands: {entry['path']}")
        found.append(Cut(entry["path"], entry["scenario"], keep, len(messages), len(before)))
    return found


def cut_trajectory(path: Path) -> Trajectory | None:
    """The cut conversation followed by the trajectory's original final report.

    The final report is reused as recorded; only the messages between the cut
    and the report are dropped.
    """
    raw = json.loads(path.read_text())["agent"]["messages"]
    keep = cut_point(raw)
    if keep is None:
        return None
    traj = trajectories.load(path)
    if len(traj.messages) != len(raw):
        raise ValueError(f"messages changed between parsing: {path}")
    final = traj.messages[-1]
    if final["role"] != "assistant" or final.get("tool_calls"):
        raise ValueError(f"trajectory does not end with a final report: {path}")
    return dataclasses.replace(traj, messages=[*traj.messages[:keep], final])


@dataclass(frozen=True)
class Rendered:
    student: str
    examples: int
    over_length: list[str]


@dataclass(frozen=True)
class Truncation:
    teacher: str
    source_revision: str  # revision of the frozen dataset the cuts are made from
    attempts: int  # kept attempts in the manifest
    with_cut: int  # attempts that have a verified cut
    revision: str  # sha256 of the cuts and the rendered files
    files: list[str]  # what the revision covers, relative to the repository
    rendered: list[Rendered]


def render_examples(directory: Path, source: Path, paths: list[str]) -> Truncation:
    out = directory / "rendered"
    out.mkdir(parents=True, exist_ok=True)
    loaded = [(p, cut_trajectory(ROOT / p)) for p in paths]
    with_cut = [t for _, t in loaded if t is not None]
    rendered = []
    for name, student in CONFIG.students.items():
        tok = render.tokenizer(student)
        kept, over = 0, []
        with (out / f"{name}.jsonl").open("w") as f:
            for path, traj in loaded:
                if traj is None:
                    continue
                row = rows.example(tok, student, traj)
                if (
                    len(tok(row.text, add_special_tokens=False)["input_ids"])
                    > CONFIG.training.max_len
                ):
                    over.append(path)
                    continue
                f.write(rows.replace_source(row, path).to_json() + "\n")
                kept += 1
        log.info(
            "truncation %s: %d examples, %d over %d tokens",
            name,
            kept,
            len(over),
            CONFIG.training.max_len,
        )
        rendered.append(Rendered(name, kept, over))
    files = [directory / "cuts.yaml", *sorted(out.glob("*.jsonl"))]
    digest = hashlib.sha256(b"".join(p.read_bytes() for p in files)).hexdigest()
    source_revision = json.loads((source / "freeze.json").read_text())["revision"]
    result = Truncation(
        directory.name,
        source_revision,
        len(paths),
        len(with_cut),
        digest,
        [str(p.relative_to(ROOT)) for p in files],
        rendered,
    )
    (directory / "freeze.json").write_text(json.dumps(asdict(result), indent=2) + "\n")
    return result


def main() -> None:
    logs.setup()
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--name", required=True, help="the teacher dataset to cut")
    args = parser.parse_args()
    source = dataset.DATASETS / args.name
    found = cuts(args.name)
    directory = dataset.DATASETS / "truncation" / args.name
    directory.mkdir(parents=True, exist_ok=True)
    out = directory / "cuts.yaml"
    out.write_text(yaml.safe_dump([asdict(c) for c in found], sort_keys=False, width=200))
    removed = sum(c.total_messages - c.keep_messages for c in found)
    log.info("%d cuts, %d messages removed, written to %s", len(found), removed, out)
    manifest = yaml.safe_load((source / "manifest.yaml").read_text())
    result = render_examples(directory, source, manifest["train"])
    log.info(
        "truncation: %d of %d attempts, revision %s",
        result.with_cut,
        result.attempts,
        result.revision,
    )


if __name__ == "__main__":
    main()
