package orchestration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

const (
	kubeconfigEnv    = "KUBECONFIG"
	phasePrepare     = "prepare"
	phaseVerifyClean = "verify clean"
	phaseInjectFault = "inject fault"
	phaseVerifyFault = "verify fault"
	phaseRepair      = "repair"
)

type phase struct {
	name string
	step scenario.Step
}

func (r Runner) runPhases(ctx context.Context, definition scenario.Definition, repair scenario.Step, modelAgent rootagent.Agent, kubeconfigPath string, phaseExecutor command.Executor, logger *slog.Logger) (GradingResult, *common.Result, error) {
	phases := []phase{
		{name: phasePrepare, step: definition.Prepare},
		{name: phaseVerifyClean, step: definition.VerifyClean},
	}
	if len(definition.InjectFault) > 0 {
		phases = append(phases, phase{name: phaseInjectFault, step: definition.InjectFault})
	}
	if len(definition.VerifyFault) > 0 {
		phases = append(phases, phase{name: phaseVerifyFault, step: definition.VerifyFault})
	}
	if len(repair) > 0 {
		phases = append(phases, phase{name: phaseRepair, step: repair})
	}
	if err := r.runPhaseSteps(ctx, phases, kubeconfigPath, phaseExecutor, logger); err != nil {
		return GradingResult{}, nil, err
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

	return result, agentResult, errors.Join(agentErr, gradingErr)
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
