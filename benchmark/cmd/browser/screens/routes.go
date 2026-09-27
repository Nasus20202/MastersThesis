package screens

import "github.com/Nasus20202/MastersThesis/benchmark/internal/results"

type Mode int

const (
	ModeRuns Mode = iota
	ModeAgents
	ModeTasks
	ModeTotals
	ModeCount
)

func (m Mode) Label() string {
	switch m {
	case ModeAgents:
		return "Agents"
	case ModeTasks:
		return "Tasks"
	case ModeTotals:
		return "Totals"
	default:
		return "Runs"
	}
}

const (
	FocusPrimary   = 0
	FocusSecondary = 1
)

type Kind int

const (
	List Kind = iota
	AgentRuns
	TaskRuns
	Run
	Attempts
	Attempt
	Totals
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

func (r *Route) Ref() results.AttemptRef {
	return results.AttemptRef{ScenarioID: r.ScenarioID, Group: r.Group, Attempt: r.Attempt, Path: r.Path}
}
