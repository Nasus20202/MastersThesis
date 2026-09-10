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

func (r Runner) Run(ctx context.Context, definition scenario.Definition) (runErr error) {
	if r.ClusterFactory == nil {
		return errors.New("lifecycle cluster factory is required")
	}
	if r.Executor == nil {
		return errors.New("lifecycle command executor is required")
	}
	if err := definition.Validate(); err != nil {
		return err
	}

	clusterName := clusterNameFor(definition.ID)
	logger := slog.With("scenario", definition.ID, "cluster", clusterName)
	cluster, err := r.newCluster(clusterName, logger)
	if err != nil {
		return err
	}

	defer func() {
		if err := r.cleanupCluster(cluster, logger); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}()

	if err := r.createCluster(ctx, cluster, clusterName, logger); err != nil {
		return err
	}
	return r.runPhases(ctx, definition, cluster.KubeconfigPath(), logger)
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

func (r Runner) runPhases(ctx context.Context, definition scenario.Definition, kubeconfigPath string, logger *slog.Logger) error {
	phases := []struct {
		name string
		step scenario.Step
	}{
		{name: phasePrepare, step: definition.Prepare},
		{name: phaseVerifyClean, step: definition.VerifyClean},
		{name: phaseInjectFault, step: definition.InjectFault},
		{name: phaseVerifyFault, step: definition.VerifyFault},
		{name: phaseReset, step: definition.Reset},
	}
	for _, phase := range phases {
		logger.Info("running scenario phase", "phase", phase.name)
		if err := r.runStep(ctx, phase.name, phase.step, kubeconfigPath); err != nil {
			logger.Error("scenario phase failed", "phase", phase.name, "error", err)
			return err
		}
		logger.Info("scenario phase completed", "phase", phase.name)
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
