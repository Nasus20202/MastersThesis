package screens

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func TestAttemptHeadShowsThroughputCards(t *testing.T) {
	root := t.TempDir()
	store, err := results.New(root, results.RunMetadata{RunID: "run-1", Agents: []string{"skill"}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1})
	require.NoError(t, err)
	result := orchestration.RunResult{
		ScenarioID: "alpha", Condition: "skill",
		Agent: &common.Result{
			Termination: common.TerminationCompleted,
			Responses: []common.ResponseEvidence{
				{Response: inference.Result{Timings: &inference.Timings{PredictedN: 60, PredictedMS: 1000, DraftN: 30, DraftNAccepted: 12}}},
				{Response: inference.Result{Timings: &inference.Timings{PredictedN: 40, PredictedMS: 3000, DraftN: 20, DraftNAccepted: 13}}},
			},
		},
	}
	require.NoError(t, store.WriteAttempt(1, "skill", result))

	read := model.NewStore(model.StoreConfig{ResultsRoot: root})
	require.NoError(t, read.Reload())
	view := NewView(read)
	route := attemptRoute(t, read)

	out := view.buildDetailsHead(route, 120)
	assert.Contains(t, out, "out t/s")
	assert.Contains(t, out, "25.0")
	assert.Contains(t, out, "draft accept %")
	assert.Contains(t, out, "50%")
}

func TestAttemptHeadOmitsThroughputWithoutTimings(t *testing.T) {
	root := t.TempDir()
	store, err := results.New(root, results.RunMetadata{RunID: "run-1", Agents: []string{"skill"}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1})
	require.NoError(t, err)
	result := orchestration.RunResult{
		ScenarioID: "alpha", Condition: "skill",
		Agent: &common.Result{Termination: common.TerminationCompleted},
	}
	require.NoError(t, store.WriteAttempt(1, "skill", result))

	read := model.NewStore(model.StoreConfig{ResultsRoot: root})
	require.NoError(t, read.Reload())
	view := NewView(read)
	route := attemptRoute(t, read)

	out := view.buildDetailsHead(route, 120)
	assert.NotContains(t, out, "out t/s")
	assert.NotContains(t, out, "draft accept %")
}

func attemptRoute(t *testing.T, store *model.Store) *Route {
	t.Helper()
	require.NoError(t, store.EnsureSnapshot("run-1"))
	refs := store.AttemptsFor("run-1", "alpha", "skill")
	require.Len(t, refs, 1)
	return &Route{RunID: "run-1", ScenarioID: "alpha", Group: "skill", Attempt: refs[0].Attempt, Path: refs[0].Path}
}
