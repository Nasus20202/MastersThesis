package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
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

func TestRunMetricsAggregatesDecodingThroughput(t *testing.T) {
	root := t.TempDir()
	store, err := results.New(root, results.RunMetadata{RunID: "run-1", Agents: []string{"skill"}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1})
	require.NoError(t, err)
	result := attemptResult("alpha", true, 1, 10, 2)
	result.Agent.Responses = []common.ResponseEvidence{
		{Response: inference.Result{Timings: &inference.Timings{PredictedN: 60, PredictedMS: 1000, DraftN: 30, DraftNAccepted: 12}}},
		{Response: inference.Result{Timings: &inference.Timings{PredictedN: 40, PredictedMS: 3000, DraftN: 20, DraftNAccepted: 13}}},
		{Response: inference.Result{}},
	}
	require.NoError(t, store.WriteAttempt(1, "skill", result))

	read := NewStore(StoreConfig{ResultsRoot: root})
	require.NoError(t, read.Reload())

	metrics := RunMetrics(read, "run-1", "", "")
	assert.Equal(t, 100, metrics.PredictedTokens)
	assert.Equal(t, 4.0, metrics.PredictedSeconds)
	assert.Equal(t, 50, metrics.DraftTokens)
	assert.Equal(t, 25, metrics.DraftAccepted)
	assert.Equal(t, 25.0, PredictedTokensPerSecond(metrics))
	rate, ok := DraftAcceptanceRate(metrics)
	require.True(t, ok)
	assert.Equal(t, 0.5, rate)
}

func TestRunMetricsOmitsThroughputWithoutTimings(t *testing.T) {
	root := t.TempDir()
	store, err := results.New(root, results.RunMetadata{RunID: "run-1", Agents: []string{"skill"}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1})
	require.NoError(t, err)
	require.NoError(t, store.WriteAttempt(1, "skill", attemptResult("alpha", true, 1, 10, 2)))

	read := NewStore(StoreConfig{ResultsRoot: root})
	require.NoError(t, read.Reload())

	metrics := RunMetrics(read, "run-1", "", "")
	assert.Zero(t, PredictedTokensPerSecond(metrics))
	_, ok := DraftAcceptanceRate(metrics)
	assert.False(t, ok)
}

func TestAttemptTokensPerSecondSumsResponseTimings(t *testing.T) {
	attempt := results.Attempt{Benchmark: &results.AttemptResult{Agent: &common.Result{Responses: []common.ResponseEvidence{
		{Response: inference.Result{Timings: &inference.Timings{PredictedN: 60, PredictedMS: 1000, DraftN: 30, DraftNAccepted: 12}}},
		{Response: inference.Result{Timings: &inference.Timings{PredictedN: 40, PredictedMS: 3000, DraftN: 20, DraftNAccepted: 13}}},
		{Response: inference.Result{}},
	}}}}

	assert.Equal(t, 25.0, AttemptTokensPerSecond(attempt))
	rate, ok := AttemptDraftAcceptanceRate(attempt)
	require.True(t, ok)
	assert.Equal(t, 0.5, rate)
}

func TestAttemptThroughputWithoutTimings(t *testing.T) {
	attempt := results.Attempt{Benchmark: &results.AttemptResult{Agent: &common.Result{}}}
	assert.Zero(t, AttemptTokensPerSecond(attempt))
	_, ok := AttemptDraftAcceptanceRate(attempt)
	assert.False(t, ok)
	assert.Zero(t, AttemptTokensPerSecond(results.Attempt{}))
	_, ok = AttemptDraftAcceptanceRate(results.Attempt{})
	assert.False(t, ok)
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
