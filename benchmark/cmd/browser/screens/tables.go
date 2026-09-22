package screens

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func (v *View) runsTable(runs []results.RunRef, route *Route, width, height int) string {
	tbl := components.NewTable(
		components.Column{Title: "RUN", Min: 18, Max: 40},
		components.Column{Title: "STATE", Min: 9, Max: 9},
		components.Column{Title: "AGENTS", Min: 7, Max: 16},
		components.Column{Title: "ATT", Min: 5, Max: 9},
		components.Column{Title: "SUCCESS", Min: 7, Max: 7},
	)
	rows := make([][]string, len(runs))
	for index, run := range runs {
		attempts, success := "0", "–"
		if run.Summary != nil {
			attempts = fmt.Sprintf("%d/%d", run.Summary.AttemptsRecorded, run.Summary.ExpectedAttempts)
			success = ui.Rate(overallRate(run.Summary))
		}
		rows[index] = []string{
			run.RunID,
			string(run.Metadata.State),
			strings.Join(run.Metadata.Agents, ","),
			attempts,
			success,
		}
	}
	styleFor := func(index int) lipgloss.Style { return ui.State(string(runs[index].Metadata.State)) }
	return tbl.Render(width, height, rows, route.Cursor, route.Offset, styleFor)
}

func (v *View) agentsTable(route *Route, width, height int) string {
	tbl := components.NewTable(
		components.Column{Title: "AGENT", Min: 8, Max: 16},
		components.Column{Title: "RUNS", Min: 4, Max: 6},
		components.Column{Title: "ATT", Min: 5, Max: 7},
		components.Column{Title: "FULL", Min: 6, Max: 7},
		components.Column{Title: "MEAN", Min: 6, Max: 7},
		components.Column{Title: "BAR", Min: 10, Max: 10},
	)
	agents := v.store.Agents()
	rows := make([][]string, len(agents))
	for index, agent := range agents {
		rows[index] = []string{
			agent.Agent,
			fmt.Sprintf("%d", agent.Runs),
			fmt.Sprintf("%d", agent.Attempts),
			ui.Rate(agent.FullSuccessRate),
			ui.Rate(agent.MeanScore),
			ui.Meter(agent.MeanScore, 10, ui.Outcome(agent.FullSuccessRate >= 1, agent.MeanScore)),
		}
	}
	styleFor := func(index int) lipgloss.Style {
		return ui.Outcome(agents[index].FullSuccessRate >= 1, agents[index].MeanScore)
	}
	return tbl.Render(width, height, rows, route.Cursor, route.Offset, styleFor)
}

func (v *View) tasksTable(route *Route, width, height int) string {
	tbl := components.NewTable(
		components.Column{Title: "TASK", Min: 20, Max: 44},
		components.Column{Title: "RUNS", Min: 4, Max: 6},
		components.Column{Title: "ATT", Min: 5, Max: 7},
		components.Column{Title: "FULL", Min: 6, Max: 7},
		components.Column{Title: "MEAN", Min: 6, Max: 7},
		components.Column{Title: "BAR", Min: 10, Max: 10},
	)
	tasks := v.store.Tasks()
	rows := make([][]string, len(tasks))
	for index, task := range tasks {
		rows[index] = []string{
			v.store.ScenarioTitle(task.ScenarioID),
			fmt.Sprintf("%d", task.Runs),
			fmt.Sprintf("%d", task.Attempts),
			ui.Rate(task.FullSuccessRate),
			ui.Rate(task.MeanScore),
			ui.Meter(task.MeanScore, 10, ui.Outcome(task.FullSuccessRate >= 1, task.MeanScore)),
		}
	}
	styleFor := func(index int) lipgloss.Style {
		return ui.Outcome(tasks[index].FullSuccessRate >= 1, tasks[index].MeanScore)
	}
	return tbl.Render(width, height, rows, route.Cursor, route.Offset, styleFor)
}

func (v *View) scenariosTable(route *Route, width, height int) string {
	rows := v.store.ScenarioRows(route.RunID, route.Agent, route.Task)
	tbl := components.NewTable(
		components.Column{Title: "SCENARIO", Min: 18, Max: 40},
		components.Column{Title: "ATT", Min: 5, Max: 7},
		components.Column{Title: "FULL", Min: 6, Max: 7},
		components.Column{Title: "MEAN", Min: 6, Max: 7},
		components.Column{Title: "BAR", Min: 10, Max: 10},
	)
	tableRows := make([][]string, len(rows))
	for index, scenario := range rows {
		tableRows[index] = []string{
			v.store.ScenarioTitle(scenario.ID),
			fmt.Sprintf("%d", scenario.Attempts),
			ui.Rate(fullRate(scenario.FullSuccess, scenario.Attempts)),
			ui.Rate(scenario.MeanScore),
			ui.Meter(scenario.MeanScore, 10, ui.Outcome(scenario.Attempts > 0 && scenario.FullSuccess == scenario.Attempts, scenario.MeanScore)),
		}
	}
	styleFor := func(index int) lipgloss.Style {
		scenario := rows[index]
		return ui.Outcome(scenario.Attempts > 0 && scenario.FullSuccess == scenario.Attempts, scenario.MeanScore)
	}
	return tbl.Render(width, height, tableRows, route.Cursor, route.Offset, styleFor)
}

func (v *View) attemptsTable(route *Route, width, height int) string {
	refs := v.store.AttemptsFor(route.RunID, route.ScenarioID, route.Group)
	tbl := components.NewTable(
		components.Column{Title: "ATT", Min: 3, Max: 5},
		components.Column{Title: "GROUP", Min: 8, Max: 16},
		components.Column{Title: "SCORE", Min: 6, Max: 7},
		components.Column{Title: "FULL", Min: 5, Max: 6},
		components.Column{Title: "DURATION", Min: 8, Max: 10},
	)
	rows := make([][]string, len(refs))
	styles := make([]lipgloss.Style, len(refs))
	for index, ref := range refs {
		attempt, err := v.store.Attempt(route.RunID, ref)
		if err != nil {
			rows[index] = []string{fmt.Sprintf("%d", ref.Attempt), ref.Group, "–", "–", "–"}
			styles[index] = ui.Danger
			continue
		}
		grading := attempt.Grading()
		styles[index] = ui.Outcome(grading.FullSuccess, grading.Score)
		rows[index] = []string{
			fmt.Sprintf("%d", ref.Attempt),
			ref.Group,
			ui.Rate(grading.Score),
			ui.YesNo(grading.FullSuccess),
			ui.Seconds(model.AttemptDuration(attempt)),
		}
	}
	styleFor := func(index int) lipgloss.Style { return styles[index] }
	return tbl.Render(width, height, rows, route.Cursor, route.Offset, styleFor)
}
