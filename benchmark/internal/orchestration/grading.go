package orchestration

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

type CriterionResult struct {
	ID              string  `json:"id"`
	Weight          float64 `json:"weight"`
	Passed          bool    `json:"passed"`
	Stdout          string  `json:"stdout"`
	Stderr          string  `json:"stderr"`
	Error           string  `json:"error,omitempty"`
	ExitCode        int     `json:"exit_code"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type GradingResult struct {
	Criteria    []CriterionResult `json:"criteria"`
	Score       float64           `json:"score"`
	FullSuccess bool              `json:"full_success"`
}

type RunResult struct {
	ScenarioID string           `json:"scenario_id"`
	Condition  string           `json:"condition"`
	Agent      *common.Result   `json:"agent,omitempty"`
	Grading    GradingResult    `json:"grading"`
	Failure    *FailureEvidence `json:"failure,omitempty"`
}

func (r Runner) runGrading(ctx context.Context, criteria []scenario.Criterion, kubeconfigPath string, executor command.Executor) (GradingResult, error) {
	results := make([]CriterionResult, 0, len(criteria))
	for _, criterion := range criteria {
		result, err := executor.Run(ctx, withKubeconfig(criterion.Check.Spec(), kubeconfigPath))
		criterionResult := CriterionResult{
			ID:              criterion.ID,
			Weight:          criterion.Weight,
			Passed:          err == nil && result.ExitCode == 0,
			Stdout:          result.Stdout,
			Stderr:          result.Stderr,
			ExitCode:        result.ExitCode,
			DurationSeconds: result.Duration.Seconds(),
		}
		if err != nil {
			criterionResult.Error = err.Error()
		}
		results = append(results, criterionResult)
	}
	return calculateGradingResult(results)
}

func calculateGradingResult(criteria []CriterionResult) (GradingResult, error) {
	if len(criteria) == 0 {
		return GradingResult{}, errors.New("grading requires at least one criterion result")
	}

	var totalWeight float64
	var passedWeight float64
	fullSuccess := true
	for _, criterion := range criteria {
		if criterion.Weight <= 0 {
			return GradingResult{}, fmt.Errorf("criterion %q has non-positive weight", criterion.ID)
		}

		totalWeight += criterion.Weight
		if criterion.Passed {
			passedWeight += criterion.Weight
		} else {
			fullSuccess = false
		}
	}

	return GradingResult{
		Criteria:    slices.Clone(criteria),
		Score:       passedWeight / totalWeight,
		FullSuccess: fullSuccess,
	}, nil
}
