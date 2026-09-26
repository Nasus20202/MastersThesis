package screens

import (
	"fmt"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

// Totals renders the full-width aggregate stats page over every run.
func (v *View) Totals(route *Route, width, height int) string {
	width = max(1, width)
	lines := v.totalsLines(width)
	view := components.NewViewport()
	view.SetSize(width, height)
	view.SetLines(lines)
	view.SetOffset(route.Offset)
	return view.View()
}

// TotalsLineCount is the number of lines in the totals page for a width.
func (v *View) TotalsLineCount(width int) int {
	return len(v.totalsLines(max(1, width)))
}

func (v *View) totalsLines(width int) []string {
	if lines, ok := v.totals[width]; ok {
		return lines
	}
	lines := v.buildTotals(width)
	v.totals[width] = lines
	return lines
}

// buildTotals lays out the all-runs summary: no grouping, just the summed
// counts, tokens and time with their means and distributions.
func (v *View) buildTotals(width int) []string {
	metrics := model.TotalMetrics(v.store)
	contentWidth := max(20, width-2)
	lines := []string{
		components.Section(width, "All runs"),
		components.CardRow([]string{
			components.MetricCard("runs", fmt.Sprintf("%d", len(v.store.Runs())), ui.Section),
			components.MetricCard("agents", fmt.Sprintf("%d", len(v.store.Agents())), ui.Section),
			components.MetricCard("tasks", fmt.Sprintf("%d", len(v.store.Tasks())), ui.Section),
			components.MetricCard("attempts", fmt.Sprintf("%d", metrics.Attempts), ui.Section),
		}, contentWidth),
		"",
	}

	rate := model.OutcomeRate(metrics)
	lines = append(lines,
		components.Section(width, "Outcome"),
		components.CardRow([]string{
			components.MetricCard("full", fmt.Sprintf("%d", metrics.Full), ui.Success),
			components.MetricCard("partial", fmt.Sprintf("%d", metrics.Partial), ui.Warning),
			components.MetricCard("failed", fmt.Sprintf("%d", metrics.Failed), ui.Danger),
			components.MetricCard("full success", ui.Rate(rate), ui.Outcome(rate >= 1, rate)),
			components.MetricCard("mean score", ui.Rate(metrics.MeanScore), ui.Outcome(metrics.MeanScore >= 1, metrics.MeanScore)),
		}, contentWidth),
		"",
		ui.Donut(contentWidth, 5, outcomes(metrics)),
		"",
	)

	if len(metrics.Tokens) > 0 {
		cards := []string{
			components.MetricCard("total", ui.Count(model.SumInts(metrics.Tokens)), ui.Section),
			components.MetricCard("in", ui.Count(model.SumInts(metrics.Prompt)), ui.Section),
			components.MetricCard("out", ui.Count(model.SumInts(metrics.Completion)), ui.Section),
			components.MetricCard("cached", ui.Count(model.SumInts(metrics.Cached)), ui.Section),
			components.MetricCard("mean cache %", fmt.Sprintf("%.2f%%", model.Mean(metrics.CacheRatios)*100), ui.Section),
		}
		if throughput := model.PredictedTokensPerSecond(metrics); throughput > 0 {
			cards = append(cards, components.MetricCard("out t/s", fmt.Sprintf("%.1f", throughput), ui.Section))
		}
		if acceptance, ok := model.DraftAcceptanceRate(metrics); ok {
			cards = append(cards, components.MetricCard("draft accept %", fmt.Sprintf("%.0f%%", acceptance*100), ui.Section))
		}
		lines = append(lines, components.Section(width, "Tokens"), components.CardRow(cards, contentWidth), "")
	}

	if len(metrics.Durations) > 0 || len(metrics.Turns) > 0 {
		cards := []string{
			components.MetricCard("total time", ui.Seconds(model.Sum(metrics.Durations)), ui.Section),
			components.MetricCard("mean time", ui.Seconds(model.Mean(metrics.Durations)), ui.Section),
		}
		if len(metrics.Turns) > 0 {
			cards = append(cards,
				components.MetricCard("total turns", fmt.Sprintf("%d", model.SumInts(metrics.Turns)), ui.Section),
				components.MetricCard("mean turns", fmt.Sprintf("%.1f", model.MeanInts(metrics.Turns)), ui.Section),
			)
		}
		lines = append(lines, components.Section(width, "Time & turns"), components.CardRow(cards, contentWidth), "")
	}

	if len(metrics.Terminations) > 0 {
		lines = append(lines, components.Section(width, "Termination"), terminationBars(contentWidth, metrics.Terminations))
	}
	return lines
}
