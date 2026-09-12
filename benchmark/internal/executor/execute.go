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
	Attempt    int
	Run        func(context.Context) (orchestration.RunResult, error)
}

type Outcome struct {
	ScenarioID string
	CaseID     string
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

	workerCount := min(parallelism, len(tasks))
	jobs := make(chan int)
	completed := make([]Outcome, len(tasks))
	var workers sync.WaitGroup
	workers.Add(workerCount)
	emit := func(index int, result Outcome) {
		completed[index] = result
		outcomes <- result
	}
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				task := tasks[index]
				if err := ctx.Err(); err != nil {
					continue
				}
				attempt := task.Attempt
				if attempt == 0 {
					attempt = 1
				}
				result, err := task.Run(ctx)
				emit(index, outcome(task, attempt, result, err))
			}
		}()
	}
	dispatchDone := make(chan struct{})
	var dispatchErr error
	go func() {
		defer close(dispatchDone)
		defer close(jobs)
	dispatch:
		for index := range tasks {
			if err := ctx.Err(); err != nil {
				dispatchErr = err
				break
			}
			select {
			case jobs <- index:
			case <-ctx.Done():
				dispatchErr = ctx.Err()
				break dispatch
			}
		}
	}()
	<-dispatchDone
	workers.Wait()

	errs := make([]error, 0, 1)
	if dispatchErr != nil {
		errs = append(errs, dispatchErr)
	}
	for index, outcome := range completed {
		if outcome.Err != nil {
			errs = append(errs, fmt.Errorf("task %d (%s): %w", index+1, outcome.ScenarioID, outcome.Err))
		}
	}
	errorChannel <- errors.Join(errs...)
}

func outcome(task Task, attempt int, result orchestration.RunResult, err error) Outcome {
	return Outcome{
		ScenarioID: task.ScenarioID,
		CaseID:     task.CaseID,
		Attempt:    attempt,
		Result:     result,
		Err:        err,
	}
}
