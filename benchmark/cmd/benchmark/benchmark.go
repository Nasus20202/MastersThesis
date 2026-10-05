package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	commandagent "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/agent"
	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/llamaenv"
	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	ragagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/rag"
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

func runBenchmark(ctx context.Context, options runOptions, inputs []string, agentNames []commandagent.Name, resumeID string) error {
	definitions, err := scenario.LoadInputs(inputs)
	if err != nil {
		return err
	}
	definitions = scenario.FilterByTags(definitions, options.tags)
	if len(definitions) == 0 {
		return fmt.Errorf("no scenarios match tag filter %q", options.tags.String())
	}
	if len(agentNames) == 0 {
		return errors.New("benchmark requires at least one agent")
	}
	llamaParallelism, err := llamaenv.Parallelism()
	if err != nil {
		return err
	}
	agentSlots := make(chan struct{}, llamaParallelism)
	repeat := options.repeat
	totalTasks := len(definitions) * repeat * len(agentNames)

	deps, err := newRunnerDeps(options.config.Containers)
	if err != nil {
		return err
	}
	sampling, err := readSampling(ctx)
	if err != nil {
		return err
	}
	var search *ragagent.Search
	var retrievalProvenance *results.RetrievalProvenance
	if slices.Contains(agentNames, commandagent.RAG) {
		var provenance results.RetrievalProvenance
		search, provenance, err = commandagent.OpenSearch(ctx, options.config.Retrieval)
		if err != nil {
			return err
		}
		defer search.Index.Close()
		retrievalProvenance = &provenance
	}
	agentFactories := make(map[commandagent.Name]rootagent.Factory, len(agentNames))
	for _, agentName := range agentNames {
		agentFactory, err := commandagent.NewFactory(agentName, options.config, search)
		if err != nil {
			return err
		}
		agentFactories[agentName] = agentFactory
	}

	var store *results.Store
	if resumeID != "" {
		store, err = results.Resume(resultsDir, resumeID)
		if err == nil {
			err = validateResumeMetadata(store.Metadata(), agentNames, repeat, scenarioIDs(definitions), retrievalProvenance, sampling)
		}
	} else {
		store, err = startRun(ctx, results.RunMetadata{
			RunType:          results.RunTypeBenchmark,
			ExpectedAttempts: totalTasks,
			Agents:           agentNamesToStrings(agentNames),
			Parallelism:      options.parallel,
			RepeatCount:      repeat,
			Scenarios:        scenarioIDs(definitions),
			TagSelector:      options.tags.String(),
			Retrieval:        retrievalProvenance,
			Sampling:         sampling,
		})
	}
	if err != nil {
		return err
	}
	metadata := store.Metadata()
	logger := slog.Default()
	logger.Info("benchmark runner started",
		"run_id", metadata.RunID,
		"agents", metadata.Agents,
		"scenarios", len(definitions),
		"parallelism", options.parallel,
		"repeat_count", repeat,
		"total_tasks", totalTasks,
	)

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
	return options.execute(ctx, store, tasks, func(outcome executor.Outcome) (bool, error) {
		if outcome.Err != nil {
			logger.Error("benchmark attempt failed",
				"run_id", metadata.RunID,
				"agent", outcome.Agent,
				"scenario", outcome.ScenarioID,
				"attempt", outcome.Attempt,
				"error", outcome.Err,
			)
			return false, store.WriteAttemptFailure(outcome.Attempt, outcome.ScenarioID, outcome.Agent, outcome.Result, outcome.Err)
		}
		logger.Info("benchmark attempt completed",
			"run_id", metadata.RunID,
			"agent", outcome.Agent,
			"scenario", outcome.ScenarioID,
			"attempt", outcome.Attempt,
			"score", outcome.Result.Grading.Score,
			"full_success", outcome.Result.Grading.FullSuccess,
		)
		return outcome.Result.Grading.FullSuccess, store.WriteAttempt(outcome.Attempt, outcome.Agent, outcome.Result)
	})
}

func validateResumeMetadata(metadata results.RunMetadata, agents []commandagent.Name, repeat int, scenarios []string, retrieval *results.RetrievalProvenance, sampling *results.SamplingProvenance) error {
	if metadata.RunType != results.RunTypeBenchmark {
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
	if !equalPointers(metadata.Retrieval, retrieval) {
		return fmt.Errorf("resume retrieval %+v does not match run retrieval %+v", retrieval, metadata.Retrieval)
	}
	if !equalPointers(metadata.Sampling, sampling) {
		return fmt.Errorf("resume sampling %+v does not match run sampling %+v", sampling, metadata.Sampling)
	}
	return nil
}

// readSampling records the sampler settings the inference server applies to
// the profile's model. It also fails the run early when the server is down.
func readSampling(ctx context.Context) (*results.SamplingProvenance, error) {
	client, err := llamaenv.NewChatClient()
	if err != nil {
		return nil, err
	}
	sampling, err := client.Sampling(ctx)
	if err != nil {
		return nil, fmt.Errorf("read sampling settings: %w; start the server with make llama-start", err)
	}
	return &results.SamplingProvenance{Model: client.Metadata().Model, Sampling: sampling}, nil
}

func equalPointers[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

type runnerDeps struct {
	executor            command.Executor
	sandboxImageBuilder sandboxintegration.ImageBuilder
	setupImageBuilder   sandboxintegration.ImageBuilder
	sandboxFactory      sandboxintegration.Factory
	setupFactory        sandboxintegration.Factory
}

func newRunnerDeps(containers benchmarkconfig.ContainersConfig) (runnerDeps, error) {
	executor := command.LocalExecutor{Environment: os.Environ()}
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
				Name:    name + setupNameSuffix,
				Image:   containers.Setup.Image,
				Network: containers.Setup.Network,
				Layout:  docker.ImageLayout{User: "root", Workdir: "/workspace"},
				Env:     map[string]string{"KUBECONFIG": kubeconfigPath},
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
