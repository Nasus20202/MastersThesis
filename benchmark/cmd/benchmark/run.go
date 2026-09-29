package main

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/executor"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

const resultsDir = "results"

// startRun stamps the run identity and repository provenance onto metadata and
// creates its result store.
func startRun(ctx context.Context, metadata results.RunMetadata) (*results.Store, error) {
	startedAt := time.Now().UTC()
	metadata.RunID = runID(startedAt)
	metadata.StartedAt = startedAt
	metadata.RepositoryRevision, metadata.WorkingTreeDirty = repositoryProvenance(ctx)
	return results.New(resultsDir, metadata)
}

func runID(startedAt time.Time) string {
	timestamp := startedAt.UTC().Format("2006-01-02-15-04-05.000Z")
	return "run-" + strings.Replace(timestamp, ".", "-", 1)
}

// executeRun runs tasks behind a progress bar, passes each outcome to record and
// finalizes the store. record reports whether the attempt succeeded and any
// error persisting it.
func executeRun(ctx context.Context, store *results.Store, terminal *ui.Terminal, tasks []executor.Task, parallelism int, record func(executor.Outcome) (bool, error)) (err error) {
	defer func() {
		err = errors.Join(err, store.Finalize(time.Now().UTC()))
	}()
	if len(tasks) == 0 {
		return nil
	}
	progress, err := terminal.NewProgress(len(tasks), parallelism)
	if err != nil {
		return err
	}
	defer func() {
		if err := progress.Finish(); err != nil {
			slog.Error("progress display failed", "error", err)
		}
	}()
	outcomes, executeErrors := executor.Execute(ctx, tasks, parallelism)
	var writeErr error
	for outcome := range outcomes {
		success, err := record(outcome)
		writeErr = errors.Join(writeErr, err)
		if err := progress.Update(ui.Outcome{Success: success}); err != nil {
			slog.Error("progress display failed", "error", err)
		}
	}
	return errors.Join(writeErr, <-executeErrors)
}
