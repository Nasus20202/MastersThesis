package retrieval

import (
	"fmt"
	"strings"
)

const (
	noResults      = "No documentation matched the query."
	truncatedLabel = "\n[truncated]"
)

// Render formats hits as the model-visible search result of at most maxBytes
// and reports whether the cap cut it.
func Render(hits []Hit, maxBytes int) (string, bool) {
	if len(hits) == 0 {
		return noResults, false
	}
	blocks := make([]string, len(hits))
	for index, hit := range hits {
		blocks[index] = fmt.Sprintf("[%d] %s", hit.Rank, hit.Chunk.Rendered())
	}
	result := strings.Join(blocks, "\n\n")
	if len(result) <= maxBytes {
		return result, false
	}
	cut := max(maxBytes-len(truncatedLabel), 0)
	return result[:runeBoundary(result, cut)] + truncatedLabel, true
}
