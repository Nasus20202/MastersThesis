package screens

import (
	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func fullRate(full, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(full) / float64(total)
}

func overallRate(summary *results.RunSummary) float64 {
	total, full := 0, 0
	for _, condition := range summary.ByCondition {
		total += condition.AttemptCount
		full += condition.FullSuccessCount
	}
	return fullRate(full, total)
}

func errorStyle(errors string) lipgloss.Style {
	if errors == "0" {
		return ui.Success
	}
	return ui.Danger
}

func overflowStyle(overflow bool) lipgloss.Style {
	if overflow {
		return ui.Danger
	}
	return ui.Section
}

// runLabel is a short, sortable label for a run in comparison lists.
func runLabel(run results.RunRef) string {
	label := run.RunID
	if !run.Metadata.StartedAt.IsZero() {
		label = ui.Time(run.Metadata.StartedAt)
	}
	if run.Model != "" {
		label += " " + run.Model
	}
	return label
}

func modelName(model string) string {
	if model == "" {
		return "unknown"
	}
	return model
}
