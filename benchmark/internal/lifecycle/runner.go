package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strconv"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

type Runner struct {
	ClusterFactory ClusterFactory
	SandboxFactory SandboxFactory
	Executor       command.Executor
	CleanupTimeout time.Duration
}

const (
	defaultCleanupTimeout = 30 * time.Second
	kubeconfigEnv         = "KUBECONFIG"
	phasePrepare          = "prepare"
	phaseVerifyClean      = "verify clean"
	phaseInjectFault      = "inject fault"
	phaseVerifyFault      = "verify fault"
	phaseReset            = "reset"
)

type phase struct {
	name string
	step scenario.Step
}

func (r Runner) Run(ctx context.Context, definition scenario.Definition) (result RunResult, runErr error) {
	if r.ClusterFactory == nil {
		return RunResult{}, errors.New("lifecycle cluster factory is required")
	}
	if r.Executor == nil {
		return RunResult{}, errors.New("lifecycle command executor is required")
	}
	if err := definition.Validate(); err != nil {
		return RunResult{}, err
	}

	clusterName := clusterNameFor(definition.ID)
	logger := slog.With("scenario", definition.ID, "cluster", clusterName)
	cluster, err := r.newCluster(clusterName, logger)
	if err != nil {
		return RunResult{}, err
	}

	defer func() {
		if err := r.cleanupCluster(cluster, logger); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}()

	if err := r.createCluster(ctx, cluster, clusterName, logger); err != nil {
		return RunResult{}, err
	}
	if r.SandboxFactory != nil {
		sandbox, err := r.newSandbox(clusterName, cluster.InternalKubeconfigPath(), logger)
		if err != nil {
			return RunResult{}, err
		}
		if err := r.startSandbox(ctx, sandbox, logger); err != nil {
			return RunResult{}, err
		}
		defer func() {
			if err := r.cleanupSandbox(sandbox, logger); err != nil {
				runErr = errors.Join(runErr, err)
			}
		}()
	}
	grading, err := r.runPhases(ctx, definition, cluster.KubeconfigPath(), logger)
	return RunResult{ScenarioID: definition.ID, Grading: grading}, err
}

func (r Runner) newCluster(name string, logger *slog.Logger) (Cluster, error) {
	cluster, err := r.ClusterFactory(name)
	if err != nil {
		logger.Error("cluster factory failed", "error", err)
		return nil, fmt.Errorf("create cluster: %w", err)
	}
	if cluster == nil {
		return nil, errors.New("create cluster: factory returned a nil cluster")
	}
	return cluster, nil
}

func (r Runner) createCluster(ctx context.Context, cluster Cluster, name string, logger *slog.Logger) error {
	if err := cluster.Create(ctx); err != nil {
		logger.Error("scenario cluster creation failed", "error", err)
		return fmt.Errorf("create cluster %q: %w", name, err)
	}
	logger.Info("scenario cluster created")
	return nil
}

func (r Runner) newSandbox(name, kubeconfigPath string, logger *slog.Logger) (Sandbox, error) {
	sandbox, err := r.SandboxFactory(name, kubeconfigPath)
	if err != nil {
		logger.Error("sandbox factory failed", "error", err)
		return nil, fmt.Errorf("create sandbox: %w", err)
	}
	if sandbox == nil {
		return nil, errors.New("create sandbox: factory returned a nil sandbox")
	}
	return sandbox, nil
}

func (r Runner) startSandbox(ctx context.Context, sandbox Sandbox, logger *slog.Logger) error {
	if err := sandbox.Build(ctx); err != nil {
		logger.Error("sandbox image build failed", "error", err)
		return fmt.Errorf("build sandbox image: %w", err)
	}
	if err := sandbox.Start(ctx); err != nil {
		logger.Error("sandbox start failed", "error", err)
		return fmt.Errorf("start sandbox: %w", err)
	}
	logger.Info("sandbox started")
	return nil
}

func (r Runner) cleanupSandbox(sandbox Sandbox, logger *slog.Logger) error {
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

func (r Runner) cleanupCluster(cluster Cluster, logger *slog.Logger) error {
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

func (r Runner) runPhases(ctx context.Context, definition scenario.Definition, kubeconfigPath string, logger *slog.Logger) (GradingResult, error) {
	beforeAgent := []phase{
		{name: phasePrepare, step: definition.Prepare},
		{name: phaseVerifyClean, step: definition.VerifyClean},
		{name: phaseInjectFault, step: definition.InjectFault},
		{name: phaseVerifyFault, step: definition.VerifyFault},
	}
	if err := r.runPhaseSteps(ctx, beforeAgent, kubeconfigPath, logger); err != nil {
		return GradingResult{}, err
	}

	// Future agent repair execution belongs between fault verification and grading.
	logger.Info("running scenario grading")
	result, gradingErr := r.runGrading(ctx, definition.Grading, kubeconfigPath)
	if gradingErr != nil {
		logger.Error("scenario grading failed", "error", gradingErr)
	} else {
		logger.Info("scenario grading completed", "score", result.Score, "full_success", result.FullSuccess)
	}

	afterGrading := []phase{{name: phaseReset, step: definition.Reset}}
	if err := r.runPhaseSteps(ctx, afterGrading, kubeconfigPath, logger); err != nil {
		if gradingErr != nil {
			return GradingResult{}, errors.Join(gradingErr, err)
		}
		return result, err
	}
	if gradingErr != nil {
		return GradingResult{}, gradingErr
	}
	return result, nil
}

func (r Runner) runPhaseSteps(ctx context.Context, phases []phase, kubeconfigPath string, logger *slog.Logger) error {
	for _, currentPhase := range phases {
		logger.Info("running scenario phase", "phase", currentPhase.name)
		if err := r.runStep(ctx, currentPhase.name, currentPhase.step, kubeconfigPath); err != nil {
			logger.Error("scenario phase failed", "phase", currentPhase.name, "error", err)
			return err
		}
		logger.Info("scenario phase completed", "phase", currentPhase.name)
	}
	return nil
}

func (r Runner) runStep(ctx context.Context, phase string, step scenario.Step, kubeconfigPath string) error {
	for index, spec := range step.Specs() {
		spec = withKubeconfig(spec, kubeconfigPath)
		if _, err := r.Executor.Run(ctx, spec); err != nil {
			return fmt.Errorf("run %s command %d: %w", phase, index+1, err)
		}
	}
	return nil
}

func (r Runner) runGrading(ctx context.Context, criteria []scenario.Criterion, kubeconfigPath string) (GradingResult, error) {
	results := make([]CriterionResult, 0, len(criteria))
	for _, criterion := range criteria {
		result, err := r.Executor.Run(ctx, withKubeconfig(criterion.Check.Spec(), kubeconfigPath))
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
	return fmt.Sprintf("benchmark-%s-%s", scenarioID, strconv.FormatInt(time.Now().UnixNano(), 10))
}
