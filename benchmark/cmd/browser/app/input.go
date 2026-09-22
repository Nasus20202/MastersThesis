package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/screens"
)

func (m *Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		return m, tea.Quit
	}
	if key == "?" {
		m.showHelp = !m.showHelp
		return m, nil
	}
	if m.showHelp {
		if key == "esc" || key == "q" {
			m.showHelp = false
		}
		return m, nil
	}

	current := m.current()
	inChat := current.Kind == screens.Attempt && current.Focus == screens.FocusSecondary
	onPreview := screens.IsList(current.Kind) && current.Focus == screens.FocusSecondary

	switch key {
	case "q":
		return m, tea.Quit
	case "1":
		m.setMode(screens.ModeRuns)
	case "2":
		m.setMode(screens.ModeAgents)
	case "3":
		m.setMode(screens.ModeTasks)
	case "tab":
		if current.Kind == screens.Attempt || (screens.IsList(current.Kind) && screens.TwoColumn(m.width)) {
			m.toggleFocus()
		} else {
			m.setMode((m.mode + 1) % screens.ModeCount)
		}
	case "esc", "backspace", "h":
		m.goBack()
	case "up", "k":
		switch {
		case inChat:
			m.moveChat(-1)
		case onPreview:
			m.scrollPreview(-1)
		default:
			m.move(-1)
		}
	case "down", "j":
		switch {
		case inChat:
			m.moveChat(1)
		case onPreview:
			m.scrollPreview(1)
		default:
			m.move(1)
		}
	case "pgup":
		switch {
		case inChat:
			m.scrollChat(-m.chatPage())
		case onPreview:
			m.scrollPreview(-m.previewPage())
		default:
			m.move(-m.pageSize())
		}
	case "pgdown":
		switch {
		case inChat:
			m.scrollChat(m.chatPage())
		case onPreview:
			m.scrollPreview(m.previewPage())
		default:
			m.move(m.pageSize())
		}
	case "home", "g":
		switch {
		case inChat:
			m.chatJump(0)
		case onPreview:
			m.previewJump(0)
		default:
			m.jump(0)
		}
	case "end", "G":
		switch {
		case inChat:
			m.chatJump(-1)
		case onPreview:
			m.previewJump(-1)
		default:
			m.jump(-1)
		}
	case "enter":
		if inChat {
			m.toggleSelectedChat()
		} else {
			m.open()
		}
	case "x", "space":
		if inChat {
			m.toggleSelectedChat()
		}
	case "c":
		if current.Kind == screens.Attempt {
			m.setFocus(screens.FocusSecondary)
		}
	}
	return m, nil
}

func (m *Model) handleWheel(msg tea.MouseWheelMsg) {
	mouse := msg.Mouse()
	var delta int
	switch mouse.Button {
	case tea.MouseWheelUp:
		delta = -3
	case tea.MouseWheelDown:
		delta = 3
	default:
		return
	}
	current := m.current()
	if current.Kind == screens.Attempt && current.Focus == screens.FocusSecondary {
		m.scrollChat(delta)
		return
	}
	if screens.IsList(current.Kind) && screens.TwoColumn(m.width) {
		if current.Focus == screens.FocusSecondary {
			m.scrollPreview(delta)
			return
		}
		if left, _ := screens.ColumnWidths(m.width); mouse.X >= left {
			m.scrollPreview(delta)
			return
		}
	}
	m.move(delta)
}

func (m *Model) handleClick(msg tea.MouseClickMsg) {
	if msg.Button != tea.MouseLeft {
		return
	}
	if m.showHelp {
		m.showHelp = false
		return
	}
	mouse := msg.Mouse()
	if m.backAt(mouse) {
		m.goBack()
		return
	}
	if mode, ok := m.tabAt(mouse); ok {
		m.setMode(mode)
		return
	}
	current := m.current()
	if current.Kind == screens.Attempt {
		m.clickAttempt(current, mouse)
		return
	}
	if !screens.IsList(current.Kind) {
		return
	}
	if screens.TwoColumn(m.width) {
		left, _ := screens.ColumnWidths(m.width)
		if mouse.X >= left {
			current.Focus = screens.FocusSecondary
			return
		}
		current.Focus = screens.FocusPrimary
	}
	paneWidth := screens.PaneWidth(m.width)
	top := headerHeight + m.view.HeadLines(current, max(1, paneWidth-1)) + 1
	if mouse.Y < top {
		return
	}
	row := mouse.Y - top + current.Offset
	count := m.view.ListCount(current)
	if row < 0 || row >= count {
		return
	}
	if current.Cursor == row {
		m.open()
		return
	}
	current.Cursor = row
	m.scrollCursorIntoView(current)
}

// backAt reports whether a click landed on the header back button.
func (m *Model) backAt(mouse tea.Mouse) bool {
	if mouse.Y != 0 {
		return false
	}
	start, end := m.backRange()
	return end > start && mouse.X >= start && mouse.X < end
}

// tabAt maps a click on the tab row to a view mode.
func (m *Model) tabAt(mouse tea.Mouse) (screens.Mode, bool) {
	if mouse.Y != headerHeight-2 {
		return 0, false
	}
	x := 0
	for current := screens.Mode(0); current < screens.ModeCount; current++ {
		label := fmt.Sprintf("%d %s", current+1, current.Label())
		width := len(label) + 2
		if mouse.X >= x && mouse.X < x+width {
			return current, true
		}
		x += width + 1
	}
	return 0, false
}

func (m *Model) clickAttempt(current *screens.Route, mouse tea.Mouse) {
	chatTop := 0
	if screens.TwoColumn(m.width) {
		leftWidth, _ := screens.ColumnWidths(m.width)
		if mouse.X < leftWidth {
			current.Focus = screens.FocusPrimary
			return
		}
		chatTop = headerHeight + 1
	} else {
		top := headerHeight + max(1, m.contentHeight()/2)
		if mouse.Y < top {
			current.Focus = screens.FocusPrimary
			return
		}
		chatTop = top + 1
	}
	current.Focus = screens.FocusSecondary
	layout, ok := m.view.ChatLayout(current, screens.PreviewWidth(m.width))
	if !ok || mouse.Y < chatTop {
		return
	}
	line := mouse.Y - chatTop + current.ChatOffset
	for index, span := range layout.CardSpans {
		if line >= span[0] && line < span[1] {
			current.Cursor = layout.CardOwners[index]
			m.toggleSelectedChat()
			return
		}
	}
}

func (m *Model) toggleFocus() {
	current := m.current()
	if current.Focus == screens.FocusSecondary {
		current.Focus = screens.FocusPrimary
		return
	}
	m.setFocus(screens.FocusSecondary)
}

func (m *Model) setFocus(focus int) {
	current := m.current()
	current.Focus = focus
	if focus == screens.FocusSecondary {
		m.chatSelectLast()
	}
}
