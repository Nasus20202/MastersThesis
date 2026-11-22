import json

from finetune.data import truncation


def call(i, command):
    return {
        "id": f"c{i}",
        "type": "function",
        "name": "bash",
        "arguments": json.dumps({"command": command}),
    }


def turn(*calls):
    return {"role": "assistant", "content": "", "tool_calls": list(calls)}


def result(i, code=0):
    return {"role": "tool", "tool_call_id": f"c{i}", "content": f"exit_code: {code}\nstdout:\nok"}


CONVERSATION = [
    {"role": "user", "content": "task"},
    turn(call(1, "kubectl get pods")),
    result(1),
    turn(call(2, "kubectl apply -f a.yaml")),
    result(2),
    turn(call(3, "kubectl get pods"), call(4, "kubectl describe pod p")),
    result(3, code=1),
    result(4),
    turn(call(5, "kubectl get pods")),
    result(5),
    {"role": "assistant", "content": "done"},
]


def test_cut_after_the_first_turn_with_a_successful_read_only_call_after_the_last_write():
    # Turn 3 succeeds through call 4; its whole result block stays, later checks go.
    assert truncation.cut_point(CONVERSATION) == 8


def test_no_cut_without_a_write_or_without_a_successful_observation_after_it():
    assert truncation.cut_point(CONVERSATION[:3]) is None
    failing = [*CONVERSATION[:5], turn(call(6, "kubectl get pods")), result(6, code=1)]
    assert truncation.cut_point(failing) is None


def test_a_write_in_the_observing_turn_does_not_count_as_observation():
    messages = [*CONVERSATION[:3], turn(call(2, "kubectl apply -f a.yaml")), result(2)]
    assert truncation.cut_point(messages) is None


def attempt(path, messages):
    body = {
        "scenario_id": "scenario",
        "condition": "prompt",
        "agent": {
            "inference": {"model": "teacher"},
            "token_usage": {"peak_context_tokens": 1},
            "tools": [],
            "messages": messages,
        },
    }
    path.write_text(json.dumps(body))
    return path


def test_the_cut_ends_with_the_original_final_report(tmp_path):
    path = attempt(tmp_path / "001.json", CONVERSATION)

    traj = truncation.cut_trajectory(path)

    assert [m["role"] for m in traj.messages] == [
        *[m["role"] for m in CONVERSATION[:8]],
        "assistant",
    ]
    assert traj.messages[-1]["content"] == "done"


def test_a_record_without_a_final_report_is_refused(tmp_path):
    path = attempt(tmp_path / "001.json", CONVERSATION[:-1])

    try:
        truncation.cut_trajectory(path)
    except ValueError as error:
        assert "final report" in str(error)
    else:
        raise AssertionError("a record without a final report was accepted")
