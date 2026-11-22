"""The continuation dataset: a teacher continues a stopped Gemma attempt.

A Gemma attempt that stopped on its own although the grader scored it below
full marks has a prefix: the conversation up to its premature final message.
A teacher continues from that prefix in a fresh cluster (``make benchmark
CONTINUE_FROM=<record>``); the loss is computed only on the teacher's turns.
This module lists the attempts that qualify, builds the dataset from the
continuation runs and freezes its renders.

    uv run python -m finetune.data.continuations
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
from finetune.data import dataset, render, review, rows, select, trajectories
from finetune.data.models import Entry, Row

log = logging.getLogger(__name__)

SOURCES = ROOT / "docs" / "research" / "kaggle-development-suite" / "raw"
OUT = dataset.DATASETS / "continuation" / "candidates.yaml"


def directory(teacher: str) -> Path:
    """One teacher's continuation dataset: the runs, the manifest, the review
    sheets and, once frozen, the rendered examples."""
    return dataset.DATASETS / "continuation" / teacher


@dataclass(frozen=True)
class Prefix:
    source: str  # recorded Gemma attempt, relative to the repository
    model: str
    scenario: str
    prefix_messages: int  # messages the teacher inherits, up to the premature final message
    score: float


def prefix(path: Path, allowed: set[str]) -> Prefix | None:
    """The attempt as a continuation prefix, or None when it does not qualify: a training
    scenario, ended by the agent itself on a final message, not a full score."""
    data = json.loads(path.read_text())
    agent = data["agent"]
    messages = agent["messages"]
    last = messages[-1]
    if data["scenario_id"] not in allowed or data["grading"]["full_success"]:
        return None
    if agent["termination"] != "completed" or last["role"] != "assistant" or last.get("tool_calls"):
        return None
    return Prefix(
        str(path.resolve().relative_to(ROOT)),
        agent["inference"]["model"],
        data["scenario_id"],
        len(messages) - 1,
        data["grading"]["score"],
    )


def candidates() -> list[Prefix]:
    allowed = dataset.training_scenarios()
    found = []
    for path in sorted(SOURCES.glob("gemma-*/**/prompt/*.json")):
        item = prefix(path, allowed)
        if item:
            found.append(item)
    return found


def sheet(entry: Entry) -> str:
    """A review sheet for a continuation: the task, the stop it inherited, the
    teacher's own write commands and its final report."""
    data = json.loads((ROOT / entry.path).read_text())
    agent = data["agent"]
    messages = agent["messages"]
    prefix = agent.get("prefix_messages")
    if not prefix:
        raise ValueError(f"{entry.path} is not a continuation record")
    task = next(m["content"] for m in messages if m["role"] == "user")
    stop = messages[prefix - 1]
    inherited = select.write_commands(messages[:prefix])
    own = select.write_commands(messages[prefix:])
    front = yaml.safe_dump(
        {"attempt": entry.path, "verdict": review.PENDING, "reason": ""}, sort_keys=False
    )

    def listed(commands: tuple[str, ...]) -> str:
        return "\n".join(f"{i}. `{w}`" for i, w in enumerate(commands, 1)) or "(none)"

    return (
        f"---\n{front}---\n"
        f"# {entry.scenario} · {entry.teacher} · {entry.run} · attempt {entry.attempt}\n\n"
        f"## Task\n\n{task}\n\n"
        f"Documented in:\n{review.sources(entry.scenario) or '(no sources)'}\n\n"
        f"## Inherited stop\n\n"
        f"The first {prefix - 1} messages come from the Gemma attempt; it stopped on:\n\n"
        f"[{stop['role']}] {(stop.get('content') or '').strip()}\n\n"
        f"Its writes:\n\n{listed(inherited)}\n\n"
        f"## Teacher's write commands\n\n{listed(own)}\n\n"
        f"## Final report\n\n{messages[-1]['content']}\n"
    )


