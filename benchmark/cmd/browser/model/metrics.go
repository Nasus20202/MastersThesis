// Package model is the browser's read model: a cached view of the results
// tree and the attempt-level numbers the dashboards render, such as outcome
// mix, speed and turn distributions, termination reasons and criteria pass
// rates.
package model

import "github.com/Nasus20202/MastersThesis/benchmark/internal/results"

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
	// PeakContext is each attempt's largest single-response context;
	// Overflows counts attempts that ran out of context.
	PeakContext []int
	Overflows   int

	// Decoding throughput summed over responses, so long responses weigh more
	// than short ones. Draft counters are zero when speculation is disabled.
	PredictedTokens  int
	PredictedSeconds float64
	DraftTokens      int
	DraftAccepted    int

	Terminations map[string]int
	Criteria     map[string]CriterionStat
}

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
			metrics.PeakContext = append(metrics.PeakContext, agent.TokenUsage.PeakContextTokens)
			if agent.ContextOverflow {
				metrics.Overflows++
			}
			metrics.Terminations[agent.Termination]++
			predicted, seconds, draft, accepted := attemptTimings(attempt)
			metrics.PredictedTokens += predicted
			metrics.PredictedSeconds += seconds
			metrics.DraftTokens += draft
			metrics.DraftAccepted += accepted
		}
		if attempt.ErrorMessage() != "" || attempt.Failure() != nil {
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
	return gather(store, runRefs(store, runID, filterAgent, filterTask))
}

// RunAgentMetrics gathers one run's attempts grouped by condition, honoring the
// same agent and task filters as RunMetrics.
func RunAgentMetrics(store *Store, runID, filterAgent, filterTask string) map[string]Metrics {
	return metricsByAgent(store, runRefs(store, runID, filterAgent, filterTask))
}

// AgentMetrics gathers every attempt of one condition across all runs.
func AgentMetrics(store *Store, agent string) Metrics {
	var refs []ref
	for _, run := range store.RunsForAgent(agent) {
		refs = append(refs, runRefs(store, run.RunID, agent, "")...)
	}
	return gather(store, refs)
}

// TaskMetrics gathers every attempt of one scenario across all runs.
func TaskMetrics(store *Store, task string) Metrics {
	return gather(store, taskRefs(store, task))
}

// TotalMetrics gathers every attempt across all runs, without any grouping.
func TotalMetrics(store *Store) Metrics {
	var refs []ref
	for _, run := range store.Runs() {
		refs = append(refs, runRefs(store, run.RunID, "", "")...)
	}
	return gather(store, refs)
}

// ModelMetrics gathers the attempts of the visible runs grouped by the run's
// model, optionally restricted to one condition and one scenario. Runs of an
// unknown model are left out.
func ModelMetrics(store *Store, filterAgent, filterTask string) map[string]Metrics {
	groups := make(map[string][]ref)
	for _, run := range store.Runs() {
		if run.Model == "" {
			continue
		}
		groups[run.Model] = append(groups[run.Model], runRefs(store, run.RunID, filterAgent, filterTask)...)
	}
	metrics := make(map[string]Metrics, len(groups))
	for name, groupRefs := range groups {
		if len(groupRefs) > 0 {
			metrics[name] = gather(store, groupRefs)
		}
	}
	return metrics
}

// TaskAgentMetrics gathers one scenario's attempts across all runs grouped by
// condition.
func TaskAgentMetrics(store *Store, task string) map[string]Metrics {
	return metricsByAgent(store, taskRefs(store, task))
}

func runRefs(store *Store, runID, filterAgent, filterTask string) []ref {
	if err := store.EnsureSnapshot(runID); err != nil {
		return nil
	}
	snapshot, ok := store.Snapshot(runID)
	if !ok {
		return nil
	}
	var refs []ref
	for _, item := range snapshot.Attempts {
		if filterAgent != "" && item.Group != filterAgent {
			continue
		}
		if filterTask != "" && item.ScenarioID != filterTask {
			continue
		}
		if !store.matchesTags(item.ScenarioID) {
			continue
		}
		refs = append(refs, ref{runID: runID, ref: item})
	}
	return refs
}

func taskRefs(store *Store, task string) []ref {
	if !store.matchesTags(task) {
		return nil
	}
	var refs []ref
	for _, run := range store.RunsForTask(task) {
		if err := store.EnsureSnapshot(run.RunID); err != nil {
			continue
		}
		for _, item := range store.AttemptsFor(run.RunID, task, "") {
			refs = append(refs, ref{runID: run.RunID, ref: item})
		}
	}
	return refs
}

func metricsByAgent(store *Store, refs []ref) map[string]Metrics {
	groups := make(map[string][]ref)
	for _, item := range refs {
		groups[item.ref.Group] = append(groups[item.ref.Group], item)
	}
	metrics := make(map[string]Metrics, len(groups))
	for agent, groupRefs := range groups {
		metrics[agent] = gather(store, groupRefs)
	}
	return metrics
}
