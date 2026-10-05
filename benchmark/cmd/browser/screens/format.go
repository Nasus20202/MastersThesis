package screens

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func overallRate(summary *results.RunSummary) float64 {
	total, full := 0, 0
	for _, condition := range summary.ByCondition {
		total += condition.AttemptCount
		full += condition.FullSuccessCount
	}
	return model.Ratio(full, total)
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
		label += " " + shortModel(run.Model)
	}
	return label
}

func modelName(model string) string {
	if model == "" {
		return "unknown"
	}
	return model
}

// shortModel drops the instruction-tuning and quantization suffixes of a model
// name, e.g. gemma-4-E4B-it-qat-UD-Q4_K_XL becomes gemma-4-E4B.
func shortModel(name string) string {
	for _, marker := range []string{"-it-", "-it", "-Q"} {
		if before, _, ok := strings.Cut(name, marker); ok && before != "" {
			return before
		}
	}
	return name
}
