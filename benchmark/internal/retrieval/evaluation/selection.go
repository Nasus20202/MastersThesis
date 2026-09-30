package evaluation

import (
	"cmp"
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

// Select applies the frozen selection rule: the highest rewrite hit@k, where
// differences of at most one probe are ties, broken by higher rewrite MRR,
// then smaller k, then the simpler mode and chunking.
func Select(configs []ConfigScore, probes int) ConfigScore {
	best := slices.MaxFunc(configs, func(a, b ConfigScore) int { return cmp.Compare(a.HitRewrite, b.HitRewrite) }).HitRewrite
	tolerance := 1/float64(probes) + 1e-9
	var tied []ConfigScore
	for _, config := range configs {
		if config.HitRewrite >= best-tolerance {
			tied = append(tied, config)
		}
	}
	slices.SortStableFunc(tied, func(a, b ConfigScore) int {
		return cmp.Or(
			cmp.Compare(b.MRRRewrite, a.MRRRewrite),
			cmp.Compare(a.K, b.K),
			cmp.Compare(slices.Index(retrieval.Modes, a.Mode), slices.Index(retrieval.Modes, b.Mode)),
			cmp.Compare(slices.Index(retrieval.Chunkings, a.Chunking), slices.Index(retrieval.Chunkings, b.Chunking)),
		)
	})
	return tied[0]
}
