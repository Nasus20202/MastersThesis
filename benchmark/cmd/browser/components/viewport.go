// The viewport is a scrollable pane built on the Bubbles viewport. It adds
// the one-column scrollbar the Bubbles component lacks, and lets the owning
// screen drive scrolling so wheel and keyboard routing stay consistent.
package components

import (
	"strings"

	bubbles "charm.land/bubbles/v2/viewport"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

// Model is one scrollable pane.
type Model struct {
	width  int
	height int
	inner  bubbles.Model
}

// New returns an empty pane.
func NewViewport() *Model {
	inner := bubbles.New()
	inner.SoftWrap = false
	inner.MouseWheelEnabled = false
	return &Model{inner: inner}
}

// SetSize sets the pane's outer size (scrollbar included).
func (m *Model) SetSize(width, height int) {
	m.width = max(1, width)
	m.height = max(1, height)
	m.inner.SetWidth(max(1, m.width-1))
	m.inner.SetHeight(m.height)
}

// SetLines replaces the content, keeping the offset in range.
func (m *Model) SetLines(lines []string) {
	m.inner.SetContentLines(lines)
}

// SetOffset scrolls to an absolute line, clamped.
func (m *Model) SetOffset(offset int) {
	m.inner.SetYOffset(offset)
}

// Scroll moves the offset by delta lines.
func (m *Model) Scroll(delta int) {
	if delta >= 0 {
		m.inner.ScrollDown(delta)
		return
	}
	m.inner.ScrollUp(-delta)
}

// Offset returns the first visible line.
func (m *Model) Offset() int { return m.inner.YOffset() }

// Total returns the number of content lines.
func (m *Model) Total() int { return m.inner.TotalLineCount() }

// Visible returns the number of visible content lines.
func (m *Model) Visible() int { return m.height }

// MaxOffset returns the largest valid offset.
func (m *Model) MaxOffset() int {
	return max(0, m.Total()-m.Visible())
}

// View renders the visible slice with a scrollbar column.
func (m *Model) View() string {
	if m.height <= 0 {
		return ""
	}
	cells := scrollbarCells(m.height, m.Total(), m.Offset(), m.Visible())
	rows := strings.Split(m.inner.View(), "\n")
	out := make([]string, m.height)
	contentWidth := max(1, m.width-1)
	for index := 0; index < m.height; index++ {
		line := ""
		if index < len(rows) {
			line = rows[index]
		}
		out[index] = ui.PadRight(line, contentWidth) + cells[index]
	}
	return strings.Join(out, "\n")
}

// Frame pads content to height and appends a one-column scrollbar driven by an
// already-windowed pane (such as a table that scrolls itself).
func Frame(content string, contentWidth, height, total, offset, visible int) string {
	cells := scrollbarCells(height, total, offset, visible)
	lines := strings.Split(content, "\n")
	out := make([]string, height)
	for index := 0; index < height; index++ {
		line := ""
		if index < len(lines) {
			line = lines[index]
		}
		out[index] = ui.PadRight(line, contentWidth) + cells[index]
	}
	return strings.Join(out, "\n")
}

func scrollbarCells(height, total, offset, visible int) []string {
	height = max(1, height)
	cells := make([]string, height)
	for index := range cells {
		cells[index] = ui.FaintStyle.Render("│")
	}
	if total <= visible || total <= 0 || visible <= 0 {
		return cells
	}
	thumb := max(1, height*visible/total)
	thumb = min(thumb, height)
	position := (height - thumb) * offset / max(1, total-visible)
	position = max(0, min(position, height-thumb))
	for index := position; index < position+thumb && index < height; index++ {
		cells[index] = ui.AccentStyle.Render("█")
	}
	return cells
}
