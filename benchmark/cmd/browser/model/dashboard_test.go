package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func TestRunMetricsAggregatesOutcomeAndDuration(t *testing.T) {
	root := t.TempDir()
	store, err := results.New(root, results.RunMetadata{RunID: "run-1", Agents: []string{"skill"}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1})
	require.NoError(t, err)
	require.NoError(t, store.WriteAttempt(1, "skill", attemptResult("alpha", true, 1, 10, 2)))
	require.NoError(t, store.WriteAttempt(2, "skill", attemptResult("alpha", false, 0.5, 45, 4)))

	read := NewStore(StoreConfig{ResultsRoot: root})
	require.NoError(t, read.Reload())

	metrics := RunMetrics(read, "run-1", "", "")
	assert.Equal(t, 2, metrics.Attempts)
	assert.Equal(t, 1, metrics.Full)
	assert.Equal(t, 1, metrics.Partial)
	assert.Equal(t, 2, metrics.Terminations[common.TerminationCompleted])
	assert.Equal(t, []float64{10, 45}, metrics.Durations)
}

func attemptResult(scenario string, full bool, score, duration float64, turns int) orchestration.RunResult {
	return orchestration.RunResult{
		ScenarioID: scenario, Condition: "skill",
		Agent:   &common.Result{Turns: turns, DurationSeconds: duration, Termination: common.TerminationCompleted},
		Grading: orchestration.GradingResult{Score: score, FullSuccess: full},
	}
}

func TestRunMetricsAggregatesTokenUsage(t *testing.T) {
	root := t.TempDir()
	store, err := results.New(root, results.RunMetadata{RunID: "run-1", Agents: []string{"skill"}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1})
	require.NoError(t, err)
	result := attemptResult("alpha", true, 1, 10, 2)
	result.Agent.TokenUsage = common.TokenUsage{PromptTokens: 25, CompletionTokens: 5, TotalTokens: 30, CachedTokens: 10}
	require.NoError(t, store.WriteAttempt(1, "skill", result))

	read := NewStore(StoreConfig{ResultsRoot: root})
	require.NoError(t, read.Reload())

	metrics := RunMetrics(read, "run-1", "", "")
	assert.Equal(t, []int{30}, metrics.Tokens)
	assert.Equal(t, []int{25}, metrics.Prompt)
	assert.Equal(t, []int{5}, metrics.Completion)
	assert.Equal(t, []int{10}, metrics.Cached)
	assert.Equal(t, []float64{0.4}, metrics.CacheRatios)
}

func TestScenarioCriteriaComparesConditions(t *testing.T) {
	root := t.TempDir()
	store, err := results.New(root, results.RunMetadata{RunID: "run-1", Agents: []string{"skill", "baseline"}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1})
	require.NoError(t, err)
	require.NoError(t, store.WriteAttempt(1, "skill", criteriaResult("alpha", "skill", map[string]bool{"ready": true})))
	require.NoError(t, store.WriteAttempt(1, "baseline", criteriaResult("alpha", "baseline", map[string]bool{"ready": false})))

	read := NewStore(StoreConfig{ResultsRoot: root})
	require.NoError(t, read.Reload())

	matrix := ScenarioCriteria(read, "run-1", "alpha")
	assert.Equal(t, []string{"baseline", "skill"}, matrix.Agents)
	require.Len(t, matrix.Rows, 1)
	assert.Equal(t, "ready", matrix.Rows[0].ID)
	assert.Equal(t, CriterionStat{Total: 1}, matrix.Rows[0].Cells["baseline"])
	assert.Equal(t, CriterionStat{Passed: 1, Total: 1}, matrix.Rows[0].Cells["skill"])
}

func criteriaResult(scenario, condition string, passed map[string]bool) orchestration.RunResult {
	var criteria []orchestration.CriterionResult
	for id, ok := range passed {
		criteria = append(criteria, orchestration.CriterionResult{ID: id, Passed: ok, Weight: 1})
	}
	return orchestration.RunResult{
		ScenarioID: scenario, Condition: condition,
		Grading: orchestration.GradingResult{Criteria: criteria},
	}
}
