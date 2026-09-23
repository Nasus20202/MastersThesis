package results

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRunsReadsMetadataAndSummaryNewestFirst(t *testing.T) {
	root := t.TempDir()
	base := time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC)

	older, err := New(root, RunMetadata{
		RunID: "run-older", StartedAt: base,
		Agents: []string{"skill"}, Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"},
	})
	require.NoError(t, err)
	require.NoError(t, older.Finalize(base.Add(time.Minute)))

	newer, err := New(root, RunMetadata{
		RunID: "run-newer", StartedAt: base.Add(time.Hour),
		Agents: []string{"skill"}, Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"},
	})
	require.NoError(t, err)
	require.NoError(t, newer.WriteAttempt(1, "skill", orchestration.RunResult{ScenarioID: "scenario", Condition: "skill"}))
	require.NoError(t, newer.Finalize(base.Add(time.Hour)))

	require.NoError(t, os.Mkdir(filepath.Join(root, "not-a-run"), 0o755))

	runs, err := ListRuns(root)
	require.NoError(t, err)
	require.Len(t, runs, 2)
	assert.Equal(t, "run-newer", runs[0].RunID)
	assert.Equal(t, "run-older", runs[1].RunID)
	require.NotNil(t, runs[0].Summary)
	assert.Equal(t, RunStateCompleted, runs[0].Summary.State)
}

func TestListRunsDerivesSummaryWhenResultsMissing(t *testing.T) {
	root := t.TempDir()
	store, err := New(root, RunMetadata{
		RunID: "run-interrupted", StartedAt: time.Now().UTC(),
		Agents: []string{"skill"}, Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"},
	})
	require.NoError(t, err)
	require.NoError(t, store.WriteAttempt(1, "skill", orchestration.RunResult{
		ScenarioID: "scenario", Condition: "skill",
		Grading: orchestration.GradingResult{Score: 0.5},
	}))
	require.NoError(t, os.Remove(filepath.Join(root, "run-interrupted", "results.json")))

	runs, err := ListRuns(root)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	require.NotNil(t, runs[0].Summary)
	assert.Equal(t, 1, runs[0].Summary.AttemptsRecorded)
	assert.Equal(t, 0.5, runs[0].Summary.ByCondition["skill"].MeanScore)
}

func TestLoadRunIndexesAttemptsAndLoadsPayloads(t *testing.T) {
	root := t.TempDir()
	store, err := New(root, RunMetadata{
		RunID: "run-1", StartedAt: time.Now().UTC(),
		Agents: []string{"baseline"}, Parallelism: 1, RepeatCount: 2, Scenarios: []string{"scenario"},
	})
	require.NoError(t, err)
	for _, attempt := range []int{1, 2} {
		require.NoError(t, store.WriteAttempt(attempt, "baseline", orchestration.RunResult{
			ScenarioID: "scenario", Condition: "baseline",
			Grading: orchestration.GradingResult{Criteria: []orchestration.CriterionResult{{ID: "ready", Weight: 1, Passed: true}}, Score: 1, FullSuccess: true},
		}))
	}

	snapshot, err := LoadRun(root, "run-1")
	require.NoError(t, err)
	assert.Equal(t, "run-1", snapshot.RunID)
	assert.Equal(t, []string{"baseline"}, snapshot.Metadata.Agents)
	require.Len(t, snapshot.Attempts, 2)
	assert.Equal(t, AttemptRef{ScenarioID: "scenario", Group: "baseline", Attempt: 1, Path: filepath.Join(root, "run-1", "scenario", "baseline", "001.json")}, snapshot.Attempts[0])

	attempt, err := LoadAttempt(snapshot.Attempts[1], snapshot.Metadata.RunType)
	require.NoError(t, err)
	require.NotNil(t, attempt.Benchmark)
	assert.Nil(t, attempt.Validation)
	assert.Equal(t, 2, attempt.Benchmark.Attempt)
	assert.True(t, attempt.Grading().FullSuccess)
	assert.Empty(t, attempt.Error())
}

func TestLoadRunLoadsValidationAttempts(t *testing.T) {
	root := t.TempDir()
	store, err := New(root, RunMetadata{
		RunID: "run-validation", StartedAt: time.Now().UTC(),
		Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"},
	})
	require.NoError(t, err)
	require.NoError(t, store.WriteValidationAttempt(1, "scenario", "repaired", 1, true, orchestration.RunResult{
		ScenarioID: "scenario", Condition: "validation",
		Grading: orchestration.GradingResult{Score: 1, FullSuccess: true},
	}, nil))

	snapshot, err := LoadRun(root, "run-validation")
	require.NoError(t, err)
	assert.Equal(t, "validation", snapshot.Metadata.RunType)
	require.Len(t, snapshot.Attempts, 1)
	assert.Equal(t, "repaired", snapshot.Attempts[0].Group)

	attempt, err := LoadAttempt(snapshot.Attempts[0], snapshot.Metadata.RunType)
	require.NoError(t, err)
	require.NotNil(t, attempt.Validation)
	assert.Nil(t, attempt.Benchmark)
	assert.Equal(t, "repaired", attempt.Validation.CaseID)
	assert.True(t, attempt.Grading().FullSuccess)
}

func TestLoadRunDerivesSummaryWhenResultsMissing(t *testing.T) {
	root := t.TempDir()
	store, err := New(root, RunMetadata{
		RunID: "run-interrupted", StartedAt: time.Now().UTC(),
		Agents: []string{"skill"}, Parallelism: 1, RepeatCount: 2, Scenarios: []string{"scenario"},
	})
	require.NoError(t, err)
	require.NoError(t, store.WriteAttempt(1, "skill", orchestration.RunResult{
		ScenarioID: "scenario", Condition: "skill",
		Grading: orchestration.GradingResult{Score: 0.5},
	}))
	require.NoError(t, os.Remove(filepath.Join(root, "run-interrupted", "results.json")))

	snapshot, err := LoadRun(root, "run-interrupted")
	require.NoError(t, err)
	require.NotNil(t, snapshot.Summary)
	assert.Equal(t, 1, snapshot.Summary.AttemptsRecorded)
	assert.Equal(t, RunStateRunning, snapshot.Summary.State)
	condition := snapshot.Summary.ByCondition["skill"]
	assert.Equal(t, 0.5, condition.MeanScore)
}

func TestListRunsDerivesThroughputWhenResultsMissing(t *testing.T) {
	root := t.TempDir()
	store, err := New(root, RunMetadata{
		RunID: "run-interrupted", StartedAt: time.Now().UTC(),
		Agents: []string{"skill"}, Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"},
	})
	require.NoError(t, err)
	require.NoError(t, store.WriteAttempt(1, "skill", orchestration.RunResult{
		ScenarioID: "scenario", Condition: "skill",
		Agent: &common.Result{Responses: []common.ResponseEvidence{
			{Response: inference.Result{Timings: &inference.Timings{PredictedN: 60, PredictedMS: 1000, DraftN: 30, DraftNAccepted: 12}}},
		}},
	}))
	require.NoError(t, os.Remove(filepath.Join(root, "run-interrupted", "results.json")))

	runs, err := ListRuns(root)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	throughput := runs[0].Summary.ByCondition["skill"].Throughput
	require.NotNil(t, throughput)
	assert.Equal(t, 60, throughput.PredictedTokens)
	assert.Equal(t, 60.0, throughput.PredictedTokensPerSecond)
	require.NotNil(t, throughput.DraftAcceptanceRate)
	assert.Equal(t, 0.4, *throughput.DraftAcceptanceRate)
}
