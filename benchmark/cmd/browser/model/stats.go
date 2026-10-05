package model

import "github.com/Nasus20202/MastersThesis/benchmark/internal/results"

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
	return Ratio(metrics.Full, metrics.Attempts)
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

type number interface{ ~int | ~float64 }

func Sum[T number](values []T) T {
	var total T
	for _, value := range values {
		total += value
	}
	return total
}

// Mean is the average of values, or zero without values.
func Mean[T number](values []T) float64 {
	if len(values) == 0 {
		return 0
	}
	return float64(Sum(values)) / float64(len(values))
}
