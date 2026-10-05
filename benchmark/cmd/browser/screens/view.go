// Package screens renders the browser's screens from the read model. Each
// screen is a method on View, which owns the render caches so that navigation
// only recomputes the pane whose inputs changed.
package screens

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

type View struct {
	store *model.Store
	mode  Mode

	previews    map[string][]string
	heads       map[string]string
	details     map[string][]string
	detailHeads map[string]string
	layouts     map[string]cachedLayout
	totals      map[int][]string
	markdown    *ui.Renderer
}

type cachedLayout struct {
	layout components.Layout
	ok     bool
}

func NewView(store *model.Store) *View {
	view := &View{store: store}
	view.Invalidate()
	return view
}

func (v *View) Store() *model.Store { return v.store }

func (v *View) SetMode(mode Mode) { v.mode = mode }

// Invalidate drops every cached rendering; call it when the data reloads.
func (v *View) Invalidate() {
	v.previews = make(map[string][]string)
	v.heads = make(map[string]string)
	v.details = make(map[string][]string)
	v.detailHeads = make(map[string]string)
	v.layouts = make(map[string]cachedLayout)
	v.totals = make(map[int][]string)
}

func (v *View) md(width int) *ui.Renderer {
	if v.markdown == nil || v.markdown.Width() != width {
		v.markdown = ui.NewRenderer(width)
	}
	return v.markdown
}

func (v *View) HeadLines(route *Route, width int) int {
	return ui.LineCount(v.screenHead(route, width))
}

// DetailsHead returns the cached header of the attempt details pane.
func (v *View) DetailsHead(route *Route, width int) string {
	key := fmt.Sprintf("%s|%d", route.Path, width)
	if head, ok := v.detailHeads[key]; ok {
		return head
	}
	head := v.buildDetailsHead(route, width)
	v.detailHeads[key] = head
	return head
}

// DetailsLines returns the cached body lines of the attempt details pane.
func (v *View) DetailsLines(route *Route, width int) []string {
	key := fmt.Sprintf("%s|%d", route.Path, width)
	if lines, ok := v.details[key]; ok {
		return lines
	}
	attempt, err := v.store.Attempt(route.RunID, route.Ref())
	if err != nil {
		return []string{err.Error()}
	}
	lines := v.attemptBodyLines(attempt, attempt.Grading(), max(1, width-1))
	v.details[key] = lines
	return lines
}

// ChatLayout renders and caches the conversation layout for a route.
func (v *View) ChatLayout(route *Route, width int) (components.Layout, bool) {
	key := fmt.Sprintf("%s|%d|%d|%s", route.Path, width, route.Cursor, expandedKey(route.Expanded))
	if entry, ok := v.layouts[key]; ok {
		return entry.layout, entry.ok
	}
	layout, ok := v.buildChatLayout(route, width)
	v.layouts[key] = cachedLayout{layout: layout, ok: ok}
	return layout, ok
}

func (v *View) buildChatLayout(route *Route, width int) (components.Layout, bool) {
	attempt, err := v.store.Attempt(route.RunID, route.Ref())
	if err != nil {
		return components.Layout{}, false
	}
	if attempt.Benchmark == nil || attempt.Benchmark.Agent == nil {
		return components.Layout{}, false
	}
	renderer := v.md(max(20, width-8))
	return components.Build(attempt.Benchmark.Agent.Messages, route.Cursor, route.Expanded, renderer, width), true
}

func expandedKey(expanded map[int]bool) string {
	if len(expanded) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, key := range slices.Sorted(maps.Keys(expanded)) {
		fmt.Fprintf(&builder, "%d=%t,", key, expanded[key])
	}
	return builder.String()
}

func previewKey(route *Route, mode Mode, width int) string {
	return fmt.Sprintf("%d|%d|%s|%s|%s|%s|%s|%d|%s|%d|%d",
		route.Kind, mode, route.RunID, route.Agent, route.Task,
		route.ScenarioID, route.Group, route.Attempt, route.Path, route.Cursor, width)
}

func headKey(route *Route, mode Mode, width int) string {
	return fmt.Sprintf("h|%d|%d|%s|%s|%s|%s|%d",
		route.Kind, mode, route.RunID, route.Agent, route.Task, route.ScenarioID, width)
}

func (v *View) previewContent(route *Route, width int) string {
	switch route.Kind {
	case Run:
		return v.runDashboard(route, width)
	case Attempts:
		return v.attemptPreview(route, width)
	case AgentRuns, TaskRuns:
		return v.runsPreview(route, width)
	case List:
		switch v.mode {
		case ModeAgents:
			return v.agentPreview(route, width)
		case ModeTasks:
			return v.taskPreview(route, width)
		default:
			return v.runsPreview(route, width)
		}
	default:
		return ""
	}
}

func (v *View) previewLines(route *Route, width int) []string {
	return strings.Split(v.previewContent(route, width), "\n")
}

func (v *View) screenHead(route *Route, width int) string {
	key := headKey(route, v.mode, width)
	if head, ok := v.heads[key]; ok {
		return head
	}
	head := v.buildScreenHead(route, width)
	v.heads[key] = head
	return head
}

func (v *View) buildScreenHead(route *Route, width int) string {
	switch route.Kind {
	case List:
		return components.Section(width, v.mode.Label())
	case AgentRuns:
		return components.Section(width, "Runs · agent "+route.Agent)
	case TaskRuns:
		return components.Section(width, "Runs · "+v.store.ScenarioTitle(route.Task))
	case Run:
		return v.runSummary(route, width)
	case Attempts:
		return v.attemptsHead(route, width)
	default:
		return ""
	}
}

func (v *View) screenTable(route *Route, width, height int) string {
	switch route.Kind {
	case List:
		switch v.mode {
		case ModeAgents:
			return v.agentsTable(route, width, height)
		case ModeTasks:
			return v.tasksTable(route, width, height)
		default:
			return v.runsTable(v.store.Runs(), route, width, height)
		}
	case AgentRuns:
		return v.runsTable(v.store.RunsForAgent(route.Agent), route, width, height)
	case TaskRuns:
		return v.runsTable(v.store.RunsForTask(route.Task), route, width, height)
	case Run:
		return v.scenariosTable(route, width, height)
	case Attempts:
		return v.attemptsTable(route, width, height)
	default:
		return ""
	}
}

func (v *View) ListCount(route *Route) int {
	switch route.Kind {
	case List:
		switch v.mode {
		case ModeAgents:
			return len(v.store.Agents())
		case ModeTasks:
			return len(v.store.Tasks())
		default:
			return len(v.store.Runs())
		}
	case AgentRuns:
		return len(v.store.RunsForAgent(route.Agent))
	case TaskRuns:
		return len(v.store.RunsForTask(route.Task))
	case Run:
		return len(v.store.ScenarioRows(route.RunID, route.Agent, route.Task))
	case Attempts:
		return len(v.store.AttemptsFor(route.RunID, route.ScenarioID, route.Group))
	default:
		return 0
	}
}
