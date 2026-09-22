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

// CardRow lays metric cards out horizontally, wrapping onto further rows so
// the strip fits within width.
func CardRow(cards []string, width int) string {
	if len(cards) == 0 {
		return ""
	}
	width = max(1, width)
	const gap = 2
	var rows []string
	var current []string
	currentWidth := 0
	for _, card := range cards {
		cardWidth := lipgloss.Width(card)
		added := cardWidth
		if len(current) > 0 {
			added += gap
		}
		if len(current) > 0 && currentWidth+added > width {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, current...))
			current, currentWidth = nil, 0
			added = cardWidth
		}
		if len(current) > 0 {
			current = append(current, strings.Repeat(" ", gap))
		}
		current = append(current, card)
		currentWidth += added
	}
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, current...))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
