package logging

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCreatesConfiguredHandler(t *testing.T) {
	var output bytes.Buffer
	logger, err := New(&output, FormatJSON, slog.LevelInfo)
	require.NoError(t, err)

	logger.Debug("hidden")
	logger.Info("visible", "component", "test")

	assert.NotContains(t, output.String(), "hidden")
	assert.Contains(t, output.String(), `"msg":"visible"`)
	assert.NotContains(t, output.String(), `"source"`)
}

func TestNewAddsSourceAtDebugLevel(t *testing.T) {
	var output bytes.Buffer
	logger, err := New(&output, FormatJSON, slog.LevelDebug)
	require.NoError(t, err)

	logger.Debug("visible")

	assert.Contains(t, output.String(), `"source"`)
}

func TestNewRejectsUnknownFormat(t *testing.T) {
	_, err := New(&bytes.Buffer{}, Format("xml"), slog.LevelInfo)
	assert.Error(t, err)
}

func TestParseLevel(t *testing.T) {
	level, err := ParseLevel(" WARN ")
	require.NoError(t, err)
	assert.Equal(t, slog.LevelWarn, level)

	_, err = ParseLevel("trace")
	assert.Error(t, err)

	level, err = ParseLevel("")
	require.NoError(t, err)
	assert.Equal(t, slog.LevelInfo, level)
}
