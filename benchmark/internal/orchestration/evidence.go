package orchestration

import (
	"errors"
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
)

type FailureEvidence struct {
	Phase           string   `json:"phase"`
	CommandIndex    int      `json:"command_index"`
	Program         string   `json:"program"`
	Args            []string `json:"args,omitempty"`
	Stdout          string   `json:"stdout"`
	Stderr          string   `json:"stderr"`
	ExitCode        int      `json:"exit_code"`
	DurationSeconds float64  `json:"duration_seconds"`
}

type stepError struct {
	phase        string
	commandIndex int
	spec         command.Spec
	result       command.Result
	err          error
}

func (e *stepError) Error() string {
	return e.err.Error()
}

func (e *stepError) Unwrap() error {
	return e.err
}

func newFailureEvidence(fallbackPhase string, err error) *FailureEvidence {
	var failedStep *stepError
	if errors.As(err, &failedStep) {
		return failureEvidence(failedStep.phase, failedStep.commandIndex, failedStep.spec, failedStep.result)
	}

	var executionErr *command.ExecutionError
	if errors.As(err, &executionErr) {
		return failureEvidence(fallbackPhase, 1, executionErr.Spec, executionErr.Result)
	}
	return nil
}

func failureEvidence(phase string, commandIndex int, spec command.Spec, result command.Result) *FailureEvidence {
	return &FailureEvidence{
		Phase:           phase,
		CommandIndex:    commandIndex,
		Program:         spec.Program,
		Args:            slices.Clone(spec.Args),
		Stdout:          result.Stdout,
		Stderr:          result.Stderr,
		ExitCode:        result.ExitCode,
		DurationSeconds: result.Duration.Seconds(),
	}
}
