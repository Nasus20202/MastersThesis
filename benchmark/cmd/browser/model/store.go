// The store is the browser's read model: it loads and caches the persisted
// run history and the scenario catalogue, and answers the aggregate queries the
// screens and dashboards are built from.
package model

import (
	"io"
	"log/slog"
	"sort"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

// Config locates the read model's inputs.
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

	runs     []results.RunRef
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
	s.agents = results.RollupAgents(runs)
	s.tasks = results.RollupTasks(runs)
	s.snapshot = make(map[string]results.RunSnapshot)
	s.attempts = make(map[string]attemptEntry)
	return nil
}

// Runs returns every discovered run, newest first.
func (s *Store) Runs() []results.RunRef { return s.runs }

// Agents returns the cross-run condition rollups.
func (s *Store) Agents() []results.AgentRollup { return s.agents }

// Tasks returns the cross-run scenario rollups.
func (s *Store) Tasks() []results.TaskRollup { return s.tasks }

// Catalog returns the scenario definitions keyed by ID.
func (s *Store) Catalog() map[string]scenario.Definition { return s.catalog }

// ScenarioTitle resolves a scenario ID to its human title.
func (s *Store) ScenarioTitle(scenarioID string) string {
	if definition, ok := s.catalog[scenarioID]; ok && definition.Title != "" {
		return definition.Title
	}
	return scenarioID
}

// Snapshot returns a loaded run snapshot.
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
	runType := ""
	if snapshot, ok := s.snapshot[runID]; ok {
		runType = snapshot.Metadata.RunType
	}
	attempt, err := results.LoadAttempt(ref, runType)
	s.attempts[ref.Path] = attemptEntry{attempt: attempt, err: err}
	return attempt, err
}

// RunsForAgent returns the runs that contain attempts for an agent.
func (s *Store) RunsForAgent(agent string) []results.RunRef {
	var runs []results.RunRef
	for _, run := range s.runs {
		if run.Summary == nil {
			continue
		}
		if condition, ok := run.Summary.ByCondition[agent]; ok && condition.AttemptCount > 0 {
			runs = append(runs, run)
		}
	}
	return runs
}

// RunsForTask returns the runs that contain attempts for a scenario.
func (s *Store) RunsForTask(task string) []results.RunRef {
	var runs []results.RunRef
	for _, run := range s.runs {
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
