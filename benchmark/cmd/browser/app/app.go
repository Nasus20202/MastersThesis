// Package app is the browser's root Bubble Tea model. It owns navigation, the
// screen stack and the responsive two-pane layout, and delegates rendering to
// the screens package.
package app

import (
	"charm.land/bubbletea/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/screens"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

// Config configures a browser model.
type Config struct {
	ResultsRoot   string
	ScenariosRoot string
}

// Model is the root Bubble Tea model for the results browser.
type Model struct {
	store *model.Store
	view  *screens.View

	width, height int
	mode          screens.Mode
	stack         []screens.Route

	showHelp bool
	err      error
}

// New loads the run history and scenario catalogue and returns the browser
// model. Missing result or scenario directories are reported as a model error
// instead of failing to start.
func New(config Config) *Model {
	store := model.NewStore(model.StoreConfig{ResultsRoot: config.ResultsRoot, ScenariosRoot: config.ScenariosRoot})
	model := &Model{
		store: store,
		view:  screens.NewView(store),
		stack: []screens.Route{{Kind: screens.List}},
	}
	if err := model.reload(); err != nil {
		model.err = err
	}
	model.clamp()
	return model
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case WatchMsg:
		if err := m.reload(); err != nil {
			m.err = err
		}
		m.clamp()
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case tea.MouseClickMsg:
		m.handleClick(msg)
		return m, nil
	case tea.MouseWheelMsg:
		m.handleWheel(msg)
		return m, nil
	}
	return m, nil
}

func (m *Model) reload() error {
	m.view.Invalidate()
	if err := m.store.Reload(); err != nil {
		return err
	}
	for _, route := range m.stack {
		if route.RunID == "" {
			continue
		}
		if err := m.store.EnsureSnapshot(route.RunID); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) current() *screens.Route {
	return &m.stack[len(m.stack)-1]
}

func (m *Model) setMode(next screens.Mode) {
	m.mode = next
	m.view.SetMode(next)
	m.stack = []screens.Route{{Kind: screens.List}}
	m.clamp()
}

func (m *Model) push(next screens.Route) {
	m.stack = append(m.stack, next)
	if next.RunID != "" {
		_ = m.store.EnsureSnapshot(next.RunID)
	}
}

func (m *Model) open() {
	current := m.current()
	switch current.Kind {
	case screens.List:
		switch m.mode {
		case screens.ModeRuns:
			if run, ok := m.selectedRun(current.Cursor); ok {
				m.push(screens.Route{Kind: screens.Run, RunID: run.RunID})
			}
		case screens.ModeAgents:
			agents := m.store.Agents()
			if current.Cursor < len(agents) {
				m.push(screens.Route{Kind: screens.AgentRuns, Agent: agents[current.Cursor].Agent})
			}
		case screens.ModeTasks:
			tasks := m.store.Tasks()
			if current.Cursor < len(tasks) {
				m.push(screens.Route{Kind: screens.TaskRuns, Task: tasks[current.Cursor].ScenarioID})
			}
		}
	case screens.AgentRuns:
		runs := m.store.RunsForAgent(current.Agent)
		if current.Cursor < len(runs) {
			m.push(screens.Route{Kind: screens.Run, RunID: runs[current.Cursor].RunID, Agent: current.Agent})
		}
	case screens.TaskRuns:
		runs := m.store.RunsForTask(current.Task)
		if current.Cursor < len(runs) {
			m.push(screens.Route{Kind: screens.Run, RunID: runs[current.Cursor].RunID, Task: current.Task})
		}
	case screens.Run:
		if _, ok := m.store.Snapshot(current.RunID); !ok {
			return
		}
		rows := m.store.ScenarioRows(current.RunID, current.Agent, current.Task)
		if current.Cursor < len(rows) {
			m.push(screens.Route{Kind: screens.Attempts, RunID: current.RunID, Agent: current.Agent, Task: current.Task, ScenarioID: rows[current.Cursor].ID, Group: current.Agent})
		}
	case screens.Attempts:
		refs := m.store.AttemptsFor(current.RunID, current.ScenarioID, current.Group)
		if current.Cursor < len(refs) {
			ref := refs[current.Cursor]
			m.push(screens.Route{Kind: screens.Attempt, RunID: current.RunID, ScenarioID: current.ScenarioID, Group: ref.Group, Attempt: ref.Attempt, Path: ref.Path, Focus: screens.FocusPrimary})
			m.chatSelectLast()
		}
	}
}

func (m *Model) selectedRun(cursor int) (results.RunRef, bool) {
	runs := m.store.Runs()
	if cursor < 0 || cursor >= len(runs) {
		return results.RunRef{}, false
	}
	return runs[cursor], true
}

func (m *Model) clamp() {
	for index := range m.stack {
		route := &m.stack[index]
		if !screens.IsList(route.Kind) {
			continue
		}
		count := m.view.ListCount(route)
		if count == 0 {
			route.Cursor, route.Offset = 0, 0
			continue
		}
		route.Cursor = min(route.Cursor, count-1)
		route.Offset = min(route.Offset, route.Cursor)
	}
}
