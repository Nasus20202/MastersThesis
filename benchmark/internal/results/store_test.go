package results

import (
	"encoding/json"
	"errors"
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

func TestStoreWritesRunMetadataAndAttemptEvidence(t *testing.T) {
	root := t.TempDir()
	startedAt := time.Date(2026, time.September, 11, 10, 0, 0, 0, time.UTC)
	store, err := New(root, RunMetadata{
		RunID:       "run-1",
		StartedAt:   startedAt,
		Parallelism: 2,
		RepeatCount: 2,
		Scenarios:   []string{"image-pull-failure"},
	})
	require.NoError(t, err)

	result := orchestration.RunResult{
		ScenarioID: "image-pull-failure",
		Condition:  "baseline",
		Agent: &common.Result{
			Task:        "Restore the application.",
			Messages:    []inference.Message{{Role: "user", Content: "Restore the application."}},
			Termination: common.TerminationCompleted,
		},
		Grading: orchestration.GradingResult{
			Criteria:    []orchestration.CriterionResult{{ID: "ready", Weight: 1, Passed: true, Stdout: "ready"}},
			Score:       1,
			FullSuccess: true,
		},
	}
	require.NoError(t, store.WriteAttempt(2, result))
	completedAt := startedAt.Add(time.Minute)
	require.NoError(t, store.Finalize(completedAt))

	runData, err := os.ReadFile(filepath.Join(root, "run-1", "run.json"))
	require.NoError(t, err)
	var metadata RunMetadata
	require.NoError(t, json.Unmarshal(runData, &metadata))
	assert.Equal(t, "run-1", metadata.RunID)
	assert.Equal(t, startedAt, metadata.StartedAt)
	require.NotNil(t, metadata.CompletedAt)
	assert.Equal(t, completedAt, *metadata.CompletedAt)
	assert.Equal(t, 2, metadata.Parallelism)
	assert.Equal(t, 2, metadata.RepeatCount)
	assert.Equal(t, []string{"image-pull-failure"}, metadata.Scenarios)

	attemptData, err := os.ReadFile(filepath.Join(root, "run-1", "image-pull-failure", "002.json"))
	require.NoError(t, err)
	var attempt AttemptResult
	require.NoError(t, json.Unmarshal(attemptData, &attempt))
	assert.Equal(t, "run-1", attempt.RunID)
	assert.Equal(t, "image-pull-failure", attempt.ScenarioID)
	assert.Equal(t, "baseline", attempt.Condition)
	assert.Equal(t, 2, attempt.Attempt)
	assert.Equal(t, result.Agent, attempt.Agent)
	assert.Equal(t, result.Grading, attempt.Grading)
}

func TestStoreRejectsInvalidMetadataAndAttempts(t *testing.T) {
	_, err := New(t.TempDir(), RunMetadata{Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"}})
	assert.ErrorContains(t, err, "run ID")

	store, err := New(t.TempDir(), RunMetadata{RunID: "run-1", Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"}})
	require.NoError(t, err)
	assert.ErrorContains(t, store.WriteAttempt(0, orchestration.RunResult{ScenarioID: "scenario"}), "attempt number")
	assert.ErrorContains(t, store.WriteAttempt(1, orchestration.RunResult{}), "scenario ID")
	assert.ErrorContains(t, store.WriteAttemptFailure(1, "scenario", orchestration.RunResult{}, nil), "execution error")

	_, err = New(t.TempDir(), RunMetadata{RunID: "run-1", Parallelism: 1, RepeatCount: 0, Scenarios: []string{"scenario"}})
	assert.ErrorContains(t, err, "repeat count")
	_, err = New(t.TempDir(), RunMetadata{RunID: "run-1", Parallelism: 0, RepeatCount: 1, Scenarios: []string{"scenario"}})
	assert.ErrorContains(t, err, "parallelism")
	_, err = New(t.TempDir(), RunMetadata{RunID: "run-1", Parallelism: 1, RepeatCount: 1})
	assert.ErrorContains(t, err, "at least one scenario")
}

func TestStoreWritesFailedAttemptEvidence(t *testing.T) {
	store, err := New(t.TempDir(), RunMetadata{RunID: "run-1", Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"}})
	require.NoError(t, err)
	wantErr := errors.New("cluster creation failed")
	failure := &orchestration.FailureEvidence{Phase: "create cluster", Program: "kind", Stderr: "kind failed", ExitCode: 1}
	require.NoError(t, store.WriteAttemptFailure(2, "scenario", orchestration.RunResult{Failure: failure}, wantErr))

	data, err := os.ReadFile(filepath.Join(store.runDir, "scenario", "002.json"))
	require.NoError(t, err)
	var artifact AttemptResult
	require.NoError(t, json.Unmarshal(data, &artifact))
	assert.Equal(t, "scenario", artifact.ScenarioID)
	assert.Equal(t, wantErr.Error(), artifact.Error)
	assert.Equal(t, failure, artifact.Failure)
}

func TestStoreWritesValidationAttemptEvidence(t *testing.T) {
	store, err := New(t.TempDir(), RunMetadata{RunID: "run-1", Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"}})
	require.NoError(t, err)
	result := orchestration.RunResult{
		ScenarioID: "scenario",
		Condition:  "validation",
		Grading:    orchestration.GradingResult{Score: 0.5, FullSuccess: false},
	}
	require.NoError(t, store.WriteValidationAttempt(1, "scenario", "partial", 0.5, false, result, nil))

	data, err := os.ReadFile(filepath.Join(store.runDir, "scenario", "partial", "001.json"))
	require.NoError(t, err)
	var artifact ValidationAttemptResult
	require.NoError(t, json.Unmarshal(data, &artifact))
	assert.Equal(t, "run-1", artifact.RunID)
	assert.Equal(t, "partial", artifact.CaseID)
	assert.Equal(t, "validation", artifact.Condition)
	assert.Equal(t, 0.5, artifact.ExpectedScore)
	assert.True(t, artifact.Passed)
}

func TestStoreWritesFailedValidationAttemptEvidence(t *testing.T) {
	store, err := New(t.TempDir(), RunMetadata{RunID: "run-1", Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"}})
	require.NoError(t, err)
	wantErr := errors.New("validation score mismatch")
	failure := &orchestration.FailureEvidence{Phase: "grading", Program: "kubectl", ExitCode: 1}
	result := orchestration.RunResult{
		ScenarioID: "scenario",
		Failure:    failure,
		Grading:    orchestration.GradingResult{Score: 0.5},
	}
	require.NoError(t, store.WriteValidationAttempt(2, "scenario", "broken", 1, true, result, wantErr))

	data, err := os.ReadFile(filepath.Join(store.runDir, "scenario", "broken", "002.json"))
	require.NoError(t, err)
	var artifact ValidationAttemptResult
	require.NoError(t, json.Unmarshal(data, &artifact))
	assert.Equal(t, "run-1", artifact.RunID)
	assert.Equal(t, "scenario", artifact.ScenarioID)
	assert.Equal(t, "broken", artifact.CaseID)
	assert.False(t, artifact.Passed)
	assert.Equal(t, wantErr.Error(), artifact.Error)
	assert.Equal(t, failure, artifact.Failure)
}

func TestStoreRejectsInvalidValidationAttempt(t *testing.T) {
	store, err := New(t.TempDir(), RunMetadata{RunID: "run-1", Parallelism: 1, RepeatCount: 1, Scenarios: []string{"scenario"}})
	require.NoError(t, err)
	result := orchestration.RunResult{ScenarioID: "scenario"}

	assert.ErrorContains(t, store.WriteValidationAttempt(0, "scenario", "case", 0, false, result, nil), "attempt number")
	assert.ErrorContains(t, store.WriteValidationAttempt(1, "", "case", 0, false, result, nil), "scenario ID")
	assert.ErrorContains(t, store.WriteValidationAttempt(1, "scenario", "", 0, false, result, nil), "case ID")
}

func TestWriteJSONReturnsMarshalErrors(t *testing.T) {
	err := writeJSON(filepath.Join(t.TempDir(), "result.json"), make(chan int))
	assert.Error(t, err)
}
