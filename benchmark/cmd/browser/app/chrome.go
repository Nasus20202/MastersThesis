package app

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
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
	right, _, _ := m.headerRight()

	tabs := make([]string, 0, screens.ModeCount)
	for current := screens.Mode(0); current < screens.ModeCount; current++ {
		label := fmt.Sprintf("%d %s", current+1, current.Label())
		if current == m.mode {
			tabs = append(tabs, ui.TabActive.Render(label))
			continue
		}
		tabs = append(tabs, ui.TabIdle.Render(label))
	}
	rule := ui.Rule.Render(strings.Repeat("─", max(1, m.width)))
	return ui.JoinSides(title, right, m.width) + "\n" +
		strings.Join(tabs, " ") + "\n" + rule
}

func (m *Model) renderFooter() string {
	line := shortHelp(ui.DefaultKeyMap.ShortHelp(), max(1, m.width-2))
	if m.err != nil {
		line = ui.Danger.Render("error: "+m.err.Error()) + "  " + line
	}
	return ui.Footer.Width(max(1, m.width)).Render(" " + line)
}

// shortHelp styles every segment with the footer background so the bar stays a
// single colour.
func shortHelp(bindings []key.Binding, limit int) string {
	keyStyle := lipgloss.NewStyle().Foreground(ui.Accent).Background(ui.Bar)
	descStyle := lipgloss.NewStyle().Foreground(ui.Muted).Background(ui.Bar)
	sepStyle := lipgloss.NewStyle().Foreground(ui.Faint).Background(ui.Bar)
	var builder strings.Builder
	width := 0
	for _, binding := range bindings {
		if !binding.Enabled() {
			continue
		}
		item := keyStyle.Render(binding.Help().Key) + descStyle.Render(" "+binding.Help().Desc)
		if width > 0 {
			item = sepStyle.Render(" • ") + item
		}
		itemWidth := lipgloss.Width(item)
		if width > 0 && width+itemWidth > limit {
			break
		}
		width += itemWidth
		builder.WriteString(item)
	}
	return builder.String()
}

func (m *Model) renderHelp() string {
	footer := help.New()
	footer.Styles.ShortKey = lipgloss.NewStyle().Foreground(ui.Accent)
	footer.Styles.ShortDesc = lipgloss.NewStyle().Foreground(ui.Muted)
	footer.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(ui.Faint)
	footer.Styles.FullKey = lipgloss.NewStyle().Foreground(ui.Text)
	footer.Styles.FullDesc = ui.MutedStyle
	footer.Styles.FullSeparator = ui.FaintStyle
	footer.ShowAll = true
	footer.SetWidth(m.width)
	groups := footer.FullHelpView(ui.DefaultKeyMap.FullHelp())
	legend := components.CardRow([]string{
		components.MetricCard("full success", "100%", ui.Success),
		components.MetricCard("partial", "0–99%", ui.Warning),
		components.MetricCard("failed / error", "0%", ui.Danger),
		components.MetricCard("running", "live", ui.Running),
	}, m.width)
	header := ui.JoinSides(ui.Title.Render("◆ Keys"), ui.Button(closeLabel), m.width)
	return header + "\n\n" + groups + "\n\n" +
		components.Section(m.width, "Legend") + "\n\n" + legend + "\n\n" +
		ui.MutedStyle.Render("press ? or esc to close")
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

// headerRight returns the header's right side and the filter control's range.
func (m *Model) headerRight() (text string, filterStart, filterEnd int) {
	control := m.filterControlLabel()
	counts := ui.MutedStyle.Render(fmt.Sprintf("%d runs · %d agents · %d tasks",
		len(m.store.Runs()), len(m.store.Agents()), len(m.store.Tasks())))
	if m.hasRunning() {
		counts = ui.BadgeRunning.Render("LIVE") + " " + counts
	}
	text = control + "  " + counts
	filterStart = m.width - lipgloss.Width(text)
	filterEnd = filterStart + lipgloss.Width(control)
	return text, filterStart, filterEnd
}

func (m *Model) filterControlLabel() string {
	store := m.view.Store()
	var parts []string
	if models := store.ModelFilter(); len(models) > 0 {
		parts = append(parts, "model="+strings.Join(models, ","))
	}
	if filter := store.TagFilter(); !filter.Empty() {
		parts = append(parts, filter.String())
	}
	if len(parts) == 0 {
		return ui.Button("⌕ filter")
	}
	return ui.TabActive.Render(" ⌕ " + ui.Truncate(strings.Join(parts, " "), 44) + " ")
}

func (m *Model) filterControlAt(x int) bool {
	_, start, end := m.headerRight()
	return end > start && x >= start && x < end
}

func (m *Model) canGoBack() bool { return len(m.stack) > 1 }

func (m *Model) backLabel() string { return ui.Button("← back") }

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
