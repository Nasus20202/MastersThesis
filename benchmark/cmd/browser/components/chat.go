// Package components holds the browser's reusable terminal widgets: the
// conversation transcript, box/section primitives, the data table and the
// scrollable pane.
package components

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const (
	collapseLimit = 12
	systemLimit   = 4
)

// Layout is the flattened transcript: rendered lines plus the spans used for
// keyboard selection and click hit-testing.
type Layout struct {
	Lines        []string
	CardSpans    [][2]int
	CardOwners   []int
	MessageSpans [][2]int
}

// Build renders the messages into a layout for the given cursor and expansion
// state.
func Build(messages []inference.Message, cursor int, expanded map[int]bool, renderer *ui.Renderer, width int) Layout {
	return flatten(render(messages, expanded, renderer), cursor, width)
}

type card struct {
	owner      int
	title      string
	body       []string
	border     color.Color
	titleStyle lipgloss.Style
}

func render(messages []inference.Message, expanded map[int]bool, renderer *ui.Renderer) []card {
	cards := make([]card, 0, len(messages))
	for index, message := range messages {
		limit := collapseLimit
		if message.Role == "system" {
			limit = systemLimit
		}
		isExpanded := expanded[index]
		if strings.TrimSpace(message.Content) != "" {
			cards = append(cards, card{
				owner:      index,
				title:      message.Role,
				body:       collapse(renderer.Render(message.Content), limit, isExpanded),
				border:     roleBorder(message.Role),
				titleStyle: roleTitleStyle(message.Role),
			})
		}
		for _, call := range message.ToolCalls {
			cards = append(cards, card{
				owner:      index,
				title:      "→ " + call.Name,
				body:       collapse(ui.ToolCallLines(call, renderer), limit, isExpanded),
				border:     ui.Amber,
				titleStyle: ui.Warning,
			})
		}
	}
	return cards
}

func flatten(cards []card, cursor, width int) Layout {
	layout := Layout{CardSpans: make([][2]int, len(cards)), CardOwners: make([]int, len(cards))}
	messageSpans := make(map[int][2]int)
	maxOwner := -1
	for index, item := range cards {
		maxOwner = max(maxOwner, item.owner)
		start := len(layout.Lines)
		layout.Lines = append(layout.Lines, renderCard(item, width, item.owner == cursor)...)
		layout.Lines = append(layout.Lines, "")
		layout.CardSpans[index] = [2]int{start, len(layout.Lines)}
		layout.CardOwners[index] = item.owner
		if span, ok := messageSpans[item.owner]; ok {
			messageSpans[item.owner] = [2]int{span[0], len(layout.Lines)}
		} else {
			messageSpans[item.owner] = [2]int{start, len(layout.Lines)}
		}
	}
	layout.MessageSpans = make([][2]int, maxOwner+1)
	for owner, span := range messageSpans {
		layout.MessageSpans[owner] = span
	}
	return layout
}

func collapse(lines []string, limit int, expanded bool) []string {
	if expanded || len(lines) <= limit {
		return lines
	}
	collapsed := append([]string{}, lines[:limit]...)
	return append(collapsed, ui.FaintStyle.Render(fmt.Sprintf("… %d more lines (x to expand)", len(lines)-limit)))
}

// renderCard draws a rounded, titled box. Selected cards use the accent color.
func renderCard(item card, width int, selected bool) []string {
	inner := max(6, width-2)
	border := lipgloss.NewStyle().Foreground(item.border)
	if selected {
		border = lipgloss.NewStyle().Foreground(ui.Accent).Bold(true)
	}
	title := item.titleStyle.Render(item.title)
	if selected {
		title = lipgloss.NewStyle().Foreground(ui.Accent).Bold(true).Render(item.title)
	}
	label := " " + title + " "
	labelWidth := lipgloss.Width(label)
	if labelWidth > inner-2 {
		value := ui.TruncatePlain(item.title, inner-4)
		label = " " + value + " "
		labelWidth = lipgloss.Width(label)
	}
	top := border.Render("╭─") + label + border.Render(strings.Repeat("─", max(0, inner-1-labelWidth))+"╮")

	contentWidth := max(1, inner-2)
	lines := []string{top}
	for _, line := range item.body {
		wrapped := lipgloss.NewStyle().Width(contentWidth).Render(line)
		for _, part := range strings.Split(wrapped, "\n") {
			lines = append(lines, border.Render("│ ")+ui.PadRight(part, contentWidth)+border.Render(" │"))
		}
	}
	if len(item.body) == 0 {
		lines = append(lines, border.Render("│ ")+strings.Repeat(" ", contentWidth)+border.Render(" │"))
	}
	lines = append(lines, border.Render("╰"+strings.Repeat("─", inner)+"╯"))
	return lines
}

func roleBorder(role string) color.Color {
	switch role {
	case "assistant":
		return ui.Green
	case "user":
		return ui.Blue
	case "tool":
		return ui.Amber
	default:
		return ui.Border
	}
}

func roleTitleStyle(role string) lipgloss.Style {
	switch role {
	case "assistant":
		return ui.BadgeSuccess
	case "user":
		return ui.BadgeRunning
	case "tool":
		return ui.BadgeWarning
	default:
		return ui.BadgeNeutral
	}
}
