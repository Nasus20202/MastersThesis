package app

import (
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/screens"
)

func (m *Model) chatWidth() int { return screens.PreviewWidth(m.width) }

func (m *Model) moveChat(delta int) {
	current := m.current()
	layout, ok := m.view.ChatLayout(current, m.chatWidth())
	if !ok || len(layout.MessageSpans) == 0 {
		return
	}
	current.Cursor = max(0, min(len(layout.MessageSpans)-1, current.Cursor+delta))
	m.ensureChatVisible(current, layout)
}

func (m *Model) chatJump(index int) {
	current := m.current()
	layout, ok := m.view.ChatLayout(current, m.chatWidth())
	if !ok || len(layout.MessageSpans) == 0 {
		return
	}
	if index < 0 {
		current.Cursor = len(layout.MessageSpans) - 1
	} else {
		current.Cursor = 0
	}
	m.ensureChatVisible(current, layout)
}

func (m *Model) scrollChat(delta int) {
	current := m.current()
	layout, ok := m.view.ChatLayout(current, m.chatWidth())
	if !ok {
		return
	}
	maxOffset := max(0, len(layout.Lines)-screens.ChatVisible(m.width, m.contentHeight()))
	current.ChatOffset = max(0, min(current.ChatOffset+delta, maxOffset))
}

func (m *Model) toggleSelectedChat() {
	current := m.current()
	if current.Expanded == nil {
		current.Expanded = make(map[int]bool)
	}
	current.Expanded[current.Cursor] = !current.Expanded[current.Cursor]
	layout, ok := m.view.ChatLayout(current, m.chatWidth())
	if !ok {
		return
	}
	m.ensureChatVisible(current, layout)
}

func (m *Model) chatSelectLast() {
	current := m.current()
	if current.Kind != screens.Attempt {
		return
	}
	layout, ok := m.view.ChatLayout(current, m.chatWidth())
	if !ok || len(layout.MessageSpans) == 0 {
		return
	}
	current.Cursor = len(layout.MessageSpans) - 1
	current.Expanded = nil
	current.ChatOffset = max(0, layout.MessageSpans[current.Cursor][1]-screens.ChatVisible(m.width, m.contentHeight()))
}

func (m *Model) ensureChatVisible(current *screens.Route, layout components.Layout) {
	if current.Cursor >= len(layout.MessageSpans) {
		return
	}
	visible := screens.ChatVisible(m.width, m.contentHeight())
	span := layout.MessageSpans[current.Cursor]
	if span[0] == 0 && span[1] == 0 {
		return
	}
	if span[0] < current.ChatOffset {
		current.ChatOffset = span[0]
	}
	if span[1] > current.ChatOffset+visible {
		current.ChatOffset = span[1] - visible
	}
	current.ChatOffset = max(0, current.ChatOffset)
}

func (m *Model) chatPage() int {
	return max(1, screens.ChatVisible(m.width, m.contentHeight()))
}
