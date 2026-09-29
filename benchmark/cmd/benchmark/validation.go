package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/config"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/executor"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/validation"
)

func runValidation(ctx context.Context, inputs []string, parallelism, repeat int, benchmarkConfig benchmarkconfig.Config, terminal *ui.Terminal, tagFilter scenario.TagFilter) error {
	cases, err := validation.LoadCases(inputs)
	if err != nil {
		return err
	}
	cases = validation.FilterCases(cases, tagFilter)
	if len(cases) == 0 {
		return fmt.Errorf("no validation cases match tag filter %q", tagFilter.String())
	}
	deps, err := newRunnerDeps(command.LocalExecutor{Environment: os.Environ()}, benchmarkConfig.Containers)
	if err != nil {
		return err
	}
	store, err := startRun(ctx, results.RunMetadata{
		RunType:          results.RunTypeValidation,
		ExpectedAttempts: len(cases) * repeat,
		Parallelism:      parallelism,
		RepeatCount:      repeat,
		Scenarios:        validationScenarioIDs(cases),
		TagSelector:      tagFilter.String(),
	})
	if err != nil {
		return err
	}
	metadata := store.Metadata()
	logger := slog.Default()
	logger.Info("validation runner started",
		"run_id", metadata.RunID,
		"cases", len(cases),
		"parallelism", parallelism,
		"repeat_count", repeat,
		"total_tasks", len(cases)*repeat,
	)

	runCase := func(ctx context.Context, definition scenario.Definition, repair scenario.Step) (orchestration.RunResult, error) {
		return newOrchestrationRunner(deps, definition, "", nil, nil).RunWithRepair(ctx, definition, repair)
	}
	tasks := make([]executor.Task, 0, len(cases)*repeat)
	for attempt := 1; attempt <= repeat; attempt++ {
		for _, task := range validation.Tasks(cases, runCase) {
			task.Attempt = attempt
			tasks = append(tasks, task)
		}
	}
	caseByKey := make(map[string]validation.ValidationCase, len(cases))
	for _, item := range cases {
		caseByKey[item.Scenario.ID+"/"+item.ID] = item
	}
	var validationErrors []error
	runErr := executeRun(ctx, store, terminal, tasks, parallelism, func(outcome executor.Outcome) (bool, error) {
		item, exists := caseByKey[outcome.ScenarioID+"/"+outcome.CaseID]
		if !exists {
			validationErrors = append(validationErrors, fmt.Errorf("validation outcome has unknown case %s/%s", outcome.ScenarioID, outcome.CaseID))
			return false, nil
		}
		caseErr := outcome.Err
		if caseErr == nil {
			caseErr = validation.Check(item, outcome.Result)
		}
		writeErr := store.WriteValidationAttempt(outcome.Attempt, item.Scenario.ID, item.ID, item.ExpectedScore, item.ExpectedFullSuccess, outcome.Result, caseErr)
		if caseErr != nil {
			logger.Error("validation case failed",
				"run_id", metadata.RunID,
				"scenario", item.Scenario.ID,
				"case", item.ID,
				"attempt", outcome.Attempt,
				"score", outcome.Result.Grading.Score,
				"full_success", outcome.Result.Grading.FullSuccess,
				"error", caseErr,
			)
			validationErrors = append(validationErrors, fmt.Errorf("%s/%s: %w", item.Scenario.ID, item.ID, caseErr))
			return false, writeErr
		}
		logger.Info("validation case passed",
			"run_id", metadata.RunID,
			"scenario", item.Scenario.ID,
			"case", item.ID,
			"attempt", outcome.Attempt,
			"score", outcome.Result.Grading.Score,
			"full_success", outcome.Result.Grading.FullSuccess,
		)
		return true, writeErr
	})
	return errors.Join(runErr, errors.Join(validationErrors...))
}

func validationScenarioIDs(cases []validation.ValidationCase) []string {
	ids := make([]string, 0, len(cases))
	seen := make(map[string]struct{}, len(cases))
	for _, item := range cases {
		if _, exists := seen[item.Scenario.ID]; exists {
			continue
		}
		seen[item.Scenario.ID] = struct{}{}
		ids = append(ids, item.Scenario.ID)
	}
	return ids
}
