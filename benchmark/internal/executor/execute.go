// Package executor runs benchmark and validation tasks concurrently and
// streams their outcomes.
package executor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
)

type Task struct {
	ScenarioID string
	CaseID     string
	Agent      string
	Attempt    int
	Run        func(context.Context) (orchestration.RunResult, error)
}

type Outcome struct {
	ScenarioID string
	CaseID     string
	Agent      string
	Attempt    int
	Result     orchestration.RunResult
	Err        error
}

// Execute runs tasks concurrently and streams each completed outcome on the
// returned channel. The error channel receives the aggregate execution error
// after the outcomes channel is closed.
func Execute(ctx context.Context, tasks []Task, parallelism int) (<-chan Outcome, <-chan error) {
	outcomes := make(chan Outcome)
	errorChannel := make(chan error, 1)
	go execute(ctx, tasks, parallelism, outcomes, errorChannel)
	return outcomes, errorChannel
}

func execute(ctx context.Context, tasks []Task, parallelism int, outcomes chan<- Outcome, errorChannel chan<- error) {
	defer close(outcomes)
	defer close(errorChannel)

	if len(tasks) == 0 {
		errorChannel <- errors.New("benchmark requires at least one task")
		return
	}
	if parallelism < 1 {
		errorChannel <- errors.New("benchmark parallelism must be at least 1")
		return
	}
	for index, task := range tasks {
		if task.Run == nil {
			errorChannel <- fmt.Errorf("benchmark task %d has no executor", index+1)
			return
		}
		if task.ScenarioID == "" {
			errorChannel <- fmt.Errorf("benchmark task %d has no scenario ID", index+1)
			return
		}
		if task.Attempt < 0 {
			errorChannel <- fmt.Errorf("benchmark task %d has invalid attempt", index+1)
			return
		}
	}

	jobs := make(chan int)
	completed := make([]Outcome, len(tasks))
	var workers sync.WaitGroup
	for range min(parallelism, len(tasks)) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				if ctx.Err() != nil {
					continue
				}
				task := tasks[index]
				result, err := task.Run(ctx)
				completed[index] = Outcome{
					ScenarioID: task.ScenarioID,
					CaseID:     task.CaseID,
					Agent:      task.Agent,
					Attempt:    max(task.Attempt, 1),
					Result:     result,
					Err:        err,
				}
				outcomes <- completed[index]
			}
		}()
	}
	dispatchErr := dispatch(ctx, jobs, len(tasks))
	workers.Wait()

	errs := []error{dispatchErr}
	for index, outcome := range completed {
		if outcome.Err != nil {
			label := outcome.ScenarioID
			if outcome.Agent != "" {
				label += "/" + outcome.Agent
			}
			errs = append(errs, fmt.Errorf("task %d (%s): %w", index+1, label, outcome.Err))
		}
	}
	errorChannel <- errors.Join(errs...)
}

// dispatch hands out task indexes until all are taken or ctx ends.
func dispatch(ctx context.Context, jobs chan<- int, count int) error {
	defer close(jobs)
	for index := range count {
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case jobs <- index:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
