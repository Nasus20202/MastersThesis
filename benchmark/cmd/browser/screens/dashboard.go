package screens

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
)

// dashboardView renders the full dashboard for a metrics set.
func (v *View) dashboardView(metrics model.Metrics, width int) string {
	chartWidth := max(20, width-2)
	var lines []string

	lines = append(lines, components.Section(width, "Outcome mix"))
	lines = append(lines, ui.Donut(chartWidth, 5, outcomes(metrics)))
	lines = append(lines, "")

	lines = append(lines, components.Section(width, "Speed distribution"))
	if labels, counts := histogram(metrics.Durations, 5, durationLabel); len(counts) > 0 {
		lines = append(lines, ui.Histogram(chartWidth, 5, labels, counts))
	} else {
		lines = append(lines, ui.MutedStyle.Render("no durations recorded"))
	}
	lines = append(lines, "")

	if len(metrics.Turns) > 0 {
		lines = append(lines, components.Section(width, "Turns distribution"))
		labels, counts := histogram(intsToFloats(metrics.Turns), 5, wholeNumber)
		lines = append(lines, ui.Histogram(chartWidth, 5, labels, counts))
		lines = append(lines, "")
	}

	if len(metrics.Terminations) > 0 {
		lines = append(lines, components.Section(width, "Termination"))
		lines = append(lines, terminationBars(chartWidth, metrics.Terminations))
	}
	return strings.Join(lines, "\n")
}

// vitals renders the headline metric cards used above a dashboard.
func (v *View) vitals(metrics model.Metrics) []string {
	rate := model.OutcomeRate(metrics)
	cards := []string{
		components.MetricCard("attempts", fmt.Sprintf("%d", metrics.Attempts), ui.Section),
		components.MetricCard("full success", ui.Rate(rate), ui.Outcome(rate >= 1, rate)),
		components.MetricCard("mean score", ui.Rate(metrics.MeanScore), ui.Outcome(metrics.MeanScore >= 1, metrics.MeanScore)),
	}
	if len(metrics.Durations) > 0 {
		cards = append(cards, components.MetricCard("mean time", ui.Seconds(model.Mean(metrics.Durations)), ui.Section))
	}
	if len(metrics.Turns) > 0 {
		cards = append(cards, components.MetricCard("mean turns", fmt.Sprintf("%.1f", model.MeanInts(metrics.Turns)), ui.Section))
	}
	if len(metrics.Tokens) > 0 {
		cards = append(cards, components.MetricCard("mean tokens", fmt.Sprintf("%.0f", model.MeanInts(metrics.Tokens)), ui.Section))
	}
	if len(metrics.Prompt) > 0 {
		cards = append(cards, components.MetricCard("mean in", fmt.Sprintf("%.0f", model.MeanInts(metrics.Prompt)), ui.Section))
	}
	if len(metrics.Completion) > 0 {
		cards = append(cards, components.MetricCard("mean out", fmt.Sprintf("%.0f", model.MeanInts(metrics.Completion)), ui.Section))
	}
	if len(metrics.Cached) > 0 {
		cards = append(cards, components.MetricCard("mean cached", fmt.Sprintf("%.0f", model.MeanInts(metrics.Cached)), ui.Section))
	}
	if len(metrics.CacheRatios) > 0 {
		cards = append(cards, components.MetricCard("mean cache %", fmt.Sprintf("%.2f%%", model.Mean(metrics.CacheRatios)*100), ui.Section))
	}
	return cards
}

func outcomes(metrics model.Metrics) []ui.Segment {
	return []ui.Segment{
		{Label: "full", Value: float64(metrics.Full), Style: ui.Success},
		{Label: "partial", Value: float64(metrics.Partial), Style: ui.Warning},
		{Label: "failed", Value: float64(metrics.Failed), Style: ui.Danger},
	}
}

func terminationBars(width int, terminations map[string]int) string {
	names := make([]string, 0, len(terminations))
	maximum := 0
	for name, count := range terminations {
		names = append(names, name)
		maximum = max(maximum, count)
	}
	sort.SliceStable(names, func(i, j int) bool { return terminations[names[i]] > terminations[names[j]] })
	barWidth := max(8, width/2)
	var lines []string
	for _, name := range names {
		value := float64(terminations[name]) / float64(max(1, maximum))
		lines = append(lines, ui.PadRight(name, 14)+
			ui.Meter(value, barWidth, ui.Outcome(name == common.TerminationCompleted, value))+
			ui.MutedStyle.Render(fmt.Sprintf(" %d", terminations[name])))
	}
	return strings.Join(lines, "\n")
}

// criteriaMatrix renders a criterion × condition grid so the checks that
// separate conditions on one scenario are visible at a glance.
func criteriaMatrix(matrix model.CriteriaMatrix, width int) string {
	if len(matrix.Rows) == 0 {
		return ui.MutedStyle.Render("no criteria recorded")
	}
	labelWidth := min(28, max(12, width/3))
	columnWidth := 8
	if len(matrix.Agents) > 0 {
		columnWidth = max(7, (width-labelWidth)/len(matrix.Agents))
	}

	header := ui.PadRight(ui.Header.Render("CRITERION"), labelWidth)
	for _, agent := range matrix.Agents {
		header += ui.PadRight(ui.Header.Render(ui.Truncate(agent, columnWidth-1)), columnWidth)
	}
	lines := []string{header}
	for _, row := range matrix.Rows {
		line := ui.PadRight(ui.Truncate(row.ID, labelWidth-1), labelWidth)
		for _, agent := range matrix.Agents {
			stat := row.Cells[agent]
			label, style := "–", ui.MutedStyle
			if stat.Total > 0 {
				rate := model.Ratio(stat.Passed, stat.Total)
				label, style = fmt.Sprintf("%d/%d", stat.Passed, stat.Total), ui.Outcome(rate >= 1, rate)
			}
			line += ui.PadRight(style.Render(label), columnWidth)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// histogram splits values into at most buckets ranges spanning the observed
// minimum and maximum, so the chart adapts to the data instead of fixed edges.
func histogram(values []float64, buckets int, format func(float64) string) ([]string, []int) {
	if len(values) == 0 {
		return nil, nil
	}
	minimum, maximum := values[0], values[0]
	for _, value := range values {
		minimum = min(minimum, value)
		maximum = max(maximum, value)
	}
	if maximum <= minimum {
		return []string{format(minimum)}, []int{len(values)}
	}
	buckets = max(1, buckets)
	step := (maximum - minimum) / float64(buckets)
	labels := make([]string, buckets)
	counts := make([]int, buckets)
	for index := range buckets {
		low := minimum + float64(index)*step
		high := minimum + float64(index+1)*step
		labels[index] = format(low) + "–" + format(high)
	}
	labels[buckets-1] = format(minimum+float64(buckets-1)*step) + "–" + format(maximum)
	for _, value := range values {
		index := int((value - minimum) / step)
		index = max(0, min(buckets-1, index))
		counts[index]++
	}
	return labels, counts
}

func durationLabel(value float64) string {
	if value < 1 {
		return "0s"
	}
	return ui.Seconds(value)
}

func wholeNumber(value float64) string {
	return fmt.Sprintf("%.0f", value)
}

func intsToFloats(values []int) []float64 {
	out := make([]float64, len(values))
	for index, value := range values {
		out[index] = float64(value)
	}
	return out
}
