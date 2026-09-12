package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/executor"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/validation"
)

func runValidation(ctx context.Context, inputs []string, parallelism, repeat int, output io.Writer) error {
	cases, err := validation.LoadCases(inputs)
	if err != nil {
		return err
	}
	startedAt := time.Now().UTC()
	metadata := results.RunMetadata{
		RunID:       runID(startedAt),
		StartedAt:   startedAt,
		Parallelism: parallelism,
		RepeatCount: repeat,
		Scenarios:   validationScenarioIDs(cases),
	}
	store, err := results.New("results", metadata)
	if err != nil {
		return err
	}
	logger := slog.Default()
	logger.Info("validation runner started",
		"run_id", metadata.RunID,
		"cases", len(cases),
		"parallelism", parallelism,
		"repeat_count", repeat,
		"total_tasks", len(cases)*repeat,
	)

	commandExecutor := command.LocalExecutor{}
	imageBuilder, err := newSandboxImageBuilder(commandExecutor)
	if err != nil {
		return err
	}
	runCase := func(ctx context.Context, definition scenario.Definition, repair scenario.Step) (orchestration.RunResult, error) {
		return newOrchestrationRunner(commandExecutor, definition, imageBuilder, nil).RunWithRepair(ctx, definition, repair)
	}
	tasks := make([]executor.Task, 0, len(cases)*repeat)
	for attempt := 1; attempt <= repeat; attempt++ {
		for _, task := range validation.Tasks(cases, runCase) {
			task.Attempt = attempt
			tasks = append(tasks, task)
		}
	}
	outcomes, executeErrors := executor.Execute(ctx, tasks, parallelism)
	caseByKey := make(map[string]validation.ValidationCase, len(cases))
	for _, item := range cases {
		caseByKey[item.Scenario.ID+"/"+item.ID] = item
	}
	validationErrors := make([]error, 0)
	var writeErr error
	var outputErr error
	for outcome := range outcomes {
		item, exists := caseByKey[outcome.ScenarioID+"/"+outcome.CaseID]
		if !exists {
			validationErrors = append(validationErrors, fmt.Errorf("validation outcome has unknown case %s/%s", outcome.ScenarioID, outcome.CaseID))
			continue
		}
		caseErr := outcome.Err
		if caseErr == nil {
			caseErr = validation.Check(item, outcome.Result)
		}
		if err := store.WriteValidationAttempt(outcome.Attempt, item.Scenario.ID, item.ID, item.ExpectedScore, item.ExpectedFullSuccess, outcome.Result, caseErr); err != nil {
			writeErr = errors.Join(writeErr, err)
		}
		if caseErr != nil {
			logger.Error("validation case failed", "scenario", item.Scenario.ID, "case", item.ID, "error", caseErr)
			validationErrors = append(validationErrors, fmt.Errorf("%s/%s: %w", item.Scenario.ID, item.ID, caseErr))
			if _, err := fmt.Fprintf(output, "%s %s/%s attempt %03d: failed (score %.3f, full_success=%t)\n", metadata.RunID, item.Scenario.ID, item.ID, outcome.Attempt, outcome.Result.Grading.Score, outcome.Result.Grading.FullSuccess); err != nil {
				outputErr = errors.Join(outputErr, fmt.Errorf("write validation summary: %w", err))
			}
			continue
		}
		if _, err := fmt.Fprintf(output, "%s %s/%s attempt %03d: passed (score %.3f, full_success=%t)\n", metadata.RunID, item.Scenario.ID, item.ID, outcome.Attempt, outcome.Result.Grading.Score, outcome.Result.Grading.FullSuccess); err != nil {
			outputErr = errors.Join(outputErr, fmt.Errorf("write validation summary: %w", err))
		}
	}
	_ = <-executeErrors
	if writeErr != nil {
		return writeErr
	}
	if outputErr != nil {
		return outputErr
	}
	if err := store.Finalize(time.Now().UTC()); err != nil {
		return err
	}
	return errors.Join(validationErrors...)
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
