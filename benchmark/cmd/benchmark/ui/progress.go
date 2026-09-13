package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	clearLine            = "\r\x1b[2K"
	defaultTerminalWidth = 80
	minimumBarWidth      = 10
)

var refreshInterval = time.Second

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

func (t *Terminal) NewProgress(total, parallelism int) (*Progress, error) {
	if total < 1 {
		return nil, fmt.Errorf("progress total must be at least 1")
	}
	if parallelism < 1 {
		return nil, fmt.Errorf("progress parallelism must be at least 1")
	}

	progress := &Progress{
		terminal:    t,
		total:       total,
		parallelism: parallelism,
		started:     time.Now(),
		refreshStop: make(chan struct{}),
		refreshDone: make(chan struct{}),
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.progress != nil {
		return nil, fmt.Errorf("progress display is already active")
	}
	t.progress = progress
	if t.interactive {
		go progress.refresh()
	}
	return progress, t.renderProgress()
}

type Outcome struct {
	Success bool
}

type Progress struct {
	terminal    *Terminal
	total       int
	parallelism int
	completed   int
	successful  int
	started     time.Time
	finished    bool
	refreshStop chan struct{}
	refreshDone chan struct{}
}

func (p *Progress) Update(outcome Outcome) error {
	p.terminal.mu.Lock()
	defer p.terminal.mu.Unlock()
	if p.finished {
		return nil
	}
	if p.completed < p.total {
		p.completed++
		if outcome.Success {
			p.successful++
		}
	}
	return p.terminal.renderProgress()
}

func (p *Progress) Finish() error {
	p.terminal.mu.Lock()
	if p.finished {
		p.terminal.mu.Unlock()
		return nil
	}
	p.finished = true
	active := p.terminal.progress == p
	if active {
		p.terminal.progress = nil
	}
	interactive := p.terminal.interactive
	p.terminal.mu.Unlock()

	if interactive {
		close(p.refreshStop)
		<-p.refreshDone
	}
	if !active || !interactive {
		return nil
	}

	p.terminal.mu.Lock()
	defer p.terminal.mu.Unlock()
	if _, err := io.WriteString(p.terminal.writer, clearLine+p.String()+"\n"); err != nil {
		return err
	}
	p.terminal.progress = nil
	return nil
}

func (p *Progress) refresh() {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()
	defer close(p.refreshDone)
	for {
		select {
		case <-ticker.C:
			p.terminal.mu.Lock()
			if p.finished || p.terminal.progress != p {
				p.terminal.mu.Unlock()
				return
			}
			_ = p.terminal.renderProgress()
			p.terminal.mu.Unlock()
		case <-p.refreshStop:
			return
		}
	}
}

func (t *Terminal) renderProgress() error {
	if !t.interactive || t.progress == nil {
		return nil
	}
	_, err := io.WriteString(t.writer, clearLine+t.progress.String())
	return err
}

func (p *Progress) String() string {
	percent := p.completed * 100 / p.total
	elapsed := time.Since(p.started)
	eta := "--"
	if p.completed > 0 && p.completed < p.total {
		remaining := p.total - p.completed
		etaDuration := time.Duration(float64(elapsed) * float64(remaining) / float64(p.completed))
		eta = formatDuration(etaDuration)
	}
	suffix := fmt.Sprintf(" %d/%d | %3d%% | passed %d | elapsed %s | ETA %s",
		p.completed, p.total, percent, p.successful, formatDuration(elapsed), eta)
	barWidth := p.terminal.width() - utf8.RuneCountInString(suffix) - 2
	if barWidth < minimumBarWidth {
		barWidth = minimumBarWidth
	}
	filled := p.completed * barWidth / p.total
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	return "[" + bar + "]" + suffix
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	totalSeconds := int(duration.Round(time.Second) / time.Second)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
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
