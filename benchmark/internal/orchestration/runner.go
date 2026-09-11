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

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	clusterintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/cluster"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

type Runner struct {
	ClusterFactory      clusterintegration.Factory
	SandboxFactory      sandboxintegration.Factory
	SandboxImageBuilder sandboxintegration.ImageBuilder
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
	phaseReset            = "reset"
)

type phase struct {
	name string
	step scenario.Step
}

func (r Runner) Run(ctx context.Context, definition scenario.Definition) (result RunResult, runErr error) {
	return r.run(ctx, definition, nil)
}

// RunWithRepair runs a scenario with a validation-only repair step between
// fault verification and grading. The repair is supplied by validation data,
// not by the scenario definition used for benchmark execution.
func (r Runner) RunWithRepair(ctx context.Context, definition scenario.Definition, repair scenario.Step) (RunResult, error) {
	return r.run(ctx, definition, repair)
}

func (r Runner) run(ctx context.Context, definition scenario.Definition, repair scenario.Step) (result RunResult, runErr error) {
	if r.ClusterFactory == nil {
		return RunResult{}, errors.New("orchestration cluster factory is required")
	}
	if r.Executor == nil {
		return RunResult{}, errors.New("orchestration command executor is required")
	}
	if err := definition.Validate(); err != nil {
		return RunResult{}, err
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
	cluster, err := r.newCluster(clusterName, logger)
	if err != nil {
		return RunResult{}, err
	}

	defer func() {
		if err := r.cleanupCluster(cluster, logger); err != nil {
			runErr = errors.Join(runErr, err)
			if result.Failure == nil {
				result.Failure = newFailureEvidence("cleanup cluster", err)
			}
		}
	}()

	if err := r.createCluster(ctx, cluster, clusterName, logger); err != nil {
		result.Failure = newFailureEvidence("create cluster", err)
		return result, err
	}
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
				runErr = errors.Join(runErr, err)
				if result.Failure == nil {
					result.Failure = newFailureEvidence("cleanup sandbox", err)
				}
			}
		}()
	}
	grading, err := r.runPhases(ctx, definition, repair, cluster.KubeconfigPath(), logger)
	result.Grading = grading
	if err != nil && result.Failure == nil {
		result.Failure = newFailureEvidence("scenario", err)
	}
	return result, err
}

func (r Runner) newCluster(name string, logger *slog.Logger) (clusterintegration.Cluster, error) {
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

func (r Runner) createCluster(ctx context.Context, cluster clusterintegration.Cluster, name string, logger *slog.Logger) error {
	if err := cluster.Create(ctx); err != nil {
		logger.Error("scenario cluster creation failed", "error", err)
		return fmt.Errorf("create cluster %q: %w", name, err)
	}
	logger.Info("scenario cluster created")
	return nil
}

func (r Runner) newSandbox(name, kubeconfigPath string, logger *slog.Logger) (sandboxintegration.Sandbox, error) {
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

func (r Runner) startSandbox(ctx context.Context, sandbox sandboxintegration.Sandbox, logger *slog.Logger) error {
	if r.SandboxImageBuilder == nil {
		if err := sandbox.Build(ctx); err != nil {
			logger.Error("sandbox image build failed", "error", err)
			return fmt.Errorf("build sandbox image: %w", err)
		}
	}
	if err := sandbox.Start(ctx); err != nil {
		logger.Error("sandbox start failed", "error", err)
		return fmt.Errorf("start sandbox: %w", err)
	}
	logger.Info("sandbox started")
	return nil
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

func (r Runner) runPhases(ctx context.Context, definition scenario.Definition, repair scenario.Step, kubeconfigPath string, logger *slog.Logger) (GradingResult, error) {
	beforeAgent := []phase{
		{name: phasePrepare, step: definition.Prepare},
		{name: phaseVerifyClean, step: definition.VerifyClean},
		{name: phaseInjectFault, step: definition.InjectFault},
		{name: phaseVerifyFault, step: definition.VerifyFault},
	}
	if err := r.runPhaseSteps(ctx, beforeAgent, kubeconfigPath, logger); err != nil {
		return GradingResult{}, err
	}
	if len(repair) > 0 {
		if err := r.runPhaseSteps(ctx, []phase{{name: phaseRepair, step: repair}}, kubeconfigPath, logger); err != nil {
			return GradingResult{}, err
		}
	}

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
		result, err := r.Executor.Run(ctx, spec)
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
