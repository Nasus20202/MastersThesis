package model

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

// TokensPerSecond is decode throughput from summed tokens and seconds, or zero
// when no generation timings were recorded.
func TokensPerSecond(tokens int, seconds float64) float64 {
	if seconds <= 0 {
		return 0
	}
	return float64(tokens) / seconds
}

// AcceptanceRate is the accepted fraction of drafted tokens. It is false when
// no drafts were recorded, as with speculative decoding disabled.
func AcceptanceRate(accepted, drafted int) (float64, bool) {
	if drafted <= 0 {
		return 0, false
	}
	return float64(accepted) / float64(drafted), true
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
