// Package orchestration drives one scenario run end-to-end: cluster and
// sandbox lifecycle, the model agent, and grading.
package orchestration

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	clusterintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/cluster"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

type Runner struct {
	ClusterFactory      clusterintegration.Factory
	SandboxFactory      sandboxintegration.Factory
	SandboxImageBuilder sandboxintegration.ImageBuilder
	SetupFactory        sandboxintegration.Factory
	SetupImageBuilder   sandboxintegration.ImageBuilder
	AgentFactory        rootagent.Factory
	AgentSlots          chan struct{} // Shared across benchmark workers; nil disables the limit.
	Condition           string
	Executor            command.Executor
	CleanupTimeout      time.Duration
}

const defaultCleanupTimeout = 30 * time.Second

// sandboxCommandExecutor adapts a sandbox executor to the command executor
// interface used by the setup phases and grading.
type sandboxCommandExecutor struct {
	executor sandboxintegration.Executor
}

func (e sandboxCommandExecutor) Run(ctx context.Context, spec command.Spec) (command.Result, error) {
	return e.executor.Exec(ctx, spec)
}

func (r Runner) Run(ctx context.Context, definition scenario.Definition) (RunResult, error) {
	condition := r.agentCondition()
	if r.AgentFactory == nil {
		return RunResult{Condition: condition}, fmt.Errorf("%s run requires a model agent factory", condition)
	}
	if r.SandboxFactory == nil {
		return RunResult{Condition: condition}, errors.New("model agent requires an orchestration sandbox")
	}
	return r.run(ctx, definition, condition, nil)
}

// RunWithRepair runs a scenario with a validation-only repair step after
// setup and optional fault verification, before grading. The repair is supplied
// by validation data, not by the scenario definition used for benchmark execution.
func (r Runner) RunWithRepair(ctx context.Context, definition scenario.Definition, repair scenario.Step) (RunResult, error) {
	if r.AgentFactory != nil {
		return RunResult{Condition: "validation"}, errors.New("validation run cannot use a model agent")
	}
	return r.run(ctx, definition, "validation", repair)
}

func (r Runner) run(ctx context.Context, definition scenario.Definition, condition string, repair scenario.Step) (result RunResult, runErr error) {
	result.Condition = condition
	if r.ClusterFactory == nil {
		return result, errors.New("orchestration cluster factory is required")
	}
	if r.Executor == nil {
		return result, errors.New("orchestration command executor is required")
	}
	if err := definition.Validate(); err != nil {
		return result, err
	}
	result.ScenarioID = definition.ID

	clusterName := clusterNameFor(definition.ID)
	logger := slog.With("scenario", definition.ID, "cluster", clusterName)
	images := []struct {
		label   string
		used    bool
		builder sandboxintegration.ImageBuilder
	}{
		{"sandbox", r.SandboxFactory != nil, r.SandboxImageBuilder},
		{"setup", r.SetupFactory != nil, r.SetupImageBuilder},
	}
	for _, image := range images {
		if !image.used || image.builder == nil {
			continue
		}
		if err := image.builder.Build(ctx); err != nil {
			logger.Error(image.label+" image build failed", "error", err)
			result.Failure = newFailureEvidence("build "+image.label+" image", err)
			return result, fmt.Errorf("build %s image: %w", image.label, err)
		}
	}
	cluster, err := create(logger, "cluster", func() (clusterintegration.Cluster, error) {
		return r.ClusterFactory(clusterName)
	})
	if err != nil {
		return result, err
	}
	defer r.deleteCluster(cluster, logger, &runErr, &result.Failure)

	if err := r.createCluster(ctx, cluster, clusterName, logger); err != nil {
		result.Failure = newFailureEvidence("create cluster", err)
		return result, err
	}
	kubeconfigPath := cluster.InternalKubeconfigPath()
	phaseExecutor := r.Executor
	if r.SetupFactory != nil {
		setup, err := startSandbox(ctx, logger, "setup", &result, func() (sandboxintegration.Sandbox, error) {
			return r.SetupFactory(clusterName, kubeconfigPath)
		})
		if err != nil {
			return result, err
		}
		defer r.stopSandbox(setup, "setup", logger, &runErr, &result.Failure)
		executor, ok := setup.(sandboxintegration.Executor)
		if !ok {
			return result, errors.New("setup sandbox does not provide command execution")
		}
		phaseExecutor = sandboxCommandExecutor{executor: executor}
	}
	var modelAgent rootagent.Agent
	if r.SandboxFactory != nil {
		sandbox, err := startSandbox(ctx, logger, "sandbox", &result, func() (sandboxintegration.Sandbox, error) {
			return r.SandboxFactory(clusterName, kubeconfigPath)
		})
		if err != nil {
			return result, err
		}
		defer r.stopSandbox(sandbox, "sandbox", logger, &runErr, &result.Failure)
		if r.AgentFactory != nil {
			executor, ok := sandbox.(sandboxintegration.Executor)
			if !ok {
				return result, errors.New("orchestration sandbox does not provide command execution")
			}
			modelAgent, err = create(logger, "model agent", func() (rootagent.Agent, error) {
				return r.AgentFactory(executor)
			})
			if err != nil {
				return result, err
			}
		}
	}
	grading, agentResult, err := r.runPhases(ctx, definition, repair, modelAgent, kubeconfigPath, phaseExecutor, logger)
	result.Agent = agentResult
	result.Grading = grading
	if err != nil && result.Failure == nil {
		result.Failure = newFailureEvidence("scenario", err)
	}
	return result, err
}

