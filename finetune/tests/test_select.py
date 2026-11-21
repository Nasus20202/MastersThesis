import json

from finetune.data import select


def attempt(scenario="pvc-pending", commands=("kubectl apply -f a.yaml",), tokens=1000, **over):
    calls = [
        {"id": f"c{i}", "type": "function", "name": "bash", "arguments": json.dumps({"command": c})}
        for i, c in enumerate(commands)
    ]
    data = {
        "scenario_id": scenario,
        "attempt": 1,
        "agent": {
            "inference": {"model": "teacher"},
            "token_usage": {"peak_context_tokens": tokens},
            "termination": "completed",
            "tool_calls": [],
            "messages": [
                {"role": "user", "content": "task"},
                {"role": "assistant", "content": "", "tool_calls": calls},
                {"role": "assistant", "content": "done"},
            ],
        },
        "grading": {"full_success": True},
    }
    data["agent"].update(over.pop("agent", {}))
    data.update(over)
    return data


def write(tmp_path, name, data, run="run-1"):
    path = tmp_path / run / data["scenario_id"] / "prompt" / f"{name}.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data))
    for meta in ("run.json", "results.json"):
        (tmp_path / run / meta).write_text("{}")
    return path


def test_filter_keeps_only_full_score_completed_without_overflow_or_malformed_calls(tmp_path):
    kept = write(tmp_path, "001", attempt())
    assert select.candidate(kept) is not None
    cases = {
        "score": attempt(grading={"full_success": False}),
        "ended": attempt(agent={"termination": "turn_limit"}),
        "overflow": attempt(agent={"context_overflow": True}),
        "cut": attempt(commands=('{"command": "kubectl apply',)),
        "reported": attempt(agent={"tool_calls": [{"error": "malformed bash arguments: eof"}]}),
    }
    cases["cut"]["agent"]["messages"][1]["tool_calls"][0]["arguments"] = '{"command": "kubectl ap'
    for name, data in cases.items():
        assert select.candidate(write(tmp_path, name, data)) is None, name


def test_write_commands_ignore_reads_and_whitespace():
    messages = [
        {
            "tool_calls": [
                {
                    "name": "bash",
                    "arguments": json.dumps(
                        {
                            "command": (
                                "kubectl get pods -A | head;\n"
                                "kubectl  -n  x   patch deploy app -p '{}' && kubectl logs app"
                            )
                        }
                    ),
                },
                {
                    "name": "edit_file",
                    "arguments": json.dumps({"path": "a", "old_string": "x", "new_string": "y"}),
                },
            ]
        }
    ]
    writes = select.write_commands(messages)
    assert writes[0] == "kubectl -n x patch deploy app -p '{}'"
    assert len(writes) == 2 and writes[1].startswith("edit_file ")


def test_select_drops_duplicate_trajectories_and_caps_per_scenario_by_tokens(tmp_path):
    files = [
        write(tmp_path, "001", attempt(commands=("kubectl apply -f a",), tokens=900)),
        write(
            tmp_path, "002", attempt(commands=("kubectl  apply -f   a",), tokens=500)
        ),  # same writes, fewer tokens
        write(tmp_path, "003", attempt(commands=("kubectl delete pod p",), tokens=700)),
        write(
            tmp_path, "004", attempt(commands=("kubectl scale deploy d --replicas=2",), tokens=800)
        ),
        write(
            tmp_path, "005", attempt(scenario="other", commands=("kubectl apply -f a",), tokens=100)
        ),
    ]
    candidates = [select.candidate(f) for f in files]

    kept = select.select(candidates, cap=2)

    assert [(c.scenario_id, c.path.stem) for c in kept] == [
        ("other", "005"),
        ("pvc-pending", "002"),
        ("pvc-pending", "003"),
    ]
