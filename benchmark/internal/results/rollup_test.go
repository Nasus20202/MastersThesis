package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRollupAgentsAggregatesAcrossRuns(t *testing.T) {
	runs := []RunRef{
		{Summary: &RunSummary{ByCondition: map[string]ConditionSummary{
			"skill": {
				AttemptCount: 2, FullSuccessCount: 1, MeanScore: 0.5,
				Scenarios: map[string]ScenarioSummary{
					"task-a": {AttemptCount: 1, FullSuccessCount: 1, MeanScore: 1},
					"task-b": {AttemptCount: 1, FullSuccessCount: 0, MeanScore: 0},
				},
			},
		}}},
		{Summary: &RunSummary{ByCondition: map[string]ConditionSummary{
			"skill": {
				AttemptCount: 1, FullSuccessCount: 0, MeanScore: 0.25,
				Scenarios: map[string]ScenarioSummary{
					"task-a": {AttemptCount: 1, FullSuccessCount: 0, MeanScore: 0.25},
				},
			},
		}}},
		{Summary: nil},
	}

	rollups := RollupAgents(runs)
	require.Len(t, rollups, 1)
	agent := rollups[0]
	assert.Equal(t, "skill", agent.Agent)
	assert.Equal(t, 2, agent.Runs)
	assert.Equal(t, 3, agent.Rollup.Attempts)
	assert.Equal(t, 1, agent.Rollup.FullSuccessCount)
	assert.InDelta(t, 1.0/3.0, agent.Rollup.FullSuccessRate, 1e-9)
	assert.InDelta(t, 1.25/3.0, agent.Rollup.MeanScore, 1e-9)
	assert.InDelta(t, 0.625, agent.Scenarios["task-a"].MeanScore, 1e-9)
	assert.Equal(t, 2, agent.Scenarios["task-a"].Attempts)
}

func TestRollupTasksAggregatesAcrossRunsAndAgents(t *testing.T) {
	runs := []RunRef{
		{Summary: &RunSummary{ByCondition: map[string]ConditionSummary{
			"skill": {
				AttemptCount: 1, FullSuccessCount: 1, MeanScore: 1,
				Scenarios: map[string]ScenarioSummary{"task-a": {AttemptCount: 1, FullSuccessCount: 1, MeanScore: 1}},
			},
			"baseline": {
				AttemptCount: 1, FullSuccessCount: 0, MeanScore: 0.5,
				Scenarios: map[string]ScenarioSummary{"task-a": {AttemptCount: 1, FullSuccessCount: 0, MeanScore: 0.5}},
			},
		}}},
	}

	tasks := RollupTasks(runs)
	require.Len(t, tasks, 1)
	task := tasks[0]
	assert.Equal(t, "task-a", task.ScenarioID)
	assert.Equal(t, 1, task.Runs)
	assert.Equal(t, 2, task.Rollup.Attempts)
	assert.Equal(t, 1, task.Rollup.FullSuccessCount)
	assert.InDelta(t, 0.75, task.Rollup.MeanScore, 1e-9)
	assert.InDelta(t, 1.0, task.Agents["skill"].MeanScore, 1e-9)
	assert.InDelta(t, 0.5, task.Agents["baseline"].MeanScore, 1e-9)
}
