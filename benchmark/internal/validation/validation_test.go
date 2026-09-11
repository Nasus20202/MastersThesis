package validation

import (
	"context"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksPreserveCasesAndPassRepairToRunner(t *testing.T) {
	definition := scenario.Definition{ID: "scenario"}
	cases := []ValidationCase{{
		Scenario:            definition,
		ID:                  "case",
		Repair:              scenario.Step{{Program: "repair"}},
		ExpectedScore:       1,
		ExpectedFullSuccess: true,
	}}
	var gotRepair scenario.Step
	tasks := Tasks(cases, func(_ context.Context, gotDefinition scenario.Definition, repair scenario.Step) (orchestration.RunResult, error) {
		assert.Equal(t, definition, gotDefinition)
		gotRepair = repair
		return orchestration.RunResult{ScenarioID: "scenario"}, nil
	})

	require.Len(t, tasks, 1)
	_, err := tasks[0].Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "repair", gotRepair[0].Program)
}

func TestCheckMatchesExpectedScoreAndSuccess(t *testing.T) {
	item := ValidationCase{
		Scenario:            scenario.Definition{ID: "scenario"},
		ExpectedScore:       0.5,
		ExpectedFullSuccess: false,
	}
	result := orchestration.RunResult{
		ScenarioID: "scenario",
		Grading:    orchestration.GradingResult{Score: 0.5, FullSuccess: false},
	}

	assert.NoError(t, Check(item, result))
	assert.ErrorContains(t, Check(item, orchestration.RunResult{
		ScenarioID: "scenario",
		Grading:    orchestration.GradingResult{Score: 1, FullSuccess: true},
	}), "score")
}
