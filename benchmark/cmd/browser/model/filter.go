package model

import (
	"maps"
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

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
		if run.Model != "" && !slices.Contains(models, run.Model) {
			models = append(models, run.Model)
		}
	}
	slices.Sort(models)
	return models
}

// SetTagFilter sets the active filter and recomputes the filtered views.
func (s *Store) SetTagFilter(filter scenario.TagFilter) {
	s.selector = filter
	s.recompute()
}

func (s *Store) TagFilter() scenario.TagFilter { return s.selector }

// TagOptions returns every catalogue tag key with its sorted, unique values.
func (s *Store) TagOptions() map[string][]string {
	values := make(map[string]map[string]bool)
	for _, definition := range s.catalog {
		for key, value := range definition.Tags {
			if values[key] == nil {
				values[key] = make(map[string]bool)
			}
			values[key][value] = true
		}
	}
	options := make(map[string][]string, len(values))
	for key, set := range values {
		options[key] = slices.Sorted(maps.Keys(set))
	}
	return options
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
