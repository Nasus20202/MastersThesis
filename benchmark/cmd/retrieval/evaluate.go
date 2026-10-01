package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/llamaenv"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval/evaluation"
)

const resultsDir = "results/retrieval"

func runEvaluate(ctx context.Context, flags *flag.FlagSet, args []string, output io.Writer) error {
	configs := configFlag(flags)
	queries := flags.String("queries", "", "query set file written by generate-queries")
	if err := parse(flags, args); err != nil {
		return err
	}
	return evaluate(ctx, *configs, *queries, output)
}

func evaluate(ctx context.Context, configPaths []string, queriesPath string, output io.Writer) error {
	if queriesPath == "" {
		return errors.New("--queries is required")
	}
	config, err := loadSettings(configPaths)
	if err != nil {
		return err
	}
	set, queriesSHA256, err := evaluation.LoadQuerySet(queriesPath)
	if err != nil {
		return err
	}
	started := time.Now().UTC()
	metadata := evaluation.Metadata{
		RunID:         results.NewRunID(started),
		StartedAt:     started,
		QueriesFile:   filepath.ToSlash(queriesPath),
		QueriesSHA256: queriesSHA256,
		Ks:            evaluation.Ks,
		MaxBytes:      config.maxBytes,
		Indexes:       make(map[retrieval.Chunking]retrieval.IndexMetadata),
		IndexSHA256:   make(map[retrieval.Chunking]string),
	}
	metadata.RepositoryRevision, metadata.WorkingTreeDirty = results.RepositoryProvenance(ctx, ".")

	embedder, err := llamaenv.NewEmbeddingClient()
	if err != nil {
		return err
	}
	indexes := make(map[retrieval.Chunking]*retrieval.Index)
	defer func() {
		for _, index := range indexes {
			_ = index.Close()
		}
	}()
	for _, chunking := range retrieval.Chunkings {
		path := retrieval.IndexPath(config.indexDir, chunking)
		index, err := openIndex(ctx, path, embedder.Metadata())
		if err != nil {
			return err
		}
		indexes[chunking] = index
		metadata.Indexes[chunking] = index.Metadata()
		if metadata.IndexSHA256[chunking], err = retrieval.FileSHA256(path); err != nil {
			return err
		}
	}
	report, err := evaluation.Evaluate(ctx, indexes, embedder, set.Probes, config.maxBytes)
	if err != nil {
		return err
	}
	metadata.FinishedAt = time.Now().UTC()

	runDir := filepath.Join(resultsDir, metadata.RunID)
	if err := evaluation.WriteReport(runDir, metadata, report); err != nil {
		return err
	}
	printReport(output, report)
	fmt.Fprintf(output, "\nresults: %s\n", runDir)
	return nil
}

func printReport(output io.Writer, report evaluation.Report) {
	table := tabwriter.NewWriter(output, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(table, "config\thit@k rewrite\tMRR rewrite\tprimary rewrite\thit@k as-is\tMRR as-is\tprimary as-is\ttruncated\t")
	for _, config := range report.Configs {
		fmt.Fprintf(table, "%s\t%.3f\t%.3f\t%.3f\t%.3f\t%.3f\t%.3f\t%d\t\n", config.Name(),
			config.Rewrite.Hit, config.Rewrite.RR, config.Rewrite.PrimaryHit,
			config.AsIs.Hit, config.AsIs.RR, config.AsIs.PrimaryHit, config.Truncated)
	}
	_ = table.Flush()
	fmt.Fprintf(output, "\nselected: %s\n\n", report.Selected.Name())

	table = tabwriter.NewWriter(output, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(table, "config\thit@k rewrite minus selected\t95% bootstrap CI\t")
	for _, difference := range report.Differences {
		fmt.Fprintf(table, "%s\t%+.3f\t[%+.3f, %+.3f]\t\n", difference.Config, difference.Mean, difference.Low, difference.High)
	}
	_ = table.Flush()
}
