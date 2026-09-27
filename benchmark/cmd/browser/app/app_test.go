package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/screens"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

func press(key string) tea.KeyPressMsg {
	switch key {
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace}
	default:
		runes := []rune(key)
		return tea.KeyPressMsg{Code: runes[0], Text: key}
	}
}

func createRun(t *testing.T, root, runID string, startedAt time.Time) {
	t.Helper()
	store, err := results.New(root, results.RunMetadata{
		RunID: runID, StartedAt: startedAt,
		Agents: []string{"skill"}, Parallelism: 1, RepeatCount: 1,
		Scenarios: []string{"scenario-alpha"},
	})
	require.NoError(t, err)
	require.NoError(t, store.WriteAttempt(1, "skill", orchestration.RunResult{
		ScenarioID: "scenario-alpha", Condition: "skill",
		Agent: &common.Result{
			Messages: []inference.Message{
				{Role: "user", Content: "Restore the deployment."},
				{Role: "assistant", Content: strings.Repeat("diagnostic detail line\n", 20)},
			},
			Turns:       2,
			Termination: common.TerminationCompleted,
		},
		Grading: orchestration.GradingResult{
			Criteria:    []orchestration.CriterionResult{{ID: "ready", Weight: 1, Passed: true, Stdout: "ready"}},
			Score:       1,
			FullSuccess: true,
		},
	}))
	require.NoError(t, store.Finalize(startedAt.Add(time.Minute)))
}

func newTestModel(t *testing.T) *Model {
	t.Helper()
	root := t.TempDir()
	base := time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC)
	createRun(t, root, "run-older", base)
	createRun(t, root, "run-newer", base.Add(time.Hour))
	model := New(Config{ResultsRoot: root})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	require.NoError(t, model.err)
	return model
}

func TestModelRendersRunList(t *testing.T) {
	model := newTestModel(t)
	view := model.View()
	assert.Contains(t, view.Content, "benchmark results")
	assert.Contains(t, view.Content, "run-newer")
	assert.True(t, view.AltScreen)
}

func TestModelNavigatesRunsToConversation(t *testing.T) {
	model := newTestModel(t)

	model.Update(press("enter"))
	require.Equal(t, screens.Run, model.current().Kind)
	assert.Equal(t, "run-newer", model.current().RunID)

	model.Update(press("enter"))
	require.Equal(t, screens.Attempts, model.current().Kind)
	assert.Equal(t, "scenario-alpha", model.current().ScenarioID)

	model.Update(press("enter"))
	require.Equal(t, screens.Attempt, model.current().Kind)
	assert.Contains(t, model.View().Content, "Grading criteria")
	assert.Contains(t, model.View().Content, "Conversation")

	model.Update(press("esc"))
	require.Equal(t, screens.Attempts, model.current().Kind)
}

func TestModelSwitchesViewModes(t *testing.T) {
	model := newTestModel(t)

	model.Update(press("2"))
	assert.Equal(t, screens.ModeAgents, model.mode)
	assert.Contains(t, model.View().Content, "skill")
	model.Update(press("enter"))
	require.Equal(t, screens.AgentRuns, model.current().Kind)
	model.Update(press("enter"))
	require.Equal(t, screens.Run, model.current().Kind)
	assert.Equal(t, "skill", model.current().Agent)

	model.Update(press("3"))
	assert.Equal(t, screens.ModeTasks, model.mode)
	require.Equal(t, screens.List, model.current().Kind)
	assert.Contains(t, model.View().Content, "scenario-alpha")
	model.Update(press("enter"))
	require.Equal(t, screens.TaskRuns, model.current().Kind)

	model.Update(press("1"))
	assert.Equal(t, screens.ModeRuns, model.mode)
}

func TestModelShowsTotalsPage(t *testing.T) {
	model := newTestModel(t)

	model.Update(press("4"))
	assert.Equal(t, screens.ModeTotals, model.mode)
	require.Equal(t, screens.Totals, model.current().Kind)
	content := model.View().Content
	assert.Contains(t, content, "All runs")
	assert.Contains(t, content, "Tokens")
}

func TestModelTogglesHelp(t *testing.T) {
	model := newTestModel(t)
	model.Update(press("?"))
	assert.Contains(t, model.View().Content, "Keys")
	model.Update(press("esc"))
	assert.False(t, model.showHelp)
}

