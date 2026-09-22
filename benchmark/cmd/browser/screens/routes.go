package screens

import "github.com/Nasus20202/MastersThesis/benchmark/internal/results"

// Mode is the top-level tab: runs, agents or tasks.
type Mode int

const (
	ModeRuns Mode = iota
	ModeAgents
	ModeTasks
	ModeCount
)

// Label is the tab label for a mode.
func (m Mode) Label() string {
	switch m {
	case ModeAgents:
		return "Agents"
	case ModeTasks:
		return "Tasks"
	default:
		return "Runs"
	}
}

// Focus selects which pane receives navigation keys.
const (
	FocusPrimary   = 0
	FocusSecondary = 1
)

// Kind identifies the screen a route points at.
type Kind int

const (
	List Kind = iota
	AgentRuns
	TaskRuns
	Run
	Attempts
	Attempt
)

// IsList reports whether a kind renders a table with an optional preview pane.
func IsList(kind Kind) bool {
	switch kind {
	case List, AgentRuns, TaskRuns, Run, Attempts:
		return true
	default:
		return false
	}
}

// Route is one entry in the navigation stack. Cursor, scroll offsets and
// expansion state are per-entry so going back restores the previous view.
type Route struct {
	Kind          Kind
	RunID         string
	Agent         string
	Task          string
	ScenarioID    string
	Group         string
	Attempt       int
	Path          string
	Cursor        int
	Offset        int
	ChatOffset    int
	PreviewOffset int
	Focus         int
	Expanded      map[int]bool
}

// Ref returns the attempt reference described by the route.
func (r *Route) Ref() results.AttemptRef {
	return results.AttemptRef{ScenarioID: r.ScenarioID, Group: r.Group, Attempt: r.Attempt, Path: r.Path}
}
