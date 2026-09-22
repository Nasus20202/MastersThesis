package orchestration

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentSlotsLimitAgentPhaseAndLeaveSetupAndGradingParallel(t *testing.T) {
	const workers = 4
	const agentLimit = 2

	prepareStarted := make(chan struct{}, workers)
	allowPrepare := make(chan struct{})
	gradingStarted := make(chan struct{}, workers)
	allowGrading := make(chan struct{})
	agentStarted := make(chan struct{}, workers)
	releaseAgent := make(chan struct{}, workers)
	executor := slotTestExecutor{
		prepareStarted: prepareStarted,
		allowPrepare:   allowPrepare,
		gradingStarted: gradingStarted,
		allowGrading:   allowGrading,
	}
	modelAgent := &slotTrackingAgent{started: agentStarted, release: releaseAgent}
	runner := Runner{Executor: executor, AgentSlots: make(chan struct{}, agentLimit)}
	definition := scenario.Definition{
		Task:    "repair the workload",
		Prepare: scenario.Step{{Program: "prepare"}},
		Grading: []scenario.Criterion{{
			ID: "healthy", Weight: 1, Check: scenario.Command{Program: "grade"},
		}},
	}
	done := make(chan error, workers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for range workers {
		go func() {
			_, _, err := runner.runPhases(ctx, definition, nil, modelAgent, "", executor, slog.Default())
			done <- err
		}()
	}

	for range workers {
		waitForSignal(t, prepareStarted, "all workers to prepare concurrently")
	}
	assert.Empty(t, agentStarted, "no agent should start before setup is released")
	close(allowPrepare)
	for range agentLimit {
		waitForSignal(t, agentStarted, "configured agent slots")
	}
	select {
	case <-agentStarted:
		t.Fatal("more model agents started than the configured limit")
	default:
	}

	// The first completed agent must release its slot before its grading step
	// finishes, allowing a third worker to enter the model-agent phase.
	releaseAgent <- struct{}{}
	waitForSignal(t, gradingStarted, "first grading step")
	waitForSignal(t, agentStarted, "third model agent while grading is blocked")

	releaseAgent <- struct{}{}
	waitForSignal(t, gradingStarted, "second grading step")
	waitForSignal(t, agentStarted, "fourth model agent")
	assert.LessOrEqual(t, modelAgent.maximum.Load(), int32(agentLimit))
	assert.Equal(t, int32(agentLimit), modelAgent.maximum.Load())

	close(allowGrading)
	for range workers - 2 {
		releaseAgent <- struct{}{}
	}
	for range workers {
		require.NoError(t, waitForError(t, done))
	}
}

func TestRunModelAgentReleasesSlotAfterError(t *testing.T) {
	wantErr := errors.New("agent failed")
	runner := Runner{AgentSlots: make(chan struct{}, 1)}

	_, err := runner.runModelAgent(context.Background(), slotResultAgent{err: wantErr}, "task")
	assert.ErrorIs(t, err, wantErr)
	assert.Empty(t, runner.AgentSlots)

	result, err := runner.runModelAgent(context.Background(), slotResultAgent{}, "task")
	require.NoError(t, err)
	assert.Equal(t, "task", result.Task)
	assert.Empty(t, runner.AgentSlots)
}

func TestRunModelAgentReleasesSlotAfterCancellation(t *testing.T) {
	runner := Runner{AgentSlots: make(chan struct{}, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{}, 1)
	done := make(chan error, 1)
	go func() {
		_, err := runner.runModelAgent(ctx, contextWaitingAgent{started: started}, "task")
		done <- err
	}()

	waitForSignal(t, started, "agent start")
	cancel()
	assert.ErrorIs(t, waitForError(t, done), context.Canceled)
	assert.Empty(t, runner.AgentSlots)

	_, err := runner.runModelAgent(context.Background(), slotResultAgent{}, "task")
	require.NoError(t, err)
}

func TestRunModelAgentWaitCanBeCancelled(t *testing.T) {
	runner := Runner{AgentSlots: make(chan struct{}, 1)}
	runner.AgentSlots <- struct{}{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := runner.runModelAgent(ctx, slotResultAgent{}, "task")
		done <- err
	}()
	cancel()
	assert.ErrorIs(t, waitForError(t, done), context.Canceled)
	assert.Len(t, runner.AgentSlots, 1, "cancelled waiter must not consume or release another agent's permit")
}

type slotTestExecutor struct {
	prepareStarted chan<- struct{}
	allowPrepare   <-chan struct{}
	gradingStarted chan<- struct{}
	allowGrading   <-chan struct{}
}

func (e slotTestExecutor) Run(ctx context.Context, spec command.Spec) (command.Result, error) {
	switch spec.Program {
	case "prepare":
		e.prepareStarted <- struct{}{}
		select {
		case <-e.allowPrepare:
			return command.Result{}, nil
		case <-ctx.Done():
			return command.Result{}, ctx.Err()
		}
	case "grade":
		e.gradingStarted <- struct{}{}
		select {
		case <-e.allowGrading:
			return command.Result{}, nil
		case <-ctx.Done():
			return command.Result{}, ctx.Err()
		}
	default:
		return command.Result{}, errors.New("unexpected command")
	}
}

type slotResultAgent struct {
	err error
}

func (a slotResultAgent) Run(_ context.Context, task string) (common.Result, error) {
	return common.Result{Task: task, Termination: common.TerminationCompleted}, a.err
}

func waitForSignal(t *testing.T, signals <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-signals:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func waitForError(t *testing.T, errors <-chan error) error {
	t.Helper()
	select {
	case err := <-errors:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for operation to complete")
		return nil
	}
}

var _ rootagent.Agent = (*slotTrackingAgent)(nil)
