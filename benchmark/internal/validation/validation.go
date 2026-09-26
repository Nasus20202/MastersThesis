// Package validation defines the validation-case YAML schema and checks a
// scenario run's grading against the expected outcome.
package validation

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/executor"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

type Definition struct {
	Scenarios []Scenario `yaml:"scenarios" validate:"required,min=1,dive"`
	sourceDir string
}

type Scenario struct {
	ScenarioFile string `yaml:"scenario_file" validate:"required,notblank"`
	Cases        []Case `yaml:"cases" validate:"required,min=1,dive"`
	sourceDir    string
}

type Case struct {
	ID                  string        `yaml:"id" validate:"required,validationid,max=32"`
	Repair              scenario.Step `yaml:"repair,omitempty" validate:"omitempty,min=1,dive"`
	ExpectedScore       *float64      `yaml:"expected_score" validate:"required,gte=0,lte=1"`
	ExpectedFullSuccess *bool         `yaml:"expected_full_success" validate:"required"`
}

type ValidationCase struct {
	Scenario            scenario.Definition
	ID                  string
	Repair              scenario.Step
	ExpectedScore       float64
	ExpectedFullSuccess bool
}

type RunFunc func(context.Context, scenario.Definition, scenario.Step) (orchestration.RunResult, error)

func LoadCases(inputs []string) ([]ValidationCase, error) {
	definitions, err := LoadInputs(inputs)
	if err != nil {
		return nil, err
	}

	cases := make([]ValidationCase, 0)
	seen := make(map[string]struct{})
	for _, definition := range definitions {
		for _, item := range definition.Scenarios {
			scenarioDefinition, err := scenario.Load(item.ScenarioPath())
			if err != nil {
				return nil, fmt.Errorf("load validation scenario %q: %w", item.ScenarioPath(), err)
			}
			for _, validationCase := range item.Cases {
				key := scenarioDefinition.ID + "/" + validationCase.ID
				if _, exists := seen[key]; exists {
					return nil, fmt.Errorf("duplicate validation case %q", key)
				}
				seen[key] = struct{}{}
				cases = append(cases, ValidationCase{
					Scenario:            scenarioDefinition,
					ID:                  validationCase.ID,
					Repair:              validationCase.Repair,
					ExpectedScore:       *validationCase.ExpectedScore,
					ExpectedFullSuccess: *validationCase.ExpectedFullSuccess,
				})
			}
		}
	}
	if len(cases) == 0 {
		return nil, errors.New("validation requires at least one case")
	}
	return cases, nil
}

func Tasks(cases []ValidationCase, run RunFunc) []executor.Task {
	tasks := make([]executor.Task, len(cases))
	for index, item := range cases {
		tasks[index] = executor.Task{
			ScenarioID: item.Scenario.ID,
			CaseID:     item.ID,
			Run: func(ctx context.Context) (orchestration.RunResult, error) {
				return run(ctx, item.Scenario, item.Repair)
			},
		}
	}
	return tasks
}

func Check(item ValidationCase, result orchestration.RunResult) error {
	if result.ScenarioID != item.Scenario.ID {
		return fmt.Errorf("scenario ID %q does not match %q", result.ScenarioID, item.Scenario.ID)
	}
	if math.Abs(result.Grading.Score-item.ExpectedScore) > 1e-9 {
		return fmt.Errorf("score %.6f does not match expected %.6f", result.Grading.Score, item.ExpectedScore)
	}
	if result.Grading.FullSuccess != item.ExpectedFullSuccess {
		return fmt.Errorf("full_success %t does not match expected %t", result.Grading.FullSuccess, item.ExpectedFullSuccess)
	}
	return nil
}
