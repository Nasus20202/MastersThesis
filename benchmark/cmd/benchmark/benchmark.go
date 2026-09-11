package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/executor"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/cluster/kind"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox/docker"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

const (
	sandboxImage          = "benchmark-sandbox:ubuntu-26.04"
	sandboxDockerfilePath = "sandbox/Dockerfile.ubuntu-26.04"
	sandboxBuildContext   = "sandbox"
)

func runBenchmark(ctx context.Context, inputs []string, parallelism, repeat int, output io.Writer) error {
	definitions, err := scenario.LoadInputs(inputs)
	if err != nil {
		return err
	}
	startedAt := time.Now().UTC()
	metadata := results.RunMetadata{
		RunID:       runID(startedAt),
		StartedAt:   startedAt,
		Parallelism: parallelism,
		RepeatCount: repeat,
		Scenarios:   scenarioIDs(definitions),
	}
	store, err := results.New("results", metadata)
	if err != nil {
		return err
	}
	logger := slog.Default()
	logger.Info("benchmark runner started",
		"run_id", metadata.RunID,
		"scenarios", len(definitions),
		"parallelism", parallelism,
		"repeat_count", repeat,
		"total_tasks", len(definitions)*repeat,
	)

	commandExecutor := command.LocalExecutor{}
	tasks := make([]executor.Task, 0, len(definitions)*repeat)
	for attempt := 1; attempt <= repeat; attempt++ {
		for _, definition := range definitions {
			definition := definition
			tasks = append(tasks, executor.Task{
				ScenarioID: definition.ID,
				Attempt:    attempt,
				Run: func(ctx context.Context) (orchestration.RunResult, error) {
					return newOrchestrationRunner(commandExecutor, definition).Run(ctx, definition)
				},
			})
		}
	}
	outcomes, runErr := executor.Execute(ctx, tasks, parallelism)
	for _, outcome := range outcomes {
		if outcome.Err != nil {
			logger.Error("benchmark task failed", "scenario", outcome.ScenarioID, "error", outcome.Err)
			if err := store.WriteAttemptFailure(outcome.Attempt, outcome.ScenarioID, outcome.Result, outcome.Err); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(output, "%s %s attempt %03d: failed\n", metadata.RunID, outcome.ScenarioID, outcome.Attempt); err != nil {
				return fmt.Errorf("write benchmark summary: %w", err)
			}
			continue
		}
		if err := store.WriteAttempt(outcome.Attempt, outcome.Result); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(output, "%s %s attempt %03d: score %.3f, full_success=%t\n", metadata.RunID, outcome.ScenarioID, outcome.Attempt, outcome.Result.Grading.Score, outcome.Result.Grading.FullSuccess); err != nil {
			return fmt.Errorf("write benchmark summary: %w", err)
		}
	}
	if err := store.Finalize(time.Now().UTC()); err != nil {
		return err
	}
	return runErr
}

func runID(startedAt time.Time) string {
	timestamp := startedAt.UTC().Format("2006-01-02-15-04-05.000Z")
	return "run-" + strings.Replace(timestamp, ".", "-", 1)
}

func newOrchestrationRunner(commandExecutor command.Executor, definition scenario.Definition) orchestration.Runner {
	return orchestration.Runner{
		Executor: commandExecutor,
		ClusterFactory: func(name string) (orchestration.Cluster, error) {
			return kind.New(commandExecutor, kind.Config{
				Name:       name,
				ConfigPath: definition.Cluster.Kind.ConfigPath(),
			})
		},
		SandboxFactory: func(name, kubeconfigPath string) (orchestration.Sandbox, error) {
			return docker.New(commandExecutor, docker.Config{
				Name:           name + "-sandbox",
				Image:          sandboxImage,
				DockerfilePath: sandboxDockerfilePath,
				BuildContext:   sandboxBuildContext,
				KubeconfigPath: kubeconfigPath,
				Network:        name + "-sandbox-network",
				NetworkTarget:  name + "-control-plane",
				Layout:         docker.DefaultImageLayout(),
			})
		},
	}
}

func scenarioIDs(definitions []scenario.Definition) []string {
	ids := make([]string, len(definitions))
	for index, definition := range definitions {
		ids[index] = definition.ID
	}
	return ids
}
