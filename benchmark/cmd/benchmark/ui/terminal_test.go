package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminalWritePassesThroughForNonInteractiveWriter(t *testing.T) {
	var output bytes.Buffer
	terminal := NewTerminal(&output)

	data := []byte("log message\n")
	n, err := terminal.Write(data)

	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, string(data), output.String())
}

func TestTerminalRedrawsProgressAfterLogOutput(t *testing.T) {
	var output bytes.Buffer
	terminal := NewTerminal(&output)
	terminal.interactive = true

	progress, err := terminal.NewProgress(1, 1)
	require.NoError(t, err)
	_, err = terminal.Write([]byte("log message\n"))
	require.NoError(t, err)
	require.NoError(t, progress.Finish())

	assert.Contains(t, output.String(), "log message\n")
	assert.GreaterOrEqual(t, strings.Count(output.String(), "["), 2)
}

func TestTerminalWidthDefaultsForNonFileWriter(t *testing.T) {
	terminal := NewTerminal(&bytes.Buffer{})

	assert.Equal(t, defaultTerminalWidth, terminal.width())
}
