package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

// modelStore has a run recording its model in the sampling metadata and an
// older run whose model is known only from its attempts.
func modelStore(t *testing.T) *Store {
	t.Helper()
	root := t.TempDir()
	write := func(runID, model string, sampling bool, score float64) {
		metadata := results.RunMetadata{RunID: runID, Agents: []string{"skill"}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1}
		if sampling {
			metadata.Sampling = &results.SamplingProvenance{Model: model}
		}
		store, err := results.New(root, metadata)
		require.NoError(t, err)
		result := attemptResult("alpha", score == 1, score, 10, 2)
		result.Agent.Inference = inference.Metadata{Model: model}
		require.NoError(t, store.WriteAttempt(1, "skill", result))
	}
	write("run-gemma", "gemma", true, 1)
	write("run-qwen", "qwen", false, 0.5)

	store := NewStore(StoreConfig{ResultsRoot: root})
	require.NoError(t, store.Reload())
	return store
}

func TestStoreFiltersRunsByModel(t *testing.T) {
	store := modelStore(t)
	assert.Equal(t, []string{"gemma", "qwen"}, store.ModelOptions())
	assert.Len(t, store.Runs(), 2)

	store.SetModelFilter([]string{"qwen"})
	require.Len(t, store.Runs(), 1)
	assert.Equal(t, "run-qwen", store.Runs()[0].RunID)
	require.Len(t, store.Agents(), 1)
	assert.InDelta(t, 0.5, store.Agents()[0].MeanScore, 1e-9)
	assert.Equal(t, []string{"qwen"}, store.ModelFilter())

	store.SetModelFilter(nil)
	assert.Len(t, store.Runs(), 2)
}

func TestModelMetricsGroupsVisibleRunsByModel(t *testing.T) {
	store := modelStore(t)

	metrics := ModelMetrics(store, "skill", "alpha")
	require.Len(t, metrics, 2)
	assert.InDelta(t, 1, metrics["gemma"].MeanScore, 1e-9)
	assert.InDelta(t, 0.5, metrics["qwen"].MeanScore, 1e-9)
	assert.Empty(t, ModelMetrics(store, "baseline", ""))

	store.SetModelFilter([]string{"gemma"})
	assert.Len(t, ModelMetrics(store, "", ""), 1)
}

func TestAgentModelMatrixRollsUpConditionsByModel(t *testing.T) {
	store := modelStore(t)

	matrix := store.AgentModelMatrix()
	assert.Equal(t, []string{"skill"}, matrix.Agents)
	assert.Equal(t, []string{"gemma", "qwen"}, matrix.Models)
	assert.Equal(t, results.Rollup{Attempts: 1, FullSuccessCount: 1, FullSuccessRate: 1, MeanScore: 1}, matrix.Cells["skill"]["gemma"])
	assert.InDelta(t, 0.5, matrix.Cells["skill"]["qwen"].MeanScore, 1e-9)

	store.SetModelFilter([]string{"qwen"})
	assert.Equal(t, []string{"qwen"}, store.AgentModelMatrix().Models)
}
