package app

import (
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

func TestTabFocusesDashboardAndKeysScrollIt(t *testing.T) {
	model := newTestModel(t)
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 12})
	model.Update(press("enter"))
	current := model.current()
	require.Equal(t, screens.Run, current.Kind)

	model.Update(press("tab"))
	require.Equal(t, screens.FocusSecondary, current.Focus)

	model.Update(press("pgdown"))
	assert.Greater(t, current.PreviewOffset, 0)

	model.Update(press("g"))
	assert.Equal(t, 0, current.PreviewOffset)

	model.Update(press("tab"))
	assert.Equal(t, screens.FocusPrimary, current.Focus)
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
