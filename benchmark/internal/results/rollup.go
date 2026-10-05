package results

import (
	"slices"
	"strings"
)

type Rollup struct {
	Attempts         int
	FullSuccessCount int
	FullSuccessRate  float64
	MeanScore        float64
}

// AgentRollup aggregates one condition across all runs.
type AgentRollup struct {
	Agent string
	Runs  int
	Rollup
	Scenarios map[string]Rollup
}

// TaskRollup aggregates one scenario across all runs and conditions.
type TaskRollup struct {
	ScenarioID string
	Runs       int
	Rollup
	Agents map[string]Rollup
}

type rollupAccumulator struct {
	attempts    int
	fullSuccess int
	scoreTotal  float64
	children    map[string]*rollupAccumulator
}

func (a *rollupAccumulator) add(fullSuccess, attempts int, scoreTotal float64) {
	a.attempts += attempts
	a.fullSuccess += fullSuccess
	a.scoreTotal += scoreTotal
}

func (a *rollupAccumulator) child(key string) *rollupAccumulator {
	if a.children == nil {
		a.children = make(map[string]*rollupAccumulator)
	}
	if child := a.children[key]; child != nil {
		return child
	}
	child := &rollupAccumulator{}
	a.children[key] = child
	return child
}

func (a *rollupAccumulator) childRollups() map[string]Rollup {
	rollups := make(map[string]Rollup, len(a.children))
	for key, child := range a.children {
		rollups[key] = child.rollup()
	}
	return rollups
}

func (a *rollupAccumulator) rollup() Rollup {
	result := Rollup{Attempts: a.attempts, FullSuccessCount: a.fullSuccess}
	if a.attempts > 0 {
		result.FullSuccessRate = float64(a.fullSuccess) / float64(a.attempts)
		result.MeanScore = a.scoreTotal / float64(a.attempts)
	}
	return result
}

// RollupAgents aggregates each condition across runs, counting only scenarios
// for which include returns true. A nil include keeps all.
func RollupAgents(runs []RunRef, include func(string) bool) []AgentRollup {
	accumulators, runCounts := rollupRuns(runs, include, func(agent, scenarioID string) (string, string) {
		return agent, scenarioID
	})
	rollups := make([]AgentRollup, 0, len(accumulators))
	for agent, accumulator := range accumulators {
		rollups = append(rollups, AgentRollup{Agent: agent, Runs: runCounts[agent], Rollup: accumulator.rollup(), Scenarios: accumulator.childRollups()})
	}
	slices.SortFunc(rollups, func(a, b AgentRollup) int { return strings.Compare(a.Agent, b.Agent) })
	return rollups
}

// RollupTasks aggregates each scenario across runs, keeping only scenarios for
// which include returns true. A nil include keeps all.
func RollupTasks(runs []RunRef, include func(string) bool) []TaskRollup {
	accumulators, runCounts := rollupRuns(runs, include, func(agent, scenarioID string) (string, string) {
		return scenarioID, agent
	})
	rollups := make([]TaskRollup, 0, len(accumulators))
	for scenarioID, accumulator := range accumulators {
		rollups = append(rollups, TaskRollup{ScenarioID: scenarioID, Runs: runCounts[scenarioID], Rollup: accumulator.rollup(), Agents: accumulator.childRollups()})
	}
	slices.SortFunc(rollups, func(a, b TaskRollup) int { return strings.Compare(a.ScenarioID, b.ScenarioID) })
	return rollups
}

// rollupRuns sums the included scenario summaries of every run by the group
// and child that key returns for a condition and scenario. It also counts the
// runs contributing to each group.
func rollupRuns(runs []RunRef, include func(string) bool, key func(agent, scenarioID string) (string, string)) (map[string]*rollupAccumulator, map[string]int) {
	accumulators := make(map[string]*rollupAccumulator)
	runCounts := make(map[string]int)
	for _, run := range runs {
		if run.Summary == nil {
			continue
		}
		counted := make(map[string]bool)
		for agent, condition := range run.Summary.ByCondition {
			for scenarioID, scenario := range condition.Scenarios {
				if scenario.AttemptCount == 0 || (include != nil && !include(scenarioID)) {
					continue
				}
				group, child := key(agent, scenarioID)
				accumulator := accumulators[group]
				if accumulator == nil {
					accumulator = &rollupAccumulator{}
					accumulators[group] = accumulator
				}
				if !counted[group] {
					counted[group] = true
					runCounts[group]++
				}
				scoreTotal := scenario.MeanScore * float64(scenario.AttemptCount)
				accumulator.add(scenario.FullSuccessCount, scenario.AttemptCount, scoreTotal)
				accumulator.child(child).add(scenario.FullSuccessCount, scenario.AttemptCount, scoreTotal)
			}
		}
	}
	return accumulators, runCounts
}
