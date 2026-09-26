package screens

import (
	"fmt"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func (v *View) runsPreview(route *Route, width int) string {
	var runs []results.RunRef
	switch route.Kind {
	case AgentRuns:
		runs = v.store.RunsForAgent(route.Agent)
	case TaskRuns:
		runs = v.store.RunsForTask(route.Task)
	default:
		runs = v.store.Runs()
	}
	lines := []string{components.Section(width, "Across runs"), v.runComparison(width)}
	if route.Cursor < len(runs) {
		run := runs[route.Cursor]
		metrics := model.RunMetrics(v.store, run.RunID, route.Agent, route.Task)
		lines = append(lines,
			components.Section(width, "Selected run"),
			v.runPreview(run, width),
			"",
			components.CardRow(v.vitals(metrics), width),
			"",
			v.dashboardView(metrics, width),
		)
	}
	return strings.Join(lines, "\n")
}

func (v *View) runComparison(width int) string {
	runs := v.store.Runs()
	ranks := make([]ui.Rank, 0, len(runs))
	for _, run := range runs {
		value := 0.0
		if run.Summary != nil {
			value = overallRate(run.Summary)
		}
		ranks = append(ranks, ui.Rank{Label: runLabel(run), Value: value})
	}
	return ui.Ranked(ranks, width, 0)
}

func (v *View) runPreview(run results.RunRef, width int) string {
	body := []string{
		ui.Header.Render(run.RunID) + "  " + ui.StateBadge(string(run.Metadata.State)),
		ui.MutedStyle.Render(fmt.Sprintf("type %s · agents %s · repeat %d · parallel %d",
			run.Metadata.RunType, strings.Join(run.Metadata.Agents, ","), run.Metadata.RepeatCount, run.Metadata.Parallelism)),
		ui.MutedStyle.Render(fmt.Sprintf("started %s · rev %s", ui.Time(run.Metadata.StartedAt), ui.ShortRevision(run.Metadata.RepositoryRevision))),
	}
	if run.Summary != nil {
		body = append(body,
			fmt.Sprintf("attempts %d/%d · errors %d", run.Summary.AttemptsRecorded, run.Summary.ExpectedAttempts, run.Summary.ErrorCount),
			fmt.Sprintf("full success %s · mean %s", ui.Rate(overallRate(run.Summary)), ui.Rate(runMean(run.Summary))),
		)
	}
	return components.Panel("", strings.Join(body, "\n"), width, ui.Border)
}

func (v *View) agentPreview(route *Route, width int) string {
	agents := v.store.Agents()
	if route.Cursor >= len(agents) {
		return ""
	}
	ranks := make([]ui.Rank, 0, len(agents))
	for _, agent := range agents {
		ranks = append(ranks, ui.Rank{Label: agent.Agent, Value: agent.MeanScore})
	}
	metrics := model.AgentMetrics(v.store, agents[route.Cursor].Agent)
	lines := []string{
		components.Section(width, "Mean score by agent"),
		ui.Ranked(ranks, width, 0),
		"",
		components.Section(width, "Selected agent"),
		components.CardRow(v.vitals(metrics), width),
		"",
		v.dashboardView(metrics, width),
	}
	return strings.Join(lines, "\n")
}

func (v *View) taskPreview(route *Route, width int) string {
	tasks := v.store.Tasks()
	if route.Cursor >= len(tasks) {
		return ""
	}
	ranks := make([]ui.Rank, 0, len(tasks))
	for _, task := range tasks {
		ranks = append(ranks, ui.Rank{Label: v.store.ScenarioTitle(task.ScenarioID), Value: task.MeanScore})
	}
	metrics := model.TaskMetrics(v.store, tasks[route.Cursor].ScenarioID)
	task := tasks[route.Cursor]
	lines := []string{
		components.Section(width, "Mean score by task"),
		ui.Ranked(ranks, width, 0),
		"",
		components.Section(width, "Selected task"),
		components.CardRow(v.vitals(metrics), width),
		"",
		v.dashboardView(metrics, width),
		"",
		agentRankGraphs(model.TaskAgentMetrics(v.store, task.ScenarioID), width),
		"",
		components.Section(width, "Criteria by agent"),
		criteriaMatrix(model.TaskCriteria(v.store, task.ScenarioID), width),
	}
	return strings.Join(lines, "\n")
}

func (v *View) attemptPreview(route *Route, width int) string {
	refs := v.store.AttemptsFor(route.RunID, route.ScenarioID, route.Group)
	if route.Cursor >= len(refs) {
		return ""
	}
	attempt, err := v.store.Attempt(route.RunID, refs[route.Cursor])
	if err != nil {
		return ui.Danger.Render(err.Error())
	}
	grading := attempt.Grading()
	body := []string{ui.OutcomeBadge(grading.FullSuccess, grading.Score) + "  " + ui.MutedStyle.Render(fmt.Sprintf("score %s", ui.Rate(grading.Score)))}
	for _, criterion := range grading.Criteria {
		mark := ui.Success.Render("PASS")
		if !criterion.Passed {
			mark = ui.Danger.Render("FAIL")
		}
		body = append(body, fmt.Sprintf("%s %s", mark, criterion.ID))
	}
	if len(grading.Criteria) == 0 {
		body = append(body, ui.MutedStyle.Render("no criteria"))
	}
	lines := []string{components.Section(width, "Selected attempt"), components.Panel("", strings.Join(body, "\n"), width, ui.Border)}
	metrics := model.RunMetrics(v.store, route.RunID, "", route.ScenarioID)
	lines = append(lines, "", components.Section(width, "Task stats"), components.CardRow(v.vitals(metrics), width))
	return strings.Join(lines, "\n")
}
