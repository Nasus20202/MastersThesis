package components

import (
	"strings"

	bubbles "charm.land/bubbles/v2/viewport"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

// Model is a scrollable pane on the Bubbles viewport. It adds the one-column
// scrollbar the Bubbles component lacks, and lets the owning screen drive
// scrolling so wheel and keyboard routing stay consistent.
type Model struct {
	width  int
	height int
	inner  bubbles.Model
}

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

// View renders the visible slice with a scrollbar column.
func (m *Model) View() string {
	if m.height <= 0 {
		return ""
	}
	return Frame(m.inner.View(), max(1, m.width-1), m.height, m.inner.TotalLineCount(), m.inner.YOffset(), m.height)
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
