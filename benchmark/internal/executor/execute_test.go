package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteRunsTasksInInputOrder(t *testing.T) {
	tasks := []Task{
		{ScenarioID: "first", Attempt: 2, Run: func(context.Context) (orchestration.RunResult, error) {
			time.Sleep(20 * time.Millisecond)
			return orchestration.RunResult{ScenarioID: "first"}, nil
		}},
		{ScenarioID: "second", Run: func(context.Context) (orchestration.RunResult, error) {
			return orchestration.RunResult{ScenarioID: "second"}, nil
		}},
	}

	outcomes, err := Execute(context.Background(), tasks, 2)

	require.NoError(t, err)
	require.Len(t, outcomes, 2)
	assert.Equal(t, "first", outcomes[0].ScenarioID)
	assert.Equal(t, 2, outcomes[0].Attempt)
	assert.Equal(t, "second", outcomes[1].ScenarioID)
	assert.Equal(t, 1, outcomes[1].Attempt)
}

func TestExecuteLimitsConcurrentTasks(t *testing.T) {
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	tasks := make([]Task, 3)
	for index := range tasks {
		tasks[index] = Task{
			ScenarioID: "scenario",
			Run: func(context.Context) (orchestration.RunResult, error) {
				started <- struct{}{}
				<-release
				return orchestration.RunResult{ScenarioID: "scenario"}, nil
			},
		}
	}

	done := make(chan struct{})
	go func() {
		_, _ = Execute(context.Background(), tasks, 2)
		close(done)
	}()

	<-started
	<-started
	select {
	case <-started:
		close(release)
		t.Fatal("started more tasks than configured parallelism")
	case <-time.After(25 * time.Millisecond):
	}
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("tasks did not complete")
	}
}

func TestExecuteCollectsAllTaskErrors(t *testing.T) {
	wantErr := errors.New("task failed")
	tasks := []Task{
		{ScenarioID: "failed", Run: func(context.Context) (orchestration.RunResult, error) {
			return orchestration.RunResult{}, wantErr
		}},
		{ScenarioID: "successful", Run: func(context.Context) (orchestration.RunResult, error) {
			return orchestration.RunResult{ScenarioID: "successful"}, nil
		}},
	}

	outcomes, err := Execute(context.Background(), tasks, 2)

	assert.ErrorIs(t, err, wantErr)
	require.Len(t, outcomes, 2)
	assert.ErrorIs(t, outcomes[0].Err, wantErr)
	assert.NoError(t, outcomes[1].Err)
}

func TestExecuteDoesNotStartQueuedTasksAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	startedQueued := make(chan struct{}, 1)
	tasks := []Task{
		{ScenarioID: "running", Run: func(ctx context.Context) (orchestration.RunResult, error) {
			close(started)
			<-ctx.Done()
			return orchestration.RunResult{}, ctx.Err()
		}},
		{ScenarioID: "queued", Run: func(context.Context) (orchestration.RunResult, error) {
			startedQueued <- struct{}{}
			return orchestration.RunResult{ScenarioID: "queued"}, nil
		}},
	}

	done := make(chan []Outcome, 1)
	go func() {
		outcomes, _ := Execute(ctx, tasks, 1)
		done <- outcomes
	}()
	<-started
	cancel()

	select {
	case <-startedQueued:
		t.Fatal("queued task started after cancellation")
	case outcomes := <-done:
		require.Len(t, outcomes, 2)
		assert.ErrorIs(t, outcomes[0].Err, context.Canceled)
		assert.ErrorIs(t, outcomes[1].Err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("executor did not stop after cancellation")
	}
}

func TestExecuteRejectsInvalidInput(t *testing.T) {
	_, err := Execute(context.Background(), nil, 1)
	assert.ErrorContains(t, err, "at least one task")

	_, err = Execute(context.Background(), []Task{{ScenarioID: "scenario", Run: func(context.Context) (orchestration.RunResult, error) {
		return orchestration.RunResult{}, nil
	}}}, 0)
	assert.ErrorContains(t, err, "parallelism")

	_, err = Execute(context.Background(), []Task{{ScenarioID: "scenario"}}, 1)
	assert.ErrorContains(t, err, "no executor")

	_, err = Execute(context.Background(), []Task{{Run: func(context.Context) (orchestration.RunResult, error) {
		return orchestration.RunResult{}, nil
	}}}, 1)
	assert.ErrorContains(t, err, "scenario ID")

	_, err = Execute(context.Background(), []Task{{ScenarioID: "scenario", Attempt: -1, Run: func(context.Context) (orchestration.RunResult, error) {
		return orchestration.RunResult{}, nil
	}}}, 1)
	assert.ErrorContains(t, err, "invalid attempt")
}
