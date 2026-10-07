package results

import (
	"testing"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeCombinesShardRuns(t *testing.T) {
	root := t.TempDir()
	startedAt := time.Date(2026, time.October, 7, 10, 0, 0, 0, time.UTC)
	writeShard := func(runID string, offset time.Duration, scenarioID string, score float64, finish bool) {
		t.Helper()
		store, err := New(root, RunMetadata{
			RunID:       runID,
			StartedAt:   startedAt.Add(offset),
			Agents:      []string{"prompt"},
			Parallelism: 2,
			RepeatCount: 1,
			Scenarios:   []string{scenarioID},
		})
		require.NoError(t, err)
		if finish {
			result := orchestration.RunResult{ScenarioID: scenarioID, Grading: orchestration.GradingResult{Score: score, FullSuccess: score == 1}}
			require.NoError(t, store.WriteAttempt(1, "prompt", result))
		}
		require.NoError(t, store.Finalize(startedAt.Add(offset+time.Hour)))
	}
	writeShard("run-a", time.Minute, "scenario-a", 1, true)
	writeShard("run-b", 0, "scenario-b", 0.5, true)

	merged, err := Merge(root, "run-merged", []string{"run-a", "run-b"})
	require.NoError(t, err)
	assert.Equal(t, startedAt, merged.StartedAt)
	assert.Equal(t, startedAt.Add(time.Minute+time.Hour), *merged.CompletedAt)
	assert.Equal(t, RunStateCompleted, merged.State)
	assert.Equal(t, 4, merged.Parallelism)
	assert.Equal(t, 2, merged.ExpectedAttempts)

	snapshot, err := LoadRun(root, "run-merged")
	require.NoError(t, err)
	require.Len(t, snapshot.Attempts, 2)
	attempt, err := LoadAttempt(snapshot.Attempts[1], RunTypeBenchmark)
	require.NoError(t, err)
	assert.Equal(t, "run-merged", attempt.Benchmark.RunID)
	prompt := snapshot.Summary.ByCondition["prompt"]
	assert.Equal(t, 2, prompt.AttemptCount)
	require.NotNil(t, prompt.MacroAverageScore)
	assert.InDelta(t, 0.75, *prompt.MacroAverageScore, 1e-9)

	writeShard("run-c", 0, "scenario-c", 0, false)
	merged, err = Merge(root, "run-unfinished", []string{"run-a", "run-c"})
	require.NoError(t, err)
	assert.Equal(t, RunStateIncomplete, merged.State)

	// A later run of the same scenario adds repeats, numbered after the
	// earlier ones; scenario-b then has fewer and the run is incomplete.
	writeShard("run-a2", 3*time.Hour, "scenario-a", 0, true)
	merged, err = Merge(root, "run-repeated", []string{"run-a", "run-a2"})
	require.NoError(t, err)
	assert.Equal(t, RunStateCompleted, merged.State)
	assert.Equal(t, 2, merged.RepeatCount)
	assert.Equal(t, 2, merged.Parallelism)
	snapshot, err = LoadRun(root, "run-repeated")
	require.NoError(t, err)
	require.Len(t, snapshot.Attempts, 2)
	assert.Equal(t, 2, snapshot.Attempts[1].Attempt)
	assert.InDelta(t, 0.5, *snapshot.Summary.ByCondition["prompt"].MacroAverageScore, 1e-9)

	merged, err = Merge(root, "run-uneven", []string{"run-a", "run-b", "run-a2"})
	require.NoError(t, err)
	assert.Equal(t, RunStateIncomplete, merged.State)
	assert.Equal(t, 4, merged.ExpectedAttempts)

	_, err = Merge(root, "run-duplicate", []string{"run-a", "run-a"})
	assert.ErrorContains(t, err, "more than once")

	store, err := New(root, RunMetadata{RunID: "run-d", Agents: []string{"rag"}, Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario-d"}})
	require.NoError(t, err)
	require.NoError(t, store.Finalize(startedAt))
	_, err = Merge(root, "run-mismatch", []string{"run-a", "run-d"})
	assert.ErrorContains(t, err, "differs from")
}
