package model

import (
	"io"
	"log/slog"
	"sort"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

type StoreConfig struct {
	ResultsRoot   string
	ScenariosRoot string
}

// Store is a cached view over the results tree. It is not safe for concurrent
// use; the TUI updates it from a single goroutine.
type Store struct {
	resultsRoot   string
	scenariosRoot string
	catalog       map[string]scenario.Definition
	selector      scenario.TagFilter
	models        []string

	runs     []results.RunRef
	visible  []results.RunRef
	agents   []results.AgentRollup
	tasks    []results.TaskRollup
	snapshot map[string]results.RunSnapshot
	attempts map[string]attemptEntry
}

type attemptEntry struct {
	attempt results.Attempt
	err     error
}

// New returns an empty store; call Reload to populate it.
func NewStore(config StoreConfig) *Store {
	return &Store{
		resultsRoot:   config.ResultsRoot,
		scenariosRoot: config.ScenariosRoot,
		catalog:       loadCatalog(config.ScenariosRoot),
		snapshot:      make(map[string]results.RunSnapshot),
		attempts:      make(map[string]attemptEntry),
	}
}

// Reload re-reads the run list and drops every cache.
func (s *Store) Reload() error {
	runs, err := results.ListRuns(s.resultsRoot)
	if err != nil {
		return err
	}
	s.runs = runs
	s.snapshot = make(map[string]results.RunSnapshot)
	s.attempts = make(map[string]attemptEntry)
	s.recompute()
	return nil
}

// recompute rebuilds the model- and tag-filtered views from the full run list.
func (s *Store) recompute() {
	byModel := s.runs
	if len(s.models) > 0 {
		byModel = make([]results.RunRef, 0, len(s.runs))
		for _, run := range s.runs {
			if containsString(s.models, run.Model) {
				byModel = append(byModel, run)
			}
		}
	}
	s.visible = s.filterRuns(byModel)
	s.agents = results.RollupAgents(byModel, s.matchesTags)
	s.tasks = results.RollupTasks(byModel, s.matchesTags)
}

// SetModelFilter keeps only runs of the given models; none keeps all.
func (s *Store) SetModelFilter(models []string) {
	s.models = models
	s.recompute()
}

func (s *Store) ModelFilter() []string { return s.models }

// ModelOptions returns the sorted models of all discovered runs.
func (s *Store) ModelOptions() []string {
	var models []string
	for _, run := range s.runs {
		if run.Model != "" && !containsString(models, run.Model) {
			models = append(models, run.Model)
		}
	}
	sort.Strings(models)
	return models
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// Runs returns every discovered run that matches the active tag filter, newest first.
func (s *Store) Runs() []results.RunRef { return s.visible }

func (s *Store) Agents() []results.AgentRollup { return s.agents }

func (s *Store) Tasks() []results.TaskRollup { return s.tasks }

// SetTagFilter sets the active filter and recomputes the filtered views.
func (s *Store) SetTagFilter(filter scenario.TagFilter) {
	s.selector = filter
	s.recompute()
}

func (s *Store) TagFilter() scenario.TagFilter { return s.selector }

// TagOptions returns every catalogue tag key with its sorted, unique values.
func (s *Store) TagOptions() map[string][]string {
	values := make(map[string]map[string]struct{})
	for _, definition := range s.catalog {
		for key, value := range definition.Tags {
			if values[key] == nil {
				values[key] = make(map[string]struct{})
			}
			values[key][value] = struct{}{}
		}
	}
	options := make(map[string][]string, len(values))
	for key, set := range values {
		keyValues := make([]string, 0, len(set))
		for value := range set {
			keyValues = append(keyValues, value)
		}
		sort.Strings(keyValues)
		options[key] = keyValues
	}
	return options
}

// ScenarioTags returns a scenario's tags, or nil when it is unknown.
func (s *Store) ScenarioTags(scenarioID string) map[string]string {
	if definition, ok := s.catalog[scenarioID]; ok {
		return definition.Tags
	}
	return nil
}

func (s *Store) matchesTags(scenarioID string) bool {
	return s.selector.Matches(s.ScenarioTags(scenarioID))
}

// filterRuns keeps the runs that have attempts for a matching scenario. A run
// without a summary cannot be evaluated and stays visible.
func (s *Store) filterRuns(runs []results.RunRef) []results.RunRef {
	if s.selector.Empty() {
		return runs
	}
	filtered := make([]results.RunRef, 0, len(runs))
	for _, run := range runs {
		if s.runMatchesTags(run) {
			filtered = append(filtered, run)
		}
	}
	return filtered
}

func (s *Store) runMatchesTags(run results.RunRef) bool {
	if run.Summary == nil {
		return true
	}
	for _, condition := range run.Summary.ByCondition {
		if s.conditionMatches(condition) {
			return true
		}
	}
	return false
}

// conditionMatches reports whether a condition has attempts from matching
// scenarios; a summary without per-scenario detail falls back to its total.
func (s *Store) conditionMatches(condition results.ConditionSummary) bool {
	if len(condition.Scenarios) == 0 {
		return condition.AttemptCount > 0
	}
	for scenarioID, scenario := range condition.Scenarios {
		if scenario.AttemptCount > 0 && s.matchesTags(scenarioID) {
			return true
		}
	}
	return false
}

func (s *Store) Catalog() map[string]scenario.Definition { return s.catalog }

func (s *Store) ScenarioTitle(scenarioID string) string {
	if definition, ok := s.catalog[scenarioID]; ok && definition.Title != "" {
		return definition.Title
	}
	return scenarioID
}

func (s *Store) Snapshot(runID string) (results.RunSnapshot, bool) {
	snapshot, ok := s.snapshot[runID]
	return snapshot, ok
}

// EnsureSnapshot loads a run snapshot if it is not cached yet.
func (s *Store) EnsureSnapshot(runID string) error {
	if _, ok := s.snapshot[runID]; ok {
		return nil
	}
	snapshot, err := results.LoadRun(s.resultsRoot, runID)
	if err != nil {
		return err
	}
	s.snapshot[runID] = snapshot
	return nil
}

// Attempt loads and caches one attempt artifact.
func (s *Store) Attempt(runID string, ref results.AttemptRef) (results.Attempt, error) {
	if entry, ok := s.attempts[ref.Path]; ok {
		return entry.attempt, entry.err
	}
	var runType results.RunType
	if snapshot, ok := s.snapshot[runID]; ok {
		runType = snapshot.Metadata.RunType
	}
	attempt, err := results.LoadAttempt(ref, runType)
	s.attempts[ref.Path] = attemptEntry{attempt: attempt, err: err}
	return attempt, err
}

func (s *Store) RunsForAgent(agent string) []results.RunRef {
	var runs []results.RunRef
	for _, run := range s.visible {
		if run.Summary == nil {
			continue
		}
		condition, ok := run.Summary.ByCondition[agent]
		if ok && s.conditionMatches(condition) {
			runs = append(runs, run)
		}
	}
	return runs
}

func (s *Store) RunsForTask(task string) []results.RunRef {
	if !s.matchesTags(task) {
		return nil
	}
	var runs []results.RunRef
	for _, run := range s.visible {
		if run.Summary == nil {
			continue
		}
		for _, condition := range run.Summary.ByCondition {
			if scenario, ok := condition.Scenarios[task]; ok && scenario.AttemptCount > 0 {
				runs = append(runs, run)
				break
			}
		}
	}
	return runs
}

// AttemptsFor returns the attempt references of a scenario within a run,
// optionally filtered to one condition (group).
func (s *Store) AttemptsFor(runID, scenarioID, group string) []results.AttemptRef {
	snapshot, ok := s.snapshot[runID]
	if !ok {
		return nil
	}
	var refs []results.AttemptRef
	for _, ref := range snapshot.Attempts {
		if ref.ScenarioID != scenarioID {
			continue
		}
		if group != "" && ref.Group != group {
			continue
		}
		refs = append(refs, ref)
	}
	return refs
}

// ScenarioRow aggregates one scenario across the conditions shown in a run.
type ScenarioRow struct {
	ID          string
	Agents      []string
	Attempts    int
	FullSuccess int
	ScoreTotal  float64
	MeanScore   float64
}

// ScenarioRows aggregates the scenarios of a run, optionally filtered by agent
// and scenario.
func (s *Store) ScenarioRows(runID, filterAgent, filterTask string) []ScenarioRow {
	snapshot, ok := s.snapshot[runID]
	if !ok || snapshot.Summary == nil {
		return nil
	}
	var rows []ScenarioRow
	for _, scenarioID := range snapshot.Metadata.Scenarios {
		if filterTask != "" && scenarioID != filterTask {
			continue
		}
		if !s.selector.Empty() && !s.matchesTags(scenarioID) {
			continue
		}
		row := ScenarioRow{ID: scenarioID}
		for _, agent := range ConditionNames(snapshot.Summary) {
			if filterAgent != "" && agent != filterAgent {
				continue
			}
			scenario, ok := snapshot.Summary.ByCondition[agent].Scenarios[scenarioID]
			if !ok || scenario.AttemptCount == 0 {
				continue
			}
			row.Agents = append(row.Agents, agent)
			row.Attempts += scenario.AttemptCount
			row.FullSuccess += scenario.FullSuccessCount
			row.ScoreTotal += scenario.MeanScore * float64(scenario.AttemptCount)
		}
		if row.Attempts > 0 {
			row.MeanScore = row.ScoreTotal / float64(row.Attempts)
		}
		rows = append(rows, row)
	}
	return rows
}

// ConditionNames returns the condition names of a summary, sorted.
func ConditionNames(summary *results.RunSummary) []string {
	if summary == nil {
		return nil
	}
	names := make([]string, 0, len(summary.ByCondition))
	for name := range summary.ByCondition {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func loadCatalog(root string) map[string]scenario.Definition {
	catalog := make(map[string]scenario.Definition)
	if strings.TrimSpace(root) == "" {
		return catalog
	}
	// Silence the loader's info logs so they never interleave with the TUI.
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer slog.SetDefault(previous)

	paths, err := scenario.Discover(root)
	if err != nil {
		return catalog
	}
	for _, path := range paths {
		definition, err := scenario.Load(path)
		if err != nil {
			continue
		}
		catalog[definition.ID] = definition
	}
	return catalog
}
