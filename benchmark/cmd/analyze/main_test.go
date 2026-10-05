package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunScoresSearchesAgainstReference(t *testing.T) {
	root := t.TempDir()
	scenarios := filepath.Join(root, "scenarios", "networking", "svc")
	require.NoError(t, os.MkdirAll(scenarios, 0o750))
	scenario := "id: svc\ntitle: Service\ntask: Fix it.\nsources: [{path: docs/service.md}]\n" +
		"prepare: [{program: prepare}]\nverify_clean: [{program: check}]\ngrading: [{id: ready, weight: 1, check: {program: check}}]\n"
	require.NoError(t, os.WriteFile(filepath.Join(scenarios, "scenario.yaml"), []byte(scenario), 0o600))
	store, err := results.New(filepath.Join(root, "results"), results.RunMetadata{RunID: "run", Agents: []string{"rag", "prompt"}, Parallelism: 1, RepeatCount: 1, Scenarios: []string{"svc"}})
	require.NoError(t, err)
	search := common.ToolCallEvidence{
		Call:    inference.ToolCall{Name: "search_docs"},
		Details: map[string]any{"query": "service port", "hits": []any{map[string]any{"rank": 1, "chunk": map[string]any{"path": "docs/service.md"}}}},
	}
	for condition, score := range map[string]float64{"rag": 1, "prompt": 0} {
		agent := &common.Result{Termination: "completed"}
		if condition == "rag" {
			agent.ToolCalls = []common.ToolCallEvidence{search}
		}
		require.NoError(t, store.WriteAttempt(1, condition, orchestration.RunResult{ScenarioID: "svc", Condition: condition, Agent: agent, Grading: orchestration.GradingResult{Score: score, FullSuccess: score == 1}}))
	}

	var output bytes.Buffer
	err = run([]string{"--results", filepath.Join(root, "results"), "--scenarios", filepath.Join(root, "scenarios"), "--condition", "run/rag", "--reference", "run/prompt"}, &output, &bytes.Buffer{})

	require.NoError(t, err)
	assert.Contains(t, output.String(), "run/rag: 1 attempts, macro 1.000, full success 1/1")
	assert.Contains(t, output.String(), "source in top k: 1/1 searches, 1/1 attempts")
	assert.Regexp(t, `source retrieved\s+1\s+1\s+1\.000\s+1/1\s+0\.000\s+0/1`, output.String())

	err = run([]string{"--results", filepath.Join(root, "results"), "--scenarios", filepath.Join(root, "scenarios")}, &output, &bytes.Buffer{})
	assert.EqualError(t, err, "--condition is required")
}
