package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressIsSilentForNonInteractiveWriters(t *testing.T) {
	var output bytes.Buffer
	terminal := NewTerminal(&output)

	progress, err := terminal.NewProgress(2, 1)
	require.NoError(t, err)
	require.NoError(t, progress.Update(Outcome{Success: true}))
	require.NoError(t, progress.Finish())

	assert.Empty(t, output.String())
}

func TestProgressRendersCompletionParallelismAndETA(t *testing.T) {
	var output bytes.Buffer
	terminal := NewTerminal(&output)
	terminal.interactive = true

	progress, err := terminal.NewProgress(4, 2)
	require.NoError(t, err)
	progress.started = time.Now().Add(-2 * time.Minute)
	require.NoError(t, progress.Update(Outcome{Success: true}))
	require.NoError(t, progress.Update(Outcome{Success: false}))
	require.NoError(t, progress.Finish())

	text := output.String()
	assert.Contains(t, text, "2/4")
	assert.Contains(t, text, "50%")
	assert.Contains(t, text, "passed 1")
	assert.Contains(t, text, "parallel 2")
	assert.Contains(t, text, "ETA")
	assert.Equal(t, 80, utf8.RuneCountInString(progress.String()))
	assert.Contains(t, text, "\n"+showCursor)
}

func TestProgressHidesCursorUntilFinished(t *testing.T) {
	var output bytes.Buffer
	terminal := NewTerminal(&output)
	terminal.interactive = true

	progress, err := terminal.NewProgress(1, 1)
	require.NoError(t, err)
	assert.Contains(t, output.String(), hideCursor)
	assert.NotContains(t, output.String(), showCursor)

	require.NoError(t, progress.Finish())
	assert.Contains(t, output.String(), showCursor)
}

func TestProgressRefreshesInBackground(t *testing.T) {
	previousInterval := refreshInterval
	refreshInterval = 10 * time.Millisecond
	t.Cleanup(func() { refreshInterval = previousInterval })

	var output bytes.Buffer
	terminal := NewTerminal(&output)
	terminal.interactive = true

	progress, err := terminal.NewProgress(2, 1)
	require.NoError(t, err)
	time.Sleep(30 * time.Millisecond)
	require.NoError(t, progress.Finish())

	assert.GreaterOrEqual(t, strings.Count(output.String(), "["), 3)
}

func TestNewProgressRejectsInvalidArguments(t *testing.T) {
	terminal := NewTerminal(&bytes.Buffer{})

	_, err := terminal.NewProgress(0, 1)
	assert.EqualError(t, err, "progress total must be at least 1")

	_, err = terminal.NewProgress(1, 0)
	assert.EqualError(t, err, "progress parallelism must be at least 1")
}