func TestMouseClickSwitchesModeTabs(t *testing.T) {
	model := newTestModel(t)
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 12, Y: 1})
	assert.Equal(t, screens.ModeAgents, model.mode)
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 1, Y: 1})
	assert.Equal(t, screens.ModeRuns, model.mode)
}

func TestMouseWheelScrollsPreviewPane(t *testing.T) {
	model := newTestModel(t)
	model.Update(tea.WindowSizeMsg{Width: 150, Height: 8})
	model.Update(press("2"))

	current := model.current()
	require.Equal(t, screens.List, current.Kind)
	model.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown, X: 145, Y: 10})
	assert.Greater(t, current.PreviewOffset, 0)
}

func TestChatSelectsCollapsesAndExpandsMessages(t *testing.T) {
	model := newTestModel(t)
	model.Update(press("enter"))
	model.Update(press("enter"))
	model.Update(press("enter"))
	model.Update(press("c"))
	require.Equal(t, screens.Attempt, model.current().Kind)
	require.Equal(t, screens.FocusSecondary, model.current().Focus)
	assert.Contains(t, model.View().Content, "Conversation")

	current := model.current()
	require.Equal(t, 1, current.Cursor)

	model.Update(press("x"))
	assert.True(t, current.Expanded[current.Cursor])

	model.Update(press("up"))
	assert.Equal(t, 0, current.Cursor)
}

func TestMouseClickOpensRun(t *testing.T) {
	model := newTestModel(t)
	current := model.current()
	y := headerHeight + model.view.HeadLines(current, screens.PaneWidth(model.width)) + 1
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 2, Y: y})
	require.Equal(t, screens.Run, model.current().Kind)
	assert.Equal(t, "run-newer", model.current().RunID)
}

func TestMouseWheelScrollsChat(t *testing.T) {
	model := newTestModel(t)
	model.Update(press("enter"))
	model.Update(press("enter"))
	model.Update(press("enter"))
	model.Update(press("c"))
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 20})
	model.Update(press("g"))

	current := model.current()
	require.Equal(t, 0, current.ChatOffset)
	model.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown, X: 5, Y: 10})
	assert.Greater(t, current.ChatOffset, 0)
	before := current.ChatOffset
	model.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp, X: 5, Y: 10})
	assert.Less(t, current.ChatOffset, before)
}

func TestMouseClickExpandsChatMessage(t *testing.T) {
	model := newTestModel(t)
	model.Update(press("enter"))
	model.Update(press("enter"))
	model.Update(press("enter"))
	model.Update(press("c"))
	model.Update(press("g"))

	current := model.current()
	layout, ok := model.view.ChatLayout(current, screens.PreviewWidth(model.width))
	require.True(t, ok)
	require.NotEmpty(t, layout.CardSpans)

	span := layout.CardSpans[0]
	y := headerHeight + 1 + span[0] - current.ChatOffset
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: model.width - 5, Y: y})

	assert.Equal(t, layout.CardOwners[0], current.Cursor)
	assert.True(t, current.Expanded[layout.CardOwners[0]])
}

func TestBackReturnsToPreviousScreen(t *testing.T) {
	model := newTestModel(t)
	model.Update(press("enter"))
	require.Equal(t, screens.Run, model.current().Kind)

	start, end := model.backRange()
	require.Greater(t, end, start)
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: (start + end) / 2, Y: 0})
	require.Equal(t, screens.List, model.current().Kind)

	model.Update(press("enter"))
	require.Equal(t, screens.Run, model.current().Kind)
	model.Update(press("backspace"))
	require.Equal(t, screens.List, model.current().Kind)
}

func TestPaneKeyFocusesDashboardAndKeysScrollIt(t *testing.T) {
	model := newTestModel(t)
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 12})
	model.Update(press("enter"))
	current := model.current()
	require.Equal(t, screens.Run, current.Kind)

	model.Update(press("v"))
	require.Equal(t, screens.FocusSecondary, current.Focus)

	model.Update(press("pgdown"))
	assert.Greater(t, current.PreviewOffset, 0)

	model.Update(press("g"))
	assert.Equal(t, 0, current.PreviewOffset)

	model.Update(press("v"))
	assert.Equal(t, screens.FocusPrimary, current.Focus)
}

