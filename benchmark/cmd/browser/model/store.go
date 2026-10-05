package model

import (
	"io"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/analysis"
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

// attemptEntry caches a parsed attempt file and its flattened record. Attempt
// files are written once, so entries survive reloads and are re-read only when
// the file's modification time changes, e.g. after a rerun replaced it.
type attemptEntry struct {
	attempt   results.Attempt
	err       error
	record    analysis.RunAttempt
	recordErr error
	modTime   time.Time
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
	s.recompute()
	return nil
}

// recompute rebuilds the model- and tag-filtered views from the full run list.
func (s *Store) recompute() {
	byModel := s.runs
	if len(s.models) > 0 {
		byModel = make([]results.RunRef, 0, len(s.runs))
		for _, run := range s.runs {
			if slices.Contains(s.models, run.Model) {
				byModel = append(byModel, run)
			}
		}
	}
	s.visible = s.filterRuns(byModel)
	s.agents = results.RollupAgents(byModel, s.matchesTags)
	s.tasks = results.RollupTasks(byModel, s.matchesTags)
}

// Runs returns every discovered run that matches the active tag filter, newest first.
func (s *Store) Runs() []results.RunRef { return s.visible }

func (s *Store) Agents() []results.AgentRollup { return s.agents }

func (s *Store) Tasks() []results.TaskRollup { return s.tasks }

// ScenarioTags returns a scenario's tags, or nil when it is unknown.
func (s *Store) ScenarioTags(scenarioID string) map[string]string {
	if definition, ok := s.catalog[scenarioID]; ok {
		return definition.Tags
	}
	return nil
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
	entry := s.attempt(runID, ref)
	return entry.attempt, entry.err
}

// RunAttempt returns one attempt flattened into the record analyze uses,
// without scenario sources.
func (s *Store) RunAttempt(runID string, ref results.AttemptRef) (analysis.RunAttempt, error) {
	entry := s.attempt(runID, ref)
	if entry.err != nil {
		return analysis.RunAttempt{}, entry.err
	}
	return entry.record, entry.recordErr
}

func (s *Store) attempt(runID string, ref results.AttemptRef) attemptEntry {
	var modTime time.Time
	if info, err := os.Stat(ref.Path); err == nil {
		modTime = info.ModTime()
	}
	if entry, ok := s.attempts[ref.Path]; ok && entry.modTime.Equal(modTime) {
		return entry
	}
	var runType results.RunType
	if snapshot, ok := s.snapshot[runID]; ok {
		runType = snapshot.Metadata.RunType
	}
	entry := attemptEntry{modTime: modTime}
	entry.attempt, entry.err = results.LoadAttempt(ref, runType)
	if entry.err == nil {
		entry.record, entry.recordErr = analysis.NewRunAttempt(runID, entry.attempt, nil)
	}
	s.attempts[ref.Path] = entry
	return entry
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
	return slices.Sorted(maps.Keys(summary.ByCondition))
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
