package orchestration

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strings"
	"sync/atomic"
	"time"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
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

const (
	defaultCleanupTimeout = 30 * time.Second
	kubeconfigEnv         = "KUBECONFIG"
	phasePrepare          = "prepare"
	phaseVerifyClean      = "verify clean"
	phaseInjectFault      = "inject fault"
	phaseVerifyFault      = "verify fault"
	phaseRepair           = "repair"
)

type phase struct {
	name string
	step scenario.Step
}

// sandboxCommandExecutor adapts a sandbox executor to the command executor
// interface used by the setup phases and grading.
type sandboxCommandExecutor struct {
	executor sandboxintegration.Executor
}

func (e sandboxCommandExecutor) Run(ctx context.Context, spec command.Spec) (command.Result, error) {
	return e.executor.Exec(ctx, spec)
}

func (r Runner) Run(ctx context.Context, definition scenario.Definition) (result RunResult, runErr error) {
	return r.run(ctx, definition, nil, true)
}

// RunWithRepair runs a scenario with a validation-only repair step after
// setup and optional fault verification, before grading. The repair is supplied
// by validation data, not by the scenario definition used for benchmark execution.
func (r Runner) RunWithRepair(ctx context.Context, definition scenario.Definition, repair scenario.Step) (RunResult, error) {
	return r.run(ctx, definition, repair, false)
}

