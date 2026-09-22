package screens

import (
	"fmt"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

// Attempt renders the merged grading-details and conversation screen.
func (v *View) Attempt(route *Route, width, height int) string {
	if TwoColumn(width) {
		leftWidth, rightWidth := ColumnWidths(width)
		left := v.attemptDetails(route, leftWidth, height)
		right := v.conversationPane(route, rightWidth, height)
		return Join(left, right, width, height)
	}
	top := max(1, height/2)
	bottom := max(1, height-top)
	left := ui.FitHeight(v.attemptDetails(route, width, top), top)
	right := ui.FitHeight(v.conversationPane(route, width, bottom), bottom)
	return left + "\n" + right
}

func (v *View) attemptDetails(route *Route, width, height int) string {
	contentWidth := max(1, width-1)
	attempt, err := v.store.Attempt(route.RunID, route.Ref())
	if err != nil {
		return ui.Danger.Render(err.Error())
	}
	head := v.DetailsHead(route)
	body := v.attemptBodyLines(attempt, attempt.Grading(), contentWidth)
	visible := max(1, height-ui.LineCount(head))
	return components.Frame(head+"\n"+ui.Window(body, route.Offset, visible), contentWidth, height, len(body), route.Offset, visible)
}

func (v *View) buildDetailsHead(route *Route) string {
	attempt, err := v.store.Attempt(route.RunID, route.Ref())
	title := ui.Title.Render(v.store.ScenarioTitle(route.ScenarioID)) + "  " +
		ui.MutedStyle.Render(fmt.Sprintf("attempt %d · %s · %s", route.Attempt, route.Group, route.RunID))
	if err != nil {
		return title + "\n" + ui.Danger.Render(err.Error())
	}
	grading := attempt.Grading()
	summary := ui.OutcomeBadge(grading.FullSuccess, grading.Score) + "  " +
		ui.MutedStyle.Render(fmt.Sprintf("score %s · duration %s", ui.Rate(grading.Score), ui.Seconds(model.AttemptDuration(attempt))))
	if attempt.Error() != "" {
		summary += "  " + ui.BadgeDanger.Render("ERROR") + " " + ui.Danger.Render(ui.FirstLine(attempt.Error()))
	}
	var cards []string
	if attempt.Benchmark != nil && attempt.Benchmark.Agent != nil {
		agent := attempt.Benchmark.Agent
		cards = []string{
			components.MetricCard("turns", fmt.Sprintf("%d", agent.Turns), ui.Section),
			components.MetricCard("tool calls", fmt.Sprintf("%d", agent.ToolCallCount), ui.Section),
			components.MetricCard("tokens", fmt.Sprintf("%d", agent.TokenUsage.TotalTokens), ui.Section),
			components.MetricCard("termination", agent.Termination, ui.Termination(agent.Termination)),
		}
	} else if attempt.Validation != nil {
		cards = []string{
			components.MetricCard("case", attempt.Validation.CaseID, ui.Section),
			components.MetricCard("expected", ui.Rate(attempt.Validation.ExpectedScore), ui.Section),
			components.MetricCard("passed", ui.YesNo(attempt.Validation.Passed), ui.Outcome(attempt.Validation.Passed, boolScore(attempt.Validation.Passed))),
		}
	}
	head := title + "\n" + summary
	if len(cards) > 0 {
		head += "\n" + components.CardRow(cards)
	}
	return head
}

func (v *View) attemptBodyLines(attempt results.Attempt, grading orchestration.GradingResult, width int) []string {
	lines := []string{components.Section(width, "Grading criteria")}
	for _, criterion := range grading.Criteria {
		badge := ui.BadgeSuccess.Render("PASS")
		if !criterion.Passed {
			badge = ui.BadgeDanger.Render("FAIL")
		}
		header := badge + "  " + ui.Section.Render(criterion.ID) +
			ui.MutedStyle.Render(fmt.Sprintf("   weight %s · exit %d · %s", ui.Float(criterion.Weight), criterion.ExitCode, ui.Seconds(criterion.DurationSeconds)))
		body := []string{header}
		body = append(body, indented(criterion.Stdout)...)
		body = append(body, indented(criterion.Stderr)...)
		lines = append(lines, ui.Card.Render(strings.Join(body, "\n")))
	}
	if len(grading.Criteria) == 0 {
		lines = append(lines, ui.MutedStyle.Render("  none recorded"))
	}
	if failure := attempt.Failure(); failure != nil {
		lines = append(lines, components.Section(width, "Lifecycle failure"))
		header := ui.BadgeDanger.Render(failure.Phase) + "  " +
			ui.MutedStyle.Render(fmt.Sprintf("command %d · %s %s · exit %d", failure.CommandIndex, failure.Program, strings.Join(failure.Args, " "), failure.ExitCode))
		body := []string{header}
		body = append(body, indented(failure.Stdout)...)
		body = append(body, indented(failure.Stderr)...)
		lines = append(lines, ui.Card.Render(strings.Join(body, "\n")))
	}
	lines = append(lines, components.Section(width, "Execution"))
	lines = append(lines, executionLines(attempt)...)
	return lines
}

func executionLines(attempt results.Attempt) []string {
	if attempt.Benchmark != nil && attempt.Benchmark.Agent != nil {
		agent := attempt.Benchmark.Agent
		inference := agent.Inference
		return []string{
			ui.MutedStyle.Render(fmt.Sprintf("  model      %s", inference.Model)),
			ui.MutedStyle.Render(fmt.Sprintf("  artifact   %s", inference.Artifact)),
			ui.MutedStyle.Render(fmt.Sprintf("  condition  %s", agent.Condition)),
			ui.MutedStyle.Render(fmt.Sprintf("  loop       max %d turns · max %d tool calls · timeout %.0fs", agent.LoopConfig.MaxTurns, agent.LoopConfig.MaxToolCalls, agent.LoopConfig.TimeoutSeconds)),
		}
	}
	if attempt.Validation != nil {
		return []string{ui.MutedStyle.Render(fmt.Sprintf("  case %s", attempt.Validation.CaseID))}
	}
	return []string{ui.MutedStyle.Render("  no execution details")}
}

func (v *View) conversationPane(route *Route, width, height int) string {
	contentWidth := max(1, width-1)
	title := components.Section(contentWidth, "Conversation")
	view := components.NewViewport()
	view.SetSize(contentWidth, height)
	layout, ok := v.ChatLayout(route, contentWidth)
	if !ok {
		body := title + "\n" + ui.MutedStyle.Render("  no conversation recorded for this attempt")
		view.SetLines(strings.Split(body, "\n"))
		view.SetOffset(0)
		return view.View()
	}
	view.SetLines(append([]string{title}, layout.Lines...))
	view.SetOffset(route.ChatOffset)
	return view.View()
}

func indented(block string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(block, "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, "    "+ui.MutedStyle.Render(line))
	}
	return lines
}

func boolScore(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
