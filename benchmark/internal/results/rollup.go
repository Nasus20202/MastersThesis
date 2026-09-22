package results

import (
	"slices"
	"strings"
)

// Rollup is an aggregate outcome over one or more attempts.
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

func (a *rollupAccumulator) rollup() Rollup {
	result := Rollup{Attempts: a.attempts, FullSuccessCount: a.fullSuccess}
	if a.attempts > 0 {
		result.FullSuccessRate = float64(a.fullSuccess) / float64(a.attempts)
		result.MeanScore = a.scoreTotal / float64(a.attempts)
	}
	return result
}

// RollupAgents aggregates every condition across the given runs.
func RollupAgents(runs []RunRef) []AgentRollup {
	accumulators := make(map[string]*rollupAccumulator)
	runCounts := make(map[string]int)
	for _, run := range runs {
		if run.Summary == nil {
			continue
		}
		for agent, condition := range run.Summary.ByCondition {
			accumulator := accumulators[agent]
			if accumulator == nil {
				accumulator = &rollupAccumulator{}
				accumulators[agent] = accumulator
			}
			runCounts[agent]++
			accumulator.add(condition.FullSuccessCount, condition.AttemptCount, condition.MeanScore*float64(condition.AttemptCount))
			for scenarioID, scenario := range condition.Scenarios {
				accumulator.child(scenarioID).add(scenario.FullSuccessCount, scenario.AttemptCount, scenario.MeanScore*float64(scenario.AttemptCount))
			}
		}
	}
	rollups := make([]AgentRollup, 0, len(accumulators))
	for agent, accumulator := range accumulators {
		scenarios := make(map[string]Rollup, len(accumulator.children))
		for scenarioID, child := range accumulator.children {
			scenarios[scenarioID] = child.rollup()
		}
		rollups = append(rollups, AgentRollup{Agent: agent, Runs: runCounts[agent], Rollup: accumulator.rollup(), Scenarios: scenarios})
	}
	slices.SortFunc(rollups, func(a, b AgentRollup) int { return strings.Compare(a.Agent, b.Agent) })
	return rollups
}

// RollupTasks aggregates every scenario across the given runs.
func RollupTasks(runs []RunRef) []TaskRollup {
	accumulators := make(map[string]*rollupAccumulator)
	runCounts := make(map[string]int)
	for _, run := range runs {
		if run.Summary == nil {
			continue
		}
		seen := make(map[string]struct{})
		for agent, condition := range run.Summary.ByCondition {
			for scenarioID, scenario := range condition.Scenarios {
				if scenario.AttemptCount == 0 {
					continue
				}
				accumulator := accumulators[scenarioID]
				if accumulator == nil {
					accumulator = &rollupAccumulator{}
					accumulators[scenarioID] = accumulator
				}
				if _, ok := seen[scenarioID]; !ok {
					seen[scenarioID] = struct{}{}
					runCounts[scenarioID]++
				}
				scoreTotal := scenario.MeanScore * float64(scenario.AttemptCount)
				accumulator.add(scenario.FullSuccessCount, scenario.AttemptCount, scoreTotal)
				accumulator.child(agent).add(scenario.FullSuccessCount, scenario.AttemptCount, scoreTotal)
			}
		}
	}
	rollups := make([]TaskRollup, 0, len(accumulators))
	for scenarioID, accumulator := range accumulators {
		agents := make(map[string]Rollup, len(accumulator.children))
		for agent, child := range accumulator.children {
			agents[agent] = child.rollup()
		}
		rollups = append(rollups, TaskRollup{ScenarioID: scenarioID, Runs: runCounts[scenarioID], Rollup: accumulator.rollup(), Agents: agents})
	}
	slices.SortFunc(rollups, func(a, b TaskRollup) int { return strings.Compare(a.ScenarioID, b.ScenarioID) })
	return rollups
}
