package lifecycle

import (
	"errors"
	"fmt"
	"slices"
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
	ScenarioID string        `json:"scenario_id"`
	Grading    GradingResult `json:"grading"`
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
