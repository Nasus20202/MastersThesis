package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestWatchReportsChanges(t *testing.T) {
	root := t.TempDir()
	messages := make(chan tea.Msg, 4)
	watcher, err := Watch(root, func(message tea.Msg) { messages <- message })
	require.NoError(t, err)
	defer func() { require.NoError(t, watcher.Close()) }()

	require.NoError(t, os.WriteFile(filepath.Join(root, "run.json"), []byte("{}"), 0o644))

	select {
	case message := <-messages:
		require.IsType(t, WatchMsg{}, message)
	case <-time.After(3 * time.Second):
		t.Fatal("no watch message received")
	}
}
