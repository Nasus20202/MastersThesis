package model

import (
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

// AgentModelMatrix holds each condition's results on each model.
type AgentModelMatrix struct {
	Agents []string
	Models []string
	// Cells is keyed by agent, then model.
	Cells map[string]map[string]results.Rollup
}

// AgentModelMatrix rolls up every visible run's conditions by run model from
// the run summaries, counting only scenarios matching the tag filter. Runs of
// an unknown model are left out.
func (s *Store) AgentModelMatrix() AgentModelMatrix {
	type totals struct {
		attempts, full int
		score          float64
	}
	sums := map[string]map[string]*totals{}
	var agents, models []string
	for _, run := range s.visible {
		if run.Model == "" || run.Summary == nil {
			continue
		}
		for agent, condition := range run.Summary.ByCondition {
			for scenarioID, scenario := range condition.Scenarios {
				if scenario.AttemptCount == 0 || !s.matchesTags(scenarioID) {
					continue
				}
				if sums[agent] == nil {
					sums[agent] = map[string]*totals{}
					agents = append(agents, agent)
				}
				cell := sums[agent][run.Model]
				if cell == nil {
					cell = &totals{}
					sums[agent][run.Model] = cell
				}
				if !slices.Contains(models, run.Model) {
					models = append(models, run.Model)
				}
				cell.attempts += scenario.AttemptCount
				cell.full += scenario.FullSuccessCount
				cell.score += scenario.MeanScore * float64(scenario.AttemptCount)
			}
		}
	}
	slices.Sort(agents)
	slices.Sort(models)
	matrix := AgentModelMatrix{Agents: agents, Models: models, Cells: map[string]map[string]results.Rollup{}}
	for agent, byModel := range sums {
		matrix.Cells[agent] = map[string]results.Rollup{}
		for model, cell := range byModel {
			matrix.Cells[agent][model] = results.Rollup{
				Attempts:         cell.attempts,
				FullSuccessCount: cell.full,
				FullSuccessRate:  float64(cell.full) / float64(cell.attempts),
				MeanScore:        cell.score / float64(cell.attempts),
			}
		}
	}
	return matrix
}

// ModelRollups sums each model's conditions in the agent-by-model matrix into
// summary-level metrics, without loading attempt files.
func (s *Store) ModelRollups() map[string]Metrics {
	matrix := s.AgentModelMatrix()
	metrics := make(map[string]Metrics, len(matrix.Models))
	for _, byModel := range matrix.Cells {
		for model, cell := range byModel {
			item := metrics[model]
			item.MeanScore = (item.MeanScore*float64(item.Attempts) + cell.MeanScore*float64(cell.Attempts)) / float64(item.Attempts+cell.Attempts)
			item.Attempts += cell.Attempts
			item.Full += cell.FullSuccessCount
			metrics[model] = item
		}
	}
	return metrics
}
