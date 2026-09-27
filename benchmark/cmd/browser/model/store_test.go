package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

func taggedRun(runID, scenarioID string) results.RunRef {
	return results.RunRef{RunID: runID, Summary: &results.RunSummary{
		ByCondition: map[string]results.ConditionSummary{
			"skill": {
				AttemptCount: 1, FullSuccessCount: 1, MeanScore: 1,
				Scenarios: map[string]results.ScenarioSummary{
					scenarioID: {AttemptCount: 1, FullSuccessCount: 1, MeanScore: 1},
				},
			},
		},
	}}
}

func taggedStore() *Store {
	return &Store{
		catalog: map[string]scenario.Definition{
			"easy": {Tags: map[string]string{"difficulty": "easy"}},
			"hard": {Tags: map[string]string{"difficulty": "hard"}},
		},
		runs: []results.RunRef{taggedRun("easy-run", "easy"), taggedRun("hard-run", "hard")},
	}
}

func TestRunsFiltersByTagSelector(t *testing.T) {
	store := taggedStore()
	store.SetTagFilter(scenario.TagFilter{"difficulty": {"easy"}})
	require.Len(t, store.Runs(), 1)
	assert.Equal(t, "easy-run", store.Runs()[0].RunID)
}

func TestRunsForAgentFiltersByTagSelector(t *testing.T) {
	store := taggedStore()
	store.SetTagFilter(scenario.TagFilter{"difficulty": {"easy"}})
	require.Len(t, store.RunsForAgent("skill"), 1)
	assert.Equal(t, "easy-run", store.RunsForAgent("skill")[0].RunID)
}

func TestRunsForTaskHonoursTagSelector(t *testing.T) {
	store := taggedStore()
	store.SetTagFilter(scenario.TagFilter{"difficulty": {"easy"}})
	assert.Len(t, store.RunsForTask("easy"), 1)
	assert.Empty(t, store.RunsForTask("hard"))
}
