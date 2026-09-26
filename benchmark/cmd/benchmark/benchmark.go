package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
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
	sandboxNameSuffix  = "-sandbox"
	setupNameSuffix    = "-setup"
	networkSuffix      = "-network"
	controlPlaneSuffix = "-control-plane"
)

func runBenchmark(ctx context.Context, inputs []string, parallelism, repeat int, agentNames []commandagent.Name, benchmarkConfig benchmarkconfig.Config, terminal *ui.Terminal, resumeID string) error {
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

	var store *results.Store
	if resumeID != "" {
		store, err = results.Resume("results", resumeID)
		if err == nil {
			if err := validateResumeMetadata(store.Metadata(), agentNames, repeat, scenarioIDs(definitions)); err != nil {
				return err
			}
		}
	} else {
		startedAt := time.Now().UTC()
		revision, workingTreeDirty := repositoryProvenance(ctx)
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
		store, err = results.New("results", metadata)
	}
	if err != nil {
		return err
	}
	metadata := store.Metadata()
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
	deps, err := newRunnerDeps(commandExecutor, benchmarkConfig.Containers)
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
				if store.HasAttempt(attempt, definition.ID, string(agentName)) {
					continue
				}
				tasks = append(tasks, executor.Task{
					ScenarioID: definition.ID,
					Agent:      string(agentName),
					Attempt:    attempt,
					Run: func(ctx context.Context) (orchestration.RunResult, error) {
						return newOrchestrationRunner(deps, definition, agentName, agentFactory, agentSlots).Run(ctx, definition)
					},
				})
			}
		}
	}
	if len(tasks) == 0 {
		finalizeErr := store.Finalize(time.Now().UTC())
		if finalizeErr == nil {
			finalized = true
		}
		return finalizeErr
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

func validateResumeMetadata(metadata results.RunMetadata, agents []commandagent.Name, repeat int, scenarios []string) error {
	if metadata.RunType != "benchmark" {
		return fmt.Errorf("run %q is not a benchmark run", metadata.RunID)
	}
	wantAgents := agentNamesToStrings(agents)
	if !slices.Equal(metadata.Agents, wantAgents) {
		return fmt.Errorf("resume agents %v do not match run agents %v", wantAgents, metadata.Agents)
	}
	if metadata.RepeatCount != repeat {
		return fmt.Errorf("resume repeat %d does not match run repeat %d", repeat, metadata.RepeatCount)
	}
	if !slices.Equal(metadata.Scenarios, scenarios) {
		return fmt.Errorf("resume scenarios %v do not match run scenarios %v", scenarios, metadata.Scenarios)
	}
	return nil
}

func runID(startedAt time.Time) string {
	timestamp := startedAt.UTC().Format("2006-01-02-15-04-05.000Z")
	return "run-" + strings.Replace(timestamp, ".", "-", 1)
}

type runnerDeps struct {
	executor            command.Executor
	sandboxImageBuilder sandboxintegration.ImageBuilder
	setupImageBuilder   sandboxintegration.ImageBuilder
	sandboxFactory      sandboxintegration.Factory
	setupFactory        sandboxintegration.Factory
}

func newRunnerDeps(executor command.Executor, containers benchmarkconfig.ContainersConfig) (runnerDeps, error) {
	sandboxImageBuilder, err := docker.NewImageBuilder(executor, docker.ImageConfig{
		Image:          containers.Sandbox.Image,
		DockerfilePath: containers.Sandbox.DockerfilePath,
		BuildContext:   containers.Sandbox.BuildContext,
	})
	if err != nil {
		return runnerDeps{}, err
	}
	setupImageBuilder, err := docker.NewImageBuilder(executor, docker.ImageConfig{
		Image:          containers.Setup.Image,
		DockerfilePath: containers.Setup.DockerfilePath,
		BuildContext:   containers.Setup.BuildContext,
	})
	if err != nil {
		return runnerDeps{}, err
	}
	repoRoot, err := os.Getwd()
	if err != nil {
		return runnerDeps{}, fmt.Errorf("resolve repository root: %w", err)
	}
	return runnerDeps{
		executor:            executor,
		sandboxImageBuilder: sandboxImageBuilder,
		setupImageBuilder:   setupImageBuilder,
		sandboxFactory: func(name, kubeconfigPath string) (sandboxintegration.Sandbox, error) {
			return docker.New(executor, docker.Config{
				Name:           name + sandboxNameSuffix,
				Image:          containers.Sandbox.Image,
				DockerfilePath: containers.Sandbox.DockerfilePath,
				BuildContext:   containers.Sandbox.BuildContext,
				KubeconfigPath: kubeconfigPath,
				Network:        name + sandboxNameSuffix + networkSuffix,
				NetworkTarget:  name + controlPlaneSuffix,
				Layout:         docker.DefaultImageLayout(),
				Security: docker.SecurityConfig{
					ReadOnlyRoot:     true,
					DropCapabilities: true,
					NoNewPrivileges:  true,
				},
			})
		},
		setupFactory: func(name, kubeconfigPath string) (sandboxintegration.Sandbox, error) {
			kubeconfigDir := filepath.Dir(kubeconfigPath)
			return docker.New(executor, docker.Config{
				Name:           name + setupNameSuffix,
				Image:          containers.Setup.Image,
				DockerfilePath: containers.Setup.DockerfilePath,
				BuildContext:   containers.Setup.BuildContext,
				Network:        containers.Setup.Network,
				Layout:         docker.ImageLayout{User: "root", Workdir: "/workspace"},
				Env:            map[string]string{"KUBECONFIG": kubeconfigPath},
				Mounts: []docker.Mount{
					{Source: repoRoot, Target: repoRoot, ReadOnly: true},
					{Source: kubeconfigDir, Target: kubeconfigDir},
				},
			})
		},
	}, nil
}

func newOrchestrationRunner(deps runnerDeps, definition scenario.Definition, agentName commandagent.Name, agentFactory rootagent.Factory, agentSlots chan struct{}) orchestration.Runner {
	return orchestration.Runner{
		Executor:            deps.executor,
		SandboxImageBuilder: deps.sandboxImageBuilder,
		SetupImageBuilder:   deps.setupImageBuilder,
		SandboxFactory:      deps.sandboxFactory,
		SetupFactory:        deps.setupFactory,
		AgentFactory:        agentFactory,
		AgentSlots:          agentSlots,
		Condition:           string(agentName),
		ClusterFactory: func(name string) (clusterintegration.Cluster, error) {
			return kind.New(deps.executor, kind.Config{
				Name:       name,
				ConfigPath: definition.Cluster.Kind.ConfigPath(),
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

func agentNamesToStrings(names []commandagent.Name) []string {
	values := make([]string, len(names))
	for index, name := range names {
		values[index] = string(name)
	}
	return values
}
