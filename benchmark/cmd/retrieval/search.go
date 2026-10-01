package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/llamaenv"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/retrievalenv"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

func runSearch(ctx context.Context, flags *flag.FlagSet, args []string, output io.Writer) error {
	configs := configFlag(flags)
	mode := flags.String("mode", "", "retrieval mode (default: from config)")
	chunking := flags.String("chunking", "", "chunking (default: from config)")
	k := flags.Int("k", 0, "number of results (default: from config)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	query := strings.TrimSpace(strings.Join(flags.Args(), " "))
	if query == "" {
		return errors.New("search query is required")
	}
	return search(ctx, *configs, searchOverrides{mode: *mode, chunking: *chunking, k: *k}, query, output)
}

type searchOverrides struct {
	mode     string
	chunking string
	k        int
}

func search(ctx context.Context, configPaths []string, overrides searchOverrides, query string, output io.Writer) error {
	config, err := loadSettings(configPaths)
	if err != nil {
		return err
	}
	if overrides.mode != "" {
		if config.Mode, err = retrieval.ParseMode(overrides.mode); err != nil {
			return err
		}
	}
	if overrides.chunking != "" {
		if config.Chunking, err = retrieval.ParseChunking(overrides.chunking); err != nil {
			return err
		}
	}
	if overrides.k < 0 {
		return errors.New("--k must be positive")
	}
	if overrides.k > 0 {
		config.TopK = overrides.k
	}
	embedder, err := llamaenv.NewEmbeddingClient()
	if err != nil {
		return err
	}
	index, err := retrievalenv.OpenIndex(ctx, retrieval.IndexPath(config.IndexDir, config.Chunking), embedder.Metadata())
	if err != nil {
		return err
	}
	defer index.Close()
	hits, err := index.Search(ctx, embedder, config.Mode, query, config.TopK)
	if err != nil {
		return err
	}
	fmt.Fprintf(output, "%s/%s/k=%d\n", config.Mode, config.Chunking, config.TopK)
	for _, hit := range hits {
		fmt.Fprintf(output, "%d\t%.4f\t%d\t%s\n", hit.Rank, hit.Score, hit.ChunkID, hit.Chunk.Path)
	}
	rendered, truncated := retrieval.Render(hits, config.MaxBytes)
	fmt.Fprintf(output, "\n--- model-visible result (%d bytes, truncated: %t) ---\n%s\n", len(rendered), truncated, rendered)
	return nil
}
