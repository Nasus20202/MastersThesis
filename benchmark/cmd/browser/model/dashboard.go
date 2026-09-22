// The model package aggregates attempt-level evidence into the numbers the
// dashboards render: outcome mix, speed and turn distributions, termination
// reasons and criteria pass rates.
package model

import "github.com/Nasus20202/MastersThesis/benchmark/internal/results"

// Metrics is the aggregated evidence for one dashboard scope.
type Metrics struct {
	Attempts  int
	Full      int
	Partial   int
	Failed    int
	MeanScore float64

	Durations   []float64
	Turns       []int
	Tokens      []int
	Prompt      []int
	Completion  []int
	Cached      []int
	CacheRatios []float64

	Terminations map[string]int
	Criteria     map[string]CriterionStat
}

// CriterionStat is a pass count for one grading criterion.
type CriterionStat struct {
	Passed int
	Total  int
}

type ref struct {
	runID string
	ref   results.AttemptRef
}

func newMetrics() Metrics {
	return Metrics{Terminations: make(map[string]int), Criteria: make(map[string]CriterionStat)}
}

func gather(store *Store, refs []ref) Metrics {
	metrics := newMetrics()
	for _, item := range refs {
		attempt, err := store.Attempt(item.runID, item.ref)
		if err != nil {
			continue
		}
		grading := attempt.Grading()
		metrics.Attempts++
		switch {
		case grading.FullSuccess:
			metrics.Full++
		case grading.Score > 0:
			metrics.Partial++
		default:
			metrics.Failed++
		}
		metrics.MeanScore += grading.Score
		for _, criterion := range grading.Criteria {
			stat := metrics.Criteria[criterion.ID]
			stat.Total++
			if criterion.Passed {
				stat.Passed++
			}
			metrics.Criteria[criterion.ID] = stat
		}
		metrics.Durations = append(metrics.Durations, AttemptDuration(attempt))
		if attempt.Benchmark != nil && attempt.Benchmark.Agent != nil {
			agent := attempt.Benchmark.Agent
			metrics.Turns = append(metrics.Turns, agent.Turns)
			metrics.Tokens = append(metrics.Tokens, agent.TokenUsage.TotalTokens)
			metrics.Prompt = append(metrics.Prompt, agent.TokenUsage.PromptTokens)
			metrics.Completion = append(metrics.Completion, agent.TokenUsage.CompletionTokens)
			metrics.Cached = append(metrics.Cached, agent.TokenUsage.CachedTokens)
			metrics.CacheRatios = append(metrics.CacheRatios, CacheRatio(agent.TokenUsage.PromptTokens, agent.TokenUsage.CachedTokens))
			metrics.Terminations[agent.Termination]++
		}
		if attempt.Error() != "" || attempt.Failure() != nil {
			metrics.Terminations["error"]++
		}
	}
	if metrics.Attempts > 0 {
		metrics.MeanScore /= float64(metrics.Attempts)
	}
	return metrics
}

// RunMetrics gathers the attempts of one run, optionally filtered by agent and task.
func RunMetrics(store *Store, runID, filterAgent, filterTask string) Metrics {
	if err := store.EnsureSnapshot(runID); err != nil {
		return newMetrics()
	}
	snapshot, ok := store.Snapshot(runID)
	if !ok {
		return newMetrics()
	}
	var refs []ref
	for _, item := range snapshot.Attempts {
		if filterAgent != "" && item.Group != filterAgent {
			continue
		}
		if filterTask != "" && item.ScenarioID != filterTask {
			continue
		}
		refs = append(refs, ref{runID: runID, ref: item})
	}
	return gather(store, refs)
}

// AgentMetrics gathers every attempt of one condition across all runs.
func AgentMetrics(store *Store, agent string) Metrics {
	var refs []ref
	for _, run := range store.RunsForAgent(agent) {
		if err := store.EnsureSnapshot(run.RunID); err != nil {
			continue
		}
		snapshot, ok := store.Snapshot(run.RunID)
		if !ok {
			continue
		}
		for _, item := range snapshot.Attempts {
			if item.Group == agent {
				refs = append(refs, ref{runID: run.RunID, ref: item})
			}
		}
	}
	return gather(store, refs)
}

// TaskMetrics gathers every attempt of one scenario across all runs.
func TaskMetrics(store *Store, task string) Metrics {
	var refs []ref
	for _, run := range store.RunsForTask(task) {
		if err := store.EnsureSnapshot(run.RunID); err != nil {
			continue
		}
		snapshot, ok := store.Snapshot(run.RunID)
		if !ok {
			continue
		}
		for _, item := range snapshot.Attempts {
			if item.ScenarioID == task {
				refs = append(refs, ref{runID: run.RunID, ref: item})
			}
		}
	}
	return gather(store, refs)
}

// AttemptDuration is the wall-clock duration of an attempt.
func AttemptDuration(attempt results.Attempt) float64 {
	if attempt.Benchmark != nil && attempt.Benchmark.Agent != nil {
		return attempt.Benchmark.Agent.DurationSeconds
	}
	var total float64
	for _, criterion := range attempt.Grading().Criteria {
		total += criterion.DurationSeconds
	}
	return total
}

// OutcomeRate is the full-success rate of a metrics set.
func OutcomeRate(metrics Metrics) float64 {
	if metrics.Attempts == 0 {
		return 0
	}
	return float64(metrics.Full) / float64(metrics.Attempts)
}

// Ratio is passed/total, or zero when there is no data.
func Ratio(passed, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(passed) / float64(total)
}

// CacheRatio is the cached fraction of a prompt, or zero with no prompt.
func CacheRatio(prompt, cached int) float64 {
	if prompt <= 0 {
		return 0
	}
	return float64(cached) / float64(prompt)
}

// Mean is the arithmetic mean of values.
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

// MeanInts is the arithmetic mean of integers.
func MeanInts(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0
	for _, value := range values {
		total += value
	}
	return float64(total) / float64(len(values))
}
