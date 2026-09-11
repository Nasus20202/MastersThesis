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

func Execute(ctx context.Context, tasks []Task, parallelism int) ([]Outcome, error) {
	if len(tasks) == 0 {
		return nil, errors.New("benchmark requires at least one task")
	}
	if parallelism < 1 {
		return nil, errors.New("benchmark parallelism must be at least 1")
	}
	for index, task := range tasks {
		if task.Run == nil {
			return nil, fmt.Errorf("benchmark task %d has no executor", index+1)
		}
		if task.ScenarioID == "" {
			return nil, fmt.Errorf("benchmark task %d has no scenario ID", index+1)
		}
		if task.Attempt < 0 {
			return nil, fmt.Errorf("benchmark task %d has invalid attempt", index+1)
		}
	}

	workerCount := min(parallelism, len(tasks))
	jobs := make(chan int)
	outcomes := make([]Outcome, len(tasks))
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				task := tasks[index]
				if err := ctx.Err(); err != nil {
					outcomes[index] = canceledOutcome(task, err)
					continue
				}
				attempt := task.Attempt
				if attempt == 0 {
					attempt = 1
				}
				result, err := task.Run(ctx)
				outcomes[index] = outcome(task, attempt, result, err)
			}
		}()
	}
dispatch:
	for index := range tasks {
		if err := ctx.Err(); err != nil {
			markCanceled(outcomes, tasks, index, err)
			break
		}
		select {
		case jobs <- index:
		case <-ctx.Done():
			markCanceled(outcomes, tasks, index, ctx.Err())
			break dispatch
		}
	}
	close(jobs)
	workers.Wait()

	errs := make([]error, 0)
	for index, outcome := range outcomes {
		if outcome.Err != nil {
			errs = append(errs, fmt.Errorf("task %d (%s): %w", index+1, outcome.ScenarioID, outcome.Err))
		}
	}
	return outcomes, errors.Join(errs...)
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

func canceledOutcome(task Task, err error) Outcome {
	attempt := task.Attempt
	if attempt == 0 {
		attempt = 1
	}
	return outcome(task, attempt, orchestration.RunResult{ScenarioID: task.ScenarioID}, err)
}

func markCanceled(outcomes []Outcome, tasks []Task, start int, err error) {
	for index := start; index < len(tasks); index++ {
		outcomes[index] = canceledOutcome(tasks[index], err)
	}
}
