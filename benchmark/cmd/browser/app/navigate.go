package app

import (
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/screens"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

func (m *Model) move(delta int) {
	current := m.current()
	if screens.IsList(current.Kind) {
		count := m.view.ListCount(current)
		if count == 0 {
			return
		}
		current.Cursor = max(0, min(count-1, current.Cursor+delta))
		m.scrollCursorIntoView(current)
		return
	}
	height := m.contentHeight()
	visible := max(1, screens.DetailsHeight(m.width, height)-ui.LineCount(m.view.DetailsHead(current)))
	maxOffset := max(0, len(m.view.DetailsLines(current, screens.PaneWidth(m.width)))-visible)
	current.Offset = max(0, min(current.Offset+delta, maxOffset))
}

func (m *Model) scrollCursorIntoView(current *screens.Route) {
	visible := m.visibleRows()
	if current.Cursor < current.Offset {
		current.Offset = current.Cursor
	}
	if current.Cursor >= current.Offset+visible {
		current.Offset = current.Cursor - visible + 1
	}
}

func (m *Model) jump(index int) {
	current := m.current()
	if !screens.IsList(current.Kind) {
		if index < 0 {
			current.Offset = 1 << 30
		} else {
			current.Offset = 0
		}
		return
	}
	count := m.view.ListCount(current)
	if count == 0 {
		return
	}
	if index < 0 {
		m.move(count)
		return
	}
	current.Cursor = 0
	current.Offset = 0
}

func (m *Model) pageSize() int {
	return max(1, m.visibleRows())
}

func (m *Model) visibleRows() int {
	content := m.contentHeight()
	current := m.current()
	if screens.IsList(current.Kind) {
		width := max(1, screens.PaneWidth(m.width)-1)
		content -= m.view.HeadLines(current, width) + 1
	}
	return max(1, content)
}

func (m *Model) scrollPreview(delta int) {
	current := m.current()
	total := len(m.view.PreviewLines(current, m.width))
	visible := m.contentHeight()
	current.PreviewOffset = max(0, min(current.PreviewOffset+delta, max(0, total-visible)))
}

func (m *Model) previewPage() int {
	return max(1, m.contentHeight()-2)
}

func (m *Model) previewJump(index int) {
	current := m.current()
	if index < 0 {
		current.PreviewOffset = 1 << 30
	} else {
		current.PreviewOffset = 0
	}
	m.scrollPreview(0)
}
