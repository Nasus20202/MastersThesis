// Box, section and card primitives shared by the screens.
package components

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

// Section renders an accent-barred heading with a rule to the edge.
func Section(width int, title string) string {
	prefix := ui.AccentStyle.Render("▌") + " " + ui.Section.Render(title) + " "
	remaining := width - lipgloss.Width(prefix)
	if remaining < 2 {
		return prefix
	}
	return prefix + ui.Rule.Render(strings.Repeat("─", remaining))
}

// Panel draws a titled box around pre-formatted content.
func Panel(title, body string, width int, borderColor color.Color) string {
	inner := max(6, width-2)
	border := lipgloss.NewStyle().Foreground(borderColor)
	label := ""
	labelWidth := 0
	if title != "" {
		label = " " + ui.Header.Render(title) + " "
		labelWidth = lipgloss.Width(label)
	}
	top := border.Render("╭─") + label + border.Render(strings.Repeat("─", max(0, inner-1-labelWidth))+"╮")
	contentWidth := max(1, inner-2)
	lines := []string{top}
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		for _, part := range strings.Split(lipgloss.NewStyle().Width(contentWidth).Render(line), "\n") {
			lines = append(lines, border.Render("│ ")+ui.PadRight(part, contentWidth)+border.Render(" │"))
		}
	}
	lines = append(lines, border.Render("╰"+strings.Repeat("─", inner)+"╯"))
	return strings.Join(lines, "\n")
}

// MetricCard renders a small labelled value card.
func MetricCard(label, value string, valueStyle lipgloss.Style) string {
	return ui.Card.Render(ui.FaintStyle.Render(label) + "\n" + valueStyle.Bold(true).Render(value))
}

// CardRow lays metric cards out horizontally.
func CardRow(cards []string) string {
	parts := make([]string, 0, len(cards)*2)
	for index, card := range cards {
		if index > 0 {
			parts = append(parts, "  ")
		}
		parts = append(parts, card)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
