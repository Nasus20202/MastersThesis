package ui

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Open       key.Binding
	Back       key.Binding
	Chat       key.Binding
	Expand     key.Binding
	Help       key.Binding
	Quit       key.Binding
	ModeRuns   key.Binding
	ModeAgents key.Binding
	ModeTasks  key.Binding
	ModeTotals key.Binding
	NextMode   key.Binding
	Pane       key.Binding
	Filter     key.Binding
}

var DefaultKeyMap = KeyMap{
	Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	PageUp:     key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
	PageDown:   key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
	Open:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
	Back:       key.NewBinding(key.WithKeys("esc", "backspace", "h"), key.WithHelp("esc", "back")),
	Chat:       key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "conversation")),
	Expand:     key.NewBinding(key.WithKeys("x", "space", "enter"), key.WithHelp("x", "expand")),
	Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	ModeRuns:   key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "runs")),
	ModeAgents: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "agents")),
	ModeTasks:  key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "tasks")),
	ModeTotals: key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "totals")),
	NextMode:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab")),
	Pane:       key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "list / dashboard")),
	Filter:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "tag filter")),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Open, k.Back, k.Chat, k.Expand, k.NextMode, k.Pane, k.Filter, k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Open, k.Back, k.Chat, k.Expand},
		{k.ModeRuns, k.ModeAgents, k.ModeTasks, k.ModeTotals},
		{k.NextMode, k.Pane, k.Filter, k.Help, k.Quit},
	}
}
