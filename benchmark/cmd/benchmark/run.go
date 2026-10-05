package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/ui"
	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/executor"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

const resultsDir = "results"

// runOptions are the settings shared by benchmark and validation runs.
type runOptions struct {
	parallel int
	repeat   int
	tags     scenario.TagFilter
	config   benchmarkconfig.Config
	terminal *ui.Terminal
}

// startRun stamps the run identity and repository provenance onto metadata and
// creates its result store.
func startRun(ctx context.Context, metadata results.RunMetadata) (*results.Store, error) {
	startedAt := time.Now().UTC()
	metadata.RunID = results.NewRunID(startedAt)
	metadata.StartedAt = startedAt
	metadata.RepositoryRevision, metadata.WorkingTreeDirty = results.RepositoryProvenance(ctx, ".")
	return results.New(resultsDir, metadata)
}

// execute runs tasks behind a progress bar, passes each outcome to record and
// finalizes the store. record reports whether the attempt succeeded and any
// error persisting it.
func (o runOptions) execute(ctx context.Context, store *results.Store, tasks []executor.Task, record func(executor.Outcome) (bool, error)) (err error) {
	defer func() {
		err = errors.Join(err, store.Finalize(time.Now().UTC()))
	}()
	if len(tasks) == 0 {
		return nil
	}
	progress, err := o.terminal.NewProgress(len(tasks), o.parallel)
	if err != nil {
		return err
	}
	defer func() {
		if err := progress.Finish(); err != nil {
			slog.Error("progress display failed", "error", err)
		}
	}()
	outcomes, executeErrors := executor.Execute(ctx, tasks, o.parallel)
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
