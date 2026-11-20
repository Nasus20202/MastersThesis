import json

from finetune.data import trajectories


def attempt(termination="completed", full_success=True):
    return {
        "scenario_id": "basic-deploy-service",
        "condition": "prompt",
        "agent": {
            "inference": {"model": "Qwen3.5-9B-Q4_K_M"},
            "token_usage": {"peak_context_tokens": 6292},
            "termination": termination,
            "tools": [{"name": "bash", "description": "Run.", "parameters": {"type": "object"}}],
            "messages": [
                {"role": "system", "content": "sys"},
                {"role": "user", "content": "task"},
                {
                    "role": "assistant",
                    "content": "",
                    "reasoning_content": "think",
                    "tool_calls": [
                        {
                            "id": "c1",
                            "type": "function",
                            "name": "bash",
                            "arguments": '{"command":"kubectl get ns"}',
                        }
                    ],
                },
                {"role": "tool", "content": "exit_code: 0", "tool_call_id": "c1"},
                {"role": "assistant", "content": "done"},
            ],
        },
        "grading": {"full_success": full_success},
    }


def write(tmp_path, data):
    path = tmp_path / "001.json"
    path.write_text(json.dumps(data))
    return path


def test_load_converts_tool_calls_to_openai_shape(tmp_path):
    traj = trajectories.load(write(tmp_path, attempt()))

    call = traj.messages[2]["tool_calls"][0]
    assert call == {
        "id": "c1",
        "type": "function",
        "function": {"name": "bash", "arguments": {"command": "kubectl get ns"}},
    }
    assert traj.messages[2]["reasoning_content"] == "think"
    assert traj.messages[3]["tool_call_id"] == "c1"
    assert "reasoning_content" not in traj.messages[4]
    assert traj.tools[0]["type"] == "function"
    assert traj.tools[0]["function"]["name"] == "bash"


def test_is_clean_requires_full_score_and_own_ending(tmp_path):
    assert trajectories.is_clean(write(tmp_path, attempt()))
    assert not trajectories.is_clean(write(tmp_path, attempt(termination="turn_limit")))
    assert not trajectories.is_clean(write(tmp_path, attempt(full_success=False)))
