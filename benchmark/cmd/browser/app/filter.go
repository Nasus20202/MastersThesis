package app

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

type filterOption struct {
	key      string
	value    string
	selected bool
}

// filterRow is one filter body line; a negative option marks a key heading.
type filterRow struct {
	option int
	text   string
}

const filterChromeHeight = 3 // heading, spacer and help line

func (m *Model) openFilter() {
	options := m.view.Store().TagOptions()
	keys := make([]string, 0, len(options))
	for key := range options {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	active := m.view.Store().TagFilter()
	m.filterOptions = m.filterOptions[:0]
	for _, key := range keys {
		for _, value := range options[key] {
			m.filterOptions = append(m.filterOptions, filterOption{
				key:      key,
				value:    value,
				selected: containsValue(active[key], value),
			})
		}
	}
	m.filterCursor = 0
	m.filterOffset = 0
	m.filterOpen = true
}

func (m *Model) handleFilterKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "/":
		m.applyFilter()
	case "c", "r":
		m.resetFilter()
	case "up", "k":
		m.moveFilterCursor(-1)
	case "down", "j":
		m.moveFilterCursor(1)
	case "pgup":
		m.moveFilterCursor(-m.filterVisibleRows())
	case "pgdown":
		m.moveFilterCursor(m.filterVisibleRows())
	case "home", "g":
		m.moveFilterCursor(-1 << 30)
	case "end", "G":
		m.moveFilterCursor(1 << 30)
	case "enter", "space", " ", "x":
		m.toggleFilterOption(m.filterCursor)
	}
	return m, nil
}

func (m *Model) filterWheel(msg tea.MouseWheelMsg) {
	switch msg.Mouse().Button {
	case tea.MouseWheelUp:
		m.moveFilterCursor(-1)
	case tea.MouseWheelDown:
		m.moveFilterCursor(1)
	}
}

func (m *Model) filterClick(msg tea.MouseClickMsg) {
	if msg.Button != tea.MouseLeft {
		return
	}
	if msg.Mouse().Y == 0 {
		resetStart, resetEnd, closeStart, closeEnd := m.filterButtons()
		x := msg.Mouse().X
		switch {
		case x >= closeStart && x < closeEnd:
			m.applyFilter()
		case x >= resetStart && x < resetEnd:
			m.resetFilter()
		}
		return
	}
	rows, _ := m.filterRows()
	contentY := msg.Mouse().Y - 2
	if contentY < 0 || contentY >= m.filterVisibleRows() {
		return
	}
	rowIndex := m.filterOffset + contentY
	if rowIndex < 0 || rowIndex >= len(rows) {
		return
	}
	if option := rows[rowIndex].option; option >= 0 && option < len(m.filterOptions) {
		m.filterCursor = option
		m.toggleFilterOption(option)
	}
}

func (m *Model) toggleFilterOption(index int) {
	if index < 0 || index >= len(m.filterOptions) {
		return
	}
	m.filterOptions[index].selected = !m.filterOptions[index].selected
}

func (m *Model) resetFilter() {
	for index := range m.filterOptions {
		m.filterOptions[index].selected = false
	}
}

func (m *Model) moveFilterCursor(delta int) {
	if len(m.filterOptions) == 0 {
		return
	}
	m.filterCursor = max(0, min(len(m.filterOptions)-1, m.filterCursor+delta))
	m.ensureFilterCursorVisible()
}

func (m *Model) ensureFilterCursorVisible() {
	rows, optionRows := m.filterRows()
	if m.filterCursor >= len(optionRows) {
		return
	}
	row := optionRows[m.filterCursor]
	visible := m.filterVisibleRows()
	if row < m.filterOffset {
		m.filterOffset = row
	}
	if row >= m.filterOffset+visible {
		m.filterOffset = row - visible + 1
	}
	m.filterOffset = max(0, min(m.filterOffset, max(0, len(rows)-visible)))
}

func (m *Model) filterVisibleRows() int {
	return max(1, m.height-filterChromeHeight)
}

func (m *Model) filterRows() ([]filterRow, []int) {
	rows := make([]filterRow, 0, len(m.filterOptions)+4)
	optionRows := make([]int, len(m.filterOptions))
	lastKey := ""
	for index, option := range m.filterOptions {
		if option.key != lastKey {
			rows = append(rows, filterRow{option: -1, text: option.key})
			lastKey = option.key
		}
		mark := "[ ]"
		if option.selected {
			mark = "[x]"
		}
		optionRows[index] = len(rows)
		rows = append(rows, filterRow{option: index, text: fmt.Sprintf("%s %s", mark, option.value)})
	}
	return rows, optionRows
}

func (m *Model) applyFilter() {
	filter := scenario.TagFilter{}
	for _, option := range m.filterOptions {
		if option.selected {
			filter[option.key] = append(filter[option.key], option.value)
		}
	}
	m.view.Store().SetTagFilter(filter)
	m.filterOpen = false
	m.view.Invalidate()
	m.clamp()
}

func (m *Model) renderFilter() string {
	rows, _ := m.filterRows()
	visible := m.filterVisibleRows()
	lines := []string{
		ui.JoinSides(ui.Title.Render("◆ filter by tags"), m.filterButtonsLabel(), m.width),
		"",
	}
	if len(rows) == 0 {
		lines = append(lines, ui.MutedStyle.Render("  no tags found in the scenario catalogue"))
	}
	for offset := 0; offset < visible; offset++ {
		index := m.filterOffset + offset
		if index >= len(rows) {
			lines = append(lines, "")
			continue
		}
		row := rows[index]
		if row.option < 0 {
			lines = append(lines, components.Section(m.width, row.text))
			continue
		}
		if row.option == m.filterCursor {
			lines = append(lines, ui.AccentStyle.Render("▸ "+row.text))
			continue
		}
		lines = append(lines, "  "+row.text)
	}
	lines = append(lines, ui.MutedStyle.Render("↑/↓ move · space toggle · c clear · esc apply"))
	return strings.Join(lines, "\n")
}

func (m *Model) filterButtonsLabel() string {
	return m.filterResetLabel() + " " + m.filterCloseLabel()
}

func (m *Model) filterResetLabel() string {
	return ui.TabIdle.Render("[") + ui.AccentStyle.Bold(true).Render("↺ reset") + ui.TabIdle.Render("]")
}

func (m *Model) filterCloseLabel() string {
	return ui.TabIdle.Render("[") + ui.AccentStyle.Bold(true).Render("× close") + ui.TabIdle.Render("]")
}

func (m *Model) filterButtons() (resetStart, resetEnd, closeStart, closeEnd int) {
	closeLabel := m.filterCloseLabel()
	resetLabel := m.filterResetLabel()
	closeEnd = m.width
	closeStart = closeEnd - lipgloss.Width(closeLabel)
	resetEnd = closeStart - 1
	resetStart = resetEnd - lipgloss.Width(resetLabel)
	return resetStart, resetEnd, closeStart, closeEnd
}

func containsValue(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