def build(teacher: str, run_dirs: list[Path]) -> Path:
    """Copy the continuation runs, keep every clean attempt and write the
    manifest with a review sheet per kept attempt."""
    out = directory(teacher)
    allowed = dataset.training_scenarios()
    dataset.copy_runs(run_dirs, allowed, out)
    clean = []
    for path in select.attempt_files(run_dirs):
        found = select.candidate(path)
        if found and found.scenario_id in allowed:
            clean.append(found)
    dropped = dataset.rejected(out)
    clean = [c for c in clean if select.relative(dataset.raw_target(c, out)) not in dropped]
    kept = dataset.entries(clean, out)
    review.write_missing(out / "review", kept, sheet)
    prefixes = {c["scenario"] for c in yaml.safe_load(OUT.read_text())}
    missing = sorted(prefixes - {e.scenario for e in kept})
    manifest = {
        "teacher": teacher,
        "train": [e.path for e in kept],
        "scenarios_without_examples": missing,
        "attempts": [asdict(e) for e in kept],
    }
    (out / "manifest.yaml").write_text(yaml.safe_dump(manifest, sort_keys=False, width=200))
    log.info(
        "%s: %d continuations kept, %d scenarios without examples",
        teacher,
        len(kept),
        len(missing),
    )
    return out


@dataclass(frozen=True)
class Rendered:
    student: str
    examples: int
    over_length: list[str]


@dataclass(frozen=True)
class Freeze:
    teacher: str
    attempts: int
    revision: str  # sha256 of the manifest, the candidates and the rendered files
    rendered: list[Rendered]


def masked_row(tok, student, path: str) -> Row:
    """The continuation rendered in full, with the inherited prefix masked: the
    loss is computed on the teacher's turns only."""
    data = json.loads((ROOT / path).read_text())
    prefix = data["agent"].get("prefix_messages")
    if not prefix:
        raise ValueError(f"{path} is not a continuation record")
    traj = trajectories.load(ROOT / path)
    text = render.render(tok, traj)
    context = render.render(tok, dataclasses.replace(traj, messages=traj.messages[:prefix]))
    spans = rows.mask_context(text, context, rows.target_spans(student.family, text))
    if not spans:
        raise ValueError(f"nothing to learn in {path}: the prefix covers the whole render")
    return Row(path, traj.scenario_id, text, spans)


def digest(files: list[Path]) -> str:
    return hashlib.sha256(b"".join(p.read_bytes() for p in files)).hexdigest()


def freeze(teacher: str, tokenizers: dict | None = None) -> Freeze:
    out = directory(teacher)
    manifest = yaml.safe_load((out / "manifest.yaml").read_text())
    waiting = [a["path"] for a in manifest["attempts"] if a["review"] == review.PENDING]
    if waiting:
        raise ValueError(f"{len(waiting)} continuations still wait for review, e.g. {waiting[0]}")
    rejected = [a["path"] for a in manifest["attempts"] if a["review"] == review.REJECTED]
    if rejected:
        raise ValueError(f"rejected continuations are in the manifest: build again: {rejected}")
    tokenizers = tokenizers or {n: render.tokenizer(s) for n, s in CONFIG.students.items()}
    folder = out / "rendered"
    folder.mkdir(exist_ok=True)
    rendered = []
    for name, student in CONFIG.students.items():
        tok = tokenizers[name]
        kept, over = 0, []
        with (folder / f"{name}.jsonl").open("w") as f:
            for path in manifest["train"]:
                row = masked_row(tok, student, path)
                if (
                    len(tok(row.text, add_special_tokens=False)["input_ids"])
                    > CONFIG.training.max_len
                ):
                    over.append(path)
                    continue
                f.write(row.to_json() + "\n")
                kept += 1
        log.info(
            "continuation %s: %d examples, %d over %d tokens",
            name,
            kept,
            len(over),
            CONFIG.training.max_len,
        )
        rendered.append(Rendered(name, kept, over))
    files = [out / "manifest.yaml", OUT, *sorted(folder.glob("*.jsonl"))]
    result = Freeze(teacher, len(manifest["train"]), digest(files), rendered)
    (out / "freeze.json").write_text(json.dumps(asdict(result), indent=2) + "\n")
    log.info("%s continuation frozen: revision %s", teacher, result.revision)
    return result


def main() -> None:
    logs.setup()
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--teacher", help="build or freeze this teacher's continuations")
    parser.add_argument("--runs", nargs="+", type=Path, help="the continuation runs to build from")
    parser.add_argument("--freeze", action="store_true", help="render and freeze the built dataset")
    args = parser.parse_args()
    if args.freeze and args.teacher:
        freeze(args.teacher)
        return
    if args.teacher and args.runs:
        build(args.teacher, args.runs)
        return
    found = candidates()
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(yaml.safe_dump([asdict(c) for c in found], sort_keys=False, width=200))
    log.info(
        "%d candidates in %d scenarios, written to %s",
        len(found),
        len({c.scenario for c in found}),
        OUT,
    )


if __name__ == "__main__":
    main()
