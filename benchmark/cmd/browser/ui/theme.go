// The design system: one accent, a muted palette and
// outcome colors so the same meaning keeps the same color on every screen.
package ui

import "charm.land/lipgloss/v2"

var (
	Accent   = lipgloss.Color("63")
	Border   = lipgloss.Color("238")
	Bar      = lipgloss.Color("236")
	Text     = lipgloss.Color("252")
	Muted    = lipgloss.Color("245")
	Faint    = lipgloss.Color("240")
	Ink      = lipgloss.Color("16")
	Green    = lipgloss.Color("42")
	Amber    = lipgloss.Color("214")
	Red      = lipgloss.Color("203")
	Blue     = lipgloss.Color("39")
	Selected = lipgloss.Color("237")

	Title       = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	Header      = lipgloss.NewStyle().Bold(true).Foreground(Text)
	Section     = lipgloss.NewStyle().Bold(true).Foreground(Text)
	AccentStyle = lipgloss.NewStyle().Foreground(Accent)
	MutedStyle  = lipgloss.NewStyle().Foreground(Muted)
	FaintStyle  = lipgloss.NewStyle().Foreground(Faint)
	Rule        = lipgloss.NewStyle().Foreground(Border)
	Success     = lipgloss.NewStyle().Foreground(Green)
	Warning     = lipgloss.NewStyle().Foreground(Amber)
	Danger      = lipgloss.NewStyle().Foreground(Red)
	Running     = lipgloss.NewStyle().Foreground(Blue)

	TableHeader = lipgloss.NewStyle().Bold(true).Foreground(Text).Background(Bar)
	SelectedRow = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(Selected)
	Gutter      = lipgloss.NewStyle().Foreground(Accent).Bold(true)
	Card        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Border).Padding(0, 1)
	Footer      = lipgloss.NewStyle().Foreground(Text).Background(Bar)
	TabActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(Accent).Padding(0, 1)
	TabIdle     = lipgloss.NewStyle().Foreground(Muted).Padding(0, 1)

	BadgeSuccess = lipgloss.NewStyle().Bold(true).Foreground(Ink).Background(Green).Padding(0, 1)
	BadgeWarning = lipgloss.NewStyle().Bold(true).Foreground(Ink).Background(Amber).Padding(0, 1)
	BadgeDanger  = lipgloss.NewStyle().Bold(true).Foreground(Ink).Background(Red).Padding(0, 1)
	BadgeRunning = lipgloss.NewStyle().Bold(true).Foreground(Ink).Background(Blue).Padding(0, 1)
	BadgeNeutral = lipgloss.NewStyle().Bold(true).Foreground(Text).Background(Border).Padding(0, 1)
)

// Outcome colors a value by how close to success it is.
func Outcome(fullSuccess bool, score float64) lipgloss.Style {
	switch {
	case fullSuccess:
		return Success
	case score > 0:
		return Warning
	default:
		return Danger
	}
}

// State colors a run state.
func State(state string) lipgloss.Style {
	switch state {
	case "completed":
		return Success
	case "running":
		return Running
	default:
		return Warning
	}
}

// StateBadge renders a run state as a filled badge.
func StateBadge(state string) string {
	switch state {
	case "completed":
		return BadgeSuccess.Render("COMPLETED")
	case "running":
		return BadgeRunning.Render("RUNNING")
	default:
		return BadgeWarning.Render("INCOMPLETE")
	}
}

// OutcomeBadge renders a grading outcome as a filled badge.
func OutcomeBadge(fullSuccess bool, score float64) string {
	switch {
	case fullSuccess:
		return BadgeSuccess.Render("FULL SUCCESS")
	case score > 0:
		return BadgeWarning.Render("PARTIAL")
	default:
		return BadgeDanger.Render("FAILED")
	}
}

// Termination colors a termination reason.
func Termination(reason string) lipgloss.Style {
	switch reason {
	case "completed":
		return Success
	case "cancellation", "timeout", "inference":
		return Danger
	default:
		return Warning
	}
}
