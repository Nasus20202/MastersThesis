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

	// Decoding throughput summed over responses, so long responses weigh more
	// than short ones. Draft counters are zero when speculation is disabled.
	PredictedTokens  int
	PredictedSeconds float64
	DraftTokens      int
	DraftAccepted    int

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
		refs = append(refs, ref{runID: runID, ref: item})
	}
	return refs
}

func taskRefs(store *Store, task string) []ref {
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

// AttemptTokensPerSecond is one attempt's decode throughput from its generation
// timings, summed over responses like the dashboard metrics.
func AttemptTokensPerSecond(attempt results.Attempt) float64 {
	predicted, seconds, _, _ := attemptTimings(attempt)
	if seconds <= 0 {
		return 0
	}
	return float64(predicted) / seconds
}

// AttemptDraftAcceptanceRate is one attempt's accepted fraction of drafted
// tokens. It is false when the attempt recorded no drafts.
func AttemptDraftAcceptanceRate(attempt results.Attempt) (float64, bool) {
	_, _, draft, accepted := attemptTimings(attempt)
	if draft <= 0 {
		return 0, false
	}
	return float64(accepted) / float64(draft), true
}

// attemptTimings sums the generation timings recorded across an attempt's
// responses.
func attemptTimings(attempt results.Attempt) (predicted int, seconds float64, draft, accepted int) {
	if attempt.Benchmark == nil || attempt.Benchmark.Agent == nil {
		return 0, 0, 0, 0
	}
	for _, response := range attempt.Benchmark.Agent.Responses {
		timings := response.Response.Timings
		if timings == nil {
			continue
		}
		predicted += timings.PredictedN
		seconds += timings.PredictedMS / 1000
		draft += timings.DraftN
		accepted += timings.DraftNAccepted
	}
	return
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

// PredictedTokensPerSecond is decode throughput from summed tokens and seconds,
// or zero when no generation timings were recorded.
func PredictedTokensPerSecond(metrics Metrics) float64 {
	if metrics.PredictedSeconds <= 0 {
		return 0
	}
	return float64(metrics.PredictedTokens) / metrics.PredictedSeconds
}

// DraftAcceptanceRate is the accepted fraction of drafted tokens. It is false
// when no drafts were recorded, as with speculative decoding disabled.
func DraftAcceptanceRate(metrics Metrics) (float64, bool) {
	if metrics.DraftTokens <= 0 {
		return 0, false
	}
	return float64(metrics.DraftAccepted) / float64(metrics.DraftTokens), true
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

// Sum is the total of values.
func Sum(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total
}

// SumInts is the total of integers.
func SumInts(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
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
