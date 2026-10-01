package evaluation

import (
	"context"
	"fmt"
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

// Ks are the compared top-k values. Each query is searched once at the largest
// k; a ranking prefix equals a search at the smaller k.
var Ks = []int{3, 5}

const (
	AsIs    = "as_is"
	Rewrite = "rewrite"
)

type QueryResult struct {
	Chunking   retrieval.Chunking `json:"chunking"`
	Mode       retrieval.Mode     `json:"mode"`
	ProbeID    string             `json:"probe_id"`
	Form       string             `json:"form"`
	QueryIndex int                `json:"query_index"`
	Query      string             `json:"query"`
	Hits       []retrieval.Hit    `json:"hits"`
}

// Scores are hit@k, reciprocal rank and hit@k on the probe's own sources.
// Averaged over queries or probes, RR becomes MRR.
type Scores struct {
	Hit        float64 `json:"hit"`
	RR         float64 `json:"rr"`
	PrimaryHit float64 `json:"primary_hit"`
}

func (s *Scores) addScaled(other Scores, weight float64) {
	s.Hit += other.Hit * weight
	s.RR += other.RR * weight
	s.PrimaryHit += other.PrimaryHit * weight
}

// Rewrite scores are means over the probe's rewritten queries.
type ProbeScore struct {
	ProbeID string `json:"probe_id"`
	AsIs    Scores `json:"as_is"`
	Rewrite Scores `json:"rewrite"`
}

// ConfigScore averages ProbeScore over probes. Truncated counts results the
// model-visible cap would have cut.
type ConfigScore struct {
	Mode      retrieval.Mode     `json:"mode"`
	Chunking  retrieval.Chunking `json:"chunking"`
	K         int                `json:"k"`
	AsIs      Scores             `json:"as_is"`
	Rewrite   Scores             `json:"rewrite"`
	Truncated int                `json:"truncated"`
	Probes    []ProbeScore       `json:"probes"`
}

func (s ConfigScore) Name() string {
	return fmt.Sprintf("%s/%s/k=%d", s.Mode, s.Chunking, s.K)
}

type Report struct {
	Configs     []ConfigScore `json:"configs"`
	Selected    ConfigScore   `json:"selected"`
	Differences []Difference  `json:"differences"`
	Results     []QueryResult `json:"-"`
}

func Evaluate(ctx context.Context, indexes map[retrieval.Chunking]*retrieval.Index, embedder inference.Embedder, probes []ProbeQueries, maxBytes int) (Report, error) {
	embedder = newCachingEmbedder(embedder)
	depth := slices.Max(Ks)
	var report Report
	for _, chunking := range retrieval.Chunkings {
		index, ok := indexes[chunking]
		if !ok {
			return Report{}, fmt.Errorf("no index for chunking %q", chunking)
		}
		for _, mode := range retrieval.Modes {
			var results []QueryResult
			for _, probe := range probes {
				for _, result := range queryForms(probe) {
					hits, err := index.Search(ctx, embedder, mode, result.Query, depth)
					if err != nil {
						return Report{}, fmt.Errorf("%s/%s %s: %w", mode, chunking, probe.ID, err)
					}
					result.Chunking, result.Mode, result.ProbeID, result.Hits = chunking, mode, probe.ID, hits
					results = append(results, result)
				}
			}
			for _, k := range Ks {
				report.Configs = append(report.Configs, scoreConfig(mode, chunking, k, probes, results, maxBytes))
			}
			report.Results = append(report.Results, results...)
		}
	}
	report.Selected = Select(report.Configs, len(probes))
	report.Differences = compareToSelected(report.Selected, report.Configs)
	return report, nil
}

func queryForms(probe ProbeQueries) []QueryResult {
	forms := []QueryResult{{Form: AsIs, Query: probe.Prompt}}
	for index, query := range probe.Queries {
		forms = append(forms, QueryResult{Form: Rewrite, QueryIndex: index + 1, Query: query})
	}
	return forms
}

func scoreConfig(mode retrieval.Mode, chunking retrieval.Chunking, k int, probes []ProbeQueries, results []QueryResult, maxBytes int) ConfigScore {
	score := ConfigScore{Mode: mode, Chunking: chunking, K: k}
	for _, probe := range probes {
		probeScore := ProbeScore{ProbeID: probe.ID}
		var rewrites []Scores
		for _, result := range results {
			if result.ProbeID != probe.ID {
				continue
			}
			hits := result.Hits[:min(k, len(result.Hits))]
			if _, truncated := retrieval.Render(hits, maxBytes); truncated {
				score.Truncated++
			}
			rr := reciprocalRank(hits, probe.Relevant)
			scores := Scores{Hit: indicator(rr > 0), RR: rr, PrimaryHit: indicator(reciprocalRank(hits, probe.Primary) > 0)}
			if result.Form == AsIs {
				probeScore.AsIs = scores
			} else {
				rewrites = append(rewrites, scores)
			}
		}
		for _, rewrite := range rewrites {
			probeScore.Rewrite.addScaled(rewrite, 1/float64(len(rewrites)))
		}
		score.Probes = append(score.Probes, probeScore)
	}
	for _, probe := range score.Probes {
		score.AsIs.addScaled(probe.AsIs, 1/float64(len(score.Probes)))
		score.Rewrite.addScaled(probe.Rewrite, 1/float64(len(score.Probes)))
	}
	return score
}

// reciprocalRank is 0 when no hit comes from a relevant file.
func reciprocalRank(hits []retrieval.Hit, relevant []string) float64 {
	for _, hit := range hits {
		if slices.Contains(relevant, hit.Chunk.Path) {
			return 1 / float64(hit.Rank)
		}
	}
	return 0
}

func indicator(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
