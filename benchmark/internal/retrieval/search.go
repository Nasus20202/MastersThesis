package retrieval

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// Hit.Score is negated BM25 for lexical, cosine similarity for semantic and
// the fused RRF score for hybrid search.
type Hit struct {
	Rank    int     `json:"rank"`
	ChunkID int64   `json:"chunk_id"`
	Score   float64 `json:"score"`
	Chunk   Chunk   `json:"chunk"`
}

type scored struct {
	id    int64
	score float64
}

// Search needs the embedder only for semantic and hybrid mode.
func (i *Index) Search(ctx context.Context, embedder inference.Embedder, mode Mode, query string, k int) ([]Hit, error) {
	if k < 1 {
		return nil, errors.New("search k must be at least 1")
	}
	var ranked []scored
	var err error
	switch mode {
	case Lexical:
		ranked, err = i.lexical(ctx, query, k)
	case Semantic:
		ranked, err = i.semantic(ctx, embedder, query, k)
	case Hybrid:
		ranked, err = i.hybrid(ctx, embedder, query, k)
	default:
		return nil, fmt.Errorf("unsupported retrieval mode %q", mode)
	}
	if err != nil {
		return nil, err
	}
	return i.hits(ctx, ranked)
}

var queryTerm = regexp.MustCompile(`[\p{L}\p{N}]+`)

// lexicalQuery quotes every term so no FTS5 syntax reaches the index and
// ORs them so a chunk does not need to contain all of them.
func lexicalQuery(query string) string {
	var terms []string
	for _, term := range queryTerm.FindAllString(strings.ToLower(query), -1) {
		quoted := `"` + term + `"`
		if !slices.Contains(terms, quoted) {
			terms = append(terms, quoted)
		}
	}
	return strings.Join(terms, " OR ")
}

func (i *Index) lexical(ctx context.Context, query string, k int) ([]scored, error) {
	match := lexicalQuery(query)
	if match == "" {
		return nil, nil
	}
	ranked, err := i.rank(ctx, searchLexicalSQL, match, k)
	if err != nil {
		return nil, fmt.Errorf("lexical search: %w", err)
	}
	for index := range ranked {
		ranked[index].score = -ranked[index].score
	}
	return ranked, nil
}

func (i *Index) semantic(ctx context.Context, embedder inference.Embedder, query string, k int) ([]scored, error) {
	embedding, err := embedQuery(ctx, embedder, query)
	if err != nil {
		return nil, err
	}
	ranked, err := i.rank(ctx, searchSemanticSQL, embedding, k)
	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}
	for index := range ranked {
		ranked[index].score = 1 - ranked[index].score
	}
	// vec0 accepts only ORDER BY distance, so ties are broken here.
	slices.SortStableFunc(ranked, compareScored)
	return ranked, nil
}

func (i *Index) hybrid(ctx context.Context, embedder inference.Embedder, query string, k int) ([]scored, error) {
	if i.Hybrid.Candidates < 1 || i.Hybrid.RRFK < 1 {
		return nil, errors.New("hybrid search parameters are not set")
	}
	lexical, err := i.lexical(ctx, query, i.Hybrid.Candidates)
	if err != nil {
		return nil, err
	}
	semantic, err := i.semantic(ctx, embedder, query, i.Hybrid.Candidates)
	if err != nil {
		return nil, err
	}
	return fuse(k, i.Hybrid.RRFK, lexical, semantic), nil
}

func (i *Index) rank(ctx context.Context, query string, match any, k int) ([]scored, error) {
	rows, err := i.db.QueryContext(ctx, query, match, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ranked []scored
	for rows.Next() {
		var hit scored
		if err := rows.Scan(&hit.id, &hit.score); err != nil {
			return nil, err
		}
		ranked = append(ranked, hit)
	}
	return ranked, rows.Err()
}

func fuse(k, rrfK int, rankings ...[]scored) []scored {
	scores := make(map[int64]float64)
	for _, ranking := range rankings {
		for rank, hit := range ranking {
			scores[hit.id] += 1 / float64(rrfK+rank+1)
		}
	}
	fused := make([]scored, 0, len(scores))
	for id, score := range scores {
		fused = append(fused, scored{id: id, score: score})
	}
	slices.SortFunc(fused, compareScored)
	return fused[:min(k, len(fused))]
}

// compareScored breaks score ties by chunk ID so rankings are deterministic.
func compareScored(a, b scored) int {
	return cmp.Or(cmp.Compare(b.score, a.score), cmp.Compare(a.id, b.id))
}

func (i *Index) hits(ctx context.Context, ranked []scored) ([]Hit, error) {
	if len(ranked) == 0 {
		return nil, nil
	}
	ids := make([]int64, len(ranked))
	for index, hit := range ranked {
		ids[index] = hit.id
	}
	encodedIDs, err := json.Marshal(ids)
	if err != nil {
		return nil, err
	}
	rows, err := i.db.QueryContext(ctx, selectChunksSQL, string(encodedIDs))
	if err != nil {
		return nil, fmt.Errorf("load chunks: %w", err)
	}
	defer rows.Close()
	chunks := make(map[int64]Chunk, len(ranked))
	for rows.Next() {
		var id int64
		var chunk Chunk
		if err := rows.Scan(&id, &chunk.Path, &chunk.Title, &chunk.Headings, &chunk.Body); err != nil {
			return nil, err
		}
		chunks[id] = chunk
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	hits := make([]Hit, len(ranked))
	for index, hit := range ranked {
		chunk, ok := chunks[hit.id]
		if !ok {
			return nil, fmt.Errorf("chunk %d is missing from the index", hit.id)
		}
		hits[index] = Hit{Rank: index + 1, ChunkID: hit.id, Score: hit.score, Chunk: chunk}
	}
	return hits, nil
}
