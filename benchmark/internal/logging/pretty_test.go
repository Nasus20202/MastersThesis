package logging

import (
	"bytes"
	"fmt"
	"log/slog"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormatsMultilineTextLogs(t *testing.T) {
	var output bytes.Buffer
	logger, err := New(&output, FormatText, slog.LevelInfo)
	require.NoError(t, err)

	logger.Info("inference message sent",
		"component", "agent",
		"turn", 8,
		"details", "exit_code: 0\nstdout:\nhealthy",
	)

	text := output.String()
	assert.Regexp(t, regexp.MustCompile(`(?m)^\d{2}:\d{2}:\d{2} INFO  \|`), text)
	assert.Contains(t, text, fmt.Sprintf("INFO  | %-32s | component=agent  turn=8", "inference message sent"))
	assert.Contains(t, text, "  details:\n    exit_code: 0\n    stdout:\n    healthy\n")
	assert.NotContains(t, text, `details="exit_code: 0\nstdout:\nhealthy"`)
}

func TestNewPadsLogLevelsToFiveCharacters(t *testing.T) {
	var output bytes.Buffer
	logger, err := New(&output, FormatText, slog.LevelDebug)
	require.NoError(t, err)

	logger.Debug("debug", "value", "one")
	logger.Info("info", "value", "two")
	logger.Warn("warn", "value", "three")
	logger.Error("error", "value", "four")

	text := output.String()
	assert.Contains(t, text, "DEBUG | debug")
	assert.Contains(t, text, "INFO  | info")
	assert.Contains(t, text, "WARN  | warn")
	assert.Contains(t, text, "ERROR | error")
}

func TestNewAlignsStructuredOutputColumn(t *testing.T) {
	var output bytes.Buffer
	logger, err := New(&output, FormatText, slog.LevelInfo)
	require.NoError(t, err)

	logger.Info("short", "first", "value")
	logger.Info("longer event", "second", "value")

	lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
	require.Len(t, lines, 2)
	firstPipe := bytes.Index(lines[0], []byte("|"))
	secondPipe := bytes.Index(lines[0][firstPipe+1:], []byte("|")) + firstPipe + 1
	firstPipeOther := bytes.Index(lines[1], []byte("|"))
	secondPipeOther := bytes.Index(lines[1][firstPipeOther+1:], []byte("|")) + firstPipeOther + 1
	assert.Equal(t, secondPipe, secondPipeOther)
}

func TestNewCanForceColorsForTextLogs(t *testing.T) {
	var output bytes.Buffer
	logger, err := NewWithColor(&output, FormatText, slog.LevelInfo, true)
	require.NoError(t, err)

	logger.Info("visible", "argument", "value")

	text := output.String()
	assert.Contains(t, text, "\x1b[")
	assert.Contains(t, text, colorDim+"argument"+colorReset)
	assert.Contains(t, text, colorDim+"="+colorReset)
}