// joinCleanupError joins a cleanup error into runErr and, if no failure has
// been recorded yet for this run, records phase as the failing step.
func joinCleanupError(runErr *error, failure **FailureEvidence, phase string, err error) {
	*runErr = errors.Join(*runErr, err)
	if *failure == nil {
		*failure = newFailureEvidence(phase, err)
	}
}

func (r Runner) agentCondition() string {
	if condition := strings.TrimSpace(r.Condition); condition != "" {
		return condition
	}
	return "baseline"
}

// create calls a factory, wrapping its error and rejecting a nil result.
func create[T any](logger *slog.Logger, label string, factory func() (T, error)) (T, error) {
	value, err := factory()
	if err != nil {
		logger.Error(label+" factory failed", "error", err)
		return value, fmt.Errorf("create %s: %w", label, err)
	}
	if any(value) == nil {
		return value, fmt.Errorf("create %s: factory returned a nil %s", label, label)
	}
	return value, nil
}

func (r Runner) createCluster(ctx context.Context, cluster clusterintegration.Cluster, name string, logger *slog.Logger) error {
	if err := cluster.Create(ctx); err != nil {
		logger.Error("scenario cluster creation failed", "error", err)
		return fmt.Errorf("create cluster %q: %w", name, err)
	}
	logger.Info("scenario cluster created")
	return nil
}

// startSandbox creates and starts a container, recording a start failure as
// the run's failure.
func startSandbox(ctx context.Context, logger *slog.Logger, label string, result *RunResult, factory func() (sandboxintegration.Sandbox, error)) (sandboxintegration.Sandbox, error) {
	sandbox, err := create(logger, label, factory)
	if err != nil {
		return nil, err
	}
	if err := sandbox.Start(ctx); err != nil {
		logger.Error(label+" start failed", "error", err)
		err = fmt.Errorf("start %s: %w", label, err)
		result.Failure = newFailureEvidence("start "+label, err)
		return nil, err
	}
	logger.Info(label + " started")
	return sandbox, nil
}

// stopSandbox stops a container when the run ends and joins a failure into
// the run's error.
func (r Runner) stopSandbox(sandbox sandboxintegration.Sandbox, label string, logger *slog.Logger, runErr *error, failure **FailureEvidence) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), r.cleanupTimeout())
	defer cancel()
	logger.Info("stopping sandbox")
	if err := sandbox.Stop(cleanupCtx); err != nil {
		logger.Error("sandbox stop failed", "error", err)
		joinCleanupError(runErr, failure, "cleanup "+label, fmt.Errorf("stop sandbox: %w", err))
		return
	}
	logger.Info("sandbox stopped")
}

func (r Runner) deleteCluster(cluster clusterintegration.Cluster, logger *slog.Logger, runErr *error, failure **FailureEvidence) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), r.cleanupTimeout())
	defer cancel()
	logger.Info("deleting scenario cluster")
	if err := cluster.Delete(cleanupCtx); err != nil {
		logger.Error("scenario cluster deletion failed", "error", err)
		joinCleanupError(runErr, failure, "cleanup cluster", fmt.Errorf("delete cluster: %w", err))
		return
	}
	logger.Info("scenario cluster deleted")
}

func (r Runner) cleanupTimeout() time.Duration {
	if r.CleanupTimeout > 0 {
		return r.CleanupTimeout
	}
	return defaultCleanupTimeout
}

func clusterNameFor(scenarioID string) string {
	const prefix = "benchmark-"
	const maxNodeNameLength = 63
	const controlPlaneSuffix = "-control-plane"

	suffix := randomClusterSuffix()
	base := prefix + scenarioID
	maxClusterNameLength := maxNodeNameLength - len(controlPlaneSuffix)
	maxBaseLength := maxClusterNameLength - len(suffix) - 1
	if len(base) > maxBaseLength {
		base = strings.TrimRight(base[:maxBaseLength], "-")
	}
	return base + "-" + suffix
}

var clusterNameFallback atomic.Uint64

func randomClusterSuffix() string {
	var value [4]byte
	if _, err := crand.Read(value[:]); err == nil {
		return hex.EncodeToString(value[:])
	}
	return fmt.Sprintf("%08x", clusterNameFallback.Add(1))
}
