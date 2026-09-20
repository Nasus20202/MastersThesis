package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	commandagent "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/agent"
	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/config"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/ui"
	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/executor"
	clusterintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/cluster"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/cluster/kind"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
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

func runBenchmark(ctx context.Context, inputs []string, parallelism, repeat int, agentNames []commandagent.Name, benchmarkConfig benchmarkconfig.Config, terminal *ui.Terminal) error {
	definitions, err := scenario.LoadInputs(inputs)
	if err != nil {
		return err
	}
	if len(agentNames) == 0 {
		return errors.New("benchmark requires at least one agent")
	}
	llamaParallelism, err := commandagent.ConfiguredLlamaParallelism()
	if err != nil {
		return err
	}
	agentSlots := make(chan struct{}, llamaParallelism)
	totalTasks := len(definitions) * repeat * len(agentNames)

	startedAt := time.Now().UTC()
	revision, workingTreeDirty := repositoryProvenance()
	metadata := results.RunMetadata{
		RunID:              runID(startedAt),
		RunType:            "benchmark",
		StartedAt:          startedAt,
		RepositoryRevision: revision,
		WorkingTreeDirty:   workingTreeDirty,
		ExpectedAttempts:   totalTasks,
		Agents:             agentNamesToStrings(agentNames),
		Parallelism:        parallelism,
		RepeatCount:        repeat,
		Scenarios:          scenarioIDs(definitions),
	}
	store, err := results.New("results", metadata)
	if err != nil {
		return err
	}
	logger := slog.Default()
	finalized := false
	defer func() {
		if finalized {
			return
		}
		if err := store.Finalize(time.Now().UTC()); err != nil {
			logger.Error("benchmark result finalization failed", "run_id", metadata.RunID, "error", err)
		}
	}()
	logger.Info("benchmark runner started",
		"run_id", metadata.RunID,
		"agents", metadata.Agents,
		"scenarios", len(definitions),
		"parallelism", parallelism,
		"repeat_count", repeat,
		"total_tasks", totalTasks,
	)

	commandExecutor := command.LocalExecutor{Environment: os.Environ()}
	imageBuilder, err := newSandboxImageBuilder(commandExecutor)
	if err != nil {
		return err
	}
	agentFactories := make(map[commandagent.Name]rootagent.Factory, len(agentNames))
	for _, agentName := range agentNames {
		agentFactory, err := commandagent.NewFactory(agentName, benchmarkConfig)
		if err != nil {
			return err
		}
		agentFactories[agentName] = agentFactory
	}
	tasks := make([]executor.Task, 0, totalTasks)
	for _, agentName := range agentNames {
		agentFactory := agentFactories[agentName]
		for attempt := 1; attempt <= repeat; attempt++ {
			for _, definition := range definitions {
				definition := definition
				tasks = append(tasks, executor.Task{
					ScenarioID: definition.ID,
					Agent:      string(agentName),
					Attempt:    attempt,
					Run: func(ctx context.Context) (orchestration.RunResult, error) {
						return newOrchestrationRunner(commandExecutor, definition, imageBuilder, agentName, agentFactory, agentSlots).Run(ctx, definition)
					},
				})
			}
		}
	}
	progress, err := terminal.NewProgress(len(tasks), parallelism)
	if err != nil {
		return err
	}
	defer func() {
		if err := progress.Finish(); err != nil {
			logger.Error("benchmark progress display failed", "error", err)
		}
	}()
	outcomes, executeErrors := executor.Execute(ctx, tasks, parallelism)
	var writeErr error
	for outcome := range outcomes {
		if err := progress.Update(ui.Outcome{Success: outcome.Err == nil && outcome.Result.Grading.FullSuccess}); err != nil {
			logger.Error("benchmark progress display failed", "error", err)
		}
		if outcome.Err != nil {
			logger.Error("benchmark attempt failed",
				"run_id", metadata.RunID,
				"agent", outcome.Agent,
				"scenario", outcome.ScenarioID,
				"attempt", outcome.Attempt,
				"error", outcome.Err,
			)
			if err := store.WriteAttemptFailure(outcome.Attempt, outcome.ScenarioID, outcome.Agent, outcome.Result, outcome.Err); err != nil {
				writeErr = errors.Join(writeErr, err)
			}
			continue
		}
		if err := store.WriteAttempt(outcome.Attempt, outcome.Agent, outcome.Result); err != nil {
			writeErr = errors.Join(writeErr, err)
		}
		logger.Info("benchmark attempt completed",
			"run_id", metadata.RunID,
			"agent", outcome.Agent,
			"scenario", outcome.ScenarioID,
			"attempt", outcome.Attempt,
			"score", outcome.Result.Grading.Score,
			"full_success", outcome.Result.Grading.FullSuccess,
		)
	}
	runErr := <-executeErrors
	finalizeErr := store.Finalize(time.Now().UTC())
	if finalizeErr == nil {
		finalized = true
	}
	if writeErr != nil || finalizeErr != nil {
		return errors.Join(writeErr, finalizeErr)
	}
	return runErr
}

func runID(startedAt time.Time) string {
	timestamp := startedAt.UTC().Format("2006-01-02-15-04-05.000Z")
	return "run-" + strings.Replace(timestamp, ".", "-", 1)
}

func newOrchestrationRunner(commandExecutor command.Executor, definition scenario.Definition, imageBuilder sandboxintegration.ImageBuilder, agentName commandagent.Name, agentFactory rootagent.Factory, agentSlots chan struct{}) orchestration.Runner {
	return orchestration.Runner{
		Executor:            commandExecutor,
		SandboxImageBuilder: imageBuilder,
		AgentFactory:        agentFactory,
		AgentSlots:          agentSlots,
		Condition:           string(agentName),
		ClusterFactory: func(name string) (clusterintegration.Cluster, error) {
			return kind.New(commandExecutor, kind.Config{
				Name:       name,
				ConfigPath: definition.Cluster.Kind.ConfigPath(),
			})
		},
		SandboxFactory: func(name, kubeconfigPath string) (sandboxintegration.Sandbox, error) {
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

func newSandboxImageBuilder(commandExecutor command.Executor) (*docker.ImageBuilder, error) {
	return docker.NewImageBuilder(commandExecutor, docker.ImageConfig{
		Image:          sandboxImage,
		DockerfilePath: sandboxDockerfilePath,
		BuildContext:   sandboxBuildContext,
	})
}

func scenarioIDs(definitions []scenario.Definition) []string {
	ids := make([]string, len(definitions))
	for index, definition := range definitions {
		ids[index] = definition.ID
	}
	return ids
}

func agentNamesToStrings(names []commandagent.Name) []string {
	values := make([]string, len(names))
	for index, name := range names {
		values[index] = string(name)
	}
	return values
}