func (r Runner) run(ctx context.Context, definition scenario.Definition, repair scenario.Step, useAgent bool) (result RunResult, runErr error) {
	if useAgent {
		result.Condition = r.agentCondition()
	} else {
		result.Condition = "validation"
	}
	if r.ClusterFactory == nil {
		return result, errors.New("orchestration cluster factory is required")
	}
	if r.Executor == nil {
		return result, errors.New("orchestration command executor is required")
	}
	if !useAgent && r.AgentFactory != nil {
		return result, errors.New("validation run cannot use a model agent")
	}
	if useAgent && r.AgentFactory == nil {
		return result, fmt.Errorf("%s run requires a model agent factory", result.Condition)
	}
	if useAgent && r.SandboxFactory == nil {
		return result, errors.New("model agent requires an orchestration sandbox")
	}
	if err := definition.Validate(); err != nil {
		return result, err
	}
	result.ScenarioID = definition.ID

	clusterName := clusterNameFor(definition.ID)
	logger := slog.With("scenario", definition.ID, "cluster", clusterName)
	if r.SandboxFactory != nil && r.SandboxImageBuilder != nil {
		if err := r.SandboxImageBuilder.Build(ctx); err != nil {
			logger.Error("sandbox image build failed", "error", err)
			result.Failure = newFailureEvidence("build sandbox image", err)
			return result, fmt.Errorf("build sandbox image: %w", err)
		}
	}
	if r.SetupFactory != nil && r.SetupImageBuilder != nil {
		if err := r.SetupImageBuilder.Build(ctx); err != nil {
			logger.Error("setup image build failed", "error", err)
			result.Failure = newFailureEvidence("build setup image", err)
			return result, fmt.Errorf("build setup image: %w", err)
		}
	}
	cluster, err := r.newCluster(clusterName, logger)
	if err != nil {
		return result, err
	}

	defer func() {
		if err := r.cleanupCluster(cluster, logger); err != nil {
			joinCleanupError(&runErr, &result.Failure, "cleanup cluster", err)
		}
	}()

	if err := r.createCluster(ctx, cluster, clusterName, logger); err != nil {
		result.Failure = newFailureEvidence("create cluster", err)
		return result, err
	}
	phaseExecutor := r.Executor
	if r.SetupFactory != nil {
		setup, err := r.newSetup(clusterName, cluster.InternalKubeconfigPath(), logger)
		if err != nil {
			return result, err
		}
		if err := r.startSetup(ctx, setup, logger); err != nil {
			result.Failure = newFailureEvidence("start setup", err)
			return result, err
		}
		defer func() {
			if err := r.cleanupSandbox(setup, logger); err != nil {
				joinCleanupError(&runErr, &result.Failure, "cleanup setup", err)
			}
		}()
		executor, ok := setup.(sandboxintegration.Executor)
		if !ok {
			return result, errors.New("setup sandbox does not provide command execution")
		}
		phaseExecutor = sandboxCommandExecutor{executor: executor}
	}
	var modelAgent rootagent.Agent
	if r.SandboxFactory != nil {
		sandbox, err := r.newSandbox(clusterName, cluster.InternalKubeconfigPath(), logger)
		if err != nil {
			return result, err
		}
		if err := r.startSandbox(ctx, sandbox, logger); err != nil {
			result.Failure = newFailureEvidence("start sandbox", err)
			return result, err
		}
		defer func() {
			if err := r.cleanupSandbox(sandbox, logger); err != nil {
				joinCleanupError(&runErr, &result.Failure, "cleanup sandbox", err)
			}
		}()
		if useAgent && r.AgentFactory != nil {
			executor, ok := sandbox.(sandboxintegration.Executor)
			if !ok {
				return result, errors.New("orchestration sandbox does not provide command execution")
			}
			modelAgent, err = r.newAgent(executor, logger)
			if err != nil {
				return result, err
			}
		}
	}
	grading, agentResult, err := r.runPhases(ctx, definition, repair, modelAgent, cluster.InternalKubeconfigPath(), phaseExecutor, logger)
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

// wrapFactory calls a component factory, logs and wraps a returned error
// under errLabel, and turns a nil result into an error, so every factory
// wrapper below only has to name its logger prefix and error label.
func wrapFactory[T any](logger *slog.Logger, logLabel, errLabel string, call func() (T, error)) (T, error) {
	value, err := call()
	if err != nil {
		logger.Error(logLabel+" factory failed", "error", err)
		return value, fmt.Errorf("create %s: %w", errLabel, err)
	}
	if any(value) == nil {
		return value, fmt.Errorf("create %s: factory returned a nil %s", errLabel, errLabel)
	}
	return value, nil
}

func (r Runner) newAgent(executor sandboxintegration.Executor, logger *slog.Logger) (rootagent.Agent, error) {
	return wrapFactory(logger, "agent", "model agent", func() (rootagent.Agent, error) {
		return r.AgentFactory(executor)
	})
}

func (r Runner) newCluster(name string, logger *slog.Logger) (clusterintegration.Cluster, error) {
	return wrapFactory(logger, "cluster", "cluster", func() (clusterintegration.Cluster, error) {
		return r.ClusterFactory(name)
	})
}

func (r Runner) createCluster(ctx context.Context, cluster clusterintegration.Cluster, name string, logger *slog.Logger) error {
	if err := cluster.Create(ctx); err != nil {
		logger.Error("scenario cluster creation failed", "error", err)
		return fmt.Errorf("create cluster %q: %w", name, err)
	}
	logger.Info("scenario cluster created")
	return nil
}

func (r Runner) newSandbox(name, kubeconfigPath string, logger *slog.Logger) (sandboxintegration.Sandbox, error) {
	return wrapFactory(logger, "sandbox", "sandbox", func() (sandboxintegration.Sandbox, error) {
		return r.SandboxFactory(name, kubeconfigPath)
	})
}

func (r Runner) newSetup(name, kubeconfigPath string, logger *slog.Logger) (sandboxintegration.Sandbox, error) {
	return wrapFactory(logger, "setup", "setup", func() (sandboxintegration.Sandbox, error) {
		return r.SetupFactory(name, kubeconfigPath)
	})
}

// startSandboxLike builds sandbox (if no imageBuilder was supplied, meaning no
// image was pre-built for it) and starts it, logging under label ("setup" or
// "sandbox"). Shared by startSetup and startSandbox, which only differ in
// which sandbox/image-builder pair they operate on.
func startSandboxLike(ctx context.Context, sandbox sandboxintegration.Sandbox, imageBuilder sandboxintegration.ImageBuilder, label string, logger *slog.Logger) error {
	if imageBuilder == nil {
		if err := sandbox.Build(ctx); err != nil {
			logger.Error(label+" image build failed", "error", err)
			return fmt.Errorf("build %s image: %w", label, err)
		}
	}
	if err := sandbox.Start(ctx); err != nil {
		logger.Error(label+" start failed", "error", err)
		return fmt.Errorf("start %s: %w", label, err)
	}
	logger.Info(label + " started")
	return nil
}

func (r Runner) startSetup(ctx context.Context, setup sandboxintegration.Sandbox, logger *slog.Logger) error {
	return startSandboxLike(ctx, setup, r.SetupImageBuilder, "setup", logger)
}

func (r Runner) startSandbox(ctx context.Context, sandbox sandboxintegration.Sandbox, logger *slog.Logger) error {
	return startSandboxLike(ctx, sandbox, r.SandboxImageBuilder, "sandbox", logger)
}

func (r Runner) cleanupSandbox(sandbox sandboxintegration.Sandbox, logger *slog.Logger) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), r.cleanupTimeout())
	defer cancel()
	logger.Info("stopping sandbox")
	if err := sandbox.Stop(cleanupCtx); err != nil {
		logger.Error("sandbox stop failed", "error", err)
		return fmt.Errorf("stop sandbox: %w", err)
	}
	logger.Info("sandbox stopped")
	return nil
}

