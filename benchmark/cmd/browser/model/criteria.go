package model

import (
	"maps"
	"slices"
	"strings"
)

// CriteriaMatrix compares grading criteria across conditions for one scenario.
// Rows are criteria, columns are conditions (agents); each cell is that
// criterion's pass count for that condition. It is only meaningful where the
// same criterion is evaluated repeatedly, i.e. within one scenario.
type CriteriaMatrix struct {
	Agents []string
	Rows   []CriteriaRow
}

type CriterionStat struct {
	Passed int
	Total  int
}

// CriteriaRow is one criterion's pass counts, keyed by condition.
type CriteriaRow struct {
	ID    string
	Cells map[string]CriterionStat
}

// ScenarioCriteria builds the criterion × condition matrix for one scenario in
// one run.
func ScenarioCriteria(store *Store, runID, scenarioID string) CriteriaMatrix {
	if err := store.EnsureSnapshot(runID); err != nil {
		return CriteriaMatrix{}
	}
	var refs []ref
	for _, item := range store.AttemptsFor(runID, scenarioID, "") {
		refs = append(refs, ref{runID: runID, ref: item})
	}
	return criteriaFromRefs(store, refs)
}

// TaskCriteria builds the criterion × condition matrix for one scenario across
// every run that contains it.
func TaskCriteria(store *Store, scenarioID string) CriteriaMatrix {
	return criteriaFromRefs(store, taskRefs(store, scenarioID))
}

func criteriaFromRefs(store *Store, refs []ref) CriteriaMatrix {
	cells := make(map[string]map[string]CriterionStat)
	agents := make(map[string]bool)
	for _, item := range refs {
		attempt, err := store.Attempt(item.runID, item.ref)
		if err != nil {
			continue
		}
		agents[item.ref.Group] = true
		for _, criterion := range attempt.Grading().Criteria {
			if cells[criterion.ID] == nil {
				cells[criterion.ID] = make(map[string]CriterionStat)
			}
			stat := cells[criterion.ID][item.ref.Group]
			stat.Total++
			if criterion.Passed {
				stat.Passed++
			}
			cells[criterion.ID][item.ref.Group] = stat
		}
	}

	matrix := CriteriaMatrix{Agents: slices.Sorted(maps.Keys(agents))}
	for id, byAgent := range cells {
		matrix.Rows = append(matrix.Rows, CriteriaRow{ID: id, Cells: byAgent})
	}
	slices.SortFunc(matrix.Rows, func(a, b CriteriaRow) int { return strings.Compare(a.ID, b.ID) })
	return matrix
}
