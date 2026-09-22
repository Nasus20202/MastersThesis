package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/screens"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

const (
	headerHeight = 3
	footerHeight = 1
)

// View implements tea.Model.
func (m *Model) View() tea.View {
	view := tea.NewView("")
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	if m.err != nil && len(m.store.Runs()) == 0 {
		view.SetContent(ui.Danger.Render("error: " + m.err.Error()))
		return view
	}
	if m.showHelp {
		view.SetContent(m.renderHelp())
		return view
	}
	view.SetContent(m.renderScreen())
	return view
}

func (m *Model) renderScreen() string {
	header := m.renderHeader()
	footer := m.renderFooter()
	height := m.contentHeight()
	body := ui.FitHeight(m.renderBody(height), height)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m *Model) renderBody(height int) string {
	current := m.current()
	if current.Kind == screens.Attempt {
		return m.view.Attempt(current, m.width, height)
	}
	if !screens.IsList(current.Kind) {
		return ""
	}
	left := m.view.List(current, m.width, height)
	if !screens.TwoColumn(m.width) {
		return left
	}
	return screens.Join(left, m.view.Right(current, m.width, height), m.width, height)
}

func (m *Model) contentHeight() int {
	return max(1, m.height-headerHeight-footerHeight)
}
