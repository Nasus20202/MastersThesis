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

func runMean(summary *results.RunSummary) float64 {
	total, score := 0, 0.0
	for _, condition := range summary.ByCondition {
		total += condition.AttemptCount
		score += condition.MeanScore * float64(condition.AttemptCount)
	}
	if total == 0 {
		return 0
	}
	return score / float64(total)
}

func errorStyle(errors string) lipgloss.Style {
	if errors == "0" {
		return ui.Success
	}
	return ui.Danger
}

// runLabel is a short, sortable label for a run in comparison lists.
func runLabel(run results.RunRef) string {
	if !run.Metadata.StartedAt.IsZero() {
		return ui.Time(run.Metadata.StartedAt)
	}
	return run.RunID
}