func (r Runner) cleanupCluster(cluster clusterintegration.Cluster, logger *slog.Logger) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), r.cleanupTimeout())
	defer cancel()
	logger.Info("deleting scenario cluster")
	if err := cluster.Delete(cleanupCtx); err != nil {
		logger.Error("scenario cluster deletion failed", "error", err)
		return fmt.Errorf("delete cluster: %w", err)
	}
	logger.Info("scenario cluster deleted")
	return nil
}

func (r Runner) runPhases(ctx context.Context, definition scenario.Definition, repair scenario.Step, modelAgent rootagent.Agent, kubeconfigPath string, phaseExecutor command.Executor, logger *slog.Logger) (GradingResult, *common.Result, error) {
	beforeAgent := []phase{
		{name: phasePrepare, step: definition.Prepare},
		{name: phaseVerifyClean, step: definition.VerifyClean},
	}
	if len(definition.InjectFault) > 0 {
		beforeAgent = append(beforeAgent, phase{name: phaseInjectFault, step: definition.InjectFault})
	}
	if len(definition.VerifyFault) > 0 {
		beforeAgent = append(beforeAgent, phase{name: phaseVerifyFault, step: definition.VerifyFault})
	}
	if err := r.runPhaseSteps(ctx, beforeAgent, kubeconfigPath, phaseExecutor, logger); err != nil {
		return GradingResult{}, nil, err
	}
	if len(repair) > 0 {
		if err := r.runPhaseSteps(ctx, []phase{{name: phaseRepair, step: repair}}, kubeconfigPath, phaseExecutor, logger); err != nil {
			return GradingResult{}, nil, err
		}
	}

	var agentResult *common.Result
	var agentErr error
	if modelAgent != nil {
		logger.Info("running model agent")
		result, err := r.runModelAgent(ctx, modelAgent, definition.Task)
		agentResult = &result
		if err != nil {
			agentErr = fmt.Errorf("run model agent: %w", err)
			logger.Error("model agent failed", "error", err)
		} else {
			logger.Info("model agent completed", "termination", result.Termination)
		}
	}

	logger.Info("running scenario grading")
	result, gradingErr := r.runGrading(ctx, definition.Grading, kubeconfigPath, phaseExecutor)
	if gradingErr != nil {
		logger.Error("scenario grading failed", "error", gradingErr)
	} else {
		logger.Info("scenario grading completed", "score", result.Score, "full_success", result.FullSuccess)
	}

	if runErr := errors.Join(agentErr, gradingErr); runErr != nil {
		return result, agentResult, runErr
	}
	return result, agentResult, nil
}

func (r Runner) runModelAgent(ctx context.Context, modelAgent rootagent.Agent, task string) (common.Result, error) {
	if r.AgentSlots != nil {
		if err := ctx.Err(); err != nil {
			return common.Result{}, err
		}
		select {
		case r.AgentSlots <- struct{}{}:
			defer func() { <-r.AgentSlots }()
		case <-ctx.Done():
			return common.Result{}, ctx.Err()
		}
	}
	return modelAgent.Run(ctx, task)
}

func (r Runner) runPhaseSteps(ctx context.Context, phases []phase, kubeconfigPath string, executor command.Executor, logger *slog.Logger) error {
	for _, currentPhase := range phases {
		logger.Info("running scenario phase", "phase", currentPhase.name)
		if err := r.runStep(ctx, currentPhase.name, currentPhase.step, kubeconfigPath, executor); err != nil {
			logger.Error("scenario phase failed", "phase", currentPhase.name, "error", err)
			return err
		}
		logger.Info("scenario phase completed", "phase", currentPhase.name)
	}
	return nil
}

func (r Runner) runStep(ctx context.Context, phase string, step scenario.Step, kubeconfigPath string, executor command.Executor) error {
	for index, spec := range step.Specs() {
		spec = withKubeconfig(spec, kubeconfigPath)
		result, err := executor.Run(ctx, spec)
		if err != nil {
			return &stepError{
				phase:        phase,
				commandIndex: index + 1,
				spec:         spec,
				result:       result,
				err:          fmt.Errorf("run %s command %d: %w", phase, index+1, err),
			}
		}
	}
	return nil
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

func withKubeconfig(spec command.Spec, kubeconfigPath string) command.Spec {
	if kubeconfigPath == "" {
		return spec
	}

	env := maps.Clone(spec.Env)
	if env == nil {
		env = make(map[string]string)
	}
	env[kubeconfigEnv] = kubeconfigPath
	spec.Env = env
	return spec
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
