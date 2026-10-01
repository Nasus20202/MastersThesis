package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

func TestRunRejectsUnknownCommandsAndMissingArguments(t *testing.T) {
	ctx := context.Background()
	assert.ErrorContains(t, run(ctx, []string{"index"}, io.Discard, io.Discard), `unknown command "index"`)
	assert.ErrorContains(t, run(ctx, []string{"generate-queries", "--probes", "p.json"}, io.Discard, io.Discard), "--probes and --out are required")
	assert.ErrorContains(t, run(ctx, []string{"search"}, io.Discard, io.Discard), "search query is required")
	assert.ErrorContains(t, run(ctx, []string{"build-index", "extra"}, io.Discard, io.Discard), "unexpected arguments")
}

func TestLoadSettingsValidatesRetrievalConfig(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("retrieval:\n  mode: hybrid\n  chunking: windows\n  top_k: 5\n  max_bytes: 8192\n  index_dir: corpus\n"), 0o600))

	config, err := loadSettings([]string{path})
	require.NoError(t, err)
	assert.Equal(t, settings{mode: retrieval.Hybrid, chunking: retrieval.Windows, topK: 5, maxBytes: 8192, indexDir: filepath.Join(directory, "corpus")}, config)

	require.NoError(t, os.WriteFile(path, []byte("retrieval:\n  mode: dense\n  chunking: windows\n  top_k: 5\n  max_bytes: 8192\n  index_dir: corpus\n"), 0o600))
	_, err = loadSettings([]string{path})
	assert.ErrorContains(t, err, `unsupported retrieval mode "dense"`)
}