func TestTabCyclesModesAtWideWidth(t *testing.T) {
	model := newTestModel(t)
	model.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	require.Equal(t, screens.ModeRuns, model.mode)

	model.Update(press("tab"))
	assert.Equal(t, screens.ModeAgents, model.mode)
	model.Update(press("tab"))
	assert.Equal(t, screens.ModeTasks, model.mode)
	model.Update(press("tab"))
	assert.Equal(t, screens.ModeTotals, model.mode)
	require.Equal(t, screens.Totals, model.current().Kind)
	model.Update(press("tab"))
	assert.Equal(t, screens.ModeRuns, model.mode)
}

func TestRunsPageShowsSelectedRunDashboard(t *testing.T) {
	model := newTestModel(t)
	model.Update(tea.WindowSizeMsg{Width: 140, Height: 48})
	content := model.View().Content
	assert.Contains(t, content, "Outcome mix")
	assert.NotContains(t, content, "no attempts")
}

func TestHelpClosesOnClick(t *testing.T) {
	model := newTestModel(t)
	model.Update(press("?"))
	require.True(t, model.showHelp)
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 5, Y: 5})
	assert.False(t, model.showHelp)
}

func writeTestScenario(t *testing.T, root, id, difficulty string) {
	t.Helper()
	dir := filepath.Join(root, id)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	body := "id: " + id + "\ntitle: " + id + "\ntags:\n  difficulty: " + difficulty +
		"\n  area: pods\ntask: Restore the workload.\nprepare:\n  - program: prepare\n" +
		"verify_clean:\n  - program: verify\ngrading:\n  - id: ready\n    weight: 1\n    check:\n      program: check\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scenario.yaml"), []byte(body), 0o600))
}

func newFilterModel(t *testing.T) *Model {
	t.Helper()
	root := t.TempDir()
	createRun(t, root, "run-older", time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC))
	scenarios := t.TempDir()
	writeTestScenario(t, scenarios, "easy-task", "easy")
	writeTestScenario(t, scenarios, "hard-task", "hard")
	model := New(Config{ResultsRoot: root, ScenariosRoot: scenarios})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	require.NoError(t, model.err)
	return model
}

func TestTagFilterOverlayTogglesWithSpaceAndApplies(t *testing.T) {
	model := newFilterModel(t)
	model.Update(press("/"))
	require.True(t, model.filterOpen)
	require.NotEmpty(t, model.filterOptions)

	first := model.filterOptions[0]
	model.Update(press("space"))
	assert.True(t, model.filterOptions[0].selected)

	model.Update(press("esc"))
	assert.False(t, model.filterOpen)
	assert.Equal(t, []string{first.value}, model.view.Store().TagFilter()[first.key])
}

func TestTagFilterOverlayMouseToggleResetAndClose(t *testing.T) {
	model := newFilterModel(t)
	model.Update(press("/"))

	row := 2 + 1 // heading, spacer, first key section row, then first option
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 4, Y: row})
	assert.True(t, model.filterOptions[0].selected)

	resetStart, _, _, _ := model.filterButtons()
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: resetStart + 1, Y: 0})
	assert.False(t, model.filterOptions[0].selected)

	_, _, closeStart, _ := model.filterButtons()
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: closeStart + 1, Y: 0})
	assert.False(t, model.filterOpen)
}

func TestHeaderFilterButtonOpensOverlay(t *testing.T) {
	model := newFilterModel(t)
	_, start, end := model.headerRight()
	require.Greater(t, end, start)
	model.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: start + 1, Y: 0})
	assert.True(t, model.filterOpen)
}

func TestHeaderShowsActiveFilterChip(t *testing.T) {
	model := newFilterModel(t)
	model.view.Store().SetTagFilter(scenario.TagFilter{"difficulty": {"hard"}})
	assert.Contains(t, model.View().Content, "difficulty=hard")
}

func TestTagFilterOverlayMouseWheelMovesCursor(t *testing.T) {
	model := newFilterModel(t)
	model.Update(press("/"))
	require.Equal(t, 0, model.filterCursor)
	model.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown, X: 2, Y: 5})
	assert.Greater(t, model.filterCursor, 0)
	model.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp, X: 2, Y: 5})
	assert.Equal(t, 0, model.filterCursor)
}
