package app

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/screens"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func (m *Model) renderHeader() string {
	title := m.headerTitle()
	if m.canGoBack() {
		title += "  " + m.backLabel()
	}
	right := ui.MutedStyle.Render(fmt.Sprintf("%d runs · %d agents · %d tasks",
		len(m.store.Runs()), len(m.store.Agents()), len(m.store.Tasks())))
	if m.hasRunning() {
		right = ui.BadgeRunning.Render("LIVE") + " " + right
	}

	tabs := make([]string, 0, screens.ModeCount)
	for current := screens.Mode(0); current < screens.ModeCount; current++ {
		label := fmt.Sprintf("%d %s", current+1, current.Label())
		if current == m.mode {
			tabs = append(tabs, ui.TabActive.Render(label))
			continue
		}
		tabs = append(tabs, ui.TabIdle.Render(label))
	}
	hint := "tab switch"
	switch {
	case m.current().Kind == screens.Attempt:
		hint = "tab pane"
	case screens.IsList(m.current().Kind) && screens.TwoColumn(m.width) && m.current().Focus == screens.FocusSecondary:
		hint = "dashboard pane · tab back to list"
	case screens.IsList(m.current().Kind) && screens.TwoColumn(m.width):
		hint = "tab dashboard"
	}
	if m.canGoBack() {
		hint = "esc back · " + hint
	}
	hint += " · ? help · q quit"
	rule := ui.Rule.Render(strings.Repeat("─", max(1, m.width)))
	return ui.JoinSides(title, right, m.width) + "\n" +
		ui.JoinSides(strings.Join(tabs, " "), ui.FaintStyle.Render(hint), m.width) + "\n" + rule
}

func (m *Model) renderFooter() string {
	footer := help.New()
	footer.SetWidth(max(1, m.width-4))
	line := footer.ShortHelpView(ui.DefaultKeyMap.ShortHelp())
	if m.err != nil {
		line = ui.Danger.Render("error: "+m.err.Error()) + "  " + line
	}
	return ui.Footer.Width(max(1, m.width)).Render(" " + line)
}

func (m *Model) renderHelp() string {
	footer := help.New()
	footer.ShowAll = true
	footer.SetWidth(m.width)
	groups := footer.FullHelpView(ui.DefaultKeyMap.FullHelp())
	legend := components.CardRow([]string{
		components.MetricCard("full success", "100%", ui.Success),
		components.MetricCard("partial", "0–99%", ui.Warning),
		components.MetricCard("failed / error", "0%", ui.Danger),
		components.MetricCard("running", "live", ui.Running),
	}, m.width)
	header := ui.JoinSides(ui.Title.Render("◆ Keys"), m.closeLabel(), m.width)
	return header + "\n\n" + groups + "\n\n" +
		components.Section(m.width, "Legend") + "\n\n" + legend + "\n\n" +
		ui.MutedStyle.Render("press ? or esc to close")
}

func (m *Model) closeLabel() string {
	return ui.TabIdle.Render("[") + ui.AccentStyle.Bold(true).Render("× close") + ui.TabIdle.Render("]")
}

// headerTitle is the title and breadcrumb, shared by the renderer and the back
// button's hit box so they stay aligned.
func (m *Model) headerTitle() string {
	title := ui.AccentStyle.Render("◆") + " " + ui.Title.Render("benchmark results")
	if crumbs := m.breadcrumb(); crumbs != "" {
		title += "  " + ui.MutedStyle.Render(crumbs)
	}
	return title
}

func (m *Model) canGoBack() bool { return len(m.stack) > 1 }

func (m *Model) backLabel() string {
	return ui.TabIdle.Render("[") + ui.AccentStyle.Bold(true).Render("← back") + ui.TabIdle.Render("]")
}

// backRange returns the columns occupied by the header back button.
func (m *Model) backRange() (int, int) {
	if !m.canGoBack() {
		return 0, 0
	}
	start := lipgloss.Width(m.headerTitle()) + 2
	return start, start + lipgloss.Width(m.backLabel())
}

func (m *Model) goBack() {
	if len(m.stack) > 1 {
		m.stack = m.stack[:len(m.stack)-1]
	}
}

func (m *Model) hasRunning() bool {
	for _, run := range m.store.Runs() {
		if run.Metadata.State == results.RunStateRunning {
			return true
		}
	}
	return false
}

func (m *Model) breadcrumb() string {
	parts := make([]string, 0, len(m.stack))
	for _, route := range m.stack {
		if label := routeLabel(route); label != "" {
			parts = append(parts, label)
		}
	}
	if len(parts) > 4 {
		parts = append([]string{"…"}, parts[len(parts)-3:]...)
	}
	return strings.Join(parts, "  ›  ")
}

func routeLabel(route screens.Route) string {
	switch route.Kind {
	case screens.AgentRuns:
		return "agent " + route.Agent
	case screens.TaskRuns:
		return "task " + route.Task
	case screens.Run:
		return route.RunID
	case screens.Attempts:
		return route.ScenarioID
	case screens.Attempt:
		return fmt.Sprintf("attempt %d", route.Attempt)
	default:
		return ""
	}
}
