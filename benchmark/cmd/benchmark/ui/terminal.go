// Package ui renders the benchmark CLI's interactive terminal output and
// progress bar.
package ui

import (
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

const (
	clearLine            = "\r\x1b[2K"
	hideCursor           = "\x1b[?25l"
	showCursor           = "\x1b[?25h"
	defaultTerminalWidth = 80
)

type Terminal struct {
	writer      io.Writer
	interactive bool
	mu          sync.Mutex
	progress    *Progress
}

func NewTerminal(writer io.Writer) *Terminal {
	return &Terminal{
		writer:      writer,
		interactive: isInteractive(writer),
	}
}

func (t *Terminal) Write(data []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.interactive || t.progress == nil {
		return t.writer.Write(data)
	}
	if _, err := io.WriteString(t.writer, clearLine); err != nil {
		return 0, err
	}
	n, err := t.writer.Write(data)
	if err != nil {
		return n, err
	}
	if renderErr := t.renderProgress(); renderErr != nil {
		return n, renderErr
	}
	return n, nil
}

func (t *Terminal) ColorEnabled() bool {
	return t.interactive
}

func (t *Terminal) renderProgress() error {
	if !t.interactive || t.progress == nil {
		return nil
	}
	_, err := io.WriteString(t.writer, hideCursor+clearLine+t.progress.String())
	return err
}

func isInteractive(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func (t *Terminal) width() int {
	file, ok := t.writer.(*os.File)
	if !ok {
		return defaultTerminalWidth
	}
	width, _, err := term.GetSize(int(file.Fd()))
	if err != nil || width < 1 {
		return defaultTerminalWidth
	}
	return width
}
