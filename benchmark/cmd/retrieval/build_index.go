package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/llamaenv"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

func runBuildIndex(ctx context.Context, flags *flag.FlagSet, args []string, _ io.Writer) error {
	configs := configFlag(flags)
	chunking := flags.String("chunking", "all", "chunking to index: sections, windows or all")
	if err := parse(flags, args); err != nil {
		return err
	}
	return buildIndexes(ctx, *configs, *chunking)
}

func buildIndexes(ctx context.Context, configPaths []string, chunkingValue string) error {
	config, err := loadSettings(configPaths)
	if err != nil {
		return err
	}
	chunkings := retrieval.Chunkings
	if chunkingValue != "all" {
		chunking, err := retrieval.ParseChunking(chunkingValue)
		if err != nil {
			return err
		}
		chunkings = []retrieval.Chunking{chunking}
	}
	checkout, subtree, revision, err := corpusCheckout(ctx)
	if err != nil {
		return err
	}
	documents, err := retrieval.LoadCorpus(os.DirFS(checkout), subtree)
	if err != nil {
		return err
	}
	embedder, err := llamaenv.NewEmbeddingClient()
	if err != nil {
		return err
	}
	for _, chunking := range chunkings {
		path := retrieval.IndexPath(config.indexDir, chunking)
		slog.InfoContext(ctx, "building index", "chunking", chunking, "documents", len(documents), "path", path)
		metadata, err := retrieval.BuildIndex(ctx, path, documents, chunking, embedder,
			retrieval.IndexMetadata{CorpusRevision: revision, Embedding: embedder.Metadata()})
		if err != nil {
			return fmt.Errorf("build %s index: %w", chunking, err)
		}
		digest, err := retrieval.FileSHA256(path)
		if err != nil {
			return err
		}
		slog.InfoContext(ctx, "index built", "chunking", chunking, "chunks", metadata.Chunks, "dimensions", metadata.Dimensions, "sha256", digest)
	}
	return nil
}
