package screens

import (
	"fmt"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

func (v *View) runSummary(route *Route, width int) string {
	snapshot, ok := v.store.Snapshot(route.RunID)
	if !ok {
		return ui.MutedStyle.Render("loading run…")
	}
	metadata := snapshot.Metadata
	title := ui.Title.Render(metadata.RunID) + "  " + ui.StateBadge(string(metadata.State)) + "  " + ui.MutedStyle.Render(metadata.RunType)

	attempts := "0"
	errors := "0"
	if snapshot.Summary != nil {
		attempts = fmt.Sprintf("%d/%d", snapshot.Summary.AttemptsRecorded, snapshot.Summary.ExpectedAttempts)
		errors = fmt.Sprintf("%d", snapshot.Summary.ErrorCount)
	}
	cards := []string{
		components.MetricCard("started", ui.Time(metadata.StartedAt), ui.Running),
		components.MetricCard("attempts", attempts, ui.Section),
		components.MetricCard("errors", errors, errorStyle(errors)),
	}
	lines := []string{title, components.CardRow(cards, width)}
	if route.Task != "" {
		lines = append(lines, ui.MutedStyle.Render("task: "+v.store.ScenarioTitle(route.Task)))
	}
	lines = append(lines, components.Section(width, "Scenarios"))
	return strings.Join(lines, "\n")
}

// runDashboard is the right-pane dashboard for a run.
func (v *View) runDashboard(route *Route, width int) string {
	if _, ok := v.store.Snapshot(route.RunID); !ok {
		return ui.MutedStyle.Render("loading…")
	}
	metrics := model.RunMetrics(v.store, route.RunID, route.Agent, route.Task)
	lines := []string{components.CardRow(v.vitals(metrics), width), "", v.dashboardView(metrics, width)}
	rows := v.store.ScenarioRows(route.RunID, route.Agent, route.Task)
	if route.Cursor < len(rows) {
		scenario := rows[route.Cursor]
		lines = append(lines, "",
			components.Section(width, "Selected scenario"), v.scenarioPreview(scenario, width),
			"",
			components.Section(width, "Criteria by agent"),
			criteriaMatrix(model.ScenarioCriteria(v.store, route.RunID, scenario.ID), width),
		)
	}
	return strings.Join(lines, "\n")
}

func (v *View) scenarioPreview(row model.ScenarioRow, width int) string {
	body := []string{ui.Header.Render(v.store.ScenarioTitle(row.ID))}
	if definition, ok := v.store.Catalog()[row.ID]; ok && strings.TrimSpace(definition.Task) != "" {
		for _, line := range ui.WrapPreview(strings.TrimSpace(definition.Task), max(16, width-4), 2) {
			body = append(body, ui.MutedStyle.Render(line))
		}
	}
	body = append(body, fmt.Sprintf("attempts %d · full %s · mean %s",
		row.Attempts, ui.Rate(fullRate(row.FullSuccess, row.Attempts)), ui.Rate(row.MeanScore)))
	for _, agent := range row.Agents {
		body = append(body, fmt.Sprintf("%s %s", ui.PadRight(agent, 10),
			ui.Meter(row.MeanScore, 10, ui.Outcome(row.Attempts > 0 && row.FullSuccess == row.Attempts, row.MeanScore))))
	}
	return components.Panel("", strings.Join(body, "\n"), width, ui.Border)
}

func (v *View) attemptsHead(route *Route, width int) string {
	lines := []string{ui.Title.Render(v.store.ScenarioTitle(route.ScenarioID)) + "  " + ui.MutedStyle.Render(route.RunID)}
	if definition, ok := v.store.Catalog()[route.ScenarioID]; ok && strings.TrimSpace(definition.Task) != "" {
		for _, line := range ui.WrapPreview(strings.TrimSpace(definition.Task), max(20, width-4), 2) {
			lines = append(lines, ui.MutedStyle.Render(line))
		}
	}
	var scores []float64
	for _, ref := range v.store.AttemptsFor(route.RunID, route.ScenarioID, route.Group) {
		if attempt, err := v.store.Attempt(route.RunID, ref); err == nil {
			scores = append(scores, attempt.Grading().Score)
		}
	}
	if len(scores) > 1 {
		lines = append(lines, ui.MutedStyle.Render("score per attempt ")+
			ui.Sparkline(min(width-18, 50), 2, scores, ui.Success))
	}
	lines = append(lines, components.Section(width, "Attempts"))
	return strings.Join(lines, "\n")
}
