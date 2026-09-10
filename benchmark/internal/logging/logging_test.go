package logging

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCreatesConfiguredHandler(t *testing.T) {
	var output bytes.Buffer
	logger, err := New(&output, FormatJSON, slog.LevelInfo)
	if !assert.NoError(t, err) {
		return
	}

	logger.Debug("hidden")
	logger.Info("visible", "component", "test")

	assert.NotContains(t, output.String(), "hidden")
	assert.Contains(t, output.String(), `"msg":"visible"`)
	assert.NotContains(t, output.String(), `"source"`)
}

func TestNewAddsSourceAtDebugLevel(t *testing.T) {
	var output bytes.Buffer
	logger, err := New(&output, FormatJSON, slog.LevelDebug)
	if !assert.NoError(t, err) {
		return
	}

	logger.Debug("visible")

	assert.Contains(t, output.String(), `"source"`)
}

func TestNewRejectsUnknownFormat(t *testing.T) {
	_, err := New(&bytes.Buffer{}, Format("xml"), slog.LevelInfo)
	assert.Error(t, err)
}

func TestParseLevel(t *testing.T) {
	level, err := ParseLevel(" WARN ")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, slog.LevelWarn, level)

	_, err = ParseLevel("trace")
	assert.Error(t, err)
	level, err = ParseLevel("")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, slog.LevelInfo, level)
}
